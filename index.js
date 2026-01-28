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
const CHROMA_BTN_CLASS = 'ame-chroma-button';
const CHROMA_PANEL_ID = 'ame-chroma-panel-container';
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
let activeChromaPanel = null;
let activeChromaButton = null;
const chromaButtonDataMap = new WeakMap();

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

function getSkinNameFromItem(item) {
  if (!item) return null;

  const attrNames = [
    'data-skin-name',
    'data-skin',
    'data-name',
    'title',
    'aria-label',
  ];
  for (const attr of attrNames) {
    const val = item.getAttribute(attr);
    if (val && val.trim()) return val.trim();
  }

  const nameEl = item.querySelector('.skin-name, .skin-name-text, .skin-selection-item-name');
  if (nameEl && nameEl.textContent.trim()) return nameEl.textContent.trim();

  return null;
}

function isItemVisible(item) {
  if (!item) return false;
  const style = window.getComputedStyle(item);
  if (style.display === 'none' || style.visibility === 'hidden') return false;
  const opacity = parseFloat(style.opacity || '1');
  return opacity > 0.05;
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
    .unlock-skin-hit-area *,
    .locked-state {
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
    .skin-selection-carousel .skin-selection-item {
      position: relative;
      overflow: visible !important;
    }
    .skin-selection-button-container,
    .skin-selection-carousel-container,
    .skin-selection-carousel {
      overflow: visible !important;
    }
    .skin-selection-carousel {
      position: relative;
    }
    .skin-selection-thumbnail {
      position: relative;
      overflow: visible !important;
      pointer-events: auto !important;
    }
    .skin-selection-carousel,
    .skin-selection-carousel-container {
      overflow: visible !important;
    }
    .${CHROMA_BTN_CLASS} {
      position: absolute;
      right: 6px;
      bottom: 6px;
      z-index: 5;
      pointer-events: auto;
      cursor: pointer;
      display: block !important;
    }
    .${CHROMA_BTN_CLASS}.hidden {
      display: none !important;
      pointer-events: none !important;
    }
    #${CHROMA_PANEL_ID} {
      position: absolute;
      z-index: 10000;
      pointer-events: auto;
    }
    #${CHROMA_PANEL_ID} lc-flyout-content {
      display: block !important;
      opacity: 1 !important;
      visibility: visible !important;
      pointer-events: auto !important;
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
  // Hide lock overlays that can intercept clicks
  document.querySelectorAll('.unlock-skin-hit-area, .locked-state, .skin-selection-item-unowned-overlay').forEach(el => {
    el.style.display = 'none';
    el.style.visibility = 'hidden';
    el.style.pointerEvents = 'none';
  });
}

// --- Chroma ---

function getSkinOffset(skinItem) {
  const match = skinItem.className.match(/skin-carousel-offset-(-?\d+)/);
  if (match) return parseInt(match[1]);
  let parent = skinItem.parentElement;
  for (let i = 0; i < 3 && parent; i++) {
    const pm = parent.className.match(/skin-carousel-offset-(-?\d+)/);
    if (pm) return parseInt(pm[1]);
    parent = parent.parentElement;
  }
  return null;
}

