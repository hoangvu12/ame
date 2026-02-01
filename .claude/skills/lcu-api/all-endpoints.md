
## /AsyncDelete

- `POST` `/AsyncDelete`

## /AsyncResult

- `POST` `/AsyncResult`

## /AsyncStatus

- `POST` `/AsyncStatus`

## /Cancel

- `POST` `/Cancel`

## /Exit

- `POST` `/Exit`

## /Help

- `POST` `/Help`

## /LoggingGetEntries

- `POST` `/LoggingGetEntries`

## /LoggingMetrics

- `POST` `/LoggingMetrics`

## /LoggingMetricsMetadata

- `POST` `/LoggingMetricsMetadata`

## /LoggingStart

- `POST` `/LoggingStart`

## /LoggingStop

- `POST` `/LoggingStop`

## /Subscribe

- `POST` `/Subscribe`

## /Unsubscribe

- `POST` `/Unsubscribe`

## /WebSocketFormat

- `POST` `/WebSocketFormat`

## /async

- `GET` `/async/v1/result/{asyncToken}`
- `DELETE, GET` `/async/v1/status/{asyncToken}`

## /client-config

- `PUT` `/client-config/v1/entitlements-token`
- `PUT` `/client-config/v1/refresh-config-status`
- `GET` `/client-config/v1/status/{type}`
- `GET` `/client-config/v2/config/{name}`
- `PUT` `/client-config/v2/namespace-changes`
- `GET` `/client-config/v2/namespace/{namespace}`
- `GET` `/client-config/v2/namespace/{namespace}/player`
- `GET` `/client-config/v2/namespace/{namespace}/public`

## /cookie-jar

- `GET, POST` `/cookie-jar/v1/cookies`

## /crash-reporting

- `GET` `/crash-reporting/v1/crash-status`

## /data-store

- `GET` `/data-store/v1/install-dir`
- `GET, POST` `/data-store/v1/install-settings/{+path}`
- `GET` `/data-store/v1/system-settings/{+path}`

## /deep-links

- `POST` `/deep-links/v1/launch-lor-link`
- `GET` `/deep-links/v1/settings`

## /entitlements

- `GET` `/entitlements/v1/entitlements`
- `GET` `/entitlements/v1/token`

## /error-monitor

- `POST` `/error-monitor/v1/logs/batches`
- `GET` `/error-monitor/v1/logs/changed`
- `POST` `/error-monitor/v1/ux/logs`

## /lol-account-verification

- `POST` `/lol-account-verification/v1/confirmActivationPin`
- `POST` `/lol-account-verification/v1/confirmDeactivationPin`
- `GET` `/lol-account-verification/v1/is-verified`
- `GET` `/lol-account-verification/v1/phone-number`
- `POST` `/lol-account-verification/v1/sendActivationPin`
- `POST` `/lol-account-verification/v1/sendDeactivationPin`

## /lol-active-boosts

- `GET` `/lol-active-boosts/v1/active-boosts`

## /lol-activity-center

- `POST` `/lol-activity-center/v1/clear-cache`
- `GET` `/lol-activity-center/v1/config-data`
- `GET` `/lol-activity-center/v1/content/client-nav`
- `GET` `/lol-activity-center/v1/content/{pageName}`
- `GET` `/lol-activity-center/v1/overrides`
- `GET` `/lol-activity-center/v1/ready`

## /lol-banners

- `GET` `/lol-banners/v1/current-summoner/flags`
- `GET, PUT` `/lol-banners/v1/current-summoner/flags/equipped`
- `GET` `/lol-banners/v1/current-summoner/frames/equipped`
- `GET` `/lol-banners/v1/players/{puuid}/flags/equipped`

## /lol-cap-missions

- `GET` `/lol-cap-missions/v1/getmissions`
- `POST` `/lol-cap-missions/v1/markasviewed`
- `GET` `/lol-cap-missions/v1/ready`

## /lol-catalog

- `GET` `/lol-catalog/v1/item-details`
- `GET` `/lol-catalog/v1/items`
- `GET` `/lol-catalog/v1/items-list-details`
- `GET` `/lol-catalog/v1/items-list-details/skip-cache`
- `GET` `/lol-catalog/v1/items/{bundleId}/related-items`
- `GET` `/lol-catalog/v1/items/{inventoryType}`
- `GET` `/lol-catalog/v1/items/{skinId}/related-bundles`
- `GET` `/lol-catalog/v1/ready`

## /lol-challenges

- `POST` `/lol-challenges/v1/ack-challenge-update/{id}`
- `GET` `/lol-challenges/v1/available-queue-ids`
- `GET` `/lol-challenges/v1/challenges/category-data`
- `GET` `/lol-challenges/v1/challenges/local-player`
- `GET` `/lol-challenges/v1/client-state`
- `GET` `/lol-challenges/v1/level-points`
- `GET` `/lol-challenges/v1/my-updated-challenges/{gameId}`
- `GET` `/lol-challenges/v1/penalty`
- `POST` `/lol-challenges/v1/rsbot-challenges/{gameId}`
- `GET` `/lol-challenges/v1/seasons`
- `GET` `/lol-challenges/v1/summary-player-data/local-player`
- `GET` `/lol-challenges/v1/summary-player-data/player/{puuid}`
- `GET` `/lol-challenges/v1/summary-players-data/players`
- `POST` `/lol-challenges/v1/update-player-preferences`
- `GET` `/lol-challenges/v1/updated-challenges/{gameId}/{puuid}`
- `GET` `/lol-challenges/v2/titles/all`
- `GET` `/lol-challenges/v2/titles/local-player`

## /lol-champ-select

- `GET` `/lol-champ-select/v1/all-grid-champions`
- `GET` `/lol-champ-select/v1/bannable-champion-ids`
- `POST` `/lol-champ-select/v1/battle-training/launch`
- `GET` `/lol-champ-select/v1/current-champion`
- `GET` `/lol-champ-select/v1/disabled-champion-ids`
- `GET` `/lol-champ-select/v1/grid-champions/{championId}`
- `GET` `/lol-champ-select/v1/muted-players`
- `GET` `/lol-champ-select/v1/ongoing-champion-swap`
- `POST` `/lol-champ-select/v1/ongoing-champion-swap/{id}/clear`
- `GET` `/lol-champ-select/v1/ongoing-pick-order-swap`
- `POST` `/lol-champ-select/v1/ongoing-pick-order-swap/{id}/clear`
- `GET` `/lol-champ-select/v1/ongoing-position-swap`
- `POST` `/lol-champ-select/v1/ongoing-position-swap/{id}/clear`
- `GET` `/lol-champ-select/v1/pickable-champion-ids`
- `GET` `/lol-champ-select/v1/pickable-skin-ids`
- `GET` `/lol-champ-select/v1/pin-drop-notification`
- `POST` `/lol-champ-select/v1/retrieve-latest-game-dto`
- `GET` `/lol-champ-select/v1/session`
- `PATCH` `/lol-champ-select/v1/session/actions/{id}`
- `POST` `/lol-champ-select/v1/session/actions/{id}/complete`
- `POST` `/lol-champ-select/v1/session/bench/swap/{championId}`
- `GET` `/lol-champ-select/v1/session/champion-swaps`
- `GET` `/lol-champ-select/v1/session/champion-swaps/{id}`
- `POST` `/lol-champ-select/v1/session/champion-swaps/{id}/accept`
- `POST` `/lol-champ-select/v1/session/champion-swaps/{id}/cancel`
- `POST` `/lol-champ-select/v1/session/champion-swaps/{id}/decline`
- `POST` `/lol-champ-select/v1/session/champion-swaps/{id}/request`
- `GET, PATCH` `/lol-champ-select/v1/session/my-selection`
- `POST` `/lol-champ-select/v1/session/my-selection/reroll`
- `GET` `/lol-champ-select/v1/session/pick-order-swaps`
- `GET` `/lol-champ-select/v1/session/pick-order-swaps/{id}`
- `POST` `/lol-champ-select/v1/session/pick-order-swaps/{id}/accept`
- `POST` `/lol-champ-select/v1/session/pick-order-swaps/{id}/cancel`
- `POST` `/lol-champ-select/v1/session/pick-order-swaps/{id}/decline`
- `POST` `/lol-champ-select/v1/session/pick-order-swaps/{id}/request`
- `GET` `/lol-champ-select/v1/session/position-swaps`
- `POST` `/lol-champ-select/v1/session/position-swaps/{id}/accept`
- `POST` `/lol-champ-select/v1/session/position-swaps/{id}/cancel`
- `POST` `/lol-champ-select/v1/session/position-swaps/{id}/decline`
- `POST` `/lol-champ-select/v1/session/position-swaps/{id}/request`
- `POST` `/lol-champ-select/v1/session/simple-inventory`
- `GET` `/lol-champ-select/v1/session/timer`
- `GET` `/lol-champ-select/v1/sfx-notifications`
- `GET` `/lol-champ-select/v1/skin-carousel-skins`
- `GET` `/lol-champ-select/v1/skin-selector-info`
- `GET` `/lol-champ-select/v1/summoners/{slotId}`
- `GET` `/lol-champ-select/v1/team-boost`
- `POST` `/lol-champ-select/v1/team-boost/purchase`
- `POST` `/lol-champ-select/v1/toggle-favorite/{championId}/{position}`
- `POST` `/lol-champ-select/v1/toggle-player-muted`

## /lol-champ-select-legacy

- `GET` `/lol-champ-select-legacy/v1/bannable-champion-ids`
- `POST` `/lol-champ-select-legacy/v1/battle-training/launch`
- `GET` `/lol-champ-select-legacy/v1/current-champion`
- `GET` `/lol-champ-select-legacy/v1/disabled-champion-ids`
- `GET` `/lol-champ-select-legacy/v1/implementation-active`
- `GET` `/lol-champ-select-legacy/v1/pickable-champion-ids`
- `GET` `/lol-champ-select-legacy/v1/pickable-skin-ids`
- `GET` `/lol-champ-select-legacy/v1/session`
- `PATCH` `/lol-champ-select-legacy/v1/session/actions/{id}`
- `POST` `/lol-champ-select-legacy/v1/session/actions/{id}/complete`
- `GET` `/lol-champ-select-legacy/v1/session/champion-swaps`
- `GET` `/lol-champ-select-legacy/v1/session/champion-swaps/{id}`
- `POST` `/lol-champ-select-legacy/v1/session/champion-swaps/{id}/accept`
- `POST` `/lol-champ-select-legacy/v1/session/champion-swaps/{id}/cancel`
- `POST` `/lol-champ-select-legacy/v1/session/champion-swaps/{id}/decline`
- `POST` `/lol-champ-select-legacy/v1/session/champion-swaps/{id}/request`
- `GET, PATCH` `/lol-champ-select-legacy/v1/session/my-selection`
- `POST` `/lol-champ-select-legacy/v1/session/my-selection/reroll`
- `GET` `/lol-champ-select-legacy/v1/session/timer`
- `GET` `/lol-champ-select-legacy/v1/team-boost`

## /lol-champion-mastery

- `GET` `/lol-champion-mastery/v1/local-player/champion-mastery`
- `GET` `/lol-champion-mastery/v1/local-player/champion-mastery-score`
- `GET` `/lol-champion-mastery/v1/local-player/champion-mastery-sets-and-rewards`
- `GET` `/lol-champion-mastery/v1/notifications`
- `POST` `/lol-champion-mastery/v1/notifications/ack`
- `GET` `/lol-champion-mastery/v1/reward-grants`
- `DELETE` `/lol-champion-mastery/v1/reward-grants/{id}`
- `POST` `/lol-champion-mastery/v1/scouting`
- `GET` `/lol-champion-mastery/v1/{puuid}/champion-mastery`
- `POST` `/lol-champion-mastery/v1/{puuid}/champion-mastery-view`
- `POST` `/lol-champion-mastery/v1/{puuid}/champion-mastery-view/top`
- `POST` `/lol-champion-mastery/v1/{puuid}/champion-mastery/top`

## /lol-champions

