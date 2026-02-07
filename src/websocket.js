import { WS_URL, WS_RECONNECT_BASE_MS, WS_RECONNECT_MAX_MS } from './constants';
import { toastError, toastPromise } from './toast';
import { el } from './dom';
import { t } from './i18n';
import { createLogger } from './logger';

const logger = createLogger('ws');

let ws = null;
let wsReconnectDelay = WS_RECONNECT_BASE_MS;
let wsReconnectTimer = null;

// Pending apply promise — resolved/rejected by incoming status messages
let applyResolve = null;
let applyReject = null;

// Overlay tracking
let lastApplyPayload = null;
let overlayActive = false;

// One-shot callback for gamePath response
let gamePathCallback = null;

// One-shot callback for logs response
let logsCallback = null;

// Settings: local cache + pub/sub listeners keyed by setting name
const settingsCache = {};
const settingsListeners = {};

// Auto-select roles: separate cache for complex (non-boolean) config
let autoSelectRolesCache = {};
const autoSelectRolesListeners = [];

// Random skin: mode cache + listeners
let randomSkinCache = '';
const randomSkinListeners = [];

// Chat status: separate cache for non-boolean config
let chatStatusCache = { availability: '', statusMessage: '' };
const chatStatusListeners = [];

// Room party: teammate update listeners
const roomPartyListeners = [];

// Connection state listeners
let wsConnected = false;
const connectionListeners = [];

function setConnected(v) {
  if (wsConnected === v) return;
  wsConnected = v;
  connectionListeners.forEach(cb => cb(wsConnected));
}

/**
 * Register a listener for a boolean setting.
 * Fires immediately with cached value (if available) and on every update.
 * Returns an unsubscribe function.
 */
export function onSetting(key, cb) {
  if (!settingsListeners[key]) settingsListeners[key] = [];
  settingsListeners[key].push(cb);

  // Fire immediately with cached value if we have one
  if (key in settingsCache) cb(settingsCache[key]);

  return () => {
    const arr = settingsListeners[key];
    if (arr) {
      const idx = arr.indexOf(cb);
      if (idx !== -1) arr.splice(idx, 1);
    }
  };
}

/** Re-fetch all settings from server. */
export function refreshSettings() {
  wsSend({ type: 'getSettings' });
}

function applySetting(key, value) {
  const v = !!value;
  settingsCache[key] = v;
  if (settingsListeners[key]) {
    settingsListeners[key].forEach(cb => cb(v));
  }
}

/**
 * Register a listener for auto-select role config changes.
 * Fires immediately with cached value and on every update.
 * Returns an unsubscribe function.
 */
export function onAutoSelectRoles(cb) {
  autoSelectRolesListeners.push(cb);
  cb(autoSelectRolesCache);
  return () => {
    const idx = autoSelectRolesListeners.indexOf(cb);
    if (idx !== -1) autoSelectRolesListeners.splice(idx, 1);
  };
}

export function getAutoSelectRolesCache() {
  return autoSelectRolesCache;
}

/**
 * Register a listener for chat status config changes.
 * Fires immediately with cached value and on every update.
 * Returns an unsubscribe function.
 */
export function onChatStatus(cb) {
  chatStatusListeners.push(cb);
  cb(chatStatusCache);
  return () => {
    const idx = chatStatusListeners.indexOf(cb);
    if (idx !== -1) chatStatusListeners.splice(idx, 1);
  };
}

/**
 * Register a listener for random skin mode changes.
 * Fires immediately with cached value and on every update.
 * Returns an unsubscribe function.
 */
export function onRandomSkin(cb) {
  randomSkinListeners.push(cb);
  cb(randomSkinCache);
  return () => {
    const idx = randomSkinListeners.indexOf(cb);
    if (idx !== -1) randomSkinListeners.splice(idx, 1);
  };
}

export function getRandomSkinMode() {
  return randomSkinCache;
}

