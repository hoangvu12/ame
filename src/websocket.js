import { WS_URL, WS_RECONNECT_BASE_MS, WS_RECONNECT_MAX_MS } from './constants';

let ws = null;
let wsReconnectDelay = WS_RECONNECT_BASE_MS;
let wsReconnectTimer = null;

// Pending apply promise — resolved/rejected by incoming status messages
let applyResolve = null;
let applyReject = null;

export function wsConnect() {
  if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) return;
  try {
    ws = new WebSocket(WS_URL);
    ws.onopen = () => {
      console.log('[ame] WebSocket connected');
      wsReconnectDelay = WS_RECONNECT_BASE_MS;
    };
    ws.onmessage = (e) => {
      try {
        const msg = JSON.parse(e.data);
        if (msg.type === 'status') {
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

export function wsSend(obj) {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(obj));
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

  const promise = new Promise((resolve, reject) => {
    applyResolve = resolve;
    applyReject = reject;
  });

  if (typeof Toast !== 'undefined') {
    Toast.promise(promise, {
      loading: 'Applying skin...',
      success: 'Skin applied!',
      error: 'Failed to apply skin',
    });
  }

  wsSend(obj);
}