- `GET` `/lol-champions/v1/inventories/{summonerId}/champions`
- `GET` `/lol-champions/v1/inventories/{summonerId}/champions-minimal`
- `GET` `/lol-champions/v1/inventories/{summonerId}/champions-playable-count`
- `GET` `/lol-champions/v1/inventories/{summonerId}/champions/{championId}`
- `GET` `/lol-champions/v1/inventories/{summonerId}/champions/{championId}/skins`
- `GET` `/lol-champions/v1/inventories/{summonerId}/champions/{championId}/skins/{championSkinId}`
- `GET` `/lol-champions/v1/inventories/{summonerId}/champions/{championId}/skins/{skinId}/chromas`
- `GET` `/lol-champions/v1/inventories/{summonerId}/skins-minimal`
- `GET` `/lol-champions/v1/inventories/{summonerId}/{queueId}/champions-minimal-per-queue`
- `GET` `/lol-champions/v1/owned-champions-minimal`

## /lol-chat

- `GET, POST` `/lol-chat/v1/blocked-players`
- `DELETE, GET` `/lol-chat/v1/blocked-players/{id}`
- `GET` `/lol-chat/v1/config`
- `GET, POST` `/lol-chat/v1/conversations`
- `DELETE, GET, PUT` `/lol-chat/v1/conversations/active`
- `POST` `/lol-chat/v1/conversations/eog-chat-toggle`
- `GET` `/lol-chat/v1/conversations/notify`
- `DELETE, GET, PUT` `/lol-chat/v1/conversations/{id}`
- `DELETE, GET, POST` `/lol-chat/v1/conversations/{id}/messages`
- `GET` `/lol-chat/v1/conversations/{id}/participants`
- `GET` `/lol-chat/v1/conversations/{id}/participants/{pid}`
- `POST` `/lol-chat/v1/discord-link`
- `GET` `/lol-chat/v1/errors`
- `DELETE` `/lol-chat/v1/errors/{id}`
- `GET` `/lol-chat/v1/friend-counts`
- `GET` `/lol-chat/v1/friend-exists/{summonerId}`
- `GET, POST` `/lol-chat/v1/friend-groups`
- `PUT` `/lol-chat/v1/friend-groups/order`
- `DELETE, GET, PUT` `/lol-chat/v1/friend-groups/{id}`
- `GET` `/lol-chat/v1/friends`
- `DELETE, GET, PUT` `/lol-chat/v1/friends/{id}`
- `GET` `/lol-chat/v1/is-discord-link-available`
- `GET` `/lol-chat/v1/is-discord-linked`
- `GET, PUT` `/lol-chat/v1/me`
- `DELETE, GET, POST` `/lol-chat/v1/player-mutes`
- `GET` `/lol-chat/v1/resources`
- `GET` `/lol-chat/v1/session`
- `GET, PUT` `/lol-chat/v1/settings`
- `POST` `/lol-chat/v1/system-mutes`
- `GET` `/lol-chat/v2/friend-exists/{puuid}`
- `GET, POST` `/lol-chat/v2/friend-requests`
- `DELETE, PUT` `/lol-chat/v2/friend-requests/{id}`

## /lol-clash

- `GET` `/lol-clash/v1/all-tournaments`
- `GET` `/lol-clash/v1/awaiting-resent-eog`
- `GET` `/lol-clash/v1/bracket/{bracketId}`
- `GET` `/lol-clash/v1/checkin-allowed`
- `GET` `/lol-clash/v1/currentTournamentIds`
- `GET` `/lol-clash/v1/disabled-config`
- `GET` `/lol-clash/v1/enabled`
- `GET` `/lol-clash/v1/eog-player-update`
- `POST` `/lol-clash/v1/eog-player-update/acknowledge`
- `GET` `/lol-clash/v1/event/{uuid}`
- `POST` `/lol-clash/v1/events`
- `GET` `/lol-clash/v1/game-end`
- `POST` `/lol-clash/v1/game-end/acknowledge`
- `GET` `/lol-clash/v1/historyandwinners`
- `GET` `/lol-clash/v1/iconconfig`
- `GET` `/lol-clash/v1/invited-roster-ids`
- `POST` `/lol-clash/v1/lft/player`
- `POST` `/lol-clash/v1/lft/player/find`
- `POST` `/lol-clash/v1/lft/team`
- `POST` `/lol-clash/v1/lft/team/fetch-requests`
- `POST` `/lol-clash/v1/lft/team/find`
- `GET` `/lol-clash/v1/lft/team/requests`
- `POST` `/lol-clash/v1/lft/team/{rosterId}/request`
- `GET` `/lol-clash/v1/notifications`
- `POST` `/lol-clash/v1/notifications/acknowledge`
- `GET` `/lol-clash/v1/ping`
- `GET` `/lol-clash/v1/player`
- `GET` `/lol-clash/v1/player/chat-rosters`
- `GET` `/lol-clash/v1/player/history`
- `GET` `/lol-clash/v1/playmode-restricted`
- `GET` `/lol-clash/v1/ready`
- `POST` `/lol-clash/v1/refresh`
- `GET` `/lol-clash/v1/rewards`
- `GET` `/lol-clash/v1/roster/{rosterId}`
- `POST` `/lol-clash/v1/roster/{rosterId}/accept`
- `POST` `/lol-clash/v1/roster/{rosterId}/cancel-withdraw`
- `POST` `/lol-clash/v1/roster/{rosterId}/change-all-details`
- `POST` `/lol-clash/v1/roster/{rosterId}/change-icon`
- `POST` `/lol-clash/v1/roster/{rosterId}/change-name`
- `POST` `/lol-clash/v1/roster/{rosterId}/change-short-name`
- `POST` `/lol-clash/v1/roster/{rosterId}/decline`
- `POST` `/lol-clash/v1/roster/{rosterId}/disband`
- `POST` `/lol-clash/v1/roster/{rosterId}/invite`
- `POST` `/lol-clash/v1/roster/{rosterId}/kick`
- `POST` `/lol-clash/v1/roster/{rosterId}/leave`
- `POST` `/lol-clash/v1/roster/{rosterId}/lockin`
- `POST` `/lol-clash/v1/roster/{rosterId}/set-position`
- `POST` `/lol-clash/v1/roster/{rosterId}/set-ticket`
- `GET` `/lol-clash/v1/roster/{rosterId}/stats`
- `POST` `/lol-clash/v1/roster/{rosterId}/suggest`
- `POST` `/lol-clash/v1/roster/{rosterId}/suggest/{summonerId}/accept`
- `POST` `/lol-clash/v1/roster/{rosterId}/suggest/{summonerId}/decline`
- `POST` `/lol-clash/v1/roster/{rosterId}/suggest/{summonerId}/revoke`
- `POST` `/lol-clash/v1/roster/{rosterId}/ticket-offer/{summonerId}/accept`
- `POST` `/lol-clash/v1/roster/{rosterId}/ticket-offer/{summonerId}/decline`
- `POST` `/lol-clash/v1/roster/{rosterId}/ticket-offer/{summonerId}/offer`
- `POST` `/lol-clash/v1/roster/{rosterId}/ticket-offer/{summonerId}/revoke`
- `POST` `/lol-clash/v1/roster/{rosterId}/transfer-captain`
- `POST` `/lol-clash/v1/roster/{rosterId}/unlockin`
- `POST` `/lol-clash/v1/roster/{rosterId}/unwithdraw`
- `POST` `/lol-clash/v1/roster/{rosterId}/update-logos`
- `POST` `/lol-clash/v1/roster/{rosterId}/withdraw`
- `GET` `/lol-clash/v1/scouting/champions`
- `GET` `/lol-clash/v1/scouting/matchhistory`
- `GET` `/lol-clash/v1/season-rewards/{seasonId}`
- `GET` `/lol-clash/v1/simple-state-flags`
- `POST` `/lol-clash/v1/simple-state-flags/{id}/acknowledge`
- `GET` `/lol-clash/v1/thirdparty/team-data`
- `GET` `/lol-clash/v1/time`
- `GET` `/lol-clash/v1/tournament-state-info`
- `GET` `/lol-clash/v1/tournament-summary`
- `GET` `/lol-clash/v1/tournament/cancelled`
- `GET` `/lol-clash/v1/tournament/get-all-player-tiers`
- `GET` `/lol-clash/v1/tournament/{tournamentId}`
- `POST` `/lol-clash/v1/tournament/{tournamentId}/create-roster`
- `GET` `/lol-clash/v1/tournament/{tournamentId}/get-player-tiers`
- `GET` `/lol-clash/v1/tournament/{tournamentId}/player`
- `GET` `/lol-clash/v1/tournament/{tournamentId}/player-honor-restricted`
- `GET` `/lol-clash/v1/tournament/{tournamentId}/stateInfo`
- `GET` `/lol-clash/v1/tournament/{tournamentId}/winners`
- `POST` `/lol-clash/v1/update-logos`
- `GET` `/lol-clash/v1/visible`
- `DELETE, POST` `/lol-clash/v1/voice`
- `DELETE, POST` `/lol-clash/v1/voice-delay/{delaySeconds}`
- `GET` `/lol-clash/v1/voice-enabled`
- `GET` `/lol-clash/v2/playmode-restricted`

## /lol-client-config

- `GET` `/lol-client-config/v3/client-config/{name}`

## /lol-collections

- `GET` `/lol-collections/v1/inventories/{summonerId}/backdrop`
- `GET` `/lol-collections/v1/inventories/{summonerId}/spells`
- `GET` `/lol-collections/v1/inventories/{summonerId}/ward-skins`
- `GET` `/lol-collections/v1/inventories/{summonerId}/ward-skins/{wardSkinId}`

## /lol-cosmetics

- `GET` `/lol-cosmetics/v1/favorites/tft/companions`
- `GET` `/lol-cosmetics/v1/favorites/tft/damage-skins`
- `GET` `/lol-cosmetics/v1/favorites/tft/map-skins`
- `PUT` `/lol-cosmetics/v1/favorites/tft/save`
- `GET` `/lol-cosmetics/v1/favorites/tft/zoom-skins`
- `DELETE, PUT` `/lol-cosmetics/v1/favorites/tft/{cosmeticType}/{contentId}`
- `GET` `/lol-cosmetics/v1/inventories/{setName}/augment-pillars`
- `GET` `/lol-cosmetics/v1/inventories/{setName}/companions`
- `GET` `/lol-cosmetics/v1/inventories/{setName}/damage-skins`
- `GET` `/lol-cosmetics/v1/inventories/{setName}/map-skins`
- `GET` `/lol-cosmetics/v1/inventories/{setName}/playbooks`
- `GET` `/lol-cosmetics/v1/inventories/{setName}/zoom-skins`
- `PATCH` `/lol-cosmetics/v1/recent/{type}`
- `DELETE, PUT` `/lol-cosmetics/v1/selection/companion`
- `DELETE, PUT` `/lol-cosmetics/v1/selection/playbook`
- `DELETE, PUT` `/lol-cosmetics/v1/selection/tft-augment-pillar`
- `DELETE, PUT` `/lol-cosmetics/v1/selection/tft-damage-skin`
- `DELETE, PUT` `/lol-cosmetics/v1/selection/tft-map-skin`
- `DELETE, PUT` `/lol-cosmetics/v1/selection/tft-zoom-skin`

## /lol-directx-upgrade

- `GET` `/lol-directx-upgrade/needs-hardware-upgrade`
- `POST` `/lol-directx-upgrade/notification-ack`
- `GET` `/lol-directx-upgrade/notification-type`

## /lol-drops

- `GET` `/lol-drops/v1/drop-tables`
- `GET` `/lol-drops/v1/drop-tables/{dropTableId}`
- `GET` `/lol-drops/v1/drop-tables/{dropTableId}/odds-list`
- `GET` `/lol-drops/v1/drop-tables/{dropTableId}/odds-tree`
- `GET` `/lol-drops/v1/drop-tables/{dropTableId}/players/{playerId}/pity-count`
- `GET` `/lol-drops/v1/players/{playerId}/pity-counts`
- `GET` `/lol-drops/v1/players/{playerId}/total-rolls-counts`
- `GET` `/lol-drops/v1/ready`

