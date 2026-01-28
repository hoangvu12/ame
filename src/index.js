import { CHAMP_SELECT_PHASES, POST_GAME_PHASES, POLL_INTERVAL_MS, CHROMA_BTN_CLASS } from './constants';
import { getMyChampionId, loadChampionSkins, resetSkinsCache } from './api';
import { injectStyles, removeStyles, unlockSkinCarousel } from './styles';
import { wsConnect, wsSend } from './websocket';
import { ensureApplyButton, removeApplyButton, updateButtonState } from './ui';
import { ensureChromaButton, closeChromaPanel, setLastChampionId, setAppliedSkinName } from './chroma';
import { handleSessionUpdate, resetAutoApply, fetchAndLogGameflow, fetchAndLogTimer, checkSkinChange } from './autoApply';

let pollTimer = null;
let observer = null;
let inChampSelect = false;
let lastChampionId = null;
let injectionTriggered = false;

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
  ensureApplyButton();
  unlockSkinCarousel();
  updateButtonState();

  const champId = lastChampionId || await getMyChampionId();
  if (champId) {
    await loadChampionSkins(champId);
    ensureChromaButton();
  }

  // Check for skin changes to reset auto-apply timer
  checkSkinChange();
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

    // Feed session to auto-apply logic
    handleSessionUpdate(session);
  });

  context.socket.observe('/lol-gameflow/v1/gameflow-phase', (event) => {
    const phase = event.data;
    console.log('[ame] Gameflow phase:', phase);

    const wasInChampSelect = inChampSelect;
    inChampSelect = CHAMP_SELECT_PHASES.includes(phase);

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
      resetAutoApply();
      injectionTriggered = true;
    }

    if (injectionTriggered && POST_GAME_PHASES.includes(phase)) {
      console.log('[ame] Game ended - sending cleanup');
      wsSend({ type: 'cleanup' });
      injectionTriggered = false;
    }
  });
}
