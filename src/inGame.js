import { IN_GAME_CONTAINER_ID } from './constants';
import { wsSend, wsSendApply, getLastApplyPayload, isOverlayActive, setOverlayActive } from './websocket';
import { getChampionSkins } from './api';
import { isDefaultSkin } from './skin';
import { removeElement } from './dom';
import { createButton, createDropdown } from './components';

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
  if (!skins) return null;

  const nonDefault = skins.filter(s => !isDefaultSkin(s));
  if (nonDefault.length === 0) return null;

  const payload = getLastApplyPayload();

  const dropdown = createDropdown(
    nonDefault.map(skin => ({
      text: skin.name,
      attrs: {
        'data-skin-id': String(skin.id),
        'data-skin-name': skin.name,
      },
      selected: payload && (
        String(skin.id) === String(payload.skinId) ||
        String(skin.id) === String(payload.baseSkinId)
      ),
    })),
    {
      class: 'ame-ingame-dropdown',
      onChange: (selected) => {
        const skinId = parseInt(selected.getAttribute('data-skin-id'), 10);
        if (isNaN(skinId)) return;

        const lastPayload = getLastApplyPayload();
        const championId = lastPayload?.championId;
        if (!championId) return;

        wsSendApply({ type: 'apply', championId, skinId });
        updateInGameStatus();
      },
    }
  );

  return dropdown;
}

export function ensureInGameUI() {
  if (document.getElementById(IN_GAME_CONTAINER_ID)) return;
  const parent = findReconnectContainer();
  if (!parent) return;

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
        const skinId = selected ? parseInt(selected.getAttribute('data-skin-id'), 10) : null;
        const lastPayload = getLastApplyPayload();
        const championId = lastPayload?.championId;

        if (championId && skinId && !isNaN(skinId)) {
          wsSendApply({ type: 'apply', championId, skinId });
          updateInGameStatus();
        }
      }
    },
  });

  container.appendChild(btn);

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
    btn.textContent = 'Remove Skin';
    btn.removeAttribute('disabled');
    btn.style.display = '';
    if (dropdown) dropdown.style.display = 'none';
  } else if (hasSkins) {
    btn.textContent = 'Apply Skin';
    btn.removeAttribute('disabled');
    btn.style.display = '';
    dropdown.style.display = '';
  } else {
    btn.style.display = 'none';
  }
}
