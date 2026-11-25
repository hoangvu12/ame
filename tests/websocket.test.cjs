const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');
const vm = require('node:vm');

// In-memory stand-in for the client's storage; pass the same one to several
// setups to simulate the client restarting.
function memoryStorage(initial = {}) {
  const data = new Map(Object.entries(initial));
  return {
    data,
    getItem: (key) => (data.has(key) ? data.get(key) : null),
    setItem: (key, value) => { data.set(key, String(value)); },
  };
}

async function setup({ storage = null, connect = true } = {}) {
  const timers = new Map();
  const sockets = [];
  const promises = [];
  const toasts = [];
  const notices = [];
  let id = 0;
  class Socket {
    static OPEN = 1;
    static CONNECTING = 0;
    constructor() { this.readyState = 0; this.sent = []; sockets.push(this); }
    open() { this.readyState = 1; this.onopen(); }
    send(data) { if (this.failSend) throw new Error('send failed'); this.sent.push(JSON.parse(data)); }
    close() { this.readyState = 3; this.onclose(); }
    status(status) { this.onmessage({ data: JSON.stringify({ type: 'status', status, message: status }) }); }
  }
  const context = vm.createContext({
    WebSocket: Socket,
    setTimeout: (fn, delay) => { timers.set(++id, { fn, delay }); return id; },
    clearTimeout: (key) => timers.delete(key),
    ...(storage ? { localStorage: storage } : {}),
  });
  const mocks = {
    './constants': { WS_URL: 'ws://test', WS_RECONNECT_BASE_MS: 1000, WS_RECONNECT_MAX_MS: 5000, WS_RECONNECT_JITTER_MS: 0 },
    './toast': {
      toastError() {},
      toastSuccess(message) { notices.push(message); },
      toastPromise(p, opts) { promises.push(p); toasts.push(opts); p.catch(() => {}); },
    },
    './dom': { el() {} },
    './i18n': { t: (key) => key },
    './logger': { createLogger: () => ({ log() {}, error() {} }) },
  };
  const module = new vm.SourceTextModule(fs.readFileSync(path.join(__dirname, '../src/websocket.js'), 'utf8'), { context });
  await module.link((name) => {
    if (name === './clientSettings') {
      return new vm.SourceTextModule(fs.readFileSync(path.join(__dirname, '../src/clientSettings.js'), 'utf8'), { context });
    }
    return new vm.SyntheticModule(Object.keys(mocks[name]), function () {
      for (const [key, value] of Object.entries(mocks[name])) this.setExport(key, value);
    }, { context });
  });
  await module.evaluate();
  const api = module.namespace;
  api.wsConnect();
  if (connect) sockets[0].open();
  const fire = (delay) => {
    const entry = [...timers].find(([, timer]) => timer.delay === delay);
    assert.ok(entry, `timer ${delay} exists`);
    timers.delete(entry[0]); entry[1].fn();
  };
  return { api, sockets, promises, toasts, notices, timers, fire };
}
const payload = { type: 'apply', championId: 1, skinId: 1001 };

test('disconnect rejects and clears pending apply; reconnect does not replay and permits retry', async () => {
  const { api, sockets, promises, timers, fire } = await setup();
  assert.equal(api.wsSendApply(payload), true);
  assert.equal(api.wsSendApply(payload), false);
  sockets[0].close();
  await assert.rejects(promises[0], /Connection closed/);
  assert.equal(api.isApplyInFlight(), false);
  assert.equal([...timers.values()].some(t => t.delay === 120000), false);
  assert.equal(api.wsSendApply(payload), false);
  fire(1000); sockets[1].open();
  assert.equal(sockets[1].sent.some(m => m.type === 'apply'), false);
  assert.equal(api.wsSendApply(payload), true);
  sockets[0].status('ready');
  assert.equal(api.isApplyInFlight(), true);
  sockets[1].status('ready');
  assert.equal(await promises[1], true);
  assert.equal(timers.size, 0);
});