## /lol-email-verification

- `POST` `/lol-email-verification/v1/confirm-email`
- `GET, PUT` `/lol-email-verification/v1/email`

## /lol-end-of-game

- `GET` `/lol-end-of-game/v1/champion-mastery-updates`
- `GET` `/lol-end-of-game/v1/eog-stats-block`
- `GET, POST` `/lol-end-of-game/v1/gameclient-eog-stats-block`
- `POST` `/lol-end-of-game/v1/state/dismiss-stats`
- `GET` `/lol-end-of-game/v1/tft-eog-stats`

## /lol-esport-stream-notifications

- `GET` `/lol-esport-stream-notifications/v1/live-streams`
- `POST` `/lol-esport-stream-notifications/v1/send-stats`
- `GET` `/lol-esport-stream-notifications/v1/stream-url`

## /lol-event-hub

- `GET` `/lol-event-hub/v1/events`
- `GET` `/lol-event-hub/v1/events/{eventId}/chapters`
- `GET` `/lol-event-hub/v1/events/{eventId}/event-details-data`
- `GET` `/lol-event-hub/v1/events/{eventId}/info`
- `GET` `/lol-event-hub/v1/events/{eventId}/is-grace-period`
- `GET` `/lol-event-hub/v1/events/{eventId}/narrative`
- `GET` `/lol-event-hub/v1/events/{eventId}/objectives-banner`
- `GET` `/lol-event-hub/v1/events/{eventId}/pass-background-data`
- `GET` `/lol-event-hub/v1/events/{eventId}/pass-bundles`
- `GET` `/lol-event-hub/v1/events/{eventId}/progress-info-data`
- `GET` `/lol-event-hub/v1/events/{eventId}/progression-purchase-data`
- `POST` `/lol-event-hub/v1/events/{eventId}/purchase-offer`
- `GET` `/lol-event-hub/v1/events/{eventId}/reward-track/bonus-items`
- `GET` `/lol-event-hub/v1/events/{eventId}/reward-track/bonus-progress`
- `POST` `/lol-event-hub/v1/events/{eventId}/reward-track/claim-all`
- `GET` `/lol-event-hub/v1/events/{eventId}/reward-track/counter`
- `GET` `/lol-event-hub/v1/events/{eventId}/reward-track/failure`
- `GET` `/lol-event-hub/v1/events/{eventId}/reward-track/items`
- `GET` `/lol-event-hub/v1/events/{eventId}/reward-track/progress`
- `GET` `/lol-event-hub/v1/events/{eventId}/reward-track/unclaimed-rewards`
- `GET` `/lol-event-hub/v1/events/{eventId}/reward-track/xp`
- `GET` `/lol-event-hub/v1/events/{eventId}/token-shop`
- `GET` `/lol-event-hub/v1/events/{eventId}/token-shop/categories-offers`
- `GET` `/lol-event-hub/v1/events/{eventId}/token-shop/token-balance`
- `GET` `/lol-event-hub/v1/navigation-button-data`
- `POST` `/lol-event-hub/v1/purchase-item`
- `GET` `/lol-event-hub/v1/skins`
- `GET` `/lol-event-hub/v1/token-upsell`

## /lol-event-mission

- `GET` `/lol-event-mission/v1/event-mission`

## /lol-game-client-chat

- `GET` `/lol-game-client-chat/v1/buddies`
- `GET` `/lol-game-client-chat/v1/ignored-summoners`
- `POST` `/lol-game-client-chat/v1/instant-messages`
- `GET` `/lol-game-client-chat/v1/muted-summoners`
- `GET` `/lol-game-client-chat/v2/buddies`
- `GET` `/lol-game-client-chat/v2/ignored-players`
- `POST` `/lol-game-client-chat/v2/instant-messages`
- `GET` `/lol-game-client-chat/v2/muted-players`

## /lol-game-data-inventory

- `GET` `/lol-game-data-inventory/v1/items/contentIds`
- `GET` `/lol-game-data-inventory/v1/items/contentIds/inventoryTypes/{inventoryType}`
- `GET` `/lol-game-data-inventory/v1/items/contentIds/{contentId}`
- `GET` `/lol-game-data-inventory/v1/items/itemIds/inventoryTypes/{inventoryType}`
- `GET` `/lol-game-data-inventory/v1/items/itemIds/inventoryTypes/{inventoryType}/itemIds/{itemId}`
- `GET` `/lol-game-data-inventory/v1/ready`

## /lol-game-queues

- `GET` `/lol-game-queues/v1/custom`
- `GET` `/lol-game-queues/v1/custom-non-default`
- `GET` `/lol-game-queues/v1/game-type-config/{gameTypeConfigId}`
- `GET` `/lol-game-queues/v1/game-type-config/{gameTypeConfigId}/map/{mapId}`
- `GET` `/lol-game-queues/v1/matchmaking-queues`
- `GET` `/lol-game-queues/v1/queues`
- `GET` `/lol-game-queues/v1/queues/type/{queueType}`
- `GET` `/lol-game-queues/v1/queues/{id}`
- `GET` `/lol-game-queues/v1/queues/{id}/isCustom`

## /lol-game-settings

- `GET` `/lol-game-settings/v1/didreset`
- `GET, PATCH` `/lol-game-settings/v1/game-settings`
- `GET` `/lol-game-settings/v1/game-settings-schema`
- `GET, PATCH` `/lol-game-settings/v1/input-settings`
- `GET` `/lol-game-settings/v1/input-settings-schema`
- `GET` `/lol-game-settings/v1/ready`
- `POST` `/lol-game-settings/v1/reload-post-game`
- `POST` `/lol-game-settings/v1/save`

## /lol-gameflow

- `POST` `/lol-gameflow/v1/ack-failed-to-launch`
- `GET` `/lol-gameflow/v1/active-patcher-lock`
- `GET` `/lol-gameflow/v1/availability`
- `GET` `/lol-gameflow/v1/basic-tutorial`
- `POST` `/lol-gameflow/v1/basic-tutorial/start`
- `GET` `/lol-gameflow/v1/battle-training`
- `POST` `/lol-gameflow/v1/battle-training/start`
- `POST` `/lol-gameflow/v1/battle-training/stop`
- `POST` `/lol-gameflow/v1/early-exit`
- `GET` `/lol-gameflow/v1/early-exit-enabled`
- `DELETE, GET` `/lol-gameflow/v1/early-exit-notifications/eog`
- `DELETE` `/lol-gameflow/v1/early-exit-notifications/eog/{key}`
- `DELETE, GET` `/lol-gameflow/v1/early-exit-notifications/missions`
- `DELETE` `/lol-gameflow/v1/early-exit-notifications/missions/{key}`
- `GET` `/lol-gameflow/v1/early-exit-quit-enabled`
- `GET, POST` `/lol-gameflow/v1/extra-game-client-args`
- `GET` `/lol-gameflow/v1/game-exit-early-vanguard`
- `GET, POST` `/lol-gameflow/v1/gameflow-metadata/player-status`
- `GET, POST` `/lol-gameflow/v1/gameflow-metadata/registration-status`
- `POST` `/lol-gameflow/v1/gameflow-monitor`
- `GET` `/lol-gameflow/v1/gameflow-phase`
- `GET` `/lol-gameflow/v1/player-kicked-vanguard`
- `POST` `/lol-gameflow/v1/pre-end-game-transition`
- `POST` `/lol-gameflow/v1/reconnect`
- `GET` `/lol-gameflow/v1/session`
- `POST` `/lol-gameflow/v1/session/champ-select/phase-time-remaining`
- `POST` `/lol-gameflow/v1/session/dodge`
- `POST` `/lol-gameflow/v1/session/event`
- `POST` `/lol-gameflow/v1/session/game-configuration`
- `GET` `/lol-gameflow/v1/session/per-position-summoner-spells/disallowed`
- `GET` `/lol-gameflow/v1/session/per-position-summoner-spells/disallowed/as-string`
- `GET` `/lol-gameflow/v1/session/per-position-summoner-spells/required`
- `GET` `/lol-gameflow/v1/session/per-position-summoner-spells/required/as-string`
- `POST` `/lol-gameflow/v1/session/request-enter-gameflow`
- `POST` `/lol-gameflow/v1/session/request-lobby`
- `POST` `/lol-gameflow/v1/session/request-tournament-checkin`
- `POST` `/lol-gameflow/v1/session/tournament-ended`
- `GET` `/lol-gameflow/v1/spectate`
- `GET` `/lol-gameflow/v1/spectate/delayed-launch`
- `POST` `/lol-gameflow/v1/spectate/launch`
- `POST` `/lol-gameflow/v1/spectate/quit`
- `POST` `/lol-gameflow/v1/tick`
- `GET` `/lol-gameflow/v1/watch`
- `POST` `/lol-gameflow/v1/watch/launch`
- `POST` `/lol-gameflow/v2/spectate/launch`

## /lol-heartbeat

- `POST` `/lol-heartbeat/v1/connection-status`

## /lol-highlights

- `GET` `/lol-highlights/v1/config`
- `POST` `/lol-highlights/v1/file-browser/{highlightId}`
- `GET, POST` `/lol-highlights/v1/highlights`
- `GET` `/lol-highlights/v1/highlights-folder-path`
- `GET` `/lol-highlights/v1/highlights-folder-path/default`
- `DELETE, GET, PUT` `/lol-highlights/v1/highlights/{id}`

## /lol-honeyfruit

- `GET, POST` `/lol-honeyfruit/v1/vng-publisher-settings`

## /lol-honor

- `POST` `/lol-honor/v1/ballot`
- `POST` `/lol-honor/v1/honor`

## /lol-honor-v2

- `POST` `/lol-honor-v2/v1/ack-honor-notification/{mailId}`
- `DELETE, GET` `/lol-honor-v2/v1/ballot`
- `POST` `/lol-honor-v2/v1/ballot/refresh`
- `GET` `/lol-honor-v2/v1/config`
- `POST` `/lol-honor-v2/v1/honor-player`
- `GET` `/lol-honor-v2/v1/late-recognition`
- `POST` `/lol-honor-v2/v1/late-recognition/ack`
- `GET` `/lol-honor-v2/v1/latest-eligible-game`
- `GET` `/lol-honor-v2/v1/level-change`
- `GET` `/lol-honor-v2/v1/level-change-notifications`
- `POST` `/lol-honor-v2/v1/level-change/ack`
- `GET` `/lol-honor-v2/v1/mutual-honor`
- `POST` `/lol-honor-v2/v1/mutual-honor/ack`
- `GET` `/lol-honor-v2/v1/profile`
- `GET` `/lol-honor-v2/v1/recognition`
- `GET` `/lol-honor-v2/v1/recognition-history`
- `GET` `/lol-honor-v2/v1/reward-granted`
- `POST` `/lol-honor-v2/v1/reward-granted/ack`
- `GET` `/lol-honor-v2/v1/team-choices`
- `GET` `/lol-honor-v2/v1/vote-completion`

## /lol-hovercard

- `GET` `/lol-hovercard/v1/friend-info/{puuid}`

## /lol-inventory

- `GET` `/lol-inventory/v1/aramInventory`
- `GET` `/lol-inventory/v1/champSelectInventory`
- `GET` `/lol-inventory/v1/cherryInventory`
- `GET` `/lol-inventory/v1/initial-configuration-complete`
- `GET` `/lol-inventory/v1/inventory`
- `GET` `/lol-inventory/v1/inventory/emotes`
- `GET` `/lol-inventory/v1/inventoryWithF2P`
- `POST` `/lol-inventory/v1/notification/acknowledge`
- `GET` `/lol-inventory/v1/notifications/{inventoryType}`
- `GET` `/lol-inventory/v1/players/{puuid}/inventory`
- `GET` `/lol-inventory/v1/signedInventory`
- `GET` `/lol-inventory/v1/signedInventory/simple`
- `GET` `/lol-inventory/v1/signedInventory/tournamentlogos`
- `GET` `/lol-inventory/v1/signedInventoryCache`
- `GET` `/lol-inventory/v1/signedWallet`
- `GET` `/lol-inventory/v1/signedWallet/{currencyType}`
- `GET` `/lol-inventory/v1/strawberryInventory`
- `GET` `/lol-inventory/v1/wallet`
- `GET` `/lol-inventory/v1/wallet/{currencyType}`
- `GET` `/lol-inventory/v1/xbox-subscription-status`
- `GET` `/lol-inventory/v2/inventory/{inventoryType}`