export function setRandomSkinMode(mode) {
  applyRandomSkinMode(mode);
}

function applyRandomSkinMode(mode) {
  randomSkinCache = mode || '';
  randomSkinListeners.forEach(cb => cb(randomSkinCache));
}

function applyChatStatusConfig(availability, statusMessage) {
  chatStatusCache = { availability: availability || '', statusMessage: statusMessage || '' };
  chatStatusListeners.forEach(cb => cb(chatStatusCache));
}

/**
 * Register a listener for room party teammate updates.
 * Fires whenever the backend sends a roomPartyUpdate message.
 * Returns an unsubscribe function.
 */
export function onRoomPartyUpdate(cb) {
  roomPartyListeners.push(cb);
  return () => {
    const idx = roomPartyListeners.indexOf(cb);
    if (idx !== -1) roomPartyListeners.splice(idx, 1);
  };
}

function applyAutoSelectRoles(roles) {
  autoSelectRolesCache = roles || {};
  autoSelectRolesListeners.forEach(cb => cb(autoSelectRolesCache));
}

function applyAutoSelectRole(role, picks, bans) {
  autoSelectRolesCache = { ...autoSelectRolesCache, [role]: { picks: picks || [], bans: bans || [] } };
  autoSelectRolesListeners.forEach(cb => cb(autoSelectRolesCache));
}

export function wsConnect() {
  if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) return;
  try {
    ws = new WebSocket(WS_URL);
    ws.onopen = () => {
      logger.log('WebSocket connected');
      setConnected(true);
      wsReconnectDelay = WS_RECONNECT_BASE_MS;
      // Hydrate all state from server on connect/reconnect
      ws.send(JSON.stringify({ type: 'query' }));
      ws.send(JSON.stringify({ type: 'getSettings' }));
    };
    ws.onmessage = (e) => {
      try {
        const msg = JSON.parse(e.data);
        if (msg.type === 'state') {
          if (msg.championId && msg.skinId) {
            lastApplyPayload = {
              championId: Number(msg.championId),
              skinId: Number(msg.skinId),
              baseSkinId: msg.baseSkinId ? Number(msg.baseSkinId) : Number(msg.skinId),
              championName: msg.championName || null,
              skinName: msg.skinName || null,
              chromaName: msg.chromaName || null,
            };
          }
          overlayActive = !!msg.overlayActive;
          logger.log('State from server:', overlayActive ? 'active' : 'inactive', lastApplyPayload);
        } else if (msg.type === 'roomPartyUpdate') {
          roomPartyListeners.forEach(cb => cb(msg.teammates || []));
        } else if (msg.type === 'gamePath') {
          if (gamePathCallback) {
            gamePathCallback(msg.path || '');
            gamePathCallback = null;
          }
        } else if (msg.type === 'logs') {
          if (logsCallback) {
            logsCallback({
              version: msg.version || 'unknown',
              entries: msg.entries || [],
            });
            logsCallback = null;
          }
        } else if (msg.type === 'settings') {
          // Batch settings snapshot — update all registered keys
          for (const key of Object.keys(settingsListeners)) {
            if (key in msg) applySetting(key, msg[key]);
          }
          if (msg.autoSelectRoles) applyAutoSelectRoles(msg.autoSelectRoles);
          if ('chatAvailability' in msg || 'chatStatusMessage' in msg) {
            applyChatStatusConfig(msg.chatAvailability, msg.chatStatusMessage);
          }
          if ('randomSkin' in msg) {
            applyRandomSkinMode(msg.randomSkin);
          }
        } else if (msg.type === 'autoSelectRole') {
          applyAutoSelectRole(msg.role, msg.picks, msg.bans);
        } else if (msg.type === 'randomSkin') {
          applyRandomSkinMode(msg.mode);
        } else if (msg.type === 'chatStatus') {
          applyChatStatusConfig(msg.availability, msg.statusMessage);
        } else if (settingsListeners[msg.type] && 'enabled' in msg) {
          // Individual setting response (from set* calls)
          applySetting(msg.type, msg.enabled);
        } else if (msg.type === 'status') {
          if (msg.status === 'ready' && applyResolve) {
            applyResolve();
            applyResolve = null;
            applyReject = null;
          } else if (msg.status === 'error') {
            if (applyReject) {
              applyReject(new Error(msg.message));
              applyResolve = null;
              applyReject = null;
            } else {
              toastError(msg.message);
            }
          }
        }
      } catch (err) {
        logger.error('onmessage error:', err);
      }
    };
    ws.onclose = () => {
      setConnected(false);
      wsScheduleReconnect();
    };
    ws.onerror = () => {};
  } catch {
    setConnected(false);
    wsScheduleReconnect();
  }
}

