import { getMyChampionId, loadChampionSkins } from './api';
import { readCurrentSkin, findSkinByName } from './skin';
import { wsSend } from './websocket';
import { setAppliedSkinName, getAppliedSkinName } from './chroma';
import { setButtonState } from './ui';

// Auto-apply configuration
const AUTO_APPLY_DELAY_MS = 10000; // 10 seconds after champion stabilizes
const LOG_PREFIX = '[ame:auto]';

// State
let autoApplyTimer = null;
let lastChampionChangeTime = null;
let lastLoggedChampionId = null;
let autoApplyTriggered = false;
let lastSessionSnapshot = null;

/**
 * Log session data for debugging
 */
function logSessionData(session, label = 'Session') {
  const timer = session.timer || {};
  const myTeam = session.myTeam || [];
  const localCellId = session.localPlayerCellId;

  const teamInfo = myTeam.map(p => ({
    cellId: p.cellId,
    isMe: p.cellId === localCellId,
    championId: p.championId || 0,
    championPickIntent: p.championPickIntent || 0,
    spell1Id: p.spell1Id,
    spell2Id: p.spell2Id,
  }));

  const allHaveChampion = myTeam.every(p => p.championId > 0);
  const me = myTeam.find(p => p.cellId === localCellId);

  console.log(`${LOG_PREFIX} === ${label} ===`);
  console.log(`${LOG_PREFIX} Phase: ${session.gamePhase || session.phase || 'unknown'}`);
  console.log(`${LOG_PREFIX} Timer:`, {
    adjustedTimeLeftInPhase: timer.adjustedTimeLeftInPhase,
    totalTimeInPhase: timer.totalTimeInPhase,
    phase: timer.phase,
    isInfinite: timer.isInfinite,
  });
  console.log(`${LOG_PREFIX} My team (${myTeam.length} players):`, teamInfo);
  console.log(`${LOG_PREFIX} All have champion: ${allHaveChampion}`);
  console.log(`${LOG_PREFIX} My champion ID: ${me?.championId || 'none'}`);
  console.log(`${LOG_PREFIX} ==================`);
}

/**
 * Check if all teammates have selected a champion
 */
function allTeammatesHaveChampion(session) {
  const myTeam = session.myTeam || [];
  if (myTeam.length === 0) return false;
  return myTeam.every(p => p.championId > 0);
}

/**
 * Get my champion ID from session
 */
function getMyChampFromSession(session) {
  const me = session.myTeam?.find(p => p.cellId === session.localPlayerCellId);
  return me?.championId || null;
}

/**
 * Reset auto-apply state (called when entering champ select)
 */
export function resetAutoApply() {
  console.log(`${LOG_PREFIX} Resetting auto-apply state`);
  clearAutoApplyTimer();
  lastChampionChangeTime = null;
  lastLoggedChampionId = null;
  autoApplyTriggered = false;
  lastSessionSnapshot = null;
}

/**
 * Clear the auto-apply timer
 */
function clearAutoApplyTimer() {
  if (autoApplyTimer) {
    clearTimeout(autoApplyTimer);
    autoApplyTimer = null;
    console.log(`${LOG_PREFIX} Auto-apply timer cleared`);
  }
}

/**
 * Trigger the auto-apply
 */
async function triggerAutoApply() {
  if (autoApplyTriggered) {
    console.log(`${LOG_PREFIX} Auto-apply already triggered, skipping`);
    return;
  }

  // Check if user has already manually applied
  const appliedSkinName = getAppliedSkinName();
  if (appliedSkinName) {
    console.log(`${LOG_PREFIX} Skin already applied: ${appliedSkinName}, skipping auto-apply`);
    return;
  }

  console.log(`${LOG_PREFIX} *** TRIGGERING AUTO-APPLY ***`);

  const skinName = readCurrentSkin();
  if (!skinName) {
    console.log(`${LOG_PREFIX} No skin selected, cannot auto-apply`);
    return;
  }

  const championId = await getMyChampionId();
  if (!championId) {
    console.log(`${LOG_PREFIX} No champion selected, cannot auto-apply`);
    return;
  }

  const skins = await loadChampionSkins(championId);
  if (!skins) {
    console.log(`${LOG_PREFIX} Could not load skins, cannot auto-apply`);
    return;
  }

  const skin = findSkinByName(skins, skinName);
  if (!skin) {
    console.log(`${LOG_PREFIX} Skin not found: ${skinName}, cannot auto-apply`);
    return;
  }

  console.log(`${LOG_PREFIX} Auto-applying: ${skinName} | Skin ID: ${skin.id} | Champion ID: ${championId}`);
  wsSend({ type: 'apply', championId, skinId: skin.id });
  setAppliedSkinName(skinName);
  setButtonState('Applied', true);
  autoApplyTriggered = true;

  console.log(`${LOG_PREFIX} Auto-apply completed successfully!`);
}

