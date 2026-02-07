import { getChampionSkins, getChampionName } from './api';
import { isDefaultSkin, getSkinKeyFromItem } from './skin';
import { getRandomSkinMode } from './websocket';
import { createLogger } from './logger';

const logger = createLogger('random');

const RANKINGS_API = 'https://api.lolskinrankings.com/api/champions';

let lastRandomChampionId = null;
let chosenSkinNum = null; // persisted pick so retries click the same skin
let done = false;         // true once the click succeeded
let picking = false;      // guard against concurrent picks

export function resetRandomSkin() {
  lastRandomChampionId = null;
  chosenSkinNum = null;
  done = false;
  picking = false;
}

async function fetchSkinRankings(championName) {
  try {
    const res = await fetch(`${RANKINGS_API}/${encodeURIComponent(championName)}/skins`);
    if (!res.ok) return null;
    const json = await res.json();
    if (!json.success || !Array.isArray(json.data)) return null;
    return json.data;
  } catch {
    return null;
  }
}

function clickCarouselSkin(skinNum) {
  const items = document.querySelectorAll('.skin-selection-item');
  for (const item of items) {
    const key = getSkinKeyFromItem(item);
    if (key === skinNum) {
      logger.log(`clicking carousel item for skinNum=${skinNum}`);
      const thumb = item.querySelector('.skin-selection-thumbnail') || item;
      thumb.dispatchEvent(new PointerEvent('pointerdown', { bubbles: true, cancelable: true }));
      thumb.dispatchEvent(new PointerEvent('pointerup', { bubbles: true, cancelable: true }));
      thumb.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true }));
      return true;
    }
  }
  return false;
}

function pickRandom(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

export async function triggerRandomSkin(championId) {
  const mode = getRandomSkinMode();
  if (!mode) return;
  if (done && lastRandomChampionId === championId) return;
  logger.log(`trigger: champId=${championId} mode=${mode} done=${done} chosenSkinNum=${chosenSkinNum}`);

  // Champion changed — reset pick
  if (lastRandomChampionId !== championId) {
    chosenSkinNum = null;
    done = false;
    lastRandomChampionId = championId;
  }

  // If we already picked a skin, just retry the click
  if (chosenSkinNum !== null) {
    if (clickCarouselSkin(chosenSkinNum)) {
      done = true;
    }
    return;
  }

  // Guard against concurrent async picks (top3 fetches API)
  if (picking) return;
  picking = true;

  try {
    const skins = getChampionSkins();
    if (!skins || skins.length === 0) return;

    const nonBaseSkins = skins.filter(s => !isDefaultSkin(s));
    if (nonBaseSkins.length === 0) return;

    let pool = nonBaseSkins;

    if (mode === 'top3') {
      const champName = await getChampionName(championId);
      if (champName) {
        const rankings = await fetchSkinRankings(champName);
        if (rankings && rankings.length > 0) {
          const sorted = [...rankings].sort((a, b) => b.vote_count - a.vote_count);
          const top3Ids = new Set(sorted.slice(0, 3).map(r => parseInt(r.skin_id, 10)));
          const matched = nonBaseSkins.filter(s => top3Ids.has(s.id));
          if (matched.length > 0) {
            pool = matched;
          } else {
            logger.log('no top 3 skins matched loaded skins, falling back to all');
          }
        } else {
          logger.log('rankings API returned no data, falling back to all');
        }
      } else {
        logger.log('could not resolve champion name for rankings, falling back to all');
      }
    }

    const chosen = pickRandom(pool);
    chosenSkinNum = chosen.id % 1000;
    logger.log(`random skin: mode=${mode} chose="${chosen.name}" (num=${chosenSkinNum})`);

    if (clickCarouselSkin(chosenSkinNum)) {
      done = true;
    }
    // If click failed, chosenSkinNum is saved — next poll cycle will retry
  } finally {
    picking = false;
  }
}