## /lol-item-sets

- `POST` `/lol-item-sets/v1/item-sets/validate`
- `GET, POST, PUT` `/lol-item-sets/v1/item-sets/{summonerId}/sets`

## /lol-kickout

- `GET` `/lol-kickout/v1/notification`

## /lol-kr-playtime-reminder

- `GET` `/lol-kr-playtime-reminder/v1/hours-played`

## /lol-kr-shutdown-law

- `GET` `/lol-kr-shutdown-law/v1/custom-status`
- `GET` `/lol-kr-shutdown-law/v1/disabled-queues`
- `GET` `/lol-kr-shutdown-law/v1/is-enabled`
- `GET` `/lol-kr-shutdown-law/v1/notification`
- `GET` `/lol-kr-shutdown-law/v1/queue-status/{queue_id}`
- `GET` `/lol-kr-shutdown-law/v1/rating-screen`
- `POST` `/lol-kr-shutdown-law/v1/rating-screen/acknowledge`
- `GET` `/lol-kr-shutdown-law/v1/status`

## /lol-leaderboard

- `GET` `/lol-leaderboard/v1/name/{name}/currentSeason`
- `GET` `/lol-leaderboard/v1/name/{name}/groupings`
- `GET` `/lol-leaderboard/v1/name/{name}/pageSize`
- `GET` `/lol-leaderboard/v1/name/{name}/season/{season}/grouping/{grouping}`
- `GET` `/lol-leaderboard/v1/name/{name}/season/{season}/grouping/{grouping}/entry-ranking/{entityId}`
- `GET` `/lol-leaderboard/v1/name/{name}/season/{season}/grouping/{grouping}/jump-to-entry/{entityId}`
- `GET` `/lol-leaderboard/v1/name/{name}/season/{season}/grouping/{grouping}/multipage`
- `GET` `/lol-leaderboard/v1/name/{name}/timer`
- `GET` `/lol-leaderboard/v1/ready`

## /lol-league-session

- `GET` `/lol-league-session/v1/league-session-token`

## /lol-leaver-buster

- `GET` `/lol-leaver-buster/v1/notifications`
- `DELETE, GET` `/lol-leaver-buster/v1/notifications/{id}`
- `GET` `/lol-leaver-buster/v1/ranked-restriction`

## /lol-license-agreement

- `GET` `/lol-license-agreement/v1/agreement`
- `GET` `/lol-license-agreement/v1/privacy-policy`

## /lol-loadouts

- `GET` `/lol-loadouts/v1/loadouts-ready`
- `GET` `/lol-loadouts/v4/loadouts/scope/account`
- `GET` `/lol-loadouts/v4/loadouts/scope/{scope}/{scopeItemId}`
- `DELETE, PATCH, PUT` `/lol-loadouts/v4/loadouts/{id}`
- `GET` `/lol-loadouts/v4/loadouts/{loadoutId}`

## /lol-lobby

- `GET, PUT` `/lol-lobby/v1/autofill-displayed`
- `DELETE, POST` `/lol-lobby/v1/clash`
- `GET` `/lol-lobby/v1/custom-games`
- `POST` `/lol-lobby/v1/custom-games/refresh`
- `GET` `/lol-lobby/v1/custom-games/{id}`
- `POST` `/lol-lobby/v1/custom-games/{id}/join`
- `GET` `/lol-lobby/v1/lobby/availability`
- `GET` `/lol-lobby/v1/lobby/countdown`
- `POST` `/lol-lobby/v1/lobby/custom/bots`
- `POST` `/lol-lobby/v1/lobby/custom/bots/{summonerInternalName}/{botUuidToDelete}`
- `DELETE` `/lol-lobby/v1/lobby/custom/bots/{summonerInternalName}/{botUuidToDelete}/{teamId}`
- `POST` `/lol-lobby/v1/lobby/custom/cancel-champ-select`
- `POST` `/lol-lobby/v1/lobby/custom/start-champ-select`
- `GET, POST` `/lol-lobby/v1/lobby/invitations`
- `GET` `/lol-lobby/v1/lobby/invitations/{id}`
- `GET, PUT` `/lol-lobby/v1/lobby/members/localMember/player-slots`
- `POST` `/lol-lobby/v1/lobby/members/localMember/player-slots/{slotsIndex}/{perksString}`
- `PUT` `/lol-lobby/v1/lobby/members/localMember/position-preferences`
- `PUT` `/lol-lobby/v1/parties/active`
- `GET` `/lol-lobby/v1/parties/gamemode`
- `PUT` `/lol-lobby/v1/parties/metadata`
- `GET` `/lol-lobby/v1/parties/player`
- `PUT` `/lol-lobby/v1/parties/queue`
- `PUT` `/lol-lobby/v1/parties/ready`
- `PUT` `/lol-lobby/v1/parties/{partyId}/members/{puuid}/role`
- `GET` `/lol-lobby/v1/party-rewards`
- `POST` `/lol-lobby/v1/tournaments/{id}/join`
- `GET` `/lol-lobby/v2/comms/members`
- `GET` `/lol-lobby/v2/comms/token`
- `GET` `/lol-lobby/v2/eligibility/game-select-eligibility-hash`
- `GET` `/lol-lobby/v2/eligibility/initial-configuration-complete`
- `POST` `/lol-lobby/v2/eligibility/party`
- `POST` `/lol-lobby/v2/eligibility/self`
- `POST` `/lol-lobby/v2/eog-invitations`
- `DELETE, GET, POST` `/lol-lobby/v2/lobby`
- `GET` `/lol-lobby/v2/lobby/custom/available-bots`
- `GET` `/lol-lobby/v2/lobby/custom/bots-enabled`
- `GET, POST` `/lol-lobby/v2/lobby/invitations`
- `POST` `/lol-lobby/v2/lobby/invitationsWithContext`
- `DELETE, POST` `/lol-lobby/v2/lobby/matchmaking/search`
- `GET` `/lol-lobby/v2/lobby/matchmaking/search-state`
- `PUT` `/lol-lobby/v2/lobby/memberData`
- `GET` `/lol-lobby/v2/lobby/members`
- `PUT` `/lol-lobby/v2/lobby/members/localMember/position-preferences`
- `POST` `/lol-lobby/v2/lobby/members/{summonerId}/grant-invite`
- `POST` `/lol-lobby/v2/lobby/members/{summonerId}/kick`
- `POST` `/lol-lobby/v2/lobby/members/{summonerId}/promote`
- `POST` `/lol-lobby/v2/lobby/members/{summonerId}/revoke-invite`
- `PUT` `/lol-lobby/v2/lobby/partyType`
- `PUT` `/lol-lobby/v2/lobby/quickplayMemberData`
- `PUT` `/lol-lobby/v2/lobby/strawberryMapId`
- `PUT` `/lol-lobby/v2/lobby/subteamData`
- `POST` `/lol-lobby/v2/lobby/team/{team}`
- `POST` `/lol-lobby/v2/matchmaking/quick-search`
- `GET, POST` `/lol-lobby/v2/notifications`
- `DELETE` `/lol-lobby/v2/notifications/{notificationId}`
- `POST` `/lol-lobby/v2/parties/overrides/EnabledForTeamBuilderQueues`
- `GET` `/lol-lobby/v2/party-active`
- `GET` `/lol-lobby/v2/party/eog-status`
- `POST` `/lol-lobby/v2/party/{partyId}/join`
- `POST` `/lol-lobby/v2/play-again`
- `POST` `/lol-lobby/v2/play-again-decline`
- `GET` `/lol-lobby/v2/received-invitations`
- `POST` `/lol-lobby/v2/received-invitations/{invitationId}/accept`
- `POST` `/lol-lobby/v2/received-invitations/{invitationId}/decline`
- `GET` `/lol-lobby/v2/registration-status`
- `POST` `/lol-lobby/v3/lobby/members/{puuid}/grant-invite`
- `POST` `/lol-lobby/v3/lobby/members/{puuid}/kick`
- `POST` `/lol-lobby/v3/lobby/members/{puuid}/promote`
- `POST` `/lol-lobby/v3/lobby/members/{puuid}/revoke-invite`

## /lol-lobby-team-builder

- `GET` `/lol-lobby-team-builder/champ-select/v1/bannable-champion-ids`
- `GET` `/lol-lobby-team-builder/champ-select/v1/crowd-favorte-champion-list`
- `GET` `/lol-lobby-team-builder/champ-select/v1/current-champion`
- `GET` `/lol-lobby-team-builder/champ-select/v1/disabled-champion-ids`
- `GET` `/lol-lobby-team-builder/champ-select/v1/has-auto-assigned-smite`
- `GET` `/lol-lobby-team-builder/champ-select/v1/implementation-active`
- `GET` `/lol-lobby-team-builder/champ-select/v1/pickable-champion-ids`
- `GET` `/lol-lobby-team-builder/champ-select/v1/pickable-skin-ids`
- `GET` `/lol-lobby-team-builder/champ-select/v1/preferences`
- `POST` `/lol-lobby-team-builder/champ-select/v1/retrieve-latest-game-dto`
- `GET` `/lol-lobby-team-builder/champ-select/v1/session`
- `PATCH` `/lol-lobby-team-builder/champ-select/v1/session/actions/{id}`
- `POST` `/lol-lobby-team-builder/champ-select/v1/session/actions/{id}/complete`
- `POST` `/lol-lobby-team-builder/champ-select/v1/session/bench/swap/{championId}`
- `GET` `/lol-lobby-team-builder/champ-select/v1/session/champion-swaps`
- `GET` `/lol-lobby-team-builder/champ-select/v1/session/champion-swaps/{id}`
- `POST` `/lol-lobby-team-builder/champ-select/v1/session/champion-swaps/{id}/accept`
- `POST` `/lol-lobby-team-builder/champ-select/v1/session/champion-swaps/{id}/cancel`
- `POST` `/lol-lobby-team-builder/champ-select/v1/session/champion-swaps/{id}/decline`
- `POST` `/lol-lobby-team-builder/champ-select/v1/session/champion-swaps/{id}/request`
- `GET, PATCH` `/lol-lobby-team-builder/champ-select/v1/session/my-selection`
- `POST` `/lol-lobby-team-builder/champ-select/v1/session/my-selection/reroll`
- `GET` `/lol-lobby-team-builder/champ-select/v1/session/obfuscated-puuids`
- `GET` `/lol-lobby-team-builder/champ-select/v1/session/obfuscated-summoner-ids`
- `GET` `/lol-lobby-team-builder/champ-select/v1/session/pick-order-swaps`
- `GET` `/lol-lobby-team-builder/champ-select/v1/session/pick-order-swaps/{id}`
- `POST` `/lol-lobby-team-builder/champ-select/v1/session/pick-order-swaps/{id}/accept`
- `POST` `/lol-lobby-team-builder/champ-select/v1/session/pick-order-swaps/{id}/cancel`
- `POST` `/lol-lobby-team-builder/champ-select/v1/session/pick-order-swaps/{id}/decline`
- `POST` `/lol-lobby-team-builder/champ-select/v1/session/pick-order-swaps/{id}/request`
- `GET` `/lol-lobby-team-builder/champ-select/v1/session/position-swaps`
- `GET` `/lol-lobby-team-builder/champ-select/v1/session/position-swaps/{id}`
- `POST` `/lol-lobby-team-builder/champ-select/v1/session/position-swaps/{id}/accept`
- `POST` `/lol-lobby-team-builder/champ-select/v1/session/position-swaps/{id}/cancel`
- `POST` `/lol-lobby-team-builder/champ-select/v1/session/position-swaps/{id}/decline`
- `POST` `/lol-lobby-team-builder/champ-select/v1/session/position-swaps/{id}/request`
- `POST` `/lol-lobby-team-builder/champ-select/v1/session/quit`
- `GET` `/lol-lobby-team-builder/champ-select/v1/session/timer`
- `POST` `/lol-lobby-team-builder/champ-select/v1/simple-inventory`
- `GET` `/lol-lobby-team-builder/champ-select/v1/subset-champion-list`
- `GET` `/lol-lobby-team-builder/champ-select/v1/team-boost`
- `POST` `/lol-lobby-team-builder/champ-select/v1/team-boost/purchase`
- `GET` `/lol-lobby-team-builder/v1/matchmaking`
- `POST` `/lol-lobby-team-builder/v1/ready-check/accept`
- `POST` `/lol-lobby-team-builder/v1/ready-check/decline`
- `GET` `/lol-lobby-team-builder/v1/ready-check/isAutoAccept`