test('timeout retires socket and ignores late old replies after retry', async () => {
  const { api, sockets, promises, fire, timers } = await setup();
  api.wsSendApply(payload);
  fire(120000);
  await assert.rejects(promises[0], /timed out/);
  assert.equal(api.isApplyInFlight(), false);
  assert.equal(api.wsSendApply(payload), false);
  fire(1000); sockets[1].open();
  assert.equal(sockets[1].sent.some(m => m.type === 'apply'), false);
  api.wsSendApply(payload);
  sockets[0].status('ready');
  assert.equal(api.isApplyInFlight(), true);
  sockets[1].status('noop');
  assert.equal(await promises[1], false);
  assert.equal(timers.size, 0);
});

for (const status of ['ready', 'noop', 'error', 'starting']) {
  test(`${status} settles apply and cancels timeout`, async () => {
    const { api, sockets, promises, timers } = await setup();
    api.wsSendApply(payload);
    sockets[0].status(status);
    if (status === 'error' || status === 'starting') await assert.rejects(promises[0]);
    else assert.equal(await promises[0], status === 'ready');
    assert.equal(api.isApplyInFlight(), false);
    assert.equal(timers.size, 0);
  });
}

// "kept": the change came too late for this match and ame kept the skins it
// had prepared. Not a failure, and not a claim that the change applied.
test('kept settles apply as success and says the change waits for the next match', async () => {
  const { api, sockets, promises, toasts, timers } = await setup();
  api.wsSendApply(payload);
  sockets[0].status('kept');
  assert.equal(await promises[0], true);
  assert.equal(api.isApplyInFlight(), false);
  assert.equal(api.isOverlayActive(), true);
  assert.equal(timers.size, 0);
  assert.equal(toasts[0].success(), 'toast.apply.kept');

  api.wsSendApply(payload);
  sockets[0].status('ready');
  await promises[1];
  assert.equal(toasts[1].success(), 'toast.apply.success');
});

test('kept without a pending apply shows a rate-limited notice', async () => {
  const { sockets, notices } = await setup();
  sockets[0].status('kept');
  sockets[0].status('kept');
  assert.deepEqual(notices, ['toast.apply.kept']);
});

// In-game "Remove Skin" flips the overlay off as it sends cleanup. While
// League runs ame declines and the skins stay on, so the state flips back;
// a server that predates the reply always cleans up and never sends it.
test('declined cleanup keeps the overlay active', async () => {
  const { api, sockets, notices } = await setup();
  api.setOverlayActive(true);
  api.wsSend({ type: 'cleanup' });
  assert.equal(api.isOverlayActive(), false);
  sockets[0].status('cleanupDeclined');
  assert.equal(api.isOverlayActive(), true);
  assert.deepEqual(notices, ['toast.apply.kept']);

  api.wsSend({ type: 'cleanup' });
  assert.equal(api.isOverlayActive(), false);
});

// ame copies the map files ahead of time; champion select that starts
// while it still does (the first time after a League update) says so.
test('map preparation status reaches listeners and the champion select notice', async () => {
  const { api, sockets, notices } = await setup();
  const seen = [];
  api.onMapPrep((status) => seen.push(status.state));
  api.noticeMapPrepIfRunning();
  assert.deepEqual(notices, []);
  const send = (msg) => sockets[0].onmessage({ data: JSON.stringify({ type: 'mapPrep', ...msg }) });
  send({ state: 'preparing', stage: 'patching', current: 1, total: 3 });
  assert.deepEqual(JSON.parse(JSON.stringify(api.getMapPrep())), { state: 'preparing', reason: '', stage: 'patching', current: 1, total: 3 });
  api.noticeMapPrepIfRunning();
  assert.deepEqual(notices, ['toast.map_prep.preparing']);
  send({ state: 'ready' });
  api.noticeMapPrepIfRunning();
  assert.deepEqual(notices, ['toast.map_prep.preparing']);
  assert.deepEqual(seen, ['preparing', 'ready']);
  const late = [];
  api.onMapPrep((status) => late.push(status.state));
  assert.deepEqual(late, ['ready']);
});

test('send exception rejects and clears timer, allowing a later explicit retry', async () => {
  const { api, sockets, promises, timers } = await setup();
  sockets[0].failSend = true;
  assert.equal(api.wsSendApply(payload), false);
  await assert.rejects(promises[0], /send failed/);
  assert.equal(api.isApplyInFlight(), false);
  assert.equal(timers.size, 0);
  sockets[0].failSend = false;
  assert.equal(api.wsSendApply(payload), true);
  sockets[0].status('ready');
  await promises[1];
});

