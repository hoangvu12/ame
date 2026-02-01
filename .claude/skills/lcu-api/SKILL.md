---
name: lcu-api
description: League of Legends LCU (League Client Update) API reference. Use when working with LCU endpoints, champ select, skin selection, gameflow phases, ownership checks, lobby, matchmaking, store, collections, or any fetch/PATCH calls to the local client API.
allowed-tools: Read, Grep, Glob
argument-hint: [endpoint or topic]
---

# LCU API Reference

Complete reference for the League Client Update (LCU) local REST API. All schemas validated against the official swagger definition (dysolix/hasagi-types).

The LCU API is a local HTTPS REST API exposed by the League of Legends client on `127.0.0.1` with a random port. In CEF plugin context, use relative paths: `fetch('/lol-champ-select/v1/session')`.

## Supporting Files

- [all-endpoints.md](all-endpoints.md) — Every LCU endpoint grouped by plugin (~1700 lines, 90+ plugins)
- [schemas.md](schemas.md) — All schema/type definitions with fields and types (~950 schemas)
- [endpoints.md](endpoints.md) — Detailed docs for commonly used endpoints with example JSON

Use `Read` or `Grep` on these files to find specific endpoint or schema details.

## Swagger Source

https://raw.githubusercontent.com/dysolix/hasagi-types/main/swagger.json

## Key Conventions

- **Skin IDs**: Base/default skin = `championId * 1000`. Example: Ahri (103) → base skin `103000`.
- **PATCH body field**: Skin selection uses `selectedSkinId` (NOT `skinId`). Confirmed across all schema variants.
- **Ownership check**: `s.ownership.owned === true` for purchased. `s.ownership.rental.rented === true` for rented/boosted.
- **Local player in session**: `session.myTeam.find(p => p.cellId === session.localPlayerCellId)`
- **WebSocket events**: `context.socket.observe(path, callback)` where callback receives `{ eventType, uri, data }`.
- **Gameflow phases** (lifecycle order): `None → Lobby → Matchmaking → ReadyCheck → ChampSelect → GameStart → InProgress → WaitingForStats → PreEndOfGame → EndOfGame`

## Plugin Index

Major LCU plugins and what they cover:

| Plugin prefix | Domain |
|---------------|--------|
| `lol-champ-select` | Champion select session, picks, bans, skin selection, swaps |
| `lol-champ-select-legacy` | Legacy champ select (same structure, older endpoints) |
| `lol-champions` | Champion/skin inventories, ownership data |
| `lol-game-data` | Static game data (champion info, skin lists, items) |
| `lol-summoner` | Current summoner info, profile |
| `lol-gameflow` | Game lifecycle phases and session |
| `lol-matchmaking` | Queue, ready check, search |
| `lol-lobby` | Lobby management, invites, party |
| `lol-lobby-team-builder` | Team builder / draft lobby |
| `lol-collections` | Player collections (ward skins, spells, etc.) |
| `lol-catalog` | Store catalog, items, bundles |
| `lol-store` | Store operations, purchases |
| `lol-inventory` | Player inventory (skins, champions, etc.) |
| `lol-loot` | Hextech crafting, loot |
| `lol-chat` | Chat, friends, conversations |
| `lol-ranked` | Ranked stats, tiers, LP |
| `lol-perks` | Runes/perks pages |
| `lol-loadouts` | Item sets, loadouts |
| `lol-cosmetics` | TFT cosmetics (companions, arenas, etc.) |
| `lol-clash` | Clash tournaments |
| `lol-challenges` | Challenges system |
| `lol-champion-mastery` | Mastery points and levels |
| `lol-end-of-game` | Post-game stats |
| `lol-event-hub` | Events, passes, reward tracks |
| `lol-active-boosts` | Active XP/skin boosts |
| `lol-patch` | Client patching status |
| `lol-player-behavior` | Honor, restrictions, reform cards |
| `lol-pre-end-of-game` | Pre-end-of-game sequence |
| `lol-replays` | Replay downloads and playback |
| `lol-settings` | Client settings |
| `lol-suggested-players` | Suggested players to play with |
| `lol-tft` | TFT-specific endpoints |

## Critical Schemas (Quick Ref)

### PATCH `/lol-champ-select/v1/session/my-selection` body

Schema: `TeamBuilderDirect-ChampSelectMySelection`

```jsonc
{
  "selectedSkinId": 0,   // int32 — THE correct field name
  "spell1Id": 0,         // uint64 (optional)
  "spell2Id": 0          // uint64 (optional)
}
```

### Champ select player (in session.myTeam[])

Schema: `TeamBuilderDirect-ChampSelectPlayerSelection`

Key fields: `cellId` (int64), `championId` (int32), `selectedSkinId` (int32), `summonerId` (uint64), `assignedPosition` (string), `spell1Id`, `spell2Id`, `gameName`, `tagLine`, `puuid`

### Skin object (from inventories endpoint)

Schema: `LolChampionsCollectionsChampionSkin`

Key fields: `id` (int32), `championId` (int32), `name` (string), `isBase` (boolean), `ownership` → `{ owned, rental: { rented } }`, `chromas[]`, `splashPath`, `tilePath`

### Ownership

Schema: `LolChampionsCollectionsOwnership`

Fields: `owned` (boolean), `loyaltyReward` (boolean), `xboxGPReward` (boolean), `rental` → `{ rented, endDate, purchaseDate, winCountRemaining, gameId }`

### Summoner

Schema: `LolSummonerSummoner`

Key fields: `summonerId` (uint64), `accountId` (uint64), `puuid` (string), `gameName` (string), `tagLine` (string), `displayName`, `summonerLevel`, `profileIconId`

### Gameflow phase

Schema: `LolGameflowGameflowPhase` (enum string)

Values: `None`, `Lobby`, `Matchmaking`, `CheckedIntoTournament`, `ReadyCheck`, `ChampSelect`, `GameStart`, `FailedToLaunch`, `InProgress`, `Reconnect`, `WaitingForStats`, `PreEndOfGame`, `EndOfGame`, `TerminatedInError`
