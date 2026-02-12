import { sendHttpProxy, requestExtensions, addExtension as wsAddExtension, addExtensionFromFile as wsAddExtensionFromFile, removeExtension as wsRemoveExtension } from './websocket';

let loadedSources = [];
let loaded = false;

const nativeFetch = window.fetch.bind(window);

function createHelpers() {
  return {
    fetch: proxyFetch,
    clientFetch: nativeFetch,
    parseHtml(html) {
      const parser = new DOMParser();
      return parser.parseFromString(html, 'text/html');
    },
  };
}

async function proxyFetch(url, opts = {}) {
  const maxRetries = 3;
  let lastError;

  for (let attempt = 0; attempt < maxRetries; attempt++) {
    if (attempt > 0) {
      await new Promise(r => setTimeout(r, attempt * 1000));
    }

    const result = await sendHttpProxy(url, opts.method, opts.headers, opts.body);

    if (result.error) {
      lastError = new Error(result.error);
      if (attempt < maxRetries - 1 && isRetryableError(result.error)) {
        continue;
      }
      throw lastError;
    }

    let bodyText = result.body;
    if (result.isBase64) {
      bodyText = atob(result.body);
    }

    return {
      ok: result.status >= 200 && result.status < 300,
      status: result.status,
      headers: result.headers || {},
      text: () => Promise.resolve(bodyText),
      json: () => Promise.resolve(JSON.parse(bodyText)),
    };
  }

  throw lastError;
}

function isRetryableError(msg) {
  return /EOF|closed|reset|refused|timeout/i.test(msg);
}

function evaluateExtension(source) {
  try {
    const fn = new Function('createSource', `
      ${source}
      if (typeof createSource === 'function') return createSource;
      return null;
    `);
    const createSource = fn();
    if (!createSource) return null;
    const helpers = createHelpers();
    return createSource(helpers);
  } catch (e) {
    console.error('[extensionManager] Failed to evaluate extension:', e);
    return null;
  }
}

export function loadExtensions() {
  return new Promise((resolve) => {
    requestExtensions((files) => {
      loadedSources = [];
      for (const file of files) {
        const source = evaluateExtension(file.source);
        if (source && source.id && source.name) {
          source._filename = file.filename;
          loadedSources.push(source);
        }
      }
      loaded = true;
      resolve(loadedSources);
    });
  });
}

export function getLoadedSources() {
  return loadedSources;
}

export function isLoaded() {
  return loaded;
}

export function addExtension(url) {
  return new Promise((resolve) => {
    wsAddExtension(url, (ext) => {
      if (ext) {
        const source = evaluateExtension(ext.source);
        if (source && source.id && source.name) {
          source._filename = ext.filename;
          loadedSources.push(source);
        }
      }
      resolve(loadedSources);
    });
  });
}

export function addExtensionFromFile() {
  return new Promise((resolve) => {
    wsAddExtensionFromFile((ext) => {
      if (ext) {
        const source = evaluateExtension(ext.source);
        if (source && source.id && source.name) {
          source._filename = ext.filename;
          loadedSources.push(source);
        }
      }
      resolve(loadedSources);
    });
  });
}

export function removeExtension(filename) {
  return new Promise((resolve) => {
    wsRemoveExtension(filename, () => {
      loadedSources = loadedSources.filter(s => s._filename !== filename);
      resolve(loadedSources);
    });
  });
}
