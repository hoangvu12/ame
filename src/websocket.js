import { WS_URL, WS_RECONNECT_BASE_MS, WS_RECONNECT_MAX_MS } from './constants';

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

// One-shot callback for autoAccept response
let autoAcceptCallback = null;

// One-shot callback for benchSwap response
let benchSwapCallback = null;

export function wsConnect() {
  if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) return;
  try {
    ws = new WebSocket(WS_URL);
    ws.onopen = () => {
      console.log('[ame] WebSocket connected');
      wsReconnectDelay = WS_RECONNECT_BASE_MS;
      // Query server for current overlay state (survives client restarts)
      ws.send(JSON.stringify({ type: 'query' }));
    };
    ws.onmessage = (e) => {
      try {
        const msg = JSON.parse(e.data);
        if (msg.type === 'state') {
          // Hydrate state from server (on connect/reconnect)
          if (msg.championId && msg.skinId) {
            lastApplyPayload = {
              championId: Number(msg.championId),
              skinId: Number(msg.skinId),
              baseSkinId: msg.baseSkinId ? Number(msg.baseSkinId) : Number(msg.skinId),
            };
          }
          overlayActive = !!msg.overlayActive;
          console.log('[ame] State from server:', overlayActive ? 'active' : 'inactive', lastApplyPayload);
        } else if (msg.type === 'gamePath') {
          if (gamePathCallback) {
            gamePathCallback(msg.path || '');
            gamePathCallback = null;
          }
        } else if (msg.type === 'autoAccept') {
          if (autoAcceptCallback) {
            autoAcceptCallback(!!msg.enabled);
            autoAcceptCallback = null;
          }
        } else if (msg.type === 'benchSwap') {
          if (benchSwapCallback) {
            benchSwapCallback(!!msg.enabled);
            benchSwapCallback = null;
          }
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
            } else if (typeof Toast !== 'undefined') {
              Toast.error(msg.message);
            }
          }
        }
      } catch (err) {
        console.log('[ame] onmessage error:', err);
      }
    };
    ws.onclose = () => wsScheduleReconnect();
    ws.onerror = () => {};
  } catch {
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
export function isOverlayActive() { return overlayActive; }
export function setOverlayActive(v) { overlayActive = v; }
export function onGamePath(cb) { gamePathCallback = cb; }
export function onAutoAccept(cb) { autoAcceptCallback = cb; }
export function onBenchSwap(cb) { benchSwapCallback = cb; }

export function wsSend(obj) {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(obj));
  }
  if (obj && obj.type === 'cleanup') {
    overlayActive = false;
  }
}

/**
 * Send an apply message and show a single Toast.promise tracking the result.
 * If an apply is already in-flight, skip (let the existing one finish).
 */
export function wsSendApply(obj) {
  if (applyResolve) {
    // Already an apply in-flight — don't send another
    return;
  }

  lastApplyPayload = { championId: obj.championId, skinId: obj.skinId, baseSkinId: obj.baseSkinId };

  const promise = new Promise((resolve, reject) => {
    applyResolve = resolve;
    applyReject = reject;
  });

  promise.then(() => { overlayActive = true; });

  if (typeof Toast !== 'undefined') {
    Toast.promise(promise, {
      loading: 'Applying skin...',
      success: 'Skin applied!',
      error: 'Failed to apply skin',
    });
  }

  wsSend(obj);
}
