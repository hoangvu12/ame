import { loadChampionSkins, getMyChampionId, getChampionName, fetchJson, forceDefaultSkin } from './api';
import { readCurrentSkin, findSkinByName, isDefaultSkin, getSkinKeyFromItem } from './skin';
import { wsSend, wsSendApply, isApplyInFlight, isOverlayActive, hasEnabledCustomMods } from './websocket';
import {
  getAppliedSkinName, setAppliedSkinName,
  getSelectedChroma, setSelectedChroma, clearSelectedChroma,
  getAppliedChromaId, setAppliedChromaId,
  setSkinForced,
  getSkinForced,
} from './state';
import { setButtonState } from './ui';
import { PREFETCH_DEBOUNCE_MS } from './constants';
import { notifySkinChange } from './roomParty';
import { t } from './i18n';
import { createLogger } from './logger';

const logger = createLogger('auto');
const AUTO_APPLY_STABLE_MS = 10000;

// Tracking state (local to this module — not shared)
let lastTrackedSkin = null;
let lastTrackedChampion = null;
let stableSince = null;
let autoApplyTriggered = false;
let epoch = 0;
let prefetchTimer = null;
let lastPrefetchPayload = null;
let retriggerTimer = null;
let retriggerRetries = 0;
let champSelectActive = false;
let pendingClickBack = null;
const MAX_RETRIGGER_RETRIES = 3;

/**
 * Notify auto-apply that a chroma was selected — resets stability timer.
 */
export function onChromaSelected(chromaId, baseSkinId, chromaName = null, baseSkinName = null) {
  logger.log(` Chroma selected: ${chromaId} (base: ${baseSkinId})`);
  setSelectedChroma(chromaId, baseSkinId, chromaName, baseSkinName);
  stableSince = Date.now();
  autoApplyTriggered = false;
}

/**
 * Send an immediate prefetch for a chroma (explicit click, no debounce).
 */
export function prefetchChroma(championId, chromaId, baseSkinId, championName = null, skinName = null, chromaName = null) {
  const payload = { type: 'prefetch', championId, skinId: chromaId, baseSkinId, championName, skinName, chromaName };
  lastPrefetchPayload = payload;
  wsSend(payload);
}

/**
 * Try to click the carousel item back to the unowned skin after forceDefaultSkin.
 * Called from the poll cycle so it runs reliably in the DOM context.
 */
export function processClickBack() {
  if (pendingClickBack === null) return;
  const skinNum = pendingClickBack;
  const items = document.querySelectorAll('.skin-selection-item');
  for (const item of items) {
    const key = getSkinKeyFromItem(item);
    if (key === skinNum) {
      logger.log(` clickBack: found skinNum=${skinNum}, dispatching events`);
      const thumb = item.querySelector('.skin-selection-thumbnail') || item;
      thumb.dispatchEvent(new PointerEvent('pointerdown', { bubbles: true, cancelable: true }));
      thumb.dispatchEvent(new PointerEvent('pointerup', { bubbles: true, cancelable: true }));
      thumb.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true }));
      pendingClickBack = null;
      return;
    }
  }
  // Item not in DOM yet — will retry on next poll cycle
}

// --- Prefetch debounce ---

function cancelPrefetch() {
  if (prefetchTimer) {
    clearTimeout(prefetchTimer);
    prefetchTimer = null;
  }
}

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

    if (lastTrackedSkin === skinName && lastTrackedChampion === championId) {
      const champName = await getChampionName(championId);
      const payload = { type: 'prefetch', championId, skinId: skin.id, championName: champName, skinName: skin.name };
      lastPrefetchPayload = payload;
      logger.log(` prefetch: sending for ${skin.name} (${skin.id})`);
      wsSend(payload);
      notifySkinChange(championId, skin.id, '', champName, skin.name, '');

      // Force base skin now so the session is on base before champ select ends.
      // skinForced prevents the poll from cleaning up the overlay.
      logger.log(` prefetch: skinForced=${getSkinForced()}, will force=${!getSkinForced()}`);
      if (!getSkinForced()) {
        const skinNum = skin.id % 1000;
        const forced = await forceDefaultSkin(championId);
        logger.log(` prefetch: forceDefaultSkin result: ${forced}`);
        if (forced) {
          setSkinForced(true);
          // Click the unowned skin's carousel item to scroll back,
          // so the UI doesn't jump to the base skin confusingly.
          pendingClickBack = skinNum;
          logger.log(` prefetch: set pendingClickBack=${skinNum}`);
        }
      }
    }
  }, PREFETCH_DEBOUNCE_MS);
}

