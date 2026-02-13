import { BUTTON_ID, CUSTOM_SKINS_BTN_ID, CONNECTION_BANNER_ID } from './constants';
import { readCurrentSkin } from './skin';
import { wsSend, isOverlayActive, isConnected, onConnection } from './websocket';
import { getAppliedSkinName, setAppliedSkinName, getSkinForced } from './state';
import { removeElement, el } from './dom';
import { openCustomSkinsModal } from './customSkins';
import { t } from './i18n';
import { createLogger } from './logger';

const logger = createLogger('ui');

const CUSTOM_SKINS_ICON = '<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2c5.523 0 10 4.477 10 10q-.002.975-.18 1.9c-.373 1.935-2.26 2.791-3.907 2.595l-.175-.025l-1.74-.29a1.29 1.29 0 0 0-1.124.36c-.37.37-.547.879-.298 1.376c.423.846.429 1.812.055 2.603C14.131 21.58 13.11 22 12 22C6.477 22 2 17.523 2 12S6.477 2 12 2m-4.5 9a1.5 1.5 0 1 0 0 3a1.5 1.5 0 0 0 0-3m7-4a1.5 1.5 0 1 0 0 3a1.5 1.5 0 0 0 0-3m-5 0a1.5 1.5 0 1 0 0 3a1.5 1.5 0 0 0 0-3"/></svg>';

export function ensureApplyButton() {
  const skinName = readCurrentSkin();
  const existing = document.getElementById(CUSTOM_SKINS_BTN_ID);

  if (existing) {
    existing.style.display = skinName ? '' : 'none';
    return;
  }

  if (!skinName) return;

  let parent;
  const teamBoost = document.querySelector('.team-boost');
  if (teamBoost && teamBoost.parentElement) {
    parent = teamBoost.parentElement;
  } else {
    const carousel = document.querySelector('.skin-selection-carousel-container');
    parent = carousel?.parentElement;
  }
  if (!parent) return;

  const btn = el('lol-uikit-flat-button', {
    class: 'ame-custom-skins-icon-btn',
    onClick: openCustomSkinsModal,
  });
  btn.id = CUSTOM_SKINS_BTN_ID;
  btn.innerHTML = '<div class="ame-csb-content">' + CUSTOM_SKINS_ICON + '<span>' + t('custom_skins.button') + '</span></div>';
  if (teamBoost) btn.style.top = '362px';
  parent.appendChild(btn);
}

export function removeApplyButton() {
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
