import { BUTTON_ID, SKIN_SELECTORS } from './constants';
import { getMyChampionId, loadChampionSkins } from './api';
import { readCurrentSkin, findSkinByName } from './skin';
import { wsSendApply } from './websocket';
import { setAppliedSkinName, getAppliedSkinName } from './chroma';
import { getSelectedChroma } from './autoApply';

export function findSkinNameElement() {
  for (const selector of SKIN_SELECTORS) {
    const nodes = document.querySelectorAll(selector);
    for (const node of nodes) {
      if (node.textContent.trim() && node.offsetParent !== null) return node;
    }
  }
  return null;
}

export function ensureApplyButton() {
  if (document.getElementById(BUTTON_ID)) return;
  const container = document.querySelector('.toggle-ability-previews-button-container');
  if (!container) return;

  container.style.justifyContent = 'center';
  container.style.alignItems = 'center';
  container.style.gap = '20px';
  container.querySelectorAll('.framing-line').forEach(line => {
    line.style.display = 'none';
  });

  const btn = document.createElement('lol-uikit-flat-button');
  btn.id = BUTTON_ID;
  btn.textContent = 'Apply Skin';
  btn.classList.add('toggle-ability-previews-button');
  btn.addEventListener('click', onApplyClick);

  container.appendChild(btn);
}

export function removeApplyButton() {
  const btn = document.getElementById(BUTTON_ID);
  if (btn) btn.remove();
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

  // Apply with chroma if one is selected, otherwise base skin
  const chroma = getSelectedChroma();
  if (chroma) {
    console.log('[ame] Apply clicked (chroma):', skinName, '| Chroma ID:', chroma.id, '| Base:', chroma.baseSkinId, '| Champion ID:', championId);
    wsSendApply({ type: 'apply', championId, skinId: chroma.id, baseSkinId: chroma.baseSkinId });
  } else {
    console.log('[ame] Apply clicked:', skinName, '| Skin ID:', skin.id, '| Champion ID:', championId);
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
