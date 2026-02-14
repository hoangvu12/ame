import { createLogger } from './logger';

const logger = createLogger('api');

let championSkins = null;
let cachedChampionId = null;
let cacheEpoch = 0; // incremented on reset; stale fetches check this after await

export async function fetchJson(url) {
  try {
    const res = await fetch(url);
    if (!res.ok) return null;
    return await res.json();
  } catch {
    return null;
  }
}

export async function getMyChampionId() {
  const session = await fetchJson('/lol-champ-select/v1/session');
  if (!session) return null;
  const me = session.myTeam?.find(p => p.cellId === session.localPlayerCellId);
  return me?.championId || null;
}

export async function loadChampionSkins(championId) {
  if (championId === cachedChampionId && championSkins) return championSkins;
  const startEpoch = cacheEpoch;
  const data = await fetchJson(`/lol-game-data/assets/v1/champions/${championId}.json`);
  if (!data) return null;
  // Stale: cache was reset (champ changed) while this fetch was in-flight
  if (cacheEpoch !== startEpoch) return null;
  cachedChampionId = championId;
  championSkins = data.skins || [];
  return championSkins;
}

export function getChampionSkins() {
  return championSkins;
}

export function resetSkinsCache() {
  championSkins = null;
  cachedChampionId = null;
  cacheEpoch++;
}

let championSummaryCache = null;

export async function loadChampionSummary() {
  if (championSummaryCache) return championSummaryCache;
  const data = await fetchJson('/lol-game-data/assets/v1/champion-summary.json');
  if (!data || !Array.isArray(data)) return null;
  championSummaryCache = data;
  return data;
}

export async function getChampionName(championId) {
  const summary = await loadChampionSummary();
  if (!summary) return null;
  const entry = summary.find(c => c.id === championId);
  return entry ? entry.name : null;
}

let englishNamesMap = null;

/** Load English champion names from Data Dragon (cached). */
export async function loadEnglishChampionNames() {
  if (englishNamesMap) return englishNamesMap;
  try {
    const res = await fetch('https://ddragon.leagueoflegends.com/api/versions.json');
    const versions = await res.json();
    const dataRes = await fetch(`https://ddragon.leagueoflegends.com/cdn/${versions[0]}/data/en_US/champion.json`);
    const data = await dataRes.json();
    englishNamesMap = new Map();
    for (const champ of Object.values(data.data)) {
      englishNamesMap.set(parseInt(champ.key), champ.name);
    }
    // Augment cached summary with English names
    if (championSummaryCache) {
      for (const c of championSummaryCache) {
        c.englishName = englishNamesMap.get(c.id) || null;
      }
    }
    return englishNamesMap;
  } catch {
    return null;
  }
}

/** Get English name for a champion by ID (sync, returns null if not loaded yet). */
export function getEnglishChampionName(championId) {
  return englishNamesMap?.get(championId) || null;
}

export async function getChampionIdFromLobbyDOM() {
  const selected = document.querySelector(
    '.quick-play-loadout-selection-hitbox.selected .champion-slot-tile'
  );
  if (!selected) return null;

  const src = selected.getAttribute('src') || '';
  const match = src.match(/\/Characters\/([^/]+)\//i);
  if (!match) return null;

  const alias = match[1];
  const summary = await loadChampionSummary();
  if (!summary) return null;

  const entry = summary.find(
    c => c.alias && c.alias.toLowerCase() === alias.toLowerCase()
  );
  return entry ? entry.id : null;
}

let cachedSummonerId = null;

export async function fetchSummonerId() {
  if (cachedSummonerId) return cachedSummonerId;
  const data = await fetchJson('/lol-summoner/v1/current-summoner');
  if (!data || !data.summonerId) return null;
  cachedSummonerId = data.summonerId;
  return cachedSummonerId;
}

export async function fetchOwnedSkins(summonerId, championId) {
  const data = await fetchJson(
    `/lol-champions/v1/inventories/${summonerId}/champions/${championId}/skins`
  );
  if (!data || !Array.isArray(data)) {
    logger.log(` fetchOwnedSkins: no data for summoner=${summonerId} champ=${championId}`);
    return null;
  }
  const owned = data
    .filter(s => s.ownership && (s.ownership.owned || s.ownership.rental?.rented))
    .map(s => s.id);
  logger.log(` fetchOwnedSkins: champ=${championId} total=${data.length} owned=[${owned.join(', ')}]`);
  return owned;
}

export async function forceDefaultSkin(championId) {
  const defaultSkinId = championId * 1000;
  try {
    const res = await fetch('/lol-champ-select/v1/session/my-selection', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ selectedSkinId: defaultSkinId }),
    });
    logger.log(` forceDefaultSkin: PATCH selectedSkinId=${defaultSkinId} status=${res.status}`);
    return res.ok;
  } catch (err) {
    logger.log(` forceDefaultSkin: error`, err);
    return false;
  }
}
