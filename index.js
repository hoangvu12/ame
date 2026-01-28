const SKIN_SELECTORS = [
  '.skin-name-text', // Classic Champ Select
  '.skin-name',      // Swiftplay lobby
];
const POLL_INTERVAL_MS = 300;
const CHAMP_SELECT_PHASES = ['ChampSelect'];
const POST_GAME_PHASES = ['None', 'Lobby', 'EndOfGame', 'PreEndOfGame', 'Matchmaking', 'ReadyCheck'];
const WS_URL = 'ws://localhost:18765';
const WS_RECONNECT_BASE_MS = 1000;
const WS_RECONNECT_MAX_MS = 30000;
const BUTTON_ID = 'ame-apply-btn';
const STYLE_ID = 'ame-styles';

let lastChampionId = null;
let pollTimer = null;
let observer = null;
let inChampSelect = false;
let championSkins = null;
let cachedChampionId = null;
let ws = null;
let wsReconnectDelay = WS_RECONNECT_BASE_MS;
let wsReconnectTimer = null;
let injectionTriggered = false;
let appliedSkinName = null;

async function getMyChampionId() {
  try {
    const res = await fetch('/lol-champ-select/v1/session');
    if (!res.ok) return null;
    const session = await res.json();
    const me = session.myTeam?.find(p => p.cellId === session.localPlayerCellId);
    return me?.championId || null;
  } catch {
    return null;
  }
}

async function loadChampionSkins(championId) {
  if (championId === cachedChampionId && championSkins) return championSkins;
  try {
    const res = await fetch(`/lol-game-data/assets/v1/champions/${championId}.json`);
    if (!res.ok) return null;
    const data = await res.json();
    cachedChampionId = championId;
    championSkins = data.skins || [];
    return championSkins;
  } catch {
    return null;
  }
}

function slugify(str) {
  return str
    .normalize('NFKD')
    .toLowerCase()
    .replace(/[^\p{L}\p{N}]/gu, '')
}

function findSkinByName(skins, name) {
  const exact = skins.find(s => s.name === name);
  if (exact) return exact;
  const slug = slugify(name);
  return skins.find(s => slugify(s.name) === slug) || null;
}

function readCurrentSkin() {
  for (const selector of SKIN_SELECTORS) {
    const nodes = document.querySelectorAll(selector);
    let candidate = null;
    for (const node of nodes) {
      const name = node.textContent.trim();
      if (name && node.offsetParent !== null) candidate = name;
    }
    if (candidate) return candidate;
  }
  return null;
}

// --- Style injection & skin unlock ---

function injectStyles() {
  if (document.getElementById(STYLE_ID)) return;
  const style = document.createElement('style');
  style.id = STYLE_ID;
  style.textContent = `
    .skin-selection-item.disabled {
      filter: none !important;
      -webkit-filter: none !important;
      opacity: 1 !important;
      pointer-events: auto !important;
    }
    .skin-selection-item .lock-icon,
    .skin-selection-item .locked-icon,
    .skin-selection-item .skin-selection-item-unowned-overlay,
    .unlock-skin-hit-area,
    .unlock-skin-hit-area * {
      display: none !important;
      visibility: hidden !important;
      pointer-events: none !important;
      width: 0 !important;
      height: 0 !important;
      overflow: hidden !important;
    }
    .skin-selection-item:hover .skin-selection-thumbnail,
    .skin-selection-item.selected .skin-selection-thumbnail {
      border: 2px solid #c8aa6e;
      box-shadow: 0 0 6px #c8aa6e80;
    }
    #ame-apply-btn[disabled] {
      opacity: 0.4 !important;
      pointer-events: none !important;
      cursor: default !important;
      filter: grayscale(0.5) !important;
    }
  `;
  document.head.appendChild(style);
}

function removeStyles() {
  const style = document.getElementById(STYLE_ID);
  if (style) style.remove();
}

function unlockSkinCarousel() {
  const items = document.querySelectorAll('.skin-selection-item.disabled');
  for (const el of items) {
    el.classList.remove('disabled');
  }
}

// --- Apply button ---

function findSkinNameElement() {
  for (const selector of SKIN_SELECTORS) {
    const nodes = document.querySelectorAll(selector);
    for (const node of nodes) {
      if (node.textContent.trim() && node.offsetParent !== null) return node;
    }
  }
  return null;
}

