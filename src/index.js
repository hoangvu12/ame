import { CHAMP_SELECT_PHASES, POST_GAME_PHASES, IN_GAME_PHASES, IN_GAME_POLL_MS, POLL_INTERVAL_MS, CHROMA_BTN_CLASS } from './constants';
import { getMyChampionId, loadChampionSkins, resetSkinsCache } from './api';
import { injectStyles, removeStyles, unlockSkinCarousel } from './styles';
import { wsConnect, wsSend, isOverlayActive } from './websocket';
import { ensureApplyButton, removeApplyButton, updateButtonState } from './ui';
import { ensureChromaButton, closeChromaPanel, setLastChampionId, setAppliedSkinName } from './chroma';
import { resetAutoApply, forceApplyIfNeeded, fetchAndLogGameflow, fetchAndLogTimer, checkAutoApply } from './autoApply';
import { ensureInGameUI, removeInGameUI, updateInGameStatus } from './inGame';
import { initSettings } from './settings';

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
  document.querySelectorAll(`.${CHROMA_BTN_CLASS}`).forEach(el => el.remove());
  removeStyles();
}

async function pollUI() {
  if (!inChampSelect) return;
  if (pollRunning) return;
  pollRunning = true;
  try {
    ensureApplyButton();
    unlockSkinCarousel();
    updateButtonState();

    const champId = lastChampionId || await getMyChampionId();
    if (champId) {
      await loadChampionSkins(champId);
      ensureChromaButton();
    }

    // Check stability for auto-apply
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
  console.log('[ame] Observing champ select UI...');
}

export function init(context) {
  console.log('[ame] Plugin loaded');
  injectStyles();
  wsConnect();
  initSettings();

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
    console.log('[ame] Gameflow phase:', phase);

    const wasInChampSelect = inChampSelect;
    inChampSelect = CHAMP_SELECT_PHASES.includes(phase);

    const wasInGame = inGame;
    inGame = IN_GAME_PHASES.includes(phase);

    if (inChampSelect && !wasInChampSelect) {
      console.log('[ame] Entered champ select');
      lastChampionId = null;
      setLastChampionId(null);
      resetSkinsCache();
      injectionTriggered = false;
      setAppliedSkinName(null);
      resetAutoApply();
      startObserving();

      // Log gameflow and timer info for debugging
      fetchAndLogGameflow();
      fetchAndLogTimer();
    } else if (!inChampSelect && wasInChampSelect) {
      console.log('[ame] Left champ select');
      stopObserving();
      // Last resort: if nothing was applied yet, apply immediately before game starts
      forceApplyIfNeeded().then(() => resetAutoApply());
      injectionTriggered = true;
    }

    // In-game UI polling
    if (inGame && !wasInGame) {
      console.log('[ame] Entered in-game phase');
      injectStyles();
      inGamePollTimer = setInterval(() => {
        ensureInGameUI();
        updateInGameStatus();
      }, IN_GAME_POLL_MS);
    } else if (!inGame && wasInGame) {
      console.log('[ame] Left in-game phase');
      if (inGamePollTimer) { clearInterval(inGamePollTimer); inGamePollTimer = null; }
      removeInGameUI();
    }

    if (injectionTriggered && POST_GAME_PHASES.includes(phase)) {
      if (isOverlayActive()) {
        console.log('[ame] Game ended - sending cleanup');
        wsSend({ type: 'cleanup' });
      }
      injectionTriggered = false;
    }
  }

  context.socket.observe('/lol-gameflow/v1/gameflow-phase', (event) => {
    handlePhase(event.data);
  });

  // Fetch current phase on init (observer only fires on changes, not initial state)
  fetch('/lol-gameflow/v1/gameflow-phase')
    .then(r => r.ok ? r.json() : null)
    .then(phase => { if (phase) handlePhase(phase); })
    .catch(() => {});
}
