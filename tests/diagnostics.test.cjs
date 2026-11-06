const test = require('node:test');
const assert = require('node:assert/strict');
const vm = require('node:vm');
const fs = require('node:fs');

test('diagnostics uses explicit local attachment download without upload', async () => {
  let clicked = false, removed = false, appended = false;
  const anchor = { click() { clicked = true; }, remove() { removed = true; } };
  const context = vm.createContext({ document: {
    createElement(tag) { assert.equal(tag, 'a'); return anchor; },
    body: { appendChild(a) { assert.equal(a, anchor); appended = true; } },
  } });
  const mod = new vm.SourceTextModule(fs.readFileSync('src/diagnostics.js', 'utf8'), { context });
  await mod.link(() => new vm.SyntheticModule(['WS_URL'], function () {
    this.setExport('WS_URL', 'ws://localhost:18765');
  }, { context }));
  await mod.evaluate();
  mod.namespace.downloadDiagnosticZIP();
  assert.equal(anchor.href, 'http://localhost:18765/diagnostics.zip');
  assert.equal(anchor.download, 'ame-diagnostics.zip');
  assert.ok(clicked && removed && appended);
});