function ensureApplyButton() {
  if (document.getElementById(BUTTON_ID)) return;
  const container = document.querySelector('.toggle-ability-previews-button-container');
  if (!container) return;

  const btn = document.createElement('lol-uikit-flat-button');
  btn.id = BUTTON_ID;
  btn.textContent = 'Apply Skin';
  btn.classList.add('toggle-ability-previews-button');
  btn.addEventListener('click', onApplyClick);

  container.appendChild(btn);
}

function removeApplyButton() {
  const btn = document.getElementById(BUTTON_ID);
  if (btn) btn.remove();
}

function setButtonState(text, disabled) {
  const btn = document.getElementById(BUTTON_ID);
  if (!btn) return;
  btn.textContent = text;
  if (disabled) {
    btn.setAttribute('disabled', '');
  } else {
    btn.removeAttribute('disabled');
  }
}

async function onApplyClick() {
  const skinName = readCurrentSkin();
  if (!skinName) return;

  const championId = await getMyChampionId();
  if (!championId) {
    if (typeof Toast !== 'undefined') Toast.error('Pick a champion first');
    return;
  }

  const skins = await loadChampionSkins(championId);
  if (!skins) {
    if (typeof Toast !== 'undefined') Toast.error('Could not load skin data');
    return;
  }

  const skin = findSkinByName(skins, skinName);
  if (!skin) {
    if (typeof Toast !== 'undefined') Toast.error('Skin not found in game data');
    return;
  }

  console.log('[ame] Apply clicked:', skinName, '| Skin ID:', skin.id, '| Champion ID:', championId);
  wsSend({ type: 'apply', championId, skinId: skin.id });
  appliedSkinName = skinName;
  setButtonState('Applied', true);
}

// --- Poll for button injection ---

function pollUI() {
  if (!inChampSelect) return;
  ensureApplyButton();
  unlockSkinCarousel();
  updateButtonState();
}

function updateButtonState() {
  if (!appliedSkinName) return;
  const current = readCurrentSkin();
  if (!current) return;
  if (current === appliedSkinName) {
    setButtonState('Applied', true);
  } else {
    setButtonState('Apply Skin', false);
  }
}

// --- Observing ---

function stopObserving() {
  if (observer) { observer.disconnect(); observer = null; }
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null; }
  removeApplyButton();
  removeStyles();
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

// --- WebSocket ---

function wsConnect() {
  if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) return;
  try {
    ws = new WebSocket(WS_URL);
    ws.onopen = () => {
      console.log('[ame] WebSocket connected');
      wsReconnectDelay = WS_RECONNECT_BASE_MS;
    };
    ws.onmessage = (e) => {
      try {
        const msg = JSON.parse(e.data);
        if (msg.type === 'status') {
          if (typeof Toast !== 'undefined') {
            if (msg.status === 'error') {
              Toast.error(msg.message);
            } else if (msg.status === 'ready') {
              Toast.success(msg.message);
            } else if (msg.status === 'downloading' || msg.status === 'injecting') {
              Toast.success(msg.message);
            }
          }
        }
      } catch (err) {
        console.log('[ame] onmessage error:', err);
      }
    };
    ws.onclose = () => wsScheduleReconnect();
    ws.onerror = () => {};
  } catch {
    wsScheduleReconnect();
  }
}

function wsScheduleReconnect() {
  if (wsReconnectTimer) return;
  wsReconnectTimer = setTimeout(() => {
    wsReconnectTimer = null;
    wsReconnectDelay = Math.min(wsReconnectDelay * 2, WS_RECONNECT_MAX_MS);
    wsConnect();
  }, wsReconnectDelay);
}

function wsSend(obj) {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(obj));
  }
}

// --- Entry point ---

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
      championSkins = null;
      cachedChampionId = null;
    }
  });

  context.socket.observe('/lol-gameflow/v1/gameflow-phase', (event) => {
    const phase = event.data;
    console.log('[ame] Gameflow phase:', phase);

    const wasInChampSelect = inChampSelect;
    inChampSelect = CHAMP_SELECT_PHASES.includes(phase);

    if (inChampSelect && !wasInChampSelect) {
      console.log('[ame] Entered champ select');
      lastChampionId = null;
      championSkins = null;
      cachedChampionId = null;
      injectionTriggered = false;
      appliedSkinName = null;
      startObserving();
    } else if (!inChampSelect && wasInChampSelect) {
      console.log('[ame] Left champ select');
      stopObserving();
      injectionTriggered = true;
    }

    if (injectionTriggered && POST_GAME_PHASES.includes(phase)) {
      console.log('[ame] Game ended - sending cleanup');
      wsSend({ type: 'cleanup' });
      injectionTriggered = false;
    }
  });
}
