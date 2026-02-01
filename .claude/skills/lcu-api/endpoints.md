# LCU API — Endpoint & Schema Details

All schemas validated against the official swagger (dysolix/hasagi-types).

---

## Champ Select

### GET `/lol-champ-select/v1/session`

Returns the full champ select session.

**Response schema**: `TeamBuilderDirect-ChampSelectSession`

```jsonc
{
  "id": "string",
  "gameId": 0,               // uint64
  "queueId": 0,              // int32
  "timer": { /* ChampSelectTimer */ },
  "chatDetails": { /* ChatRoomDetails */ },
  "myTeam": [                // array of ChampSelectPlayerSelection
    {
      "cellId": 0,           // int64 — use to find local player
      "championId": 0,       // int32 — 0 if not yet picked
      "selectedSkinId": 0,   // int32 — current skin selection
      "wardSkinId": 0,       // int64
      "spell1Id": 0,         // uint64
      "spell2Id": 0,         // uint64
      "team": 1,             // int32 (1 = blue, 2 = red)
      "assignedPosition": "MIDDLE", // string
      "championPickIntent": 0,
      "playerType": "HUMAN",
      "summonerId": 0,       // uint64
      "gameName": "string",
      "tagLine": "string",
      "puuid": "string"
    }
  ],
  "theirTeam": [ /* same structure */ ],
  "trades": [ /* SwapContract[] */ ],
  "pickOrderSwaps": [ /* SwapContract[] */ ],
  "actions": [ /* array of action objects */ ],
  "bans": { /* BannedChampions */ },
  "localPlayerCellId": 0,    // int64 — match against myTeam[].cellId
  "allowSkinSelection": true,
  "allowRerolling": true,
  "rerollsRemaining": 0,
  "benchEnabled": false,
  "benchChampions": []
}
```

**Finding the local player**:
```js
const me = session.myTeam?.find(p => p.cellId === session.localPlayerCellId);
const championId = me?.championId || null;
```

### PATCH `/lol-champ-select/v1/session/my-selection`

Change your skin or summoner spells during champ select.

**Request body schema**: `TeamBuilderDirect-ChampSelectMySelection`

```jsonc
{
  "selectedSkinId": 0,   // int32 — REQUIRED field name (NOT "skinId")
  "spell1Id": 0,         // uint64 (optional)
  "spell2Id": 0          // uint64 (optional)
}
```

**Important**: The field is `selectedSkinId`. This is confirmed across all schema variants:
- `TeamBuilderDirect-ChampSelectMySelection`
- `LolChampSelectLegacyChampSelectMySelection`
- `ChampSelectMySelection`

All use `selectedSkinId`.

**Forcing default skin**: PATCH with `{ "selectedSkinId": championId * 1000 }`.

**Constraints**: This endpoint only works during champ select phase. Returns error/404 outside of champ select.

### PATCH `/lol-champ-select/v1/session/actions/{actionId}`

Hover or lock in a champion during your pick/ban action.

```jsonc
// Hover a champion (without locking)
{ "championId": 103 }

// Lock in the champion
{ "championId": 103, "completed": true }
```

### POST `/lol-champ-select/v1/session/bench/swap/{championId}`

Swap your current champion with one on the ARAM bench. No request body needed.

### GET `/lol-champ-select/v1/session/timer`

Returns current champ select timer information.

---

## Champion & Skin Data

### GET `/lol-champions/v1/inventories/{summonerId}/champions/{championId}/skins`

Returns all skins for a champion with ownership info.

**Path params**:
- `summonerId`: uint64
- `championId`: int32

**Response**: Array of `LolChampionsCollectionsChampionSkin`

```jsonc
[
  {
    "championId": 103,          // int32
    "id": 103000,               // int32 — skin ID (base = championId * 1000)
    "name": "default",          // string
    "isBase": true,             // boolean
    "disabled": false,          // boolean
    "stillObtainable": true,    // boolean
    "lastSelected": false,      // boolean
    "ownership": {              // LolChampionsCollectionsOwnership
      "loyaltyReward": false,   // boolean
      "xboxGPReward": false,    // boolean
      "owned": true,            // boolean — PRIMARY ownership flag
      "rental": {               // LolChampionsCollectionsRental
        "rented": false,        // boolean — check for rented skins
        "endDate": 0,
        "purchaseDate": 0,
        "winCountRemaining": 0,
        "gameId": 0
      }
    },
    "splashPath": "/lol-game-data/assets/...",
    "tilePath": "/lol-game-data/assets/...",
    "chromaPath": "",
    "chromas": [                // array of ChampionChroma
      {
        "id": 103001,
        "championId": 103,
        "name": "Chroma Name",
        "chromaPath": "...",
        "ownership": { /* same structure */ }
      }
    ],
    "skinAugments": { /* augments data */ },
    "questSkinInfo": { /* quest skin data */ },
    "emblems": []
  }
]
```

