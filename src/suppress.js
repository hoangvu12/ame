import { createLogger } from './logger';

const logger = createLogger('suppress');

let suppressFlag = false;
let suppressTimeout = null;

/**
 * Arm the one-shot suppression flag. The next `skin-selector-info` WebSocket
 * event will be silently dropped before the champ-select UI processes it.
 * Auto-expires after 3 s if the event never arrives.
 */
export function suppressNextSkinEvent() {
  suppressFlag = true;
  if (suppressTimeout) clearTimeout(suppressTimeout);
  suppressTimeout = setTimeout(() => {
    if (suppressFlag) {
      logger.log('suppress timeout expired, disarming');
      suppressFlag = false;
    }
    suppressTimeout = null;
  }, 3000);
  logger.log('armed skin-selector-info suppression');
}

/**
 * Clear the suppression flag (e.g. on champion change or champ-select exit).
 */
export function disarmSkinSuppression() {
  if (suppressFlag) logger.log('disarming suppression');
  suppressFlag = false;
  if (suppressTimeout) { clearTimeout(suppressTimeout); suppressTimeout = null; }
}

/**
 * Hook into the champ-select internal WebSocket so we can drop
 * `skin-selector-info` events before the UI reacts to them.
 *
 * Must be called once during init(). Falls back gracefully if the
 * RCP plugin API is unavailable.
 */
export function initSkinSuppression() {
  if (!window.rcp?.postInit) {
    logger.log('window.rcp.postInit not available, suppression disabled');
    return;
  }

  window.rcp.postInit('rcp-fe-lol-champ-select', (api) => {
    try {
      const ws = api?.champSelectBinding?.socket?._websocket;
      if (!ws) {
        logger.log('could not locate champ-select WebSocket');
        return;
      }

      const origOnMessage = ws.onmessage;
      ws.onmessage = function (event) {
        if (suppressFlag) {
          try {
            const payload = JSON.parse(event.data);
            if (
              Array.isArray(payload) &&
              payload[1] === 'OnJsonApiEvent' &&
              payload[2]?.uri === '/lol-champ-select/v1/skin-selector-info'
            ) {
              suppressFlag = false;
              if (suppressTimeout) { clearTimeout(suppressTimeout); suppressTimeout = null; }
              logger.log('suppressed skin-selector-info event');
              return; // drop the event
            }
          } catch {
            // not JSON — pass through
          }
        }
        return origOnMessage.call(this, event);
      };

      logger.log('hooked champ-select WebSocket onmessage');
    } catch (err) {
      logger.log('failed to hook WebSocket:', err);
    }
  });
}
