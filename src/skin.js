import { SKIN_SELECTORS } from './constants';

export function slugify(str) {
  return str
    .normalize('NFKD')
    .toLowerCase()
    .replace(/[^\p{L}\p{N}]/gu, '');
}

export function findSkinByName(skins, name) {
  const exact = skins.find(s => s.name === name);
  if (exact) return exact;
  const slug = slugify(name);
  return skins.find(s => slugify(s.name) === slug) || null;
}

export function readCurrentSkin() {
  for (const selector of SKIN_SELECTORS) {
    const nodes = document.querySelectorAll(selector);
    let candidate = null;
    for (const node of nodes) {
      const name = node.textContent.trim();
      if (name && node.offsetParent !== null) candidate = name;
    }
    if (candidate) return candidate;
  }
  return null;
}

export function getSkinNameFromItem(item) {
  if (!item) return null;

  const attrNames = [
    'data-skin-name',
    'data-skin',
    'data-name',
    'title',
    'aria-label',
  ];
  for (const attr of attrNames) {
    const val = item.getAttribute(attr);
    if (val && val.trim()) return val.trim();
  }

  const nameEl = item.querySelector('.skin-name, .skin-name-text, .skin-selection-item-name');
  if (nameEl && nameEl.textContent.trim()) return nameEl.textContent.trim();

  return null;
}

export function isItemVisible(item) {
  if (!item) return false;
  const style = window.getComputedStyle(item);
  if (style.display === 'none' || style.visibility === 'hidden') return false;
  const opacity = parseFloat(style.opacity || '1');
  return opacity > 0.05;
}

export function getSkinOffset(skinItem) {
  const match = skinItem.className.match(/skin-carousel-offset-(-?\d+)/);
  if (match) return parseInt(match[1]);
  let parent = skinItem.parentElement;
  for (let i = 0; i < 3 && parent; i++) {
    const pm = parent.className.match(/skin-carousel-offset-(-?\d+)/);
    if (pm) return parseInt(pm[1]);
    parent = parent.parentElement;
  }
  return null;
}

export function getSkinKeyFromItem(item) {
  if (!item) return null;
  const thumb = item.querySelector('.skin-selection-thumbnail');
  if (!thumb) return null;
  const bg = thumb.style.backgroundImage || window.getComputedStyle(thumb).backgroundImage || '';
  if (!bg) return null;
  const match = bg.match(/Skins\/(Base|Skin(\d+))\//i);
  if (!match) return null;
  if (match[1].toLowerCase() === 'base') return 0;
  const id = parseInt(match[2], 10);
  return Number.isFinite(id) ? id : null;
}

export function isDefaultSkin(skin) {
  if (!skin) return false;
  return skin.isBase === true || skin.num === 0 || skin.id % 1000 === 0;
}

export function findSkinByCarouselKey(skins, key) {
  if (!skins || key === null || key === undefined) return null;
  let skin = skins.find(s => s.num === key);
  if (skin) return skin;
  skin = skins.find(s => typeof s.id === 'number' && (s.id % 1000) === key);
  return skin || null;
}

export function getChromaData(skins, currentSkinName) {
  if (!skins || !currentSkinName) return null;
  const skin = findSkinByName(skins, currentSkinName);
  console.log('[ame][chroma] findSkinByName("' + currentSkinName + '") =>', skin ? `id:${skin.id} name:"${skin.name}" chromas:${skin.chromas?.length}` : 'NOT FOUND',
    '| slugified input:', slugify(currentSkinName),
    '| slugified names:', skins.map(s => slugify(s.name)).join(', '));
  if (!skin || !skin.chromas || skin.chromas.length === 0) return null;
  return { skin, chromas: skin.chromas };
}
