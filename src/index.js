import { CHAMP_SELECT_PHASES, POST_GAME_PHASES, IN_GAME_PHASES, IN_GAME_POLL_MS, POLL_INTERVAL_MS, CHROMA_BTN_CLASS } from './constants';
import { ensureSwiftplayButton, removeSwiftplayButton, updateSwiftplayButtonState, unlockSwiftplayCarousel, isSwiftplaySkinPanelOpen, ensureSwiftplayConnectionBanner, updateSwiftplayConnectionBanner } from './swiftplay';
import { getMyChampionId, getChampionSkins, loadChampionSkins, resetSkinsCache, fetchJson, fetchSummonerId, fetchOwnedSkins, forceDefaultSkin, getChampionIdFromLobbyDOM } from './api';
import { injectStyles, unlockSkinCarousel } from './styles';
import { wsConnect, wsSend, isOverlayActive } from './websocket';
import { ensureApplyButton, removeApplyButton, updateButtonState, ensureConnectionBanner, updateConnectionBanner, initConnectionStatus } from './ui';
import { ensureChromaButton, closeChromaPanel } from './chroma';
import { resetAutoApply, forceApplyIfNeeded, fetchAndLogGameflow, fetchAndLogTimer, checkAutoApply, lockRetrigger, setChampSelectActive, processClickBack } from './autoApply';
import { ensureInGameUI, removeInGameUI, updateInGameStatus } from './inGame';
import { initSettings } from './settings';
import { handleReadyCheck, cancelPendingAccept, loadAutoAcceptSetting } from './autoAcceptMatch';
import { ensureBenchSwap, cleanupBenchSwap, loadBenchSwapSetting } from './benchSwap';
import { loadAutoSelectSetting, handleChampSelectSession, resetAutoSelect } from './autoSelect';
import { setLastChampionId, setAppliedSkinName, setOwnedSkinIds, resetOwnedSkins, getOwnedSkinChampionId, isOwnedSkin, getPendingForceDefault, setPendingForceDefault } from './state';
import { readCurrentSkin, findSkinByName } from './skin';
import { joinRoom, leaveRoom, loadRoomPartySetting, flushPendingRetrigger } from './roomParty';
import { triggerRandomSkin, resetRandomSkin } from './randomSkin';
import { initChatStatus } from './chatStatus';
import { initI18n } from './i18n';
import { createLogger } from './logger';

const logger = createLogger('core');


let pollTimer = null;
let observer = null;
let inChampSelect = false;
let inGame = false;
let inSwiftplay = false;
let inGamePollTimer = null;
let swiftplayPollTimer = null;
let swiftplayObserver = null;
let swiftplayPollRunning = false;
let lastChampionId = null;
let injectionTriggered = false;
let pollRunning = false;
let ownershipLoading = false;

async function ensureOwnershipCache(championId) {
  if (!championId) return;
  if (getOwnedSkinChampionId() === championId) return;
  if (ownershipLoading) return;
  ownershipLoading = true;
  try {
    const summonerId = await fetchSummonerId();
    if (!summonerId) return;
    const ownedIds = await fetchOwnedSkins(summonerId, championId);
    if (ownedIds) {
      setOwnedSkinIds(championId, ownedIds);
    }
  } finally {
    ownershipLoading = false;
  }
}

// Returns true (owned), false (not owned), or null (data not ready yet).
function resolveOwnership(champId) {
  const skinName = readCurrentSkin();
  if (!skinName) return null;
  const skins = getChampionSkins();
  if (!skins) return null;
  if (getOwnedSkinChampionId() !== champId) return null;
  const skin = findSkinByName(skins, skinName);
  if (!skin) return null;
  const owned = isOwnedSkin(skin.id);
  logger.log(`resolveOwnership: skin="${skinName}" id=${skin.id} owned=${owned}`);
  return owned;
}

function stopObserving() {
  if (observer) { observer.disconnect(); observer = null; }
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null; }
  removeApplyButton();
  closeChromaPanel();
  cleanupBenchSwap();
  document.querySelectorAll(`.${CHROMA_BTN_CLASS}`).forEach(el => el.remove());
}

async function pollUI() {
  if (!inChampSelect) return;
  if (pollRunning) return;
  pollRunning = true;
  try {
    ensureApplyButton();
    ensureConnectionBanner();
    updateConnectionBanner();
    ensureBenchSwap();
    unlockSkinCarousel();

    const champId = lastChampionId || await getMyChampionId();

    if (champId && getPendingForceDefault()) {
      await forceDefaultSkin(champId);
      setPendingForceDefault(false);
    }

    let ownership = null;
    if (champId) {
      await loadChampionSkins(champId);
      await ensureOwnershipCache(champId);
      ensureChromaButton();
      triggerRandomSkin(champId);
      ownership = resolveOwnership(champId);
    }

    updateButtonState(ownership);
    checkAutoApply(champId, ownership);
    processClickBack();
  } finally {
    pollRunning = false;
  }
}

function startObserving() {
  if (!document.body) {
    setTimeout(startObserving, 250);
    return;
  }
  stopObserving();
  observer = new MutationObserver(pollUI);
  observer.observe(document.body, { childList: true, subtree: true });
  pollTimer = setInterval(pollUI, POLL_INTERVAL_MS);
  pollUI();
}