**Ownership check — complete logic**:
```js
// Owned: purchased or received as reward
s.ownership.owned === true

// Rented: temporary access (e.g., skin boost, rental)
s.ownership.rental?.rented === true

// Effectively available to the player:
s.ownership.owned || s.ownership.rental?.rented
```

### GET `/lol-game-data/assets/v1/champions/{championId}.json`

Returns champion data including skins array. This is a static data endpoint (CDN-like), different from the inventory endpoint above. Does NOT include ownership info.

```jsonc
{
  "id": 103,
  "name": "Ahri",
  "skins": [
    { "id": 103000, "name": "default", "isBase": true },
    { "id": 103001, "name": "Dynasty Ahri", "isBase": false },
    // ...
  ]
}
```

### GET `/lol-game-data/assets/v1/champion-summary.json`

Returns summary of all champions. Used for name/alias lookups.

```jsonc
[
  {
    "id": 103,
    "name": "Ahri",
    "alias": "Ahri",         // internal alias, used in asset paths
    "squarePortraitPath": "...",
    "roles": ["mage", "assassin"]
  },
  // ...
]
```

---

## Summoner

### GET `/lol-summoner/v1/current-summoner`

Returns the currently logged-in summoner.

**Response schema**: `LolSummonerSummoner`

```jsonc
{
  "summonerId": 0,            // uint64 — use for inventory endpoints
  "accountId": 0,             // uint64
  "displayName": "string",
  "internalName": "string",
  "profileIconId": 0,         // int32
  "summonerLevel": 0,         // uint32
  "xpSinceLastLevel": 0,      // uint64
  "xpUntilNextLevel": 0,      // uint64
  "percentCompleteForNextLevel": 0,
  "puuid": "string",
  "gameName": "string",
  "tagLine": "string",
  "nameChangeFlag": false,
  "unnamed": false
}
```

---

## Gameflow

### GET `/lol-gameflow/v1/gameflow-phase`

Returns the current phase as a plain string (not JSON object).

**Response**: `LolGameflowGameflowPhase` enum string

```
"None"
"Lobby"
"Matchmaking"
"CheckedIntoTournament"
"ReadyCheck"
"ChampSelect"
"GameStart"
"FailedToLaunch"
"InProgress"
"Reconnect"
"WaitingForStats"
"PreEndOfGame"
"EndOfGame"
"TerminatedInError"
```

**Phase lifecycle** (typical game):
```
None → Lobby → Matchmaking → ReadyCheck → ChampSelect → GameStart → InProgress → WaitingForStats → PreEndOfGame → EndOfGame → None
```

**Swiftplay note**: Swiftplay uses `Lobby` phase for champion/skin selection (no separate champ select). The PATCH my-selection endpoint does NOT work during `Lobby` — only during `ChampSelect`.

### GET `/lol-gameflow/v1/session`

Returns the full gameflow session with game details, map info, queue info, etc.

---

## Matchmaking

### GET `/lol-matchmaking/v1/ready-check`

Returns current ready check status.

```jsonc
{
  "state": "InProgress",        // "Invalid", "PartyNotReady", "InProgress", "EveryoneReady", etc.
  "playerResponse": "None",     // "None", "Accepted", "Declined"
  "dodgeWarning": "None",
  "timer": 0.0,
  "declinerIds": []
}
```

### POST `/lol-matchmaking/v1/ready-check/accept`

Accept the ready check. No request body needed.

---

## Skin ID Convention

Skin IDs follow a predictable pattern:

```
Base skin:   championId * 1000         (e.g., Ahri base = 103000)
First skin:  championId * 1000 + 1     (e.g., Dynasty Ahri = 103001)
Chromas:     baseSkinId + offset        (varies per skin)
```

Example for Ahri (championId = 103):
- 103000 = default (base)
- 103001 = Dynasty Ahri
- 103002 = Midnight Ahri
- 103003 = Foxfire Ahri
- etc.

---

## WebSocket Event Format

Events received via `context.socket.observe(path, callback)` have this shape:

```jsonc
{
  "eventType": "Create" | "Update" | "Delete",
  "uri": "/lol-gameflow/v1/gameflow-phase",
  "data": /* the endpoint's response payload */
}
```

For gameflow-phase, `event.data` is the phase string directly.
For champ-select session, `event.data` is the full session object.

---

## Common Patterns

### Fetching with error handling
```js
async function fetchJson(url) {
  try {
    const res = await fetch(url);
    if (!res.ok) return null;
    return await res.json();
  } catch {
    return null;
  }
}
```

### PATCH with JSON body
```js
await fetch('/lol-champ-select/v1/session/my-selection', {
  method: 'PATCH',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ selectedSkinId: championId * 1000 }),
});
```

### Finding local player in session
```js
const session = await fetchJson('/lol-champ-select/v1/session');
const me = session?.myTeam?.find(p => p.cellId === session.localPlayerCellId);
const championId = me?.championId || null;
const currentSkinId = me?.selectedSkinId || null;
```
