import { onChatStatus } from './websocket';
import { fetchJson } from './api';

let currentAvailability = '';
let reapplyTimer = null;
let applying = false;

function getGameStatus(availability) {
  return availability === 'dnd' ? 'inGame' : 'outOfGame';
}

async function applyToLCU() {
  if (!currentAvailability || applying) return false;
  applying = true;
  try {
    const res = await fetch('/lol-chat/v1/me', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        availability: currentAvailability,
        lol: { gameStatus: getGameStatus(currentAvailability) },
      }),
    });
    return res.ok;
  } catch {
    return false;
  } finally {
    applying = false;
  }
}

function startReapply() {
  stopReapply();
  reapplyTimer = setInterval(applyToLCU, 5000);
}

function stopReapply() {
  if (reapplyTimer) {
    clearInterval(reapplyTimer);
    reapplyTimer = null;
  }
}

/**
 * Apply chat status directly. Called from settings UI for immediate effect.
 * Returns the verified availability string on success, null on failure.
 */
export async function applyChatStatus(availability) {
  currentAvailability = availability;
  if (!availability) {
    stopReapply();
    return null;
  }
  const ok = await applyToLCU();
  if (!ok) return null;
  startReapply();
  const me = await fetchJson('/lol-chat/v1/me');
  return me?.availability || null;
}

export function initChatStatus(context) {
  onChatStatus(({ availability }) => {
    currentAvailability = availability;
    if (availability) {
      applyToLCU();
      startReapply();
    } else {
      stopReapply();
    }
  });

  context.socket.observe('/lol-chat/v1/me', (event) => {
    if (!currentAvailability || applying) return;
    const data = event.data;
    if (data && data.availability !== currentAvailability) {
      applyToLCU();
    }
  });
}