function stopSwiftplayObserving() {
  if (swiftplayObserver) { swiftplayObserver.disconnect(); swiftplayObserver = null; }
  if (swiftplayPollTimer) { clearInterval(swiftplayPollTimer); swiftplayPollTimer = null; }
  removeSwiftplayButton();
}

async function pollSwiftplayUI() {
  if (!inSwiftplay) return;
  if (swiftplayPollRunning) return;
  swiftplayPollRunning = true;
  try {
    if (!isSwiftplaySkinPanelOpen()) {
      removeSwiftplayButton();
      return;
    }
    ensureSwiftplayConnectionBanner();
    updateSwiftplayConnectionBanner();
    ensureSwiftplayButton();
    unlockSwiftplayCarousel();

    let ownership = null;
    const champId = await getChampionIdFromLobbyDOM();
    if (champId) {
      await loadChampionSkins(champId);
      await ensureOwnershipCache(champId);
      ownership = resolveOwnership(champId);
    }

    updateSwiftplayButtonState(ownership);
  } finally {
    swiftplayPollRunning = false;
  }
}

function startSwiftplayObserving() {
  if (!document.body) {
    setTimeout(startSwiftplayObserving, 250);
    return;
  }
  stopSwiftplayObserving();
  swiftplayObserver = new MutationObserver(pollSwiftplayUI);
  swiftplayObserver.observe(document.body, { childList: true, subtree: true });
  swiftplayPollTimer = setInterval(pollSwiftplayUI, POLL_INTERVAL_MS);
  pollSwiftplayUI();
}

export async function init(context) {
  await initI18n();
  injectStyles();
  wsConnect();
  initConnectionStatus();
  initSettings();
  loadAutoAcceptSetting();
  loadBenchSwapSetting();
  loadAutoSelectSetting();
  loadRoomPartySetting();
  initChatStatus(context);

  context.socket.observe('/lol-champ-select/v1/session', (event) => {
    if (event.eventType === 'Delete' || !inChampSelect) return;
    const session = event.data;
    handleChampSelectSession(session);
    const me = session.myTeam?.find(p => p.cellId === session.localPlayerCellId);
    const champId = me?.championId || null;
    if (champId && champId !== lastChampionId) {
      lastChampionId = champId;
      setLastChampionId(champId);
      resetSkinsCache();
      resetOwnedSkins();
    }
    joinRoom(session);
  });

  function handlePhase(phase) {
    if (phase === 'ReadyCheck') {
      handleReadyCheck();
    } else {
      cancelPendingAccept();
    }

    const wasInChampSelect = inChampSelect;
    inChampSelect = CHAMP_SELECT_PHASES.includes(phase);

    const wasInSwiftplay = inSwiftplay;
    inSwiftplay = phase === 'Lobby';

    const wasInGame = inGame;
    inGame = IN_GAME_PHASES.includes(phase);

    if (inChampSelect && !wasInChampSelect) {
      if (wasInSwiftplay && isOverlayActive()) {
        setPendingForceDefault(true);
      }
      lastChampionId = null;
      setLastChampionId(null);
      resetSkinsCache();
      resetOwnedSkins();
      injectionTriggered = false;
      setAppliedSkinName(null);
      resetAutoApply();
      resetAutoSelect();
      resetRandomSkin();
      setChampSelectActive(true);
      stopSwiftplayObserving();
      startObserving();
      fetchAndLogGameflow();
      fetchAndLogTimer();
      leaveRoom();
      joinRoom();
    } else if (!inChampSelect && wasInChampSelect) {
      logger.log('Champ select ended, flushing retrigger + forceApply');
      setChampSelectActive(false);
      stopObserving();
      flushPendingRetrigger();
      forceApplyIfNeeded().finally(() => {
        logger.log('forceApplyIfNeeded settled');
        resetAutoApply(true);
      });
      resetAutoSelect();
      injectionTriggered = true;
    }

    if (inSwiftplay && !wasInSwiftplay) {
      setAppliedSkinName(null);
      resetOwnedSkins();
      startSwiftplayObserving();
    } else if (!inSwiftplay && wasInSwiftplay) {
      stopSwiftplayObserving();
    }

    if (inGame && !wasInGame) {
      lockRetrigger();
      injectStyles();
      inGamePollTimer = setInterval(() => {
        ensureInGameUI();
        updateInGameStatus();
      }, IN_GAME_POLL_MS);
    } else if (!inGame && wasInGame) {
      if (inGamePollTimer) { clearInterval(inGamePollTimer); inGamePollTimer = null; }
      removeInGameUI();
    }

    if (injectionTriggered && POST_GAME_PHASES.includes(phase)) {
      leaveRoom();
      if (isOverlayActive()) {
        wsSend({ type: 'cleanup' });
      }
      injectionTriggered = false;
    }
  }

  context.socket.observe('/lol-gameflow/v1/gameflow-phase', (event) => {
    handlePhase(event.data);
  });

  fetchJson('/lol-gameflow/v1/gameflow-phase').then(phase => {
    if (phase) handlePhase(phase);
  });
}