function wsScheduleReconnect() {
  if (wsReconnectTimer) return;
  wsReconnectTimer = setTimeout(() => {
    wsReconnectTimer = null;
    wsReconnectDelay = Math.min(wsReconnectDelay * 2, WS_RECONNECT_MAX_MS);
    wsConnect();
  }, wsReconnectDelay);
}

export function getLastApplyPayload() { return lastApplyPayload; }
export function isApplyInFlight() { return !!applyResolve; }
export function isOverlayActive() { return overlayActive; }
export function setOverlayActive(v) { overlayActive = v; }
export function onGamePath(cb) { gamePathCallback = cb; }
export function isConnected() { return wsConnected; }
export function requestLogs(cb) {
  logsCallback = cb;
  wsSend({ type: 'getLogs' });
}
export function onConnection(cb) {
  connectionListeners.push(cb);
  cb(wsConnected);
  return () => {
    const idx = connectionListeners.indexOf(cb);
    if (idx !== -1) connectionListeners.splice(idx, 1);
  };
}

export function wsSend(obj) {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(obj));
  }
  if (obj && obj.type === 'cleanup') {
    overlayActive = false;
  }
}

/**
 * Send an unstuck message to kill the game process.
 * Shows a Toast.promise tracking the result.
 */
export function wsSendUnstuck() {
  const promise = new Promise((resolve, reject) => {
    const originalOnMessage = ws.onmessage;
    const timeout = setTimeout(() => {
      ws.onmessage = originalOnMessage;
      reject(new Error('Timeout'));
    }, 10000);

    ws.onmessage = (e) => {
      originalOnMessage(e);
      try {
        const msg = JSON.parse(e.data);
        if (msg.type === 'status') {
          clearTimeout(timeout);
          ws.onmessage = originalOnMessage;
          if (msg.status === 'ready') {
            resolve();
          } else if (msg.status === 'error') {
            reject(new Error(msg.message));
          }
        }
      } catch {}
    };
  });

  promise.then(() => { overlayActive = false; });

  toastPromise(promise, {
    loading: t('toast.unstuck.loading'),
    success: t('toast.unstuck.success'),
    error: t('toast.unstuck.error'),
  });

  wsSend({ type: 'unstuck' });
}

/**
 * Send an apply message and show a single Toast.promise tracking the result.
 * If an apply is already in-flight, skip (let the existing one finish).
 */
export function wsSendApply(obj) {
  if (applyResolve) {
    return;
  }

  lastApplyPayload = {
    championId: obj.championId, skinId: obj.skinId, baseSkinId: obj.baseSkinId,
    championName: obj.championName || null, skinName: obj.skinName || null, chromaName: obj.chromaName || null,
  };

  const promise = new Promise((resolve, reject) => {
    applyResolve = resolve;
    applyReject = reject;
  });

  promise.then(() => { overlayActive = true; });

  toastPromise(promise, {
    loading: t('toast.apply.loading'),
    success: t('toast.apply.success'),
    error: t('toast.apply.error'),
  });

  wsSend(obj);
}
