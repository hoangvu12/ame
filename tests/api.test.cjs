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

async function loadApi(fetch, suppressed) {
  const filePath = path.resolve(__dirname, '../src/api.js');
  const context = vm.createContext({ fetch, JSON, Map, Set });
  const module = new vm.SourceTextModule(fs.readFileSync(filePath, 'utf8'), {
    context,
    identifier: pathToFileURL(filePath).href,
  });
  await module.link(async (specifier) => {
    if (specifier === './suppress') {
      return syntheticModule(context, specifier, {
        suppressNextSkinEvent: (skinId) => suppressed.push(skinId),
      });
    }
    if (specifier === './logger') {
      return syntheticModule(context, specifier, {
        createLogger: () => ({ log: () => {} }),
      });
    }
    throw new Error(`Missing mock for ${specifier}`);
  });
  await module.evaluate();
  return module.namespace;
}

test('owned selection is sent after an in-flight base selection', async () => {
  const requests = [];
  const suppressed = [];
  let finishBase;
  const fetch = async (_url, options) => {
    const skinId = JSON.parse(options.body).selectedSkinId;
    requests.push(skinId);
    if (skinId === 1000) {
      await new Promise((resolve) => { finishBase = resolve; });
    }
    return { ok: true, status: 204 };
  };
  const api = await loadApi(fetch, suppressed);

  const baseRequest = api.forceDefaultSkin(1);
  await new Promise((resolve) => setImmediate(resolve));
  const ownedRequest = api.selectSkin(1002);
  await new Promise((resolve) => setImmediate(resolve));

  assert.deepEqual(requests, [1000]);
  finishBase();
  await Promise.all([baseRequest, ownedRequest]);
  assert.deepEqual(requests, [1000, 1002]);
  assert.deepEqual(suppressed, [1000]);
});

test('a newer unowned selection skips a queued owned selection', async () => {
  const requests = [];
  let finishFirstBase;
  const fetch = async (_url, options) => {
    const skinId = JSON.parse(options.body).selectedSkinId;
    requests.push(skinId);
    if (requests.length === 1) {
      await new Promise((resolve) => { finishFirstBase = resolve; });
    }
    return { ok: true, status: 204 };
  };
  const api = await loadApi(fetch, []);

  const firstBase = api.forceDefaultSkin(1);
  await new Promise((resolve) => setImmediate(resolve));
  const staleOwned = api.selectSkin(1002);
  api.cancelPendingSkinSelections();
  const latestBase = api.forceDefaultSkin(1);

  finishFirstBase();
  await Promise.all([firstBase, staleOwned, latestBase]);
  assert.deepEqual(requests, [1000, 1000]);
});