/**
 * Schedule auto-apply after delay
 */
function scheduleAutoApply() {
  clearAutoApplyTimer();

  console.log(`${LOG_PREFIX} Scheduling auto-apply in ${AUTO_APPLY_DELAY_MS / 1000} seconds...`);

  autoApplyTimer = setTimeout(() => {
    console.log(`${LOG_PREFIX} Auto-apply timer fired!`);
    triggerAutoApply();
  }, AUTO_APPLY_DELAY_MS);
}

/**
 * Main handler for session updates - call this from the session observer
 */
export function handleSessionUpdate(session) {
  if (!session || !session.myTeam) {
    console.log(`${LOG_PREFIX} Invalid session data`);
    return;
  }

  lastSessionSnapshot = session;

  const myChampId = getMyChampFromSession(session);
  const allHaveChamp = allTeammatesHaveChampion(session);

  // Log when champion changes
  if (myChampId !== lastLoggedChampionId) {
    console.log(`${LOG_PREFIX} Champion changed: ${lastLoggedChampionId || 'none'} -> ${myChampId || 'none'}`);
    logSessionData(session, 'Champion Changed');
    lastLoggedChampionId = myChampId;
    lastChampionChangeTime = Date.now();

    // Reset auto-apply when champion changes (user might be rerolling in ARAM)
    if (!autoApplyTriggered) {
      clearAutoApplyTimer();
    }
  }

  // Log periodically for debugging (every 5 seconds based on session updates)
  const now = Date.now();
  const timeSinceChange = lastChampionChangeTime ? now - lastChampionChangeTime : null;

  // Check if we should start the auto-apply timer
  if (!autoApplyTriggered && myChampId && allHaveChamp) {
    if (!autoApplyTimer) {
      console.log(`${LOG_PREFIX} Conditions met: all teammates have champions, starting timer`);
      logSessionData(session, 'Auto-apply conditions met');
      scheduleAutoApply();
    }
  } else if (autoApplyTimer && (!myChampId || !allHaveChamp)) {
    // Conditions no longer met, clear timer
    console.log(`${LOG_PREFIX} Conditions no longer met, clearing timer`);
    clearAutoApplyTimer();
  }

  // Debug logging
  if (timeSinceChange !== null && timeSinceChange < AUTO_APPLY_DELAY_MS) {
    const remaining = Math.ceil((AUTO_APPLY_DELAY_MS - timeSinceChange) / 1000);
    if (remaining % 3 === 0) { // Log every ~3 seconds
      console.log(`${LOG_PREFIX} Time since last champion change: ${Math.floor(timeSinceChange / 1000)}s, auto-apply in: ${remaining}s`);
    }
  }
}

/**
 * Fetch and log the timer data separately (for detailed debugging)
 */
export async function fetchAndLogTimer() {
  try {
    const res = await fetch('/lol-champ-select/v1/session/timer');
    if (!res.ok) {
      console.log(`${LOG_PREFIX} Timer fetch failed: ${res.status}`);
      return null;
    }
    const timer = await res.json();
    console.log(`${LOG_PREFIX} Timer data:`, timer);
    return timer;
  } catch (e) {
    console.log(`${LOG_PREFIX} Timer fetch error:`, e.message);
    return null;
  }
}

/**
 * Fetch gameflow session for queue info (ARAM detection)
 */
export async function fetchAndLogGameflow() {
  try {
    const res = await fetch('/lol-gameflow/v1/session');
    if (!res.ok) {
      console.log(`${LOG_PREFIX} Gameflow fetch failed: ${res.status}`);
      return null;
    }
    const gameflow = await res.json();
    console.log(`${LOG_PREFIX} Gameflow data:`, {
      gameMode: gameflow.gameData?.queue?.gameMode,
      queueId: gameflow.gameData?.queue?.id,
      mapId: gameflow.map?.id,
      phase: gameflow.phase,
    });
    return gameflow;
  } catch (e) {
    console.log(`${LOG_PREFIX} Gameflow fetch error:`, e.message);
    return null;
  }
}
