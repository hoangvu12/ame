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
  if (!data || !Array.isArray(data)) return null;
  return data
    .filter(s => s.ownership && (s.ownership.owned || s.ownership.rental?.rented))
    .map(s => s.id);
}

export async function forceDefaultSkin(championId) {
  const defaultSkinId = championId * 1000;
  try {
    const res = await fetch('/lol-champ-select/v1/session/my-selection', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ selectedSkinId: defaultSkinId }),
    });
    return res.ok;
  } catch {
    return false;
  }
}
