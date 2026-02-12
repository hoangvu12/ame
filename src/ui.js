import { BUTTON_ID, CUSTOM_SKINS_BTN_ID, CONNECTION_BANNER_ID } from './constants';
import { getMyChampionId, loadChampionSkins, getChampionName, forceDefaultSkin } from './api';
import { readCurrentSkin, findSkinByName, isDefaultSkin } from './skin';
import { wsSend, wsSendApply, isOverlayActive, isConnected, onConnection } from './websocket';
import { toastError } from './toast';
import { getAppliedSkinName, setAppliedSkinName, getSelectedChroma, getSkinForced, setSkinForced } from './state';
import { ensureElement, removeElement, el } from './dom';
import { createButton } from './components';
import { notifySkinChange } from './roomParty';
import { openCustomSkinsModal } from './customSkins';
import { t } from './i18n';
import { createLogger } from './logger';

const logger = createLogger('ui');

export function ensureApplyButton() {
  ensureElement(BUTTON_ID, '.toggle-ability-previews-button-container', (container) => {
    container.style.justifyContent = 'center';
    container.style.alignItems = 'center';
    container.style.gap = '20px';
    container.querySelectorAll('.framing-line').forEach(line => {
      line.style.display = 'none';
    });
    return createButton(t('ui.apply_skin'), {
      class: 'toggle-ability-previews-button',
      onClick: onApplyClick,
    });
  });
  ensureElement(CUSTOM_SKINS_BTN_ID, '.toggle-ability-previews-button-container', () => {
    return createButton(t('custom_skins.button'), {
      class: 'toggle-ability-previews-button',
      onClick: openCustomSkinsModal,
    });
  });
}

export function removeApplyButton() {
  removeElement(BUTTON_ID);
  removeElement(CUSTOM_SKINS_BTN_ID);
}

let connectionHooked = false;

export function ensureConnectionBanner() {
  const existing = document.getElementById(CONNECTION_BANNER_ID);
  if (existing) return existing;

  const banner = el('div', { class: 'ame-connection-banner' },
    el('span', { class: 'ame-connection-dot' }),
    el('span', { class: 'ame-connection-text' }, t('ui.connection_banner'))
  );
  banner.id = CONNECTION_BANNER_ID;

  const teamBoost = document.querySelector('.team-boost');
  if (teamBoost && teamBoost.parentElement) {
    teamBoost.parentElement.insertBefore(banner, teamBoost);
  } else {
    const container = document.querySelector('.skin-selection-carousel-container');
    if (!container) return null;
    container.prepend(banner);
  }

  return banner;
}

function setConnectionBanner(connected) {
  const banner = ensureConnectionBanner();
  if (!banner) return;
  banner.style.display = connected ? 'none' : 'flex';
}

export function updateConnectionBanner() {
  setConnectionBanner(isConnected());
}

export function initConnectionStatus() {
  if (connectionHooked) return;
  connectionHooked = true;
  onConnection((connected) => {
    setConnectionBanner(connected);
    if (!connected) {
      setButtonState(t('ui.open_ame'), true);
    }
  });
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
    toastError(t('errors.pick_champion_first'));
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

  logger.log(` forceDefaultSkin(${championId}) calling...`);
  const forced = await forceDefaultSkin(championId);
  logger.log(` forceDefaultSkin result: ${forced}`);
  if (!forced) {
    toastError(t('errors.could_not_set_default'));
    return;
  }
  setSkinForced(true);
  logger.log(' skinForced set to true');

  const champName = await getChampionName(championId);
  const chroma = getSelectedChroma();
  if (chroma) {
    wsSendApply({
      type: 'apply', championId, skinId: chroma.id, baseSkinId: chroma.baseSkinId,
      championName: champName, skinName: chroma.baseSkinName || skin.name, chromaName: chroma.chromaName,
    });
    notifySkinChange(championId, chroma.id, chroma.baseSkinId, champName, chroma.baseSkinName || skin.name, chroma.chromaName);
  } else {
    wsSendApply({ type: 'apply', championId, skinId: skin.id, championName: champName, skinName: skin.name });
    notifySkinChange(championId, skin.id, '', champName, skin.name, '');
  }
  setAppliedSkinName(skinName);
  setButtonState(t('ui.applied'), true);
}

export function updateButtonState(ownership) {
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
    const forced = getSkinForced();
    const overlay = isOverlayActive();
    logger.log(` updateButtonState: owned=true skin="${current}" skinForced=${forced} overlayActive=${overlay}`);
    if (forced) return;
    if (overlay) {
      logger.log(' updateButtonState: sending cleanup (owned + overlay active)');
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