// --- Force apply (last resort before game starts) ---

export async function forceApplyIfNeeded() {
  if (getAppliedSkinName()) { logger.log(` forceApply: skipped (already applied: ${getAppliedSkinName()})`); return; }
  if (isApplyInFlight()) { logger.log(` forceApply: skipped (apply in-flight)`); return; }
  if (isOverlayActive()) { logger.log(` forceApply: skipped (overlay active)`); return; }

  // Use saved prefetch payload — after forceDefaultSkin the DOM shows the base
  // skin so lastTrackedSkin / readCurrentSkin() would resolve to default.
  // Capture in a local var because lockRetrigger() can null lastPrefetchPayload
  // if the InProgress phase fires during the forceDefaultSkin await.
  const savedPayload = lastPrefetchPayload;
  if (savedPayload) {
    const championId = savedPayload.championId || lastTrackedChampion;
    if (championId) {
      const forced = await forceDefaultSkin(championId);
      logger.log(` forceApply: forceDefaultSkin result: ${forced}`);
    }
    const chroma = getSelectedChroma();
    logger.log(` forceApply: applying from payload ${savedPayload.skinName} (${savedPayload.skinId}${chroma ? ', chroma: ' + chroma.id : ''})`);
    if (chroma) {
      wsSendApply({
        type: 'apply', championId: savedPayload.championId,
        skinId: chroma.id, baseSkinId: chroma.baseSkinId,
        championName: savedPayload.championName,
        skinName: chroma.baseSkinName || savedPayload.skinName,
        chromaName: chroma.chromaName,
      });
    } else {
      wsSendApply({ ...savedPayload, type: 'apply' });
    }
    setAppliedSkinName(savedPayload.skinName);
    setAppliedChromaId(chroma?.id || null);
    setButtonState(t('ui.applied'), true);
    return;
  }

  // Fallback: resolve from DOM / tracking state (no prefetch happened)
  const skinName = lastTrackedSkin || readCurrentSkin();
  const championId = lastTrackedChampion || await getMyChampionId();
  if (!skinName || !championId) {
    // No skin data — still apply custom mods if enabled
    await applyCustomModsOnly();
    return;
  }

  const skins = await loadChampionSkins(championId);
  if (!skins) {
    await applyCustomModsOnly();
    return;
  }

  const skin = findSkinByName(skins, skinName);
  if (!skin || isDefaultSkin(skin)) {
    // Default skin selected — still apply custom mods if enabled
    await applyCustomModsOnly();
    return;
  }

  const forced = await forceDefaultSkin(championId);
  logger.log(` forceApply: forceDefaultSkin (fallback) result: ${forced}`);

  const champName = await getChampionName(championId);
  const chroma = getSelectedChroma();
  logger.log(` forceApply: applying ${skinName} (champ: ${champName}, skin: ${skin.id}${chroma ? ', chroma: ' + chroma.id : ''})`);
  if (chroma) {
    const payload = {
      type: 'apply', championId, skinId: chroma.id, baseSkinId: chroma.baseSkinId,
      championName: champName, skinName: chroma.baseSkinName || skin.name, chromaName: chroma.chromaName,
    };
    lastPrefetchPayload = payload;
    wsSendApply(payload);
  } else {
    const payload = { type: 'apply', championId, skinId: skin.id, championName: champName, skinName: skin.name };
    lastPrefetchPayload = payload;
    wsSendApply(payload);
  }

  setAppliedSkinName(skinName);
  setAppliedChromaId(chroma?.id || null);
  setButtonState(t('ui.applied'), true);
}

/**
 * Apply only custom mods (no skin). Used when default skin is selected but custom mods are enabled.
 */