## /lol-lock-and-load

- `GET` `/lol-lock-and-load/v1/home-hubs-waits`
- `GET` `/lol-lock-and-load/v1/should-show-progress-bar-text`
- `GET` `/lol-lock-and-load/v1/should-wait-for-home-hubs`

## /lol-login

- `GET, POST` `/lol-login/v1/account-state`
- `POST` `/lol-login/v1/change-summoner-name`
- `POST` `/lol-login/v1/delete-rso-on-close`
- `POST` `/lol-login/v1/leagueSessionStatus`
- `GET` `/lol-login/v1/login-connection-state`
- `GET` `/lol-login/v1/login-data-packet`
- `GET` `/lol-login/v1/login-in-game-creds`
- `GET` `/lol-login/v1/login-platform-credentials`
- `GET` `/lol-login/v1/login-queue-state`
- `DELETE, POST` `/lol-login/v1/service-proxy-async-requests/{serviceName}/{methodName}`
- `POST` `/lol-login/v1/service-proxy-uuid-requests`
- `DELETE, GET` `/lol-login/v1/session`
- `POST` `/lol-login/v1/session/invoke`
- `DELETE, PUT` `/lol-login/v1/shutdown-locks/{lockName}`
- `POST` `/lol-login/v1/summoner-session`
- `POST` `/lol-login/v1/summoner-session-failed`
- `GET` `/lol-login/v1/wallet`
- `GET` `/lol-login/v2/league-session-init-token`

## /lol-loot

- `POST` `/lol-loot/v1/craft/mass`
- `GET` `/lol-loot/v1/currency-configuration`
- `GET` `/lol-loot/v1/enabled`
- `GET` `/lol-loot/v1/loot-grants`
- `DELETE` `/lol-loot/v1/loot-grants/{id}`
- `GET` `/lol-loot/v1/loot-items`
- `GET` `/lol-loot/v1/loot-items/{lootName}`
- `PUT` `/lol-loot/v1/loot-odds/evaluateQuery`
- `GET` `/lol-loot/v1/loot-odds/{recipeName}`
- `GET` `/lol-loot/v1/loot-odds/{recipeName}/visibility`
- `GET` `/lol-loot/v1/mass-disenchant-recipes`
- `GET` `/lol-loot/v1/mass-disenchant/configuration`
- `GET` `/lol-loot/v1/milestones`
- `GET` `/lol-loot/v1/milestones/counters`
- `GET` `/lol-loot/v1/milestones/items`
- `GET` `/lol-loot/v1/milestones/{lootMilestonesId}`
- `POST` `/lol-loot/v1/milestones/{lootMilestonesId}/claim`
- `GET` `/lol-loot/v1/milestones/{lootMilestonesId}/claimProgress`
- `GET` `/lol-loot/v1/milestones/{lootMilestonesId}/counter`
- `GET` `/lol-loot/v1/player-display-categories`
- `GET` `/lol-loot/v1/player-loot`
- `GET` `/lol-loot/v1/player-loot-map`
- `GET` `/lol-loot/v1/player-loot-notifications`
- `POST` `/lol-loot/v1/player-loot-notifications/{id}/acknowledge`
- `GET` `/lol-loot/v1/player-loot/{lootId}`
- `GET, POST` `/lol-loot/v1/player-loot/{lootId}/context-menu`
- `DELETE` `/lol-loot/v1/player-loot/{lootId}/new-notification`
- `POST` `/lol-loot/v1/player-loot/{lootName}/redeem`
- `GET` `/lol-loot/v1/ready`
- `GET` `/lol-loot/v1/recipes/configuration`
- `GET, POST` `/lol-loot/v1/recipes/initial-item/{lootId}`
- `POST` `/lol-loot/v1/recipes/{recipeName}/craft`
- `POST` `/lol-loot/v1/refresh`
- `GET` `/lol-loot/v2/player-loot-map`

## /lol-loyalty

- `GET` `/lol-loyalty/v1/status-notification`
- `POST` `/lol-loyalty/v1/updateLoyaltyInventory`

## /lol-mac-graphics-upgrade

- `GET` `/lol-mac-graphics-upgrade/needs-hardware-upgrade`
- `GET` `/lol-mac-graphics-upgrade/notification-type`

## /lol-maps

- `POST` `/lol-maps/v1/map`
- `GET` `/lol-maps/v1/map/{id}`
- `GET` `/lol-maps/v1/maps`
- `GET` `/lol-maps/v2/map/{id}/{gameMode}`
- `GET` `/lol-maps/v2/map/{id}/{gameMode}/{gameMutator}`
- `GET` `/lol-maps/v2/maps`

## /lol-marketing-preferences

- `GET, POST` `/lol-marketing-preferences/v1/partition/{partitionKey}`
- `GET` `/lol-marketing-preferences/v1/ready`

## /lol-marketplace

- `GET` `/lol-marketplace/v1/products/tft/configs`
- `GET` `/lol-marketplace/v1/products/tft/stores/upgrades`
- `GET, POST` `/lol-marketplace/v1/products/{product}/purchases`
- `POST` `/lol-marketplace/v1/products/{product}/refunds`
- `GET` `/lol-marketplace/v1/products/{product}/stores`
- `GET` `/lol-marketplace/v1/purchases/{purchaseId}`
- `GET` `/lol-marketplace/v1/ready`

## /lol-match-history

- `GET` `/lol-match-history/v1/game-timelines/{gameId}`
- `GET` `/lol-match-history/v1/games/{gameId}`
- `GET` `/lol-match-history/v1/products/lol/current-summoner/matches`
- `GET` `/lol-match-history/v1/products/lol/{puuid}/matches`
- `GET` `/lol-match-history/v1/products/tft/{puuid}/matches`
- `GET` `/lol-match-history/v1/recently-played-summoners`

## /lol-matchmaking

- `GET` `/lol-matchmaking/v1/ready-check`
- `POST` `/lol-matchmaking/v1/ready-check/accept`
- `POST` `/lol-matchmaking/v1/ready-check/decline`
- `DELETE, GET, POST, PUT` `/lol-matchmaking/v1/search`
- `GET` `/lol-matchmaking/v1/search/errors`
- `GET` `/lol-matchmaking/v1/search/errors/{id}`

## /lol-metagames

- `POST` `/lol-metagames/v1/player-events/{metagameId}/{eventName}`
- `GET` `/lol-metagames/v1/purchases/{purchaseId}`
- `GET` `/lol-metagames/v1/ready`
- `PUT` `/lol-metagames/v1/{metagameId}/player-data`

## /lol-missions

- `GET` `/lol-missions/v1/data`
- `POST` `/lol-missions/v1/force`
- `GET` `/lol-missions/v1/missions`
- `GET` `/lol-missions/v1/missions/multiseries`
- `GET` `/lol-missions/v1/missions/series/{seriesName}`
- `GET` `/lol-missions/v1/missions/seriesId`
- `GET` `/lol-missions/v1/missions/seriesName`
- `PUT` `/lol-missions/v1/player`
- `PUT` `/lol-missions/v1/player/{missionId}`
- `GET` `/lol-missions/v1/series`
- `PUT` `/lol-missions/v2/player/opt`

## /lol-nacho

- `GET` `/lol-nacho/v1/banner-odds`
- `GET, POST` `/lol-nacho/v1/banner/active`
- `GET` `/lol-nacho/v1/banners`
- `GET` `/lol-nacho/v1/counters/{counterId}`
- `GET` `/lol-nacho/v1/get-active-store-catalog`
- `GET` `/lol-nacho/v1/get-current-catalog-item`
- `POST` `/lol-nacho/v1/purchase/blessing-token`
- `POST` `/lol-nacho/v1/purchase/roll`
- `GET` `/lol-nacho/v1/purchases/{purchaseId}`
- `GET` `/lol-nacho/v1/ready`
- `GET` `/lol-nacho/v1/roll-purchase-enabled`
- `POST` `/lol-nacho/v1/set-product-id`

## /lol-npe-rewards

- `POST` `/lol-npe-rewards/v1/challenges/opt`
- `GET` `/lol-npe-rewards/v1/challenges/progress`
- `GET` `/lol-npe-rewards/v1/get_poro_experiments`
- `GET` `/lol-npe-rewards/v1/level-rewards`
- `GET` `/lol-npe-rewards/v1/level-rewards/state`
- `GET` `/lol-npe-rewards/v1/login-rewards`
- `GET` `/lol-npe-rewards/v1/login-rewards/state`

## /lol-npe-tutorial-path

- `GET` `/lol-npe-tutorial-path/v1/rewards/champ`
- `GET, PUT` `/lol-npe-tutorial-path/v1/settings`
- `GET` `/lol-npe-tutorial-path/v1/tutorials`
- `PATCH` `/lol-npe-tutorial-path/v1/tutorials/init`
- `PUT` `/lol-npe-tutorial-path/v1/tutorials/{tutorialId}/view`

## /lol-objectives

- `GET` `/lol-objectives/v1/missions-by-tag`
- `GET` `/lol-objectives/v1/objectives/{gameType}`
- `GET` `/lol-objectives/v1/ready`

## /lol-patch

- `GET` `/lol-patch/v1/checking-enabled`
- `GET` `/lol-patch/v1/environment`
- `PUT` `/lol-patch/v1/game-patch-url`
- `GET` `/lol-patch/v1/game-version`
- `GET` `/lol-patch/v1/notifications`
- `DELETE` `/lol-patch/v1/notifications/{id}`
- `GET` `/lol-patch/v1/product-integration/app-update/available`
- `POST` `/lol-patch/v1/products/league_of_legends/detect-corruption-request`
- `GET` `/lol-patch/v1/products/league_of_legends/install-location`
- `POST` `/lol-patch/v1/products/league_of_legends/partial-repair-request`
- `POST` `/lol-patch/v1/products/league_of_legends/start-checking-request`
- `POST` `/lol-patch/v1/products/league_of_legends/start-patching-request`
- `GET` `/lol-patch/v1/products/league_of_legends/state`
- `POST` `/lol-patch/v1/products/league_of_legends/stop-checking-request`
- `POST` `/lol-patch/v1/products/league_of_legends/stop-patching-request`
- `GET` `/lol-patch/v1/products/league_of_legends/supported-game-releases`
- `GET` `/lol-patch/v1/status`
- `PUT` `/lol-patch/v1/ux`

## /lol-perks

- `GET, PUT` `/lol-perks/v1/currentpage`
- `GET` `/lol-perks/v1/inventory`
- `DELETE, GET, POST` `/lol-perks/v1/pages`
- `PUT` `/lol-perks/v1/pages/validate`
- `DELETE, GET, PUT` `/lol-perks/v1/pages/{id}`
- `DELETE` `/lol-perks/v1/pages/{id}/auto-modified-selections`
- `GET` `/lol-perks/v1/perks`
- `PUT` `/lol-perks/v1/perks/ack-gameplay-updated`
- `GET` `/lol-perks/v1/perks/disabled`
- `GET` `/lol-perks/v1/perks/gameplay-updated`
- `GET` `/lol-perks/v1/quick-play-selections/champion/{championId}/position/{position}`
- `GET` `/lol-perks/v1/recommended-champion-positions`
- `GET` `/lol-perks/v1/recommended-pages-position/champion/{championId}`
- `POST` `/lol-perks/v1/recommended-pages-position/champion/{championId}/position/{position}`
- `GET` `/lol-perks/v1/recommended-pages/champion/{championId}/position/{position}/map/{mapId}`
- `GET, POST` `/lol-perks/v1/rune-recommender-auto-select`
- `GET, PUT` `/lol-perks/v1/settings`
- `GET, POST` `/lol-perks/v1/show-auto-modified-pages-notification`
- `GET` `/lol-perks/v1/styles`
- `POST` `/lol-perks/v1/update-page-order`

