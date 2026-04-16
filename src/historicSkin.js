import { getSkinKeyFromItem } from './skin';
import { el } from './dom';
import { createLogger } from './logger';

const logger = createLogger('historic');

const STORAGE_PREFIX = 'ame:historic:';
const MARKER_CLASS = 'ame-historic-marker';

const cache = {};

function loadFromStorage(championId) {
  try {
    const raw = localStorage.getItem(STORAGE_PREFIX + championId);
    if (!raw) return null;
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

export function recordHistoricSkin(championId, skinNum, skinName) {
  if (!championId || skinNum == null) return;
  const entry = { skinNum, skinName };
  cache[championId] = entry;
  try {
    localStorage.setItem(STORAGE_PREFIX + championId, JSON.stringify(entry));
    logger.log(`recorded: champ=${championId} skinNum=${skinNum} name="${skinName}"`);
  } catch { /* quota exceeded */ }
}

export function getHistoricSkin(championId) {
  if (!championId) return null;
  if (championId in cache) return cache[championId];
  const entry = loadFromStorage(championId);
  cache[championId] = entry;
  return entry;
}

export function markHistoricSkin(championId) {
  const historic = getHistoricSkin(championId);

  if (!historic) {
    clearHistoricMarkers();
    return;
  }

  // Check if the marker is already on the correct item
  const existing = document.querySelector(`.${MARKER_CLASS}`);
  if (existing) {
    const parent = existing.closest('.skin-selection-item');
    if (parent && getSkinKeyFromItem(parent) === historic.skinNum) return;
    existing.remove();
  }

  const items = document.querySelectorAll('.skin-selection-item');
  for (const item of items) {
    if (getSkinKeyFromItem(item) === historic.skinNum) {
      const badge = el('div', { class: MARKER_CLASS, title: historic.skinName });
      item.appendChild(badge);
      return;
    }
  }
}

export function clearHistoricMarkers() {
  document.querySelectorAll(`.${MARKER_CLASS}`).forEach(node => node.remove());
}
