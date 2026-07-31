const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');
const vm = require('node:vm');
const { pathToFileURL } = require('node:url');

function syntheticModule(context, identifier, exports) {
  const names = Object.keys(exports);
  return new vm.SyntheticModule(names, function () {
    for (const name of names) this.setExport(name, exports[name]);
  }, { context, identifier });
}

async function loadAutoApply(mocks) {
  const context = vm.createContext({
    console,
    Date,
    setTimeout,
    clearTimeout,
  });
  const sourcePath = path.resolve(__dirname, '../src/autoApply.js');
  const statePath = path.resolve(__dirname, '../src/state.js');
  const modules = new Map();

  const loadSource = (filePath) => {
    if (modules.has(filePath)) return modules.get(filePath);
    const module = new vm.SourceTextModule(fs.readFileSync(filePath, 'utf8'), {
      context,
      identifier: pathToFileURL(filePath).href,
    });
    modules.set(filePath, module);
    return module;
  };

  const autoApply = loadSource(sourcePath);
  await autoApply.link(async (specifier) => {
    if (specifier === './state') return loadSource(statePath);
    const exports = mocks[specifier];
    if (!exports) throw new Error(`Missing mock for ${specifier}`);
    return syntheticModule(context, specifier, exports);
  });
  await autoApply.evaluate();
  return autoApply.namespace;
}

test('owned selection invalidates a previously forced unowned skin', async () => {
  const skins = [
    { id: 1000, name: 'Default', isBase: true },
    { id: 1001, name: 'Unowned' },
    { id: 1002, name: 'Owned' },
  ];
  let currentSkin = 'Unowned';
  const applied = [];
  const selected = [];

  const autoApply = await loadAutoApply({
    './api': {
      loadChampionSkins: async () => skins,
      getChampionSkins: () => skins,
      getMyChampionId: async () => 1,
      getChampionName: async () => 'Champion',
      fetchJson: async () => null,
      forceDefaultSkin: async () => true,
      selectSkin: async (skinId) => { selected.push(skinId); return true; },
      cancelPendingSkinSelections: () => {},
    },
    './skin': {
      readCurrentSkin: () => currentSkin,
      findSkinByName: (list, name) => list.find((skin) => skin.name === name) || null,
      isDefaultSkin: (skin) => skin.isBase === true,
    },
    './websocket': {
      wsSend: () => true,
      wsSendApply: (payload) => { applied.push(payload); return true; },
      isApplyInFlight: () => false,
      isOverlayActive: () => false,
      hasEnabledCustomMods: () => false,
    },
    './ui': { setButtonState: () => {} },
    './constants': { PREFETCH_DEBOUNCE_MS: 0, OWNED_SELECTION_DELAY_MS: 0 },
    './roomParty': { notifySkinChange: () => {} },
    './historicSkin': { recordHistoricSkin: () => {} },
    './i18n': { t: (key) => key },
    './logger': { createLogger: () => ({ log: () => {} }) },
  });

  autoApply.checkAutoApply(1, false);
  await new Promise((resolve) => setTimeout(resolve, 10));

  currentSkin = 'Owned';
  autoApply.checkAutoApply(1, true);
  await autoApply.forceApplyIfNeeded();

  assert.deepEqual(applied, []);
  assert.deepEqual(selected, [1002]);
});

test('explicit chroma payload remains applicable on an owned base skin', async () => {
  const skins = [{ id: 1002, name: 'Owned' }];
  const applied = [];
  const autoApply = await loadAutoApply({
    './api': {
      loadChampionSkins: async () => skins,
      getChampionSkins: () => skins,
      getMyChampionId: async () => 1,
      getChampionName: async () => 'Champion',
      fetchJson: async () => null,
      forceDefaultSkin: async () => true,
      selectSkin: async () => true,
      cancelPendingSkinSelections: () => {},
    },
    './skin': {
      readCurrentSkin: () => 'Owned',
      findSkinByName: (list, name) => list.find((skin) => skin.name === name) || null,
      isDefaultSkin: () => false,
    },
    './websocket': {
      wsSend: () => true,
      wsSendApply: (payload) => { applied.push(payload); return true; },
      isApplyInFlight: () => false,
      isOverlayActive: () => false,
      hasEnabledCustomMods: () => false,
    },
    './ui': { setButtonState: () => {} },
    './constants': { PREFETCH_DEBOUNCE_MS: 0, OWNED_SELECTION_DELAY_MS: 0 },
    './roomParty': { notifySkinChange: () => {} },
    './historicSkin': { recordHistoricSkin: () => {} },
    './i18n': { t: (key) => key },
    './logger': { createLogger: () => ({ log: () => {} }) },
  });

  autoApply.checkAutoApply(1, true);
  autoApply.prefetchChroma(1, 10021, 1002, 'Champion', 'Owned', 'Chroma');
  await autoApply.forceApplyIfNeeded();

  assert.equal(applied.length, 1);
  assert.equal(applied[0].skinId, 10021);
});

test('visible forced base does not discard the unowned payload', async () => {
  const skins = [
    { id: 1000, name: 'Default', isBase: true },
    { id: 1001, name: 'Unowned' },
  ];
  let currentSkin = 'Unowned';
  const applied = [];
  const autoApply = await loadAutoApply({
    './api': {
      loadChampionSkins: async () => skins,
      getChampionSkins: () => skins,
      getMyChampionId: async () => 1,
      getChampionName: async () => 'Champion',
      fetchJson: async () => null,
      forceDefaultSkin: async () => true,
      selectSkin: async () => true,
      cancelPendingSkinSelections: () => {},
    },
    './skin': {
      readCurrentSkin: () => currentSkin,
      findSkinByName: (list, name) => list.find((skin) => skin.name === name) || null,
      isDefaultSkin: (skin) => skin.isBase === true,
    },
    './websocket': {
      wsSend: () => true,
      wsSendApply: (payload) => { applied.push(payload); return true; },
      isApplyInFlight: () => false,
      isOverlayActive: () => false,
      hasEnabledCustomMods: () => false,
    },
    './ui': { setButtonState: () => {} },
    './constants': { PREFETCH_DEBOUNCE_MS: 0, OWNED_SELECTION_DELAY_MS: 0 },
    './roomParty': { notifySkinChange: () => {} },
    './historicSkin': { recordHistoricSkin: () => {} },
    './i18n': { t: (key) => key },
    './logger': { createLogger: () => ({ log: () => {} }) },
  });

  autoApply.checkAutoApply(1, false);
  await new Promise((resolve) => setTimeout(resolve, 10));
  currentSkin = 'Default';
  autoApply.checkAutoApply(1, true);
  await autoApply.forceApplyIfNeeded();

  assert.equal(applied.length, 1);
  assert.equal(applied[0].skinId, 1001);
});