## /lol-pft

- `POST` `/lol-pft/v2/events`
- `GET, POST` `/lol-pft/v2/survey`

## /lol-platform-config

- `GET` `/lol-platform-config/v1/initial-configuration-complete`
- `GET` `/lol-platform-config/v1/namespaces`
- `GET` `/lol-platform-config/v1/namespaces/{ns}`
- `GET` `/lol-platform-config/v1/namespaces/{ns}/{key}`

## /lol-player-behavior

- `PUT` `/lol-player-behavior/v1/ack-credibility-behavior-warning/{mailId}`
- `GET` `/lol-player-behavior/v1/ban`
- `GET` `/lol-player-behavior/v1/chat-restriction`
- `DELETE, GET` `/lol-player-behavior/v1/code-of-conduct-notification`
- `GET` `/lol-player-behavior/v1/config`
- `GET` `/lol-player-behavior/v1/credibility-behavior-warnings`
- `GET` `/lol-player-behavior/v1/reform-card`
- `GET` `/lol-player-behavior/v1/reporter-feedback`
- `DELETE, GET` `/lol-player-behavior/v1/reporter-feedback/{id}`
- `GET` `/lol-player-behavior/v2/reform-card`
- `GET` `/lol-player-behavior/v2/reporter-feedback`
- `POST` `/lol-player-behavior/v2/reporter-feedback/{key}`
- `PUT` `/lol-player-behavior/v3/reform-card/{id}`
- `GET` `/lol-player-behavior/v3/reform-cards`

## /lol-player-messaging

- `GET` `/lol-player-messaging/v1/celebration/notification`
- `DELETE` `/lol-player-messaging/v1/celebration/notification/{id}/acknowledge`
- `GET` `/lol-player-messaging/v1/notification`
- `DELETE` `/lol-player-messaging/v1/notification/{id}/acknowledge`

## /lol-player-preferences

- `POST` `/lol-player-preferences/v1/player-preferences-endpoint-override`
- `GET` `/lol-player-preferences/v1/player-preferences-ready`
- `PUT` `/lol-player-preferences/v1/preference`
- `GET` `/lol-player-preferences/v1/preference/{type}`

## /lol-player-report-sender

- `POST` `/lol-player-report-sender/v1/champ-select-reports`
- `POST` `/lol-player-report-sender/v1/end-of-game-reports`
- `GET` `/lol-player-report-sender/v1/game-ids-with-verbal-abuse-report`
- `GET, POST` `/lol-player-report-sender/v1/in-game-reports`
- `POST` `/lol-player-report-sender/v1/match-history-reports`
- `DELETE, GET` `/lol-player-report-sender/v1/reported-players/gameId/{gameId}`

## /lol-pre-end-of-game

- `POST` `/lol-pre-end-of-game/v1/complete/{sequenceEventName}`
- `GET` `/lol-pre-end-of-game/v1/currentSequenceEvent`
- `DELETE` `/lol-pre-end-of-game/v1/registration/{sequenceEventName}`
- `POST` `/lol-pre-end-of-game/v1/registration/{sequenceEventName}/{priority}`
- `POST` `/lol-pre-end-of-game/v1/skip-pre-end-of-game`

## /lol-premade-voice

- `GET` `/lol-premade-voice/v1/availability`
- `GET, PUT` `/lol-premade-voice/v1/capturedevices`
- `GET` `/lol-premade-voice/v1/devices/capture/permission`
- `GET` `/lol-premade-voice/v1/first-experience`
- `POST` `/lol-premade-voice/v1/first-experience/game`
- `POST` `/lol-premade-voice/v1/first-experience/lcu`
- `POST` `/lol-premade-voice/v1/first-experience/reset`
- `POST` `/lol-premade-voice/v1/gameClientUpdatedPTTKey`
- `DELETE, GET, POST` `/lol-premade-voice/v1/mic-test`
- `GET` `/lol-premade-voice/v1/participant-records`
- `GET` `/lol-premade-voice/v1/participants`
- `PUT` `/lol-premade-voice/v1/participants/{puuid}/mute`
- `PUT` `/lol-premade-voice/v1/participants/{puuid}/volume`
- `POST` `/lol-premade-voice/v1/push-to-talk/check-available`
- `PUT` `/lol-premade-voice/v1/self/activationSensitivity`
- `PUT` `/lol-premade-voice/v1/self/inputMode`
- `PUT` `/lol-premade-voice/v1/self/micLevel`
- `PUT` `/lol-premade-voice/v1/self/mute`
- `DELETE, POST` `/lol-premade-voice/v1/session`
- `GET` `/lol-premade-voice/v1/settings`
- `POST` `/lol-premade-voice/v1/settings/reset`

## /lol-progression

- `GET` `/lol-progression/v1/groups/configuration`
- `GET` `/lol-progression/v1/groups/{groupId}/configuration`
- `GET` `/lol-progression/v1/groups/{groupId}/instanceData`
- `GET` `/lol-progression/v1/ready`

## /lol-publishing-content

- `GET` `/lol-publishing-content/v1/listeners/allow-list/{region}`
- `GET` `/lol-publishing-content/v1/listeners/client-data`
- `GET` `/lol-publishing-content/v1/listeners/pubhub-config`
- `GET` `/lol-publishing-content/v1/ready`
- `GET` `/lol-publishing-content/v1/settings`
- `GET` `/lol-publishing-content/v1/tft-hub-cards`

## /lol-purchase-widget

- `GET` `/lol-purchase-widget/v1/configuration`
- `GET` `/lol-purchase-widget/v1/items/{itemId}/related-bundles`
- `GET` `/lol-purchase-widget/v1/order-notifications`
- `GET` `/lol-purchase-widget/v1/purchasable-item`
- `POST` `/lol-purchase-widget/v1/purchasable-items/{inventoryType}`
- `POST` `/lol-purchase-widget/v2/purchaseItems`
- `GET` `/lol-purchase-widget/v3/base-skin-line-data/{offerId}`
- `GET` `/lol-purchase-widget/v3/purchase-offer-order-statuses`
- `POST` `/lol-purchase-widget/v3/purchaseOffer`
- `POST` `/lol-purchase-widget/v3/purchaseOfferViaCap`
- `POST` `/lol-purchase-widget/v3/validateOffer`

## /lol-ranked

- `GET` `/lol-ranked/v1/apex-leagues/{queueType}/{tier}`
- `GET` `/lol-ranked/v1/cached-ranked-stats/{puuid}`
- `GET` `/lol-ranked/v1/challenger-ladders-enabled`
- `GET` `/lol-ranked/v1/current-lp-change-notification`
- `GET` `/lol-ranked/v1/current-ranked-stats`
- `GET` `/lol-ranked/v1/eligibleTiers/queueType/{queueType}`
- `GET` `/lol-ranked/v1/eos-notifications`
- `POST` `/lol-ranked/v1/eos-notifications/{id}/acknowledge`
- `GET` `/lol-ranked/v1/eos-rewards/{seasonId}`
- `GET` `/lol-ranked/v1/global-notifications`
- `GET` `/lol-ranked/v1/league-ladders/{puuid}`
- `GET` `/lol-ranked/v1/notifications`
- `POST` `/lol-ranked/v1/notifications/{id}/acknowledge`
- `GET` `/lol-ranked/v1/ranked-stats/{puuid}`
- `GET` `/lol-ranked/v1/rated-ladder/{queueType}`
- `GET` `/lol-ranked/v1/signed-ranked-stats`
- `GET` `/lol-ranked/v1/social-leaderboard-ranked-queue-stats-for-puuids`
- `GET` `/lol-ranked/v1/top-rated-ladders-enabled`
- `GET` `/lol-ranked/v2/tiers`

## /lol-regalia

- `GET` `/lol-regalia/v2/config`
- `GET, PUT` `/lol-regalia/v2/current-summoner/regalia`
- `GET` `/lol-regalia/v2/summoners/{summonerId}/queues/{queue}/positions/{position}/regalia`
- `GET` `/lol-regalia/v2/summoners/{summonerId}/queues/{queue}/regalia`
- `GET` `/lol-regalia/v2/summoners/{summonerId}/regalia`
- `GET` `/lol-regalia/v2/summoners/{summonerId}/regalia/async`
- `GET` `/lol-regalia/v3/inventory/{inventoryType}`
- `GET` `/lol-regalia/v3/summoners/{summonerId}/regalia`

## /lol-remedy

- `PUT` `/lol-remedy/v1/ack-remedy-notification/{mailId}`
- `GET` `/lol-remedy/v1/config/is-verbal-abuse-remedy-modal-enabled`
- `GET` `/lol-remedy/v1/remedy-notifications`

## /lol-replays

- `GET` `/lol-replays/v1/configuration`
- `GET` `/lol-replays/v1/metadata/{gameId}`
- `POST` `/lol-replays/v1/metadata/{gameId}/create/gameVersion/{gameVersion}/gameType/{gameType}/queueId/{queueId}`
- `GET` `/lol-replays/v1/rofls/path`
- `GET` `/lol-replays/v1/rofls/path/default`
- `POST` `/lol-replays/v1/rofls/scan`
- `POST` `/lol-replays/v1/rofls/{gameId}/download`
- `POST` `/lol-replays/v1/rofls/{gameId}/download/graceful`
- `POST` `/lol-replays/v1/rofls/{gameId}/watch`
- `POST` `/lol-replays/v2/metadata/{gameId}/create`

## /lol-reward-track

- `GET` `/lol-reward-track/register/{progressionGroupId}`
- `GET` `/lol-reward-track/{progressionGroupId}/reward-track/bonus-items`
- `GET` `/lol-reward-track/{progressionGroupId}/reward-track/bonus-progress`
- `GET` `/lol-reward-track/{progressionGroupId}/reward-track/failure`
- `GET` `/lol-reward-track/{progressionGroupId}/reward-track/items`
- `GET` `/lol-reward-track/{progressionGroupId}/reward-track/progress`
- `GET` `/lol-reward-track/{progressionGroupId}/reward-track/unclaimed-rewards`
- `GET` `/lol-reward-track/{progressionGroupId}/reward-track/xp`

## /lol-rewards

- `POST` `/lol-rewards/v1/celebrations/fsc`
- `GET` `/lol-rewards/v1/grants`
- `PATCH` `/lol-rewards/v1/grants/view`
- `POST` `/lol-rewards/v1/grants/{grantId}/select`
- `GET` `/lol-rewards/v1/groups`
- `POST` `/lol-rewards/v1/reward/replay`
- `POST` `/lol-rewards/v1/select-bulk`

## /lol-rms

- `GET` `/lol-rms/v1/champion-mastery-leaveup-update`
- `DELETE` `/lol-rms/v1/champion-mastery-leaveup-update/{id}`

## /lol-rso-auth

- `GET` `/lol-rso-auth/configuration/v3/ready-state`
- `DELETE, GET` `/lol-rso-auth/v1/authorization`
- `GET` `/lol-rso-auth/v1/authorization/access-token`
- `GET` `/lol-rso-auth/v1/authorization/country`
- `GET` `/lol-rso-auth/v1/authorization/error`
- `POST` `/lol-rso-auth/v1/authorization/gas`
- `GET` `/lol-rso-auth/v1/authorization/id-token`
- `POST` `/lol-rso-auth/v1/authorization/refresh`
- `GET, POST` `/lol-rso-auth/v1/authorization/userinfo`
- `POST` `/lol-rso-auth/v1/device-id`
- `POST` `/lol-rso-auth/v1/external-session-config`
- `DELETE` `/lol-rso-auth/v1/session`
- `GET` `/lol-rso-auth/v1/status/{platformId}`
- `DELETE, POST` `/lol-rso-auth/v2/config`

