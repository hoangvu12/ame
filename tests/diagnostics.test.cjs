const test = require('node:test');
const assert = require('node:assert/strict');
const vm = require('node:vm');
const fs = require('node:fs');

async function load(exporter) {
  let feedback;
  const context = vm.createContext({});
  const mod = new vm.SourceTextModule(fs.readFileSync('src/diagnostics.js', 'utf8'), { context });
  await mod.link((name) => {
    const values = name === './websocket' ? { exportDiagnostics: exporter } : {
      toastPromise: (promise, options) => { feedback = { promise, options }; },
    };
    return new vm.SyntheticModule(Object.keys(values), function () {
      for (const [key, value] of Object.entries(values)) this.setExport(key, value);
    }, { context });
  });
  await mod.evaluate();
  return { download: mod.namespace.downloadDiagnosticZIP, feedback: () => feedback };
}

test('export saves through backend and reports the actual saved path, without browser downloads', async () => {
  const result = { path: 'C:\\logs\\ame.zip' };
  const app = await load(() => Promise.resolve(result));
  assert.equal(await app.download(), result);
  assert.match(app.feedback().options.loading, /Saving/);
  assert.match(app.feedback().options.success(result), /C:\\logs\\ame.zip/);
});

test('export surfaces backend failure instead of silently doing nothing', async () => {
  const error = new Error('Ame is disconnected');
  const app = await load(() => Promise.reject(error));
  await assert.rejects(app.download(), /disconnected/);
  assert.equal(app.feedback().options.error(error), error.message);
});
