// Settings for the features that run entirely inside the League client
// (auto accept, bench swap, auto pick/ban, chat status, random skin). They are
// kept here, in the client, so those features work even when ame is not
// running. ame used to be the only place they were saved; its copy is taken
// over once (see mergeServerSettings) and after that the client copy wins.

const STORAGE_KEY = 'ame:settings';

export const CLIENT_BOOL_KEYS = ['autoAccept', 'benchSwap', 'benchSwapSkipCooldown', 'autoSelect'];

export function defaultValues() {
  return {
    autoAccept: false,
    benchSwap: false,
    benchSwapSkipCooldown: false,
    autoSelect: false,
    autoSelectRoles: {},
    chatAvailability: '',
    chatStatusMessage: '',
    randomSkin: '',
  };
}

// migrated: ame's copy has already been taken over.
// touched: keys changed here before that happened; ame's older value must not
// overwrite them. Role lists are tracked per role ("autoSelectRoles.top").
function emptyState() {
  return { migrated: false, touched: [], values: defaultValues() };
}

function normalize(raw) {
  const state = emptyState();
  if (!raw || typeof raw !== 'object') return state;
  state.migrated = !!raw.migrated;
  state.touched = Array.isArray(raw.touched) ? raw.touched.filter(k => typeof k === 'string') : [];
  const v = raw.values && typeof raw.values === 'object' ? raw.values : {};
  for (const key of CLIENT_BOOL_KEYS) {
    if (key in v) state.values[key] = !!v[key];
  }
  if (v.autoSelectRoles && typeof v.autoSelectRoles === 'object') {
    state.values.autoSelectRoles = normalizeRoles(v.autoSelectRoles);
  }
  if (typeof v.chatAvailability === 'string') state.values.chatAvailability = v.chatAvailability;
  if (typeof v.chatStatusMessage === 'string') state.values.chatStatusMessage = v.chatStatusMessage;
  if (typeof v.randomSkin === 'string') state.values.randomSkin = v.randomSkin;
  return state;
}

function normalizeRoles(roles) {
  const out = {};
  for (const [role, cfg] of Object.entries(roles || {})) {
    out[role] = {
      picks: Array.isArray(cfg?.picks) ? cfg.picks.filter(Number.isFinite) : [],
      bans: Array.isArray(cfg?.bans) ? cfg.bans.filter(Number.isFinite) : [],
    };
  }
  return out;
}

// Pengu's DataStore is written to a file next to Pengu, so it survives client
// restarts. localStorage is kept as a second copy in case DataStore is missing.
function dataStore() {
  return typeof DataStore !== 'undefined' && DataStore && typeof DataStore.get === 'function' ? DataStore : null;
}

function webStorage() {
  try {
    return typeof localStorage !== 'undefined' ? localStorage : null;
  } catch {
    return null;
  }
}

function parse(raw) {
  if (raw == null) return null;
  if (typeof raw === 'object') return raw;
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

function readStored() {
  const ds = dataStore();
  if (ds) {
    try {
      const value = parse(ds.get(STORAGE_KEY));
      if (value) return value;
    } catch { /* fall through to localStorage */ }
  }
  const ls = webStorage();
  if (ls) {
    try {
      return parse(ls.getItem(STORAGE_KEY));
    } catch { /* unreadable */ }
  }
  return null;
}

function writeStored(state) {
  const json = JSON.stringify(state);
  const ds = dataStore();
  if (ds) {
    try {
      const result = ds.set(STORAGE_KEY, json);
      if (result && typeof result.catch === 'function') result.catch(() => {});
    } catch { /* keep the localStorage copy */ }
  }
  const ls = webStorage();
  if (ls) {
    try { ls.setItem(STORAGE_KEY, json); } catch { /* quota exceeded */ }
  }
}

let state = null;

function current() {
  if (!state) state = normalize(readStored());
  return state;
}

/** Current client-side setting values (do not mutate). */
export function getClientSettings() {
  return current().values;
}

function touch(key) {
  const s = current();
  if (!s.migrated && !s.touched.includes(key)) s.touched.push(key);
}

export function setClientBool(key, enabled) {
  const s = current();
  s.values[key] = !!enabled;
  touch(key);
  writeStored(s);
}

export function setClientAutoSelectRole(role, picks, bans) {
  const s = current();
  s.values.autoSelectRoles = {
    ...s.values.autoSelectRoles,
    [role]: { picks: [...(picks || [])], bans: [...(bans || [])] },
  };
  touch(`autoSelectRoles.${role}`);
  writeStored(s);
}

export function setClientChatStatus(availability, statusMessage) {
  const s = current();
  s.values.chatAvailability = availability || '';
  s.values.chatStatusMessage = statusMessage || '';
  touch('chatStatus');
  writeStored(s);
}

export function setClientRandomSkin(mode) {
  const s = current();
  s.values.randomSkin = mode || '';
  touch('randomSkin');
  writeStored(s);
}

/**
 * Take over the settings ame saved, the first time ame is reached. Values
 * changed here before that are kept. Afterwards ame's copy is ignored.
 * Returns true when anything was taken from ame.
 */
export function mergeServerSettings(msg) {
  const s = current();
  if (s.migrated || !msg) return false;
  const touched = new Set(s.touched);
  for (const key of CLIENT_BOOL_KEYS) {
    if (key in msg && !touched.has(key)) s.values[key] = !!msg[key];
  }
  if (msg.autoSelectRoles && typeof msg.autoSelectRoles === 'object') {
    const serverRoles = normalizeRoles(msg.autoSelectRoles);
    const roles = { ...s.values.autoSelectRoles };
    for (const [role, cfg] of Object.entries(serverRoles)) {
      if (!touched.has(`autoSelectRoles.${role}`)) roles[role] = cfg;
    }
    s.values.autoSelectRoles = roles;
  }
  if (('chatAvailability' in msg || 'chatStatusMessage' in msg) && !touched.has('chatStatus')) {
    s.values.chatAvailability = msg.chatAvailability || '';
    s.values.chatStatusMessage = msg.chatStatusMessage || '';
  }
  if ('randomSkin' in msg && !touched.has('randomSkin')) {
    s.values.randomSkin = msg.randomSkin || '';
  }
  s.migrated = true;
  s.touched = [];
  writeStored(s);
  return true;
}

// Test hook: forget the in-memory copy so the next read goes to storage.
export function resetClientSettingsForTests() {
  state = null;
}