## /lol-sanctum

- `GET` `/lol-sanctum/v1/banners`
- `GET, POST` `/lol-sanctum/v1/banners/selected/id`
- `GET` `/lol-sanctum/v1/banners/{bannerId}/odds`
- `GET` `/lol-sanctum/v1/banners/{bannerId}/token-catalog-entry`
- `GET` `/lol-sanctum/v1/counters/{counterId}`
- `POST` `/lol-sanctum/v1/purchase/roll`
- `POST` `/lol-sanctum/v1/purchase/token`
- `GET` `/lol-sanctum/v1/purchases/{purchaseId}`
- `GET` `/lol-sanctum/v1/ready`
- `GET` `/lol-sanctum/v1/velocity-limits`

## /lol-seasons

- `POST` `/lol-seasons/v1/allSeasons/product/{product}`
- `GET` `/lol-seasons/v1/season/LOL/current-split-seasons`
- `GET` `/lol-seasons/v1/season/name/{name}`
- `GET` `/lol-seasons/v1/season/product/{product}`
- `GET` `/lol-seasons/v1/season/recent-final-split`
- `GET` `/lol-seasons/v1/season/should-ac-refresh`

## /lol-service-status

- `GET` `/lol-service-status/v1/lcu-status`
- `GET` `/lol-service-status/v1/ticker-messages`

## /lol-settings

- `GET` `/lol-settings/v1/account/didreset`
- `POST` `/lol-settings/v1/account/save`
- `GET, PATCH, PUT` `/lol-settings/v1/account/{category}`
- `GET, PATCH` `/lol-settings/v1/local/{category}`
- `GET, PATCH, PUT` `/lol-settings/v2/account/{ppType}/{category}`
- `GET` `/lol-settings/v2/config`
- `GET` `/lol-settings/v2/didreset/{ppType}`
- `GET, PATCH` `/lol-settings/v2/local/{category}`
- `GET` `/lol-settings/v2/ready`
- `POST` `/lol-settings/v2/reload/{ppType}`

## /lol-shoppefront

- `POST` `/lol-shoppefront/v1/bulk-purchases`
- `POST` `/lol-shoppefront/v1/purchases`
- `GET` `/lol-shoppefront/v1/purchases/{purchaseId}`
- `GET` `/lol-shoppefront/v1/ready`
- `GET, POST` `/lol-shoppefront/v1/store-digests`
- `GET` `/lol-shoppefront/v1/stores`
- `GET` `/lol-shoppefront/v1/stores/{shoppefrontId}`
- `GET` `/lol-shoppefront/v1/velocity`

## /lol-shutdown

- `GET` `/lol-shutdown/v1/notification`

## /lol-simple-dialog-messages

- `GET, POST` `/lol-simple-dialog-messages/v1/messages`
- `DELETE` `/lol-simple-dialog-messages/v1/messages/{messageId}`

## /lol-social-leaderboard

- `GET` `/lol-social-leaderboard/v1/leaderboard-next-update-time`
- `GET` `/lol-social-leaderboard/v1/social-leaderboard-data`

## /lol-spectator

- `POST` `/lol-spectator/v1/buddy/spectate`
- `GET` `/lol-spectator/v1/spectate`
- `GET` `/lol-spectator/v1/spectate/config`
- `POST` `/lol-spectator/v1/spectate/launch`
- `POST` `/lol-spectator/v2/buddy/spectate`
- `GET` `/lol-spectator/v3/buddy/can-spectate/{puuid}/{spectatorKey}`
- `POST` `/lol-spectator/v3/buddy/spectate`

## /lol-statstones

- `GET` `/lol-statstones/v1/eog-notifications/{gameId}`
- `GET` `/lol-statstones/v1/featured-champion-statstones/{championItemId}`
- `POST` `/lol-statstones/v1/featured-champion-statstones/{championItemId}/{statstoneId}`
- `GET` `/lol-statstones/v1/profile-summary/{puuid}`
- `GET` `/lol-statstones/v1/statstone/{contentId}/owned`
- `GET` `/lol-statstones/v1/statstones-enabled-queue-ids`
- `DELETE, GET` `/lol-statstones/v1/vignette-notifications`
- `DELETE` `/lol-statstones/v1/vignette-notifications/{key}`
- `GET` `/lol-statstones/v2/player-statstones-self/{championItemId}`
- `GET` `/lol-statstones/v2/player-summary-self`

## /lol-store

- `GET` `/lol-store/v1/alias-change-notifications`
- `GET` `/lol-store/v1/catalog`
- `GET` `/lol-store/v1/catalog/items/skip-cache`
- `GET` `/lol-store/v1/catalog/sales`
- `GET` `/lol-store/v1/catalog/{inventoryType}`
- `GET` `/lol-store/v1/catalogByInstanceIds`
- `GET` `/lol-store/v1/getStoreUrl`
- `GET` `/lol-store/v1/giftablefriends`
- `GET` `/lol-store/v1/itemKeysFromInstanceIds`
- `GET` `/lol-store/v1/itemKeysFromOfferIds`
- `GET, POST` `/lol-store/v1/lastPage`
- `POST` `/lol-store/v1/notifications/acknowledge`
- `GET` `/lol-store/v1/offers`
- `GET` `/lol-store/v1/offers/{offerId}`
- `GET` `/lol-store/v1/order-notifications`
- `GET` `/lol-store/v1/order-notifications/{id}`
- `GET` `/lol-store/v1/paymentDetails`
- `GET` `/lol-store/v1/skins/{skinId}`
- `GET` `/lol-store/v1/status`
- `GET` `/lol-store/v1/store-ready`
- `GET` `/lol-store/v1/{pageType}`
- `GET` `/lol-store/v2/offers`
- `POST` `/lol-store/v3/purchase`

## /lol-suggested-players

- `POST` `/lol-suggested-players/v1/reported-player`
- `GET` `/lol-suggested-players/v1/suggested-players`
- `DELETE` `/lol-suggested-players/v1/suggested-players/{summonerId}`
- `POST` `/lol-suggested-players/v1/victorious-comrade`

## /lol-summoner

- `GET` `/lol-summoner/v1/alias/lookup`
- `GET` `/lol-summoner/v1/check-name-availability-new-summoners/{name}`
- `GET` `/lol-summoner/v1/check-name-availability/{name}`
- `GET` `/lol-summoner/v1/current-summoner`
- `GET` `/lol-summoner/v1/current-summoner/account-and-summoner-ids`
- `GET` `/lol-summoner/v1/current-summoner/autofill`
- `PUT` `/lol-summoner/v1/current-summoner/icon`
- `GET` `/lol-summoner/v1/current-summoner/jwt`
- `POST` `/lol-summoner/v1/current-summoner/name`
- `GET, PUT` `/lol-summoner/v1/current-summoner/profile-privacy`
- `GET` `/lol-summoner/v1/current-summoner/rerollPoints`
- `GET, POST` `/lol-summoner/v1/current-summoner/summoner-profile`
- `GET` `/lol-summoner/v1/player-alias-state`
- `GET` `/lol-summoner/v1/player-name-mode`
- `GET` `/lol-summoner/v1/profile-privacy-enabled`
- `GET` `/lol-summoner/v1/riot-alias-free-eligibility`
- `GET` `/lol-summoner/v1/riot-alias-purchase-eligibility`
- `POST` `/lol-summoner/v1/save-alias`
- `GET` `/lol-summoner/v1/status`
- `POST` `/lol-summoner/v1/summoner-aliases-by-ids`
- `POST` `/lol-summoner/v1/summoner-aliases-by-puuids`
- `GET` `/lol-summoner/v1/summoner-profile`
- `GET` `/lol-summoner/v1/summoner-requests-ready`
- `GET, POST` `/lol-summoner/v1/summoners`
- `GET` `/lol-summoner/v1/summoners-by-puuid-cached/{puuid}`
- `POST` `/lol-summoner/v1/summoners/aliases`
- `GET` `/lol-summoner/v1/summoners/{id}`
- `POST` `/lol-summoner/v1/validate-alias`
- `GET` `/lol-summoner/v2/summoner-icons`
- `GET` `/lol-summoner/v2/summoner-icons-by-puuids`
- `GET` `/lol-summoner/v2/summoner-names`
- `GET` `/lol-summoner/v2/summoner-names-by-puuids`
- `GET` `/lol-summoner/v2/summoners`
- `POST` `/lol-summoner/v2/summoners/names`
- `POST` `/lol-summoner/v2/summoners/puuid`
- `GET` `/lol-summoner/v2/summoners/puuid/{puuid}`

## /lol-summoner-profiles

- `GET` `/lol-summoner-profiles/v1/get-champion-mastery-view`
- `GET` `/lol-summoner-profiles/v1/get-honor-view`
- `GET` `/lol-summoner-profiles/v1/get-lol-eos-rewards-view`
- `GET` `/lol-summoner-profiles/v1/get-privacy-view`
- `GET` `/lol-summoner-profiles/v1/get-restriction-view`
- `GET` `/lol-summoner-profiles/v1/get-summoner-level-view`
- `POST` `/lol-summoner-profiles/v1/pco/{category}`

## /lol-tastes

- `GET` `/lol-tastes/v1/ready`
- `GET` `/lol-tastes/v1/skins-model`
- `GET` `/lol-tastes/v1/tft-overview-model`

## /lol-tft

- `GET` `/lol-tft/v1/tft/backgrounds`
- `GET` `/lol-tft/v1/tft/battlePassHub`
- `GET` `/lol-tft/v1/tft/events`
- `GET` `/lol-tft/v1/tft/homeHub`
- `POST` `/lol-tft/v1/tft/homeHub/redirect`
- `GET` `/lol-tft/v1/tft/newsHub`
- `GET` `/lol-tft/v1/tft/promoButtons`
- `GET` `/lol-tft/v1/tft/tencentEventhubConfigs`
- `PUT` `/lol-tft/v1/tft_experiment_bucket`

## /lol-tft-event-pve

- `GET` `/lol-tft-event-pve/v1/buddies`
- `GET` `/lol-tft-event-pve/v1/buddy`
- `PUT` `/lol-tft-event-pve/v1/buddy/{id}/equip`
- `GET` `/lol-tft-event-pve/v1/difficulty`
- `PUT` `/lol-tft-event-pve/v1/difficulty/{id}/equip`
- `GET` `/lol-tft-event-pve/v1/enabled`
- `GET` `/lol-tft-event-pve/v1/eogmissionrewards`
- `POST` `/lol-tft-event-pve/v1/eogmissionrewards/clear`
- `GET` `/lol-tft-event-pve/v1/eventpvehub`
- `GET` `/lol-tft-event-pve/v1/journeytrack/bonuses`
- `GET` `/lol-tft-event-pve/v1/ready`
- `PUT` `/lol-tft-event-pve/v1/seenBuddies`
- `PUT` `/lol-tft-event-pve/v1/seenBuddies/{id}`
- `PUT` `/lol-tft-event-pve/v1/seenLevels`
- `PUT` `/lol-tft-event-pve/v1/seenUltimateVictory`

## /lol-tft-pass

- `GET` `/lol-tft-pass/v1/active-passes`
- `GET` `/lol-tft-pass/v1/battle-pass`
- `GET` `/lol-tft-pass/v1/config-fetched`
- `GET` `/lol-tft-pass/v1/daily-login-pass`
- `GET` `/lol-tft-pass/v1/enabled`
- `GET` `/lol-tft-pass/v1/event-pass`
- `GET` `/lol-tft-pass/v1/objectives-banner/{id}`
- `POST` `/lol-tft-pass/v1/pass/{id}`
- `PUT` `/lol-tft-pass/v1/pass/{id}/milestone/claimAllRewards`
- `PUT` `/lol-tft-pass/v1/pass/{id}/milestone/{milestoneId}/reward`
- `POST` `/lol-tft-pass/v1/passes`
- `GET` `/lol-tft-pass/v1/pm-ultimate-victory-pass`
- `GET` `/lol-tft-pass/v1/ready`
- `GET` `/lol-tft-pass/v1/reward-notification`
- `GET` `/lol-tft-pass/v1/skill-tree-pass`

