import { loadChampionSkins, getMyChampionId, getChampionName, fetchJson, forceDefaultSkin } from './api';
import { readCurrentSkin, findSkinByName, isDefaultSkin } from './skin';
import { wsSend, wsSendApply, isApplyInFlight, isOverlayActive } from './websocket';
import {
  getAppliedSkinName, setAppliedSkinName,
  getSelectedChroma, setSelectedChroma, clearSelectedChroma,
  getAppliedChromaId, setAppliedChromaId,
} from './state';
import { setButtonState } from './ui';
import { PREFETCH_DEBOUNCE_MS } from './constants';

const AUTO_APPLY_STABLE_MS = 10000;
const LOG = '[ame:auto]';

// Tracking state (local to this module — not shared)
let lastTrackedSkin = null;
let lastTrackedChampion = null;
let stableSince = null;
let autoApplyTriggered = false;
let epoch = 0;
let prefetchTimer = null;

/**
 * Notify auto-apply that a chroma was selected — resets stability timer.
 */
export function onChromaSelected(chromaId, baseSkinId, chromaName = null, baseSkinName = null) {
  console.log(`${LOG} Chroma selected: ${chromaId} (base: ${baseSkinId})`);
  setSelectedChroma(chromaId, baseSkinId, chromaName, baseSkinName);
  stableSince = Date.now();
  autoApplyTriggered = false;
}

/**
 * Send an immediate prefetch for a chroma (explicit click, no debounce).
 */
export function prefetchChroma(championId, chromaId, baseSkinId, championName = null, skinName = null, chromaName = null) {
  wsSend({ type: 'prefetch', championId, skinId: chromaId, baseSkinId, championName, skinName, chromaName });
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
      wsSend({ type: 'prefetch', championId, skinId: skin.id, championName: champName, skinName: skin.name });
    }
  }, PREFETCH_DEBOUNCE_MS);
}

// --- Force apply (last resort before game starts) ---

export async function forceApplyIfNeeded() {
  if (getAppliedSkinName()) return;
  if (isApplyInFlight()) return;
  if (isOverlayActive()) return;

  const skinName = lastTrackedSkin || readCurrentSkin();
  const championId = lastTrackedChampion || await getMyChampionId();
  if (!skinName || !championId) return;

  const skins = await loadChampionSkins(championId);
  if (!skins) return;

  const skin = findSkinByName(skins, skinName);
  if (!skin || isDefaultSkin(skin)) return;

  // No forceDefaultSkin here — this runs after leaving champ select, so the
  // PATCH session endpoint is already gone. The overlay still needs to be
  // applied as a last resort before the game loads.

  const champName = await getChampionName(championId);
  const chroma = getSelectedChroma();
  if (chroma) {
    wsSendApply({
      type: 'apply', championId, skinId: chroma.id, baseSkinId: chroma.baseSkinId,
      championName: champName, skinName: chroma.baseSkinName || skin.name, chromaName: chroma.chromaName,
    });
  } else {
    wsSendApply({ type: 'apply', championId, skinId: skin.id, championName: champName, skinName: skin.name });
  }

  setAppliedSkinName(skinName);
  setAppliedChromaId(chroma?.id || null);
  setButtonState('Applied', true);
}

// --- Reset ---

export function resetAutoApply() {
  lastTrackedSkin = null;
  lastTrackedChampion = null;
  stableSince = null;
  autoApplyTriggered = false;
  epoch++;
  clearSelectedChroma();
  setAppliedChromaId(null);
  cancelPrefetch();
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
      setButtonState('Apply Skin', false);
      autoApplyTriggered = false;
    } else {
      return;
    }
  }

  if (!skinName || !championId) return;

  // null = ownership data not loaded yet; true = owned skin
  if (isCurrentSkinOwned !== false) {
    autoApplyTriggered = false;
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

  const forced = await forceDefaultSkin(championId);
  if (!forced) {
    autoApplyTriggered = false;
    stableSince = Date.now();
    return;
  }

  const champName = await getChampionName(championId);
  if (startChroma) {
    wsSendApply({
      type: 'apply', championId, skinId: startChroma.id, baseSkinId: startChroma.baseSkinId,
      championName: champName, skinName: startChroma.baseSkinName || skin.name, chromaName: startChroma.chromaName,
    });
  } else {
    wsSendApply({ type: 'apply', championId, skinId: skin.id, championName: champName, skinName: skin.name });
  }

  setAppliedSkinName(startSkin);
  setAppliedChromaId(startChroma?.id || null);
  setButtonState('Applied', true);
}

// --- Debug helpers ---

export function fetchAndLogTimer() {
  return fetchJson('/lol-champ-select/v1/session/timer');
}

export function fetchAndLogGameflow() {
  return fetchJson('/lol-gameflow/v1/session');
}
