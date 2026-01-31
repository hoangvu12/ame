import { CHAMP_SELECT_PHASES, POST_GAME_PHASES, IN_GAME_PHASES, IN_GAME_POLL_MS, POLL_INTERVAL_MS, CHROMA_BTN_CLASS } from './constants';
import { getMyChampionId, loadChampionSkins, resetSkinsCache, fetchJson } from './api';
import { injectStyles, unlockSkinCarousel } from './styles';
import { wsConnect, wsSend, isOverlayActive } from './websocket';
import { ensureApplyButton, removeApplyButton, updateButtonState } from './ui';
import { ensureChromaButton, closeChromaPanel } from './chroma';
import { resetAutoApply, forceApplyIfNeeded, fetchAndLogGameflow, fetchAndLogTimer, checkAutoApply } from './autoApply';
import { ensureInGameUI, removeInGameUI, updateInGameStatus } from './inGame';
import { initSettings } from './settings';
import { handleReadyCheck, cancelPendingAccept, loadAutoAcceptSetting } from './autoAcceptMatch';
import { ensureBenchSwap, cleanupBenchSwap, loadBenchSwapSetting } from './benchSwap';
import { setLastChampionId, setAppliedSkinName } from './state';

let pollTimer = null;
let observer = null;
let inChampSelect = false;
let inGame = false;
let inGamePollTimer = null;
let lastChampionId = null;
let injectionTriggered = false;
let pollRunning = false;

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
    ensureBenchSwap();
    unlockSkinCarousel();
    updateButtonState();

    const champId = lastChampionId || await getMyChampionId();
    if (champId) {
      await loadChampionSkins(champId);
      ensureChromaButton();
    }

    checkAutoApply(champId);
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

export function init(context) {
  injectStyles();
  wsConnect();
  initSettings();
  loadAutoAcceptSetting();
  loadBenchSwapSetting();

  context.socket.observe('/lol-champ-select/v1/session', (event) => {
    if (event.eventType === 'Delete' || !inChampSelect) return;
    const session = event.data;
    const me = session.myTeam?.find(p => p.cellId === session.localPlayerCellId);
    const champId = me?.championId || null;
    if (champId && champId !== lastChampionId) {
      lastChampionId = champId;
      setLastChampionId(champId);
      resetSkinsCache();
    }
  });

  function handlePhase(phase) {
    if (phase === 'ReadyCheck') {
      handleReadyCheck();
    } else {
      cancelPendingAccept();
    }

    const wasInChampSelect = inChampSelect;
    inChampSelect = CHAMP_SELECT_PHASES.includes(phase);

    const wasInGame = inGame;
    inGame = IN_GAME_PHASES.includes(phase);

    if (inChampSelect && !wasInChampSelect) {
      lastChampionId = null;
      setLastChampionId(null);
      resetSkinsCache();
      injectionTriggered = false;
      setAppliedSkinName(null);
      resetAutoApply();
      startObserving();
      fetchAndLogGameflow();
      fetchAndLogTimer();
    } else if (!inChampSelect && wasInChampSelect) {
      stopObserving();
      forceApplyIfNeeded().then(() => resetAutoApply());
      injectionTriggered = true;
    }

    if (inGame && !wasInGame) {
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