function getSkinKeyFromItem(item) {
  if (!item) return null;
  const thumb = item.querySelector('.skin-selection-thumbnail');
  if (!thumb) return null;
  const bg = thumb.style.backgroundImage || window.getComputedStyle(thumb).backgroundImage || '';
  if (!bg) return null;
  const match = bg.match(/Skins\/(Base|Skin(\d+))\//i);
  if (!match) return null;
  if (match[1].toLowerCase() === 'base') return 0;
  const id = parseInt(match[2], 10);
  return Number.isFinite(id) ? id : null;
}

function findSkinByCarouselKey(skins, key) {
  if (!skins || key === null || key === undefined) return null;
  // Prefer "num" which matches SkinXX in the asset path
  let skin = skins.find(s => s.num === key);
  if (skin) return skin;
  // Fallback: some ids encode the skin number in the last 3 digits
  skin = skins.find(s => typeof s.id === 'number' && (s.id % 1000) === key);
  return skin || null;
}

function getChromaData(skins, currentSkinName) {
  if (!skins || !currentSkinName) return null;
  const skin = findSkinByName(skins, currentSkinName);
  console.log('[ame][chroma] findSkinByName("' + currentSkinName + '") =>', skin ? `id:${skin.id} name:"${skin.name}" chromas:${skin.chromas?.length}` : 'NOT FOUND',
    '| slugified input:', slugify(currentSkinName),
    '| slugified names:', skins.map(s => slugify(s.name)).join(', '));
  if (!skin || !skin.chromas || skin.chromas.length === 0) return null;
  return { skin, chromas: skin.chromas };
}

function closeChromaPanel() {
  if (activeChromaPanel) {
    activeChromaPanel.remove();
  }
  activeChromaPanel = null;
  activeChromaButton = null;
}

function onClickOutsideChroma(e) {
  const panel = activeChromaPanel || document.getElementById(CHROMA_PANEL_ID);
  const btn = document.querySelector(`.${CHROMA_BTN_CLASS}`);
  if (panel && !panel.contains(e.target) && (!btn || !btn.contains(e.target))) {
    closeChromaPanel();
    document.removeEventListener('click', onClickOutsideChroma, true);
  }
}

function createChromaButton() {
  const button = document.createElement('div');
  button.className = CHROMA_BTN_CLASS;
  button.classList.add('chroma-button', 'chroma-selection', 'uikit-framed-icon', 'ember-view');

  const outerMask = document.createElement('div');
  outerMask.className = 'outer-mask interactive';

  const frameColor = document.createElement('div');
  frameColor.className = 'frame-color';

  const content = document.createElement('div');
  content.className = 'content';

  const innerMask = document.createElement('div');
  innerMask.className = 'inner-mask inner-shadow';

  frameColor.appendChild(content);
  frameColor.appendChild(innerMask);
  outerMask.appendChild(frameColor);
  button.appendChild(outerMask);

  return button;
}

function getOrCreateChromaButton(container, data) {
  if (!container) return null;
  let btn = container.querySelector(`.${CHROMA_BTN_CLASS}`);
  if (!btn) {
    btn = createChromaButton();
    btn.addEventListener('click', (e) => {
      e.stopPropagation();
      e.preventDefault();
      const payload = chromaButtonDataMap.get(btn);
      if (!payload) return;
      console.log('[ame][chroma] Button click:', payload.skin?.name, '| chromas:', payload.chromas?.length);
      if (activeChromaPanel) {
        closeChromaPanel();
        document.removeEventListener('click', onClickOutsideChroma, true);
      } else {
        createChromaPanel(payload.skin, payload.chromas, btn, payload.championId || null);
      }
    });
    container.appendChild(btn);
  }
  if (data) chromaButtonDataMap.set(btn, data);
  return btn;
}

function createChromaPanel(skinData, chromas, buttonEl, championId) {
  closeChromaPanel();

  const carousel = buttonEl.closest('.skin-selection-carousel') || document.querySelector('.skin-selection-carousel');

  // Create flyout frame (provides the border styling)
  const flyout = document.createElement('lol-uikit-flyout-frame');
  flyout.id = CHROMA_PANEL_ID;
  flyout.className = 'flyout';
  flyout.setAttribute('orientation', 'top');
  flyout.setAttribute('show', 'true');
  flyout.style.position = 'absolute';
  flyout.style.overflow = 'visible';
  flyout.style.zIndex = '10000';

  (carousel || document.body).appendChild(flyout);

  const flyoutContent = document.createElement('lc-flyout-content');
  flyoutContent.dataset.ameChroma = '1';

  const modal = document.createElement('div');
  modal.className = 'champ-select-chroma-modal chroma-view ember-view';

  const champId = championId || lastChampionId || null;
  const getChromaImageUrl = (chroma) => {
    if (!chroma) return '';
    if (champId && chroma.id) {
      return `/lol-game-data/assets/v1/champion-chroma-images/${champId}/${chroma.id}.png`;
    }
    return '';
  };
  const getChromaPreviewPath = (chroma) => {
    if (!chroma) return '';
    return (
      chroma.chromaPreviewPath ||
      chroma.chromaPath ||
      chroma.chromaPreview ||
      chroma.imagePath ||
      chroma.splashPath ||
      chroma.tilePath ||
      ''
    );
  };

  // Chroma preview area
  const chromaInfo = document.createElement('div');
  chromaInfo.className = 'chroma-information';
  chromaInfo.style.backgroundImage = "url('lol-game-data/assets/content/src/LeagueClient/GameModeAssets/Classic_SRU/img/champ-select-flyout-background.jpg')";

  const chromaImage = document.createElement('div');
  chromaImage.className = 'chroma-information-image';
  if (chromas.length > 0) {
    const first = chromas[0];
    const imgUrl = getChromaImageUrl(first) || getChromaPreviewPath(first);
    if (imgUrl) chromaImage.style.backgroundImage = `url('${imgUrl}')`;
  }

  const skinName = document.createElement('div');
  skinName.className = 'child-skin-name';
  skinName.textContent = skinData.name;
  const disabledNote = document.createElement('div');
  disabledNote.className = 'child-skin-disabled-notification';
  skinName.appendChild(disabledNote);

  chromaInfo.appendChild(chromaImage);
  chromaInfo.appendChild(skinName);

  // Chroma selection dots
  const scrollable = document.createElement('lol-uikit-scrollable');
  scrollable.className = 'chroma-selection';
  scrollable.setAttribute('overflow-masks', 'enabled');
  scrollable.setAttribute('scrolled-bottom', 'false');
  scrollable.setAttribute('scrolled-top', 'true');
  const listHost = scrollable;

  for (let i = 0; i < chromas.length; i++) {
    const chroma = chromas[i];
    const wrapper = document.createElement('div');
    wrapper.className = 'ember-view';

    const btn = document.createElement('div');
    btn.className = 'chroma-skin-button';
    if (i === 0) btn.classList.add('selected');

    const contents = document.createElement('div');
    contents.className = 'contents';
    const colors = chroma.colors || [];
    if (colors.length >= 2) {
      contents.style.background = `linear-gradient(135deg, ${colors[0]} 50%, ${colors[1]} 50%)`;
    } else if (colors.length === 1) {
      contents.style.background = colors[0];
    } else {
      contents.style.background = '#27211C';
    }

    btn.appendChild(contents);
    wrapper.appendChild(btn);
    listHost.appendChild(wrapper);

    btn.addEventListener('click', () => {
      selectChroma(skinData, chroma);
      listHost.querySelectorAll('.chroma-skin-button').forEach(b => b.classList.remove('selected'));
      btn.classList.add('selected');
    });
    btn.addEventListener('mouseenter', () => {
      const preview = getChromaImageUrl(chroma) || getChromaPreviewPath(chroma);
      if (preview) chromaImage.style.backgroundImage = `url('${preview}')`;
      if (skinName.firstChild) {
        skinName.firstChild.nodeValue = chroma.name || skinData.name;
      } else {
        skinName.textContent = chroma.name || skinData.name;
        skinName.appendChild(disabledNote);
      }
    });
  }
  modal.appendChild(chromaInfo);
  modal.appendChild(scrollable);
  flyoutContent.appendChild(modal);
  // Remove any previous chroma content before attaching
  flyout.querySelectorAll('lc-flyout-content[data-ame-chroma="1"]').forEach(el => el.remove());
  flyout.appendChild(flyoutContent);

  activeChromaPanel = flyout;
  activeChromaButton = buttonEl;

  const container = carousel || document.body;
  const containerRect = container.getBoundingClientRect();
  const btnRect = buttonEl.getBoundingClientRect();
  const modalRect = modal.getBoundingClientRect();
  const width = modalRect.width || 305;
  const height = modalRect.height || 340;
  let left = btnRect.left - containerRect.left + btnRect.width / 2 - width / 2;
  let top = btnRect.top - containerRect.top - height - 8;
  // Prefer above the card; allow negative top if needed.
  flyout.style.left = `${Math.round(left)}px`;
  flyout.style.top = `${Math.round(top)}px`;
  flyout.style.bottom = '';

  setTimeout(() => document.addEventListener('click', onClickOutsideChroma, true), 0);
}

async function selectChroma(skinData, chroma) {
  const championId = await getMyChampionId();
  if (!championId) return;

  // Save reference before closing panel (which resets activeChromaButton)
  const triggerButton = activeChromaButton;

  console.log('[ame] Chroma selected:', chroma.name, '| ID:', chroma.id, '| Base skin:', skinData.id);
  wsSend({ type: 'apply', championId, skinId: chroma.id, baseSkinId: skinData.id });
  appliedSkinName = chroma.name || skinData.name;
  setButtonState('Applied', true);
  closeChromaPanel();
  document.removeEventListener('click', onClickOutsideChroma, true);

  // Update the chroma button that triggered the panel
  const contentEl = triggerButton?.querySelector('.content');
  if (contentEl) {
    const colors = chroma.colors || [];
    if (colors.length >= 2) {
      contentEl.style.background = `linear-gradient(135deg, ${colors[0]} 50%, ${colors[1]} 50%)`;
    } else if (colors.length === 1) {
      contentEl.style.background = colors[0];
    }
  }
}

function ensureChromaButton() {
  const skinItems = document.querySelectorAll('.skin-selection-item');
  if (skinItems.length === 0) return;

  // Prefer the selected item; fallback to visible center (offset-0), then closest.
  const selectedItem = document.querySelector('.skin-selection-item-selected, .skin-selection-item.selected');
  let centerItem = null;
  let centerReason = 'none';

  // Debug: log skin names and chroma counts once per skin set
  const debugKey = `debug-${championSkins?.length}`;
  if (championSkins && debugKey !== ensureChromaButton._lastDebug) {
    ensureChromaButton._lastDebug = debugKey;
    console.log('[ame][chroma] All skins:', championSkins.map(s => `${s.id} (num:${s.num}) :"${s.name}" chromas:${s.chromas?.length || 0}`).join(' | '));
  }

  // Find the center item (offset-0) as primary
  const offsets = [];
  const entries = [];
  let closestItem = null;
  let closestDelta = Number.POSITIVE_INFINITY;
  for (const item of skinItems) {
    const offset = getSkinOffset(item);
    offsets.push(offset);
    const visible = isItemVisible(item);
    entries.push({ item, offset, visible });
    if (offset !== null && offset !== undefined) {
      const delta = Math.abs(offset);
      if (delta < closestDelta) {
        closestDelta = delta;
        closestItem = item;
      }
    }
  }

  const offset0 = entries.find(e => e.visible && e.offset === 0)?.item;
  if (selectedItem && isItemVisible(selectedItem)) {
    centerItem = selectedItem;
    centerReason = 'selected-visible';
  } else if (offset0) {
    centerItem = offset0;
    centerReason = 'offset-0';
  } else if (closestItem) {
    centerItem = closestItem;
    centerReason = 'closest';
  }

  const nameFromItem = getSkinNameFromItem(centerItem);
  const skinKeyFromItem = getSkinKeyFromItem(centerItem);
  let skinName = nameFromItem || readCurrentSkin();
  let chromaData = null;

  if (skinKeyFromItem !== null && championSkins) {
    const skinByKey = findSkinByCarouselKey(championSkins, skinKeyFromItem);
    if (skinByKey && skinByKey.chromas && skinByKey.chromas.length > 0) {
      chromaData = { skin: skinByKey, chromas: skinByKey.chromas };
    }
  }

  if (!chromaData) {
    chromaData = (skinName && championSkins) ? getChromaData(championSkins, skinName) : null;
  }
  if (!chromaData && nameFromItem) {
    const altName = readCurrentSkin();
    if (altName && altName !== skinName) {
      const altData = getChromaData(championSkins, altName);
      if (altData) {
        skinName = altName;
        chromaData = altData;
      }
    }
  }

  // Log on state change only
  const centerOffset = getSkinOffset(centerItem);
  const centerOpacity = centerItem ? window.getComputedStyle(centerItem).opacity : 'n/a';
  const logKey = `${skinName}|${skinKeyFromItem}|${championSkins?.length}|${chromaData?.chromas?.length}|${offsets.join(',')}|${!!centerItem}|${centerReason}|${centerOffset}|${centerOpacity}`;
  if (logKey !== ensureChromaButton._lastLog) {
    ensureChromaButton._lastLog = logKey;
    console.log('[ame][chroma] skinName:', skinName,
      '| skinKeyFromItem:', skinKeyFromItem,
      '| skins loaded:', !!championSkins, '(count:', championSkins?.length, ')',
      '| chromaData:', chromaData ? `${chromaData.chromas.length} chromas for "${chromaData.skin.name}"` : 'null',
      '| offsets:', offsets.join(','),
      '| centerItem:', !!centerItem,
      '| centerReason:', centerReason,
      '| centerOffset:', centerOffset,
      '| centerOpacity:', centerOpacity,
      '| centerClasses:', centerItem?.className);
  }

  if (!centerItem || !chromaData) {
    closeChromaPanel();
  }

  // Show chroma button on every visible carousel item that has chromas
  for (const item of skinItems) {
    const key = getSkinKeyFromItem(item);
    let data = null;
    if (key !== null && championSkins) {
      const skinByKey = findSkinByCarouselKey(championSkins, key);
      if (skinByKey && skinByKey.chromas && skinByKey.chromas.length > 0) {
        data = { skin: skinByKey, chromas: skinByKey.chromas, championId: lastChampionId || null };
      }
    }
    const thumb = item.querySelector('.skin-selection-thumbnail') || item;
    if (!thumb) continue;
    if (!thumb.style.position) thumb.style.position = 'relative';
    thumb.style.pointerEvents = 'auto';
    thumb.style.overflow = 'visible';

    if (!data) {
      const existing = thumb.querySelector(`.${CHROMA_BTN_CLASS}`);
      if (existing) existing.remove();
      continue;
    }

    const btn = getOrCreateChromaButton(thumb, data);
    if (!btn) continue;
    btn.classList.remove('hidden');
    btn.style.position = 'absolute';
    btn.style.right = '6px';
    btn.style.bottom = '6px';
    btn.style.left = '';
    btn.style.top = '';
    btn.style.zIndex = '50';
    btn.style.pointerEvents = 'auto';
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

async function pollUI() {
  if (!inChampSelect) return;
  ensureApplyButton();
  unlockSkinCarousel();
  updateButtonState();

  // Load skins for chroma button
  const champId = lastChampionId || await getMyChampionId();
  if (champId) {
    await loadChampionSkins(champId);
    ensureChromaButton();
  }
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
  closeChromaPanel();
  document.querySelectorAll(`.${CHROMA_BTN_CLASS}`).forEach(el => el.remove());
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
