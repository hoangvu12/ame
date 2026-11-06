const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');
const vm = require('node:vm');

async function setup() {
  const timers = new Map();
  const sockets = [];
  const promises = [];
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
  });
  const mocks = {
    './constants': { WS_URL: 'ws://test', WS_RECONNECT_BASE_MS: 1000, WS_RECONNECT_MAX_MS: 5000, WS_RECONNECT_JITTER_MS: 0 },
    './toast': { toastError() {}, toastPromise(p) { promises.push(p); p.catch(() => {}); } },
    './dom': { el() {} },
    './i18n': { t: (key) => key },
    './logger': { createLogger: () => ({ log() {}, error() {} }) },
  };
  const module = new vm.SourceTextModule(fs.readFileSync(path.join(__dirname, '../src/websocket.js'), 'utf8'), { context });
  await module.link((name) => new vm.SyntheticModule(Object.keys(mocks[name]), function () {
    for (const [key, value] of Object.entries(mocks[name])) this.setExport(key, value);
  }, { context }));
  await module.evaluate();
  const api = module.namespace;
  api.wsConnect();
  sockets[0].open();
  const fire = (delay) => {
    const entry = [...timers].find(([, timer]) => timer.delay === delay);
    assert.ok(entry, `timer ${delay} exists`);
    timers.delete(entry[0]); entry[1].fn();
  };
  return { api, sockets, promises, timers, fire };
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
