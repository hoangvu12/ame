import { AUTO_ACCEPT_DELAY_MS } from './constants';
import { wsSend, onAutoAccept } from './websocket';

let enabled = false;
let pendingTimer = null;

export function setAutoAcceptEnabled(v) {
  enabled = v;
}

export function loadAutoAcceptSetting() {
  onAutoAccept((v) => { enabled = v; });
  wsSend({ type: 'getAutoAccept' });
}

export function handleReadyCheck() {
  if (!enabled) return;
  if (pendingTimer) return;

  console.log('[ame] ReadyCheck detected, will auto-accept in', AUTO_ACCEPT_DELAY_MS, 'ms');

  pendingTimer = setTimeout(async () => {
    pendingTimer = null;
    try {
      // Check current ready-check state to respect manual decline
      const res = await fetch('/lol-matchmaking/v1/ready-check');
      if (!res.ok) {
        console.log('[ame] Ready check endpoint returned', res.status);
        return;
      }
      const data = await res.json();
      const response = data.playerResponse;

      if (response === 'Declined') {
        console.log('[ame] User declined — skipping auto-accept');
        return;
      }
      if (response === 'Accepted') {
        console.log('[ame] Already accepted');
        return;
      }

      // playerResponse is "None" — accept
      await fetch('/lol-matchmaking/v1/ready-check/accept', { method: 'POST' });
      console.log('[ame] Auto-accepted match');
    } catch (err) {
      console.log('[ame] Auto-accept error:', err);
    }
  }, AUTO_ACCEPT_DELAY_MS);
}

export function cancelPendingAccept() {
  if (pendingTimer) {
    clearTimeout(pendingTimer);
    pendingTimer = null;
  }
}