## /lol-tft-skill-tree

- `GET` `/lol-tft-skill-tree/v1/enabled`
- `GET` `/lol-tft-skill-tree/v1/player-progression`
- `GET` `/lol-tft-skill-tree/v1/ready`
- `GET` `/lol-tft-skill-tree/v1/skill-tree`
- `GET` `/lol-tft-skill-tree/v1/skill-tree-rank/{rank}`
- `PUT` `/lol-tft-skill-tree/v1/skill-tree-rank/{rank}/claim-rewards`
- `GET` `/lol-tft-skill-tree/v1/skill/{skillId}`
- `PUT` `/lol-tft-skill-tree/v1/skill/{skillId}/equip`

## /lol-tft-team-planner

- `GET` `/lol-tft-team-planner/v1/config`
- `GET, PATCH` `/lol-tft-team-planner/v1/ftue/hasViewed`
- `POST` `/lol-tft-team-planner/v1/is-name-valid/{name}`
- `GET, PUT` `/lol-tft-team-planner/v1/previous-context`
- `PATCH` `/lol-tft-team-planner/v1/set`
- `PATCH` `/lol-tft-team-planner/v1/set/lastViewed`
- `GET, POST` `/lol-tft-team-planner/v1/sets/dirty`
- `PUT` `/lol-tft-team-planner/v1/sets/save-all`
- `DELETE, PATCH` `/lol-tft-team-planner/v1/sets/{set}/reminders/{team}`
- `POST` `/lol-tft-team-planner/v1/sets/{set}/team-code/{team}`
- `DELETE, POST` `/lol-tft-team-planner/v1/sets/{set}/teams/{team}`
- `DELETE, PATCH, POST` `/lol-tft-team-planner/v1/sets/{set}/teams/{team}/champions`
- `DELETE, POST` `/lol-tft-team-planner/v1/sets/{set}/teams/{team}/champions/{championId}`
- `DELETE, POST` `/lol-tft-team-planner/v1/sets/{set}/teams/{team}/champions/{championId}/{index}`
- `POST` `/lol-tft-team-planner/v1/sets/{set}/teams/{team}/import`
- `POST` `/lol-tft-team-planner/v1/sets/{set}/teams/{team}/lastView`
- `PATCH` `/lol-tft-team-planner/v1/sets/{set}/teams/{team}/{name}`
- `GET, PATCH` `/lol-tft-team-planner/v1/sort-option`
- `POST` `/lol-tft-team-planner/v1/team-code/clipboard/{set}`
- `GET` `/lol-tft-team-planner/v2/reminders`

## /lol-tft-troves

- `GET` `/lol-tft-troves/v1/banners`
- `GET` `/lol-tft-troves/v1/config`
- `GET` `/lol-tft-troves/v1/loot-odds/{dropTableId}`
- `GET` `/lol-tft-troves/v1/milestone-notifications`
- `GET` `/lol-tft-troves/v1/milestones`
- `POST` `/lol-tft-troves/v1/purchase`
- `POST` `/lol-tft-troves/v1/roll`
- `DELETE, GET` `/lol-tft-troves/v1/roll-rewards`
- `GET` `/lol-tft-troves/v1/status-notifications`
- `POST` `/lol-tft-troves/v2/roll`

## /lol-trophies

- `GET` `/lol-trophies/v1/current-summoner/trophies/profile`
- `GET` `/lol-trophies/v1/players/{puuid}/trophies/profile`

## /lol-vanguard

- `PUT` `/lol-vanguard/v1/ack-notification/{mailId}`
- `GET` `/lol-vanguard/v1/config/days-to-reshow-modal`
- `GET` `/lol-vanguard/v1/config/enabled`
- `GET` `/lol-vanguard/v1/is-playing-in-pcb`
- `GET` `/lol-vanguard/v1/machine-specs`
- `GET` `/lol-vanguard/v1/notification`
- `GET` `/lol-vanguard/v1/session`
- `POST` `/lol-vanguard/v1/telemetry/system-check`

## /lol-yourshop

- `GET` `/lol-yourshop/v1/has-permissions`
- `GET` `/lol-yourshop/v1/modal`
- `GET` `/lol-yourshop/v1/offers`
- `POST` `/lol-yourshop/v1/offers/{id}/purchase`
- `POST` `/lol-yourshop/v1/offers/{id}/reveal`
- `POST` `/lol-yourshop/v1/permissions`
- `GET` `/lol-yourshop/v1/ready`
- `GET` `/lol-yourshop/v1/status`
- `GET` `/lol-yourshop/v1/themed`

## /memory

- `GET` `/memory/v1/fe-processes-ready`
- `POST` `/memory/v1/notify-fe-processes-ready`
- `POST` `/memory/v1/snapshot`

## /patcher

- `GET, POST` `/patcher/v1/notifications`
- `DELETE` `/patcher/v1/notifications/{id}`
- `GET, PATCH` `/patcher/v1/p2p/status`
- `GET` `/patcher/v1/products`
- `DELETE` `/patcher/v1/products/{product-id}`
- `POST` `/patcher/v1/products/{product-id}/detect-corruption-request`
- `POST` `/patcher/v1/products/{product-id}/partial-repair-request`
- `GET` `/patcher/v1/products/{product-id}/paths`
- `POST` `/patcher/v1/products/{product-id}/signal-start-patching-delayed`
- `POST` `/patcher/v1/products/{product-id}/start-checking-request`
- `POST` `/patcher/v1/products/{product-id}/start-patching-request`
- `GET` `/patcher/v1/products/{product-id}/state`
- `POST` `/patcher/v1/products/{product-id}/stop-checking-request`
- `POST` `/patcher/v1/products/{product-id}/stop-patching-request`
- `GET` `/patcher/v1/products/{product-id}/tags`
- `GET` `/patcher/v1/status`
- `PUT` `/patcher/v1/ux`

## /payments

- `POST` `/payments/v1/pmc-start-url`
- `POST` `/payments/v1/updatePaymentTelemetryState`

## /performance

- `GET` `/performance/v1/memory`
- `POST` `/performance/v1/process/{processId}`
- `GET` `/performance/v1/report`
- `POST` `/performance/v1/report/restart`
- `GET` `/performance/v1/system-info`

## /player-notifications

- `GET, PUT` `/player-notifications/v1/config`
- `GET, POST` `/player-notifications/v1/notifications`
- `DELETE, GET, PUT` `/player-notifications/v1/notifications/{id}`

## /plugin-manager

- `GET` `/plugin-manager/v1/external-plugins/availability`
- `GET` `/plugin-manager/v1/status`
- `GET` `/plugin-manager/v2/descriptions`
- `GET` `/plugin-manager/v2/descriptions/{plugin}`
- `GET` `/plugin-manager/v2/plugins`
- `GET` `/plugin-manager/v2/plugins/{plugin}`
- `GET` `/plugin-manager/v3/plugins-manifest`

## /process-control

- `GET` `/process-control/v1/process`
- `POST` `/process-control/v1/process/quit`

## /riot-messaging-service

- `DELETE, POST` `/riot-messaging-service/v1/connect`
- `DELETE, POST` `/riot-messaging-service/v1/entitlements`
- `GET` `/riot-messaging-service/v1/message/{a}`
- `GET` `/riot-messaging-service/v1/message/{a}/{b}`
- `GET` `/riot-messaging-service/v1/message/{a}/{b}/{c}`
- `GET` `/riot-messaging-service/v1/message/{a}/{b}/{c}/{d}`
- `GET` `/riot-messaging-service/v1/message/{a}/{b}/{c}/{d}/{e}`
- `GET` `/riot-messaging-service/v1/message/{a}/{b}/{c}/{d}/{e}/{f}`
- `DELETE, GET` `/riot-messaging-service/v1/session`
- `GET` `/riot-messaging-service/v1/state`

## /riotclient

- `POST` `/riotclient/addorupdatemetric`
- `DELETE, GET, POST` `/riotclient/affinity`
- `GET` `/riotclient/app-name`
- `GET` `/riotclient/app-port`
- `GET` `/riotclient/auth-token`
- `GET, POST` `/riotclient/clipboard`
- `GET` `/riotclient/command-line-args`
- `POST` `/riotclient/kill-and-restart-ux`
- `POST` `/riotclient/kill-ux`
- `POST` `/riotclient/launch-ux`
- `GET` `/riotclient/machine-id`
- `POST` `/riotclient/new-args`
- `POST` `/riotclient/open-url-in-browser`
- `GET` `/riotclient/region-locale`
- `POST` `/riotclient/show-swagger`
- `DELETE, PUT` `/riotclient/splash`
- `GET` `/riotclient/system-info/v1/basic-info`
- `GET` `/riotclient/trace`
- `POST` `/riotclient/unload`
- `POST` `/riotclient/ux-allow-foreground`
- `GET` `/riotclient/ux-crash-count`
- `POST` `/riotclient/ux-flash`
- `PUT` `/riotclient/ux-load-complete`
- `POST` `/riotclient/ux-minimize`
- `POST` `/riotclient/ux-show`
- `GET` `/riotclient/ux-state`
- `PUT` `/riotclient/ux-state/ack`
- `DELETE, PUT` `/riotclient/v1/auth-tokens/{authToken}`
- `GET, PUT` `/riotclient/v1/crash-reporting/environment`
- `POST` `/riotclient/v1/crash-reporting/logs`
- `POST` `/riotclient/v1/elevation-requests`
- `GET, POST` `/riotclient/zoom-scale`

## /sanitizer

- `POST` `/sanitizer/v1/containsSanitized`
- `POST` `/sanitizer/v1/sanitize`
- `GET` `/sanitizer/v1/status`

## /services-api

- `POST` `/services-api/config/v2/client-config/{prefix}/{scope}`
- `PUT` `/services-api/game-session/v1/game-session-token`
- `PUT` `/services-api/identity/v1/player-alias`
- `DELETE, PUT` `/services-api/player-session/v1/access-token`
- `DELETE, PUT` `/services-api/player-session/v1/entitlements-token`
- `DELETE, PUT` `/services-api/player-session/v1/user-info-token`
- `POST` `/services-api/riot-message-service/v1/data`

## /system

- `GET` `/system/v1/builds`

## /telemetry

- `GET` `/telemetry/v1/application-start-time`
- `GET` `/telemetry/v1/common-data`
- `POST` `/telemetry/v1/common-data/{key}`
- `POST` `/telemetry/v1/events-with-perf-info/{eventType}`
- `POST` `/telemetry/v1/events/{eventType}`
- `PATCH` `/telemetry/v3/slis/add-bool-diagnostic`
- `PATCH` `/telemetry/v3/slis/add-double-diagnostic`
- `PATCH` `/telemetry/v3/slis/add-int-diagnostic`
- `PATCH` `/telemetry/v3/slis/add-label`
- `PATCH` `/telemetry/v3/slis/add-string-diagnostic`
- `POST` `/telemetry/v3/slis/counts`
- `POST` `/telemetry/v3/uptime-tracking/notify-failure`
- `POST` `/telemetry/v3/uptime-tracking/notify-success`

## /tracing

- `DELETE, POST` `/tracing/v1/performance/{name}`
- `POST` `/tracing/v1/trace/critical-flow`
- `POST` `/tracing/v1/trace/event`
- `POST` `/tracing/v1/trace/module`
- `POST` `/tracing/v1/trace/non-timing-event/{eventName}`
- `POST` `/tracing/v1/trace/phase/begin`
- `POST` `/tracing/v1/trace/phase/end`
- `POST` `/tracing/v1/trace/step-event`
- `DELETE, POST` `/tracing/v1/trace/time-series-event/{eventName}`
- `POST` `/tracing/v1/trace/time-series-event/{eventName}/marker/{markerName}`

## /{plugin}

- `GET, HEAD` `/{plugin}/assets/{+path}`
