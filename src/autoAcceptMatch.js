import { AUTO_ACCEPT_DELAY_MS } from './constants';
import { onSetting } from './websocket';
import { fetchJson } from './api';

let enabled = false;
let pendingTimer = null;

export function loadAutoAcceptSetting() {
  onSetting('autoAccept', (v) => { enabled = v; });
}

export function handleReadyCheck() {
  if (!enabled) return;
  if (pendingTimer) return;

  pendingTimer = setTimeout(async () => {
    pendingTimer = null;
    const data = await fetchJson('/lol-matchmaking/v1/ready-check');
    if (!data) return;
    if (data.playerResponse === 'Declined' || data.playerResponse === 'Accepted') return;
    fetch('/lol-matchmaking/v1/ready-check/accept', { method: 'POST' }).catch(() => {});
  }, AUTO_ACCEPT_DELAY_MS);
}

export function cancelPendingAccept() {
  if (pendingTimer) {
    clearTimeout(pendingTimer);
    pendingTimer = null;
  }
}