async function applyCustomModsOnly() {
  if (!hasEnabledCustomMods()) return;
  if (isApplyInFlight()) return;
  if (isOverlayActive()) return;

  const championId = lastTrackedChampion || await getMyChampionId();
  if (!championId) return;
  const champName = await getChampionName(championId);
  logger.log(` applyCustomModsOnly: applying custom mods for ${champName}`);
  wsSendApply({ type: 'apply', championId, skinId: 0, championName: champName, skinName: '' });
}

// --- Reset ---

export function resetAutoApply(keepPayload = false) {
  lastTrackedSkin = null;
  lastTrackedChampion = null;
  stableSince = null;
  autoApplyTriggered = false;
  epoch++;
  clearSelectedChroma();
  setAppliedChromaId(null);
  logger.log(` resetAutoApply: clearing skinForced`);
  setSkinForced(false);
  pendingClickBack = null;
  cancelPrefetch();
  if (!keepPayload) {
    lastPrefetchPayload = null;
    if (retriggerTimer) { clearTimeout(retriggerTimer); retriggerTimer = null; }
  }
}

/**
 * Track champ select phase so retrigger knows whether to send prefetch or apply.
 */
export function setChampSelectActive(active) {
  champSelectActive = active;
}

/**
 * Stop retriggering — call when the game starts (InProgress) so room party
 * updates can no longer rebuild the overlay mid-game.
 */
export function lockRetrigger() {
  lastPrefetchPayload = null;
  if (retriggerTimer) { clearTimeout(retriggerTimer); retriggerTimer = null; }
  retriggerRetries = 0;
}

export function retriggerPrefetch() {
  if (retriggerTimer) { clearTimeout(retriggerTimer); retriggerTimer = null; }

  if (!lastPrefetchPayload) {
    logger.log(` retriggerPrefetch: no saved payload, skipping`);
    return;
  }

  // Staleness check: only applies during champ select where the DOM is live.
  if (champSelectActive) {
    const currentSkin = readCurrentSkin();
    if (currentSkin && lastPrefetchPayload.skinName !== currentSkin) {
      logger.log(` retriggerPrefetch: stale payload (${lastPrefetchPayload.skinName} != ${currentSkin}), skipping`);
      return;
    }
  }

  if (isApplyInFlight()) {
    if (retriggerRetries >= MAX_RETRIGGER_RETRIES) {
      logger.log(` retriggerPrefetch: max retries, scheduling final deferred retry`);
      retriggerRetries = 0;
      retriggerTimer = setTimeout(() => { retriggerTimer = null; retriggerPrefetch(); }, 10000);
      return;
    }
    retriggerRetries++;
    logger.log(` retriggerPrefetch: apply in-flight, scheduling retry ${retriggerRetries}/${MAX_RETRIGGER_RETRIES}`);
    retriggerTimer = setTimeout(() => { retriggerTimer = null; retriggerPrefetch(); }, 2000);
    return;
  }
  retriggerRetries = 0;

  // During champ select: send prefetch (just build, don't start overlay).
  // After champ select: send apply (rebuild + restart overlay).
  if (champSelectActive && !isOverlayActive()) {
    logger.log(` retriggerPrefetch: re-sending prefetch for ${lastPrefetchPayload.skinId} (${lastPrefetchPayload.skinName})`);
    wsSend(lastPrefetchPayload);
  } else {
    logger.log(` retriggerPrefetch: sending as apply for ${lastPrefetchPayload.skinId} (${lastPrefetchPayload.skinName})`);
    wsSendApply({ ...lastPrefetchPayload, type: 'apply' });
  }
}

// --- Stability check (called every poll cycle) ---

