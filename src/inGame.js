import { IN_GAME_CONTAINER_ID } from './constants';
import { wsSend, wsSendApply, wsSendUnstuck, getLastApplyPayload, isOverlayActive, setOverlayActive } from './websocket';
import { openCustomSkinsModal } from './customSkins';
import { getChampionSkins, loadChampionSkins } from './api';
import { isDefaultSkin } from './skin';
import { removeElement } from './dom';
import { createButton, createDropdown } from './components';
import { t } from './i18n';

const RECONNECT_SELECTORS = [
  '.game-in-progress-container',
  '.rcp-fe-lol-game-in-progress',
  '.reconnect-container',
];

function findReconnectContainer() {
  for (const sel of RECONNECT_SELECTORS) {
    const el = document.querySelector(sel);
    if (el) return el;
  }
  return null;
}

function buildDropdown() {
  const skins = getChampionSkins();
  if (!skins || skins.length === 0) return null;

  const payload = getLastApplyPayload();
  const defaultSkin = skins.find(s => isDefaultSkin(s));
  const nonDefault = skins.filter(s => !isDefaultSkin(s));

  if (nonDefault.length === 0) return null;

  const options = [];

  // Add default skin option first
  if (defaultSkin) {
    options.push({
      text: defaultSkin.name,
      attrs: {
        'data-skin-id': '0',
        'data-skin-name': defaultSkin.name,
        'data-is-default': 'true',
      },
      selected: !payload || !payload.skinId || payload.skinId === 0,
    });
  }

  // Add non-default skins
  for (const skin of nonDefault) {
    options.push({
      text: skin.name,
      attrs: {
        'data-skin-id': String(skin.id),
        'data-skin-name': skin.name,
      },
      selected: payload && (
        String(skin.id) === String(payload.skinId) ||
        String(skin.id) === String(payload.baseSkinId)
      ),
    });
  }

  if (options.length === 0) return null;

  const dropdown = createDropdown(options, {
    class: 'ame-ingame-dropdown',
    onChange: (selected) => {
      const isDefault = selected.getAttribute('data-is-default') === 'true';
      const lastPayload = getLastApplyPayload();
      const championId = lastPayload?.championId;
      const championName = lastPayload?.championName || '';
      if (!championId) return;

      if (isDefault) {
        wsSendApply({ type: 'apply', championId, skinId: 0, championName, skinName: '' });
      } else {
        const skinId = parseInt(selected.getAttribute('data-skin-id'), 10);
        if (isNaN(skinId)) return;
        const skinName = selected.getAttribute('data-skin-name') || '';
        wsSendApply({ type: 'apply', championId, skinId, championName, skinName });
      }
      updateInGameStatus();
    },
  });

  return dropdown;
}

export function ensureInGameUI() {
  if (document.getElementById(IN_GAME_CONTAINER_ID)) return;
  const parent = findReconnectContainer();
  if (!parent) return;

  // Ensure skins are loaded for the current champion
  const payload = getLastApplyPayload();
  if (payload?.championId && !getChampionSkins()) {
    loadChampionSkins(payload.championId);
    return; // Wait for next poll cycle when skins are cached
  }

  const container = document.createElement('div');
  container.id = IN_GAME_CONTAINER_ID;

  const dropdown = buildDropdown();
  if (dropdown) container.appendChild(dropdown);

  const btn = createButton('', {
    class: 'ame-ingame-action',
    onClick: () => {
      if (isOverlayActive()) {
        wsSend({ type: 'cleanup' });
        setOverlayActive(false);
        updateInGameStatus();
      } else {
        const dd = container.querySelector('.ame-ingame-dropdown');
        const selected = dd?.querySelector('lol-uikit-dropdown-option[selected]');
        const isDefault = selected?.getAttribute('data-is-default') === 'true';
        const lastPayload = getLastApplyPayload();
        const championId = lastPayload?.championId;
        const championName = lastPayload?.championName || '';

        if (championId) {
          if (isDefault || !selected) {
            wsSendApply({ type: 'apply', championId, skinId: 0, championName, skinName: '' });
          } else {
            const skinId = parseInt(selected.getAttribute('data-skin-id'), 10);
            const skinName = selected?.getAttribute('data-skin-name') || '';
            if (skinId && !isNaN(skinId)) {
              wsSendApply({ type: 'apply', championId, skinId, championName, skinName });
            }
          }
          updateInGameStatus();
        }
      }
    },
  });

  container.appendChild(btn);

  const customModsBtn = createButton(t('custom_skins.button'), {
    class: 'ame-ingame-custom-mods',
    onClick: () => openCustomSkinsModal(),
  });
  container.appendChild(customModsBtn);

  const unstuckBtn = createButton(t('ui.im_stuck'), {
    class: 'ame-ingame-unstuck',
    onClick: () => {
      wsSendUnstuck();
    },
  });
  container.appendChild(unstuckBtn);

  if (!dropdown) btn.style.display = 'none';

  const pos = window.getComputedStyle(parent).position;
  if (pos === 'static' || !pos) parent.style.position = 'relative';

  parent.appendChild(container);
  updateInGameStatus();
}

export function removeInGameUI() {
  removeElement(IN_GAME_CONTAINER_ID);
}

export function updateInGameStatus() {
  const container = document.getElementById(IN_GAME_CONTAINER_ID);
  if (!container) return;

  const btn = container.querySelector('.ame-ingame-action');
  const dropdown = container.querySelector('.ame-ingame-dropdown');
  if (!btn) return;

  const active = isOverlayActive();
  const hasSkins = !!dropdown;

  if (active) {
    btn.textContent = t('ui.remove_skin');
    btn.removeAttribute('disabled');
    btn.style.display = '';
    if (dropdown) dropdown.style.display = 'none';
  } else if (hasSkins) {
    btn.textContent = t('ui.apply_skin');
    btn.removeAttribute('disabled');
    btn.style.display = '';
    dropdown.style.display = '';
  } else {
    btn.style.display = 'none';
  }
}
