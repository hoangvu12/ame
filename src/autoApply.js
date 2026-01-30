import { loadChampionSkins, getMyChampionId } from './api';
import { readCurrentSkin, findSkinByName, isDefaultSkin } from './skin';
import { wsSend, wsSendApply } from './websocket';
import { setAppliedSkinName, getAppliedSkinName } from './chroma';
import { setButtonState } from './ui';
import { PREFETCH_DEBOUNCE_MS } from './constants';

const AUTO_APPLY_STABLE_MS = 10000; // 10 seconds of stability
const LOG_PREFIX = '[ame:auto]';

// State
let lastTrackedSkin = null;
let lastTrackedChampion = null;
let stableSince = null;
let autoApplyTriggered = false;
let epoch = 0; // incremented on reset; async flows bail if epoch changes

// Chroma selection state
let selectedChromaId = null;
let selectedChromaBaseSkinId = null;

// Applied chroma tracking (to detect chroma-only changes for re-arm)
let appliedChromaId = null;

// Prefetch debounce
let prefetchTimer = null;

/**
 * Set the selected chroma (called from chroma.js when user clicks a chroma).
 * Resets stability timer so auto-apply waits another 10s.
 */
export function setSelectedChroma(chromaId, baseSkinId) {
  console.log(`${LOG_PREFIX} Chroma selected: ${chromaId} (base: ${baseSkinId})`);
  selectedChromaId = chromaId;
  selectedChromaBaseSkinId = baseSkinId;
  stableSince = Date.now();
  autoApplyTriggered = false;
}

/**
 * Clear chroma selection (called when base skin changes or on reset).
 */
export function clearSelectedChroma() {
  selectedChromaId = null;
  selectedChromaBaseSkinId = null;
}

/**
 * Get the currently selected chroma, or null if none.
 */
export function getSelectedChroma() {
  if (!selectedChromaId) return null;
  return { id: selectedChromaId, baseSkinId: selectedChromaBaseSkinId };
}

/**
 * Cancel any pending prefetch.
 */
function cancelPrefetch() {
  if (prefetchTimer) {
    clearTimeout(prefetchTimer);
    prefetchTimer = null;
  }
}

/**
 * Start a debounced prefetch for the given skin.
 * Cancels any previous pending prefetch.
 */
function debouncePrefetch(championId, skinName) {
  cancelPrefetch();
  const startEpoch = epoch;
  prefetchTimer = setTimeout(async () => {
    prefetchTimer = null;
    if (epoch !== startEpoch) return;

    const skins = await loadChampionSkins(championId);
    if (!skins || epoch !== startEpoch) return;

    const skin = findSkinByName(skins, skinName);
    if (!skin || isDefaultSkin(skin)) return;

    // Only send prefetch if still on this skin
    if (lastTrackedSkin === skinName && lastTrackedChampion === championId) {
      console.log(`${LOG_PREFIX} Prefetching skin: ${skinName} (id: ${skin.id})`);
      wsSend({ type: 'prefetch', championId, skinId: skin.id });
    }
  }, PREFETCH_DEBOUNCE_MS);
}

/**
 * Send an immediate prefetch for a chroma (no debounce, since it's an explicit click).
 */
export function prefetchChroma(championId, chromaId, baseSkinId) {
  console.log(`${LOG_PREFIX} Prefetching chroma: ${chromaId} (base: ${baseSkinId})`);
  wsSend({ type: 'prefetch', championId, skinId: chromaId, baseSkinId });
}

/**
 * Force an immediate apply if nothing has been applied yet.
 * Called when leaving champ select (game about to start) as a last resort.
 */