// Settings for the client-side features live in the client, so they work
// without ame. ame's copy is taken over once, then the client copy wins.
const plain = (value) => JSON.parse(JSON.stringify(value));
const serverSays = (socket, msg) => socket.onmessage({ data: JSON.stringify(msg) });

test('client-side settings changed while ame is off apply at once and survive a restart', async () => {
  const storage = memoryStorage();
  const first = await setup({ storage, connect: false });
  const seen = [];
  first.api.onSetting('autoAccept', (v) => seen.push(v));
  first.api.setSetting('autoAccept', true);
  first.api.setAutoSelectRole('top', [1, 2], [3]);
  first.api.setChatStatus('away');
  first.api.setRandomSkinMode('top3');
  assert.deepEqual(seen, [false, true]);

  const second = await setup({ storage, connect: false });
  const restarted = [];
  second.api.onSetting('autoAccept', (v) => restarted.push(v));
  assert.deepEqual(restarted, [true]);
  assert.deepEqual(plain(second.api.getAutoSelectRolesCache()), { top: { picks: [1, 2], bans: [3] } });
  assert.equal(second.api.getRandomSkinMode(), 'top3');
  let chat = null;
  second.api.onChatStatus((c) => { chat = plain(c); });
  assert.deepEqual(chat, { availability: 'away', statusMessage: '' });
});

test('ame settings are taken over once, then ame cannot overwrite them', async () => {
  const storage = memoryStorage();
  const { api, sockets } = await setup({ storage });
  const autoAccept = [];
  const roomParty = [];
  api.onSetting('autoAccept', (v) => autoAccept.push(v));
  api.onSetting('roomParty', (v) => roomParty.push(v));
  serverSays(sockets[0], {
    type: 'settings', autoAccept: true, roomParty: true, randomSkin: 'all',
    autoSelectRoles: { mid: { picks: [7], bans: [] } }, chatAvailability: 'mobile', chatStatusMessage: '',
  });
  assert.deepEqual(autoAccept, [false, true]);
  assert.equal(api.getRandomSkinMode(), 'all');
  assert.deepEqual(plain(api.getAutoSelectRolesCache()), { mid: { picks: [7], bans: [] } });

  // A later snapshot or an echo from ame does not touch client-side settings,
  // but ame still owns the settings that need it.
  serverSays(sockets[0], { type: 'settings', autoAccept: false, roomParty: false, randomSkin: '', autoSelectRoles: {} });
  serverSays(sockets[0], { type: 'autoAccept', enabled: false });
  serverSays(sockets[0], { type: 'randomSkin', mode: '' });
  assert.deepEqual(autoAccept, [false, true]);
  assert.equal(api.getRandomSkinMode(), 'all');
  assert.deepEqual(plain(api.getAutoSelectRolesCache()), { mid: { picks: [7], bans: [] } });
  assert.deepEqual(roomParty, [true, false]);

  // Changes still reach ame, which keeps its own copy.
  api.setSetting('autoAccept', false);
  assert.deepEqual(plain(sockets[0].sent.at(-1)), { type: 'setAutoAccept', enabled: false });
  api.setSetting('roomParty', true);
  assert.deepEqual(plain(sockets[0].sent.at(-1)), { type: 'setRoomParty', enabled: true });
  assert.deepEqual(roomParty, [true, false]);
});

test('changes made before ame was first reached are not overwritten by its older copy', async () => {
  const storage = memoryStorage();
  const offline = await setup({ storage, connect: false });
  offline.api.setSetting('autoAccept', true);
  offline.api.setAutoSelectRole('top', [1], []);

  const { api, sockets } = await setup({ storage });
  serverSays(sockets[0], {
    type: 'settings', autoAccept: false, benchSwap: true, randomSkin: 'all',
    autoSelectRoles: { top: { picks: [9], bans: [] }, jungle: { picks: [2], bans: [4] } },
  });
  const values = {};
  for (const key of ['autoAccept', 'benchSwap']) api.onSetting(key, (v) => { values[key] = v; });
  assert.deepEqual(values, { autoAccept: true, benchSwap: true });
  assert.equal(api.getRandomSkinMode(), 'all');
  assert.deepEqual(plain(api.getAutoSelectRolesCache()), {
    top: { picks: [1], bans: [] },
    jungle: { picks: [2], bans: [4] },
  });
});
