const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');
const vm = require('node:vm');

async function loadRoomParty() {
  const timers = new Map();
  let id = 0;
  let onUpdate = null;
  let retriggers = 0;
  const session = { myTeam: [{ puuid: 'me', cellId: 0 }, { puuid: 'mate', cellId: 1 }], chatDetails: { multiUserChatId: 'room' } };
  const context = vm.createContext({
    console,
    document: { querySelectorAll: () => [] },
    setTimeout: (fn, delay) => { timers.set(++id, { fn, delay }); return id; },
    clearTimeout: (key) => timers.delete(key),
  });
  const mocks = {
    './websocket': {
      wsSend: () => true,
      onRoomPartyUpdate: (cb) => { onUpdate = cb; return () => {}; },
      onSetting: (key, cb) => { cb(true); return () => {}; },
      clearRoomPartySkins: () => {},
    },
    './api': { fetchJson: async (url) => (url.includes('current-summoner') ? { puuid: 'me' } : session) },
    './dom': { el: () => ({}) },
    './constants': { ROOM_PARTY_INDICATOR_CLASS: 'indicator' },
    './autoApply': { retriggerPrefetch: () => { retriggers++; } },
    './splash': { updateTeammateSplashes: async () => {} },
    './i18n': { t: (key) => key },
    './logger': { createLogger: () => ({ log() {} }) },
  };
  const module = new vm.SourceTextModule(fs.readFileSync(path.join(__dirname, '../src/roomParty.js'), 'utf8'), { context });
  await module.link((name) => new vm.SyntheticModule(Object.keys(mocks[name]), function () {
    for (const [key, value] of Object.entries(mocks[name])) this.setExport(key, value);
  }, { context }));
  await module.evaluate();
  const api = module.namespace;
  api.loadRoomPartySetting();
  for (let i = 0; i < 10 && !onUpdate; i++) await new Promise((resolve) => setImmediate(resolve));
  assert.ok(onUpdate, 'joined the room');
  return {
    api,
    timers,
    retriggers: () => retriggers,
    teammatePicks: (skinId) => onUpdate([{ puuid: 'mate', skinInfo: { skinId, skinName: 'Skin' } }]),
  };
}

test('a teammate change pending when champion select ends waits for the next match', async () => {
  const room = await loadRoomParty();
  room.teammatePicks('238068');
  assert.equal(room.timers.size, 1, 'retrigger debounced');
  room.api.discardPendingRetrigger();
  assert.equal(room.timers.size, 0);
  assert.equal(room.retriggers(), 0);
});

test('a teammate change during champion select still retriggers after the debounce', async () => {
  const room = await loadRoomParty();
  room.teammatePicks('36021');
  const [[key, timer]] = [...room.timers];
  room.timers.delete(key);
  timer.fn();
  assert.equal(room.retriggers(), 1);
});
