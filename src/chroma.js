import { CHROMA_BTN_CLASS, CHROMA_PANEL_ID } from './constants';
import {
  getSkinOffset,
  getSkinKeyFromItem,
  findSkinByCarouselKey,
  getSkinNameFromItem,
  readCurrentSkin,
  getChromaData,
  isItemVisible
} from './skin';
import { getMyChampionId, getChampionSkins } from './api';
import { getLastChampionId } from './state';
import { onChromaSelected, prefetchChroma } from './autoApply';
import { el } from './dom';

let activeChromaPanel = null;
let activeChromaButton = null;
const chromaButtonDataMap = new WeakMap();

export function closeChromaPanel() {
  if (activeChromaPanel) activeChromaPanel.remove();
  activeChromaPanel = null;
  activeChromaButton = null;
}

// --- Click-outside handler ---

function onClickOutsideChroma(e) {
  const panel = activeChromaPanel || document.getElementById(CHROMA_PANEL_ID);
  const btn = document.querySelector(`.${CHROMA_BTN_CLASS}`);
  if (panel && !panel.contains(e.target) && (!btn || !btn.contains(e.target))) {
    closeChromaPanel();
    document.removeEventListener('click', onClickOutsideChroma, true);
  }
}

// --- Button creation ---

function createChromaButton() {
  return el('div', { class: `${CHROMA_BTN_CLASS} chroma-button chroma-selection uikit-framed-icon ember-view` },
    el('div', { class: 'outer-mask interactive' },
      el('div', { class: 'frame-color' },
        el('div', { class: 'content' }),
        el('div', { class: 'inner-mask inner-shadow' })
      )
    )
  );
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

// --- Panel creation ---

function getChromaImageUrl(champId, chroma) {
  if (!chroma || !champId || !chroma.id) return '';
  return `/lol-game-data/assets/v1/champion-chroma-images/${champId}/${chroma.id}.png`;
}

function getChromaPreviewPath(chroma) {
  if (!chroma) return '';
  return chroma.chromaPreviewPath || chroma.chromaPath || chroma.chromaPreview
    || chroma.imagePath || chroma.splashPath || chroma.tilePath || '';
}

function chromaBackground(colors) {
  if (colors.length >= 2) return `linear-gradient(135deg, ${colors[0]} 50%, ${colors[1]} 50%)`;
  if (colors.length === 1) return colors[0];
  return '#27211C';
}

function createChromaPanel(skinData, chromas, buttonEl, championId) {
  closeChromaPanel();

  const carousel = buttonEl.closest('.skin-selection-carousel') || document.querySelector('.skin-selection-carousel');
  const champId = championId || getLastChampionId();

  const flyout = el('lol-uikit-flyout-frame', {
    id: CHROMA_PANEL_ID,
    class: 'flyout',
    orientation: 'top',
    show: 'true',
    style: { position: 'absolute', overflow: 'visible', zIndex: '10000' },
  });

  (carousel || document.body).appendChild(flyout);

  const chromaImage = el('div', { class: 'chroma-information-image' });
  if (chromas.length > 0) {
    const imgUrl = getChromaImageUrl(champId, chromas[0]) || getChromaPreviewPath(chromas[0]);
    if (imgUrl) chromaImage.style.backgroundImage = `url('${imgUrl}')`;
  }

  const disabledNote = el('div', { class: 'child-skin-disabled-notification' });
  const skinName = el('div', { class: 'child-skin-name' }, skinData.name, disabledNote);

  const chromaInfo = el('div', {
    class: 'chroma-information',
    style: { backgroundImage: "url('lol-game-data/assets/content/src/LeagueClient/GameModeAssets/Classic_SRU/img/champ-select-flyout-background.jpg')" },
  }, chromaImage, skinName);

  const scrollable = el('lol-uikit-scrollable', {
    class: 'chroma-selection',
    'overflow-masks': 'enabled',
    'scrolled-bottom': 'false',
    'scrolled-top': 'true',
  });

  for (let i = 0; i < chromas.length; i++) {
    const chroma = chromas[i];
    const colors = chroma.colors || [];

    const contents = el('div', {
      class: 'contents',
      style: { background: chromaBackground(colors) },
    });

    const btn = el('div', { class: i === 0 ? 'chroma-skin-button selected' : 'chroma-skin-button' }, contents);
    scrollable.appendChild(el('div', { class: 'ember-view' }, btn));

    btn.addEventListener('click', () => {
      selectChroma(skinData, chroma);
      scrollable.querySelectorAll('.chroma-skin-button').forEach(b => b.classList.remove('selected'));
      btn.classList.add('selected');
    });
    btn.addEventListener('mouseenter', () => {
      const preview = getChromaImageUrl(champId, chroma) || getChromaPreviewPath(chroma);
      if (preview) chromaImage.style.backgroundImage = `url('${preview}')`;
      if (skinName.firstChild) {
        skinName.firstChild.nodeValue = chroma.name || skinData.name;
      } else {
        skinName.textContent = chroma.name || skinData.name;
        skinName.appendChild(disabledNote);
      }
    });
  }

  const modal = el('div', { class: 'champ-select-chroma-modal chroma-view ember-view' },
    chromaInfo, scrollable
  );

  const flyoutContent = el('lc-flyout-content', { dataset: { ameChroma: '1' } }, modal);
  flyout.querySelectorAll('lc-flyout-content[data-ame-chroma="1"]').forEach(old => old.remove());
  flyout.appendChild(flyoutContent);

  activeChromaPanel = flyout;
  activeChromaButton = buttonEl;

  // Position the flyout above the button
  const panelContainer = carousel || document.body;
  const containerRect = panelContainer.getBoundingClientRect();
  const btnRect = buttonEl.getBoundingClientRect();
  const modalRect = modal.getBoundingClientRect();
  const width = modalRect.width || 305;
  const height = modalRect.height || 340;
  flyout.style.left = `${Math.round(btnRect.left - containerRect.left + btnRect.width / 2 - width / 2)}px`;
  flyout.style.top = `${Math.round(btnRect.top - containerRect.top - height - 8)}px`;
  flyout.style.bottom = '';

  setTimeout(() => document.addEventListener('click', onClickOutsideChroma, true), 0);
}

async function selectChroma(skinData, chroma) {
  const championId = await getMyChampionId();
  if (!championId) return;

  const triggerButton = activeChromaButton;

  onChromaSelected(chroma.id, skinData.id);
  prefetchChroma(championId, chroma.id, skinData.id);

  closeChromaPanel();
  document.removeEventListener('click', onClickOutsideChroma, true);

  // Update button swatch color to reflect selected chroma
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

// --- ensureChromaButton: broken into focused helpers ---

function findCenterItem(skinItems) {
  const selectedItem = document.querySelector('.skin-selection-item-selected, .skin-selection-item.selected');

  let closestItem = null;
  let closestDelta = Number.POSITIVE_INFINITY;

  for (const item of skinItems) {
    const offset = getSkinOffset(item);
    if (offset !== null && offset !== undefined) {
      const delta = Math.abs(offset);
      if (delta < closestDelta) {
        closestDelta = delta;
        closestItem = item;
      }
    }
  }

  const offset0Item = Array.from(skinItems).find(item => isItemVisible(item) && getSkinOffset(item) === 0);

  if (selectedItem && isItemVisible(selectedItem)) return selectedItem;
  if (offset0Item) return offset0Item;
  return closestItem;
}

function resolveChromaData(centerItem, championSkins) {
  const skinKeyFromItem = getSkinKeyFromItem(centerItem);
  let skinName = getSkinNameFromItem(centerItem) || readCurrentSkin();

  // Try matching by carousel key first (most reliable)
  if (skinKeyFromItem !== null && championSkins) {
    const skinByKey = findSkinByCarouselKey(championSkins, skinKeyFromItem);
    if (skinByKey?.chromas?.length > 0) {
      return { skin: skinByKey, chromas: skinByKey.chromas };
    }
  }

  // Fall back to name-based matching
  let chromaData = (skinName && championSkins) ? getChromaData(championSkins, skinName) : null;

  if (!chromaData) {
    const altName = readCurrentSkin();
    if (altName && altName !== skinName) {
      chromaData = getChromaData(championSkins, altName);
    }
  }

  return chromaData;
}

function syncChromaButtons(skinItems, championSkins) {
  const champId = getLastChampionId();

  for (const item of skinItems) {
    const key = getSkinKeyFromItem(item);
    let data = null;

    if (key !== null && championSkins) {
      const skinByKey = findSkinByCarouselKey(championSkins, key);
      if (skinByKey?.chromas?.length > 0) {
        data = { skin: skinByKey, chromas: skinByKey.chromas, championId: champId };
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

    const chromaBtn = getOrCreateChromaButton(thumb, data);
    if (!chromaBtn) continue;
    chromaBtn.classList.remove('hidden');
    Object.assign(chromaBtn.style, {
      position: 'absolute',
      right: '6px',
      bottom: '6px',
      left: '',
      top: '',
      zIndex: '50',
      pointerEvents: 'auto',
    });
  }
}

export function ensureChromaButton() {
  const championSkins = getChampionSkins();
  const skinItems = document.querySelectorAll('.skin-selection-item');
  if (skinItems.length === 0) return;

  const centerItem = findCenterItem(skinItems);
  const chromaData = centerItem ? resolveChromaData(centerItem, championSkins) : null;

  if (!centerItem || !chromaData) {
    closeChromaPanel();
  }

  syncChromaButtons(skinItems, championSkins);
}
