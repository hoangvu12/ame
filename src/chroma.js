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
import { setSelectedChroma, prefetchChroma } from './autoApply';

let activeChromaPanel = null;
let activeChromaButton = null;
let appliedSkinName = null;
let lastChampionId = null;
const chromaButtonDataMap = new WeakMap();

export function setLastChampionId(id) {
  lastChampionId = id;
}

export function getLastChampionId() {
  return lastChampionId;
}

export function setAppliedSkinName(name) {
  appliedSkinName = name;
}

export function getAppliedSkinName() {
  return appliedSkinName;
}

export function closeChromaPanel() {
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
  flyout.style.left = `${Math.round(left)}px`;
  flyout.style.top = `${Math.round(top)}px`;
  flyout.style.bottom = '';

  setTimeout(() => document.addEventListener('click', onClickOutsideChroma, true), 0);
}

async function selectChroma(skinData, chroma) {
  const championId = await getMyChampionId();
  if (!championId) return;

  const triggerButton = activeChromaButton;

  console.log('[ame] Chroma selected:', chroma.name, '| ID:', chroma.id, '| Base skin:', skinData.id);

  // Just select the chroma — auto-apply will handle applying after 10s stability
  setSelectedChroma(chroma.id, skinData.id);
  prefetchChroma(championId, chroma.id, skinData.id);

  closeChromaPanel();
  document.removeEventListener('click', onClickOutsideChroma, true);

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

export function ensureChromaButton() {
  const championSkins = getChampionSkins();
  const skinItems = document.querySelectorAll('.skin-selection-item');
  if (skinItems.length === 0) return;

  const selectedItem = document.querySelector('.skin-selection-item-selected, .skin-selection-item.selected');
  let centerItem = null;
  let centerReason = 'none';

  const debugKey = `debug-${championSkins?.length}`;
  if (championSkins && debugKey !== ensureChromaButton._lastDebug) {
    ensureChromaButton._lastDebug = debugKey;
    console.log('[ame][chroma] All skins:', championSkins.map(s => `${s.id} (num:${s.num}) :"${s.name}" chromas:${s.chromas?.length || 0}`).join(' | '));
  }

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
