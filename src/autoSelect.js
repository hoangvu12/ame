import { onSetting, onAutoSelectRoles } from './websocket';

let enabled = false;
let rolesConfig = {};
let acting = false;

export function loadAutoSelectSetting() {
  onSetting('autoSelect', (v) => { enabled = v; });
  onAutoSelectRoles((roles) => { rolesConfig = roles; });
}

export function resetAutoSelect() {
  acting = false;
}

export function handleChampSelectSession(session) {
  if (!enabled || acting) return;

  const me = session.myTeam?.find(p => p.cellId === session.localPlayerCellId);
  if (!me) return;

  const position = me.assignedPosition?.toLowerCase();
  if (!position) return;

  const roleConfig = rolesConfig[position];
  if (!roleConfig) return;

  // Find all incomplete actions for local player (no isInProgress check —
  // the API rejects PATCHes on non-active actions, we just retry next event)
  const myActions = (session.actions?.flat() || []).filter(
    a => a.actorCellId === session.localPlayerCellId && !a.completed
  );
  if (myActions.length === 0) return;

  // Banned champions (from session.bans, more reliable than inferring from actions)
  const banned = new Set(
    [...(session.bans?.myTeamBans || []), ...(session.bans?.theirTeamBans || [])]
      .filter(id => id > 0)
  );

  // Picked champions (both teams)
  const picked = new Set();
  for (const p of [...(session.myTeam || []), ...(session.theirTeam || [])]) {
    if (p.championId) picked.add(p.championId);
  }

  // Teammate intents (avoid banning what a teammate wants to play)
  const teamIntents = new Set(
    (session.myTeam || [])
      .filter(p => p.cellId !== session.localPlayerCellId && p.championPickIntent)
      .map(p => p.championPickIntent)
  );

  for (const action of myActions) {
    const isBan = action.type === 'ban';
    const priorityList = isBan ? roleConfig.bans : roleConfig.picks;
    if (!priorityList || priorityList.length === 0) continue;

    const championId = priorityList.find(id => {
      if (banned.has(id)) return false;
      if (isBan && teamIntents.has(id)) return false;
      if (!isBan && picked.has(id)) return false;
      return true;
    });
    if (!championId) continue;

    // Skip if this action already has our desired champion (avoid spamming)
    if (action.championId === championId) continue;

    acting = true;
    performAction(action.id, championId);
    return; // one action at a time
  }
}

async function performAction(actionId, championId) {
  try {
    await fetch(`/lol-champ-select/v1/session/actions/${actionId}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ championId, completed: true }),
    });
  } catch {
    // Silently fail — will retry on next session event
  } finally {
    acting = false;
  }
}
