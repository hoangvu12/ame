const fs = require('fs');
const path = require('path');

const localesDir = path.join(__dirname, '..', 'src', 'locales');
const baseLocale = 'en_US';
const manifestPath = path.join(localesDir, 'manifest.json');

function readJson(filePath) {
  const raw = fs.readFileSync(filePath, 'utf8');
  return JSON.parse(raw);
}

function flatten(obj, prefix = '', out = {}) {
  for (const [key, value] of Object.entries(obj)) {
    const next = prefix ? `${prefix}.${key}` : key;
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      flatten(value, next, out);
    } else {
      out[next] = true;
    }
  }
  return out;
}

function loadManifest() {
  if (!fs.existsSync(manifestPath)) {
    throw new Error(`Missing manifest: ${manifestPath}`);
  }
  const list = readJson(manifestPath);
  if (!Array.isArray(list)) {
    throw new Error('manifest.json must be an array of locale codes');
  }
  return list;
}

function main() {
  const basePath = path.join(localesDir, `${baseLocale}.json`);
  if (!fs.existsSync(basePath)) {
    console.error(`Missing base locale: ${basePath}`);
    process.exit(1);
  }

  const manifest = loadManifest();
  const base = readJson(basePath);
  const baseKeys = Object.keys(flatten(base));
  const baseSet = new Set(baseKeys);

  let ok = true;

  for (const code of manifest) {
    const filePath = path.join(localesDir, `${code}.json`);
    if (!fs.existsSync(filePath)) {
      console.error(`[MISSING] ${code}.json (listed in manifest)`);
      ok = false;
      continue;
    }

    let data;
    try {
      data = readJson(filePath);
    } catch (err) {
      console.error(`[INVALID] ${code}.json: ${err.message}`);
      ok = false;
      continue;
    }

    const keys = Object.keys(flatten(data));
    const keySet = new Set(keys);

    const missing = baseKeys.filter(k => !keySet.has(k));
    const extra = keys.filter(k => !baseSet.has(k));

    if (missing.length || extra.length) {
      ok = false;
      console.error(`\n[DIFF] ${code}.json`);
      if (missing.length) {
        console.error(`  Missing (${missing.length}):`);
        for (const k of missing) console.error(`  - ${k}`);
      }
      if (extra.length) {
        console.error(`  Extra (${extra.length}):`);
        for (const k of extra) console.error(`  - ${k}`);
      }
    }
  }

  const localeFiles = fs.readdirSync(localesDir)
    .filter(f => f.endsWith('.json') && f !== 'manifest.json')
    .map(f => path.basename(f, '.json'));

  for (const code of localeFiles) {
    if (!manifest.includes(code)) {
      ok = false;
      console.error(`[UNLISTED] ${code}.json (missing from manifest)`);
    }
  }

  if (!ok) {
    console.error('\nLocale validation failed.');
    process.exit(1);
  }

  console.log('Locale validation passed.');
}

main();