export async function forceApplyIfNeeded() {
  // Already applied — nothing to do
  if (getAppliedSkinName()) {
    console.log(`${LOG_PREFIX} forceApply: already applied, skipping`);
    return;
  }

  const skinName = lastTrackedSkin || readCurrentSkin();
  const championId = lastTrackedChampion || await getMyChampionId();
  if (!skinName || !championId) {
    console.log(`${LOG_PREFIX} forceApply: no skin or champion available, skipping`);
    return;
  }

  const skins = await loadChampionSkins(championId);
  if (!skins) {
    console.log(`${LOG_PREFIX} forceApply: could not load skins, skipping`);
    return;
  }

  const skin = findSkinByName(skins, skinName);
  if (!skin) {
    console.log(`${LOG_PREFIX} forceApply: skin "${skinName}" not found, skipping`);
    return;
  }

  if (isDefaultSkin(skin)) {
    console.log(`${LOG_PREFIX} forceApply: default skin "${skinName}", skipping`);
    return;
  }

  if (selectedChromaId) {
    console.log(`${LOG_PREFIX} *** FORCE APPLY (chroma) *** ${selectedChromaId} (base: ${selectedChromaBaseSkinId}) | Champion: ${championId}`);
    wsSendApply({ type: 'apply', championId, skinId: selectedChromaId, baseSkinId: selectedChromaBaseSkinId });
  } else {
    console.log(`${LOG_PREFIX} *** FORCE APPLY *** ${skinName} | Skin ID: ${skin.id} | Champion: ${championId}`);
    wsSendApply({ type: 'apply', championId, skinId: skin.id });
  }

  setAppliedSkinName(skinName);
  appliedChromaId = selectedChromaId || null;
  setButtonState('Applied', true);
}

/**
 * Reset auto-apply state (called when entering/leaving champ select)
 */
export function resetAutoApply() {
  console.log(`${LOG_PREFIX} Resetting auto-apply state`);
  lastTrackedSkin = null;
  lastTrackedChampion = null;
  stableSince = null;
  autoApplyTriggered = false;
  epoch++;
  clearSelectedChroma();
  appliedChromaId = null;
  cancelPrefetch();
}

/**
 * Check if skin and champion have been stable long enough to auto-apply.
 * Called from pollUI every 300ms.
 *
 * Tracking (lastTrackedSkin/lastTrackedChampion) always runs so that
 * in-flight triggerAutoApply() can detect mid-await changes.
 * Only the trigger decision is gated by autoApplyTriggered.
 */
export function checkAutoApply(championId) {
  const skinName = readCurrentSkin();

  // Build a combined key that includes chroma selection
  const currentKey = `${skinName || ''}|${selectedChromaId || ''}`;
  const appliedKey = `${getAppliedSkinName() || ''}|${appliedChromaId || ''}`;

  // Re-arm: if user changed skin or chroma since last apply, clear and re-arm
  const applied = getAppliedSkinName();
  if (applied) {
    if (currentKey !== appliedKey) {
      console.log(`${LOG_PREFIX} Selection changed from applied "${appliedKey}" to "${currentKey}", re-arming auto-apply`);
      setAppliedSkinName(null);
      appliedChromaId = null;
      setButtonState('Apply Skin', false);
      autoApplyTriggered = false;
      // Fall through to start tracking the new selection
    } else {
      return;
    }
  }

  // Need both values present — but don't clear existing tracking,
  // the DOM may temporarily show nothing (e.g. game loading transition)
  if (!skinName || !championId) {
    return;
  }

  // Detect changes in skin name, champion, or chroma
  const skinChanged = skinName !== lastTrackedSkin;
  const champChanged = championId !== lastTrackedChampion;

  if (skinChanged || champChanged) {
    if (skinChanged) {
      // Different base skin means different chromas — clear selection
      clearSelectedChroma();
    }

    console.log(`${LOG_PREFIX} Change detected: champ ${lastTrackedChampion} -> ${championId}, skin "${lastTrackedSkin}" -> "${skinName}"`);
    lastTrackedSkin = skinName;
    lastTrackedChampion = championId;
    stableSince = Date.now();
    autoApplyTriggered = false;

    // Start debounced prefetch for the new skin
    debouncePrefetch(championId, skinName);

    return;
  }

  // Only trigger if not already in-flight
  if (autoApplyTriggered) return;

  // Both values stable — check if long enough
  if (stableSince && Date.now() - stableSince >= AUTO_APPLY_STABLE_MS) {
    triggerAutoApply();
  }
}

