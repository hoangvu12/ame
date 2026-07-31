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

async function loadSuppress(window) {
  const filePath = path.resolve(__dirname, '../src/suppress.js');
  const context = vm.createContext({ window, JSON, setTimeout, clearTimeout });
  const module = new vm.SourceTextModule(fs.readFileSync(filePath, 'utf8'), {
    context,
    identifier: pathToFileURL(filePath).href,
  });
  await module.link((specifier) => syntheticModule(context, specifier, {
    createLogger: () => ({ log: () => {} }),
  }));
  await module.evaluate();
  return module.namespace;
}

test('suppression only drops the forced base skin event', async () => {
  const forwarded = [];
  const socket = { onmessage: (event) => forwarded.push(JSON.parse(event.data)[2].data.selectedSkinId) };
  const window = {
    rcp: {
      postInit: (_name, callback) => callback({ champSelectBinding: { socket: { _websocket: socket } } }),
    },
  };
  const suppress = await loadSuppress(window);
  suppress.initSkinSuppression();
  suppress.suppressNextSkinEvent(1000);

  const event = (selectedSkinId) => ({
    data: JSON.stringify([8, 'OnJsonApiEvent', {
      uri: '/lol-champ-select/v1/skin-selector-info',
      data: { selectedSkinId },
    }]),
  });
  socket.onmessage(event(1002));
  socket.onmessage(event(1000));
  socket.onmessage(event(1000));

  assert.deepEqual(forwarded, [1002, 1000]);
  suppress.disarmSkinSuppression();
});
