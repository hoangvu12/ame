import { SWIFTPLAY_BUTTON_ID } from './constants';
import { getChampionIdFromLobbyDOM, loadChampionSkins, getChampionName } from './api';
import { readCurrentSkin, findSkinByName, isDefaultSkin } from './skin';
import { wsSend, wsSendApply, isOverlayActive } from './websocket';
import { toastError } from './toast';
import { getAppliedSkinName, setAppliedSkinName, getSelectedChroma } from './state';
import { ensureElement, removeElement } from './dom';
import { createButton } from './components';

const CONTAINER_SELECTOR = '.quick-play-skin-select-component .top-part';

export function ensureSwiftplayButton() {
  ensureElement(SWIFTPLAY_BUTTON_ID, CONTAINER_SELECTOR, (container) => {
    container.style.display = 'flex';
    container.style.justifyContent = 'center';
    container.style.alignItems = 'center';
    container.style.padding = '8px 0';
    return createButton('Apply Skin', {
      class: 'toggle-ability-previews-button',
      onClick: onSwiftplayApplyClick,
    });
  });
}

export function removeSwiftplayButton() {
  removeElement(SWIFTPLAY_BUTTON_ID);
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
  const current = readCurrentSkin();
  if (!current) return;

  if (ownership === null) {
    setButtonState('Loading...', true);
    return;
  }

  if (ownership) {
    if (isOverlayActive()) {
      wsSend({ type: 'cleanup' });
    }
    if (getAppliedSkinName()) {
      setAppliedSkinName(null);
    }
    setButtonState('Owned', true);
    return;
  }

  const appliedSkinName = getAppliedSkinName();
  if (appliedSkinName && current === appliedSkinName) {
    setButtonState('Applied', true);
  } else {
    setButtonState('Apply Skin', false);
  }
}

async function onSwiftplayApplyClick() {
  const skinName = readCurrentSkin();
  if (!skinName) return;

  const championId = await getChampionIdFromLobbyDOM();
  if (!championId) {
    toastError('Could not identify champion');
    return;
  }

  const skins = await loadChampionSkins(championId);
  if (!skins) {
    toastError('Could not load skin data');
    return;
  }

  const skin = findSkinByName(skins, skinName);
  if (!skin) {
    toastError('Skin not found in game data');
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
  setButtonState('Applied', true);
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
