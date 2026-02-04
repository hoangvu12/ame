import { wsSend, onRoomPartyUpdate, onSetting } from './websocket';
import { fetchJson } from './api';
import { el } from './dom';
import { ROOM_PARTY_INDICATOR_CLASS } from './constants';
import { retriggerPrefetch } from './autoApply';
import { t } from './i18n';

const RETRIGGER_DEBOUNCE_MS = 5000;

let enabled = false;
let joined = false;
let joining = false;
let currentTeammates = [];
let unsubUpdate = null;
let retriggerDebounceTimer = null;

function teammateSkinKey(teammates) {
  return teammates.map(t => t.skinInfo?.skinId || '').sort().join(',');
}

export function loadRoomPartySetting() {
  onSetting('roomParty', (v) => {
    enabled = v;
    if (!v && joined) {
      leaveRoom();
    } else if (v && !joined) {
      joinRoom();
    }
  });
}

export async function joinRoom(existingSession) {
  if (!enabled || joined || joining) {
    console.log('[ame] joinRoom skipped:', !enabled ? 'not enabled' : joined ? 'already joined' : 'join in progress');
    return;
  }
  joining = true;

  try {
    const summoner = await fetchJson('/lol-summoner/v1/current-summoner');
    if (!summoner?.puuid) { console.log('[ame] joinRoom: no puuid'); return; }

    const session = existingSession || await fetchJson('/lol-champ-select/v1/session');
    if (!session?.myTeam) { console.log('[ame] joinRoom: no myTeam'); return; }

    const roomKey = session.chatDetails?.multiUserChatId;
    if (!roomKey) { console.log('[ame] joinRoom: no multiUserChatId'); return; }

    const teamPuuids = session.myTeam
      .map(p => p.puuid)
      .filter(p => p && p !== '' && p !== summoner.puuid);

    console.log('[ame] joinRoom: joining room', roomKey, 'with', teamPuuids.length, 'teammates');

    wsSend({
      type: 'roomPartyJoin',
      roomKey,
      puuid: summoner.puuid,
      teamPuuids,
    });

    joined = true;

    unsubUpdate = onRoomPartyUpdate((teammates) => {
      const oldKey = teammateSkinKey(currentTeammates);
      const newKey = teammateSkinKey(teammates);
      console.log(`[ame] roomPartyUpdate: ${teammates.length} teammates, skinKey: "${oldKey}" -> "${newKey}"`);
      currentTeammates = teammates;
      renderTeammateIndicators();
      if (newKey !== oldKey) {
        // Only retrigger if there are new or changed skins — not when teammates
        // disappear (transient polling gaps should not remove already-applied skins).
        const oldIds = new Set(oldKey.split(',').filter(Boolean));
        const hasNewSkins = newKey.split(',').filter(Boolean).some(id => !oldIds.has(id));
        if (hasNewSkins) {
          if (retriggerDebounceTimer) clearTimeout(retriggerDebounceTimer);
          console.log(`[ame] roomPartyUpdate: new teammate skins detected, debouncing retrigger`);
          retriggerDebounceTimer = setTimeout(() => {
            retriggerDebounceTimer = null;
            console.log(`[ame] roomPartyUpdate: debounce fired, retriggering prefetch`);
            retriggerPrefetch();
          }, RETRIGGER_DEBOUNCE_MS);
        } else {
          console.log(`[ame] roomPartyUpdate: teammate skins removed, skipping retrigger`);
        }
      }
    });
  } finally {
    joining = false;
  }
}

export function notifySkinChange(championId, skinId, baseSkinId, championName, skinName, chromaName) {
  if (!enabled || !joined) return;
  console.log(`[ame] Room party: notifying skin change: ${skinName} (${skinId})`);
  wsSend({
    type: 'roomPartySkin',
    championId,
    skinId,
    baseSkinId: baseSkinId || '',
    championName: championName || '',
    skinName: skinName || '',
    chromaName: chromaName || '',
  });
}

export function flushPendingRetrigger() {
  if (retriggerDebounceTimer) {
    clearTimeout(retriggerDebounceTimer);
    retriggerDebounceTimer = null;
    console.log(`[ame] flushPendingRetrigger: flushing debounced retrigger`);
    retriggerPrefetch();
  }
}

export function leaveRoom() {
  if (!joined) return;
  if (retriggerDebounceTimer) { clearTimeout(retriggerDebounceTimer); retriggerDebounceTimer = null; }
  console.log('[ame] leaveRoom: leaving room party');
  wsSend({ type: 'roomPartyLeave' });
  joined = false;
  currentTeammates = [];
  if (unsubUpdate) {
    unsubUpdate();
    unsubUpdate = null;
  }
  removeTeammateIndicators();
}

export function resetRoomPartyJoin() {
  joined = false;
}

function removeTeammateIndicators() {
  document.querySelectorAll(`.${ROOM_PARTY_INDICATOR_CLASS}`).forEach(e => e.remove());
}

async function renderTeammateIndicators() {
  removeTeammateIndicators();
  if (currentTeammates.length === 0) return;

  const session = await fetchJson('/lol-champ-select/v1/session');
  if (!session?.myTeam) return;

  const teamOrdered = session.myTeam.slice().sort((a, b) => a.cellId - b.cellId);
  const slots = document.querySelectorAll('.ally-slot');
  if (!slots.length) return;

  for (const tm of currentTeammates) {
    if (!tm.skinInfo?.skinName) continue;

    const label = tm.skinInfo.chromaName
      ? `${tm.skinInfo.skinName} (${tm.skinInfo.chromaName})`
      : tm.skinInfo.skinName;

    const targetIndex = teamOrdered.findIndex(p => p.puuid === tm.puuid);
    if (targetIndex < 0 || targetIndex >= slots.length) continue;

    const slot = slots[targetIndex];
    slot.style.position = 'relative';

    const badge = el('div', {
      class: ROOM_PARTY_INDICATOR_CLASS,
      title: t('room_party.tooltip', { label }),
    }, el('span', null, label));

    slot.appendChild(badge);
  }
}