export function checkAutoApply(championId, isCurrentSkinOwned) {
  const skinName = readCurrentSkin();
  const chroma = getSelectedChroma();

  const currentKey = `${skinName || ''}|${chroma?.id || ''}`;
  const appliedKey = `${getAppliedSkinName() || ''}|${getAppliedChromaId() || ''}`;

  // Re-arm if user changed selection since last apply
  if (getAppliedSkinName()) {
    if (currentKey !== appliedKey) {
      setAppliedSkinName(null);
      setAppliedChromaId(null);
      setButtonState(t('ui.apply_skin'), false);
      autoApplyTriggered = false;
    } else {
      return;
    }
  }

  if (!skinName || !championId) return;

  // null = ownership data not loaded yet; true = owned skin
  if (isCurrentSkinOwned !== false) {
    autoApplyTriggered = false;
    // Don't clear the saved payload when we deliberately forced to base skin
    if (!getSkinForced()) lastPrefetchPayload = null;
    return;
  }

  const skinChanged = skinName !== lastTrackedSkin;
  const champChanged = championId !== lastTrackedChampion;

  if (skinChanged || champChanged) {
    if (skinChanged) clearSelectedChroma();

    lastTrackedSkin = skinName;
    lastTrackedChampion = championId;
    stableSince = Date.now();
    autoApplyTriggered = false;
    debouncePrefetch(championId, skinName);
    return;
  }

  if (autoApplyTriggered) return;

  if (stableSince && Date.now() - stableSince >= AUTO_APPLY_STABLE_MS) {
    triggerAutoApply();
  }
}

// --- Auto-apply trigger ---

async function triggerAutoApply() {
  if (autoApplyTriggered) return;
  autoApplyTriggered = true;

  const startEpoch = epoch;
  const startSkin = lastTrackedSkin;
  const startChampion = lastTrackedChampion;
  const startChroma = getSelectedChroma();

  if (getAppliedSkinName()) {
    autoApplyTriggered = false;
    return;
  }

  const championId = startChampion;
  if (!championId) {
    autoApplyTriggered = false;
    stableSince = Date.now();
    return;
  }

  const skins = await loadChampionSkins(championId);

  if (epoch !== startEpoch) return;

  if (!skins) {
    autoApplyTriggered = false;
    stableSince = Date.now();
    return;
  }

  // Verify nothing changed during the await
  const currentChroma = getSelectedChroma();
  if (lastTrackedSkin !== startSkin || lastTrackedChampion !== startChampion || currentChroma?.id !== startChroma?.id) {
    autoApplyTriggered = false;
    stableSince = Date.now();
    return;
  }

  const skin = findSkinByName(skins, startSkin);
  if (!skin) {
    autoApplyTriggered = false;
    stableSince = Date.now();
    return;
  }

  if (isDefaultSkin(skin)) {
    autoApplyTriggered = false;
    return;
  }

  if (getAppliedSkinName()) {
    autoApplyTriggered = false;
    return;
  }

  // Final DOM check
  if (readCurrentSkin() !== startSkin) {
    autoApplyTriggered = false;
    stableSince = Date.now();
    return;
  }

  logger.log(` forceDefaultSkin(${championId}) calling...`);
  const forced = await forceDefaultSkin(championId);
  logger.log(` forceDefaultSkin result: ${forced}`);
  if (!forced) {
    autoApplyTriggered = false;
    stableSince = Date.now();
    return;
  }
  setSkinForced(true);
  logger.log(` skinForced set to true`);

  const champName = await getChampionName(championId);
  if (startChroma) {
    wsSendApply({
      type: 'apply', championId, skinId: startChroma.id, baseSkinId: startChroma.baseSkinId,
      championName: champName, skinName: startChroma.baseSkinName || skin.name, chromaName: startChroma.chromaName,
    });
    notifySkinChange(championId, startChroma.id, startChroma.baseSkinId, champName, startChroma.baseSkinName || skin.name, startChroma.chromaName);
  } else {
    wsSendApply({ type: 'apply', championId, skinId: skin.id, championName: champName, skinName: skin.name });
    notifySkinChange(championId, skin.id, '', champName, skin.name, '');
  }

  setAppliedSkinName(startSkin);
  setAppliedChromaId(startChroma?.id || null);
  setButtonState(t('ui.applied'), true);
}

// --- Debug helpers ---

export function fetchAndLogTimer() {
  return fetchJson('/lol-champ-select/v1/session/timer');
}

export function fetchAndLogGameflow() {
  return fetchJson('/lol-gameflow/v1/session');
}
