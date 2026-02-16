import { readCurrentSkin, findSkinByName } from './skin';
import { getSkinForced } from './state';
import { fetchJson } from './api';
import { createLogger } from './logger';

const logger = createLogger('splash');

const STYLE_ID = 'ame-splash-overrides';
// .image-magic-background on low-spec, .video-magic-background-state-machine on normal
const BG_SELECTOR = '.skin-showcase';

let selfSplashUrl = null;
let lastNonDefaultUrl = null;
let teammateSplashUrls = {};
let selfLogOnce = false;

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
  for (const [index, url] of Object.entries(teammateSplashUrls)) {
    const nth = parseInt(index, 10) + 1;
    css += `.your-party .party > .summoner-wrapper:nth-child(${nth}) ${BG_SELECTOR} { background-image: url('${url}') !important; }\n`;
  }
  tag.textContent = css;
}

export function updateSelfSplash(skins) {
  if (!skins) return;

  if (!document.querySelector(`.summoner-object.is-self ${BG_SELECTOR}`)) return;

  const skinName = readCurrentSkin();
  if (!skinName) return;

  const skin = findSkinByName(skins, skinName);
  if (!skin) {
    if (!selfLogOnce) { logger.log(`self splash: skin not found for "${skinName}"`); selfLogOnce = true; }
    return;
  }

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

  const url = skin.splashPath || skin.uncenteredSplashPath || skin.tilePath;
  if (!url) {
    if (!selfLogOnce) { logger.log(`self splash: no splashPath for ${skin.name} (id=${skin.id})`); selfLogOnce = true; }
    return;
  }

  if (skinNum !== 0) lastNonDefaultUrl = url;

  if (selfSplashUrl === url) return;
  selfSplashUrl = url;
  selfLogOnce = false;
  applyStyles();
  logger.log(`self splash: skinNum=${skinNum}`);
}

export async function updateTeammateSplashes(teammates, session) {
  if (!teammates || teammates.length === 0 || !session?.myTeam) return;

  const teamOrdered = session.myTeam.slice().sort((a, b) => a.cellId - b.cellId);
  const wrappers = document.querySelectorAll('.your-party .party > .summoner-wrapper');

  if (!wrappers.length) {
    logger.log(`teammate splash: no .your-party wrappers found`);
    return;
  }

  let changed = false;
  for (const tm of teammates) {
    if (!tm.skinInfo?.skinId) continue;

    const teamIndex = teamOrdered.findIndex(p => p.puuid === tm.puuid);
    if (teamIndex < 0 || teamIndex >= wrappers.length) continue;

    if (!wrappers[teamIndex].querySelector(BG_SELECTOR)) continue;

    const championId = parseInt(tm.skinInfo.championId, 10);
    if (!championId) continue;

    const rawSkinId = parseInt(tm.skinInfo.skinId, 10) || 0;
    const rawBaseSkinId = parseInt(tm.skinInfo.baseSkinId, 10) || 0;
    // For chromas, use baseSkinId to get the parent skin's splash
    const effectiveSkinId = (rawBaseSkinId > 0 && rawBaseSkinId !== rawSkinId) ? rawBaseSkinId : rawSkinId;

    const skins = await loadTeammateChampionSkins(championId);
    if (!skins) continue;

    const skin = skins.find(s => s.id === effectiveSkinId);
    if (!skin) {
      logger.log(`teammate splash: skin id ${effectiveSkinId} not found for champ ${championId}`);
      continue;
    }

    const url = skin.splashPath || skin.uncenteredSplashPath || skin.tilePath;
    if (!url) {
      logger.log(`teammate splash: no splashPath for ${skin.name} (id=${skin.id})`);
      continue;
    }

    if (teammateSplashUrls[teamIndex] === url) continue;

    teammateSplashUrls[teamIndex] = url;
    changed = true;
    logger.log(`teammate splash: team index ${teamIndex} skin=${skin.name}`);
  }

  if (changed) applyStyles();
}

export function resetSplashState() {
  selfSplashUrl = null;
  lastNonDefaultUrl = null;
  teammateSplashUrls = {};
  selfLogOnce = false;
  Object.keys(teammateSkinsCache).forEach(k => delete teammateSkinsCache[k]);
  const tag = document.getElementById(STYLE_ID);
  if (tag) tag.textContent = '';
}