/**
 * Trigger the auto-apply
 */
async function triggerAutoApply() {
  if (autoApplyTriggered) return;
  autoApplyTriggered = true;

  const startEpoch = epoch;
  const startSkin = lastTrackedSkin;
  const startChampion = lastTrackedChampion;
  const startChromaId = selectedChromaId;
  const startChromaBaseSkinId = selectedChromaBaseSkinId;

  // Re-check manual apply (may have happened between poll cycles)
  if (getAppliedSkinName()) {
    console.log(`${LOG_PREFIX} Skin already applied manually, skipping auto-apply`);
    autoApplyTriggered = false;
    return;
  }

  console.log(`${LOG_PREFIX} *** TRIGGERING AUTO-APPLY ***`);

  const skinName = startSkin;
  const championId = startChampion;
  if (!championId) {
    console.log(`${LOG_PREFIX} No champion, cannot auto-apply`);
    autoApplyTriggered = false;
    stableSince = Date.now();
    return;
  }

  const skins = await loadChampionSkins(championId);

  // Bail if phase changed (left champ select) during await
  if (epoch !== startEpoch) {
    console.log(`${LOG_PREFIX} Epoch changed during skin load, aborting auto-apply`);
    return;
  }

  if (!skins) {
    console.log(`${LOG_PREFIX} Could not load skins, cannot auto-apply`);
    autoApplyTriggered = false;
    stableSince = Date.now();
    return;
  }

  // Re-validate: champ/skin/chroma may have changed during the await
  if (lastTrackedSkin !== startSkin || lastTrackedChampion !== startChampion || selectedChromaId !== startChromaId) {
    console.log(`${LOG_PREFIX} Selection changed during skin load, aborting auto-apply`);
    autoApplyTriggered = false;
    stableSince = Date.now();
    return;
  }

  const skin = findSkinByName(skins, skinName);
  if (!skin) {
    console.log(`${LOG_PREFIX} Skin not found: ${skinName}, cannot auto-apply`);
    autoApplyTriggered = false;
    stableSince = Date.now();
    return;
  }

  if (isDefaultSkin(skin)) {
    console.log(`${LOG_PREFIX} Default skin "${skinName}", skipping auto-apply`);
    autoApplyTriggered = false;
    return;
  }

  // Re-check: user may have manually applied during the awaits above
  if (getAppliedSkinName()) {
    console.log(`${LOG_PREFIX} Skin manually applied during auto-apply, skipping`);
    autoApplyTriggered = false;
    return;
  }

  // Final DOM check: abort if user changed skin since trigger (poll may not have run yet)
  if (readCurrentSkin() !== startSkin) {
    console.log(`${LOG_PREFIX} Skin changed since trigger, aborting auto-apply`);
    autoApplyTriggered = false;
    stableSince = Date.now();
    return;
  }

  // Apply with chroma if selected, otherwise base skin
  if (startChromaId) {
    console.log(`${LOG_PREFIX} Auto-applying chroma: ${startChromaId} (base: ${startChromaBaseSkinId}) | Champion ID: ${championId}`);
    wsSendApply({ type: 'apply', championId, skinId: startChromaId, baseSkinId: startChromaBaseSkinId });
  } else {
    console.log(`${LOG_PREFIX} Auto-applying: ${skinName} | Skin ID: ${skin.id} | Champion ID: ${championId}`);
    wsSendApply({ type: 'apply', championId, skinId: skin.id });
  }

  // Track what was applied (base skin name for DOM matching + chroma ID)
  setAppliedSkinName(skinName);
  appliedChromaId = startChromaId || null;
  setButtonState('Applied', true);

  console.log(`${LOG_PREFIX} Auto-apply completed successfully!`);
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
