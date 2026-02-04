import { SWIFTPLAY_BUTTON_ID, CONNECTION_BANNER_ID } from './constants';
import { getChampionIdFromLobbyDOM, loadChampionSkins, getChampionName } from './api';
import { readCurrentSkin, findSkinByName, isDefaultSkin } from './skin';
import { wsSend, wsSendApply, isOverlayActive, isConnected } from './websocket';
import { toastError } from './toast';
import { getAppliedSkinName, setAppliedSkinName, getSelectedChroma } from './state';
import { ensureElement, removeElement, el } from './dom';
import { createButton } from './components';
import { t } from './i18n';

const CONTAINER_SELECTOR = '.quick-play-skin-select-component .top-part';
const SWIFTPLAY_BANNER_ID = `${CONNECTION_BANNER_ID}-swiftplay`;

export function ensureSwiftplayButton() {
  ensureElement(SWIFTPLAY_BUTTON_ID, CONTAINER_SELECTOR, (container) => {
    container.style.display = 'flex';
    container.style.justifyContent = 'center';
    container.style.alignItems = 'center';
    container.style.padding = '8px 0';
    return createButton(t('ui.apply_skin'), {
      class: 'toggle-ability-previews-button',
      onClick: onSwiftplayApplyClick,
    });
  });
}

export function removeSwiftplayButton() {
  removeElement(SWIFTPLAY_BUTTON_ID);
}

export function ensureSwiftplayConnectionBanner() {
  const container = document.querySelector(CONTAINER_SELECTOR);
  if (!container) return null;
  const existing = document.getElementById(SWIFTPLAY_BANNER_ID);
  if (existing) return existing;

  const banner = el('div', { class: 'ame-connection-banner' },
    el('span', { class: 'ame-connection-dot' }),
    el('span', { class: 'ame-connection-text' }, t('ui.connection_banner'))
  );
  banner.id = SWIFTPLAY_BANNER_ID;
  container.prepend(banner);
  return banner;
}

export function updateSwiftplayConnectionBanner() {
  const banner = ensureSwiftplayConnectionBanner();
  if (!banner) return;
  banner.style.display = isConnected() ? 'none' : 'flex';
}

function setButtonState(text, disabled) {
  const btn = document.getElementById(SWIFTPLAY_BUTTON_ID);
  if (!btn) return;
  btn.textContent = text;
  if (disabled) {
    btn.setAttribute('disabled', '');
  } else {
    btn.removeAttribute('disabled');
  }
}

export function updateSwiftplayButtonState(ownership) {
  if (!isConnected()) {
    setButtonState(t('ui.open_ame'), true);
    return;
  }

  const current = readCurrentSkin();
  if (!current) return;

  if (ownership === null) {
    setButtonState(t('ui.loading'), true);
    return;
  }

  if (ownership) {
    if (isOverlayActive()) {
      wsSend({ type: 'cleanup' });
    }
    if (getAppliedSkinName()) {
      setAppliedSkinName(null);
    }
    setButtonState(t('ui.owned'), true);
    return;
  }

  const appliedSkinName = getAppliedSkinName();
  if (appliedSkinName && current === appliedSkinName) {
    setButtonState(t('ui.applied'), true);
  } else {
    setButtonState(t('ui.apply_skin'), false);
  }
}

async function onSwiftplayApplyClick() {
  const skinName = readCurrentSkin();
  if (!skinName) return;

  const championId = await getChampionIdFromLobbyDOM();
  if (!championId) {
    toastError(t('errors.could_not_identify_champion'));
    return;
  }

  const skins = await loadChampionSkins(championId);
  if (!skins) {
    toastError(t('errors.could_not_load_skin_data'));
    return;
  }

  const skin = findSkinByName(skins, skinName);
  if (!skin) {
    toastError(t('errors.skin_not_found'));
    return;
  }

  if (isDefaultSkin(skin)) return;

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
  setButtonState(t('ui.applied'), true);
}

export function unlockSwiftplayCarousel() {
  document.querySelectorAll(
    '.quick-play-skin-select-component .unlock-skin-hit-area, ' +
    '.quick-play-skin-select-component .locked-state, ' +
    '.quick-play-skin-select-component .skin-purchase-button-container'
  ).forEach(el => {
    el.style.display = 'none';
    el.style.visibility = 'hidden';
    el.style.pointerEvents = 'none';
  });
}

export function isSwiftplaySkinPanelOpen() {
  return !!document.querySelector('.quick-play-skin-select-component');
}
