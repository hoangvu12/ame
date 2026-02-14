import { readCurrentSkin, findSkinByName } from './skin';
import { getSkinForced } from './state';
import { fetchJson } from './api';
import { createLogger } from './logger';

const logger = createLogger('splash');

const STYLE_ID = 'ame-splash-overrides';
const BG_SELECTOR = '.image-magic-background';

let selfSplashUrl = null;
let lastNonDefaultUrl = null;
let teammateSplashUrls = {};

// Cache teammate champion skins so we don't re-fetch every update.
// Separate from the main skins cache which only holds the local player's champion.
const teammateSkinsCache = {};

async function loadTeammateChampionSkins(championId) {
  if (teammateSkinsCache[championId]) return teammateSkinsCache[championId];
  const data = await fetchJson(`/lol-game-data/assets/v1/champions/${championId}.json`);
  if (!data?.skins) return null;
  teammateSkinsCache[championId] = data.skins;
  return data.skins;
}

function applyStyles() {
  let tag = document.getElementById(STYLE_ID);
  if (!tag) {
    tag = document.createElement('style');
    tag.id = STYLE_ID;
    document.head.appendChild(tag);
  }

  let css = '';
  if (selfSplashUrl) {
    css += `.summoner-object.is-self ${BG_SELECTOR} { background-image: url('${selfSplashUrl}') !important; }\n`;
  }
  for (const [cellId, url] of Object.entries(teammateSplashUrls)) {
    css += `.summoner-object.slot-id-${cellId} ${BG_SELECTOR} { background-image: url('${url}') !important; }\n`;
  }
  tag.textContent = css;
}

export function updateSelfSplash(skins) {
  if (!skins) return;

  if (!document.querySelector(`.summoner-object.is-self ${BG_SELECTOR}`)) return;

  const skinName = readCurrentSkin();
  if (!skinName) return;

  const skin = findSkinByName(skins, skinName);
  if (!skin) return;

  const skinNum = skin.id % 1000;

  // After forceDefaultSkin the client resets the background to base.
  // Keep showing the last selected splash instead of flashing to default.
  if (skinNum === 0 && getSkinForced() && lastNonDefaultUrl) {
    if (selfSplashUrl !== lastNonDefaultUrl) {
      selfSplashUrl = lastNonDefaultUrl;
      applyStyles();
    }
    return;
  }

  const url = skin.splashPath;
  if (!url) return;

  if (skinNum !== 0) lastNonDefaultUrl = url;

  if (selfSplashUrl === url) return;
  selfSplashUrl = url;
  applyStyles();
  logger.log(`self splash: skinNum=${skinNum}`);
}

export async function updateTeammateSplashes(teammates, session) {
  if (!teammates || teammates.length === 0 || !session?.myTeam) return;

  let changed = false;
  for (const tm of teammates) {
    if (!tm.skinInfo?.skinId) continue;

    const member = session.myTeam.find(p => p.puuid === tm.puuid);
    if (!member) continue;

    const cellId = member.cellId;
    if (!document.querySelector(`.summoner-object.slot-id-${cellId} ${BG_SELECTOR}`)) continue;

    const championId = parseInt(tm.skinInfo.championId, 10);
    if (!championId) continue;

    const rawSkinId = parseInt(tm.skinInfo.skinId, 10) || 0;
    const rawBaseSkinId = parseInt(tm.skinInfo.baseSkinId, 10) || 0;
    // For chromas, use baseSkinId to get the parent skin's splash
    const effectiveSkinId = (rawBaseSkinId > 0 && rawBaseSkinId !== rawSkinId) ? rawBaseSkinId : rawSkinId;

    const skins = await loadTeammateChampionSkins(championId);
    if (!skins) continue;

    const skin = skins.find(s => s.id === effectiveSkinId);
    if (!skin?.splashPath) continue;

    const url = skin.splashPath;
    if (teammateSplashUrls[cellId] === url) continue;

    teammateSplashUrls[cellId] = url;
    changed = true;
    logger.log(`teammate splash: slot ${cellId} skin=${skin.name}`);
  }

  if (changed) applyStyles();
}

export function resetSplashState() {
  selfSplashUrl = null;
  lastNonDefaultUrl = null;
  teammateSplashUrls = {};
  Object.keys(teammateSkinsCache).forEach(k => delete teammateSkinsCache[k]);
  const tag = document.getElementById(STYLE_ID);
  if (tag) tag.textContent = '';
}
