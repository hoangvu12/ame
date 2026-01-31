import { BUTTON_ID } from './constants';
import { getMyChampionId, loadChampionSkins } from './api';
import { readCurrentSkin, findSkinByName, isDefaultSkin } from './skin';
import { wsSendApply } from './websocket';
import { toastError } from './toast';
import { getAppliedSkinName, setAppliedSkinName, getSelectedChroma } from './state';
import { ensureElement, removeElement } from './dom';
import { createButton } from './components';

export function ensureApplyButton() {
  ensureElement(BUTTON_ID, '.toggle-ability-previews-button-container', (container) => {
    container.style.justifyContent = 'center';
    container.style.alignItems = 'center';
    container.style.gap = '20px';
    container.querySelectorAll('.framing-line').forEach(line => {
      line.style.display = 'none';
    });
    return createButton('Apply Skin', {
      class: 'toggle-ability-previews-button',
      onClick: onApplyClick,
    });
  });
}

export function removeApplyButton() {
  removeElement(BUTTON_ID);
}

export function setButtonState(text, disabled) {
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
    toastError('Pick a champion first');
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

  const chroma = getSelectedChroma();
  if (chroma) {
    wsSendApply({ type: 'apply', championId, skinId: chroma.id, baseSkinId: chroma.baseSkinId });
  } else {
    wsSendApply({ type: 'apply', championId, skinId: skin.id });
  }
  setAppliedSkinName(skinName);
  setButtonState('Applied', true);
}

export function updateButtonState() {
  const appliedSkinName = getAppliedSkinName();
  if (!appliedSkinName) return;
  const current = readCurrentSkin();
  if (!current) return;
  if (current === appliedSkinName) {
    setButtonState('Applied', true);
  } else {
    setButtonState('Apply Skin', false);
  }
}
