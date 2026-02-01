# LCU API — Schema Reference

Key type definitions extracted from the official LCU swagger.

Only includes schemas directly used by API endpoints and their immediate dependencies.


---

## ChemtechShoppe


### ChemtechShoppe-CatalogEntryDto

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `name` | string | Yes |
| `productId` | string | Yes |
| `purchaseUnits` | ChemtechShoppe-PurchaseUnitDto[] | Yes |
| `displayMetadata` | object (map) | Yes |
| `prerequisites` | ChemtechShoppe-PrerequisiteDto[] | Yes |
| `purchaseVisibility` | string | Yes |
| `refundRule` | string | Yes |
| `giftRule` | string | Yes |
| `purchaseLimits` | ChemtechShoppe-VelocityLimitDeltaDto[] | Yes |
| `refundLimits` | ChemtechShoppe-VelocityLimitDeltaDto[] | Yes |
| `endTime` | string | Yes |

### ChemtechShoppe-PagingDto

| Field | Type | Required |
|-------|------|----------|
| `offset` | int32 | Yes |
| `limit` | int32 | Yes |
| `maxLimit` | int32 | Yes |
| `total` | int32 | Yes |
| `previous` | string | Yes |
| `next` | string | Yes |

### ChemtechShoppe-PrerequisiteDto

| Field | Type | Required |
|-------|------|----------|
| `status` | string | Yes |
| `itemTypeId` | string | Yes |
| `itemId` | string | Yes |
| `requiredQuantity` | int64 | Yes |
| `ownedQuantity` | int64 | Yes |
| `catalogEntry` | object (map) | Yes |
| `milestoneId` | string | Yes |

### ChemtechShoppe-PurchaseUnitDto

| Field | Type | Required |
|-------|------|----------|
| `paymentOptions` | ChemtechShoppe-PaymentOptionDto[] | Yes |
| `fulfillment` | ChemtechShoppe-FulfillmentDto | Yes |

### ChemtechShoppe-RotatingStoreMetadataDto

| Field | Type | Required |
|-------|------|----------|
| `slots` | ChemtechShoppe-RotatingStoreSlotMetadataDto[] | Yes |
| `currRotationStartTime` | string | Yes |
| `nextRotationStartTime` | string | Yes |
| `rotationCadence` | string | Yes |

### ChemtechShoppe-StatsDto

| Field | Type | Required |
|-------|------|----------|
| `durationMs` | int64 | Yes |

### ChemtechShoppe-StoreDigestDto

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `productId` | string | Yes |
| `name` | string | Yes |
| `displayMetaData` | object (map) | Yes |
| `startTime` | string | Yes |
| `endTime` | string | Yes |
| `rotatingStoreMetadata` | ChemtechShoppe-RotatingStoreMetadataDto | Yes |

### ChemtechShoppe-StoreDigestsDto

| Field | Type | Required |
|-------|------|----------|
| `digests` | ChemtechShoppe-StoreDigestDto[] | Yes |

### ChemtechShoppe-StoreDto

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `type` | string | Yes |
| `productId` | string | Yes |
| `name` | string | Yes |
| `catalogEntries` | ChemtechShoppe-CatalogEntryDto[] | Yes |
| `displayMetadata` | object (map) | Yes |
| `startTime` | string | Yes |
| `endTime` | string | Yes |
| `rotatingStoreMetadata` | ChemtechShoppe-RotatingStoreMetadataDto | Yes |

### ChemtechShoppe-StoresResponseDto

| Field | Type | Required |
|-------|------|----------|
| `data` | ChemtechShoppe-StoreDto[] | Yes |
| `paging` | ChemtechShoppe-PagingDto | Yes |
| `stats` | ChemtechShoppe-StatsDto | Yes |
| `notes` | string[] | Yes |

### ChemtechShoppe-VelocityLimitDeltaDto

| Field | Type | Required |
|-------|------|----------|
| `ruleId` | string | Yes |
| `delta` | int64 | Yes |
| `remaining` | int64 | Yes |

---

## ChemtechShoppeV2


### ChemtechShoppeV2-CatalogEntryDto

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `name` | string | Yes |
| `productId` | string | Yes |
| `purchaseUnits` | ChemtechShoppeV2-PurchaseUnitDto[] | Yes |
| `displayMetadata` | object (map) | Yes |
| `prerequisites` | ChemtechShoppeV2-PrerequisiteDto[] | Yes |
| `purchaseVisibility` | string | Yes |
| `refundRule` | string | Yes |
| `giftRule` | string | Yes |
| `purchaseLimits` | ChemtechShoppeV2-VelocityLimitDeltaDto[] | Yes |
| `refundLimits` | ChemtechShoppeV2-VelocityLimitDeltaDto[] | Yes |
| `endTime` | string | Yes |

### ChemtechShoppeV2-PrerequisiteDto

| Field | Type | Required |
|-------|------|----------|
| `status` | string | Yes |
| `itemTypeId` | string | Yes |
| `itemId` | string | Yes |
| `requiredQuantity` | int64 | Yes |
| `ownedQuantity` | int64 | Yes |
| `catalogEntry` | object (map) | Yes |
| `milestoneId` | string | Yes |

### ChemtechShoppeV2-PurchaseUnitDto

| Field | Type | Required |
|-------|------|----------|
| `paymentOptions` | ChemtechShoppeV2-PaymentOptionDto[] | Yes |
| `fulfillment` | ChemtechShoppeV2-FulfillmentDto | Yes |

### ChemtechShoppeV2-VelocityLimitDeltaDto

| Field | Type | Required |
|-------|------|----------|
| `ruleId` | string | Yes |
| `delta` | int64 | Yes |
| `remaining` | int64 | Yes |

---

## LolAccount


### LolAccountVerificationConfirmActivationPinRequest

| Field | Type | Required |
|-------|------|----------|
| `oneTimePin` | string | Yes |

### LolAccountVerificationConfirmDeactivationPinRequest

| Field | Type | Required |
|-------|------|----------|
| `oneTimePin` | string | Yes |

### LolAccountVerificationIsVerifiedResponse

| Field | Type | Required |
|-------|------|----------|
| `success` | boolean | Yes |
| `message` | string | Yes |
| `status` | int32 | Yes |

### LolAccountVerificationPhoneNumberResponse

| Field | Type | Required |
|-------|------|----------|
| `data` | LolAccountVerificationPhoneNumberResponseData | Yes |
| `error` | LolAccountVerificationResponseError | Yes |
| `clientMessageId` | string | Yes |

### LolAccountVerificationPhoneNumberResponseData

| Field | Type | Required |
|-------|------|----------|
| `phoneNumberObfuscated` | LolAccountVerificationPhoneNumberObfuscated | Yes |

### LolAccountVerificationResponseError

| Field | Type | Required |
|-------|------|----------|
| `errorCode` | string | Yes |
| `message` | string | Yes |

### LolAccountVerificationSendActivationPinRequest

| Field | Type | Required |
|-------|------|----------|
| `phoneNumber` | string | Yes |
| `locale` | string | Yes |

---

## LolActive


### LolActiveBoostsActiveBoosts

| Field | Type | Required |
|-------|------|----------|
| `xpBoostEndDate` | string | Yes |
| `xpBoostPerWinCount` | uint64 | Yes |
| `xpLoyaltyBoost` | int32 | Yes |
| `firstWinOfTheDayStartTime` | string | Yes |

---

## LolActivity


### LolActivityCenterConfigData

| Field | Type | Required |
|-------|------|----------|
| `puuid` | string | Yes |
| `sessionId` | string | Yes |
| `region` | string | Yes |
| `locale` | string | Yes |
| `webRegion` | string | Yes |
| `webLocale` | string | Yes |
| `publishingLocale` | string | Yes |
| `rsoPlatformId` | string | Yes |
| `enabled` | boolean | Yes |
| `clientNavEnabled` | boolean | Yes |

### LolActivityCenterTencentOverrides

| Field | Type | Required |
|-------|------|----------|
| `infoHubAlternativeExperienceUrl` | string | Yes |
| `activityCenterAlternativeExperienceUrl` | string | Yes |

---

## LolBanners


### LolBannersBannerFlag

| Field | Type | Required |
|-------|------|----------|
| `itemId` | int32 | Yes |
| `theme` | string | Yes |
| `level` | int64 | Yes |
| `seasonId` | int64 | Yes |
| `earnedDateIso8601` | string | Yes |

### LolBannersBannerFrame

| Field | Type | Required |
|-------|------|----------|
| `level` | int64 | Yes |

---

## LolCap


### LolCapMissionsCapMissionSeries

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `configurationId` | string | Yes |
| `missions` | LolCapMissionsCapMissionSeriesMission[] | Yes |

### LolCapMissionsCapMissionsMeResponse

| Field | Type | Required |
|-------|------|----------|
| `productId` | string | Yes |
| `ownerId` | string | Yes |
| `series` | LolCapMissionsCapMissionSeries[] | Yes |

---

## LolCatalog


### LolCatalogCatalogPluginItem

| Field | Type | Required |
|-------|------|----------|
| `itemId` | int32 | Yes |
| `itemInstanceId` | string | Yes |
| `owned` | boolean | Yes |
| `ownershipType` | LolCatalogInventoryOwnership |  |
| `inventoryType` | string | Yes |
| `subInventoryType` | string | Yes |
| `name` | string | Yes |
| `subTitle` | string | Yes |
| `description` | string | Yes |
| `imagePath` | string | Yes |
| `tilePath` | string | Yes |
| `loadScreenPath` | string | Yes |
| `rarity` | string | Yes |
| `taggedChampionsIds` | integer[] | Yes |
| `purchaseDate` | uint64 | Yes |
| `releaseDate` | uint64 | Yes |
| `inactiveDate` | uint64 | Yes |
| `maxQuantity` | int32 | Yes |
| `prices` | LolCatalogCatalogPluginPrice[] | Yes |
| `tags` | string[] |  |
| `metadata` | LolCatalogItemMetadataEntry[] |  |
| `active` | boolean | Yes |
| `sale` | LolCatalogSale |  |
| `questSkinInfo` | LolCatalogSkinLineInfo |  |
| `offerId` | string |  |

### LolCatalogCatalogPluginItemAssets

| Field | Type | Required |
|-------|------|----------|
| `splashPath` | string | Yes |
| `iconPath` | string | Yes |
| `tilePath` | string | Yes |
| `loadScreenPath` | string | Yes |
| `previewVideoUrl` | string | Yes |
| `emblems` | LolCatalogChampionSkinEmblem[] | Yes |
| `colors` | string[] | Yes |

### LolCatalogCatalogPluginItemWithDetails

| Field | Type | Required |
|-------|------|----------|
| `item` | LolCatalogCatalogPluginItem | Yes |
| `quantity` | uint32 | Yes |
| `requiredItems` | LolCatalogCatalogPluginItemWithDetails[] |  |
| `bundledItems` | LolCatalogCatalogPluginItemWithDetails[] |  |
| `minimumBundlePrices` | LolCatalogCatalogPluginPrice[] |  |
| `bundledDiscountPrices` | LolCatalogCatalogPluginPrice[] |  |
| `assets` | LolCatalogCatalogPluginItemAssets | Yes |
| `metadata` | LolCatalogItemMetadataEntry[] | Yes |
| `bundleFinalPrice` | int64 | Yes |
| `flexible` | boolean | Yes |

### LolCatalogCatalogPluginPrice

| Field | Type | Required |
|-------|------|----------|
| `currency` | string | Yes |
| `cost` | int64 | Yes |
| `costType` | string |  |
| `sale` | LolCatalogCatalogPluginRetailDiscount |  |

### LolCatalogInventoryOwnership

**Type**: enum string

**Values**: `F2P`, `LOYALTY`, `RENTED`, `OWNED`


### LolCatalogItemChoiceDetails

| Field | Type | Required |
|-------|------|----------|
| `item` | LolCatalogCatalogPluginItem | Yes |
| `backgroundImage` | string | Yes |
| `contents` | LolCatalogItemDetails[] | Yes |
| `discount` | string | Yes |
| `fullPrice` | int64 | Yes |
| `displayType` | string | Yes |
| `metadata` | object (map) | Yes |

### LolCatalogItemDetails

| Field | Type | Required |
|-------|------|----------|
| `title` | string | Yes |
| `subTitle` | string | Yes |
| `description` | string | Yes |
| `iconUrl` | string | Yes |

### LolCatalogItemMetadataEntry

| Field | Type | Required |
|-------|------|----------|
| `type` | string | Yes |
| `value` | string | Yes |

### LolCatalogSale

| Field | Type | Required |
|-------|------|----------|
| `startDate` | string | Yes |
| `endDate` | string | Yes |
| `prices` | LolCatalogItemCost[] | Yes |

### LolCatalogSkinLineInfo

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `descriptionInfo` | LolCatalogSkinLineDescriptionInfo[] | Yes |
| `splashPath` | string | Yes |
| `tilePath` | string | Yes |
| `collectionCardPath` | string | Yes |
| `uncenteredSplashPath` | string | Yes |
| `collectionDescription` | string | Yes |
| `tiers` | LolCatalogSkinLineTier[] | Yes |
| `mostProgressedSkinTier` | LolCatalogSkinLineTier | Yes |
| `productType` | string | Yes |

---

## LolChallenges


### LolChallengesChallengeSeason

| Field | Type | Required |
|-------|------|----------|
| `seasonId` | int32 | Yes |
| `seasonStart` | int64 | Yes |
| `seasonEnd` | int64 | Yes |
| `isActive` | boolean | Yes |

### LolChallengesChallengeSignedUpdatePayload

| Field | Type | Required |
|-------|------|----------|
| `tokensByType` | object (map) | Yes |

### LolChallengesChallengeTitleData

| Field | Type | Required |
|-------|------|----------|
| `challengeId` | int64 | Yes |
| `challengeName` | string | Yes |
| `challengeDescription` | string | Yes |
| `level` | string | Yes |
| `levelToIconPath` | object (map) | Yes |

### LolChallengesChallengesPlayerPreferences

| Field | Type | Required |
|-------|------|----------|
| `bannerAccent` | string | Yes |
| `title` | string | Yes |
| `challengeIds` | integer[] | Yes |
| `crestBorder` | string | Yes |
| `prestigeCrestBorderLevel` | int32 | Yes |
| `signedJWTPayload` | LolChallengesChallengeSignedUpdatePayload | Yes |

### LolChallengesClientState

**Type**: enum string

**Values**: `Enabled`, `DarkDisabled`, `DarkHidden`, `Disabled`, `Hidden`


### LolChallengesUICategoryProgress

| Field | Type | Required |
|-------|------|----------|
| `level` | string | Yes |
| `category` | string | Yes |
| `positionPercentile` | number | Yes |
| `current` | int32 | Yes |
| `max` | int32 | Yes |

### LolChallengesUIChallenge

| Field | Type | Required |
|-------|------|----------|
| `id` | int64 | Yes |
| `name` | string | Yes |
| `description` | string | Yes |
| `descriptionShort` | string | Yes |
| `iconPath` | string | Yes |
| `category` | string | Yes |
| `nextLevelIconPath` | string | Yes |
| `currentLevel` | string | Yes |
| `nextLevel` | string | Yes |
| `previousLevel` | string | Yes |
| `previousValue` | number | Yes |
| `currentValue` | number | Yes |
| `currentThreshold` | number | Yes |
| `nextThreshold` | number | Yes |
| `pointsAwarded` | int64 | Yes |
| `percentile` | number | Yes |
| `currentLevelAchievedTime` | int64 | Yes |
| `position` | int32 | Yes |
| `playersInLevel` | int32 | Yes |
| `isApex` | boolean | Yes |
| `isCapstone` | boolean | Yes |
| `gameModes` | string[] | Yes |
| `friendsAtLevels` | LolChallengesFriendLevelsData[] | Yes |
| `parentId` | int64 | Yes |
| `parentName` | string | Yes |
| `childrenIds` | integer[] | Yes |
| `capstoneGroupId` | int64 | Yes |
| `capstoneGroupName` | string | Yes |
| `source` | string | Yes |
| `thresholds` | object (map) | Yes |
| `levelToIconPath` | object (map) | Yes |
| `valueMapping` | string | Yes |
| `hasLeaderboard` | boolean | Yes |
| `isReverseDirection` | boolean | Yes |
| `priority` | number | Yes |
| `idListType` | LolChallengesChallengeRequirementMappingName | Yes |
| `availableIds` | integer[] | Yes |
| `completedIds` | integer[] | Yes |
| `retireTimestamp` | int64 | Yes |

### LolChallengesUIChallengePenalty

| Field | Type | Required |
|-------|------|----------|
| `reason` | string | Yes |

### LolChallengesUIPlayerSummary

| Field | Type | Required |
|-------|------|----------|
| `totalChallengeScore` | int64 | Yes |
| `pointsUntilNextRank` | int64 | Yes |
| `overallChallengeLevel` | string | Yes |
| `positionPercentile` | number | Yes |
| `isApex` | boolean | Yes |
| `apexLeaderboardPosition` | int32 | Yes |
| `title` | LolChallengesUITitle | Yes |
| `bannerId` | string | Yes |
| `crestId` | string | Yes |
| `prestigeCrestBorderLevel` | int32 | Yes |
| `categoryProgress` | LolChallengesUICategoryProgress[] | Yes |
| `topChallenges` | LolChallengesUIChallenge[] | Yes |
| `apexLadderUpdateTime` | int64 | Yes |
| `selectedChallengesString` | string | Yes |

### LolChallengesUITitle

| Field | Type | Required |
|-------|------|----------|
| `itemId` | int32 | Yes |
| `contentId` | string | Yes |
| `name` | string | Yes |
| `purchaseDate` | string | Yes |
| `titleAcquisitionType` | string | Yes |
| `titleAcquisitionName` | string |  |
| `titleRequirementDescription` | string |  |
| `isPermanentTitle` | boolean |  |
| `challengeTitleData` | LolChallengesChallengeTitleData |  |
| `iconPath` | string |  |
| `backgroundImagePath` | string |  |

---

## LolChamp


### LolChampSelectLegacyChampSelectAction

| Field | Type | Required |
|-------|------|----------|
| `id` | int64 | Yes |
| `actorCellId` | int64 | Yes |
| `championId` | int32 | Yes |
| `type` | string | Yes |
| `completed` | boolean | Yes |
| `isAllyAction` | boolean | Yes |
| `isInProgress` | boolean | Yes |
| `pickTurn` | int32 | Yes |

### LolChampSelectLegacyChampSelectBannedChampions

| Field | Type | Required |
|-------|------|----------|
| `myTeamBans` | integer[] | Yes |
| `theirTeamBans` | integer[] | Yes |
| `numBans` | int32 | Yes |

### LolChampSelectLegacyChampSelectChatRoomDetails

| Field | Type | Required |
|-------|------|----------|
| `multiUserChatId` | string | Yes |
| `multiUserChatPassword` | string | Yes |
| `mucJwtDto` | LolChampSelectLegacyMucJwtDto | Yes |

### LolChampSelectLegacyChampSelectMySelection

| Field | Type | Required |
|-------|------|----------|
| `selectedSkinId` | int32 |  |
| `spell1Id` | uint64 |  |
| `spell2Id` | uint64 |  |

### LolChampSelectLegacyChampSelectPlayerSelection

| Field | Type | Required |
|-------|------|----------|
| `cellId` | int64 | Yes |
| `championId` | int32 | Yes |
| `selectedSkinId` | int32 | Yes |
| `spell1Id` | uint64 | Yes |
| `spell2Id` | uint64 | Yes |
| `team` | int32 | Yes |
| `assignedPosition` | string | Yes |
| `championPickIntent` | int32 | Yes |
| `playerType` | string | Yes |
| `summonerId` | uint64 | Yes |
| `puuid` | string | Yes |

### LolChampSelectLegacyChampSelectSession

| Field | Type | Required |
|-------|------|----------|
| `timer` | LolChampSelectLegacyChampSelectTimer | Yes |
| `chatDetails` | LolChampSelectLegacyChampSelectChatRoomDetails | Yes |
| `myTeam` | LolChampSelectLegacyChampSelectPlayerSelection[] | Yes |
| `theirTeam` | LolChampSelectLegacyChampSelectPlayerSelection[] | Yes |
| `trades` | LolChampSelectLegacyChampSelectSwapContract[] | Yes |
| `actions` | object[] | Yes |
| `bans` | LolChampSelectLegacyChampSelectBannedChampions | Yes |
| `localPlayerCellId` | int64 | Yes |
| `isSpectating` | boolean | Yes |
| `allowSkinSelection` | boolean | Yes |
| `allowBattleBoost` | boolean | Yes |
| `allowRerolling` | boolean | Yes |
| `rerollsRemaining` | uint64 | Yes |
| `hasSimultaneousBans` | boolean | Yes |
| `hasSimultaneousPicks` | boolean | Yes |
| `showQuitButton` | boolean | Yes |
| `isLegacyChampSelect` | boolean | Yes |
| `isCustomGame` | boolean | Yes |

### LolChampSelectLegacyChampSelectSwapContract

| Field | Type | Required |
|-------|------|----------|
| `id` | int64 | Yes |
| `cellId` | int64 | Yes |
| `state` | LolChampSelectLegacyChampSelectSwapState | Yes |

### LolChampSelectLegacyChampSelectSwapState

**Type**: enum string

**Values**: `SENT`, `RECEIVED`, `INVALID`, `BUSY`, `AVAILABLE`


### LolChampSelectLegacyChampSelectTimer

| Field | Type | Required |
|-------|------|----------|
| `adjustedTimeLeftInPhase` | int64 | Yes |
| `totalTimeInPhase` | int64 | Yes |
| `phase` | string | Yes |
| `isInfinite` | boolean | Yes |
| `internalNowInEpochMs` | uint64 | Yes |

### LolChampSelectLegacyTeamBoost

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | int64 | Yes |
| `puuid` | string | Yes |
| `skinUnlockMode` | string | Yes |
| `price` | int64 | Yes |
| `ipReward` | int64 | Yes |
| `ipRewardForPurchaser` | int64 | Yes |
| `availableSkins` | integer[] | Yes |
| `unlocked` | boolean | Yes |

---

## LolChampion


### LolChampionMasteryChampionMastery

| Field | Type | Required |
|-------|------|----------|
| `puuid` | string | Yes |
| `championId` | int32 | Yes |
| `championLevel` | int32 | Yes |
| `championPoints` | int32 | Yes |
| `lastPlayTime` | uint64 | Yes |
| `championPointsSinceLastLevel` | int32 | Yes |
| `championPointsUntilNextLevel` | int32 | Yes |
| `markRequiredForNextLevel` | int32 | Yes |
| `tokensEarned` | int32 | Yes |
| `championSeasonMilestone` | int32 | Yes |
| `milestoneGrades` | string[] | Yes |
| `nextSeasonMilestone` | LolChampionMasterySeasonMilestoneRequireAndRewards | Yes |
| `highestGrade` | string | Yes |

### LolChampionMasteryChampionMasteryChangeNotification

| Field | Type | Required |
|-------|------|----------|
| `gameId` | int64 | Yes |
| `puuid` | string | Yes |
| `championId` | int32 | Yes |
| `championLevel` | int32 | Yes |
| `championPointsBeforeGame` | int32 | Yes |
| `championPointsGained` | int32 | Yes |
| `championPointsGainedIndividualContribution` | int32 | Yes |
| `bonusChampionPointsGained` | int32 | Yes |
| `playerGrade` | string | Yes |
| `championPointsSinceLastLevelBeforeGame` | int32 | Yes |
| `championPointsUntilNextLevelBeforeGame` | int32 | Yes |
| `championPointsUntilNextLevelAfterGame` | int32 | Yes |
| `championLevelUp` | boolean | Yes |
| `score` | int32 | Yes |
| `levelUpList` | LolChampionMasteryChampionMasteryMini[] | Yes |
| `memberGrades` | LolChampionMasteryChampionMasteryGrade[] | Yes |
| `win` | boolean | Yes |
| `mapId` | int32 | Yes |
| `tokensEarned` | int32 | Yes |
| `tokenEarnedAfterGame` | boolean | Yes |
| `markRequiredForNextLevel` | int32 | Yes |
| `championSeasonMilestone` | int32 | Yes |
| `championSeasonMilestoneUp` | boolean | Yes |
| `milestoneGrades` | string[] | Yes |
| `seasonMilestone` | LolChampionMasterySeasonMilestoneRequireAndRewards | Yes |

### LolChampionMasteryChampionMasteryGrade

| Field | Type | Required |
|-------|------|----------|
| `puuid` | string | Yes |
| `championId` | int32 | Yes |
| `grade` | string | Yes |

### LolChampionMasteryChampionMasteryMini

| Field | Type | Required |
|-------|------|----------|
| `puuid` | string | Yes |
| `championId` | int32 | Yes |
| `championLevel` | int32 | Yes |

### LolChampionMasteryChampionMasteryRewardGrantNotification

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `gameId` | int64 | Yes |
| `puuid` | string | Yes |
| `championId` | int32 | Yes |
| `playerGrade` | string | Yes |
| `messageKey` | string | Yes |

### LolChampionMasteryChampionMasteryViewData

| Field | Type | Required |
|-------|------|----------|
| `championId` | int64 | Yes |
| `championLevel` | int32 | Yes |
| `championPoints` | int32 | Yes |
| `highestGrade` | string |  |

### LolChampionMasteryChampionSet

| Field | Type | Required |
|-------|------|----------|
| `champions` | integer[] | Yes |
| `totalMilestone` | int32 | Yes |
| `completed` | boolean | Yes |

### LolChampionMasterySeasonMilestoneRequireAndRewards

| Field | Type | Required |
|-------|------|----------|
| `requireGradeCounts` | object (map) | Yes |
| `rewardMarks` | uint16 | Yes |
| `bonus` | boolean | Yes |
| `rewardConfig` | LolChampionMasteryRewardConfigurationEntry | Yes |

### LolChampionMasteryTopChampionMasteries

| Field | Type | Required |
|-------|------|----------|
| `puuid` | string | Yes |
| `summonerId` | uint64 | Yes |
| `score` | uint64 | Yes |
| `masteries` | LolChampionMasteryChampionMastery[] | Yes |

### LolChampionMasteryUIAllChampionMasteryWithSets

| Field | Type | Required |
|-------|------|----------|
| `championMasteries` | LolChampionMasteryChampionMastery[] | Yes |
| `championSet` | LolChampionMasteryChampionSet | Yes |
| `championSetRewards` | object (map) | Yes |
| `seasonMilestoneRequireAndRewards` | object (map) | Yes |
| `defaultChampionMastery` | LolChampionMasteryChampionMastery | Yes |
| `customRewards` | LolChampionMasteryUIChampionMasteryCustomReward[] | Yes |
| `totalScore` | int32 | Yes |
| `championCountByMilestone` | object (map) | Yes |

### LolChampionMasteryUIChampionMasteryCustomReward

| Field | Type | Required |
|-------|------|----------|
| `type` | string | Yes |
| `level` | int32 | Yes |
| `rewardValue` | string | Yes |
| `quantity` | int32 | Yes |

---

## LolChampions


### LolChampionsChampionQuestSkinInfo

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `descriptionInfo` | LolChampionsQuestSkinDescriptionInfo[] | Yes |
| `splashPath` | string | Yes |
| `uncenteredSplashPath` | string | Yes |
| `tilePath` | string | Yes |
| `collectionCardPath` | string | Yes |
| `collectionDescription` | string | Yes |
| `tiers` | LolChampionsCollectionsChampionQuestSkin[] | Yes |
| `productType` | LolChampionsQuestSkinProductType |  |

### LolChampionsCollectionsChampion

| Field | Type | Required |
|-------|------|----------|
| `alias` | string | Yes |
| `title` | string | Yes |
| `banVoPath` | string | Yes |
| `chooseVoPath` | string | Yes |
| `disabledQueues` | string[] | Yes |
| `id` | int32 | Yes |
| `name` | string | Yes |
| `ownership` | LolChampionsCollectionsOwnership | Yes |
| `purchased` | uint64 | Yes |
| `roles` | string[] | Yes |
| `squarePortraitPath` | string | Yes |
| `stingerSfxPath` | string | Yes |
| `baseLoadScreenPath` | string | Yes |
| `baseSplashPath` | string | Yes |
| `active` | boolean | Yes |
| `botEnabled` | boolean | Yes |
| `freeToPlay` | boolean | Yes |
| `rankedPlayEnabled` | boolean | Yes |
| `isVisibleInClient` | boolean | Yes |
| `passive` | LolChampionsCollectionsChampionSpell | Yes |
| `skins` | LolChampionsCollectionsChampionSkin[] | Yes |
| `spells` | LolChampionsCollectionsChampionSpell[] | Yes |
| `tacticalInfo` | LolChampionsCollectionsChampionTacticalInfo | Yes |

### LolChampionsCollectionsChampionChroma

| Field | Type | Required |
|-------|------|----------|
| `championId` | int32 | Yes |
| `chromaPath` | string |  |
| `id` | int32 | Yes |
| `name` | string | Yes |
| `ownership` | LolChampionsCollectionsOwnership | Yes |
| `disabled` | boolean | Yes |
| `stillObtainable` | boolean | Yes |
| `lastSelected` | boolean | Yes |
| `skinAugments` | LolChampionsCollectionsChampionSkinAugments | Yes |
| `colors` | string[] | Yes |

### LolChampionsCollectionsChampionMinimal

| Field | Type | Required |
|-------|------|----------|
| `alias` | string | Yes |
| `title` | string | Yes |
| `banVoPath` | string | Yes |
| `chooseVoPath` | string | Yes |
| `disabledQueues` | string[] | Yes |
| `id` | int32 | Yes |
| `name` | string | Yes |
| `ownership` | LolChampionsCollectionsOwnership | Yes |
| `purchased` | uint64 | Yes |
| `roles` | string[] | Yes |
| `squarePortraitPath` | string | Yes |
| `stingerSfxPath` | string | Yes |
| `baseLoadScreenPath` | string | Yes |
| `baseSplashPath` | string | Yes |
| `active` | boolean | Yes |
| `botEnabled` | boolean | Yes |
| `freeToPlay` | boolean | Yes |
| `rankedPlayEnabled` | boolean | Yes |
| `isVisibleInClient` | boolean | Yes |

### LolChampionsCollectionsChampionPlayableCounts

| Field | Type | Required |
|-------|------|----------|
| `championsOwned` | uint32 | Yes |
| `championsRented` | uint32 | Yes |
| `championsFreeToPlay` | uint32 | Yes |
| `championsLoyaltyReward` | uint32 | Yes |
| `championsXboxGPReward` | uint32 | Yes |

### LolChampionsCollectionsChampionSkin

| Field | Type | Required |
|-------|------|----------|
| `championId` | int32 | Yes |
| `chromaPath` | string |  |
| `id` | int32 | Yes |
| `name` | string | Yes |
| `ownership` | LolChampionsCollectionsOwnership | Yes |
| `isBase` | boolean | Yes |
| `disabled` | boolean | Yes |
| `stillObtainable` | boolean | Yes |
| `lastSelected` | boolean | Yes |
| `skinAugments` | LolChampionsCollectionsChampionSkinAugments | Yes |
| `splashPath` | string | Yes |
| `tilePath` | string | Yes |
| `chromas` | LolChampionsCollectionsChampionChroma[] | Yes |
| `questSkinInfo` | LolChampionsChampionQuestSkinInfo | Yes |
| `emblems` | LolChampionsCollectionsChampionSkinEmblem[] | Yes |
| `uncenteredSplashPath` | string | Yes |
| `loadScreenPath` | string | Yes |
| `rarityGemPath` | string | Yes |
| `splashVideoPath` | string |  |
| `collectionSplashVideoPath` | string |  |
| `skinType` | string |  |
| `featuresText` | string |  |

### LolChampionsCollectionsChampionSkinAugments

| Field | Type | Required |
|-------|------|----------|
| `augments` | LolChampionsCollectionsChampionSkinAugment[] | Yes |

### LolChampionsCollectionsChampionSkinEmblem

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `emblemPath` | LolChampionsCollectionsChampionSkinEmblemPath | Yes |
| `positions` | LolChampionsCollectionsChampionSkinEmblemPosition | Yes |

### LolChampionsCollectionsChampionSkinMinimal

| Field | Type | Required |
|-------|------|----------|
| `championId` | int32 | Yes |
| `chromaPath` | string |  |
| `id` | int32 | Yes |
| `name` | string | Yes |
| `ownership` | LolChampionsCollectionsOwnership | Yes |
| `isBase` | boolean | Yes |
| `disabled` | boolean | Yes |
| `stillObtainable` | boolean | Yes |
| `lastSelected` | boolean | Yes |
| `skinAugments` | LolChampionsCollectionsChampionSkinAugments | Yes |
| `splashPath` | string | Yes |
| `tilePath` | string | Yes |

### LolChampionsCollectionsChampionSpell

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `description` | string | Yes |

### LolChampionsCollectionsChampionTacticalInfo

| Field | Type | Required |
|-------|------|----------|
| `style` | uint32 | Yes |
| `difficulty` | uint32 | Yes |
| `damageType` | string | Yes |

### LolChampionsCollectionsOwnership

| Field | Type | Required |
|-------|------|----------|
| `loyaltyReward` | boolean | Yes |
| `xboxGPReward` | boolean | Yes |
| `owned` | boolean | Yes |
| `rental` | LolChampionsCollectionsRental | Yes |

---

## LolChat


### LolChatActiveConversationResource

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |

### LolChatBlockedPlayerResource

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `icon` | int32 | Yes |
| `id` | string | Yes |
| `name` | string | Yes |
| `pid` | string | Yes |
| `puuid` | string | Yes |
| `gameName` | string | Yes |
| `gameTag` | string | Yes |

### LolChatChatDomainConfig

| Field | Type | Required |
|-------|------|----------|
| `P2PDomainName` | string |  |
| `CustomGameDomainName` | string |  |
| `ChampSelectDomainName` | string |  |
| `PostGameDomainName` | string |  |
| `ClubDomainName` | string |  |

### LolChatChatServiceDynamicClientConfig

| Field | Type | Required |
|-------|------|----------|
| `LcuSocial` | LolChatLcuSocialConfig |  |
| `ChatDomain` | LolChatChatDomainConfig |  |

### LolChatConversationMessageResource

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `type` | string | Yes |
| `fromSummonerId` | uint64 | Yes |
| `fromPuuid` | string | Yes |
| `fromId` | string | Yes |
| `fromPid` | string | Yes |
| `fromObfuscatedSummonerId` | uint64 | Yes |
| `fromObfuscatedPuuid` | string | Yes |
| `body` | string | Yes |
| `timestamp` | string | Yes |
| `isHistorical` | boolean | Yes |

### LolChatConversationResource

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `name` | string | Yes |
| `pid` | string | Yes |
| `gameName` | string | Yes |
| `gameTag` | string | Yes |
| `type` | string | Yes |
| `inviterId` | string | Yes |
| `password` | string | Yes |
| `targetRegion` | string | Yes |
| `isMuted` | boolean | Yes |
| `unreadMessageCount` | uint64 | Yes |
| `lastMessage` | LolChatConversationMessageResource |  |
| `mucJwtDto` | LolChatMucJwtDto | Yes |

### LolChatErrorResource

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `from` | string | Yes |
| `code` | uint64 | Yes |
| `message` | string | Yes |
| `text` | string | Yes |

### LolChatFriendCountsResource

| Field | Type | Required |
|-------|------|----------|
| `numFriends` | uint32 | Yes |
| `numFriendsOnline` | uint32 | Yes |
| `numFriendsAvailable` | uint32 | Yes |
| `numFriendsAway` | uint32 | Yes |
| `numFriendsInQueue` | uint32 | Yes |
| `numFriendsInChampSelect` | uint32 | Yes |
| `numFriendsInGame` | uint32 | Yes |
| `numFriendsMobile` | uint32 | Yes |

### LolChatFriendGroupOrder

| Field | Type | Required |
|-------|------|----------|
| `groups` | string[] | Yes |

### LolChatFriendRequestDirection

**Type**: enum string

**Values**: `both`, `out`, `in`


### LolChatFriendRequestResource

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `icon` | int32 | Yes |
| `id` | string | Yes |
| `name` | string | Yes |
| `pid` | string | Yes |
| `puuid` | string | Yes |
| `gameName` | string | Yes |
| `tagLine` | string | Yes |
| `note` | string | Yes |
| `direction` | LolChatFriendRequestDirection | Yes |

### LolChatFriendResource

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `id` | string | Yes |
| `name` | string | Yes |
| `pid` | string | Yes |
| `puuid` | string | Yes |
| `gameName` | string | Yes |
| `gameTag` | string | Yes |
| `icon` | int32 | Yes |
| `availability` | string | Yes |
| `platformId` | string | Yes |
| `patchline` | string | Yes |
| `product` | string | Yes |
| `productName` | string | Yes |
| `summary` | string | Yes |
| `time` | uint64 | Yes |
| `statusMessage` | string | Yes |
| `note` | string | Yes |
| `lastSeenOnlineTimestamp` | string |  |
| `isP2PConversationMuted` | boolean | Yes |
| `groupId` | uint32 | Yes |
| `displayGroupId` | uint32 | Yes |
| `groupName` | string | Yes |
| `displayGroupName` | string | Yes |
| `lol` | object (map) | Yes |
| `discordInfo` | LolChatSocialV4DiscordInfo |  |
| `discordOnlineStatus` | string |  |
| `discordId` | string |  |

### LolChatGroupResource

| Field | Type | Required |
|-------|------|----------|
| `id` | uint32 | Yes |
| `name` | string | Yes |
| `isMetaGroup` | boolean | Yes |
| `isLocalized` | boolean | Yes |
| `priority` | int32 | Yes |
| `collapsed` | boolean | Yes |

### LolChatLcuSocialConfig

| Field | Type | Required |
|-------|------|----------|
| `ForceChatFilter` | boolean | Yes |
| `QueueJobGraceSeconds` | uint64 | Yes |
| `SilenceChatWhileInGame` | boolean | Yes |
| `AggressiveScanning` | boolean | Yes |
| `ReplaceRichMessages` | boolean | Yes |
| `gameNameTaglineEnabled` | boolean | Yes |
| `allowGroupByGame` | boolean | Yes |
| `platformToRegionMap` | object (map) | Yes |

### LolChatMucJwtDto

| Field | Type | Required |
|-------|------|----------|
| `jwt` | string | Yes |
| `channelClaim` | string | Yes |
| `domain` | string | Yes |
| `targetRegion` | string | Yes |

### LolChatPlayerMuteUpdate

| Field | Type | Required |
|-------|------|----------|
| `puuids` | string[] | Yes |
| `isMuted` | boolean | Yes |

### LolChatProductMetadataMap

| Field | Type | Required |
|-------|------|----------|
| `products` | object (map) | Yes |

### LolChatSessionResource

| Field | Type | Required |
|-------|------|----------|
| `sessionState` | LolChatSessionState | Yes |
| `sessionExpire` | uint32 | Yes |

### LolChatSessionState

**Type**: enum string

**Values**: `shuttingdown`, `disconnected`, `loaded`, `connected`, `initializing`


### LolChatSocialV4DiscordInfo

| Field | Type | Required |
|-------|------|----------|
| `displayName` | string | Yes |
| `onlineStatus` | string | Yes |
| `isPlayingSameTitle` | boolean | Yes |
| `relationship` | string | Yes |
| `discordId` | string | Yes |

### LolChatUserResource

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `id` | string | Yes |
| `name` | string | Yes |
| `pid` | string | Yes |
| `puuid` | string | Yes |
| `obfuscatedSummonerId` | uint64 | Yes |
| `gameName` | string | Yes |
| `gameTag` | string | Yes |
| `icon` | int32 | Yes |
| `availability` | string | Yes |
| `platformId` | string | Yes |
| `patchline` | string | Yes |
| `product` | string | Yes |
| `productName` | string | Yes |
| `summary` | string | Yes |
| `time` | uint64 | Yes |
| `statusMessage` | string |  |
| `lastSeenOnlineTimestamp` | string |  |
| `lol` | object (map) | Yes |

---

## LolClash


### LolClashBracket

| Field | Type | Required |
|-------|------|----------|
| `tournamentId` | int64 | Yes |
| `id` | int64 | Yes |
| `size` | int32 | Yes |
| `matches` | BracketMatch[] | Yes |
| `rosters` | BracketRoster[] | Yes |
| `version` | int32 | Yes |
| `period` | int32 | Yes |
| `isComplete` | boolean | Yes |

### LolClashChangeIconRequest

| Field | Type | Required |
|-------|------|----------|
| `iconId` | int32 | Yes |
| `iconColorId` | int32 | Yes |

### LolClashChangeNameRequest

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |

### LolClashClashDisabledConfig

| Field | Type | Required |
|-------|------|----------|
| `disabledReason` | string | Yes |
| `estimatedEnableTimeMillis` | uint64 | Yes |

### LolClashClientFailedInvite

| Field | Type | Required |
|-------|------|----------|
| `playerId` | uint64 | Yes |
| `exception` | string | Yes |

### LolClashEogPlayerUpdateDTO

| Field | Type | Required |
|-------|------|----------|
| `tournamentId` | int64 | Yes |
| `gameId` | int64 | Yes |
| `winner` | boolean | Yes |
| `themeVp` | int32 | Yes |
| `seasonVp` | int32 | Yes |
| `lowestPosition` | int32 | Yes |
| `bracketSize` | int32 | Yes |
| `bid` | int32 | Yes |
| `tier` | int32 | Yes |
| `earnedRewards` | ClashRewardDefinition[] | Yes |
| `rewardProgress` | ClashRewardDefinition[] | Yes |

### LolClashFindPlayers

| Field | Type | Required |
|-------|------|----------|
| `invitationId` | string | Yes |
| `memberId` | int64 | Yes |
| `page` | int32 | Yes |
| `count` | int32 | Yes |

### LolClashFindTeams

| Field | Type | Required |
|-------|------|----------|
| `tournamentId` | int64 | Yes |
| `page` | int32 | Yes |
| `count` | int32 | Yes |

### LolClashKickRequest

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |

### LolClashLftState

| Field | Type | Required |
|-------|------|----------|
| `lft` | boolean | Yes |
| `primaryPos` | string | Yes |
| `secondaryPos` | string | Yes |

### LolClashMucJwtDto

| Field | Type | Required |
|-------|------|----------|
| `jwt` | string | Yes |
| `channelClaim` | string | Yes |
| `domain` | string | Yes |
| `targetRegion` | string | Yes |

### LolClashNotifyReason

**Type**: enum string

**Values**: `TEAMMATE_UNBAN`, `TEAMMATE_BAN`, `MEMBER_BAN`, `UNBAN`, `BAN`, `REVERTED_REGISTRATION`, `REWARD_GRANT_RETRY`, `REWARD_GRANT_FAILED`, `ACCEPT_TICKET`, `DECLINE_TICKET`, `REVOKED_TICKET`, `OFFER_TICKET`, `SET_TICKET`, `KICK`, `CAPTAIN_LEAVE`, `LEAVE`, `REVOKE_INVITE`, `ACCEPT_INVITE`, `DECLINE_INVITE`, `RESENT_INVITE`, `INVITE`, `CHANGE_LFT`, `CHANGE_NAMETAGLOGO`, `CHANGE_POSITION`, `CHANGE_SHORTNAME`, `CHANGE_NAME`, `CHANGE_LOGO`, `OWNER_TRANSFER`, `ROSTER_DELETE`, `DISMISS`, `OWNER_CLOSE`, `UNREADY`, `READY`, `SELFJOIN`, `REVOKE_SELFJOIN`, `ACCEPT_SELFJOIN`, `DECLINE_SELFJOIN`, `REVOKE_SUGGESTION`, `ACCEPT_SUGGESTION`, `DECLINE_SUGGESTION`, `SUGGESTION`


### LolClashOfferTicketRequest

| Field | Type | Required |
|-------|------|----------|
| `ticketAmount` | int32 | Yes |
| `ticketType` | TicketType | Yes |

### LolClashPlayerChatRoster

| Field | Type | Required |
|-------|------|----------|
| `tournamentId` | int64 | Yes |
| `startTimeMs` | int64 | Yes |
| `endTimeMs` | int64 | Yes |
| `tournamentState` | LolClashTournamentState | Yes |
| `playerState` | LolClashPlayerState | Yes |
| `isRegistered` | boolean | Yes |
| `name` | string | Yes |
| `shortName` | string | Yes |
| `iconId` | int32 | Yes |
| `iconColorId` | int32 | Yes |
| `logoUrl` | string | Yes |
| `invitationId` | string | Yes |
| `multiUserChatId` | string | Yes |
| `multiUserChatPassword` | string | Yes |
| `mucJwtDto` | LolClashMucJwtDto | Yes |

### LolClashPlayerData

| Field | Type | Required |
|-------|------|----------|
| `tickets` | object (map) | Yes |
| `isClashBanned` | boolean | Yes |
| `tier` | int32 | Yes |
| `lft` | boolean | Yes |
| `primaryPos` | string | Yes |
| `secondaryPos` | string | Yes |

### LolClashPlayerNotification

| Field | Type | Required |
|-------|------|----------|
| `source` | string | Yes |
| `type` | string | Yes |
| `id` | uint64 | Yes |
| `backgroundUrl` | string | Yes |
| `data` | object (map) | Yes |
| `state` | string | Yes |
| `iconUrl` | string | Yes |
| `titleKey` | string | Yes |
| `detailKey` | string | Yes |
| `created` | string | Yes |
| `expires` | string | Yes |
| `critical` | boolean | Yes |
| `dismissible` | boolean | Yes |

### LolClashPlayerNotificationData

| Field | Type | Required |
|-------|------|----------|
| `notifyReason` | LolClashNotifyReason | Yes |
| `rosterNotifyReason` | LolClashRosterNotifyReason | Yes |
| `tournamentNotifyReason` | LolClashTournamentNotifyReason | Yes |
| `sourceSummonerId` | uint64 | Yes |
| `targetSummonerId` | uint64 | Yes |
| `notification` | LolClashPlayerNotification | Yes |
| `keySuffix` | string | Yes |

### LolClashPlayerRewards

| Field | Type | Required |
|-------|------|----------|
| `seasonVp` | int32 | Yes |
| `themeVp` | LolClashThemeVp[] | Yes |

### LolClashPlayerState

**Type**: enum string

**Values**: `ELIMINATED`, `BRACKET_ROSTER`, `REGISTERED_ROSTER`, `PENDING_ROSTER`, `NO_ROSTER`


### LolClashPlayerTournamentData

| Field | Type | Required |
|-------|------|----------|
| `state` | LolClashPlayerState | Yes |
| `rosterId` | string | Yes |
| `bracketId` | int64 | Yes |

### LolClashPlaymodeRestrictedInfo

| Field | Type | Required |
|-------|------|----------|
| `isRestricted` | boolean | Yes |
| `tournamentId` | int64 | Yes |
| `presenceState` | LolClashPresenceState | Yes |
| `rosterId` | string | Yes |
| `phaseId` | int64 | Yes |
| `readyForVoice` | boolean | Yes |

### LolClashPresenceState

**Type**: enum string

**Values**: `SCOUTING`, `LOCKED_IN`, `NONE`


### LolClashRoster

| Field | Type | Required |
|-------|------|----------|
| `tournamentId` | int64 | Yes |
| `invitationId` | string | Yes |
| `id` | string | Yes |
| `name` | string | Yes |
| `shortName` | string | Yes |
| `iconId` | int32 | Yes |
| `iconColorId` | int32 | Yes |
| `captainSummonerId` | uint64 | Yes |
| `tier` | int32 | Yes |
| `points` | int32 | Yes |
| `wins` | int32 | Yes |
| `losses` | int32 | Yes |
| `currentBracketWins` | int32 | Yes |
| `numCompletedPeriods` | int32 | Yes |
| `isEliminated` | boolean | Yes |
| `isRegistered` | boolean | Yes |
| `isActiveInCurrentPhase` | boolean | Yes |
| `isCurrentBracketComplete` | boolean | Yes |
| `highTierVariance` | boolean | Yes |
| `members` | LolClashRosterMember[] | Yes |
| `availableLogos` | RewardLogo[] | Yes |
| `suggestedInvites` | LolClashSuggestedInvite[] | Yes |
| `phaseInfos` | LolClashRosterPhaseInfo[] | Yes |
| `withdraw` | RosterWithdraw |  |
| `isClashBanned` | boolean | Yes |
| `lft` | boolean | Yes |

### LolClashRosterDetails

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `shortName` | string | Yes |
| `iconId` | int32 | Yes |
| `iconColorId` | int32 | Yes |

### LolClashRosterMember

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `state` | LolClashRosterMemberState | Yes |
| `currentBuyin` | int32 | Yes |
| `buyinType` | TicketType | Yes |
| `previousBuyin` | int32 | Yes |
| `incomingOffers` | LolClashTicketOffer[] | Yes |
| `position` | Position | Yes |
| `replacedSummonerId` | uint64 | Yes |
| `tier` | int32 | Yes |
| `inviteType` | InviteType | Yes |
| `inviterId` | uint64 | Yes |

### LolClashRosterNotifyReason

**Type**: enum string

**Values**: `GAME_RESCHEDULED`, `GAME_START_FAILED_OPPONENT`, `GAME_START_FAILED_SUMMONERS`, `GAME_START_FAILED`, `GAME_START_RETRY_OPPONENT`, `GAME_START_RETRY_SUMMONERS`, `GAME_START_RETRY`, `TICKET_COULD_NOT_BE_CHARGED`, `TICKET_REFUNDED`, `TICKET_CHARGED`, `BANNED_SMURF_OPPONENT`, `BANNED_SMURF_TEAMMATE`, `BANNED_SMURF`, `CANNOT_FIND_MATCH`, `BRACKET_ROSTER_REPLACED`, `BRACKET_ROSTER_REMOVED`, `TIER_CHANGED`, `NO_SHOW_PING`, `ROUND_COMPLETE`, `WITHDRAW`, `VOTE_WITHDRAW_DISMISS`, `VOTE_WITHDRAW_UPDATE`, `OWNER_TRANSFER`, `QUEUE_DODGE`, `GAME_END_ERROR`, `GAME_STARTED_ERROR`, `GAME_STARTED`, `GAME_SCHEDULED`, `GAME_COMPLETED`, `PERIOD_SPLIT`, `PERIOD_CANCEL`, `PHASE_BACKOUT`, `PHASE_CHECKIN`, `PHASE_READY`, `PHASE_UNREADY`, `RESTRICTION_AUTO_WIN`, `REGISTERED`, `EOG_PLAYER_UPDATE`, `CHEATER_DETECT`, `CHANGE_POSITION`, `BRACKET_READY`, `BYE_AUTO_WIN`, `ROSTER_REVOKED_TICKET`, `ROSTER_DECLINE_TICKET`, `ROSTER_ACCEPT_TICKET`, `ROSTER_OFFER_TICKET`, `ROSTER_SET_TICKET`


### LolClashRosterPeriodAggregatedStats

| Field | Type | Required |
|-------|------|----------|
| `period` | int32 | Yes |
| `bracketSize` | int32 | Yes |
| `time` | int64 | Yes |
| `matchStats` | LolClashRosterMatchAggregatedStats[] | Yes |
| `playerBids` | object (map) | Yes |

### LolClashRosterPhaseInfo

| Field | Type | Required |
|-------|------|----------|
| `phaseId` | int64 | Yes |
| `period` | int32 | Yes |
| `checkinTime` | int64 | Yes |
| `isBracketComplete` | boolean | Yes |

### LolClashRosterStats

| Field | Type | Required |
|-------|------|----------|
| `rosterId` | int64 | Yes |
| `tournamentThemeId` | int32 | Yes |
| `tournamentNameLocKey` | string | Yes |
| `tournamentNameLocKeySecondary` | string | Yes |
| `startTimeMs` | int64 | Yes |
| `endTimeMs` | int64 | Yes |
| `tournamentPeriods` | int32 | Yes |
| `tier` | int32 | Yes |
| `rosterName` | string | Yes |
| `rosterShortName` | string | Yes |
| `rosterIconId` | int32 | Yes |
| `rosterIconColorId` | int32 | Yes |
| `periodStats` | LolClashRosterPeriodAggregatedStats[] | Yes |
| `playerStats` | object (map) | Yes |

### LolClashScoutingChampionMastery

| Field | Type | Required |
|-------|------|----------|
| `championId` | int32 | Yes |
| `championLevel` | int32 | Yes |
| `championPoints` | int32 | Yes |

### LolClashScoutingChampions

| Field | Type | Required |
|-------|------|----------|
| `playerId` | uint64 | Yes |
| `totalMasteryScore` | uint64 | Yes |
| `topMasteries` | LolClashScoutingChampionMastery[] | Yes |
| `topSeasonChampions` | LolClashScoutingSeasonChampion[] | Yes |

### LolClashScoutingSeasonChampion

| Field | Type | Required |
|-------|------|----------|
| `championId` | int32 | Yes |
| `winCount` | int32 | Yes |
| `gameCount` | int32 | Yes |
| `winRate` | int32 | Yes |
| `kda` | string | Yes |
| `kdaClassification` | LolClashKdaClassification | Yes |

### LolClashSetPositionRequest

| Field | Type | Required |
|-------|------|----------|
| `position` | Position | Yes |

### LolClashSetTicketRequest

| Field | Type | Required |
|-------|------|----------|
| `ticketAmount` | int32 | Yes |
| `ticketType` | TicketType | Yes |

### LolClashSimpleStateFlag

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `status` | LolClashSimpleStateStatus | Yes |

### LolClashSimpleStateStatus

**Type**: enum string

**Values**: `ACKNOWLEDGED`, `UNACKNOWLEDGED`


### LolClashSuggestedInvite

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `suggesterSummonerId` | uint64 | Yes |

### LolClashTeamOpenState

| Field | Type | Required |
|-------|------|----------|
| `invitationId` | string | Yes |
| `captainId` | int64 | Yes |
| `openTeam` | boolean | Yes |

### LolClashThemeVp

| Field | Type | Required |
|-------|------|----------|
| `themeId` | int32 | Yes |
| `vp` | int32 | Yes |

### LolClashThirdPartyApiPlayer

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `role` | string | Yes |

### LolClashThirdPartyApiRoster

| Field | Type | Required |
|-------|------|----------|
| `captain` | LolClashThirdPartyApiPlayer | Yes |
| `members` | LolClashThirdPartyApiPlayer[] | Yes |

### LolClashTournament

| Field | Type | Required |
|-------|------|----------|
| `id` | int64 | Yes |
| `themeId` | int32 | Yes |
| `nameLocKey` | string | Yes |
| `nameLocKeySecondary` | string | Yes |
| `buyInOptions` | integer[] | Yes |
| `buyInOptionsPremium` | integer[] | Yes |
| `entryFee` | int32 | Yes |
| `rosterSize` | int32 | Yes |
| `allowRosterCreation` | boolean | Yes |
| `rosterCreateDeadline` | int64 | Yes |
| `scoutingDurationMs` | int64 | Yes |
| `startTimeMs` | int64 | Yes |
| `endTimeMs` | int64 | Yes |
| `lastThemeOfSeason` | boolean | Yes |
| `bracketSize` | string | Yes |
| `queueId` | int32 | Yes |
| `isSmsRestrictionEnabled` | boolean | Yes |
| `isHonorRestrictionEnabled` | boolean | Yes |
| `isRankedRestrictionEnabled` | boolean | Yes |
| `phases` | LolClashTournamentPhase[] | Yes |
| `rewardConfig` | ClashRewardConfigClient[] | Yes |
| `tierConfigs` | TierConfig[] | Yes |
| `bracketFormationInitDelayMs` | int64 | Yes |
| `bracketFormationIntervalMs` | int64 | Yes |
| `status` | TournamentStatusEnum | Yes |
| `resumeTime` | int64 | Yes |
| `lft` | boolean | Yes |
| `maxInvites` | int32 | Yes |
| `maxSuggestionsPerPlayer` | int32 | Yes |

### LolClashTournamentGameEnd

| Field | Type | Required |
|-------|------|----------|
| `tournamentId` | int64 | Yes |
| `tournamentNameLocKey` | string | Yes |
| `tournamentNameLocKeySecondary` | string | Yes |
| `bracketId` | int64 | Yes |
| `oldBracket` | LolClashBracket |  |

### LolClashTournamentHistoryAndWinners

| Field | Type | Required |
|-------|------|----------|
| `tournamentHistory` | LolClashTournament[] | Yes |
| `tournamentWinners` | LolClashTournamentWinnerHistory | Yes |

### LolClashTournamentNotifyReason

**Type**: enum string

**Values**: `UPDATE_STATUS`, `REVERT_PHASE`, `UPDATE_PHASE`, `ADD_PHASE`, `CANCEL_PERIOD`, `CANCEL_TOURNAMENT`, `UPDATE_TOURNAMENT`, `NEW_TOURNAMENT`


### LolClashTournamentPhase

| Field | Type | Required |
|-------|------|----------|
| `id` | int64 | Yes |
| `tournamentId` | int64 | Yes |
| `period` | int32 | Yes |
| `lockinStartTime` | int64 | Yes |
| `scoutingStartTime` | int64 | Yes |
| `cancelled` | boolean | Yes |
| `limitTiers` | integer[] | Yes |
| `capacityStatus` | CapacityEnum | Yes |

### LolClashTournamentState

**Type**: enum string

**Values**: `RESULTS`, `IN_GAME`, `SCOUTING`, `LOCK_IN`, `IDLE`, `UPCOMING`


### LolClashTournamentStateInfo

| Field | Type | Required |
|-------|------|----------|
| `tournamentId` | int64 | Yes |
| `state` | LolClashTournamentState | Yes |
| `currentPhaseId` | int64 | Yes |
| `nextPhaseId` | int64 | Yes |
| `nextStateChangeTime` | int64 | Yes |
| `numRemainingPeriods` | int32 | Yes |

### LolClashTournamentSummary

| Field | Type | Required |
|-------|------|----------|
| `state` | LolClashTournamentState | Yes |
| `tournamentId` | int64 | Yes |
| `rosterId` | string | Yes |
| `bracketId` | int64 | Yes |

### LolClashTournamentWinnerHistory

| Field | Type | Required |
|-------|------|----------|
| `tournamentId` | int64 | Yes |
| `winners` | LolClashTournamentWinnerInfo[] | Yes |

### LolClashTournamentWinnerInfo

| Field | Type | Required |
|-------|------|----------|
| `rosterId` | int64 | Yes |
| `tier` | int32 | Yes |
| `shortName` | string | Yes |
| `name` | string | Yes |
| `logo` | int32 | Yes |
| `logoColor` | int32 | Yes |
| `createTime` | int64 | Yes |
| `averageWinDuration` | int64 | Yes |
| `playerIds` | integer[] | Yes |

---

## LolCollections


### LolCollectionsCollectionsOwnership

| Field | Type | Required |
|-------|------|----------|
| `loyaltyReward` | boolean | Yes |
| `xboxGPReward` | boolean | Yes |
| `owned` | boolean | Yes |
| `rental` | LolCollectionsCollectionsRental | Yes |

### LolCollectionsCollectionsSummonerBackdrop

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `accountId` | uint64 | Yes |
| `profileIconId` | int32 | Yes |
| `championId` | int32 | Yes |
| `backdropType` | LolCollectionsCollectionsSummonerBackdropType | Yes |
| `backdropImage` | string | Yes |
| `backdropAugments` | LolCollectionsCollectionsSummonerBackdropAugments[] | Yes |
| `backdropVideo` | string | Yes |
| `backdropMaskColor` | string | Yes |
| `puuid` | string | Yes |

### LolCollectionsCollectionsSummonerBackdropAugments

| Field | Type | Required |
|-------|------|----------|
| `centeredLCOverlayPath` | string | Yes |
| `socialCardLCOverlayPath` | string | Yes |

### LolCollectionsCollectionsSummonerBackdropType

**Type**: enum string

**Values**: `specified-skin`, `highest-mastery`, `summoner-icon`, `default`


### LolCollectionsCollectionsSummonerSpells

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `spells` | integer[] | Yes |

### LolCollectionsCollectionsWardSkin

| Field | Type | Required |
|-------|------|----------|
| `id` | int64 | Yes |
| `name` | string | Yes |
| `description` | string | Yes |
| `ownership` | LolCollectionsCollectionsOwnership | Yes |
| `wardImagePath` | string | Yes |
| `wardShadowImagePath` | string | Yes |

---

## LolCosmetics


### LolCosmeticsCompanionsFavoritesViewModel

| Field | Type | Required |
|-------|------|----------|
| `favoriteItems` | LolCosmeticsCosmeticsCompanionViewModel[] | Yes |

### LolCosmeticsCompanionsGroupViewModel

| Field | Type | Required |
|-------|------|----------|
| `groupName` | string | Yes |
| `groupId` | uint32 | Yes |
| `numOwned` | uint32 | Yes |
| `numAvailable` | uint32 | Yes |
| `purchaseDate` | int64 | Yes |
| `items` | LolCosmeticsCosmeticsCompanionViewModel[] | Yes |

### LolCosmeticsCompanionsGroupedViewModel

| Field | Type | Required |
|-------|------|----------|
| `selectedLoadoutItem` | LolCosmeticsCosmeticsCompanionViewModel | Yes |
| `defaultItemId` | int32 | Yes |
| `groups` | LolCosmeticsCompanionsGroupViewModel[] | Yes |

### LolCosmeticsCosmeticsCompanionViewModel

| Field | Type | Required |
|-------|------|----------|
| `contentId` | string | Yes |
| `itemId` | int32 | Yes |
| `name` | string | Yes |
| `description` | string | Yes |
| `loadoutsIcon` | string | Yes |
| `owned` | boolean | Yes |
| `selected` | boolean | Yes |
| `favorited` | boolean | Yes |
| `loyalty` | boolean | Yes |
| `f2p` | boolean | Yes |
| `rarityValue` | uint32 | Yes |
| `purchaseDate` | string | Yes |
| `isRecentItem` | boolean | Yes |
| `species` | string | Yes |
| `groupId` | uint32 | Yes |
| `color` | string | Yes |
| `level` | uint32 | Yes |
| `upgrades` | LolCosmeticsCosmeticsCompanionViewModel[] | Yes |
| `offerData` | LolCosmeticsCapOffer |  |
| `starShardsPrice` | LolCosmeticsCosmeticsOfferPrice | Yes |
| `companionType` | string | Yes |
| `TFTRarity` | string | Yes |

### LolCosmeticsCosmeticsTFTAugmentPillarViewModel

| Field | Type | Required |
|-------|------|----------|
| `contentId` | string | Yes |
| `itemId` | int32 | Yes |
| `name` | string | Yes |
| `description` | string | Yes |
| `loadoutsIcon` | string | Yes |
| `owned` | boolean | Yes |
| `selected` | boolean | Yes |
| `favorited` | boolean | Yes |
| `loyalty` | boolean | Yes |
| `f2p` | boolean | Yes |
| `rarityValue` | uint32 | Yes |
| `purchaseDate` | string | Yes |
| `isRecentItem` | boolean | Yes |
| `groupId` | uint32 | Yes |
| `groupName` | string | Yes |
| `TFTRarity` | string | Yes |

### LolCosmeticsCosmeticsTFTDamageSkinViewModel

| Field | Type | Required |
|-------|------|----------|
| `contentId` | string | Yes |
| `itemId` | int32 | Yes |
| `name` | string | Yes |
| `description` | string | Yes |
| `loadoutsIcon` | string | Yes |
| `owned` | boolean | Yes |
| `selected` | boolean | Yes |
| `favorited` | boolean | Yes |
| `loyalty` | boolean | Yes |
| `f2p` | boolean | Yes |
| `rarityValue` | uint32 | Yes |
| `purchaseDate` | string | Yes |
| `isRecentItem` | boolean | Yes |
| `level` | uint32 | Yes |
| `groupId` | uint32 | Yes |
| `groupName` | string | Yes |
| `upgrades` | LolCosmeticsCosmeticsTFTDamageSkinViewModel[] | Yes |
| `TFTRarity` | string | Yes |

### LolCosmeticsCosmeticsTFTMapSkinViewModel

| Field | Type | Required |
|-------|------|----------|
| `contentId` | string | Yes |
| `itemId` | int32 | Yes |
| `name` | string | Yes |
| `description` | string | Yes |
| `loadoutsIcon` | string | Yes |
| `owned` | boolean | Yes |
| `selected` | boolean | Yes |
| `favorited` | boolean | Yes |
| `loyalty` | boolean | Yes |
| `f2p` | boolean | Yes |
| `rarityValue` | uint32 | Yes |
| `purchaseDate` | string | Yes |
| `isRecentItem` | boolean | Yes |
| `groupId` | uint32 | Yes |
| `groupName` | string | Yes |
| `TFTRarity` | string | Yes |

### LolCosmeticsCosmeticsTFTPlaybookViewModel

| Field | Type | Required |
|-------|------|----------|
| `contentId` | string | Yes |
| `itemId` | int32 | Yes |
| `name` | string | Yes |
| `description` | string | Yes |
| `loadoutsIcon` | string | Yes |
| `owned` | boolean | Yes |
| `selected` | boolean | Yes |
| `loyalty` | boolean | Yes |
| `f2p` | boolean | Yes |
| `rarityValue` | uint32 | Yes |
| `purchaseDate` | string | Yes |
| `isRecentItem` | boolean | Yes |
| `iconPath` | string | Yes |
| `iconPathSmall` | string | Yes |
| `splashPath` | string | Yes |
| `earlyAugments` | LolCosmeticsCosmeticsTFTPlaybookAugment[] | Yes |
| `midAugments` | LolCosmeticsCosmeticsTFTPlaybookAugment[] | Yes |
| `lateAugments` | LolCosmeticsCosmeticsTFTPlaybookAugment[] | Yes |
| `isDisabledInDoubleUp` | boolean | Yes |

### LolCosmeticsCosmeticsTFTZoomSkinViewModel

| Field | Type | Required |
|-------|------|----------|
| `contentId` | string | Yes |
| `itemId` | int32 | Yes |
| `name` | string | Yes |
| `description` | string | Yes |
| `loadoutsIcon` | string | Yes |
| `owned` | boolean | Yes |
| `selected` | boolean | Yes |
| `favorited` | boolean | Yes |
| `loyalty` | boolean | Yes |
| `f2p` | boolean | Yes |
| `rarityValue` | uint32 | Yes |
| `purchaseDate` | string | Yes |
| `isRecentItem` | boolean | Yes |
| `groupId` | uint32 | Yes |
| `groupName` | string | Yes |
| `TFTRarity` | string | Yes |

### LolCosmeticsTFTAugmentPillarGroupViewModel

| Field | Type | Required |
|-------|------|----------|
| `groupName` | string | Yes |
| `groupId` | uint32 | Yes |
| `numOwned` | uint32 | Yes |
| `numAvailable` | uint32 | Yes |
| `purchaseDate` | int64 | Yes |
| `items` | LolCosmeticsCosmeticsTFTZoomSkinViewModel[] | Yes |

### LolCosmeticsTFTAugmentPillarGroupedViewModel

| Field | Type | Required |
|-------|------|----------|
| `selectedLoadoutItem` | LolCosmeticsCosmeticsTFTAugmentPillarViewModel | Yes |
| `defaultItemId` | int32 | Yes |
| `groups` | LolCosmeticsTFTAugmentPillarGroupViewModel[] | Yes |

### LolCosmeticsTFTDamageSkinFavoritesViewModel

| Field | Type | Required |
|-------|------|----------|
| `favoriteItems` | LolCosmeticsCosmeticsTFTDamageSkinViewModel[] | Yes |

### LolCosmeticsTFTDamageSkinGroupViewModel

| Field | Type | Required |
|-------|------|----------|
| `groupName` | string | Yes |
| `groupId` | uint32 | Yes |
| `numOwned` | uint32 | Yes |
| `numAvailable` | uint32 | Yes |
| `purchaseDate` | int64 | Yes |
| `items` | LolCosmeticsCosmeticsTFTDamageSkinViewModel[] | Yes |

### LolCosmeticsTFTDamageSkinGroupedViewModel

| Field | Type | Required |
|-------|------|----------|
| `selectedLoadoutItem` | LolCosmeticsCosmeticsTFTDamageSkinViewModel | Yes |
| `defaultItemId` | int32 | Yes |
| `groups` | LolCosmeticsTFTDamageSkinGroupViewModel[] | Yes |

### LolCosmeticsTFTMapSkinFavoritesViewModel

| Field | Type | Required |
|-------|------|----------|
| `favoriteItems` | LolCosmeticsCosmeticsTFTMapSkinViewModel[] | Yes |

### LolCosmeticsTFTMapSkinGroupViewModel

| Field | Type | Required |
|-------|------|----------|
| `groupName` | string | Yes |
| `groupId` | uint32 | Yes |
| `numOwned` | uint32 | Yes |
| `numAvailable` | uint32 | Yes |
| `purchaseDate` | int64 | Yes |
| `items` | LolCosmeticsCosmeticsTFTMapSkinViewModel[] | Yes |

### LolCosmeticsTFTMapSkinGroupedViewModel

| Field | Type | Required |
|-------|------|----------|
| `selectedLoadoutItem` | LolCosmeticsCosmeticsTFTMapSkinViewModel | Yes |
| `defaultItemId` | int32 | Yes |
| `groups` | LolCosmeticsTFTMapSkinGroupViewModel[] | Yes |

### LolCosmeticsTFTPlaybookGroupViewModel

| Field | Type | Required |
|-------|------|----------|
| `groupName` | string | Yes |
| `groupId` | uint32 | Yes |
| `numOwned` | uint32 | Yes |
| `numAvailable` | uint32 | Yes |
| `items` | LolCosmeticsCosmeticsTFTPlaybookViewModel[] | Yes |

### LolCosmeticsTFTPlaybookGroupedViewModel

| Field | Type | Required |
|-------|------|----------|
| `selectedLoadoutItem` | LolCosmeticsCosmeticsTFTPlaybookViewModel | Yes |
| `defaultItemId` | int32 | Yes |
| `groups` | LolCosmeticsTFTPlaybookGroupViewModel[] | Yes |

### LolCosmeticsTFTZoomSkinFavoritesViewModel

| Field | Type | Required |
|-------|------|----------|
| `favoriteItems` | LolCosmeticsCosmeticsTFTZoomSkinViewModel[] | Yes |

### LolCosmeticsTFTZoomSkinGroupViewModel

| Field | Type | Required |
|-------|------|----------|
| `groupName` | string | Yes |
| `groupId` | uint32 | Yes |
| `numOwned` | uint32 | Yes |
| `numAvailable` | uint32 | Yes |
| `purchaseDate` | int64 | Yes |
| `items` | LolCosmeticsCosmeticsTFTZoomSkinViewModel[] | Yes |

### LolCosmeticsTFTZoomSkinGroupedViewModel

| Field | Type | Required |
|-------|------|----------|
| `selectedLoadoutItem` | LolCosmeticsCosmeticsTFTZoomSkinViewModel | Yes |
| `defaultItemId` | int32 | Yes |
| `groups` | LolCosmeticsTFTZoomSkinGroupViewModel[] | Yes |

---

## LolDirectx


### LolDirectxUpgradeDirectXUpgradeNotificationType

**Type**: enum string

**Values**: `FUTURE_HARDWARE_UPGRADE`, `HARDWARE_UPGRADE`, `NONE`


---

## LolDrops


### LolDropsCapDropTableCounterDTO

| Field | Type | Required |
|-------|------|----------|
| `dropTableId` | string | Yes |
| `count` | uint8 | Yes |

### LolDropsCapDropsDropTableDisplayMetadata

| Field | Type | Required |
|-------|------|----------|
| `isCollectorsBounty` | boolean | Yes |
| `dataAssetId` | string | Yes |
| `nameTraKey` | string | Yes |
| `progressionId` | string | Yes |
| `priority` | uint8 | Yes |
| `tables` | object (map) | Yes |
| `version` | uint8 | Yes |
| `chaseContentId` | string | Yes |
| `chaseContentIds` | string[] | Yes |
| `prestigeContentIds` | string[] | Yes |
| `bountyType` | string | Yes |
| `oddsTree` | LolDropsCapDropsOddsTreeNodeDTO | Yes |

### LolDropsCapDropsDropTableWithPityDTO

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `sourceId` | string | Yes |
| `productId` | string | Yes |
| `startDate` | string | Yes |
| `endDate` | string | Yes |
| `currencyId` | string | Yes |
| `rollOffer` | string | Yes |
| `cost` | uint16 | Yes |
| `displayMetadata` | LolDropsCapDropsDropTableDisplayMetadata | Yes |

### LolDropsCapDropsOddsListEntryDTO

| Field | Type | Required |
|-------|------|----------|
| `contentId` | string | Yes |
| `nodeId` | string | Yes |
| `sourceId` | string | Yes |
| `odds` | number | Yes |

### LolDropsCapDropsOddsTreeNodeDTO

| Field | Type | Required |
|-------|------|----------|
| `nodeId` | string | Yes |
| `sourceId` | string | Yes |
| `odds` | number | Yes |
| `children` | LolDropsCapDropsOddsTreeNodeDTO[] | Yes |
| `quantity` | uint16 | Yes |
| `nameTraKey` | string | Yes |
| `priority` | uint8 | Yes |

---

## LolEmail


### LolEmailVerificationEmailUpdate

| Field | Type | Required |
|-------|------|----------|
| `email` | string | Yes |
| `password` | string | Yes |

### LolEmailVerificationEmailVerificationSession

| Field | Type | Required |
|-------|------|----------|
| `email` | string | Yes |
| `emailVerified` | boolean | Yes |
| `fatalError` | boolean | Yes |

---

## LolEnd


### LolEndOfGameChampionMasteryGrade

| Field | Type | Required |
|-------|------|----------|
| `puuid` | string | Yes |
| `championId` | int32 | Yes |
| `grade` | string | Yes |

### LolEndOfGameChampionMasteryMini

| Field | Type | Required |
|-------|------|----------|
| `puuid` | string | Yes |
| `championId` | int32 | Yes |
| `championLevel` | int64 | Yes |

### LolEndOfGameChampionMasteryUpdate

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `gameId` | uint64 | Yes |
| `puuid` | string | Yes |
| `championId` | int32 | Yes |
| `hasLeveledUp` | boolean | Yes |
| `level` | int64 | Yes |
| `pointsBeforeGame` | int64 | Yes |
| `pointsGained` | int64 | Yes |
| `pointsGainedIndividualContribution` | int64 | Yes |
| `bonusPointsGained` | int64 | Yes |
| `pointsSinceLastLevelBeforeGame` | int64 | Yes |
| `pointsUntilNextLevelBeforeGame` | int64 | Yes |
| `pointsUntilNextLevelAfterGame` | int64 | Yes |
| `tokensEarned` | int64 | Yes |
| `tokenEarnedAfterGame` | boolean | Yes |
| `grade` | string | Yes |
| `score` | int64 | Yes |
| `levelUpList` | LolEndOfGameChampionMasteryMini[] | Yes |
| `memberGrades` | LolEndOfGameChampionMasteryGrade[] | Yes |

### LolEndOfGameEndOfGamePlayer

| Field | Type | Required |
|-------|------|----------|
| `stats` | object (map) | Yes |
| `items` | integer[] | Yes |
| `puuid` | string | Yes |
| `riotIdGameName` | string | Yes |
| `riotIdTagLine` | string | Yes |
| `botPlayer` | boolean | Yes |
| `championId` | int32 | Yes |
| `gameId` | uint64 | Yes |
| `leaver` | boolean | Yes |
| `leaves` | int32 | Yes |
| `level` | int32 | Yes |
| `losses` | int32 | Yes |
| `profileIconId` | int32 | Yes |
| `spell1Id` | int32 | Yes |
| `spell2Id` | int32 | Yes |
| `summonerName` | string | Yes |
| `teamId` | int32 | Yes |
| `wins` | int32 | Yes |
| `summonerId` | uint64 | Yes |
| `selectedPosition` | string | Yes |
| `detectedTeamPosition` | string | Yes |
| `skinSplashPath` | string | Yes |
| `skinTilePath` | string | Yes |
| `skinEmblemPaths` | string[] | Yes |
| `championName` | string | Yes |
| `championSquarePortraitPath` | string | Yes |
| `isLocalPlayer` | boolean | Yes |

### LolEndOfGameEndOfGamePoints

| Field | Type | Required |
|-------|------|----------|
| `pointChangeFromChampionsOwned` | int32 | Yes |
| `pointChangeFromGameplay` | int32 | Yes |
| `pointsUsed` | int32 | Yes |
| `previousPoints` | int32 | Yes |
| `pointsUntilNextReroll` | int32 | Yes |
| `rerollCount` | int32 | Yes |
| `totalPoints` | int32 | Yes |

### LolEndOfGameEndOfGameStats

| Field | Type | Required |
|-------|------|----------|
| `difficulty` | string | Yes |
| `gameId` | uint64 | Yes |
| `gameLength` | int32 | Yes |
| `endOfGameTimestamp` | int64 | Yes |
| `gameMode` | string | Yes |
| `gameMutators` | string[] | Yes |
| `gameType` | string | Yes |
| `invalid` | boolean | Yes |
| `queueType` | string | Yes |
| `ranked` | boolean | Yes |
| `reportGameId` | uint64 | Yes |
| `multiUserChatId` | string | Yes |
| `multiUserChatPassword` | string | Yes |
| `mucJwtDto` | LolEndOfGameMucJwtDto | Yes |
| `teams` | LolEndOfGameEndOfGameTeam[] | Yes |
| `localPlayer` | LolEndOfGameEndOfGamePlayer | Yes |
| `myTeamStatus` | string | Yes |
| `leveledUp` | boolean | Yes |
| `newSpells` | integer[] | Yes |
| `previousLevel` | uint64 | Yes |
| `rpEarned` | int32 | Yes |
| `basePoints` | int32 | Yes |
| `battleBoostIpEarned` | int32 | Yes |
| `boostIpEarned` | int32 | Yes |
| `firstWinBonus` | int32 | Yes |
| `ipEarned` | int32 | Yes |
| `ipTotal` | int32 | Yes |
| `boostXpEarned` | int32 | Yes |
| `experienceEarned` | int32 | Yes |
| `experienceTotal` | int32 | Yes |
| `globalBoostXpEarned` | int32 | Yes |
| `loyaltyBoostXpEarned` | int32 | Yes |
| `xbgpBoostXpEarned` | int32 | Yes |
| `missionsXpEarned` | int32 | Yes |
| `previousXpTotal` | uint64 | Yes |
| `nextLevelXp` | uint64 | Yes |
| `currentLevel` | uint64 | Yes |
| `preLevelUpExperienceTotal` | uint64 | Yes |
| `preLevelUpNextLevelXp` | uint64 | Yes |
| `timeUntilNextFirstWinBonus` | int32 | Yes |
| `causedEarlySurrender` | boolean | Yes |
| `earlySurrenderAccomplice` | boolean | Yes |
| `teamEarlySurrendered` | boolean | Yes |
| `gameEndedInEarlySurrender` | boolean | Yes |
| `rerollData` | LolEndOfGameEndOfGamePoints | Yes |
| `teamBoost` | LolEndOfGameEndOfGameTeamBoost |  |

### LolEndOfGameEndOfGameTeam

| Field | Type | Required |
|-------|------|----------|
| `stats` | object (map) | Yes |
| `players` | LolEndOfGameEndOfGamePlayer[] | Yes |
| `memberStatusString` | string | Yes |
| `name` | string | Yes |
| `tag` | string | Yes |
| `fullId` | string | Yes |
| `teamId` | int32 | Yes |
| `isBottomTeam` | boolean | Yes |
| `isPlayerTeam` | boolean | Yes |
| `isWinningTeam` | boolean | Yes |

### LolEndOfGameEndOfGameTeamBoost

| Field | Type | Required |
|-------|------|----------|
| `summonerName` | string | Yes |
| `skinUnlockMode` | string | Yes |
| `price` | int64 | Yes |
| `ipReward` | int64 | Yes |
| `ipRewardForPurchaser` | int64 | Yes |
| `availableSkins` | integer[] | Yes |
| `unlocked` | boolean | Yes |

### LolEndOfGameGameClientEndOfGameStats

| Field | Type | Required |
|-------|------|----------|
| `gameId` | uint64 | Yes |
| `gameMode` | string | Yes |
| `statsBlock` | object (map) | Yes |
| `queueId` | int32 | Yes |
| `queueType` | string | Yes |
| `isRanked` | boolean | Yes |

### LolEndOfGameMucJwtDto

| Field | Type | Required |
|-------|------|----------|
| `jwt` | string | Yes |
| `channelClaim` | string | Yes |
| `domain` | string | Yes |
| `targetRegion` | string | Yes |

### LolEndOfGameTFTEndOfGameEventPVEViewModel

| Field | Type | Required |
|-------|------|----------|
| `buddyName` | string | Yes |
| `buddyIcon` | string | Yes |
| `bossName` | string | Yes |
| `affixes` | string[] | Yes |

### LolEndOfGameTFTEndOfGamePlayerViewModel

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `summonerName` | string | Yes |
| `riotIdGameName` | string | Yes |
| `riotIdTagLine` | string | Yes |
| `iconId` | int32 | Yes |
| `puuid` | string | Yes |
| `ffaStanding` | uint8 | Yes |
| `health` | uint8 | Yes |
| `rank` | uint8 | Yes |
| `isLocalPlayer` | boolean | Yes |
| `isInteractable` | boolean | Yes |
| `partnerGroupId` | uint8 | Yes |
| `boardPieces` | LolEndOfGameTFTEndOfGamePieceViewModel[] | Yes |
| `augments` | LolEndOfGameTFTEndOfGameItemViewModel[] | Yes |
| `companion` | LolEndOfGameTFTEndOfGameCompanionViewModel | Yes |
| `playbook` | LolEndOfGameTFTEndOfGamePlaybookViewModel | Yes |
| `customAugmentContainer` | LolEndOfGameTFTEndOfGameCustomAugmentContainerViewModel | Yes |
| `setCoreName` | string | Yes |

### LolEndOfGameTFTEndOfGameSkillTreeViewModel

| Field | Type | Required |
|-------|------|----------|
| `eventSkillToScore` | LolEndOfGameTFTEndOfGameSkillScoreTreeViewModel[] | Yes |
| `delta` | int16 | Yes |

### LolEndOfGameTFTEndOfGameViewModel

| Field | Type | Required |
|-------|------|----------|
| `players` | LolEndOfGameTFTEndOfGamePlayerViewModel[] | Yes |
| `localPlayer` | LolEndOfGameTFTEndOfGamePlayerViewModel |  |
| `playerSkillTreeEoG` | LolEndOfGameTFTEndOfGameSkillTreeViewModel |  |
| `eventPVEData` | LolEndOfGameTFTEndOfGameEventPVEViewModel |  |
| `gameLength` | uint32 | Yes |
| `gameId` | uint64 | Yes |
| `queueId` | int32 | Yes |
| `queueType` | string | Yes |
| `isRanked` | boolean | Yes |

---

## LolEsport


### LolEsportStreamNotificationsESportsLiveStreams

| Field | Type | Required |
|-------|------|----------|
| `liveStreams` | LolEsportStreamNotificationsESportsStreams[] | Yes |

### LolEsportStreamNotificationsESportsStreams

| Field | Type | Required |
|-------|------|----------|
| `title` | string | Yes |
| `tournamentDescription` | string | Yes |
| `teamAGuid` | string | Yes |
| `teamAId` | int64 | Yes |
| `teamBGuid` | string | Yes |
| `teamBId` | int64 | Yes |
| `teamAName` | string | Yes |
| `teamBName` | string | Yes |
| `teamAAcronym` | string | Yes |
| `teamBAcronym` | string | Yes |
| `teamALogoUrl` | string | Yes |
| `teamBLogoUrl` | string | Yes |

---

## LolEvent


### LolEventHubActiveEventUIData

| Field | Type | Required |
|-------|------|----------|
| `eventId` | string | Yes |
| `eventInfo` | LolEventHubEventInfoUIData | Yes |

### LolEventHubBundleOfferUIData

| Field | Type | Required |
|-------|------|----------|
| `details` | LolEventHubBundledItemUIData | Yes |
| `initialPrice` | int64 | Yes |
| `finalPrice` | int64 | Yes |
| `futureBalance` | int64 | Yes |
| `isPurchasable` | boolean | Yes |
| `discountPercentage` | number | Yes |
| `bundledItems` | LolEventHubBundledItemUIData[] | Yes |

### LolEventHubBundledItemUIData

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `itemId` | int32 | Yes |
| `description` | string | Yes |
| `inventoryType` | string | Yes |
| `subInventoryType` | string | Yes |
| `splashImage` | string | Yes |
| `owned` | boolean | Yes |
| `quantity` | uint32 | Yes |
| `decoratorBadgeURL` | string | Yes |

### LolEventHubCapOrdersOrderDto

| Field | Type | Required |
|-------|------|----------|
| `data` | LolEventHubCapOrdersDataDto | Yes |
| `meta` | LolEventHubCapOrdersMetaDto | Yes |

### LolEventHubCatalogEntry

| Field | Type | Required |
|-------|------|----------|
| `contentId` | string | Yes |
| `itemId` | int32 | Yes |
| `offerId` | string | Yes |
| `typeId` | string | Yes |

### LolEventHubCategoryOffersUIData

| Field | Type | Required |
|-------|------|----------|
| `category` | LolEventHubOfferCategory | Yes |
| `categoryIconPath` | string | Yes |
| `offers` | LolEventHubOfferUIData[] | Yes |

### LolEventHubChapter

| Field | Type | Required |
|-------|------|----------|
| `localizedTitle` | string | Yes |
| `localizedDescription` | string | Yes |
| `cardImage` | string | Yes |
| `backgroundImage` | string | Yes |
| `objectiveBannerImage` | string | Yes |
| `chapterStart` | uint16 | Yes |
| `chapterEnd` | uint16 | Yes |
| `chapterNumber` | uint16 | Yes |
| `levelFocus` | uint16 | Yes |

### LolEventHubChaptersUIData

| Field | Type | Required |
|-------|------|----------|
| `currentChapter` | uint16 | Yes |
| `chapters` | LolEventHubChapter[] | Yes |

### LolEventHubEventBackgroundUIData

| Field | Type | Required |
|-------|------|----------|
| `backgroundImagePath` | string | Yes |

### LolEventHubEventDetailsUIData

| Field | Type | Required |
|-------|------|----------|
| `eventIconPath` | string | Yes |
| `eventName` | string | Yes |
| `headerTitleImagePath` | string | Yes |
| `progressEndDate` | string | Yes |
| `shopEndDate` | string | Yes |
| `eventStartDate` | string | Yes |
| `helpModalImagePath` | string | Yes |
| `inducteeName` | string | Yes |
| `promotionBannerImage` | string | Yes |
| `objectiveBannerImage` | string | Yes |
| `memoryBookBackgroundImage` | string | Yes |
| `spotlightSkinId` | uint32 | Yes |
| `actBackgroundImage` | string | Yes |

### LolEventHubEventInfoUIData

| Field | Type | Required |
|-------|------|----------|
| `eventId` | string | Yes |
| `eventName` | string | Yes |
| `eventType` | string | Yes |
| `eventIcon` | string | Yes |
| `navBarIcon` | string | Yes |
| `battleExpIcon` | string | Yes |
| `localizedShortName` | string | Yes |
| `eventTokenImage` | string | Yes |
| `startDate` | string | Yes |
| `progressEndDate` | string | Yes |
| `endDate` | string | Yes |
| `seasonPassSubType` | LolEventHubSeasonPassSubType | Yes |
| `localizedLogo` | string | Yes |
| `localizedEventSubtitle` | string | Yes |
| `localizedHelpUrl` | string | Yes |
| `objectiveCard` | LolEventHubObjectiveCard | Yes |
| `currentTokenBalance` | int32 | Yes |
| `lockedTokenCount` | int32 | Yes |
| `unclaimedRewardCount` | int32 | Yes |
| `timeOfLastUnclaimedReward` | int64 | Yes |
| `isPassPurchased` | boolean | Yes |
| `isGameModeEvent` | boolean | Yes |
| `queueId` | int32 | Yes |
| `eventPassBundles` | LolEventHubCatalogEntry[] | Yes |
| `tokenBundles` | LolEventHubCatalogEntry[] | Yes |

### LolEventHubItemOrderDTO

| Field | Type | Required |
|-------|------|----------|
| `inventoryType` | string | Yes |
| `itemId` | int32 | Yes |
| `quantity` | uint32 | Yes |
| `rpCost` | uint32 | Yes |

### LolEventHubNarrativeElement

| Field | Type | Required |
|-------|------|----------|
| `localizedNarrativeTitle` | string | Yes |
| `localizedNarrativeDescription` | string | Yes |
| `narrativeBackgroundImage` | string | Yes |
| `narrativeStartingTrackLevel` | uint16 | Yes |
| `narrativeVideo` | LolEventHubNarrativeVideo | Yes |

### LolEventHubNarrativeVideo

| Field | Type | Required |
|-------|------|----------|
| `localizedNarrativeVideoUrl` | string | Yes |
| `localizedPlayNarrativeButtonLabel` | string | Yes |
| `localizedVideoTitle` | string | Yes |
| `thumbnailImage` | string | Yes |
| `narrativeVideoIsLockedOnLevel` | boolean |  |
| `localizedNarrativeVideoDescription` | string | Yes |

### LolEventHubNavigationButtonUIData

| Field | Type | Required |
|-------|------|----------|
| `activeEventId` | string | Yes |
| `showPip` | boolean | Yes |
| `showGlow` | boolean | Yes |
| `iconPath` | string | Yes |
| `eventName` | string | Yes |
| `eventType` | string | Yes |

### LolEventHubObjectiveCard

| Field | Type | Required |
|-------|------|----------|
| `missionSeriesName` | string | Yes |
| `objectiveGroup` | string | Yes |
| `objectiveCategoryId` | string | Yes |

### LolEventHubObjectivesBanner

| Field | Type | Required |
|-------|------|----------|
| `eventName` | string | Yes |
| `promotionBannerImage` | string | Yes |
| `objectiveBannerImage` | string | Yes |
| `isPassPurchased` | boolean | Yes |
| `currentChapter` | LolEventHubChapter | Yes |
| `trackProgressNextReward` | LolEventHubTrackProgressNextReward | Yes |
| `trackProgress` | LolEventHubTrackProgressNextReward | Yes |
| `rewardTrackProgress` | LolEventHubRewardTrackProgress | Yes |

### LolEventHubOfferCategory

**Type**: enum string

**Values**: `Currencies`, `Tft`, `Loot`, `Borders`, `Skins`, `Chromas`, `Featured`


### LolEventHubOfferUIData

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `localizedTitle` | string | Yes |
| `localizedDescription` | string | Yes |
| `image` | string | Yes |
| `highlighted` | boolean | Yes |
| `offerState` | LolEventHubOfferStates | Yes |
| `price` | uint32 | Yes |
| `maxQuantity` | uint32 | Yes |
| `items` | LolEventHubItemUIData[] | Yes |

### LolEventHubProgressInfoUIData

| Field | Type | Required |
|-------|------|----------|
| `tokenImage` | string | Yes |
| `passPurchased` | boolean | Yes |
| `eventPassBundlesCatalogEntry` | LolEventHubCatalogEntry[] | Yes |

### LolEventHubProgressionPurchaseUIData

| Field | Type | Required |
|-------|------|----------|
| `offerId` | string | Yes |
| `pricePerLevel` | int64 | Yes |
| `rpBalance` | int64 | Yes |

### LolEventHubPurchaseOfferRequest

| Field | Type | Required |
|-------|------|----------|
| `offerId` | string | Yes |
| `purchaseQuantity` | uint32 | Yes |

### LolEventHubPurchaseOfferResponseV3

| Field | Type | Required |
|-------|------|----------|
| `legacy` | boolean | Yes |
| `orderDto` | LolEventHubCapOrdersOrderDto |  |

### LolEventHubPurchaseOrderResponseDTO

| Field | Type | Required |
|-------|------|----------|
| `rpBalance` | int64 | Yes |
| `ipBalance` | int64 | Yes |
| `transactions` | LolEventHubTransactionResponseDTO[] | Yes |

### LolEventHubRewardTrackError

| Field | Type | Required |
|-------|------|----------|
| `errorMessage` | string | Yes |
| `errorId` | string | Yes |

### LolEventHubRewardTrackItem

| Field | Type | Required |
|-------|------|----------|
| `state` | LolEventHubRewardTrackItemStates | Yes |
| `rewardOptions` | LolEventHubRewardTrackItemOption[] | Yes |
| `rewardTags` | LolEventHubRewardTrackItemTag[] | Yes |
| `progressRequired` | int64 | Yes |
| `threshold` | string | Yes |

### LolEventHubRewardTrackItemOption

| Field | Type | Required |
|-------|------|----------|
| `state` | LolEventHubRewardTrackItemStates | Yes |
| `thumbIconPath` | string | Yes |
| `splashImagePath` | string | Yes |
| `selected` | boolean | Yes |
| `overrideFooter` | string | Yes |
| `headerType` | LolEventHubRewardTrackItemHeaderType | Yes |
| `rewardName` | string | Yes |
| `rewardDescription` | string | Yes |
| `rewardItemType` | string | Yes |
| `rewardItemId` | string | Yes |
| `rewardFulfillmentSource` | string | Yes |
| `cardSize` | string | Yes |
| `rewardGroupId` | string | Yes |
| `celebrationType` | LolEventHubCelebrationType | Yes |
| `rewardInventoryTypes` | string[] | Yes |

### LolEventHubRewardTrackItemStates

**Type**: enum string

**Values**: `Selected`, `Unselected`, `Unlocked`, `Locked`


### LolEventHubRewardTrackItemTag

**Type**: enum string

**Values**: `Multiple`, `Choice`, `Instant`, `Free`, `Rare`


### LolEventHubRewardTrackProgress

| Field | Type | Required |
|-------|------|----------|
| `level` | int16 | Yes |
| `totalLevels` | int16 | Yes |
| `levelProgress` | uint16 | Yes |
| `futureLevelProgress` | uint16 | Yes |
| `passProgress` | int64 | Yes |
| `currentLevelXP` | int64 | Yes |
| `totalLevelXP` | int64 | Yes |
| `iteration` | uint32 | Yes |

### LolEventHubRewardTrackXP

| Field | Type | Required |
|-------|------|----------|
| `currentLevel` | int64 | Yes |
| `currentLevelXP` | int64 | Yes |
| `totalLevelXP` | int64 | Yes |
| `isBonusPhase` | boolean | Yes |
| `iteration` | uint32 | Yes |

### LolEventHubSeasonPassSubType

**Type**: enum string

**Values**: `Mayhem`, `Default`


### LolEventHubTokenShopUIData

| Field | Type | Required |
|-------|------|----------|
| `tokenName` | string | Yes |
| `tokenImage` | string | Yes |
| `tokenUuid` | string | Yes |
| `offersVersion` | uint32 | Yes |
| `tokenBundlesCatalogEntry` | LolEventHubCatalogEntry[] | Yes |

### LolEventHubTokenUpsell

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `internalName` | string | Yes |
| `title` | string | Yes |
| `buttonText` | string | Yes |
| `tooltipTitle` | string | Yes |
| `tooltipDescription` | string | Yes |
| `purchaseUrl` | string | Yes |
| `tooltipBackgroundUrl` | string | Yes |
| `backgroundUrl` | string | Yes |
| `currencyUrl` | string | Yes |
| `premiumCurrencyName` | string | Yes |
| `dependentInventoryType` | string | Yes |
| `dependentInventoryId` | int32 | Yes |
| `currentlyLocked` | LolEventHubTokenUpsellLockedType | Yes |
| `lockedCount` | int32 | Yes |
| `startDate` | string | Yes |
| `endDate` | string | Yes |

### LolEventHubTokenUpsellLockedType

**Type**: enum string

**Values**: `UNLOCKED`, `LOCKED`, `UNASSIGNED`


### LolEventHubTrackProgressNextReward

| Field | Type | Required |
|-------|------|----------|
| `currentXP` | int64 | Yes |
| `nextLevelXP` | int64 | Yes |
| `currentLevel` | int64 | Yes |
| `nextReward` | LolEventHubNextRewardUIData | Yes |

### LolEventHubTransactionResponseDTO

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `inventoryType` | string | Yes |
| `itemId` | int32 | Yes |

### LolEventHubUnclaimedRewardsUIData

| Field | Type | Required |
|-------|------|----------|
| `rewardsCount` | int32 | Yes |
| `lockedTokensCount` | int32 | Yes |
| `timeOfLastUnclaimedReward` | int64 | Yes |

---

## LolGame


### LolGameClientChatBuddy

| Field | Type | Required |
|-------|------|----------|
| `gameName` | string | Yes |
| `tagLine` | string | Yes |
| `puuid` | string | Yes |
| `summonerId` | uint64 | Yes |

### LolGameClientChatMessageToPlayer

| Field | Type | Required |
|-------|------|----------|
| `gameName` | string | Yes |
| `tagLine` | string | Yes |
| `body` | string | Yes |

### LolGameQueuesQueue

| Field | Type | Required |
|-------|------|----------|
| `id` | int32 | Yes |
| `mapId` | int32 | Yes |
| `name` | string | Yes |
| `shortName` | string | Yes |
| `description` | string | Yes |
| `detailedDescription` | string | Yes |
| `type` | string | Yes |
| `gameMode` | string | Yes |
| `assetMutator` | string | Yes |
| `maxTierForPremadeSize2` | string | Yes |
| `maxDivisionForPremadeSize2` | string | Yes |
| `category` | LolGameQueuesQueueGameCategory | Yes |
| `gameTypeConfig` | LolGameQueuesQueueGameTypeConfig | Yes |
| `numPlayersPerTeam` | int32 | Yes |
| `minimumParticipantListSize` | int32 | Yes |
| `maximumParticipantListSize` | int32 | Yes |
| `minLevel` | uint32 | Yes |
| `isRanked` | boolean | Yes |
| `areFreeChampionsAllowed` | boolean | Yes |
| `isTeamBuilderManaged` | boolean | Yes |
| `queueAvailability` | LolGameQueuesQueueAvailability | Yes |
| `isEnabled` | boolean | Yes |
| `isVisible` | boolean | Yes |
| `queueRewards` | LolGameQueuesQueueReward | Yes |
| `spectatorEnabled` | boolean | Yes |
| `championsRequiredToPlay` | uint32 | Yes |
| `allowablePremadeSizes` | integer[] | Yes |
| `showPositionSelector` | boolean | Yes |
| `showQuickPlaySlotSelection` | boolean | Yes |
| `lastToggledOffTime` | uint64 | Yes |
| `lastToggledOnTime` | uint64 | Yes |
| `removalFromGameAllowed` | boolean | Yes |
| `removalFromGameDelayMinutes` | int32 | Yes |
| `hidePlayerPosition` | boolean | Yes |
| `gameSelectModeGroup` | string | Yes |
| `gameSelectCategory` | string | Yes |
| `gameSelectPriority` | uint8 | Yes |
| `isLimitedTimeQueue` | boolean | Yes |
| `isSkillTreeQueue` | boolean | Yes |
| `isBotHonoringAllowed` | boolean | Yes |
| `isCustom` | boolean | Yes |
| `numberOfTeamsInLobby` | uint32 | Yes |
| `maxLobbySpectatorCount` | uint32 | Yes |

### LolGameQueuesQueueAvailability

**Type**: enum string

**Values**: `DoesntMeetRequirements`, `PlatformDisabled`, `Available`


### LolGameQueuesQueueCustomGame

| Field | Type | Required |
|-------|------|----------|
| `subcategories` | LolGameQueuesQueueCustomGameSubcategory[] | Yes |
| `queueAvailability` | LolGameQueuesQueueAvailability | Yes |
| `spectatorPolicies` | LolGameQueuesQueueCustomGameSpectatorPolicy[] | Yes |
| `spectatorSlotLimit` | uint32 | Yes |
| `gameServerRegions` | string[] |  |

### LolGameQueuesQueueCustomGameSpectatorPolicy

**Type**: enum string

**Values**: `AllAllowed`, `FriendsAllowed`, `LobbyAllowed`, `NotAllowed`


### LolGameQueuesQueueCustomGameSubcategory

| Field | Type | Required |
|-------|------|----------|
| `mapId` | int32 | Yes |
| `gameMode` | string | Yes |
| `mutators` | LolGameQueuesQueueGameTypeConfig[] | Yes |
| `numPlayersPerTeam` | int32 | Yes |
| `minimumParticipantListSize` | int32 | Yes |
| `maximumParticipantListSize` | int32 | Yes |
| `maxPlayerCount` | int32 | Yes |
| `minLevel` | uint32 | Yes |
| `queueAvailability` | LolGameQueuesQueueAvailability | Yes |
| `customSpectatorPolicies` | LolGameQueuesQueueCustomGameSpectatorPolicy[] | Yes |

### LolGameQueuesQueueGameCategory

**Type**: enum string

**Values**: `Alpha`, `VersusAi`, `PvP`, `Custom`, `None`


### LolGameQueuesQueueGameTypeConfig

| Field | Type | Required |
|-------|------|----------|
| `id` | int64 | Yes |
| `name` | string | Yes |
| `maxAllowableBans` | int32 | Yes |
| `allowTrades` | boolean | Yes |
| `exclusivePick` | boolean | Yes |
| `duplicatePick` | boolean | Yes |
| `teamChampionPool` | boolean | Yes |
| `crossTeamChampionPool` | boolean | Yes |
| `advancedLearningQuests` | boolean | Yes |
| `battleBoost` | boolean | Yes |
| `deathMatch` | boolean | Yes |
| `doNotRemove` | boolean | Yes |
| `learningQuests` | boolean | Yes |
| `onboardCoopBeginner` | boolean | Yes |
| `reroll` | boolean | Yes |
| `mainPickTimerDuration` | int32 | Yes |
| `postPickTimerDuration` | int32 | Yes |
| `banTimerDuration` | int32 | Yes |
| `pickMode` | string | Yes |
| `banMode` | string | Yes |
| `gameModeOverride` | string |  |
| `numPlayersPerTeamOverride` | int32 |  |

### LolGameQueuesQueueReward

| Field | Type | Required |
|-------|------|----------|
| `isIpEnabled` | boolean | Yes |
| `isXpEnabled` | boolean | Yes |
| `isChampionPointsEnabled` | boolean | Yes |
| `partySizeIpRewards` | integer[] | Yes |

### LolGameQueuesQueueTranslation

| Field | Type | Required |
|-------|------|----------|
| `id` | int32 | Yes |
| `name` | string | Yes |
| `shortName` | string | Yes |
| `description` | string | Yes |
| `detailedDescription` | string | Yes |
| `gameSelectModeGroup` | string | Yes |
| `gameSelectCategory` | string | Yes |
| `gameSelectPriority` | uint8 | Yes |
| `isLimitedTimeQueue` | boolean | Yes |
| `isSkillTreeQueue` | boolean | Yes |
| `isBotHonoringAllowed` | boolean | Yes |
| `hidePlayerPosition` | boolean | Yes |
| `viableChampionRoster` | integer[] | Yes |

---

## LolGameflow


### LolGameflowGameflowAvailability

| Field | Type | Required |
|-------|------|----------|
| `isAvailable` | boolean | Yes |
| `state` | LolGameflowGameflowAvailabilityState | Yes |

### LolGameflowGameflowAvailabilityState

**Type**: enum string

**Values**: `EligibilityInfoMissing`, `Configuration`, `InGameFlow`, `PlayerBanned`, `Patching`, `Initializing`, `Available`


### LolGameflowGameflowGameClient

| Field | Type | Required |
|-------|------|----------|
| `serverIp` | string | Yes |
| `serverPort` | uint16 | Yes |
| `observerServerIp` | string | Yes |
| `observerServerPort` | uint16 | Yes |
| `running` | boolean | Yes |
| `visible` | boolean | Yes |

### LolGameflowGameflowGameData

| Field | Type | Required |
|-------|------|----------|
| `gameId` | uint64 | Yes |
| `queue` | LolGameflowQueue | Yes |
| `isCustomGame` | boolean | Yes |
| `gameName` | string | Yes |
| `password` | string | Yes |
| `teamOne` | object[] | Yes |
| `teamTwo` | object[] | Yes |
| `playerChampionSelections` | object[] | Yes |
| `spectatorsAllowed` | boolean | Yes |
| `spectatorKey` | string | Yes |

### LolGameflowGameflowGameDodge

| Field | Type | Required |
|-------|------|----------|
| `state` | LolGameflowGameflowGameDodgeState | Yes |
| `dodgeIds` | integer[] | Yes |
| `phase` | LolGameflowGameflowPhase | Yes |

### LolGameflowGameflowGameDodgeState

**Type**: enum string

**Values**: `TournamentDodged`, `StrangerDodged`, `PartyDodged`, `Invalid`


### LolGameflowGameflowGameMap

| Field | Type | Required |
|-------|------|----------|
| `id` | int64 | Yes |
| `name` | string | Yes |
| `mapStringId` | string | Yes |
| `gameMode` | string | Yes |
| `gameModeName` | string | Yes |
| `gameModeShortName` | string | Yes |
| `gameMutator` | string | Yes |
| `isRGM` | boolean | Yes |
| `description` | string | Yes |
| `platformId` | string | Yes |
| `platformName` | string | Yes |
| `assets` | object (map) | Yes |
| `categorizedContentBundles` | object (map) | Yes |
| `properties` | object (map) | Yes |
| `perPositionRequiredSummonerSpells` | object (map) | Yes |
| `perPositionDisallowedSummonerSpells` | object (map) | Yes |

### LolGameflowGameflowPhase

**Type**: enum string

**Values**: `TerminatedInError`, `EndOfGame`, `PreEndOfGame`, `WaitingForStats`, `Reconnect`, `InProgress`, `FailedToLaunch`, `GameStart`, `ChampSelect`, `ReadyCheck`, `CheckedIntoTournament`, `Matchmaking`, `Lobby`, `None`


### LolGameflowGameflowSession

| Field | Type | Required |
|-------|------|----------|
| `phase` | LolGameflowGameflowPhase | Yes |
| `gameData` | LolGameflowGameflowGameData | Yes |
| `gameClient` | LolGameflowGameflowGameClient | Yes |
| `map` | LolGameflowGameflowGameMap | Yes |
| `gameDodge` | LolGameflowGameflowGameDodge | Yes |

### LolGameflowGameflowWatchPhase

**Type**: enum string

**Values**: `WatchFailedToLaunch`, `WatchInProgress`, `WatchStarted`, `None`


### LolGameflowLobbyStatus

| Field | Type | Required |
|-------|------|----------|
| `queueId` | int32 | Yes |
| `isCustom` | boolean | Yes |
| `isPracticeTool` | boolean | Yes |
| `isLeader` | boolean | Yes |
| `isSpectator` | boolean | Yes |
| `allowedPlayAgain` | boolean | Yes |
| `isNoSpectateDelay` | boolean | Yes |
| `memberSummonerIds` | integer[] | Yes |
| `invitedSummonerIds` | integer[] | Yes |
| `lobbyId` | string |  |
| `customSpectatorPolicy` | LolGameflowQueueCustomGameSpectatorPolicy | Yes |

### LolGameflowPlayerStatus

| Field | Type | Required |
|-------|------|----------|
| `currentLobbyStatus` | LolGameflowLobbyStatus |  |
| `lastQueuedLobbyStatus` | LolGameflowLobbyStatus |  |
| `canInviteOthersAtEog` | boolean | Yes |

### LolGameflowQueue

| Field | Type | Required |
|-------|------|----------|
| `id` | int32 | Yes |
| `mapId` | int32 | Yes |
| `name` | string | Yes |
| `shortName` | string | Yes |
| `description` | string | Yes |
| `detailedDescription` | string | Yes |
| `type` | string | Yes |
| `gameMode` | string | Yes |
| `assetMutator` | string | Yes |
| `category` | LolGameflowQueueGameCategory | Yes |
| `gameTypeConfig` | LolGameflowQueueGameTypeConfig | Yes |
| `numPlayersPerTeam` | int32 | Yes |
| `minimumParticipantListSize` | int32 | Yes |
| `maximumParticipantListSize` | int32 | Yes |
| `minLevel` | uint32 | Yes |
| `isRanked` | boolean | Yes |
| `areFreeChampionsAllowed` | boolean | Yes |
| `isTeamBuilderManaged` | boolean | Yes |
| `queueAvailability` | LolGameflowQueueAvailability | Yes |
| `queueRewards` | LolGameflowQueueReward | Yes |
| `spectatorEnabled` | boolean | Yes |
| `championsRequiredToPlay` | uint32 | Yes |
| `allowablePremadeSizes` | integer[] | Yes |
| `showPositionSelector` | boolean | Yes |
| `lastToggledOffTime` | uint64 | Yes |
| `lastToggledOnTime` | uint64 | Yes |
| `removalFromGameAllowed` | boolean | Yes |
| `removalFromGameDelayMinutes` | int32 | Yes |
| `isCustom` | boolean | Yes |
| `isBotHonoringAllowed` | boolean | Yes |

### LolGameflowQueueAvailability

**Type**: enum string

**Values**: `DoesntMeetRequirements`, `PlatformDisabled`, `Available`


### LolGameflowQueueGameCategory

**Type**: enum string

**Values**: `Alpha`, `VersusAi`, `PvP`, `Custom`, `None`


### LolGameflowQueueGameTypeConfig

| Field | Type | Required |
|-------|------|----------|
| `id` | int64 | Yes |
| `name` | string | Yes |
| `maxAllowableBans` | int32 | Yes |
| `allowTrades` | boolean | Yes |
| `exclusivePick` | boolean | Yes |
| `duplicatePick` | boolean | Yes |
| `teamChampionPool` | boolean | Yes |
| `crossTeamChampionPool` | boolean | Yes |
| `advancedLearningQuests` | boolean | Yes |
| `battleBoost` | boolean | Yes |
| `deathMatch` | boolean | Yes |
| `doNotRemove` | boolean | Yes |
| `learningQuests` | boolean | Yes |
| `onboardCoopBeginner` | boolean | Yes |
| `reroll` | boolean | Yes |
| `mainPickTimerDuration` | int32 | Yes |
| `postPickTimerDuration` | int32 | Yes |
| `banTimerDuration` | int32 | Yes |
| `pickMode` | string | Yes |
| `banMode` | string | Yes |

### LolGameflowQueueReward

| Field | Type | Required |
|-------|------|----------|
| `isIpEnabled` | boolean | Yes |
| `isXpEnabled` | boolean | Yes |
| `isChampionPointsEnabled` | boolean | Yes |
| `partySizeIpRewards` | integer[] | Yes |

### LolGameflowRegistrationStatus

| Field | Type | Required |
|-------|------|----------|
| `complete` | boolean | Yes |
| `errorCodes` | string[] | Yes |

### LolGameflowSpectateGameInfoResource

| Field | Type | Required |
|-------|------|----------|
| `dropInSpectateGameId` | string | Yes |
| `gameQueueType` | string | Yes |
| `allowObserveMode` | string | Yes |
| `puuid` | string | Yes |
| `spectatorKey` | string | Yes |

---

## LolHeartbeat


### LolHeartbeatLcdsConnection

| Field | Type | Required |
|-------|------|----------|
| `stableConnection` | boolean | Yes |

---

## LolHighlights


### LolHighlightsHighlight

| Field | Type | Required |
|-------|------|----------|
| `id` | uint64 | Yes |
| `name` | string | Yes |
| `filepath` | string | Yes |
| `url` | string | Yes |
| `mtimeMsUtc` | uint64 | Yes |
| `mtimeIso8601` | string | Yes |
| `fileSizeBytes` | uint64 | Yes |

### LolHighlightsHighlightsConfig

| Field | Type | Required |
|-------|------|----------|
| `isHighlightsEnabled` | boolean | Yes |
| `invalidHighlightNameCharacters` | string | Yes |

---

## LolHonor


### LolHonorV2ApiHonorPlayerServerRequest

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `puuid` | string | Yes |
| `honorType` | string | Yes |
| `gameId` | uint64 | Yes |

### LolHonorV2ApiHonorPlayerServerRequestV3

| Field | Type | Required |
|-------|------|----------|
| `recipientPuuid` | string | Yes |
| `honorType` | string | Yes |

### LolHonorV2Ballot

| Field | Type | Required |
|-------|------|----------|
| `eligibleAllies` | LolHonorV2EligiblePlayer[] | Yes |
| `eligibleOpponents` | LolHonorV2EligiblePlayer[] | Yes |
| `votePool` | LolHonorV2VotePool | Yes |
| `gameId` | uint64 | Yes |
| `honoredPlayers` | LolHonorV2ApiHonorPlayerServerRequestV3[] | Yes |

### LolHonorV2DynamicHonorMessage

| Field | Type | Required |
|-------|------|----------|
| `messageId` | string | Yes |
| `value` | int32 | Yes |

### LolHonorV2EligiblePlayer

| Field | Type | Required |
|-------|------|----------|
| `puuid` | string | Yes |
| `summonerId` | uint64 | Yes |
| `summonerName` | string | Yes |
| `championName` | string | Yes |
| `championId` | int32 | Yes |
| `skinSplashPath` | string | Yes |
| `role` | string | Yes |
| `botPlayer` | boolean | Yes |

### LolHonorV2Honor

| Field | Type | Required |
|-------|------|----------|
| `honorCategory` | string | Yes |
| `voterRelationship` | string | Yes |
| `senderPuuid` | string | Yes |

### LolHonorV2HonorConfig

| Field | Type | Required |
|-------|------|----------|
| `Enabled` | boolean | Yes |
| `SecondsToVote` | int32 | Yes |
| `HonorVisibilityEnabled` | boolean | Yes |
| `HonorSuggestionsEnabled` | boolean | Yes |
| `honorEndpointsV2Enabled` | boolean | Yes |
| `ceremonyV3Enabled` | boolean | Yes |
| `useHonorInPostgamePlugin` | boolean | Yes |

### LolHonorV2HonorInteraction

| Field | Type | Required |
|-------|------|----------|
| `puuid` | string | Yes |
| `displayName` | string | Yes |
| `gameId` | uint64 | Yes |
| `summonerId` | uint64 | Yes |

### LolHonorV2Mail

| Field | Type | Required |
|-------|------|----------|
| `mailId` | string | Yes |
| `message` | string | Yes |
| `state` | string | Yes |
| `createdAt` | uint64 | Yes |

### LolHonorV2MutualHonor

| Field | Type | Required |
|-------|------|----------|
| `gameId` | uint64 | Yes |
| `summoners` | LolHonorV2MutualHonorPlayer[] | Yes |

### LolHonorV2MutualHonorPlayer

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `championId` | int32 | Yes |
| `skinId` | int32 | Yes |

### LolHonorV2ProfileInfo

| Field | Type | Required |
|-------|------|----------|
| `honorLevel` | int32 | Yes |
| `checkpoint` | int32 | Yes |
| `rewardsLocked` | boolean | Yes |
| `redemptions` | LolHonorV2Redemption[] | Yes |

### LolHonorV2Redemption

| Field | Type | Required |
|-------|------|----------|
| `required` | int32 | Yes |
| `remaining` | int32 | Yes |
| `eventType` | string | Yes |

### LolHonorV2Reward

| Field | Type | Required |
|-------|------|----------|
| `rewardType` | string | Yes |
| `quantity` | int32 | Yes |

### LolHonorV2VendedHonorChange

| Field | Type | Required |
|-------|------|----------|
| `actionType` | string | Yes |
| `previousState` | LolHonorV2VendedHonorState | Yes |
| `currentState` | LolHonorV2VendedHonorState | Yes |
| `reward` | LolHonorV2Reward | Yes |
| `dynamicHonorMessage` | LolHonorV2DynamicHonorMessage | Yes |

### LolHonorV2VendedHonorState

| Field | Type | Required |
|-------|------|----------|
| `level` | int32 | Yes |
| `checkpoint` | int32 | Yes |
| `rewardsLocked` | boolean | Yes |

### LolHonorV2VendedReward

| Field | Type | Required |
|-------|------|----------|
| `rewardType` | string | Yes |
| `quantity` | int32 | Yes |
| `dynamicHonorMessage` | LolHonorV2DynamicHonorMessage | Yes |

### LolHonorV2VoteCompletion

| Field | Type | Required |
|-------|------|----------|
| `gameId` | uint64 | Yes |
| `fullTeamVote` | boolean | Yes |

### LolHonorV2VotePool

| Field | Type | Required |
|-------|------|----------|
| `votes` | int32 | Yes |
| `fromGamePlayed` | int32 | Yes |
| `fromHighHonor` | int32 | Yes |
| `fromRecentHonors` | int32 | Yes |
| `fromRollover` | int32 | Yes |

---

## LolHovercard


### LolHovercardHovercardUserInfo

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `puuid` | string | Yes |
| `summonerId` | uint64 | Yes |
| `name` | string | Yes |
| `accountId` | uint64 | Yes |
| `icon` | int32 | Yes |
| `gameName` | string | Yes |
| `gameTag` | string | Yes |
| `availability` | string | Yes |
| `note` | string | Yes |
| `masteryScore` | uint64 | Yes |
| `legendaryMasteryScore` | uint64 | Yes |
| `patchline` | string | Yes |
| `platformId` | string | Yes |
| `product` | string | Yes |
| `productName` | string | Yes |
| `statusMessage` | string | Yes |
| `summonerIcon` | int32 | Yes |
| `summonerLevel` | uint32 | Yes |
| `remoteProduct` | boolean | Yes |
| `remotePlatform` | boolean | Yes |
| `remoteProductIconUrl` | string | Yes |
| `remoteProductBackdropUrl` | string | Yes |
| `partySummoners` | string[] | Yes |
| `lol` | object (map) | Yes |
| `discordOnlineStatus` | string |  |
| `discordId` | string |  |

---

## LolInventory


### LolInventoryInventoryItemWithPayload

| Field | Type | Required |
|-------|------|----------|
| `uuid` | string | Yes |
| `itemId` | int32 | Yes |
| `inventoryType` | string | Yes |
| `purchaseDate` | string | Yes |
| `quantity` | uint64 | Yes |
| `ownershipType` | LolInventoryItemOwnershipType | Yes |
| `usedInGameDate` | string | Yes |
| `expirationDate` | string | Yes |
| `f2p` | boolean | Yes |
| `rental` | boolean | Yes |
| `loyalty` | boolean | Yes |
| `loyaltySources` | string[] | Yes |
| `owned` | boolean | Yes |
| `wins` | uint64 | Yes |
| `payload` | object (map) | Yes |

### LolInventoryInventoryNotification

| Field | Type | Required |
|-------|------|----------|
| `id` | int64 | Yes |
| `itemId` | int32 | Yes |
| `inventoryType` | string | Yes |
| `type` | string | Yes |
| `acknowledged` | boolean | Yes |

### LolInventoryItemOwnershipType

**Type**: enum string

**Values**: `F2P`, `LOYALTY`, `RENTED`, `OWNED`


### LolInventoryXboxSubscriptionStatus

| Field | Type | Required |
|-------|------|----------|
| `active` | string | Yes |
| `subscriptionId` | string | Yes |

---

## LolItem


### LolItemSetsItemSet

| Field | Type | Required |
|-------|------|----------|
| `uid` | string | Yes |
| `title` | string | Yes |
| `mode` | string | Yes |
| `map` | string | Yes |
| `type` | string | Yes |
| `sortrank` | int32 | Yes |
| `startedFrom` | string | Yes |
| `associatedChampions` | integer[] | Yes |
| `associatedMaps` | integer[] | Yes |
| `blocks` | LolItemSetsItemSetBlock[] | Yes |
| `preferredItemSlots` | LolItemSetsPreferredItemSlot[] | Yes |

### LolItemSetsItemSetBlock

| Field | Type | Required |
|-------|------|----------|
| `type` | string | Yes |
| `hideIfSummonerSpell` | string | Yes |
| `showIfSummonerSpell` | string | Yes |
| `items` | LolItemSetsItemSetItem[] | Yes |

### LolItemSetsItemSets

| Field | Type | Required |
|-------|------|----------|
| `timestamp` | uint64 | Yes |
| `accountId` | uint64 | Yes |
| `itemSets` | LolItemSetsItemSet[] | Yes |

### LolItemSetsNamecheckResponse

| Field | Type | Required |
|-------|------|----------|
| `errors` | string[] | Yes |

### LolItemSetsPreferredItemSlot

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `preferredItemSlot` | int16 | Yes |

### LolItemSetsValidateItemSetNameInput

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |

### LolItemSetsValidateItemSetNameResponse

| Field | Type | Required |
|-------|------|----------|
| `success` | boolean | Yes |
| `nameCheckResponse` | LolItemSetsNamecheckResponse | Yes |

---

## LolKickout


### LolKickoutKickoutMessage

| Field | Type | Required |
|-------|------|----------|
| `message` | string | Yes |

---

## LolKr


### LolKrShutdownLawAllQueueShutdownStatus

| Field | Type | Required |
|-------|------|----------|
| `isAllQueuesDisabled` | boolean | Yes |

### LolKrShutdownLawQueueShutdownStatus

| Field | Type | Required |
|-------|------|----------|
| `isDisabled` | boolean | Yes |

### LolKrShutdownLawRatingScreenInfo

| Field | Type | Required |
|-------|------|----------|
| `shown` | boolean | Yes |

### LolKrShutdownLawShutdownLawNotification

| Field | Type | Required |
|-------|------|----------|
| `type` | LolKrShutdownLawShutdownLawStatus | Yes |

### LolKrShutdownLawShutdownLawStatus

**Type**: enum string

**Values**: `CUT_OFF`, `WARNING`, `NONE`


---

## LolL10n


### LolL10nRegionLocale

| Field | Type | Required |
|-------|------|----------|
| `region` | string | Yes |
| `locale` | string | Yes |
| `webRegion` | string | Yes |
| `webLanguage` | string | Yes |

---

## LolLeaderboard


### LolLeaderboardLeaderboardPageResponse

| Field | Type | Required |
|-------|------|----------|
| `grouping` | string | Yes |
| `season` | string | Yes |
| `region` | string | Yes |
| `name` | string | Yes |
| `startRank` | uint32 | Yes |
| `endRank` | uint32 | Yes |
| `size` | uint32 | Yes |
| `rankings` | LolLeaderboardPagedLeaderboardEntry[] | Yes |

### LolLeaderboardLeaderboardPlayerRanking

| Field | Type | Required |
|-------|------|----------|
| `entityId` | string | Yes |
| `ranking` | uint32 | Yes |

### LolLeaderboardPagedLeaderboardEntry

| Field | Type | Required |
|-------|------|----------|
| `entityId` | string | Yes |
| `entityName` | string | Yes |
| `ranking` | uint32 | Yes |
| `score` | uint32 | Yes |
| `anonymous` | boolean | Yes |

---

## LolLeaver


### LolLeaverBusterLeaverBusterNotificationResource

| Field | Type | Required |
|-------|------|----------|
| `id` | uint32 | Yes |
| `msgId` | string | Yes |
| `accountId` | uint64 | Yes |
| `type` | LolLeaverBusterLeaverBusterNotificationType | Yes |
| `punishedGamesRemaining` | int32 | Yes |
| `queueLockoutTimerExpiryUtcMillisDiff` | uint64 | Yes |
| `fromRms` | boolean | Yes |

### LolLeaverBusterLeaverBusterNotificationType

**Type**: enum string

**Values**: `WinBasedRankedRestrictionsEnabled`, `RankedRestrictedGames`, `OnLockoutWarning`, `PreLockoutWarning`, `Reforming`, `PunishedGamesRemaining`, `PunishmentIncurred`, `TaintedWarning`, `Invalid`


### LolLeaverBusterRankedRestrictionInfo

| Field | Type | Required |
|-------|------|----------|
| `punishedGamesRemaining` | int32 | Yes |
| `needsAck` | boolean | Yes |

---

## LolLoadouts


### LolLoadoutsScopedLoadout

| Field | Type | Required |
|-------|------|----------|
| `scope` | string | Yes |
| `itemId` | uint32 |  |
| `name` | string | Yes |
| `loadout` | object (map) | Yes |
| `refreshTime` | string | Yes |
| `id` | string | Yes |

### LolLoadoutsUpdateLoadoutDTO

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `name` | string | Yes |
| `loadout` | object (map) | Yes |

---

## LolLobby


### LolLobbyCustomJoinOptionsDto

| Field | Type | Required |
|-------|------|----------|
| `lobbyPassword` | string | Yes |
| `team` | string |  |

### LolLobbyEligibility

| Field | Type | Required |
|-------|------|----------|
| `queueId` | int32 | Yes |
| `eligible` | boolean | Yes |
| `restrictions` | LolLobbyEligibilityRestriction[] | Yes |

### LolLobbyEligibilityRestriction

| Field | Type | Required |
|-------|------|----------|
| `restrictionCode` | LolLobbyEligibilityRestrictionCode | Yes |
| `restrictionArgs` | object (map) | Yes |
| `expiredTimestamp` | uint64 | Yes |
| `summonerIds` | integer[] | Yes |
| `summonerIdsString` | string | Yes |
| `puuids` | string[] | Yes |
| `puuidsString` | string | Yes |

### LolLobbyGameModeDto

| Field | Type | Required |
|-------|------|----------|
| `gameType` | string | Yes |
| `queueId` | int32 |  |
| `maxPartySize` | int32 | Yes |
| `maxTeamSize` | int32 | Yes |
| `allowSpectators` | string |  |
| `mapId` | int32 |  |
| `gameTypeConfigId` | int64 |  |
| `gameCustomization` | object (map) |  |
| `lobbyName` | string |  |
| `lobbyPassword` | string |  |
| `customGameMode` | string |  |

### LolLobbyInvitationType

**Type**: enum string

**Values**: `party`, `lobby`, `invalid`


### LolLobbyInviteContext

**Type**: enum string

**Values**: `RIOT`, `DISCORD`, `LOL`


### LolLobbyLobbyBotChampion

| Field | Type | Required |
|-------|------|----------|
| `active` | boolean | Yes |
| `id` | int32 | Yes |
| `name` | string | Yes |
| `botDifficulties` | LolLobbyLobbyBotDifficulty[] | Yes |

### LolLobbyLobbyBotDifficulty

**Type**: enum string

**Values**: `RSWARMINTRO`, `RSINTERMEDIATE`, `RSBEGINNER`, `RSINTRO`, `INTRO`, `TUTORIAL`, `UBER`, `HARD`, `MEDIUM`, `EASY`, `NONE`


### LolLobbyLobbyBotParams

| Field | Type | Required |
|-------|------|----------|
| `championId` | int32 | Yes |
| `botDifficulty` | LolLobbyLobbyBotDifficulty | Yes |
| `teamId` | string | Yes |
| `position` | string | Yes |
| `botUuid` | string | Yes |

### LolLobbyLobbyChangeGameDto

| Field | Type | Required |
|-------|------|----------|
| `queueId` | int32 | Yes |
| `customGameLobby` | LolLobbyLobbyCustomGameLobby |  |
| `gameCustomization` | object (map) |  |

### LolLobbyLobbyCustomChampSelectStartResponse

| Field | Type | Required |
|-------|------|----------|
| `success` | boolean | Yes |
| `failedPlayers` | LolLobbyLobbyCustomFailedPlayer[] | Yes |

### LolLobbyLobbyCustomFailedPlayer

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `summonerName` | string | Yes |
| `reason` | string | Yes |

### LolLobbyLobbyCustomGame

| Field | Type | Required |
|-------|------|----------|
| `id` | uint64 | Yes |
| `gameType` | string | Yes |
| `mapId` | int32 | Yes |
| `ownerDisplayName` | string | Yes |
| `spectatorPolicy` | string | Yes |
| `filledSpectatorSlots` | int32 | Yes |
| `maxSpectatorSlots` | uint64 | Yes |
| `filledPlayerSlots` | int32 | Yes |
| `maxPlayerSlots` | int32 | Yes |
| `lobbyName` | string | Yes |
| `hasPassword` | boolean | Yes |
| `passbackUrl` | string | Yes |
| `partyId` | string | Yes |

### LolLobbyLobbyCustomGameLobby

| Field | Type | Required |
|-------|------|----------|
| `lobbyName` | string | Yes |
| `lobbyPassword` | string | Yes |
| `configuration` | LolLobbyLobbyCustomGameConfiguration | Yes |
| `teamOne` | LolLobbyLobbyMember[] | Yes |
| `teamTwo` | LolLobbyLobbyMember[] | Yes |
| `spectators` | LolLobbyLobbyMember[] | Yes |
| `practiceGameRewardsDisabledReasons` | string[] | Yes |
| `gameId` | uint64 | Yes |

### LolLobbyLobbyCustomJoinParameters

| Field | Type | Required |
|-------|------|----------|
| `password` | string |  |
| `asSpectator` | boolean |  |

### LolLobbyLobbyDto

| Field | Type | Required |
|-------|------|----------|
| `partyId` | string | Yes |
| `partyType` | string | Yes |
| `members` | LolLobbyLobbyParticipantDto[] | Yes |
| `localMember` | LolLobbyLobbyParticipantDto | Yes |
| `invitations` | LolLobbyLobbyInvitationDto[] | Yes |
| `canStartActivity` | boolean | Yes |
| `restrictions` | LolLobbyEligibilityRestriction[] |  |
| `warnings` | LolLobbyEligibilityRestriction[] |  |
| `gameConfig` | LolLobbyLobbyGameConfigDto | Yes |
| `multiUserChatId` | string | Yes |
| `multiUserChatPassword` | string | Yes |
| `mucJwtDto` | LolLobbyMucJwtDto | Yes |
| `scarcePositions` | string[] | Yes |
| `popularChampions` | integer[] | Yes |

### LolLobbyLobbyGameConfigDto

| Field | Type | Required |
|-------|------|----------|
| `gameMode` | string | Yes |
| `mapId` | int32 | Yes |
| `queueId` | int32 | Yes |
| `pickType` | string | Yes |
| `maxTeamSize` | int32 | Yes |
| `maxLobbySize` | int32 | Yes |
| `maxHumanPlayers` | int32 | Yes |
| `numberOfTeamsInLobby` | int32 | Yes |
| `maxLobbySpectatorCount` | int32 | Yes |
| `numPlayersPerTeam` | int32 | Yes |
| `allowablePremadeSizes` | integer[] | Yes |
| `premadeSizeAllowed` | boolean | Yes |
| `isTeamBuilderManaged` | boolean | Yes |
| `isCustom` | boolean | Yes |
| `showPositionSelector` | boolean | Yes |
| `showQuickPlaySlotSelection` | boolean | Yes |
| `isLobbyFull` | boolean | Yes |
| `shouldForceScarcePositionSelection` | boolean | Yes |
| `customLobbyName` | string | Yes |
| `customMutatorName` | string | Yes |
| `customTeam100` | LolLobbyLobbyParticipantDto[] | Yes |
| `customTeam200` | LolLobbyLobbyParticipantDto[] | Yes |
| `customSpectators` | LolLobbyLobbyParticipantDto[] | Yes |
| `customSpectatorPolicy` | LolLobbyQueueCustomGameSpectatorPolicy | Yes |
| `customRewardsDisabledReasons` | string[] | Yes |

### LolLobbyLobbyInvitation

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `fromSummonerId` | uint64 | Yes |
| `toSummonerId` | uint64 | Yes |
| `state` | LolLobbyLobbyInvitationState | Yes |
| `errorType` | string | Yes |
| `timestamp` | string | Yes |
| `invitationMetaData` | object (map) | Yes |
| `eligibility` | LolLobbyEligibility | Yes |
| `fromSummonerName` | string | Yes |
| `toSummonerName` | string | Yes |

### LolLobbyLobbyInvitationDto

| Field | Type | Required |
|-------|------|----------|
| `invitationId` | string | Yes |
| `toSummonerId` | uint64 | Yes |
| `state` | LolLobbyLobbyInvitationState | Yes |
| `timestamp` | string | Yes |
| `toSummonerName` | string | Yes |
| `invitationType` | LolLobbyInvitationType | Yes |

### LolLobbyLobbyInvitationState

**Type**: enum string

**Values**: `Error`, `OnHold`, `Kicked`, `Declined`, `Joined`, `Accepted`, `Pending`, `Requested`


### LolLobbyLobbyMatchmakingLowPriorityDataResource

| Field | Type | Required |
|-------|------|----------|
| `penalizedSummonerIds` | integer[] | Yes |
| `penaltyTime` | number | Yes |
| `penaltyTimeRemaining` | number | Yes |
| `bustedLeaverAccessToken` | string | Yes |
| `reason` | string | Yes |

### LolLobbyLobbyMatchmakingSearchErrorResource

| Field | Type | Required |
|-------|------|----------|
| `id` | int32 | Yes |
| `errorType` | string | Yes |
| `penalizedSummonerId` | uint64 | Yes |
| `penaltyTimeRemaining` | number | Yes |
| `message` | string | Yes |

### LolLobbyLobbyMatchmakingSearchResource

| Field | Type | Required |
|-------|------|----------|
| `searchState` | LolLobbyLobbyMatchmakingSearchState | Yes |
| `lowPriorityData` | LolLobbyLobbyMatchmakingLowPriorityDataResource | Yes |
| `errors` | LolLobbyLobbyMatchmakingSearchErrorResource[] | Yes |

### LolLobbyLobbyMatchmakingSearchState

**Type**: enum string

**Values**: `ServiceShutdown`, `ServiceError`, `Error`, `Found`, `Searching`, `Canceled`, `AbandonedLowPriorityQueue`, `Invalid`


### LolLobbyLobbyNotification

| Field | Type | Required |
|-------|------|----------|
| `notificationId` | string | Yes |
| `notificationReason` | string | Yes |
| `summonerIds` | integer[] | Yes |
| `timestamp` | uint64 | Yes |

### LolLobbyLobbyParticipantDto

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `summonerIconId` | int32 | Yes |
| `summonerName` | string | Yes |
| `summonerInternalName` | string | Yes |
| `puuid` | string | Yes |
| `summonerLevel` | uint32 | Yes |
| `allowedStartActivity` | boolean | Yes |
| `allowedChangeActivity` | boolean | Yes |
| `allowedToggleInvite` | boolean | Yes |
| `allowedKickOthers` | boolean | Yes |
| `allowedInviteOthers` | boolean | Yes |
| `isLeader` | boolean | Yes |
| `isSpectator` | boolean | Yes |
| `teamId` | int32 | Yes |
| `firstPositionPreference` | string | Yes |
| `secondPositionPreference` | string | Yes |
| `memberData` | object (map) |  |
| `subteamIndex` | int8 |  |
| `intraSubteamPosition` | int8 |  |
| `strawberryMapId` | string |  |
| `playerSlots` | LolLobbyQuickPlayPresetSlotDto[] | Yes |
| `ready` | boolean | Yes |
| `showGhostedBanner` | boolean | Yes |
| `autoFillEligible` | boolean | Yes |
| `autoFillProtectedForStreaking` | boolean | Yes |
| `autoFillProtectedForPromos` | boolean | Yes |
| `autoFillProtectedForSoloing` | boolean | Yes |
| `autoFillProtectedForRemedy` | boolean | Yes |
| `isBot` | boolean | Yes |
| `botId` | string | Yes |
| `botDifficulty` | LolLobbyLobbyBotDifficulty | Yes |
| `botChampionId` | int32 | Yes |
| `botPosition` | string | Yes |
| `botUuid` | string | Yes |

### LolLobbyLobbyPartyRewards

| Field | Type | Required |
|-------|------|----------|
| `isEnabled` | boolean | Yes |
| `queueId` | int32 | Yes |
| `isCustom` | boolean | Yes |
| `partyRewards` | LolLobbyPartyReward[] | Yes |

### LolLobbyLobbyPositionPreferences

| Field | Type | Required |
|-------|------|----------|
| `firstPreference` | string | Yes |
| `secondPreference` | string | Yes |

### LolLobbyMucJwtDto

| Field | Type | Required |
|-------|------|----------|
| `jwt` | string | Yes |
| `channelClaim` | string | Yes |
| `domain` | string | Yes |
| `targetRegion` | string | Yes |

### LolLobbyPartyDto

| Field | Type | Required |
|-------|------|----------|
| `partyId` | string | Yes |
| `platformId` | string | Yes |
| `players` | LolLobbyPartyMemberDto[] | Yes |
| `chat` | LolLobbyPartyChatDto | Yes |
| `maxPartySize` | int32 | Yes |
| `partyType` | string | Yes |
| `gameMode` | LolLobbyGameModeDto | Yes |
| `activityLocked` | boolean | Yes |
| `version` | uint64 | Yes |
| `activityStartedUtcMillis` | uint64 | Yes |
| `activityResumeUtcMillis` | uint64 | Yes |
| `activeRestrictions` | LolLobbyQueueRestrictionDto | Yes |
| `eligibilityHash` | int64 | Yes |
| `eligibilityRestrictions` | LolLobbyGatekeeperRestrictionDto[] |  |
| `eligibilityWarnings` | LolLobbyGatekeeperRestrictionDto[] |  |
| `botParticipants` | LolLobbyBotParticipantDto[] |  |

### LolLobbyPartyMemberDto

| Field | Type | Required |
|-------|------|----------|
| `platformId` | string | Yes |
| `puuid` | string | Yes |
| `accountId` | uint64 | Yes |
| `summonerId` | uint64 | Yes |
| `partyId` | string | Yes |
| `partyVersion` | int64 | Yes |
| `role` | LolLobbyPartyMemberRoleEnum | Yes |
| `gameMode` | LolLobbyGameModeDto |  |
| `ready` | boolean |  |
| `metadata` | LolLobbyPartyMemberMetadataDto | Yes |
| `invitedBySummonerId` | uint64 |  |
| `inviteTimestamp` | uint64 |  |
| `canInvite` | boolean |  |
| `team` | string | Yes |

### LolLobbyPartyMemberMetadataDto

| Field | Type | Required |
|-------|------|----------|
| `positionPref` | string[] | Yes |
| `properties` | object (map) |  |
| `memberData` | object (map) |  |
| `playerSlots` | LolLobbyQuickPlayPresetSlotDto[] | Yes |
| `subteamData` | LolLobbySubteamDataDto |  |

### LolLobbyPartyReward

| Field | Type | Required |
|-------|------|----------|
| `premadeSize` | int32 | Yes |
| `type` | LolLobbyLobbyPartyRewardType | Yes |
| `value` | string | Yes |

### LolLobbyPartyStatusDto

| Field | Type | Required |
|-------|------|----------|
| `readyPlayers` | string[] | Yes |
| `leftPlayers` | string[] | Yes |
| `eogPlayers` | string[] | Yes |
| `partySize` | int32 | Yes |

### LolLobbyPlayerDto

| Field | Type | Required |
|-------|------|----------|
| `puuid` | string | Yes |
| `platformId` | string | Yes |
| `accountId` | uint64 | Yes |
| `summonerId` | uint64 | Yes |
| `eligibilityHash` | int64 | Yes |
| `serverUtcMillis` | int64 | Yes |
| `parties` | LolLobbyPartyMemberDto[] |  |
| `currentParty` | LolLobbyPartyDto |  |
| `registration` | LolLobbyRegistrationCredentials | Yes |
| `createdAt` | uint64 | Yes |
| `version` | uint64 | Yes |

### LolLobbyPlayerInviteDto

| Field | Type | Required |
|-------|------|----------|
| `puuid` | string | Yes |
| `inviteContext` | LolLobbyInviteContext | Yes |

### LolLobbyPremadePartyDto

| Field | Type | Required |
|-------|------|----------|
| `partyId` | string | Yes |
| `commsEnabled` | boolean | Yes |
| `players` | object (map) | Yes |

### LolLobbyQueueAvailability

**Type**: enum string

**Values**: `DoesntMeetRequirements`, `PlatformDisabled`, `Available`


### LolLobbyQuickPlayPresetSlotDto

| Field | Type | Required |
|-------|------|----------|
| `championId` | int32 | Yes |
| `skinId` | int32 | Yes |
| `positionPreference` | string | Yes |
| `perks` | string | Yes |
| `spell1` | uint64 | Yes |
| `spell2` | uint64 | Yes |

### LolLobbyReceivedInvitationDto

| Field | Type | Required |
|-------|------|----------|
| `invitationId` | string | Yes |
| `fromSummonerId` | uint64 | Yes |
| `state` | LolLobbyLobbyInvitationState | Yes |
| `timestamp` | string | Yes |
| `fromSummonerName` | string | Yes |
| `canAcceptInvitation` | boolean | Yes |
| `restrictions` | LolLobbyEligibilityRestriction[] | Yes |
| `gameConfig` | LolLobbyReceivedInvitationGameConfigDto | Yes |
| `invitationType` | LolLobbyInvitationType | Yes |

### LolLobbyReceivedInvitationGameConfigDto

| Field | Type | Required |
|-------|------|----------|
| `gameMode` | string | Yes |
| `queueId` | int32 | Yes |
| `mapId` | int32 | Yes |
| `inviteGameType` | string | Yes |

### LolLobbyRegistrationCredentials

| Field | Type | Required |
|-------|------|----------|
| `summonerToken` | string |  |
| `inventoryToken` | string |  |
| `inventoryTokens` | string[] |  |
| `simpleInventoryToken` | string |  |
| `rankedOverviewToken` | string |  |
| `gameClientVersion` | string |  |
| `playerTokens` | object (map) |  |
| `experiments` | object (map) |  |

### LolLobbyStrawberryMapUpdateDto

| Field | Type | Required |
|-------|------|----------|
| `contentId` | string | Yes |
| `itemId` | int32 | Yes |

### LolLobbySubteamDataDto

| Field | Type | Required |
|-------|------|----------|
| `subteamIndex` | int8 | Yes |
| `intraSubteamPosition` | int8 | Yes |

---

## LolLock


### LolLockAndLoadHomeHubsWaits

| Field | Type | Required |
|-------|------|----------|
| `initialWait` | uint8 | Yes |
| `additionalWait` | uint8 | Yes |

---

## LolLogin


### LolLoginAccountStateResource

| Field | Type | Required |
|-------|------|----------|
| `state` | LolLoginAccountStateType | Yes |

### LolLoginAccountStateType

**Type**: enum string

**Values**: `GENERATING`, `TRANSFERRED_OUT`, `TRANSFERRING_IN`, `TRANSFERRING_OUT`, `ENABLED`, `CREATING`


### LolLoginLcdsResponse

| Field | Type | Required |
|-------|------|----------|
| `typeName` | string | Yes |
| `body` | object (map) | Yes |

### LolLoginLeagueSessionStatus

**Type**: enum string

**Values**: `ANTI_ADDICTION_EXPIRED`, `DUPLICATED`, `EXPIRED`, `INITIALIZED`, `UNINITIALIZED`


### LolLoginLeagueSessionTokenEnvelope

| Field | Type | Required |
|-------|------|----------|
| `token` | string |  |
| `logoutOnFailure` | boolean | Yes |

### LolLoginLoginConnectionMode

**Type**: enum string

**Values**: `RiotClient`, `Partner`, `Legacy`, `Preparing`


### LolLoginLoginConnectionState

| Field | Type | Required |
|-------|------|----------|
| `mode` | LolLoginLoginConnectionMode | Yes |
| `isUsingDeveloperAuthToken` | boolean | Yes |
| `isPartnerRiotClient` | boolean | Yes |

### LolLoginLoginError

| Field | Type | Required |
|-------|------|----------|
| `errorCode` | string | Yes |
| `messageId` | string | Yes |
| `description` | string | Yes |

### LolLoginLoginQueue

| Field | Type | Required |
|-------|------|----------|
| `estimatedPositionInQueue` | uint64 | Yes |
| `approximateWaitTimeSeconds` | uint64 |  |
| `maxDisplayedPosition` | uint64 |  |
| `maxDisplayedWaitTimeSeconds` | uint64 |  |

### LolLoginLoginSession

| Field | Type | Required |
|-------|------|----------|
| `state` | LolLoginLoginSessionStates | Yes |
| `username` | string | Yes |
| `userAuthToken` | string | Yes |
| `accountId` | uint64 | Yes |
| `summonerId` | uint64 |  |
| `isInLoginQueue` | boolean | Yes |
| `error` | LolLoginLoginError |  |
| `idToken` | string | Yes |
| `puuid` | string | Yes |
| `isNewPlayer` | boolean | Yes |
| `connected` | boolean | Yes |

### LolLoginLoginSessionStates

**Type**: enum string

**Values**: `ERROR`, `LOGGING_OUT`, `SUCCEEDED`, `IN_PROGRESS`


### LolLoginLoginSessionWallet

| Field | Type | Required |
|-------|------|----------|
| `ip` | int64 | Yes |
| `rp` | int64 | Yes |

### LolLoginPlatformGeneratedCredentials

| Field | Type | Required |
|-------|------|----------|
| `username` | string | Yes |
| `password` | string | Yes |

### LolLoginSummonerSessionResource

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `displayName` | string | Yes |
| `isNewPlayer` | boolean | Yes |

---

## LolLoot


### LolLootContextMenu

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `actionType` | string | Yes |
| `recipeDescription` | string | Yes |
| `recipeContextMenuAction` | string | Yes |
| `enabled` | boolean | Yes |
| `essenceType` | string | Yes |
| `essenceQuantity` | int32 | Yes |
| `requiredTokens` | string | Yes |
| `requiredOthers` | string | Yes |
| `requiredOthersName` | string | Yes |
| `requiredOthersCount` | int32 | Yes |

### LolLootItemOwnershipStatus

**Type**: enum string

**Values**: `OWNED`, `RENTAL`, `FREE`, `NONE`


### LolLootLootGrantNotification

| Field | Type | Required |
|-------|------|----------|
| `id` | int64 | Yes |
| `gameId` | uint64 | Yes |
| `playerId` | uint64 | Yes |
| `championId` | int32 | Yes |
| `playerGrade` | string | Yes |
| `lootName` | string | Yes |
| `messageKey` | string | Yes |
| `msgId` | string | Yes |
| `accountId` | uint64 | Yes |

### LolLootLootItem

| Field | Type | Required |
|-------|------|----------|
| `lootName` | string | Yes |
| `asset` | string | Yes |
| `type` | string | Yes |
| `rarity` | string | Yes |
| `value` | int32 | Yes |
| `storeItemId` | int32 | Yes |
| `upgradeLootName` | string | Yes |
| `expiryTime` | int64 | Yes |
| `tags` | string | Yes |
| `displayCategories` | string | Yes |
| `rentalSeconds` | int64 | Yes |
| `rentalGames` | int32 | Yes |

### LolLootLootItemGdsResource

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `name` | string | Yes |
| `description` | string | Yes |
| `image` | string | Yes |
| `startDate` | string | Yes |
| `endDate` | string | Yes |
| `recipeMenuTitle` | string | Yes |
| `recipeMenuSubtitle` | string | Yes |
| `mappedStoreId` | int32 | Yes |
| `lifetimeMax` | int32 | Yes |
| `autoRedeem` | boolean | Yes |
| `recipeMenuActive` | boolean | Yes |
| `rarity` | LolLootLootRarity | Yes |
| `type` | LolLootLootType | Yes |

### LolLootLootMilestone

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `threshold` | uint64 | Yes |
| `rewards` | LolLootLootMilestoneReward[] | Yes |

### LolLootLootMilestoneClaimStatus

**Type**: enum string

**Values**: `FAILED`, `COMPLETED`, `IN_PROGRESS`, `NOT_STARTED`


### LolLootLootMilestoneRepeat

| Field | Type | Required |
|-------|------|----------|
| `repeatCount` | int32 | Yes |
| `repeatScope` | int32 | Yes |
| `multiplier` | number | Yes |

### LolLootLootMilestones

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `progressionConfigId` | string | Yes |
| `active` | boolean | Yes |
| `startDate` | string | Yes |
| `endDate` | string | Yes |
| `storeGroupTitle` | string | Yes |
| `repeat` | LolLootLootMilestoneRepeat | Yes |
| `lootItems` | string[] | Yes |
| `recipes` | string[] | Yes |
| `milestones` | LolLootLootMilestone[] | Yes |
| `errorCachingMilestoneSet` | boolean | Yes |

### LolLootLootMilestonesClaimResponse

| Field | Type | Required |
|-------|------|----------|
| `lootMilestoneSetId` | string | Yes |
| `claimedMilestones` | string[] | Yes |
| `status` | LolLootLootMilestoneClaimStatus | Yes |

### LolLootLootMilestonesCounter

| Field | Type | Required |
|-------|------|----------|
| `lootMilestonesId` | string | Yes |
| `counterValue` | int64 | Yes |
| `completedLoops` | int64 | Yes |
| `readyToClaimMilestones` | string[] | Yes |

### LolLootLootOddsResponse

| Field | Type | Required |
|-------|------|----------|
| `lootId` | string | Yes |
| `parentId` | string | Yes |
| `dropRate` | number | Yes |
| `quantity` | int32 | Yes |
| `label` | string | Yes |
| `query` | string | Yes |
| `lootOrder` | int32 | Yes |
| `children` | LolLootLootOddsResponse[] | Yes |

### LolLootLootRarity

**Type**: enum string

**Values**: `Ultimate`, `Mythic`, `Legendary`, `Epic`, `Default`


### LolLootLootType

**Type**: enum string

**Values**: `Nexus_Finisher`, `TFT_Damage_Skin`, `TFT_Map_Skin`, `SkinBorder`, `Boost`, `Statstone_Shard`, `Statstone`, `Egg_Color`, `Egg`, `Companion`, `SummonerIcon`, `Skin_Rental`, `Skin`, `WardSkin`, `Material`, `Currency`, `Chest`


### LolLootMassDisenchantClientConfig

| Field | Type | Required |
|-------|------|----------|
| `maxLootItemsSizeMassCraft` | int16 | Yes |
| `enabled` | boolean | Yes |

### LolLootPlayerLoot

| Field | Type | Required |
|-------|------|----------|
| `lootName` | string | Yes |
| `lootId` | string | Yes |
| `refId` | string | Yes |
| `localizedName` | string | Yes |
| `localizedDescription` | string | Yes |
| `itemDesc` | string | Yes |
| `displayCategories` | string | Yes |
| `rarity` | string | Yes |
| `tags` | string | Yes |
| `type` | string | Yes |
| `asset` | string | Yes |
| `tilePath` | string | Yes |
| `splashPath` | string | Yes |
| `shadowPath` | string | Yes |
| `upgradeLootName` | string | Yes |
| `upgradeEssenceName` | string | Yes |
| `disenchantLootName` | string | Yes |
| `localizedRecipeTitle` | string | Yes |
| `localizedRecipeSubtitle` | string | Yes |
| `itemStatus` | LolLootItemOwnershipStatus | Yes |
| `parentItemStatus` | LolLootItemOwnershipStatus | Yes |
| `redeemableStatus` | LolLootRedeemableStatus | Yes |
| `count` | int32 | Yes |
| `rentalGames` | int32 | Yes |
| `storeItemId` | int32 | Yes |
| `parentStoreItemId` | int32 | Yes |
| `value` | int32 | Yes |
| `upgradeEssenceValue` | int32 | Yes |
| `disenchantValue` | int32 | Yes |
| `disenchantRecipeName` | string | Yes |
| `expiryTime` | int64 | Yes |
| `rentalSeconds` | int64 | Yes |
| `isNew` | boolean | Yes |
| `isRental` | boolean | Yes |

### LolLootPlayerLootDelta

| Field | Type | Required |
|-------|------|----------|
| `deltaCount` | int32 | Yes |
| `playerLoot` | LolLootPlayerLoot | Yes |

### LolLootPlayerLootMap

| Field | Type | Required |
|-------|------|----------|
| `version` | int64 | Yes |
| `playerLoot` | object (map) | Yes |

### LolLootPlayerLootNotification

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `count` | int32 | Yes |
| `acknowledged` | boolean | Yes |

### LolLootPlayerLootUpdate

| Field | Type | Required |
|-------|------|----------|
| `added` | LolLootPlayerLootDelta[] | Yes |
| `removed` | LolLootPlayerLootDelta[] | Yes |
| `redeemed` | LolLootPlayerLootDelta[] | Yes |

### LolLootQueryEvaluatedLootItem

| Field | Type | Required |
|-------|------|----------|
| `lootName` | string | Yes |
| `localizedName` | string | Yes |

### LolLootRecipeMetadata

| Field | Type | Required |
|-------|------|----------|
| `guaranteedDescriptions` | LolLootLootDescription[] | Yes |
| `bonusDescriptions` | LolLootLootDescription[] | Yes |
| `tooltipsDisabled` | boolean | Yes |

### LolLootRecipeOutput

| Field | Type | Required |
|-------|------|----------|
| `lootName` | string | Yes |
| `quantity` | int32 | Yes |

### LolLootRecipeSlot

| Field | Type | Required |
|-------|------|----------|
| `slotNumber` | int32 | Yes |
| `lootIds` | string[] | Yes |
| `tags` | string | Yes |
| `quantity` | int32 | Yes |

### LolLootRecipeWithMilestones

| Field | Type | Required |
|-------|------|----------|
| `recipeName` | string | Yes |
| `type` | string | Yes |
| `description` | string | Yes |
| `contextMenuText` | string | Yes |
| `requirementText` | string | Yes |
| `imagePath` | string | Yes |
| `introVideoPath` | string | Yes |
| `loopVideoPath` | string | Yes |
| `outroVideoPath` | string | Yes |
| `displayCategories` | string | Yes |
| `crafterName` | string | Yes |
| `slots` | LolLootRecipeSlot[] | Yes |
| `outputs` | LolLootRecipeOutput[] | Yes |
| `metadata` | LolLootRecipeMetadata | Yes |
| `singleOpen` | boolean | Yes |
| `hasVisibleLootOdds` | boolean | Yes |
| `lootMilestoneIds` | string[] | Yes |

### LolLootRedeemableStatus

**Type**: enum string

**Values**: `SKIN_NOT_OWNED`, `CHAMPION_NOT_OWNED`, `ALREADY_RENTED`, `ALREADY_OWNED`, `NOT_REDEEMABLE_RENTAL`, `NOT_REDEEMABLE`, `REDEEMABLE_RENTAL`, `REDEEMABLE`, `UNKNOWN`


### LolLootVerboseLootOddsResponse

| Field | Type | Required |
|-------|------|----------|
| `recipeName` | string | Yes |
| `chanceToContain` | LolLootLootOddsResponse[] | Yes |
| `guaranteedToContain` | LolLootLootOddsResponse[] | Yes |
| `lootItem` | LolLootPlayerLoot | Yes |
| `hasPityRules` | boolean | Yes |
| `checksOwnership` | boolean | Yes |

---

## LolLoyalty


### LolLoyaltyGlobalRewards

| Field | Type | Required |
|-------|------|----------|
| `allChampions` | boolean | Yes |

### LolLoyaltyLoyaltyRewards

| Field | Type | Required |
|-------|------|----------|
| `freeRewardedChampionsCount` | int32 | Yes |
| `championIds` | integer[] | Yes |
| `freeRewardedSkinsCount` | int32 | Yes |
| `skinIds` | integer[] | Yes |
| `global` | LolLoyaltyGlobalRewards | Yes |
| `ipBoost` | int32 | Yes |
| `xpBoost` | object (map) | Yes |
| `loyaltyTFTMapSkinCount` | int32 | Yes |
| `loyaltyTFTCompanionCount` | int32 | Yes |
| `loyaltyTFTDamageSkinCount` | int32 | Yes |
| `loyaltySources` | object (map) | Yes |

### LolLoyaltyLoyaltyRewardsSimplified

| Field | Type | Required |
|-------|------|----------|
| `freeRewardedChampionsCount` | int32 | Yes |
| `championIds` | integer[] | Yes |
| `freeRewardedSkinsCount` | int32 | Yes |
| `skinIds` | integer[] | Yes |
| `global` | LolLoyaltyGlobalRewards | Yes |
| `ipBoost` | int32 | Yes |
| `xpBoost` | int32 | Yes |
| `loyaltyTFTMapSkinCount` | int32 | Yes |
| `loyaltyTFTCompanionCount` | int32 | Yes |
| `loyaltyTFTDamageSkinCount` | int32 | Yes |
| `loyaltySources` | object (map) | Yes |

### LolLoyaltyLoyaltyStatus

**Type**: enum string

**Values**: `DISABLED`, `REVOKE`, `CHANGE`, `EXPIRY`, `REWARDS_GRANT`, `LEGACY`


### LolLoyaltyLoyaltyStatusNotification

| Field | Type | Required |
|-------|------|----------|
| `status` | LolLoyaltyLoyaltyStatus | Yes |
| `rewards` | LolLoyaltyLoyaltyRewardsSimplified | Yes |
| `reloadInventory` | boolean | Yes |

---

## LolMac


### LolMacGraphicsUpgradeMacGraphicsUpgradeNotificationType

**Type**: enum string

**Values**: `HARDWARE_UPGRADE`, `NONE`


---

## LolMaps


### LolMapsMaps

| Field | Type | Required |
|-------|------|----------|
| `id` | int64 | Yes |
| `isDefault` | boolean | Yes |
| `gameMode` | string | Yes |
| `gameModeName` | string | Yes |
| `gameModeShortName` | string | Yes |
| `gameModeDescription` | string | Yes |
| `gameMutator` | string | Yes |
| `isRGM` | boolean | Yes |
| `name` | string | Yes |
| `description` | string | Yes |
| `mapStringId` | string | Yes |
| `platformId` | string | Yes |
| `platformName` | string | Yes |
| `assets` | object (map) | Yes |
| `locStrings` | object (map) | Yes |
| `setModalButtonBottom` | int32 | Yes |
| `setModalButtonLeft` | int32 | Yes |
| `categorizedContentBundles` | object (map) | Yes |
| `tutorialCards` | LolMapsTutorialCard[] | Yes |
| `properties` | object (map) | Yes |
| `perPositionRequiredSummonerSpells` | object (map) | Yes |
| `perPositionDisallowedSummonerSpells` | object (map) | Yes |
| `tftSetOverride` | string | Yes |

### LolMapsTutorialCard

| Field | Type | Required |
|-------|------|----------|
| `header` | string |  |
| `footer` | string |  |
| `description` | string |  |
| `imagePath` | string | Yes |

---

## LolMarketplace


### LolMarketplaceCatalogEntryDto

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `productId` | string | Yes |
| `name` | string | Yes |
| `description` | string | Yes |
| `endTime` | string | Yes |
| `purchaseUnits` | LolMarketplacePurchaseUnitDto[] | Yes |
| `displayMetadata` | object (map) | Yes |
| `refundRule` | string | Yes |
| `giftRule` | string | Yes |
| `prerequisites` | LolMarketplacePrerequisiteDto[] | Yes |
| `purchaseLimits` | LolMarketplaceVelocityLimitDeltaDto[] | Yes |

### LolMarketplacePagination

| Field | Type | Required |
|-------|------|----------|
| `offset` | uint32 | Yes |
| `limit` | uint32 | Yes |
| `maxLimit` | uint32 | Yes |
| `total` | uint32 | Yes |
| `previous` | string | Yes |
| `next` | string | Yes |

### LolMarketplacePurchaseDto

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `productId` | string | Yes |
| `storeId` | string | Yes |
| `storeName` | string | Yes |
| `catalogEntryId` | string | Yes |
| `catalogEntryName` | string | Yes |
| `purchaserId` | string | Yes |
| `recipientId` | string | Yes |
| `purchaseUnits` | LolMarketplaceFinalPurchaseUnitDto[] | Yes |
| `createdTime` | string | Yes |
| `completedTime` | string | Yes |
| `purchaseState` | string | Yes |
| `purchaseVisibility` | string | Yes |
| `refund` | LolMarketplaceRefundDto | Yes |
| `refundRule` | string | Yes |
| `refundable` | boolean | Yes |
| `quantity` | int64 | Yes |
| `source` | string | Yes |

### LolMarketplacePurchaseHistoryResponse

| Field | Type | Required |
|-------|------|----------|
| `data` | LolMarketplacePurchaseDto[] | Yes |
| `paging` | LolMarketplacePagination | Yes |
| `stats` | LolMarketplaceResponseStats | Yes |
| `notes` | string[] | Yes |
| `errors` | LolMarketplaceResponseError[] | Yes |

### LolMarketplacePurchaseRequest

| Field | Type | Required |
|-------|------|----------|
| `storeId` | string | Yes |
| `catalogEntryId` | string | Yes |
| `paymentOptionsKeys` | string[] | Yes |
| `quantity` | uint32 | Yes |

### LolMarketplacePurchaseResponse

| Field | Type | Required |
|-------|------|----------|
| `data` | LolMarketplacePurchaseDto | Yes |
| `paging` | LolMarketplacePagination | Yes |
| `stats` | LolMarketplaceResponseStats | Yes |
| `notes` | string[] | Yes |
| `errors` | LolMarketplaceResponseError[] | Yes |

### LolMarketplacePurchaseTransaction

| Field | Type | Required |
|-------|------|----------|
| `purchaseId` | string | Yes |
| `productId` | string | Yes |
| `storeId` | string | Yes |
| `catalogEntryId` | string | Yes |
| `purchaseState` | string | Yes |
| `refundId` | string | Yes |
| `refundState` | string | Yes |

### LolMarketplaceRefundDto

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `purchaseId` | string | Yes |
| `createdTime` | string | Yes |
| `completedTime` | string | Yes |
| `state` | string | Yes |
| `source` | string | Yes |

### LolMarketplaceRefundRequest

| Field | Type | Required |
|-------|------|----------|
| `purchaseId` | string | Yes |

### LolMarketplaceRefundResponse

| Field | Type | Required |
|-------|------|----------|
| `data` | LolMarketplaceRefundDto | Yes |
| `paging` | LolMarketplacePagination | Yes |
| `stats` | LolMarketplaceResponseStats | Yes |
| `notes` | string[] | Yes |
| `errors` | LolMarketplaceResponseError[] | Yes |

### LolMarketplaceResponseError

| Field | Type | Required |
|-------|------|----------|
| `message` | string | Yes |
| `type` | string | Yes |
| `code` | uint32 | Yes |

### LolMarketplaceResponseStats

| Field | Type | Required |
|-------|------|----------|
| `durationMs` | uint32 | Yes |

### LolMarketplaceStoreDto

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `productId` | string | Yes |
| `name` | string | Yes |
| `catalogEntries` | LolMarketplaceCatalogEntryDto[] | Yes |
| `displayMetadata` | object (map) | Yes |
| `startTime` | string | Yes |
| `endTime` | string | Yes |

### LolMarketplaceStoresResponse

| Field | Type | Required |
|-------|------|----------|
| `data` | LolMarketplaceStoreDto[] | Yes |
| `paging` | LolMarketplacePagination | Yes |
| `stats` | LolMarketplaceResponseStats | Yes |
| `notes` | string[] | Yes |
| `errors` | LolMarketplaceResponseError[] | Yes |

### LolMarketplaceTFTRotationalShopConfig

| Field | Type | Required |
|-------|------|----------|
| `enabled` | boolean | Yes |
| `contentPreviewConfig` | object (map) | Yes |
| `uiToggleEnabled` | boolean | Yes |
| `navs` | LolMarketplaceTFTRotationalShopNavConfig[] | Yes |
| `refund` | LolMarketplaceTFTRotationalShopRefundConfig | Yes |
| `littleLegendsUpgradeEnabled` | boolean | Yes |
| `eventsStoreEnabled` | boolean | Yes |
| `eventsStoreData` | object (map) | Yes |
| `eventsMusicDisabled` | boolean | Yes |

### LolMarketplaceTFTRotationalShopNavConfig

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `enabled` | boolean | Yes |
| `supportedCurrencies` | string[] | Yes |

### LolMarketplaceTFTRotationalShopRefundConfig

| Field | Type | Required |
|-------|------|----------|
| `enabled` | boolean | Yes |
| `allowedTypes` | string[] | Yes |
| `thresholdDays` | uint8 | Yes |

---

## LolMatch


### LolMatchHistoryGAMHSMatchHistoryData

| Field | Type | Required |
|-------|------|----------|
| `metadata` | LolMatchHistoryGAMHSMatchHistoryMetadata | Yes |
| `json` | object (map) | Yes |

### LolMatchHistoryGAMHSMatchHistoryList

| Field | Type | Required |
|-------|------|----------|
| `games` | LolMatchHistoryGAMHSMatchHistoryData[] | Yes |
| `active_puuid` | string | Yes |

### LolMatchHistoryMatchHistoryGame

| Field | Type | Required |
|-------|------|----------|
| `endOfGameResult` | string | Yes |
| `gameId` | uint64 | Yes |
| `platformId` | string | Yes |
| `gameCreation` | uint64 | Yes |
| `gameCreationDate` | string | Yes |
| `gameDuration` | uint32 | Yes |
| `queueId` | int32 | Yes |
| `mapId` | uint16 | Yes |
| `seasonId` | uint16 | Yes |
| `gameVersion` | string | Yes |
| `gameMode` | string | Yes |
| `gameModeMutators` | string[] | Yes |
| `gameType` | string | Yes |
| `teams` | LolMatchHistoryMatchHistoryTeam[] | Yes |
| `participants` | LolMatchHistoryMatchHistoryParticipant[] | Yes |
| `participantIdentities` | LolMatchHistoryMatchHistoryParticipantIdentities[] | Yes |

### LolMatchHistoryMatchHistoryGameList

| Field | Type | Required |
|-------|------|----------|
| `gameIndexBegin` | uint64 | Yes |
| `gameIndexEnd` | uint64 | Yes |
| `gameBeginDate` | string | Yes |
| `gameEndDate` | string | Yes |
| `gameCount` | uint64 | Yes |
| `games` | LolMatchHistoryMatchHistoryGame[] | Yes |

### LolMatchHistoryMatchHistoryList

| Field | Type | Required |
|-------|------|----------|
| `platformId` | string | Yes |
| `accountId` | uint64 | Yes |
| `games` | LolMatchHistoryMatchHistoryGameList | Yes |

### LolMatchHistoryMatchHistoryParticipant

| Field | Type | Required |
|-------|------|----------|
| `participantId` | uint16 | Yes |
| `teamId` | uint16 | Yes |
| `championId` | int32 | Yes |
| `spell1Id` | uint16 | Yes |
| `spell2Id` | uint16 | Yes |
| `highestAchievedSeasonTier` | string | Yes |
| `stats` | LolMatchHistoryMatchHistoryParticipantStatistics | Yes |
| `timeline` | LolMatchHistoryMatchHistoryTimeline | Yes |

### LolMatchHistoryMatchHistoryParticipantIdentities

| Field | Type | Required |
|-------|------|----------|
| `participantId` | uint16 | Yes |
| `player` | LolMatchHistoryMatchHistoryParticipantIdentityPlayer | Yes |

### LolMatchHistoryMatchHistoryTeam

| Field | Type | Required |
|-------|------|----------|
| `teamId` | uint16 | Yes |
| `win` | string | Yes |
| `firstBlood` | boolean | Yes |
| `firstTower` | boolean | Yes |
| `firstInhibitor` | boolean | Yes |
| `firstBaron` | boolean | Yes |
| `firstDargon` | boolean | Yes |
| `towerKills` | uint32 | Yes |
| `inhibitorKills` | uint32 | Yes |
| `baronKills` | uint32 | Yes |
| `dragonKills` | uint32 | Yes |
| `vilemawKills` | uint32 | Yes |
| `riftHeraldKills` | uint32 | Yes |
| `hordeKills` | uint32 | Yes |
| `dominionVictoryScore` | uint32 | Yes |
| `bans` | LolMatchHistoryMatchHistoryTeamBan[] | Yes |

### LolMatchHistoryMatchHistoryTimelineFrame

| Field | Type | Required |
|-------|------|----------|
| `participantFrames` | object (map) | Yes |
| `events` | LolMatchHistoryMatchHistoryEvent[] | Yes |
| `timestamp` | uint64 | Yes |

### LolMatchHistoryMatchHistoryTimelineFrames

| Field | Type | Required |
|-------|------|----------|
| `frames` | LolMatchHistoryMatchHistoryTimelineFrame[] | Yes |

### LolMatchHistoryRecentlyPlayedSummoner

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `summonerName` | string | Yes |
| `gameName` | string | Yes |
| `tagLine` | string | Yes |
| `gameId` | uint64 | Yes |
| `gameCreationDate` | string | Yes |
| `championId` | uint64 | Yes |
| `teamId` | uint64 | Yes |
| `puuid` | string | Yes |

---

## LolMatchmaking


### LolMatchmakingMatchmakingDodgeData

| Field | Type | Required |
|-------|------|----------|
| `state` | LolMatchmakingMatchmakingDodgeState | Yes |
| `dodgerId` | uint64 | Yes |

### LolMatchmakingMatchmakingDodgeWarning

**Type**: enum string

**Values**: `ConnectionWarning`, `Penalty`, `Warning`, `None`


### LolMatchmakingMatchmakingLowPriorityData

| Field | Type | Required |
|-------|------|----------|
| `penalizedSummonerIds` | integer[] | Yes |
| `penaltyTime` | number | Yes |
| `penaltyTimeRemaining` | number | Yes |
| `bustedLeaverAccessToken` | string | Yes |
| `reason` | string | Yes |

### LolMatchmakingMatchmakingReadyCheckResource

| Field | Type | Required |
|-------|------|----------|
| `state` | LolMatchmakingMatchmakingReadyCheckState | Yes |
| `playerResponse` | LolMatchmakingMatchmakingReadyCheckResponse | Yes |
| `dodgeWarning` | LolMatchmakingMatchmakingDodgeWarning | Yes |
| `timer` | number | Yes |
| `declinerIds` | integer[] | Yes |
| `suppressUx` | boolean | Yes |

### LolMatchmakingMatchmakingReadyCheckResponse

**Type**: enum string

**Values**: `Declined`, `Accepted`, `None`


### LolMatchmakingMatchmakingReadyCheckState

**Type**: enum string

**Values**: `Error`, `PartyNotReady`, `StrangerNotReady`, `EveryoneReady`, `InProgress`, `Invalid`


### LolMatchmakingMatchmakingSearchErrorResource

| Field | Type | Required |
|-------|------|----------|
| `id` | int32 | Yes |
| `errorType` | string | Yes |
| `penalizedSummonerId` | uint64 | Yes |
| `penaltyTimeRemaining` | number | Yes |
| `message` | string | Yes |

### LolMatchmakingMatchmakingSearchResource

| Field | Type | Required |
|-------|------|----------|
| `queueId` | int32 | Yes |
| `isCurrentlyInQueue` | boolean | Yes |
| `lobbyId` | string | Yes |
| `searchState` | LolMatchmakingMatchmakingSearchState | Yes |
| `timeInQueue` | number | Yes |
| `estimatedQueueTime` | number | Yes |
| `readyCheck` | LolMatchmakingMatchmakingReadyCheckResource | Yes |
| `dodgeData` | LolMatchmakingMatchmakingDodgeData | Yes |
| `lowPriorityData` | LolMatchmakingMatchmakingLowPriorityData | Yes |
| `errors` | LolMatchmakingMatchmakingSearchErrorResource[] | Yes |

### LolMatchmakingMatchmakingSearchState

**Type**: enum string

**Values**: `ServiceShutdown`, `ServiceError`, `Error`, `Found`, `Searching`, `Canceled`, `AbandonedLowPriorityQueue`, `Invalid`


---

## LolMetagames


### LolMetagamesPurchaseResponse

| Field | Type | Required |
|-------|------|----------|
| `idempotencyId` | string | Yes |
| `status` | LolMetagamesPurchaseResponseStatus | Yes |
| `numberOfPendingPurchases` | uint8 | Yes |
| `errorMessage` | string | Yes |

### LolMetagamesPurchaseResponseStatus

**Type**: enum string

**Values**: `Failure`, `Success`, `Pending`, `None`


---

## LolMetagamesplayer


### LolMetagamesplayerEventPayload

| Field | Type | Required |
|-------|------|----------|
| `playerGameData` | object (map) | Yes |
| `inventoryTypes` | string[] | Yes |
| `currencyTypes` | string[] | Yes |
| `storeId` | string | Yes |
| `paymentOption` | string | Yes |

### LolMetagamesplayerEventResponseDTO

| Field | Type | Required |
|-------|------|----------|
| `playerGameData` | object (map) | Yes |
| `attemptedPurchase` | boolean | Yes |
| `purchaseIdToObserve` | string | Yes |

---

## LolMissions


### LolMissionsRewardGroupsSelection

| Field | Type | Required |
|-------|------|----------|
| `rewardGroups` | string[] | Yes |

### LolMissionsSeriesOpt

| Field | Type | Required |
|-------|------|----------|
| `seriesId` | string | Yes |
| `option` | string | Yes |

---

## LolNacho


### LolNachoBannerOddsInfo

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `name` | string | Yes |
| `productId` | string | Yes |
| `rewardTables` | LolNachoNachoRollRewardsTable[] | Yes |
| `endDateMilis` | string | Yes |
| `bundledMythicEssence` | boolean | Yes |

### LolNachoBlessingTokenPurchaseRequest

| Field | Type | Required |
|-------|------|----------|
| `quantity` | uint32 | Yes |

### LolNachoCapCounterData

| Field | Type | Required |
|-------|------|----------|
| `amount` | uint32 | Yes |
| `version` | uint32 | Yes |
| `active` | boolean | Yes |
| `lastModifiedDate` | string | Yes |

### LolNachoGameDataBannerSkin

| Field | Type | Required |
|-------|------|----------|
| `id` | uint32 | Yes |
| `name` | string | Yes |
| `rarity` | string | Yes |

### LolNachoGameDataNachoBannerVo

| Field | Type | Required |
|-------|------|----------|
| `path` | string | Yes |
| `defaultDelayMillis` | uint32 | Yes |
| `localeOverrides` | LolNachoGameDataNachoBannerVoOverrideOptions[] | Yes |

### LolNachoGameDataNachoCurrency

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `name` | string | Yes |
| `currencyId` | string | Yes |
| `capCatalogEntryId` | string | Yes |

### LolNachoGameDataPityCounter

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `name` | string | Yes |

### LolNachoNachoBannersResponse

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `bannerBackgroundTexture` | string | Yes |
| `bannerBackgroundParallax` | string | Yes |
| `bannerChaseAnimationWebmPath` | string | Yes |
| `bannerChaseAnimationParallax` | string | Yes |
| `chasePityCounter` | LolNachoGameDataPityCounter | Yes |
| `chasePityThreshold` | uint32 | Yes |
| `highlightPityThreshold` | uint32 | Yes |
| `rollVignetteSkinIntroWebmPath` | string | Yes |
| `rollVignetteSkinIntroSfxPath` | string | Yes |
| `chaseCelebrationIntroWebmPath` | string | Yes |
| `chaseCelebrationVo` | LolNachoGameDataNachoBannerVo | Yes |
| `hubIntroVo` | LolNachoGameDataNachoBannerVo | Yes |
| `rollVignette` | LolNachoNachoVignette | Yes |
| `bannerSkin` | LolNachoGameDataBannerSkin | Yes |
| `bannerCurrency` | LolNachoGameDataNachoCurrency | Yes |
| `capCatalogStoreId` | string | Yes |
| `capCatalogEntryId` | string | Yes |
| `pityCounter` | LolNachoCapCounterData | Yes |
| `startDate` | int64 | Yes |
| `endDate` | int64 | Yes |

### LolNachoNachoPurchaseResponse

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `status` | LolNachoNachoPurchaseResponseStatus | Yes |
| `rollResults` | LolNachoNachoRewardData[] | Yes |

### LolNachoNachoPurchaseResponseStatus

**Type**: enum string

**Values**: `Failure`, `Success`, `Pending`, `None`


### LolNachoNachoRewardData

| Field | Type | Required |
|-------|------|----------|
| `odds` | number | Yes |
| `itemInstanceId` | string | Yes |
| `translatedName` | string | Yes |
| `id` | uint32 | Yes |
| `type` | string | Yes |
| `parentId` | uint32 | Yes |
| `priority` | uint32 | Yes |
| `isChaseItem` | boolean | Yes |
| `quantity` | uint32 | Yes |

### LolNachoNachoRollRewardsTable

| Field | Type | Required |
|-------|------|----------|
| `odds` | number | Yes |
| `translatedName` | string | Yes |
| `priority` | uint32 | Yes |
| `children` | LolNachoNachoRewardData[] | Yes |
| `fallbackChildren` | LolNachoNachoRewardData[] | Yes |

### LolNachoNachoVignette

| Field | Type | Required |
|-------|------|----------|
| `introTierOneWebmPath` | string | Yes |
| `introTierOneMultiWebmPath` | string | Yes |
| `introTierTwoWebmPath` | string | Yes |
| `introTierTwoMultiWebmPath` | string | Yes |
| `introTierThreeWebmPath` | string | Yes |
| `introTierThreeMultiWebmPath` | string | Yes |

---

## LolNpe


### LolNpeRewardsChallengesProgress

| Field | Type | Required |
|-------|------|----------|
| `progress` | LolNpeRewardsProgress | Yes |

### LolNpeRewardsCoopAiRoutingExperiment

| Field | Type | Required |
|-------|------|----------|
| `enabled` | boolean | Yes |
| `routingData` | LolNpeRewardsRoutingData[] | Yes |

### LolNpeRewardsPoroExperimentData

| Field | Type | Required |
|-------|------|----------|
| `coopAiRoutingExperiment` | LolNpeRewardsCoopAiRoutingExperiment | Yes |
| `summonerLevel` | int32 | Yes |

### LolNpeRewardsProgress

| Field | Type | Required |
|-------|------|----------|
| `lastViewedProgress` | int32 | Yes |
| `currentProgress` | int32 | Yes |
| `totalCount` | int32 | Yes |

### LolNpeRewardsRewardPack

| Field | Type | Required |
|-------|------|----------|
| `index` | int32 | Yes |
| `type` | string | Yes |
| `requirements` | LolNpeRewardsRequirements | Yes |
| `state` | string | Yes |
| `premiumReward` | boolean | Yes |
| `rewardKey` | string | Yes |
| `majorReward` | LolNpeRewardsReward | Yes |
| `minorRewards` | LolNpeRewardsReward[] | Yes |
| `delay` | int64 | Yes |
| `unlockTime` | int64 | Yes |

### LolNpeRewardsRewardSeries

| Field | Type | Required |
|-------|------|----------|
| `rewardPacks` | LolNpeRewardsRewardPack[] | Yes |

### LolNpeRewardsRewardSeriesState

| Field | Type | Required |
|-------|------|----------|
| `allRewardsClaimed` | boolean | Yes |

### LolNpeTutorialPathAccountSettingsTutorial

| Field | Type | Required |
|-------|------|----------|
| `hasSeenTutorialPath` | boolean | Yes |
| `hasSkippedTutorialPath` | boolean | Yes |
| `shouldSeeNewPlayerExperience` | boolean | Yes |

### LolNpeTutorialPathCollectionsChampion

| Field | Type | Required |
|-------|------|----------|
| `alias` | string | Yes |
| `banVoPath` | string | Yes |
| `chooseVoPath` | string | Yes |
| `id` | int32 | Yes |
| `name` | string | Yes |
| `roles` | string[] | Yes |
| `squarePortraitPath` | string | Yes |
| `stingerSfxPath` | string | Yes |
| `passive` | LolNpeTutorialPathCollectionsChampionSpell | Yes |
| `spells` | LolNpeTutorialPathCollectionsChampionSpell[] | Yes |

### LolNpeTutorialPathCollectionsChampionSpell

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `description` | string | Yes |

### LolNpeTutorialPathTutorial

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `stepNumber` | int32 | Yes |
| `title` | string | Yes |
| `description` | string | Yes |
| `backgroundUrl` | string | Yes |
| `queueId` | string | Yes |
| `useQuickSearchMatchmaking` | boolean | Yes |
| `useChosenChampion` | boolean | Yes |
| `status` | LolNpeTutorialPathTutorialStatus | Yes |
| `isViewed` | boolean | Yes |
| `type` | LolNpeTutorialPathTutorialType | Yes |
| `rewards` | LolNpeTutorialPathTutorialReward[] | Yes |

### LolNpeTutorialPathTutorialReward

| Field | Type | Required |
|-------|------|----------|
| `rewardType` | string | Yes |
| `description` | string | Yes |
| `rewardFulfilled` | boolean | Yes |
| `iconUrl` | string | Yes |
| `itemId` | string | Yes |
| `sequence` | int32 | Yes |
| `uniqueName` | string | Yes |

### LolNpeTutorialPathTutorialStatus

**Type**: enum string

**Values**: `COMPLETED`, `UNLOCKED`, `LOCKED`


### LolNpeTutorialPathTutorialType

**Type**: enum string

**Values**: `REWARD`, `CARD`


---

## LolObjectives


### LolObjectivesUIObjectivesCategory

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `startDate` | uint64 | Yes |
| `progressEndDate` | uint64 | Yes |
| `endDate` | uint64 | Yes |
| `categorySectionImage` | string | Yes |
| `categoryName` | string | Yes |
| `overrideBackgroundImage` | string | Yes |
| `objectives` | LolObjectivesUIObjectives[] | Yes |
| `categoryType` | LolObjectivesObjectiveCategoryType | Yes |
| `tftPassType` | LolObjectivesTftPassType | Yes |
| `lolEventHubType` | LolObjectivesEventHubType | Yes |
| `objectiveCategoryFilter` | LolObjectivesObjectiveCategoryFilter | Yes |

### LolObjectivesUIObjectivesGroup

| Field | Type | Required |
|-------|------|----------|
| `gameType` | string | Yes |
| `objectivesCategories` | LolObjectivesUIObjectivesCategory[] | Yes |

---

## LolPatch


### LolPatchChunkingPatcherEnvironment

| Field | Type | Required |
|-------|------|----------|
| `game_patcher_available` | boolean | Yes |
| `game_patcher_enabled` | boolean | Yes |

### LolPatchComponentState

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `action` | LolPatchComponentStateAction | Yes |
| `isUpToDate` | boolean | Yes |
| `isUpdateAvailable` | boolean | Yes |
| `timeOfLastUpToDateCheckISO8601` | string |  |
| `isCorrupted` | boolean | Yes |
| `progress` | LolPatchComponentActionProgress |  |

### LolPatchComponentStateAction

**Type**: enum string

**Values**: `Migrating`, `Repairing`, `Patching`, `CheckingForUpdates`, `Idle`


### LolPatchInstallPaths

| Field | Type | Required |
|-------|------|----------|
| `gameInstallRoot` | string | Yes |
| `gameExecutablePath` | string | Yes |

### LolPatchNotification

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `notificationId` | LolPatchNotificationId | Yes |
| `data` | object (map) | Yes |

### LolPatchNotificationId

**Type**: enum string

**Values**: `BrokenPermissions`, `NotEnoughDiskSpace`, `DidRestoreClientBackup`, `FailedToWriteError`, `MissingFilesError`, `ConnectionError`, `UnspecifiedError`


### LolPatchProductState

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `action` | LolPatchComponentStateAction | Yes |
| `isUpToDate` | boolean | Yes |
| `isUpdateAvailable` | boolean | Yes |
| `isCorrupted` | boolean | Yes |
| `isStopped` | boolean | Yes |
| `percentPatched` | number | Yes |
| `components` | LolPatchComponentState[] | Yes |

### LolPatchStatus

| Field | Type | Required |
|-------|------|----------|
| `connectedToPatchServer` | boolean | Yes |

### LolPatchSupportedGameRelease

| Field | Type | Required |
|-------|------|----------|
| `artifact_id` | string | Yes |
| `download` | LolPatchPatchSieveDownload | Yes |
| `selected` | boolean | Yes |

### LolPatchSupportedGameReleases

| Field | Type | Required |
|-------|------|----------|
| `supported_game_releases` | LolPatchSupportedGameRelease[] | Yes |

### LolPatchUxResource

| Field | Type | Required |
|-------|------|----------|
| `visible` | boolean | Yes |

---

## LolPerks


### LolPerksNamecheckResponse

| Field | Type | Required |
|-------|------|----------|
| `errors` | string[] | Yes |

### LolPerksPerkPageResource

| Field | Type | Required |
|-------|------|----------|
| `current` | boolean | Yes |
| `id` | int32 | Yes |
| `isActive` | boolean | Yes |
| `isValid` | boolean | Yes |
| `isEditable` | boolean | Yes |
| `isDeletable` | boolean | Yes |
| `isTemporary` | boolean | Yes |
| `name` | string | Yes |
| `order` | int32 | Yes |
| `primaryStyleId` | int32 | Yes |
| `selectedPerkIds` | integer[] | Yes |
| `subStyleId` | int32 | Yes |
| `autoModifiedSelections` | integer[] | Yes |
| `lastModified` | uint64 | Yes |
| `runeRecommendationId` | string | Yes |
| `recommendationIndex` | int32 | Yes |
| `isRecommendationOverride` | boolean | Yes |
| `recommendationChampionId` | int32 | Yes |
| `quickPlayChampionIds` | integer[] | Yes |
| `primaryStyleName` | string | Yes |
| `secondaryStyleName` | string | Yes |
| `primaryStyleIconPath` | string | Yes |
| `secondaryStyleIconPath` | string | Yes |
| `tooltipBgPath` | string | Yes |
| `pageKeystone` | LolPerksUIPerkMinimal | Yes |
| `uiPerks` | LolPerksUIPerkMinimal[] | Yes |

### LolPerksPerkSubStyleBonusResource

| Field | Type | Required |
|-------|------|----------|
| `perkId` | int32 | Yes |
| `styleId` | int32 | Yes |

### LolPerksPerkUIPerk

| Field | Type | Required |
|-------|------|----------|
| `iconPath` | string | Yes |
| `id` | int32 | Yes |
| `styleId` | int32 | Yes |
| `styleIdName` | string | Yes |
| `longDesc` | string | Yes |
| `name` | string | Yes |
| `shortDesc` | string | Yes |
| `tooltip` | string | Yes |
| `recommendationDescriptor` | string | Yes |
| `slotType` | string | Yes |

### LolPerksPerkUIRecommendedPage

| Field | Type | Required |
|-------|------|----------|
| `position` | string | Yes |
| `isDefaultPosition` | boolean | Yes |
| `keystone` | LolPerksPerkUIPerk | Yes |
| `perks` | LolPerksPerkUIPerk[] | Yes |
| `primaryPerkStyleId` | int32 | Yes |
| `secondaryPerkStyleId` | int32 | Yes |
| `primaryRecommendationAttribute` | string | Yes |
| `secondaryRecommendationAttribute` | string | Yes |
| `summonerSpellIds` | integer[] | Yes |
| `recommendationId` | string | Yes |
| `isRecommendationOverride` | boolean | Yes |
| `recommendationChampionId` | int32 | Yes |

### LolPerksPerkUISlot

| Field | Type | Required |
|-------|------|----------|
| `perks` | integer[] | Yes |
| `type` | string | Yes |
| `slotLabel` | string | Yes |

### LolPerksPerkUIStyle

| Field | Type | Required |
|-------|------|----------|
| `allowedSubStyles` | integer[] | Yes |
| `iconPath` | string | Yes |
| `assetMap` | object (map) | Yes |
| `id` | int32 | Yes |
| `name` | string | Yes |
| `slots` | LolPerksPerkUISlot[] | Yes |
| `subStyleBonus` | LolPerksPerkSubStyleBonusResource[] | Yes |
| `tooltip` | string | Yes |
| `defaultSubStyle` | int32 | Yes |
| `defaultPerks` | integer[] | Yes |
| `defaultPageName` | string | Yes |
| `idName` | string | Yes |

### LolPerksPlayerInventory

| Field | Type | Required |
|-------|------|----------|
| `ownedPageCount` | uint32 | Yes |
| `customPageCount` | uint32 | Yes |
| `canAddCustomPage` | boolean | Yes |
| `isCustomPageCreationUnlocked` | boolean | Yes |

### LolPerksUIPerkMinimal

| Field | Type | Required |
|-------|------|----------|
| `id` | int32 | Yes |
| `styleId` | int32 | Yes |
| `name` | string | Yes |
| `slotType` | string | Yes |
| `iconPath` | string | Yes |

### LolPerksUISettings

| Field | Type | Required |
|-------|------|----------|
| `showLongDescriptions` | boolean | Yes |
| `gridModeEnabled` | boolean | Yes |
| `showPresetPages` | boolean | Yes |
| `gameplayPatchVersionSeen` | string | Yes |
| `gameplayUpdatedPerksSeen` | integer[] | Yes |

### LolPerksUpdatePageOrderRequest

| Field | Type | Required |
|-------|------|----------|
| `targetPageId` | int32 | Yes |
| `destinationPageId` | int32 | Yes |
| `offset` | int32 | Yes |

### LolPerksValidateItemSetNameResponse

| Field | Type | Required |
|-------|------|----------|
| `success` | boolean | Yes |
| `nameCheckResponse` | LolPerksNamecheckResponse | Yes |

### LolPerksValidatePageNameData

| Field | Type | Required |
|-------|------|----------|
| `id` | int32 | Yes |
| `name` | string | Yes |

---

## LolPft


### LolPftPFTEvent

| Field | Type | Required |
|-------|------|----------|
| `playerSurveyId` | uint64 | Yes |
| `action` | string | Yes |
| `data` | object[] | Yes |

### LolPftPFTSurvey

| Field | Type | Required |
|-------|------|----------|
| `id` | uint64 | Yes |
| `title` | string | Yes |
| `caption` | string | Yes |
| `type` | string | Yes |
| `display` | string | Yes |
| `data` | object (map) | Yes |

---

## LolPlayer


### LolPlayerBehaviorBanNotification

| Field | Type | Required |
|-------|------|----------|
| `id` | uint64 | Yes |
| `source` | LolPlayerBehaviorNotificationSource | Yes |
| `reason` | string | Yes |
| `timeUntilBanExpires` | uint64 | Yes |
| `isPermaBan` | boolean | Yes |
| `displayReformCard` | boolean | Yes |

### LolPlayerBehaviorCodeOfConductNotification

| Field | Type | Required |
|-------|------|----------|
| `message` | string | Yes |

### LolPlayerBehaviorNotificationSource

**Type**: enum string

**Values**: `Message`, `ForcedShutdown`, `Login`, `Invalid`


### LolPlayerBehaviorPlayerBehaviorConfig

| Field | Type | Required |
|-------|------|----------|
| `IsLoaded` | boolean | Yes |
| `CodeOfConductEnabled` | boolean | Yes |

### LolPlayerBehaviorReformCard

| Field | Type | Required |
|-------|------|----------|
| `id` | uint64 | Yes |
| `punishmentType` | string | Yes |
| `reason` | string | Yes |
| `timeWhenPunishmentExpires` | uint64 | Yes |
| `punishmentLengthTime` | uint64 | Yes |
| `punishmentLengthGames` | int64 | Yes |
| `restrictedChatGamesRemaining` | int64 | Yes |
| `chatLogs` | string[] | Yes |
| `gameIds` | integer[] | Yes |
| `playerFacingMessage` | string | Yes |

### LolPlayerBehaviorReformCardChatLogs

| Field | Type | Required |
|-------|------|----------|
| `preGameChatLogs` | string[] | Yes |
| `inGameChatLogs` | string[] | Yes |
| `postGameChatLogs` | string[] | Yes |

### LolPlayerBehaviorReformCardV2

| Field | Type | Required |
|-------|------|----------|
| `id` | uint64 | Yes |
| `punishmentType` | string | Yes |
| `punishmentReason` | string | Yes |
| `punishedUntilDateMillis` | uint64 | Yes |
| `punishmentLengthMillis` | uint64 | Yes |
| `punishmentLengthGames` | int64 | Yes |
| `punishedForReformCardChatLogs` | LolPlayerBehaviorReformCardChatLogs[] | Yes |
| `punishedForGameIds` | integer[] | Yes |
| `playerFacingMessage` | string | Yes |

### LolPlayerBehaviorReporterFeedback

| Field | Type | Required |
|-------|------|----------|
| `id` | uint64 | Yes |
| `accountId` | uint64 | Yes |
| `messageId` | string | Yes |
| `type` | string | Yes |

### LolPlayerBehaviorReporterFeedbackMessage

| Field | Type | Required |
|-------|------|----------|
| `title` | string | Yes |
| `message` | string | Yes |
| `category` | string | Yes |
| `locale` | string | Yes |
| `key` | string | Yes |

### LolPlayerBehaviorRestrictionNotification

| Field | Type | Required |
|-------|------|----------|
| `id` | uint64 | Yes |
| `source` | LolPlayerBehaviorNotificationSource | Yes |
| `gamesRemaining` | int64 | Yes |
| `expirationMillis` | uint64 | Yes |
| `displayReformCard` | boolean | Yes |

### LolPlayerMessagingDynamicCelebrationMessagingNotificationResource

| Field | Type | Required |
|-------|------|----------|
| `id` | int32 | Yes |
| `accountId` | uint64 | Yes |
| `msgId` | string | Yes |
| `celebrationTitle` | string | Yes |
| `celebrationBody` | string | Yes |
| `celebrationMessage` | string | Yes |
| `inventoryType` | string | Yes |
| `itemId` | string | Yes |
| `itemQuantity` | string | Yes |
| `celebrationType` | string | Yes |
| `status` | int32 | Yes |

### LolPlayerMessagingPlayerMessagingNotificationResource

| Field | Type | Required |
|-------|------|----------|
| `id` | int32 | Yes |
| `accountId` | uint64 | Yes |
| `msgId` | string | Yes |
| `title` | string | Yes |
| `body` | string | Yes |
| `status` | int32 | Yes |

### LolPlayerPreferencesPlayerPreferences

| Field | Type | Required |
|-------|------|----------|
| `type` | string | Yes |
| `version` | string | Yes |
| `data` | string | Yes |

### LolPlayerPreferencesPlayerPreferencesEndpoint

| Field | Type | Required |
|-------|------|----------|
| `Enabled` | boolean | Yes |
| `EnforceHttps` | boolean | Yes |
| `ServiceEndpoint` | string | Yes |
| `Version` | string | Yes |
| `Retries` | int64 | Yes |

### LolPlayerReportSenderPlayerReport

| Field | Type | Required |
|-------|------|----------|
| `offenderPuuid` | string | Yes |
| `obfuscatedOffenderPuuid` | string | Yes |
| `categories` | string[] | Yes |
| `gameId` | uint64 | Yes |
| `comment` | string | Yes |

---

## LolPre


### LolPreEndOfGameSequenceEvent

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `priority` | int32 | Yes |

---

## LolPremade


### LolPremadeVoiceAudioPropertiesResource

| Field | Type | Required |
|-------|------|----------|
| `isLoopbackEnabled` | boolean | Yes |
| `micEnergy` | uint32 | Yes |

### LolPremadeVoiceDeviceResource

| Field | Type | Required |
|-------|------|----------|
| `handle` | string | Yes |
| `name` | string | Yes |
| `usable` | boolean | Yes |
| `is_current_device` | boolean | Yes |
| `is_default` | boolean | Yes |

### LolPremadeVoiceFirstExperience

| Field | Type | Required |
|-------|------|----------|
| `showFirstExperienceInLCU` | boolean | Yes |
| `showFirstExperienceInGame` | boolean | Yes |

### LolPremadeVoiceInputMode

**Type**: enum string

**Values**: `pushToTalk`, `voiceActivity`


### LolPremadeVoicePremadeVoiceParticipantDto

| Field | Type | Required |
|-------|------|----------|
| `participantId` | string | Yes |
| `summonerId` | uint64 | Yes |
| `puuid` | string | Yes |
| `displayName` | string | Yes |
| `volume` | uint32 | Yes |
| `energy` | uint32 | Yes |
| `isMuted` | boolean | Yes |
| `isSpeaking` | boolean | Yes |

### LolPremadeVoiceSettingsResource

| Field | Type | Required |
|-------|------|----------|
| `currentCaptureDeviceHandle` | string | Yes |
| `vadHangoverTime` | uint32 | Yes |
| `vadSensitivity` | uint32 | Yes |
| `micLevel` | uint32 | Yes |
| `localMicMuted` | boolean | Yes |
| `loopbackEnabled` | boolean | Yes |
| `autoJoin` | boolean | Yes |
| `muteOnConnect` | boolean | Yes |
| `vadActive` | boolean | Yes |
| `pttActive` | boolean | Yes |
| `inputMode` | LolPremadeVoiceInputMode | Yes |
| `pttKey` | string |  |

### LolPremadeVoiceVoiceAvailability

| Field | Type | Required |
|-------|------|----------|
| `enabled` | boolean | Yes |
| `connectedToVoiceServer` | boolean | Yes |
| `voiceChannelAvailable` | boolean | Yes |
| `disabledAfterLogin` | boolean | Yes |
| `showUI` | boolean | Yes |
| `showDisconnectedState` | boolean | Yes |

---

## LolProgression


### LolProgressionCounter

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `name` | string | Yes |
| `groupId` | string | Yes |
| `direction` | string | Yes |
| `startValue` | int64 | Yes |

### LolProgressionCounterInstance

| Field | Type | Required |
|-------|------|----------|
| `ownerId` | string | Yes |
| `productId` | string | Yes |
| `groupId` | string | Yes |
| `counterId` | string | Yes |
| `counterValue` | int64 | Yes |

### LolProgressionEntityInstance

| Field | Type | Required |
|-------|------|----------|
| `schemaVersion` | string | Yes |
| `groupId` | string | Yes |
| `counters` | LolProgressionCounterInstance[] | Yes |
| `milestones` | LolProgressionMilestoneInstance[] | Yes |

### LolProgressionGroup

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `productId` | string | Yes |
| `name` | string | Yes |
| `repeat` | LolProgressionRepeat | Yes |
| `counters` | LolProgressionCounter[] | Yes |
| `milestones` | LolProgressionMilestone[] | Yes |

### LolProgressionMilestone

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `name` | string | Yes |
| `groupId` | string | Yes |
| `counterId` | string | Yes |
| `triggerValue` | int64 | Yes |
| `properties` | object (map) | Yes |

### LolProgressionMilestoneInstance

| Field | Type | Required |
|-------|------|----------|
| `milestoneId` | string | Yes |
| `instanceId` | string | Yes |
| `ownerId` | string | Yes |
| `productId` | string | Yes |
| `groupId` | string | Yes |
| `counterId` | string | Yes |
| `triggerValue` | int64 | Yes |
| `repeatSequence` | uint32 | Yes |
| `triggered` | boolean | Yes |
| `triggeredTimestamp` | string | Yes |
| `triggers` | LolProgressionTrigger[] | Yes |

### LolProgressionRepeat

| Field | Type | Required |
|-------|------|----------|
| `count` | int32 | Yes |
| `scope` | uint32 | Yes |
| `multiplier` | number | Yes |
| `milestones` | LolProgressionMilestone[] | Yes |
| `repeatTriggers` | LolProgressionRepeatGroupTrigger[] | Yes |

---

## LolPublishing


### LolPublishingContentClientData

| Field | Type | Required |
|-------|------|----------|
| `puuid` | string | Yes |
| `account_id` | uint64 | Yes |
| `env` | string | Yes |
| `web_region` | string | Yes |
| `locale` | string | Yes |
| `summoner_level` | uint16 | Yes |
| `summoner_name` | string | Yes |
| `app_name` | string | Yes |
| `app_version` | string | Yes |
| `system_os` | string | Yes |
| `protocol` | string | Yes |
| `port` | uint16 | Yes |
| `assetUrls` | object (map) | Yes |

### LolPublishingContentPubHubConfig

| Field | Type | Required |
|-------|------|----------|
| `edge` | LolPublishingContentPubHubConfigEdge | Yes |
| `appContext` | LolPublishingContentPubHubConfigAppContext | Yes |

### LolPublishingContentPubHubConfigAppContext

| Field | Type | Required |
|-------|------|----------|
| `userId` | string | Yes |
| `userRegion` | string | Yes |
| `deviceCategory` | string | Yes |
| `deviceOperatingSystem` | string | Yes |
| `deviceOperatingSystemVersion` | string | Yes |
| `appId` | string | Yes |
| `appVersion` | string | Yes |
| `appLocale` | string | Yes |
| `appLanguage` | string | Yes |
| `publishingLocale` | string | Yes |
| `appSessionId` | string | Yes |

### LolPublishingContentPubHubConfigEdge

| Field | Type | Required |
|-------|------|----------|
| `clientId` | string | Yes |
| `clientRegion` | string | Yes |

### LolPublishingContentPublishingSettings

| Field | Type | Required |
|-------|------|----------|
| `region` | string | Yes |
| `locale` | string | Yes |
| `webRegion` | string | Yes |
| `webLocale` | string | Yes |
| `publishingLocale` | string | Yes |
| `rsoPlatformId` | string | Yes |

---

## LolPurchase


### LolPurchaseWidgetBaseSkinLineDto

| Field | Type | Required |
|-------|------|----------|
| `items` | LolPurchaseWidgetSkinLineItemDto[] | Yes |
| `localizedName` | string | Yes |
| `skinLineDescriptions` | LolPurchaseWidgetSkinLineDescriptionDto[] | Yes |
| `pricingOptions` | LolPurchaseWidgetPriceOptionDto[] | Yes |
| `splashPath` | string | Yes |
| `uncenteredSplashPath` | string | Yes |
| `collectionCardPath` | string | Yes |
| `collectionDescription` | string | Yes |
| `tilePath` | string | Yes |

### LolPurchaseWidgetCapOrdersOrderDto

| Field | Type | Required |
|-------|------|----------|
| `data` | LolPurchaseWidgetCapOrdersDataDto | Yes |
| `meta` | LolPurchaseWidgetCapOrdersMetaDto | Yes |

### LolPurchaseWidgetCatalogPluginItem

| Field | Type | Required |
|-------|------|----------|
| `itemId` | int32 | Yes |
| `itemInstanceId` | string | Yes |
| `owned` | boolean | Yes |
| `inventoryType` | string | Yes |
| `subInventoryType` | string | Yes |
| `name` | string | Yes |
| `subTitle` | string | Yes |
| `description` | string | Yes |
| `imagePath` | string | Yes |
| `purchaseDate` | uint64 | Yes |
| `releaseDate` | uint64 | Yes |
| `inactiveDate` | uint64 | Yes |
| `maxQuantity` | int32 | Yes |
| `prices` | LolPurchaseWidgetCatalogPluginPrice[] | Yes |
| `tags` | string[] |  |
| `metadata` | LolPurchaseWidgetItemMetadataEntry[] |  |
| `questSkinInfo` | LolPurchaseWidgetSkinLineInfo |  |
| `active` | boolean | Yes |
| `sale` | LolPurchaseWidgetSale |  |
| `ownershipType` | LolPurchaseWidgetInventoryOwnership |  |

### LolPurchaseWidgetCatalogPluginItemAssets

| Field | Type | Required |
|-------|------|----------|
| `splashPath` | string | Yes |
| `iconPath` | string | Yes |
| `tilePath` | string | Yes |
| `previewVideoUrl` | string | Yes |
| `emblems` | LolPurchaseWidgetChampionSkinEmblem[] | Yes |
| `colors` | string[] | Yes |

### LolPurchaseWidgetCatalogPluginItemWithDetails

| Field | Type | Required |
|-------|------|----------|
| `item` | LolPurchaseWidgetCatalogPluginItem | Yes |
| `quantity` | uint32 | Yes |
| `requiredItems` | LolPurchaseWidgetCatalogPluginItemWithDetails[] |  |
| `bundledItems` | LolPurchaseWidgetCatalogPluginItemWithDetails[] |  |
| `minimumBundlePrices` | LolPurchaseWidgetCatalogPluginPrice[] |  |
| `bundledDiscountPrices` | LolPurchaseWidgetCatalogPluginPrice[] |  |
| `assets` | LolPurchaseWidgetCatalogPluginItemAssets | Yes |
| `bundleFinalPrice` | int32 | Yes |
| `flexible` | boolean | Yes |

### LolPurchaseWidgetCatalogPluginPrice

| Field | Type | Required |
|-------|------|----------|
| `currency` | string | Yes |
| `cost` | int32 | Yes |
| `costType` | string |  |
| `sale` | LolPurchaseWidgetCatalogPluginSale |  |

### LolPurchaseWidgetItemChoiceDetails

| Field | Type | Required |
|-------|------|----------|
| `item` | LolPurchaseWidgetCatalogPluginItem | Yes |
| `backgroundImage` | string | Yes |
| `contents` | LolPurchaseWidgetItemDetails[] | Yes |
| `discount` | string | Yes |
| `fullPrice` | uint32 | Yes |
| `displayType` | string | Yes |
| `purchaseOptions` | LolPurchaseWidgetPurchaseOption[] | Yes |

### LolPurchaseWidgetItemChoices

| Field | Type | Required |
|-------|------|----------|
| `choices` | LolPurchaseWidgetItemChoiceDetails[] | Yes |
| `validationErrors` | LolPurchaseWidgetValidationErrorEntry[] | Yes |

### LolPurchaseWidgetItemDefinition

| Field | Type | Required |
|-------|------|----------|
| `itemId` | int32 | Yes |
| `inventoryType` | string | Yes |
| `subInventoryType` | string | Yes |
| `name` | string | Yes |
| `description` | string | Yes |
| `subTitle` | string | Yes |
| `imagePath` | string | Yes |
| `owned` | boolean | Yes |
| `assets` | LolPurchaseWidgetCatalogPluginItemAssets | Yes |
| `tags` | string[] | Yes |
| `metadata` | LolPurchaseWidgetItemMetadataEntry[] | Yes |
| `bundledItemPrice` | LolPurchaseWidgetBundledItemPricingInfo |  |
| `loyaltyUnlocked` | boolean | Yes |
| `hasVisibleLootOdds` | boolean | Yes |
| `inactiveDate` | uint64 | Yes |

### LolPurchaseWidgetItemSale

| Field | Type | Required |
|-------|------|----------|
| `startDate` | string | Yes |
| `endDate` | string | Yes |
| `discount` | number |  |

### LolPurchaseWidgetOrderNotificationResource

| Field | Type | Required |
|-------|------|----------|
| `eventTypeId` | string | Yes |
| `eventType` | string | Yes |
| `status` | string | Yes |

### LolPurchaseWidgetPriceOptionDto

| Field | Type | Required |
|-------|------|----------|
| `price` | int64 | Yes |
| `currencyType` | string | Yes |
| `currencyPaymentOption` | string |  |
| `currencyName` | string |  |
| `currencyImagePath` | string |  |

### LolPurchaseWidgetPurchasableItem

| Field | Type | Required |
|-------|------|----------|
| `item` | LolPurchaseWidgetItemDefinition | Yes |
| `dependencies` | LolPurchaseWidgetItemDefinition[] | Yes |
| `bundledItems` | LolPurchaseWidgetItemDefinition[] | Yes |
| `sale` | LolPurchaseWidgetItemSale |  |
| `purchaseOptions` | LolPurchaseWidgetPurchaseOption[] | Yes |
| `validationErrors` | LolPurchaseWidgetValidationErrorEntry[] | Yes |

### LolPurchaseWidgetPurchaseItem

| Field | Type | Required |
|-------|------|----------|
| `itemKey` | LolPurchaseWidgetItemKey | Yes |
| `quantity` | int32 | Yes |
| `source` | string | Yes |
| `purchaseCurrencyInfo` | LolPurchaseWidgetItemPrice | Yes |

### LolPurchaseWidgetPurchaseOfferOrderStatuses

| Field | Type | Required |
|-------|------|----------|
| `statuses` | object (map) | Yes |

### LolPurchaseWidgetPurchaseOfferRequestV3

| Field | Type | Required |
|-------|------|----------|
| `offerId` | string | Yes |
| `currencyType` | string | Yes |
| `quantity` | uint32 | Yes |
| `price` | uint32 | Yes |

### LolPurchaseWidgetPurchaseOfferResponseV3

| Field | Type | Required |
|-------|------|----------|
| `legacy` | boolean | Yes |
| `orderDto` | LolPurchaseWidgetCapOrdersOrderDto |  |

### LolPurchaseWidgetPurchaseOption

| Field | Type | Required |
|-------|------|----------|
| `priceDetails` | LolPurchaseWidgetPriceDetail[] | Yes |

### LolPurchaseWidgetPurchaseRequest

| Field | Type | Required |
|-------|------|----------|
| `items` | LolPurchaseWidgetPurchaseItem[] | Yes |

### LolPurchaseWidgetPurchaseResponse

| Field | Type | Required |
|-------|------|----------|
| `items` | LolPurchaseWidgetPurchaseItem[] | Yes |
| `transactions` | LolPurchaseWidgetTransaction[] | Yes |
| `useRMSConfirmation` | boolean | Yes |

### LolPurchaseWidgetPurchaseWidgetConfig

| Field | Type | Required |
|-------|------|----------|
| `enabled` | boolean | Yes |
| `nonRefundableDisclaimerEnabled` | boolean | Yes |
| `alwaysShowPurchaseDisclaimer` | boolean | Yes |

### LolPurchaseWidgetSkinLineDescriptionDto

| Field | Type | Required |
|-------|------|----------|
| `title` | string | Yes |
| `description` | string | Yes |
| `iconImagePath` | string | Yes |

### LolPurchaseWidgetSkinLineItemDto

| Field | Type | Required |
|-------|------|----------|
| `thumbnailImagePath` | string | Yes |
| `largeImagePath` | string |  |
| `localizedLongName` | string | Yes |
| `localizedShortName` | string | Yes |
| `largeVideoPath` | string |  |

### LolPurchaseWidgetTransaction

| Field | Type | Required |
|-------|------|----------|
| `transactionId` | string | Yes |
| `itemKey` | LolPurchaseWidgetItemKey | Yes |
| `itemName` | string | Yes |
| `iconUrl` | string | Yes |

### LolPurchaseWidgetValidateOfferError

| Field | Type | Required |
|-------|------|----------|
| `errorKey` | string | Yes |
| `meta` | string | Yes |

### LolPurchaseWidgetValidateOfferRequestV3

| Field | Type | Required |
|-------|------|----------|
| `offerId` | string | Yes |

### LolPurchaseWidgetValidateOfferResponseV3

| Field | Type | Required |
|-------|------|----------|
| `validationErrors` | LolPurchaseWidgetValidateOfferError[] | Yes |

### LolPurchaseWidgetValidationErrorEntry

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |

---

## LolRanked


### LolRankedAchievedTier

| Field | Type | Required |
|-------|------|----------|
| `queueType` | LolRankedLeagueQueueType | Yes |
| `tier` | string | Yes |
| `division` | uint64 | Yes |

### LolRankedEosNotificationResource

| Field | Type | Required |
|-------|------|----------|
| `notificationName` | string | Yes |
| `notificationType` | string | Yes |
| `seasonEndTime` | int64 | Yes |
| `queue` | string | Yes |
| `tier` | string | Yes |
| `division` | string | Yes |

### LolRankedGlobalNotification

| Field | Type | Required |
|-------|------|----------|
| `notifyReason` | string | Yes |
| `participantId` | string | Yes |
| `queueType` | string | Yes |
| `tier` | string | Yes |
| `position` | int32 | Yes |

### LolRankedLcuLeagueNotification

| Field | Type | Required |
|-------|------|----------|
| `id` | uint64 | Yes |
| `msgId` | string | Yes |
| `displayType` | LolRankedNotificationDisplayType | Yes |
| `notifyReason` | string | Yes |
| `changeReason` | string | Yes |
| `queueType` | LolRankedLeagueQueueType | Yes |
| `gameId` | uint64 | Yes |
| `provisionalGamesRemaining` | int32 | Yes |
| `tier` | string | Yes |
| `division` | LolRankedLeagueDivision | Yes |
| `numberOfPromotions` | uint64 | Yes |
| `miniseriesProgress` | string | Yes |
| `leaguePoints` | int32 | Yes |
| `leaguePointsDelta` | int32 | Yes |
| `ratedTier` | LolRankedRatedTier | Yes |
| `ratedRating` | int32 | Yes |
| `ratedRatingDelta` | int32 | Yes |
| `eligibleForPromoHelper` | boolean | Yes |
| `miniseriesWins` | int32 | Yes |
| `timeUntilInactivityStatusChanges` | int64 | Yes |
| `rewardEarnedId` | string | Yes |
| `rewardEarnedType` | string | Yes |
| `rewardEarnedTitle` | string | Yes |
| `rewardEarnedDescription` | string | Yes |
| `rewardEarnedTier` | string | Yes |
| `rewardOverrideImagePath` | string | Yes |
| `promoSeriesForRanksEnabled` | boolean | Yes |
| `consolationLpUsed` | int32 | Yes |
| `afkLpPenaltyAmount` | int32 | Yes |
| `afkLpPenaltyLevel` | int32 | Yes |
| `wasAfkOrLeaver` | boolean | Yes |
| `canDemoteFromTier` | boolean | Yes |
| `winStreak` | int32 | Yes |
| `wins` | int32 | Yes |
| `losses` | int32 | Yes |
| `lpBonusAppliedReason` | string | Yes |
| `leaguePointDeltaBreakdown` | object (map) | Yes |

### LolRankedLeagueDivision

**Type**: enum string

**Values**: `NA`, `V`, `IV`, `III`, `II`, `I`


### LolRankedLeagueDivisionInfo

| Field | Type | Required |
|-------|------|----------|
| `tier` | string | Yes |
| `division` | LolRankedLeagueDivision | Yes |
| `maxLeagueSize` | int32 | Yes |
| `apexUnlockTimeMillis` | int64 | Yes |
| `minLpForApexTier` | int32 | Yes |
| `topNumberOfPlayers` | int64 | Yes |
| `standings` | LolRankedLeagueStanding[] | Yes |

### LolRankedLeagueLadderInfo

| Field | Type | Required |
|-------|------|----------|
| `queueType` | LolRankedLeagueQueueType | Yes |
| `tier` | string | Yes |
| `provisionalGameThreshold` | int32 | Yes |
| `divisions` | LolRankedLeagueDivisionInfo[] | Yes |
| `nextApexUpdateMillis` | int64 | Yes |
| `requestedRankedEntry` | LolRankedLeagueStanding |  |

### LolRankedLeagueQueueType

**Type**: enum string

**Values**: `RANKED_TFT_DOUBLE_UP`, `RANKED_TFT_PAIRS`, `RANKED_TFT_TURBO`, `RANKED_TFT`, `RANKED_FLEX_TT`, `CHERRY`, `RANKED_FLEX_SR`, `RANKED_SOLO_5x5`, `NONE`


### LolRankedLeagueStanding

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `puuid` | string | Yes |
| `summonerName` | string | Yes |
| `position` | int32 | Yes |
| `positionDelta` | int32 | Yes |
| `previousPosition` | int32 | Yes |
| `tier` | string | Yes |
| `division` | LolRankedLeagueDivision | Yes |
| `leaguePoints` | int64 | Yes |
| `miniseriesResults` | LolRankedMiniseries[] | Yes |
| `wins` | uint64 | Yes |
| `losses` | uint64 | Yes |
| `provisionalGamesRemaining` | int32 | Yes |
| `isProvisional` | boolean | Yes |
| `previousSeasonEndTier` | string | Yes |
| `previousSeasonEndDivision` | LolRankedLeagueDivision | Yes |
| `earnedRegaliaRewardIds` | string[] | Yes |
| `rankedRegaliaLevel` | int32 | Yes |
| `pendingPromotion` | boolean | Yes |
| `pendingDemotion` | boolean | Yes |

### LolRankedLolEosRewardGameData

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `name` | string | Yes |
| `title` | string | Yes |
| `description` | string | Yes |
| `tier` | LolRankedLolEosRewardTier | Yes |
| `type` | LolRankedLolEosRewardType | Yes |
| `overrideImagePath` | string | Yes |

### LolRankedLolEosRewardGroupGameData

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `rewardNames` | string[] | Yes |

### LolRankedLolEosRewardsConfigGameData

| Field | Type | Required |
|-------|------|----------|
| `seasonId` | int64 | Yes |
| `rewards` | LolRankedLolEosRewardGameData[] | Yes |
| `rewardGroups` | LolRankedLolEosRewardGroupGameData[] | Yes |

### LolRankedNotificationDisplayType

**Type**: enum string

**Values**: `VIGNETTE`, `MODAL`, `TOAST`, `NONE`


### LolRankedParticipantTiers

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `achievedTiers` | LolRankedAchievedTier[] | Yes |

### LolRankedRankedQueueStats

| Field | Type | Required |
|-------|------|----------|
| `queueType` | LolRankedLeagueQueueType | Yes |
| `provisionalGameThreshold` | int32 | Yes |
| `provisionalGamesRemaining` | int32 | Yes |
| `isProvisional` | boolean | Yes |
| `tier` | string | Yes |
| `division` | LolRankedLeagueDivision | Yes |
| `leaguePoints` | int32 | Yes |
| `miniSeriesProgress` | string | Yes |
| `ratedTier` | LolRankedRatedTier | Yes |
| `ratedRating` | int32 | Yes |
| `wins` | int32 | Yes |
| `losses` | int32 | Yes |
| `currentSeasonWinsForRewards` | int32 | Yes |
| `previousSeasonWinsForRewards` | int32 | Yes |
| `highestTier` | string | Yes |
| `highestDivision` | LolRankedLeagueDivision | Yes |
| `previousSeasonEndTier` | string | Yes |
| `previousSeasonEndDivision` | LolRankedLeagueDivision | Yes |
| `previousSeasonHighestTier` | string | Yes |
| `previousSeasonHighestDivision` | LolRankedLeagueDivision | Yes |
| `climbingIndicatorActive` | boolean | Yes |
| `warnings` | LolRankedRankedQueueWarnings |  |

### LolRankedRankedQueueStatsDTO

| Field | Type | Required |
|-------|------|----------|
| `queueType` | string | Yes |
| `provisionalGameThreshold` | int32 | Yes |
| `provisionalGamesRemaining` | int32 | Yes |
| `tier` | string | Yes |
| `rank` | string | Yes |
| `leaguePoints` | int32 | Yes |
| `miniSeriesProgress` | string | Yes |
| `ratedTier` | string | Yes |
| `ratedRating` | int32 | Yes |
| `wins` | int32 | Yes |
| `losses` | int32 | Yes |
| `currentSeasonWinsForRewards` | int32 | Yes |
| `previousSeasonWinsForRewards` | int32 | Yes |
| `highestTier` | string | Yes |
| `highestRank` | string | Yes |
| `previousSeasonEndTier` | string | Yes |
| `previousSeasonEndRank` | string | Yes |
| `previousSeasonHighestTier` | string | Yes |
| `previousSeasonHighestRank` | string | Yes |
| `climbingIndicatorActive` | boolean | Yes |
| `warnings` | LolRankedRankedQueueWarningsDTO |  |

### LolRankedRankedStats

| Field | Type | Required |
|-------|------|----------|
| `queues` | LolRankedRankedQueueStats[] | Yes |
| `queueMap` | object (map) | Yes |
| `highestRankedEntry` | LolRankedRankedQueueStats |  |
| `highestRankedEntrySR` | LolRankedRankedQueueStats |  |
| `earnedRegaliaRewardIds` | string[] | Yes |
| `rankedRegaliaLevel` | int32 | Yes |
| `highestCurrentSeasonReachedTierSR` | string | Yes |
| `highestPreviousSeasonEndTier` | string | Yes |
| `highestPreviousSeasonEndDivision` | LolRankedLeagueDivision | Yes |
| `splitsProgress` | object (map) | Yes |
| `currentSeasonSplitPoints` | int32 | Yes |
| `previousSeasonSplitPoints` | int32 | Yes |
| `seasons` | object (map) | Yes |

### LolRankedRatedLadderInfo

| Field | Type | Required |
|-------|------|----------|
| `queueType` | LolRankedLeagueQueueType | Yes |
| `standings` | LolRankedRatedLadderStanding[] | Yes |

### LolRankedRatedLadderStanding

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `puuid` | string | Yes |
| `summonerName` | string | Yes |
| `ratedTier` | LolRankedRatedTier | Yes |
| `leaguePoints` | int32 | Yes |
| `wins` | int32 | Yes |
| `position` | int32 | Yes |
| `positionDelta` | int32 | Yes |
| `previousPosition` | int32 | Yes |

### LolRankedRatedTier

**Type**: enum string

**Values**: `ORANGE`, `PURPLE`, `BLUE`, `GREEN`, `GRAY`, `NONE`


### LolRankedSignedRankedStatsDTO

| Field | Type | Required |
|-------|------|----------|
| `queues` | LolRankedRankedQueueStatsDTO[] | Yes |
| `earnedRegaliaRewardIds` | string[] | Yes |
| `highestPreviousSeasonEndTier` | string | Yes |
| `highestPreviousSeasonEndRank` | string | Yes |
| `splitsProgress` | object (map) | Yes |
| `currentSeasonSplitPoints` | int32 | Yes |
| `previousSeasonSplitPoints` | int32 | Yes |
| `seasons` | object (map) | Yes |
| `shouldShowIndicator` | boolean | Yes |
| `jwt` | string | Yes |

---

## LolRegalia


### LolRegaliaRegalia

| Field | Type | Required |
|-------|------|----------|
| `profileIconId` | int32 | Yes |
| `crestType` | string | Yes |
| `bannerType` | string | Yes |
| `summonerLevel` | uint32 | Yes |
| `lastSeasonHighestRank` | string |  |
| `highestRankedEntry` | LolRegaliaRegaliaRankedEntry |  |
| `selectedPrestigeCrest` | uint8 | Yes |

### LolRegaliaRegaliaAsync

| Field | Type | Required |
|-------|------|----------|
| `md5` | string | Yes |

### LolRegaliaRegaliaFrontendConfig

| Field | Type | Required |
|-------|------|----------|
| `hovercardEnabled` | boolean | Yes |
| `selectionsEnabled` | boolean | Yes |

### LolRegaliaRegaliaPreferences

| Field | Type | Required |
|-------|------|----------|
| `preferredCrestType` | string | Yes |
| `preferredBannerType` | string | Yes |
| `selectedPrestigeCrest` | uint8 | Yes |

### LolRegaliaRegaliaRankedEntry

| Field | Type | Required |
|-------|------|----------|
| `queueType` | LolRegaliaLeagueQueueType | Yes |
| `tier` | string | Yes |
| `division` | LolRegaliaLeagueDivision | Yes |
| `splitRewardLevel` | int32 | Yes |

### LolRegaliaRegaliaWithPreferences

| Field | Type | Required |
|-------|------|----------|
| `profileIconId` | int32 | Yes |
| `crestType` | string | Yes |
| `bannerType` | string | Yes |
| `preferredCrestType` | string | Yes |
| `preferredBannerType` | string | Yes |
| `selectedPrestigeCrest` | uint8 | Yes |
| `summonerLevel` | uint32 | Yes |
| `lastSeasonHighestRank` | string |  |
| `highestRankedEntry` | LolRegaliaRegaliaRankedEntry |  |

---

## LolRemedy


### LolRemedyMail

| Field | Type | Required |
|-------|------|----------|
| `mailId` | string | Yes |
| `message` | string | Yes |
| `state` | string | Yes |
| `createdAt` | uint64 | Yes |

---

## LolReplays


### LolReplaysMetadataState

**Type**: enum string

**Values**: `error`, `unsupported`, `lost`, `retryDownload`, `missingOrExpired`, `incompatible`, `downloading`, `download`, `watch`, `found`, `checking`


### LolReplaysReplayContextData

| Field | Type | Required |
|-------|------|----------|
| `componentType` | string | Yes |

### LolReplaysReplayCreateMetadata

| Field | Type | Required |
|-------|------|----------|
| `gameVersion` | string | Yes |
| `gameType` | string | Yes |
| `queueId` | int32 | Yes |
| `gameEnd` | uint64 | Yes |

### LolReplaysReplayMetadata

| Field | Type | Required |
|-------|------|----------|
| `state` | LolReplaysMetadataState | Yes |
| `gameId` | uint64 | Yes |
| `downloadProgress` | uint32 | Yes |

### LolReplaysReplaysConfiguration

| Field | Type | Required |
|-------|------|----------|
| `isReplaysEnabled` | boolean | Yes |
| `isReplaysForEndOfGameEnabled` | boolean | Yes |
| `isReplaysForMatchHistoryEnabled` | boolean | Yes |
| `isPatching` | boolean | Yes |
| `isInTournament` | boolean | Yes |
| `isPlayingGame` | boolean | Yes |
| `isPlayingReplay` | boolean | Yes |
| `isLoggedIn` | boolean | Yes |
| `gameVersion` | string | Yes |
| `minServerVersion` | string | Yes |
| `minutesUntilReplayConsideredLost` | int32 | Yes |

---

## LolReward


### LolRewardTrackRewardTrackError

| Field | Type | Required |
|-------|------|----------|
| `errorMessage` | string | Yes |
| `errorId` | string | Yes |

### LolRewardTrackRewardTrackItem

| Field | Type | Required |
|-------|------|----------|
| `state` | LolRewardTrackRewardTrackItemStates | Yes |
| `rewardOptions` | LolRewardTrackRewardTrackItemOption[] | Yes |
| `rewardTags` | LolRewardTrackRewardTrackItemTag[] | Yes |
| `progressRequired` | int64 | Yes |
| `threshold` | string | Yes |

### LolRewardTrackRewardTrackItemOption

| Field | Type | Required |
|-------|------|----------|
| `state` | LolRewardTrackRewardTrackItemStates | Yes |
| `thumbIconPath` | string | Yes |
| `splashImagePath` | string | Yes |
| `selected` | boolean | Yes |
| `overrideFooter` | string | Yes |
| `headerType` | LolRewardTrackRewardTrackItemHeaderType | Yes |
| `rewardName` | string | Yes |
| `rewardDescription` | string | Yes |
| `rewardItemType` | string | Yes |
| `rewardItemId` | string | Yes |
| `rewardFulfillmentSource` | string | Yes |
| `cardSize` | string | Yes |
| `rewardGroupId` | string | Yes |
| `celebrationType` | LolRewardTrackCelebrationType | Yes |
| `rewardInventoryTypes` | string[] | Yes |

### LolRewardTrackRewardTrackItemStates

**Type**: enum string

**Values**: `Selected`, `Unselected`, `Unlocked`, `Locked`


### LolRewardTrackRewardTrackItemTag

**Type**: enum string

**Values**: `Multiple`, `Choice`, `Instant`, `Free`, `Rare`


### LolRewardTrackRewardTrackProgress

| Field | Type | Required |
|-------|------|----------|
| `level` | int16 | Yes |
| `totalLevels` | int16 | Yes |
| `levelProgress` | uint16 | Yes |
| `futureLevelProgress` | uint16 | Yes |
| `passProgress` | int64 | Yes |
| `currentLevelXP` | int64 | Yes |
| `totalLevelXP` | int64 | Yes |
| `iteration` | uint32 | Yes |

### LolRewardTrackRewardTrackXP

| Field | Type | Required |
|-------|------|----------|
| `currentLevel` | int64 | Yes |
| `currentLevelXP` | int64 | Yes |
| `totalLevelXP` | int64 | Yes |
| `isBonusPhase` | boolean | Yes |
| `iteration` | uint32 | Yes |

### LolRewardTrackUnclaimedRewardsUIData

| Field | Type | Required |
|-------|------|----------|
| `rewardsCount` | int32 | Yes |
| `lockedTokensCount` | int32 | Yes |
| `timeOfLastUnclaimedReward` | int64 | Yes |

---

## LolRewards


### LolRewardsCelebrationType

**Type**: enum string

**Values**: `FULLSCREEN`, `TOAST`, `NONE`


### LolRewardsFSC

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `canvas` | LolRewardsFSCCanvas | Yes |
| `media` | LolRewardsFSCMedia | Yes |
| `rewards` | LolRewardsFSCRewards[] | Yes |

### LolRewardsFSCCanvas

| Field | Type | Required |
|-------|------|----------|
| `title` | string | Yes |
| `subtitle` | string | Yes |
| `canvasBackgroundImage` | string | Yes |
| `canvasSize` | LolRewardsFSCCanvasSize | Yes |
| `canvasDesign` | string | Yes |

### LolRewardsFSCMedia

| Field | Type | Required |
|-------|------|----------|
| `introAnimation` | string | Yes |
| `introAnimationAudio` | string | Yes |
| `introLowSpecImage` | string | Yes |
| `loopAnimation` | string | Yes |
| `loopAnimationAudio` | string | Yes |
| `transitionAnimation` | string | Yes |
| `transitionAnimationAudio` | string | Yes |

### LolRewardsFSCRewards

| Field | Type | Required |
|-------|------|----------|
| `itemId` | int32 | Yes |
| `imageOverlay` | string | Yes |
| `title` | string | Yes |
| `subtitle` | string | Yes |

### LolRewardsReward

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `itemId` | string | Yes |
| `itemType` | string | Yes |
| `quantity` | int32 | Yes |
| `fulfillmentSource` | string | Yes |
| `media` | object (map) | Yes |
| `localizations` | object (map) | Yes |

### LolRewardsRewardGrant

| Field | Type | Required |
|-------|------|----------|
| `info` | LolRewardsSvcRewardGrant | Yes |
| `rewardGroup` | LolRewardsSvcRewardGroup | Yes |

### LolRewardsRewardStrategy

**Type**: enum string

**Values**: `SELECTION`, `RANDOM`, `ALL`


### LolRewardsSelectionRequestDTO

| Field | Type | Required |
|-------|------|----------|
| `grantId` | string | Yes |
| `rewardGroupId` | string | Yes |
| `selections` | string[] | Yes |

### LolRewardsSelectionStrategyConfig

| Field | Type | Required |
|-------|------|----------|
| `minSelectionsAllowed` | uint32 | Yes |
| `maxSelectionsAllowed` | uint32 | Yes |

### LolRewardsSvcRewardGrant

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `granteeId` | string | Yes |
| `rewardGroupId` | string | Yes |
| `dateCreated` | string | Yes |
| `status` | LolRewardsGrantStatus | Yes |
| `grantElements` | LolRewardsSvcRewardGrantElement[] | Yes |
| `selectedIds` | string[] | Yes |
| `viewed` | boolean | Yes |
| `grantorDescription` | LolRewardsGrantorDescription | Yes |
| `messageParameters` | object (map) | Yes |

### LolRewardsSvcRewardGroup

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `productId` | string | Yes |
| `types` | string[] | Yes |
| `rewards` | LolRewardsReward[] | Yes |
| `childRewardGroupIds` | string[] | Yes |
| `rewardStrategy` | LolRewardsRewardStrategy | Yes |
| `selectionStrategyConfig` | LolRewardsSelectionStrategyConfig |  |
| `active` | boolean | Yes |
| `media` | object (map) | Yes |
| `localizations` | object (map) | Yes |
| `celebrationType` | LolRewardsCelebrationType | Yes |

---

## LolRiot


### LolRiotMessagingServiceChampionMasteryLevelUp

| Field | Type | Required |
|-------|------|----------|
| `id` | uint64 | Yes |
| `puuid` | string | Yes |
| `championId` | int32 | Yes |
| `hasLeveledUp` | boolean | Yes |
| `championLevel` | int64 | Yes |

---

## LolRso


### LolRsoAuthAccessToken

| Field | Type | Required |
|-------|------|----------|
| `token` | string | Yes |
| `scopes` | string[] | Yes |
| `expiry` | uint64 | Yes |

### LolRsoAuthAuthError

| Field | Type | Required |
|-------|------|----------|
| `error` | string | Yes |
| `errorDescription` | string | Yes |

### LolRsoAuthAuthorization

| Field | Type | Required |
|-------|------|----------|
| `currentPlatformId` | string | Yes |
| `currentAccountId` | uint64 | Yes |
| `subject` | string | Yes |

### LolRsoAuthDeviceId

| Field | Type | Required |
|-------|------|----------|
| `collectorServerName` | string | Yes |
| `merchantId` | string | Yes |
| `sessionId` | string | Yes |
| `installId` | string | Yes |
| `frameUrl` | string | Yes |

### LolRsoAuthIdToken

| Field | Type | Required |
|-------|------|----------|
| `token` | string | Yes |
| `expiry` | uint64 | Yes |

### LolRsoAuthPublicClientConfig

| Field | Type | Required |
|-------|------|----------|
| `url` | string | Yes |
| `clientId` | string | Yes |

### LolRsoAuthRSOConfigReadyState

| Field | Type | Required |
|-------|------|----------|
| `ready` | boolean | Yes |

### LolRsoAuthRSOPlayerCredentials

| Field | Type | Required |
|-------|------|----------|
| `username` | string | Yes |
| `password` | string | Yes |
| `platformId` | string | Yes |

### LolRsoAuthRegionStatus

| Field | Type | Required |
|-------|------|----------|
| `platformId` | string | Yes |
| `enabled` | boolean | Yes |
| `isLQFallbackAllowed` | boolean | Yes |
| `isUserInfoEnabled` | boolean | Yes |

### LolRsoAuthUserInfo

| Field | Type | Required |
|-------|------|----------|
| `userInfo` | string | Yes |

---

## LolSanctum


### LolSanctumBannerOddsInfo

| Field | Type | Required |
|-------|------|----------|
| `bannerId` | string | Yes |
| `sTierRewards` | LolSanctumSanctumTierDrops | Yes |
| `aTierRewards` | LolSanctumSanctumTierDrops | Yes |
| `bTierRewards` | LolSanctumSanctumTierDrops | Yes |
| `nodeIdToDropDetailsMap` | object (map) | Yes |

### LolSanctumCapCounterData

| Field | Type | Required |
|-------|------|----------|
| `amount` | uint32 | Yes |
| `version` | uint32 | Yes |
| `active` | boolean | Yes |
| `lastModifiedDate` | string | Yes |

### LolSanctumGameDataBannerSkin

| Field | Type | Required |
|-------|------|----------|
| `id` | uint32 | Yes |
| `name` | string | Yes |
| `rarity` | string | Yes |

### LolSanctumGameDataPityCounter

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `name` | string | Yes |

### LolSanctumGameDataSanctumBannerVo

| Field | Type | Required |
|-------|------|----------|
| `path` | string | Yes |
| `defaultDelayMillis` | uint32 | Yes |
| `localeOverrides` | LolSanctumGameDataSanctumBannerVoOverrideOptions[] | Yes |

### LolSanctumGameDataSanctumCurrency

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `name` | string | Yes |
| `currencyId` | string | Yes |
| `capCatalogEntryId` | string | Yes |

### LolSanctumSanctumBannerResponse

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `bannerBackgroundTexture` | string | Yes |
| `bannerBackgroundParallax` | string | Yes |
| `bannerChaseAnimationWebmPath` | string | Yes |
| `bannerChaseAnimationParallax` | string | Yes |
| `chasePityCounter` | LolSanctumGameDataPityCounter | Yes |
| `chasePityThreshold` | uint32 | Yes |
| `highlightPityThreshold` | uint32 | Yes |
| `skinIntroAnimationPath` | string | Yes |
| `skinIntroSfxPath` | string | Yes |
| `chaseCelebrationIntroAnimationPath` | string | Yes |
| `chaseCelebrationVo` | LolSanctumGameDataSanctumBannerVo | Yes |
| `hubIntroVo` | LolSanctumGameDataSanctumBannerVo | Yes |
| `sanctumRollVideos` | LolSanctumSanctumVignette | Yes |
| `bannerSkin` | LolSanctumGameDataBannerSkin | Yes |
| `bannerCurrency` | LolSanctumGameDataSanctumCurrency | Yes |
| `capCatalogStoreId` | string | Yes |
| `dropTableCatalogEntryId` | string | Yes |
| `pityCounter` | LolSanctumCapCounterData | Yes |
| `startDate` | int64 | Yes |
| `endDate` | int64 | Yes |

### LolSanctumSanctumPurchaseRequest

| Field | Type | Required |
|-------|------|----------|
| `bannerId` | string | Yes |
| `quantity` | uint32 | Yes |

### LolSanctumSanctumPurchaseResponse

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `status` | LolSanctumSanctumPurchaseResponseStatus | Yes |
| `rollResultNodeIds` | string[] | Yes |

### LolSanctumSanctumPurchaseResponseStatus

**Type**: enum string

**Values**: `Failure`, `Success`, `Pending`, `None`


### LolSanctumSanctumTierDrops

| Field | Type | Required |
|-------|------|----------|
| `tier` | uint32 | Yes |
| `odds` | number | Yes |
| `mainDrops` | LolSanctumSanctumDropDetails[] | Yes |
| `fallbackDrops` | LolSanctumSanctumDropDetails[] | Yes |

### LolSanctumSanctumVignette

| Field | Type | Required |
|-------|------|----------|
| `introTierOneWebmPath` | string | Yes |
| `introTierOneMultiWebmPath` | string | Yes |
| `introTierTwoWebmPath` | string | Yes |
| `introTierTwoMultiWebmPath` | string | Yes |
| `introTierThreeWebmPath` | string | Yes |
| `introTierThreeMultiWebmPath` | string | Yes |

### LolSanctumSetSelectedBannerIdRequest

| Field | Type | Required |
|-------|------|----------|
| `bannerId` | string | Yes |

---

## LolSeasons


### LolSeasonsAllProductSeasonQuery

| Field | Type | Required |
|-------|------|----------|
| `lastNYears` | uint16 | Yes |

### LolSeasonsAllSeasonsProduct

| Field | Type | Required |
|-------|------|----------|
| `seasonId` | int32 | Yes |
| `seasonStart` | int64 | Yes |
| `seasonEnd` | int64 | Yes |
| `act` | boolean | Yes |
| `metadata` | LolSeasonsSeasonMetaData | Yes |

### LolSeasonsSeasonMetaData

| Field | Type | Required |
|-------|------|----------|
| `year` | uint16 | Yes |
| `locKey` | string | Yes |
| `publicName` | string | Yes |
| `currentSplit` | int32 | Yes |
| `totalSplit` | int32 | Yes |

---

## LolService


### LolServiceStatusServiceStatusResource

| Field | Type | Required |
|-------|------|----------|
| `status` | string | Yes |
| `humanReadableUrl` | string | Yes |

### LolServiceStatusTickerMessage

| Field | Type | Required |
|-------|------|----------|
| `severity` | string | Yes |
| `createdAt` | string | Yes |
| `updatedAt` | string | Yes |
| `heading` | string | Yes |
| `message` | string | Yes |

---

## LolSettings


### LolSettingsHoneyfruitVNGPublisherSettings

| Field | Type | Required |
|-------|------|----------|
| `visible` | boolean | Yes |

### LolSettingsSettingCategory

| Field | Type | Required |
|-------|------|----------|
| `schemaVersion` | int32 | Yes |
| `data` | object (map) | Yes |

### LolSettingsSettingsConfig

| Field | Type | Required |
|-------|------|----------|
| `isHotkeysEnabled` | boolean | Yes |
| `isSoundEnabled` | boolean | Yes |
| `isInterfaceEnabled` | boolean | Yes |
| `isGameplayEnabled` | boolean | Yes |
| `isReplaysEnabled` | boolean | Yes |
| `isPrivacyNoticeEnabled` | boolean | Yes |
| `isTermsEnabled` | boolean | Yes |
| `isLegalStatementsEnabled` | boolean | Yes |
| `localizedLicensesURL` | string | Yes |

---

## LolShoppefront


### LolShoppefrontBulkPurchaseRequest

| Field | Type | Required |
|-------|------|----------|
| `purchaseItems` | LolShoppefrontPurchaseRequest[] | Yes |
| `purchaseTimeOut` | uint32 | Yes |

### LolShoppefrontPurchaseRequest

| Field | Type | Required |
|-------|------|----------|
| `storeId` | string | Yes |
| `catalogEntryId` | string | Yes |
| `quantity` | uint32 | Yes |
| `paymentOptions` | string[] | Yes |
| `purchaseTimeOut` | uint32 | Yes |

### LolShoppefrontPurchaseResponse

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `status` | LolShoppefrontPurchaseResponseStatus | Yes |
| `numberOfPendingPurchases` | uint8 | Yes |

### LolShoppefrontPurchaseResponseStatus

**Type**: enum string

**Values**: `Failure`, `Success`, `Pending`, `None`


---

## LolShutdown


### LolShutdownShutdownNotification

| Field | Type | Required |
|-------|------|----------|
| `reason` | LolShutdownShutdownReason | Yes |
| `countdown` | number | Yes |
| `additionalInfo` | string | Yes |

### LolShutdownShutdownReason

**Type**: enum string

**Values**: `PlayerBanned`, `LcuAlphaDisabled`, `PlatformMaintenance`, `Invalid`


---

## LolSimple


### LolSimpleDialogMessagesLocalMessageRequest

| Field | Type | Required |
|-------|------|----------|
| `msgType` | string | Yes |
| `msgBody` | string[] | Yes |

### LolSimpleDialogMessagesMessage

| Field | Type | Required |
|-------|------|----------|
| `id` | int64 | Yes |
| `type` | string | Yes |
| `body` | object (map) | Yes |

---

## LolSocial


### LolSocialLeaderboardSocialLeaderboardData

| Field | Type | Required |
|-------|------|----------|
| `rowData` | LolSocialLeaderboardSocialLeaderboardRowData[] | Yes |
| `nextUpdateTime` | int64 | Yes |

### LolSocialLeaderboardSocialLeaderboardRowData

| Field | Type | Required |
|-------|------|----------|
| `puuid` | string | Yes |
| `summonerId` | uint64 | Yes |
| `summonerName` | string | Yes |
| `gameName` | string | Yes |
| `tagLine` | string | Yes |
| `provisionalGamesRemaining` | int32 | Yes |
| `isProvisional` | boolean | Yes |
| `tier` | string | Yes |
| `division` | LolSocialLeaderboardLeagueDivision | Yes |
| `leaguePoints` | int32 | Yes |
| `wins` | int32 | Yes |
| `summonerLevel` | int32 | Yes |
| `profileIconId` | int32 | Yes |
| `availability` | string | Yes |
| `leaderboardPosition` | int32 | Yes |
| `isGiftable` | boolean | Yes |

---

## LolSpectator


### LolSpectatorSpectateGameInfo

| Field | Type | Required |
|-------|------|----------|
| `dropInSpectateGameId` | string | Yes |
| `gameQueueType` | string | Yes |
| `allowObserveMode` | string | Yes |
| `puuid` | string | Yes |
| `spectatorKey` | string | Yes |

### LolSpectatorSpectatorConfig

| Field | Type | Required |
|-------|------|----------|
| `isEnabled` | boolean | Yes |
| `isSpectatorDelayConfigurable` | boolean | Yes |
| `isBracketSpectatingEnabled` | boolean | Yes |
| `isUsingClientConfigForSpectator` | boolean | Yes |
| `spectatableQueues` | integer[] | Yes |

### LolSpectatorSpectatorKeySpectateResource

| Field | Type | Required |
|-------|------|----------|
| `availableForWatching` | boolean | Yes |
| `reason` | string | Yes |

### LolSpectatorSummonerIdAvailability

| Field | Type | Required |
|-------|------|----------|
| `availableForWatching` | integer[] | Yes |

### LolSpectatorSummonerOrTeamAvailabilty

| Field | Type | Required |
|-------|------|----------|
| `availableForWatching` | string[] | Yes |

### LolSpectatorSummonerPuuidsSpectateResource

| Field | Type | Required |
|-------|------|----------|
| `availableForWatching` | string[] | Yes |

---

## LolStatstones


### LolStatstonesChampionStatstoneSetSummary

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `stonesAvailable` | uint32 | Yes |
| `stonesOwned` | uint32 | Yes |
| `stonesIlluminated` | uint32 | Yes |
| `milestonesPassed` | uint32 | Yes |

### LolStatstonesChampionStatstoneSummary

| Field | Type | Required |
|-------|------|----------|
| `championId` | int32 | Yes |
| `stonesAvailable` | uint32 | Yes |
| `stonesOwned` | uint32 | Yes |
| `stonesIlluminated` | uint32 | Yes |
| `milestonesPassed` | uint32 | Yes |
| `sets` | LolStatstonesChampionStatstoneSetSummary[] | Yes |

### LolStatstonesEogNotificationEnvelope

| Field | Type | Required |
|-------|------|----------|
| `selfStatstoneProgress` | LolStatstonesStatstoneProgress[] | Yes |
| `selfPersonalBests` | LolStatstonesPersonalBestNotification[] | Yes |
| `selfMilestoneProgress` | LolStatstonesMilestoneProgressNotification[] | Yes |
| `othersPersonalBests` | LolStatstonesPersonalBestNotification[] | Yes |

### LolStatstonesGameDataStatstonePack

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `description` | string | Yes |
| `contentId` | string | Yes |
| `itemId` | int32 | Yes |

### LolStatstonesMilestoneProgressNotification

| Field | Type | Required |
|-------|------|----------|
| `statstoneId` | string | Yes |
| `statstoneName` | string | Yes |
| `threshold` | int32 | Yes |
| `imageUrl` | string | Yes |
| `level` | string | Yes |

### LolStatstonesPersonalBestNotification

| Field | Type | Required |
|-------|------|----------|
| `summoner` | LolStatstonesSummoner | Yes |
| `statstoneId` | string | Yes |
| `statstoneName` | string | Yes |
| `personalBest` | string | Yes |
| `imageUrl` | string | Yes |

### LolStatstonesPriceInfo

| Field | Type | Required |
|-------|------|----------|
| `currency` | string | Yes |
| `price` | uint32 | Yes |

### LolStatstonesProfileStatstoneSummary

| Field | Type | Required |
|-------|------|----------|
| `championId` | int32 | Yes |
| `name` | string | Yes |
| `value` | string | Yes |
| `imageUrl` | string | Yes |
| `category` | string | Yes |

### LolStatstonesStatstone

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `statstoneId` | string | Yes |
| `boundChampionItemId` | uint32 | Yes |
| `nextMilestone` | string | Yes |
| `completionValue` | number | Yes |
| `isComplete` | boolean | Yes |
| `isFeatured` | boolean | Yes |
| `isEpic` | boolean | Yes |
| `isRetired` | boolean | Yes |
| `category` | string | Yes |
| `imageUrl` | string | Yes |
| `description` | string | Yes |
| `formattedValue` | string | Yes |
| `formattedPersonalBest` | string | Yes |
| `formattedMilestoneLevel` | string | Yes |
| `playerRecord` | LolStatstonesStatstonePlayerRecord |  |

### LolStatstonesStatstoneFeaturedRequest

| Field | Type | Required |
|-------|------|----------|
| `index` | int32 | Yes |
| `existingFeatured` | LolStatstonesStatstone[] | Yes |

### LolStatstonesStatstonePlayerRecord

| Field | Type | Required |
|-------|------|----------|
| `puuid` | string | Yes |
| `statstoneId` | string | Yes |
| `value` | uint32 | Yes |
| `personalBest` | uint32 | Yes |
| `milestoneLevel` | uint32 | Yes |
| `dateAcquired` | string | Yes |
| `dateModified` | string | Yes |
| `dateCompleted` | string | Yes |
| `dateArchived` | string | Yes |
| `entitled` | boolean | Yes |

### LolStatstonesStatstoneProgress

| Field | Type | Required |
|-------|------|----------|
| `statstoneId` | string | Yes |
| `statstoneName` | string | Yes |
| `description` | string | Yes |
| `imageUrl` | string | Yes |
| `delta` | string | Yes |
| `value` | string | Yes |
| `nextMilestone` | string | Yes |
| `existingProgressPercent` | string | Yes |
| `newProgressPercent` | string | Yes |
| `newMilestoneDifference` | string | Yes |
| `totalProgressPercent` | string | Yes |
| `category` | string | Yes |
| `level` | int32 | Yes |
| `best` | int32 | Yes |
| `isNewBest` | boolean | Yes |

### LolStatstonesStatstoneSet

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `statstones` | LolStatstonesStatstone[] | Yes |
| `stonesOwned` | uint32 | Yes |
| `milestonesPassed` | uint32 | Yes |
| `itemId` | uint32 | Yes |
| `inventoryType` | string | Yes |
| `subInventoryType` | string | Yes |
| `itemInstanceID` | string | Yes |
| `prices` | LolStatstonesPriceInfo[] | Yes |
| `ownedFromPacks` | LolStatstonesGameDataStatstonePack[] | Yes |

---

## LolStore


### LolStoreAliasChangeNotificationResource

| Field | Type | Required |
|-------|------|----------|
| `id` | uint64 | Yes |
| `type` | string | Yes |
| `details` | LolStoreAliasDetail | Yes |

### LolStoreAliasDetail

| Field | Type | Required |
|-------|------|----------|
| `field` | string | Yes |
| `old_value` | string | Yes |
| `new_value` | string | Yes |

### LolStoreBundled

| Field | Type | Required |
|-------|------|----------|
| `flexible` | boolean | Yes |
| `items` | LolStoreBundledItem[] | Yes |
| `minimumPrices` | LolStoreBundledItemCost[] | Yes |

### LolStoreCapOffer

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `typeId` | string | Yes |
| `label` | string | Yes |
| `productId` | string | Yes |
| `merchantId` | string | Yes |
| `payload` | object (map) | Yes |
| `active` | boolean | Yes |
| `startDate` | string | Yes |
| `createdDate` | string | Yes |

### LolStoreCatalogItem

| Field | Type | Required |
|-------|------|----------|
| `itemId` | int32 | Yes |
| `inventoryType` | string | Yes |
| `iconUrl` | string |  |
| `localizations` | object (map) |  |
| `active` | boolean |  |
| `bundled` | LolStoreBundled |  |
| `inactiveDate` | string |  |
| `maxQuantity` | int32 |  |
| `prices` | LolStoreItemCost[] |  |
| `releaseDate` | string |  |
| `sale` | LolStoreSale |  |
| `subInventoryType` | string |  |
| `tags` | string[] |  |
| `itemRequirements` | LolStoreItemKey[] |  |
| `metadata` | LolStoreItemMetadataEntry[] |  |
| `itemInstanceId` | string |  |
| `offerId` | string |  |

### LolStoreGiftingFriend

| Field | Type | Required |
|-------|------|----------|
| `friendsSince` | string | Yes |
| `oldFriends` | boolean | Yes |
| `summonerId` | uint64 | Yes |
| `nick` | string | Yes |

### LolStoreItemCost

| Field | Type | Required |
|-------|------|----------|
| `currency` | string | Yes |
| `cost` | int64 | Yes |
| `discount` | number |  |

### LolStoreItemKey

| Field | Type | Required |
|-------|------|----------|
| `inventoryType` | string | Yes |
| `itemId` | int32 | Yes |

### LolStoreItemMetadataEntry

| Field | Type | Required |
|-------|------|----------|
| `type` | string | Yes |
| `value` | string | Yes |

### LolStoreItemSale

| Field | Type | Required |
|-------|------|----------|
| `id` | uint64 | Yes |
| `active` | boolean | Yes |
| `item` | LolStoreItemKey | Yes |
| `sale` | LolStoreSale | Yes |

### LolStoreOrderNotificationResource

| Field | Type | Required |
|-------|------|----------|
| `id` | uint64 | Yes |
| `eventTypeId` | string | Yes |
| `eventType` | string | Yes |
| `status` | string | Yes |

### LolStorePurchaseOrderResponseDTO

| Field | Type | Required |
|-------|------|----------|
| `rpBalance` | int64 | Yes |
| `ipBalance` | int64 | Yes |
| `transactions` | LolStoreTransactionResponseDTO[] | Yes |

### LolStoreSale

| Field | Type | Required |
|-------|------|----------|
| `startDate` | string | Yes |
| `endDate` | string | Yes |
| `prices` | LolStoreItemCost[] | Yes |

### LolStoreStoreStatus

| Field | Type | Required |
|-------|------|----------|
| `storefrontIsRunning` | boolean | Yes |

### LolStoreTransactionResponseDTO

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `inventoryType` | string | Yes |
| `itemId` | int32 | Yes |

---

## LolSuggested


### LolSuggestedPlayersSuggestedPlayersReason

**Type**: enum string

**Values**: `LegacyPlayAgain`, `HonorInteractions`, `VictoriousComrade`, `FriendOfFriend`, `OnlineFriend`, `PreviousPremade`


### LolSuggestedPlayersSuggestedPlayersReportedPlayer

| Field | Type | Required |
|-------|------|----------|
| `offenderSummonerId` | uint64 | Yes |

### LolSuggestedPlayersSuggestedPlayersSuggestedPlayer

| Field | Type | Required |
|-------|------|----------|
| `summonerName` | string | Yes |
| `summonerId` | uint64 | Yes |
| `commonFriendName` | string | Yes |
| `commonFriendId` | uint64 | Yes |
| `reason` | LolSuggestedPlayersSuggestedPlayersReason | Yes |
| `gameId` | uint64 | Yes |

### LolSuggestedPlayersSuggestedPlayersVictoriousComrade

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `summonerName` | string | Yes |

---

## LolSummoner


### LolSummonerAccountIdAndSummonerId

| Field | Type | Required |
|-------|------|----------|
| `accountId` | uint64 | Yes |
| `summonerId` | uint64 | Yes |

### LolSummonerAlias

| Field | Type | Required |
|-------|------|----------|
| `gameName` | string | Yes |
| `tagLine` | string | Yes |

### LolSummonerAliasAvailability

| Field | Type | Required |
|-------|------|----------|
| `alias` | LolSummonerAlias | Yes |
| `errorCode` | LolSummonerAliasAvailabilityCode | Yes |
| `errorMessage` | string | Yes |
| `isSuccess` | boolean | Yes |

### LolSummonerAliasAvailabilityCode

**Type**: enum string

**Values**: `server_error`, `rate_limited`, `name_not_available`, `name_change_forbidden`, `no_error`


### LolSummonerAliasLookupResponse

| Field | Type | Required |
|-------|------|----------|
| `alias` | LolSummonerAlias | Yes |
| `puuid` | string | Yes |

### LolSummonerAutoFillQueueDto

| Field | Type | Required |
|-------|------|----------|
| `queueId` | int32 | Yes |
| `autoFillEligible` | boolean | Yes |
| `autoFillProtectedForStreaking` | boolean | Yes |
| `autoFillProtectedForPromos` | boolean | Yes |
| `overridePositionSelectionsWithFill` | boolean | Yes |
| `autoFillProtectedForRemedy` | boolean | Yes |

### LolSummonerPlayerNameMode

**Type**: enum string

**Values**: `ALIAS`, `DARKMODE`, `SUMMONER`


### LolSummonerPlayerNameState

| Field | Type | Required |
|-------|------|----------|
| `isAliasChangeRequired` | boolean | Yes |
| `isAliasMissing` | boolean | Yes |
| `isTaglineCustomizable` | boolean | Yes |

### LolSummonerProfilePrivacy

| Field | Type | Required |
|-------|------|----------|
| `enabledState` | LolSummonerProfilePrivacyEnabledState | Yes |
| `setting` | LolSummonerProfilePrivacySetting | Yes |

### LolSummonerProfilePrivacyEnabledState

**Type**: enum string

**Values**: `DISABLED`, `ENABLED`, `UNKNOWN`


### LolSummonerProfilePrivacySetting

**Type**: enum string

**Values**: `PUBLIC`, `PRIVATE`


### LolSummonerProfilesChampionMasteryData

| Field | Type | Required |
|-------|------|----------|
| `championId` | int64 | Yes |
| `championLevel` | int32 | Yes |
| `championPoints` | int32 | Yes |
| `highestGrade` | string |  |

### LolSummonerProfilesChampionMasteryView

| Field | Type | Required |
|-------|------|----------|
| `puuid` | string | Yes |
| `data` | LolSummonerProfilesChampionMasteryData[] | Yes |

### LolSummonerProfilesHonorView

| Field | Type | Required |
|-------|------|----------|
| `honorLevel` | int32 | Yes |
| `redemptions` | LolSummonerProfilesRedemption[] | Yes |

### LolSummonerProfilesLolEosRewardView

| Field | Type | Required |
|-------|------|----------|
| `seasonId` | int32 | Yes |
| `rewardIds` | string[] | Yes |
| `highestRankPerQueueId` | object (map) | Yes |
| `eligibility` | LolSummonerProfilesRewardsEligibility | Yes |

### LolSummonerProfilesPrivacyView

| Field | Type | Required |
|-------|------|----------|
| `anonymityEnabled` | boolean | Yes |
| `nameOnlyAnonymityEnabled` | boolean | Yes |

### LolSummonerProfilesRedemption

| Field | Type | Required |
|-------|------|----------|
| `required` | int32 | Yes |
| `remaining` | int32 | Yes |
| `eventType` | string | Yes |

### LolSummonerProfilesRestriction

| Field | Type | Required |
|-------|------|----------|
| `restrictionType` | string | Yes |
| `restrictionReason` | string | Yes |
| `expirationData` | LolSummonerProfilesRestrictionExpiration | Yes |

### LolSummonerProfilesRestrictionsView

| Field | Type | Required |
|-------|------|----------|
| `restrictions` | LolSummonerProfilesRestriction[] | Yes |

### LolSummonerProfilesRewardsEligibility

| Field | Type | Required |
|-------|------|----------|
| `honorRequirementEnabled` | boolean | Yes |
| `minHonorLevelForRewards` | int32 | Yes |
| `honorLevel` | int32 |  |

### LolSummonerProfilesSummonerLevel

| Field | Type | Required |
|-------|------|----------|
| `puuid` | string | Yes |
| `level` | uint32 | Yes |
| `xpSinceLastLevel` | uint32 | Yes |
| `xpToNextLevel` | uint32 | Yes |

### LolSummonerStatus

| Field | Type | Required |
|-------|------|----------|
| `ready` | boolean | Yes |

### LolSummonerSummoner

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `accountId` | uint64 | Yes |
| `displayName` | string | Yes |
| `internalName` | string | Yes |
| `profileIconId` | int32 | Yes |
| `summonerLevel` | uint32 | Yes |
| `xpSinceLastLevel` | uint64 | Yes |
| `xpUntilNextLevel` | uint64 | Yes |
| `percentCompleteForNextLevel` | uint32 | Yes |
| `rerollPoints` | LolSummonerSummonerRerollPoints | Yes |
| `puuid` | string | Yes |
| `nameChangeFlag` | boolean | Yes |
| `unnamed` | boolean | Yes |
| `privacy` | LolSummonerProfilePrivacySetting | Yes |
| `gameName` | string | Yes |
| `tagLine` | string | Yes |

### LolSummonerSummonerIcon

| Field | Type | Required |
|-------|------|----------|
| `profileIconId` | int32 | Yes |
| `inventoryToken` | string | Yes |

### LolSummonerSummonerIdAndIcon

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `profileIconId` | int32 | Yes |
| `puuid` | string | Yes |

### LolSummonerSummonerIdAndName

| Field | Type | Required |
|-------|------|----------|
| `summonerId` | uint64 | Yes |
| `displayName` | string | Yes |
| `puuid` | string | Yes |

### LolSummonerSummonerProfileUpdate

| Field | Type | Required |
|-------|------|----------|
| `key` | string | Yes |
| `value` | object (map) | Yes |
| `inventory` | string | Yes |

### LolSummonerSummonerRequestedName

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |

### LolSummonerSummonerRerollPoints

| Field | Type | Required |
|-------|------|----------|
| `pointsToReroll` | uint32 | Yes |
| `currentPoints` | uint32 | Yes |
| `numberOfRolls` | uint32 | Yes |
| `maxRolls` | uint32 | Yes |
| `pointsCostToRoll` | uint32 | Yes |

---

## LolTastes


### LolTastesDataModelResponse

| Field | Type | Required |
|-------|------|----------|
| `responseCode` | int64 | Yes |
| `modelData` | object (map) | Yes |

---

## LolTft


### LolTftEventPveEventPVELevelState

**Type**: enum string

**Values**: `kError`, `kCleared`, `kUnlocked`, `kUnseenUnlocked`, `kLocked`


### LolTftEventPveTFTEventBuddy

| Field | Type | Required |
|-------|------|----------|
| `isPremium` | boolean | Yes |
| `completedSummit` | boolean | Yes |
| `completedPerfectRun` | boolean | Yes |
| `journeyTrackUnlockLevel` | uint8 | Yes |
| `id` | uint32 | Yes |
| `contentId` | string | Yes |
| `name` | string | Yes |
| `description` | string | Yes |
| `selectionVOKey` | string | Yes |
| `shortDescription` | string | Yes |
| `iconTexture` | string | Yes |
| `portraitTexture` | string | Yes |
| `hubTexture` | string | Yes |
| `summitMissionId` | string | Yes |
| `perfectRunMissionId` | string | Yes |
| `equipState` | LolTftEventPveTFTEventBuddyEquipState | Yes |

### LolTftEventPveTFTEventBuddyEquipState

**Type**: enum string

**Values**: `kEquipped`, `kUnlocked`, `kUnseenUnlocked`, `kLocked`


### LolTftEventPveTFTEventPVEEoGMissionReward

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `iconTexture` | string | Yes |
| `type` | string | Yes |
| `itemTypeId` | string | Yes |
| `quantity` | uint16 | Yes |
| `firstCompletionLevelMission` | boolean | Yes |

### LolTftEventPveTFTEventPVEHub

| Field | Type | Required |
|-------|------|----------|
| `levels` | LolTftEventPveTFTEventPVELevel[] | Yes |
| `ultimateVictory` | LolTftEventPveTFTEventPVEUltimateVictory | Yes |

### LolTftEventPveTFTEventPVEJourneyTrackBonuses

| Field | Type | Required |
|-------|------|----------|
| `names` | string[] | Yes |

### LolTftEventPveTFTEventPVELevel

| Field | Type | Required |
|-------|------|----------|
| `status` | LolTftEventPveEventPVELevelState | Yes |
| `isSelected` | boolean | Yes |
| `difficultyLevel` | uint8 | Yes |
| `id` | uint8 | Yes |
| `contentId` | string | Yes |
| `name` | string | Yes |
| `recommended` | string | Yes |
| `description` | string | Yes |
| `descriptionUnlocked` | string | Yes |
| `descriptionCleared` | string | Yes |
| `missionId` | string | Yes |
| `lockedPortrait` | string | Yes |
| `unlockedPortrait` | string | Yes |
| `clearedPortrait` | string | Yes |
| `levelSymbol` | string | Yes |
| `rewards` | LolTftEventPveTFTEventPVEReward[] | Yes |

### LolTftEventPveTFTEventPVEReward

| Field | Type | Required |
|-------|------|----------|
| `rewardName` | string | Yes |
| `rewardImage` | string | Yes |
| `rewardQuantity` | uint32 | Yes |

### LolTftEventPveTFTEventPVEUltimateVictory

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `description` | string | Yes |
| `reward` | LolTftEventPveTFTEventPVEReward | Yes |
| `achievedUltimateVictory` | boolean | Yes |
| `seenCelebration` | boolean | Yes |

### LolTftEventTFTEventMissionChain

| Field | Type | Required |
|-------|------|----------|
| `chainIndex` | int32 | Yes |
| `chainSize` | uint32 | Yes |
| `missions` | PlayerMissionDTO[] | Yes |

### LolTftEventTFTEventMissionChains

| Field | Type | Required |
|-------|------|----------|
| `missionChains` | LolTftEventTFTEventMissionChain[] | Yes |
| `seriesId` | string | Yes |

### LolTftLolTftBackgrounds

| Field | Type | Required |
|-------|------|----------|
| `backgrounds` | object (map) | Yes |

### LolTftLolTftBattlePassHub

| Field | Type | Required |
|-------|------|----------|
| `battlePassXPBoosted` | boolean | Yes |

### LolTftLolTftEvent

| Field | Type | Required |
|-------|------|----------|
| `titleTranslationKey` | string | Yes |
| `eventId` | string | Yes |
| `enabled` | boolean | Yes |
| `url` | string | Yes |
| `urlFaq` | string | Yes |
| `startDate` | string | Yes |
| `endDate` | string | Yes |
| `seriesId` | string | Yes |
| `dailyLoginSeriesId` | string | Yes |
| `queueIds` | integer[] | Yes |
| `defaultLandingPage` | boolean | Yes |
| `eventHubAssetKey` | string | Yes |
| `eventHubTemplateType` | string | Yes |
| `eventPassId` | string | Yes |
| `skillTreePassId` | string | Yes |
| `eventFuture` | boolean | Yes |
| `weblinkSubnavs` | LolTftLolTftEventWebLinkSubnav[] | Yes |

### LolTftLolTftEvents

| Field | Type | Required |
|-------|------|----------|
| `subNavTabs` | LolTftLolTftEvent[] | Yes |

### LolTftLolTftHomeHub

| Field | Type | Required |
|-------|------|----------|
| `enabled` | boolean | Yes |
| `tacticianPromoOfferIds` | string[] | Yes |
| `primeGamingPromoOffer` | LolTftLolTftPrimeGaming |  |
| `overrideUrl` | string | Yes |
| `headerButtonsOverrideUrl` | string | Yes |
| `rotatingShopPromos` | LolTftTFTRotatingShopPromos | Yes |

### LolTftLolTftNewsHub

| Field | Type | Required |
|-------|------|----------|
| `enabled` | boolean | Yes |
| `url` | string | Yes |

### LolTftLolTftPrimeGaming

| Field | Type | Required |
|-------|------|----------|
| `url` | string | Yes |
| `assetId` | string | Yes |

### LolTftLolTftPromoButton

| Field | Type | Required |
|-------|------|----------|
| `enabled` | boolean | Yes |
| `showTimerWhileEventActive` | boolean | Yes |
| `eventAssetId` | string | Yes |
| `eventKey` | string | Yes |
| `url` | string | Yes |

### LolTftLolTftPromoButtons

| Field | Type | Required |
|-------|------|----------|
| `promoButtons` | LolTftLolTftPromoButton[] | Yes |

### LolTftLolTftTencentEventHubConfig

| Field | Type | Required |
|-------|------|----------|
| `troveAssetId` | string | Yes |
| `troveURL` | string | Yes |
| `logoAssetId` | string | Yes |

### LolTftLolTftTencentEventHubConfigs

| Field | Type | Required |
|-------|------|----------|
| `tencentEventhubConfigs` | LolTftLolTftTencentEventHubConfig[] | Yes |

### LolTftPassChapter

| Field | Type | Required |
|-------|------|----------|
| `localizedTitle` | string | Yes |
| `localizedDescription` | string | Yes |
| `cardImage` | string | Yes |
| `backgroundImage` | string | Yes |
| `objectiveBannerImage` | string | Yes |
| `chapterStart` | uint16 | Yes |
| `chapterEnd` | uint16 | Yes |
| `chapterNumber` | uint16 | Yes |
| `levelFocus` | uint16 | Yes |

### LolTftPassObjectivesBanner

| Field | Type | Required |
|-------|------|----------|
| `eventName` | string | Yes |
| `promotionBannerImage` | string | Yes |
| `objectiveBannerImage` | string | Yes |
| `isPassPurchased` | boolean | Yes |
| `currentChapter` | LolTftPassChapter | Yes |
| `trackProgressNextReward` | LolTftPassTrackProgressNextReward | Yes |
| `trackProgress` | LolTftPassTrackProgressNextReward | Yes |
| `rewardTrackProgress` | LolTftPassRewardTrackProgress | Yes |

### LolTftPassRewardTrackProgress

| Field | Type | Required |
|-------|------|----------|
| `level` | int16 | Yes |
| `totalLevels` | int16 | Yes |
| `levelProgress` | uint16 | Yes |
| `futureLevelProgress` | uint16 | Yes |
| `passProgress` | int64 | Yes |
| `currentLevelXP` | int64 | Yes |
| `totalLevelXP` | int64 | Yes |
| `iteration` | uint32 | Yes |

### LolTftPassTFTPassRewardNotification

| Field | Type | Required |
|-------|------|----------|
| `title` | string | Yes |
| `description` | string | Yes |
| `iconURL` | string | Yes |
| `framedIcon` | boolean | Yes |

### LolTftPassTftBattlepass

| Field | Type | Required |
|-------|------|----------|
| `totalPointsEarned` | int32 | Yes |
| `milestones` | LolTftPassTftBattlepassMilestone[] | Yes |
| `bonuses` | LolTftPassTftBattlepassMilestone[] | Yes |
| `activeMilestone` | LolTftPassTftBattlepassMilestone | Yes |
| `info` | LolTftPassTftBattlepassInfo | Yes |
| `lastViewedProgress` | int32 | Yes |
| `lastViewedMilestone` | LolTftPassTftBattlepassMilestone | Yes |
| `currentLevel` | int32 | Yes |

### LolTftPassTftBattlepassInfo

| Field | Type | Required |
|-------|------|----------|
| `title` | string | Yes |
| `description` | string | Yes |
| `startDate` | uint64 | Yes |
| `endDate` | uint64 | Yes |
| `premium` | boolean | Yes |
| `premiumTitle` | string | Yes |
| `premiumEntitlementId` | string | Yes |
| `pcPurchaseRequirement` | string | Yes |
| `passId` | string | Yes |
| `hasLevelPurchasing` | boolean | Yes |
| `media` | object (map) | Yes |
| `passType` | LolTftPassTftPassType | Yes |

### LolTftPassTftBattlepassMilestone

| Field | Type | Required |
|-------|------|----------|
| `milestoneId` | string | Yes |
| `title` | string | Yes |
| `description` | string | Yes |
| `status` | string | Yes |
| `pointsNeededForMilestone` | int32 | Yes |
| `pointsEarnedForMilestone` | int32 | Yes |
| `totalPointsForMilestone` | int32 | Yes |
| `level` | int32 | Yes |
| `iconImageUrl` | string | Yes |
| `iconNeedsFrame` | boolean | Yes |
| `rewards` | LolTftPassTftBattlepassReward[] | Yes |
| `isPaid` | boolean | Yes |
| `isLocked` | boolean | Yes |
| `isKeystone` | boolean | Yes |
| `isBonus` | boolean | Yes |
| `isClaimRequestPending` | boolean | Yes |

### LolTftPassTrackProgressNextReward

| Field | Type | Required |
|-------|------|----------|
| `currentXP` | int64 | Yes |
| `nextLevelXP` | int64 | Yes |
| `currentLevel` | int64 | Yes |
| `nextReward` | LolTftPassNextRewardUIData | Yes |

### LolTftSkillTreeEventSkill

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `itemId` | uint8 | Yes |
| `name` | string | Yes |
| `description` | string | Yes |
| `texture` | string | Yes |
| `owned` | boolean | Yes |
| `equipped` | boolean | Yes |

### LolTftSkillTreeEventSkillPlayerProgression

| Field | Type | Required |
|-------|------|----------|
| `rank` | uint8 | Yes |
| `currDivision` | uint8 | Yes |
| `divisionEventPoints` | uint32 | Yes |
| `totalEventPoints` | uint32 | Yes |
| `lastViewedTotalEventPoints` | uint32 | Yes |

### LolTftSkillTreeEventSkillRankReward

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `name` | string | Yes |
| `description` | string | Yes |
| `texture` | string | Yes |

### LolTftSkillTreeEventSkillRankState

**Type**: enum string

**Values**: `kError`, `kClaimed`, `kClaiming`, `kClaimable`, `kUnclaimable`


### LolTftSkillTreeEventSkillTree

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `ranks` | LolTftSkillTreeEventSkillTreeRank[] | Yes |

### LolTftSkillTreeEventSkillTreeData

| Field | Type | Required |
|-------|------|----------|
| `eventSkillTree` | LolTftSkillTreeEventSkillTree | Yes |
| `playerProgression` | LolTftSkillTreeEventSkillPlayerProgression | Yes |

### LolTftSkillTreeEventSkillTreeRank

| Field | Type | Required |
|-------|------|----------|
| `state` | LolTftSkillTreeEventSkillRankState | Yes |
| `numDivisions` | uint8 | Yes |
| `totalEventPointsForRank` | uint32 | Yes |
| `rankId` | string | Yes |
| `name` | string | Yes |
| `texture` | string | Yes |
| `skills` | LolTftSkillTreeEventSkill[] | Yes |
| `rewards` | LolTftSkillTreeEventSkillRankReward[] | Yes |

### LolTftTFTRotatingShopPromos

| Field | Type | Required |
|-------|------|----------|
| `firstPromos` | LolTftTFTRotatingShopPromo[] | Yes |
| `secondPromos` | LolTftTFTRotatingShopPromo[] | Yes |
| `fallbackPromo` | LolTftTFTRotatingShopPromo | Yes |

### LolTftTeamPlannerImportedTeamCodeChampionData

| Field | Type | Required |
|-------|------|----------|
| `championId` | string | Yes |
| `icon` | string | Yes |
| `price` | uint32 | Yes |

### LolTftTeamPlannerImportedTeamCodeData

| Field | Type | Required |
|-------|------|----------|
| `teamCode` | string | Yes |
| `teamPlan` | LolTftTeamPlannerImportedTeamCodeChampionData[] | Yes |

### LolTftTeamPlannerPreviouslyUsedContext

| Field | Type | Required |
|-------|------|----------|
| `optionalTeamId` | string | Yes |
| `setId` | string | Yes |
| `sortOption` | uint8 | Yes |
| `viewId` | uint8 | Yes |

### LolTftTeamPlannerTFTTeamPlannerConfig

| Field | Type | Required |
|-------|------|----------|
| `enabled` | boolean | Yes |
| `multipleSetsEnabled` | boolean | Yes |
| `tencentNameCheckEnabled` | boolean | Yes |
| `globalNameSanitizationEnabled` | boolean | Yes |

### LolTftTrovesCapOrdersResponseDTO

| Field | Type | Required |
|-------|------|----------|
| `data` | object (map) | Yes |

### LolTftTrovesLootOddsResponse

| Field | Type | Required |
|-------|------|----------|
| `lootId` | string | Yes |
| `parentId` | string | Yes |
| `dropRate` | number | Yes |
| `quantity` | int32 | Yes |
| `label` | string | Yes |
| `query` | string | Yes |
| `displayPriority` | int32 | Yes |
| `children` | LolTftTrovesLootOddsResponse[] | Yes |

### LolTftTrovesPlayerLoot

| Field | Type | Required |
|-------|------|----------|
| `lootName` | string | Yes |
| `localizedName` | string | Yes |
| `itemDesc` | string | Yes |

### LolTftTrovesPurchaseRequest

| Field | Type | Required |
|-------|------|----------|
| `storeId` | string | Yes |
| `catalogEntryId` | string | Yes |
| `paymentOptionsKeys` | string[] | Yes |
| `quantity` | uint32 | Yes |

### LolTftTrovesTroves

| Field | Type | Required |
|-------|------|----------|
| `enabled` | boolean | Yes |
| `capCatalogEnabled` | boolean | Yes |
| `shoppeOddsTreeEnabled` | boolean | Yes |
| `bannerList` | LolTftTrovesTrovesActiveBanner[] |  |

### LolTftTrovesTrovesActiveBanner

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `version` | uint8 | Yes |
| `videoId` | string | Yes |

### LolTftTrovesTrovesBanner

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `version` | uint8 | Yes |
| `sourceId` | string | Yes |
| `startDate` | string | Yes |
| `endDate` | string | Yes |
| `pityLimit` | uint32 | Yes |
| `rollOffer` | string | Yes |
| `bannerTexture` | string | Yes |
| `thumbnailTexture` | string | Yes |
| `backgroundTexture` | string | Yes |
| `platformTexture` | string | Yes |
| `eventHubBannerTexture` | string | Yes |
| `name` | string | Yes |
| `description` | string | Yes |
| `isCollectorBounty` | boolean | Yes |
| `maxTotalRolls` | uint32 | Yes |
| `pullCost` | uint64 | Yes |
| `chaseContentIds` | string[] | Yes |
| `prestigeContentIds` | string[] | Yes |
| `celebrationTheme` | LolTftTrovesTrovesCelebrationThemeData | Yes |
| `status` | LolTftTrovesTrovesStatus | Yes |
| `bountyType` | string | Yes |
| `videoId` | string | Yes |
| `storeId` | string | Yes |
| `catalogEntryId` | string | Yes |
| `paymentOptionsKeys` | string[] | Yes |

### LolTftTrovesTrovesCelebrationThemeData

| Field | Type | Required |
|-------|------|----------|
| `currencySegmentData` | LolTftTrovesTrovesCelebrationCurrencySegmentData | Yes |
| `portalSegmentData` | LolTftTrovesTrovesCelebrationPortalSegmentData | Yes |
| `highlightSegmentData` | LolTftTrovesTrovesCelebrationHighlightSegmentData | Yes |
| `standardSegmentData` | LolTftTrovesTrovesCelebrationStandardSegmentData | Yes |

### LolTftTrovesTrovesMilestone

| Field | Type | Required |
|-------|------|----------|
| `milestoneId` | string | Yes |
| `currencyId` | string | Yes |
| `currencyAmount` | uint32 | Yes |
| `instanceId` | string | Yes |
| `triggerValue` | uint64 | Yes |
| `repeatSequence` | uint32 | Yes |
| `triggeredTimestamp` | string | Yes |
| `triggered` | boolean | Yes |
| `name` | string | Yes |
| `iconURL` | string | Yes |
| `resetValue` | uint16 | Yes |

### LolTftTrovesTrovesMilestoneCounter

| Field | Type | Required |
|-------|------|----------|
| `counterId` | string | Yes |
| `counterValue` | uint64 | Yes |
| `startTriggerValue` | uint16 | Yes |
| `increaseBy` | uint16 | Yes |
| `multiplier` | number | Yes |
| `resetValue` | uint16 | Yes |

### LolTftTrovesTrovesMilestones

| Field | Type | Required |
|-------|------|----------|
| `groupId` | string | Yes |
| `name` | string | Yes |
| `milestones` | LolTftTrovesTrovesMilestone[] | Yes |
| `counter` | LolTftTrovesTrovesMilestoneCounter | Yes |

### LolTftTrovesTrovesPurchaseRequest

| Field | Type | Required |
|-------|------|----------|
| `offerId` | string | Yes |
| `quantity` | uint32 | Yes |
| `paymentOption` | string | Yes |

### LolTftTrovesTrovesRollRequest

| Field | Type | Required |
|-------|------|----------|
| `offerId` | string | Yes |
| `numberOfRolls` | uint32 | Yes |
| `isMythic` | boolean | Yes |
| `dropTableId` | string | Yes |

### LolTftTrovesTrovesStatus

| Field | Type | Required |
|-------|------|----------|
| `ownedStatus` | LolTftTrovesTrovesOwnedStatus[] | Yes |
| `dropTableId` | string | Yes |
| `hasPullError` | boolean | Yes |
| `totalRollsCount` | uint16 | Yes |
| `isCollectorBountyMaxRollsMet` | boolean | Yes |

### LolTftTrovesVerboseLootOddsResponse

| Field | Type | Required |
|-------|------|----------|
| `recipeName` | string | Yes |
| `chanceToContain` | LolTftTrovesLootOddsResponse[] | Yes |
| `guaranteedToContain` | LolTftTrovesLootOddsResponse[] | Yes |
| `lootItem` | LolTftTrovesPlayerLoot | Yes |
| `hasPityRules` | boolean | Yes |
| `checksOwnership` | boolean | Yes |

---

## LolTrophies


### LolTrophiesTrophyProfileData

| Field | Type | Required |
|-------|------|----------|
| `theme` | string | Yes |
| `tier` | int64 | Yes |
| `bracket` | int64 | Yes |
| `seasonId` | int64 | Yes |
| `pedestal` | string | Yes |
| `cup` | string | Yes |
| `gem` | string | Yes |

---

## LolVanguard


### LolVanguardVanguardMachineSpecs

| Field | Type | Required |
|-------|------|----------|
| `tpm2Enabled` | boolean | Yes |
| `secureBootEnabled` | boolean | Yes |

### LolVanguardVanguardSession

| Field | Type | Required |
|-------|------|----------|
| `state` | LolVanguardVanguardSessionState | Yes |
| `vanguardStatus` | int32 | Yes |

### LolVanguardVanguardSessionState

**Type**: enum string

**Values**: `ERROR`, `CONNECTED`, `IN_PROGRESS`


### LolVanguardVanguardSystemCheckTelemetryEvent

| Field | Type | Required |
|-------|------|----------|
| `passedOsCheck` | boolean | Yes |
| `passedSecureFeaturesCheck` | boolean | Yes |

---

## LolYourshop


### LolYourshopPlayerPermissions

| Field | Type | Required |
|-------|------|----------|
| `useData` | string | Yes |

### LolYourshopPurchaseItem

| Field | Type | Required |
|-------|------|----------|
| `offerId` | string | Yes |
| `inventoryType` | string | Yes |
| `itemId` | int32 | Yes |
| `pricePaid` | int64 | Yes |
| `orderId` | string | Yes |

### LolYourshopPurchaseResponse

| Field | Type | Required |
|-------|------|----------|
| `items` | LolYourshopPurchaseItem[] | Yes |
| `wallet` | LolYourshopWallet | Yes |

### LolYourshopUIOffer

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `originalPrice` | int64 | Yes |
| `discountPrice` | int64 | Yes |
| `type` | string | Yes |
| `skinName` | string | Yes |
| `championId` | int32 | Yes |
| `skinId` | int32 | Yes |
| `owned` | boolean | Yes |
| `revealed` | boolean | Yes |
| `purchasing` | boolean | Yes |
| `expirationDate` | string | Yes |

### LolYourshopUIStatus

| Field | Type | Required |
|-------|------|----------|
| `hubEnabled` | boolean | Yes |
| `name` | string | Yes |
| `startTime` | string | Yes |
| `endTime` | string | Yes |

### LolYourshopWallet

| Field | Type | Required |
|-------|------|----------|
| `rp` | int64 | Yes |

---

## Other


### AlertDTO

| Field | Type | Required |
|-------|------|----------|
| `alertTime` | int64 | Yes |

### BracketMatch

| Field | Type | Required |
|-------|------|----------|
| `id` | int64 | Yes |
| `round` | int32 | Yes |
| `order` | int32 | Yes |
| `rosterId1` | int64 | Yes |
| `rosterId2` | int64 | Yes |
| `resultHistory` | string | Yes |
| `lowestPossiblePosition` | int32 | Yes |
| `highestPossiblePosition` | int32 | Yes |
| `roundStartTime` | int64 | Yes |
| `gameStartTime` | int64 | Yes |
| `status` | ClientBracketMatchStatus | Yes |
| `winnerId` | int64 | Yes |
| `gameId` | int64 | Yes |
| `loserBracket` | boolean | Yes |
| `forfeitRosterId` | int64 | Yes |
| `failRosterStatus` | int32 | Yes |

### BracketRoster

| Field | Type | Required |
|-------|------|----------|
| `rosterId` | int64 | Yes |
| `name` | string | Yes |
| `shortName` | string | Yes |
| `logo` | int32 | Yes |
| `logoColor` | int32 | Yes |

### BuildInfo

| Field | Type | Required |
|-------|------|----------|
| `branch` | string | Yes |
| `patchline` | string | Yes |
| `version` | string | Yes |
| `patchlineVisibleName` | string | Yes |
| `buildType` | string | Yes |

### ChampionMasteryPublicDTO

| Field | Type | Required |
|-------|------|----------|
| `championId` | int32 | Yes |
| `championLevel` | int32 | Yes |
| `championPoints` | int32 | Yes |

### ChampionScoutingDTO

| Field | Type | Required |
|-------|------|----------|
| `championId` | int32 | Yes |
| `winCount` | int32 | Yes |
| `gameCount` | int32 | Yes |
| `kda` | number | Yes |

### ClashEventData

| Field | Type | Required |
|-------|------|----------|
| `earnedDate` | string | Yes |
| `rewardType` | string | Yes |
| `tournamentId` | int64 | Yes |
| `tournamentName` | string | Yes |
| `tier` | string | Yes |
| `bracket` | int64 | Yes |
| `seasonId` | int32 | Yes |
| `theme` | string | Yes |
| `rosterId` | int64 | Yes |
| `teamName` | string | Yes |
| `teamShortName` | string | Yes |
| `teamLogoName` | string | Yes |
| `teamLogoChromaId` | string | Yes |
| `playerUUIDs` | string[] | Yes |
| `rewardSpec` | ClashRewardSpec | Yes |

### ClashRewardConfigClient

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `keyDef` | ClashRewardKeyType[] | Yes |
| `entries` | ClashRewardConfigEntry[] | Yes |

### ClashRewardDefinition

| Field | Type | Required |
|-------|------|----------|
| `rewardType` | ClashRewardType | Yes |
| `rewardSpec` | ClashRewardSpec | Yes |

### ClashRewardSpec

| Field | Type | Required |
|-------|------|----------|
| `pedestal` | string | Yes |
| `cup` | string | Yes |
| `gem` | string | Yes |
| `tier` | string | Yes |
| `bracket` | string | Yes |
| `theme` | string | Yes |
| `level` | string | Yes |
| `seasonId` | string | Yes |
| `name` | string | Yes |
| `quantity` | string | Yes |

### ClashSeasonRewardResult

| Field | Type | Required |
|-------|------|----------|
| `playerId` | uint64 | Yes |
| `seasonId` | int32 | Yes |
| `seasonVp` | int32 | Yes |
| `banned` | boolean | Yes |
| `honorLevel` | int32 | Yes |
| `eligible` | boolean | Yes |
| `rewards` | ClashRewardDefinition[] | Yes |

### ClientConfigConfigNamespaceUpdate

| Field | Type | Required |
|-------|------|----------|
| `public` | string[] | Yes |
| `player` | string[] | Yes |

### ClientConfigConfigReadinessEnum

**Type**: enum string

**Values**: `Disabled`, `Ready`, `NotReady`


### ClientConfigConfigStatus

| Field | Type | Required |
|-------|------|----------|
| `readiness` | ClientConfigConfigReadinessEnum | Yes |
| `updateId` | uint64 | Yes |

### ClientConfigEntitlements

| Field | Type | Required |
|-------|------|----------|
| `accessToken` | string | Yes |
| `token` | string | Yes |
| `subject` | string | Yes |
| `issuer` | string | Yes |
| `entitlements` | string[] | Yes |

### ClientConfigEntitlementsUpdate

| Field | Type | Required |
|-------|------|----------|
| `UpdateType` | ClientConfigUpdateType | Yes |
| `EntitlementsTokenResource` | ClientConfigEntitlements | Yes |

### ClientConfigUpdateType

**Type**: enum string

**Values**: `Delete`, `Update`, `Create`


### CrashReportingEnvironment

| Field | Type | Required |
|-------|------|----------|
| `environment` | string | Yes |
| `userName` | string | Yes |
| `userId` | string | Yes |

### DeepLinksDeepLinksSettings

| Field | Type | Required |
|-------|------|----------|
| `isSchemeReady` | boolean | Yes |
| `externalClientScheme` | string | Yes |
| `launchLorEnabled` | boolean | Yes |
| `launchLorUrl` | string | Yes |

### ElevationAction

**Type**: enum string

**Values**: `FixBrokenPermissions`


### ElevationRequest

| Field | Type | Required |
|-------|------|----------|
| `action` | ElevationAction | Yes |

### EntitlementsToken

| Field | Type | Required |
|-------|------|----------|
| `accessToken` | string | Yes |
| `token` | string | Yes |
| `subject` | string | Yes |
| `issuer` | string | Yes |
| `entitlements` | string[] | Yes |

### ErrorMonitorLogBatch

| Field | Type | Required |
|-------|------|----------|
| `logType` | ErrorMonitorLogType | Yes |
| `logEntries` | LogEntry[] | Yes |

### ErrorMonitorLogType

**Type**: enum string

**Values**: `UX`, `FOUNDATION`


### ExternalPluginsAvailability

**Type**: enum string

**Values**: `Error`, `Recovering`, `Connected`, `Preparing`, `NotAvailable`


### ExternalPluginsResource

| Field | Type | Required |
|-------|------|----------|
| `state` | ExternalPluginsAvailability | Yes |
| `errorString` | string | Yes |

### IdsDTO

| Field | Type | Required |
|-------|------|----------|
| `missionIds` | string[] | Yes |
| `seriesIds` | string[] | Yes |

### LogEntry

| Field | Type | Required |
|-------|------|----------|
| `severity` | LogSeverityLevels | Yes |
| `message` | string | Yes |

### LogEvent

| Field | Type | Required |
|-------|------|----------|
| `severity` | LogSeverityLevels | Yes |
| `message` | string | Yes |

### LogSeverityLevels

**Type**: enum string

**Values**: `Always`, `Error`, `Warning`, `Okay`


### LootLcdsRecipeClientDTO

| Field | Type | Required |
|-------|------|----------|
| `recipeName` | string | Yes |
| `type` | string | Yes |
| `displayCategories` | string | Yes |
| `crafterName` | string | Yes |
| `slots` | LootLcdsRecipeSlotClientDTO[] | Yes |
| `outputs` | LootLcdsRecipeOutputDTO[] | Yes |
| `metadata` | LootLcdsRecipeMetadata | Yes |
| `singleOpen` | boolean | Yes |

### LootLcdsRecipeMetadata

| Field | Type | Required |
|-------|------|----------|
| `guaranteedDescriptions` | LootLcdsLootDescriptionDTO[] | Yes |
| `bonusDescriptions` | LootLcdsLootDescriptionDTO[] | Yes |
| `tooltipsDisabled` | boolean | Yes |

### LootLcdsRecipeOutputDTO

| Field | Type | Required |
|-------|------|----------|
| `lootName` | string | Yes |
| `quantityExpression` | string | Yes |
| `probability` | number | Yes |
| `allowDuplicates` | boolean | Yes |

### LootLcdsRecipeSlotClientDTO

| Field | Type | Required |
|-------|------|----------|
| `slotNumber` | int32 | Yes |
| `query` | string | Yes |
| `quantityExpression` | string | Yes |

### MissionAlertDTO

| Field | Type | Required |
|-------|------|----------|
| `type` | string | Yes |
| `message` | string | Yes |
| `alertTime` | int64 | Yes |

### MissionDisplay

| Field | Type | Required |
|-------|------|----------|
| `attributes` | string[] | Yes |
| `locations` | string[] | Yes |

### MissionMetadata

| Field | Type | Required |
|-------|------|----------|
| `tutorial` | TutorialMetadata | Yes |
| `npeRewardPack` | NpeRewardPackMetadata | Yes |
| `missionType` | string | Yes |
| `weekNum` | int32 | Yes |
| `xpReward` | int32 | Yes |
| `chain` | int32 | Yes |
| `order` | int32 | Yes |
| `chainSize` | int32 | Yes |
| `minRequired` | uint32 |  |
| `objectiveMetadataMap` | object (map) |  |

### NotifyFailureRequest

| Field | Type | Required |
|-------|------|----------|
| `availabilityItemName` | string | Yes |
| `failureInfo` | string | Yes |

### NotifySuccessRequest

| Field | Type | Required |
|-------|------|----------|
| `availabilityItemName` | string | Yes |

### OpenedTeamDTO

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `shortName` | string | Yes |
| `logo` | int32 | Yes |
| `logoColor` | int32 | Yes |
| `invitationId` | string | Yes |
| `captainId` | uint64 | Yes |
| `tier` | int32 | Yes |
| `members` | OpenedTeamMemberDTO[] | Yes |
| `invitees` | PendingRosterInviteeDTO[] | Yes |
| `openPositions` | Position[] | Yes |

### OpenedTeamMemberDTO

| Field | Type | Required |
|-------|------|----------|
| `playerId` | int64 | Yes |
| `position` | Position | Yes |
| `tier` | int32 | Yes |
| `friendship` | int32 | Yes |

### PatcherComponentState

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `action` | PatcherComponentStateAction | Yes |
| `isUpToDate` | boolean | Yes |
| `isUpdateAvailable` | boolean | Yes |
| `timeOfLastUpToDateCheckISO8601` | string |  |
| `isCorrupted` | boolean | Yes |
| `progress` | PatcherComponentActionProgress |  |

### PatcherComponentStateAction

**Type**: enum string

**Values**: `Migrating`, `Repairing`, `Patching`, `CheckingForUpdates`, `Idle`


### PatcherNotification

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `notificationId` | PatcherNotificationId | Yes |
| `data` | object (map) | Yes |

### PatcherNotificationId

**Type**: enum string

**Values**: `BrokenPermissions`, `NotEnoughDiskSpace`, `DidRestoreClientBackup`, `FailedToWriteError`, `MissingFilesError`, `ConnectionError`, `UnspecifiedError`


### PatcherP2PStatus

| Field | Type | Required |
|-------|------|----------|
| `isEnabledForPatchline` | boolean | Yes |
| `isAllowedByUser` | boolean | Yes |
| `requiresRestart` | boolean | Yes |

### PatcherP2PStatusUpdate

| Field | Type | Required |
|-------|------|----------|
| `isAllowedByUser` | boolean | Yes |

### PatcherProductState

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `action` | PatcherComponentStateAction | Yes |
| `isUpToDate` | boolean | Yes |
| `isUpdateAvailable` | boolean | Yes |
| `isCorrupted` | boolean | Yes |
| `isStopped` | boolean | Yes |
| `percentPatched` | number | Yes |
| `components` | PatcherComponentState[] | Yes |

### PatcherStatus

| Field | Type | Required |
|-------|------|----------|
| `connectedToPatchServer` | boolean | Yes |
| `successfullyInstalledVersion` | uint32 |  |

### PatcherUxResource

| Field | Type | Required |
|-------|------|----------|
| `visible` | boolean | Yes |

### PaymentsFrontEndRequest

| Field | Type | Required |
|-------|------|----------|
| `isPrepaid` | boolean | Yes |
| `localeId` | string | Yes |
| `summonerLevel` | int16 | Yes |
| `gifteeAccountId` | string | Yes |
| `gifteeMessage` | string | Yes |
| `rsoToken` | string | Yes |
| `usePmcSessions` | boolean | Yes |
| `game` | string | Yes |
| `openedFrom` | string | Yes |
| `minVirtualAmount` | int32 | Yes |
| `orderDetailsJSON` | string | Yes |

### PaymentsFrontEndResult

| Field | Type | Required |
|-------|------|----------|
| `url` | string | Yes |

### PaymentsPaymentsTelemetryTransitions

**Type**: enum string

**Values**: `PMCCompleteToIdle`, `PMCClosedToIdle`, `PMCOpenToPMCComplete`, `PMCOpenToPMCClose`, `IdleToPMCOpen`


### PendingOpenedTeamDTO

| Field | Type | Required |
|-------|------|----------|
| `invitationId` | string | Yes |
| `name` | string | Yes |
| `shortName` | string | Yes |
| `logo` | int32 | Yes |
| `logoColor` | int32 | Yes |
| `tier` | int32 | Yes |

### PendingRosterInviteeDTO

| Field | Type | Required |
|-------|------|----------|
| `inviteeId` | uint64 | Yes |
| `inviteeState` | PendingRosterInviteeState | Yes |
| `inviter` | uint64 | Yes |
| `inviteTime` | int64 | Yes |
| `inviteType` | InviteType | Yes |

### PlayerFinderDTO

| Field | Type | Required |
|-------|------|----------|
| `playerId` | uint64 | Yes |
| `tier` | int32 | Yes |
| `primaryPos` | Position | Yes |
| `secondaryPos` | Position | Yes |
| `type` | PlayerFinderEnum | Yes |
| `friendId` | int64 | Yes |

### PlayerFinderEnum

**Type**: enum string

**Values**: `FRIEND`, `FREEAGENT`


### PlayerInventory

| Field | Type | Required |
|-------|------|----------|
| `wardSkins` | integer[] | Yes |
| `champions` | integer[] | Yes |
| `skins` | integer[] | Yes |
| `icons` | integer[] | Yes |
| `inventoryJwts` | string[] | Yes |

### PlayerMissionDTO

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `title` | string | Yes |
| `helperText` | string | Yes |
| `description` | string | Yes |
| `missionLineText` | string | Yes |
| `backgroundImageUrl` | string | Yes |
| `iconImageUrl` | string | Yes |
| `seriesName` | string | Yes |
| `locale` | string | Yes |
| `sequence` | int32 | Yes |
| `metadata` | MissionMetadata | Yes |
| `startTime` | int64 | Yes |
| `endTime` | int64 | Yes |
| `lastUpdatedTimestamp` | int64 | Yes |
| `objectives` | PlayerMissionObjectiveDTO[] | Yes |
| `rewards` | PlayerMissionRewardDTO[] | Yes |
| `expiringWarnings` | MissionAlertDTO[] | Yes |
| `requirements` | string[] | Yes |
| `rewardStrategy` | RewardStrategy | Yes |
| `display` | MissionDisplay | Yes |
| `completionExpression` | string | Yes |
| `viewed` | boolean | Yes |
| `isNew` | boolean | Yes |
| `status` | string | Yes |
| `missionType` | string | Yes |
| `displayType` | string | Yes |
| `earnedDate` | int64 | Yes |
| `completedDate` | int64 | Yes |
| `cooldownTimeMillis` | int64 | Yes |
| `celebrationType` | string | Yes |
| `clientNotifyLevel` | string | Yes |
| `internalName` | string | Yes |
| `media` | object (map) | Yes |

### PlayerMissionEligibilityData

| Field | Type | Required |
|-------|------|----------|
| `level` | int32 | Yes |
| `loyaltyEnabled` | boolean | Yes |
| `playerInventory` | PlayerInventory | Yes |
| `userInfoToken` | string |  |

### PlayerMissionObjectiveDTO

| Field | Type | Required |
|-------|------|----------|
| `type` | string | Yes |
| `description` | string | Yes |
| `progress` | MissionProgressDTO | Yes |
| `sequence` | int32 | Yes |
| `rewardGroups` | string[] | Yes |
| `hasObjectiveBasedReward` | boolean | Yes |
| `status` | string | Yes |
| `requirements` | string[] | Yes |

### PlayerMissionRewardDTO

| Field | Type | Required |
|-------|------|----------|
| `rewardType` | string | Yes |
| `rewardGroup` | string | Yes |
| `description` | string | Yes |
| `iconUrl` | string | Yes |
| `smallIconUrl` | string | Yes |
| `itemId` | string | Yes |
| `itemTypeId` | string | Yes |
| `uniqueName` | string | Yes |
| `rewardFulfilled` | boolean | Yes |
| `rewardGroupSelected` | boolean | Yes |
| `sequence` | int32 | Yes |
| `quantity` | int32 | Yes |
| `isObjectiveBasedReward` | boolean | Yes |
| `media` | object (map) | Yes |
| `iconNeedsFrame` | boolean | Yes |

### PlayerNotificationsPlayerNotificationConfigResource

| Field | Type | Required |
|-------|------|----------|
| `ExpirationCheckFrequency` | uint64 |  |

### PlayerNotificationsPlayerNotificationResource

| Field | Type | Required |
|-------|------|----------|
| `backgroundUrl` | string | Yes |
| `created` | string | Yes |
| `critical` | boolean | Yes |
| `data` | object (map) | Yes |
| `detailKey` | string | Yes |
| `expires` | string | Yes |
| `iconUrl` | string | Yes |
| `id` | uint64 | Yes |
| `source` | string | Yes |
| `state` | string | Yes |
| `titleKey` | string | Yes |
| `type` | string | Yes |
| `dismissible` | boolean | Yes |

### PlayerTierDTO

| Field | Type | Required |
|-------|------|----------|
| `playerId` | uint64 | Yes |
| `tier` | int32 | Yes |
| `primaryPos` | Position | Yes |
| `secondPos` | Position | Yes |

### PluginDescriptionResource

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `riotMeta` | PluginMetadataResource | Yes |
| `pluginDependencies` | string[] | Yes |

### PluginManagerResource

| Field | Type | Required |
|-------|------|----------|
| `state` | PluginManagerState | Yes |

### PluginManagerState

**Type**: enum string

**Values**: `PluginsInitialized`, `NotReady`


### PluginMetadataResource

| Field | Type | Required |
|-------|------|----------|
| `type` | string | Yes |
| `subtype` | string | Yes |
| `app` | string | Yes |
| `feature` | string | Yes |
| `mock` | string | Yes |
| `hasBundledAssets` | boolean | Yes |
| `globalAssetBundles` | string[] | Yes |
| `perLocaleAssetBundles` | object (map) | Yes |
| `implements` | string[] | Yes |
| `threading` | PluginThreadingModel | Yes |

### PluginResource

| Field | Type | Required |
|-------|------|----------|
| `fullName` | string | Yes |
| `shortName` | string | Yes |
| `supertype` | string | Yes |
| `subtype` | string | Yes |
| `app` | string | Yes |
| `feature` | string | Yes |
| `threadingModel` | string | Yes |
| `assetBundleNames` | string[] | Yes |
| `mountedAssetBundles` | object (map) | Yes |
| `orderWadFileMounted` | int32 | Yes |
| `dependencies` | PluginResourceContract[] | Yes |
| `implementedContracts` | PluginResourceContract[] | Yes |

### PluginResourceContract

| Field | Type | Required |
|-------|------|----------|
| `fullName` | string | Yes |

### Position

**Type**: enum string

**Values**: `UNSELECTED`, `FILL`, `UTILITY`, `JUNGLE`, `BOTTOM`, `MIDDLE`, `TOP`


### ProcessControlProcess

| Field | Type | Required |
|-------|------|----------|
| `status` | string | Yes |

### QueryEvaluationRequestDTO

| Field | Type | Required |
|-------|------|----------|
| `query` | string | Yes |

### RankedScoutingDTO

| Field | Type | Required |
|-------|------|----------|
| `playerId` | uint64 | Yes |
| `puuid` | string | Yes |
| `totalMasteryScore` | uint64 | Yes |
| `topMasteries` | ChampionMasteryPublicDTO[] | Yes |
| `topSeasonChampions` | ChampionScoutingDTO[] | Yes |

### RemotingSerializedFormat

**Type**: enum string

**Values**: `MsgPack`, `YAML`, `JSON`


### RewardLogo

| Field | Type | Required |
|-------|------|----------|
| `logo` | int32 | Yes |
| `memberOwnedCount` | int32 | Yes |

### RewardStrategy

| Field | Type | Required |
|-------|------|----------|
| `groupStrategy` | string | Yes |
| `selectMaxGroupCount` | uint16 | Yes |
| `selectMinGroupCount` | uint16 | Yes |

### RiotMessagingServiceEntitlementsToken

| Field | Type | Required |
|-------|------|----------|
| `accessToken` | string | Yes |
| `token` | string | Yes |
| `subject` | string | Yes |
| `issuer` | string | Yes |
| `entitlements` | string[] | Yes |

### RiotMessagingServiceSession

| Field | Type | Required |
|-------|------|----------|
| `state` | RiotMessagingServiceState | Yes |
| `token` | string | Yes |
| `tokenType` | RiotMessagingServiceTokenType | Yes |

### RiotMessagingServiceState

**Type**: enum string

**Values**: `Connected`, `Connecting`, `Disconnected`, `Disconnecting`


### RiotMessagingServiceTokenType

**Type**: enum string

**Values**: `Identity`, `Access`, `Unavailable`


### RmsMessage

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `resource` | string | Yes |
| `service` | string | Yes |
| `version` | string | Yes |
| `timestamp` | int64 | Yes |
| `payload` | string | Yes |
| `ackRequired` | boolean | Yes |

### RosterWithdraw

| Field | Type | Required |
|-------|------|----------|
| `initVoteTime` | int64 | Yes |
| `initVoteMember` | uint64 | Yes |
| `voteTimeoutMs` | int64 | Yes |
| `lockoutTimeMs` | int64 | Yes |
| `gameStartBufferMs` | int64 | Yes |
| `voteWithdrawMembers` | integer[] | Yes |
| `declineWithdrawMembers` | integer[] | Yes |

### SLIBoolDiagnostic

| Field | Type | Required |
|-------|------|----------|
| `key` | string | Yes |
| `value` | boolean | Yes |

### SLICount

| Field | Type | Required |
|-------|------|----------|
| `sliName` | string | Yes |
| `idempotencyKey` | string | Yes |
| `successes` | number | Yes |
| `failures` | number | Yes |
| `startTimeEpochMs` | int64 | Yes |
| `endTimeEpochMs` | int64 | Yes |
| `labels` | object (map) | Yes |
| `boolDiagnostics` | object (map) | Yes |
| `doubleDiagnostics` | object (map) | Yes |
| `intDiagnostics` | object (map) | Yes |
| `stringDiagnostics` | object (map) | Yes |

### SLIDoubleDiagnostic

| Field | Type | Required |
|-------|------|----------|
| `key` | string | Yes |
| `value` | number | Yes |

### SLIIntDiagnostic

| Field | Type | Required |
|-------|------|----------|
| `key` | string | Yes |
| `value` | int64 | Yes |

### SLILabel

| Field | Type | Required |
|-------|------|----------|
| `key` | string | Yes |
| `value` | string | Yes |

### SLIStringDiagnostic

| Field | Type | Required |
|-------|------|----------|
| `key` | string | Yes |
| `value` | string | Yes |

### SanitizerContainsSanitizedRequest

| Field | Type | Required |
|-------|------|----------|
| `text` | string | Yes |
| `level` | uint32 |  |
| `aggressiveScan` | boolean |  |

### SanitizerContainsSanitizedResponse

| Field | Type | Required |
|-------|------|----------|
| `contains` | boolean | Yes |

### SanitizerSanitizeRequest

| Field | Type | Required |
|-------|------|----------|
| `texts` | string[] |  |
| `text` | string |  |
| `level` | uint32 |  |
| `aggressiveScan` | boolean |  |

### SanitizerSanitizeResponse

| Field | Type | Required |
|-------|------|----------|
| `texts` | string[] |  |
| `text` | string |  |
| `modified` | boolean | Yes |

### SanitizerSanitizerStatus

| Field | Type | Required |
|-------|------|----------|
| `ready` | boolean | Yes |
| `region` | string | Yes |
| `locale` | string | Yes |
| `filteredWordCountsByLevel` | object (map) | Yes |
| `whitelistedWordCountsByLevel` | object (map) | Yes |
| `breakingCharsCount` | uint32 | Yes |
| `projectedCharsCount` | uint32 | Yes |

### SeriesDTO

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `internalName` | string | Yes |
| `parentInternalName` | string | Yes |
| `type` | string | Yes |
| `eligibilityType` | string | Yes |
| `displayType` | string | Yes |
| `title` | string | Yes |
| `description` | string | Yes |
| `optInButtonText` | string | Yes |
| `optOutButtonText` | string | Yes |
| `status` | string | Yes |
| `startDate` | int64 | Yes |
| `endDate` | int64 | Yes |
| `createdDate` | int64 | Yes |
| `lastUpdatedTimestamp` | int64 | Yes |
| `viewed` | boolean | Yes |
| `media` | SeriesMediaDTO | Yes |
| `tags` | string[] | Yes |
| `warnings` | AlertDTO[] | Yes |

### SeriesMediaDTO

| Field | Type | Required |
|-------|------|----------|
| `backgroundUrl` | string | Yes |
| `backgroundImageLargeUrl` | string | Yes |
| `backgroundImageSmallUrl` | string | Yes |
| `trackerIconUrl` | string | Yes |
| `trackerIcon` | string | Yes |
| `accentColor` | string | Yes |

### TicketType

**Type**: enum string

**Values**: `PREMIUM`, `BASIC`


### TierConfig

| Field | Type | Required |
|-------|------|----------|
| `tier` | int32 | Yes |
| `delayTime` | int64 | Yes |
| `estimateTime` | int64 | Yes |

### TournamentDTO

| Field | Type | Required |
|-------|------|----------|
| `id` | int64 | Yes |
| `scheduleTime` | int64 | Yes |
| `scheduleEndTime` | int64 | Yes |
| `rosterCreateDeadline` | int64 | Yes |
| `entryFee` | int32 | Yes |
| `rosterSize` | int32 | Yes |
| `themeId` | int32 | Yes |
| `nameLocKey` | string | Yes |
| `nameLocKeySecondary` | string | Yes |
| `buyInOptions` | integer[] | Yes |
| `buyInOptionsPremium` | integer[] | Yes |
| `queueId` | int32 | Yes |
| `scoutingTimeMs` | int64 | Yes |
| `lastThemeOfSeason` | boolean | Yes |
| `bracketSize` | string | Yes |
| `minGames` | int32 | Yes |
| `smsRestriction` | boolean | Yes |
| `honorRestriction` | boolean | Yes |
| `rankRestriction` | boolean | Yes |
| `voiceEnabled` | boolean | Yes |
| `phases` | TournamentPhaseDTO[] | Yes |
| `rewardConfig` | ClashRewardConfigClient[] | Yes |
| `tierConfigs` | TierConfig[] | Yes |
| `bracketFormationInitDelayMs` | int64 | Yes |
| `bracketFormationIntervalMs` | int64 | Yes |
| `status` | TournamentStatusEnum | Yes |
| `resumeTime` | int64 | Yes |
| `lft` | boolean | Yes |
| `maxInvites` | int32 | Yes |
| `maxSuggestionsPerPlayer` | int32 | Yes |

### TournamentPhaseDTO

| Field | Type | Required |
|-------|------|----------|
| `id` | int64 | Yes |
| `tournamentId` | int64 | Yes |
| `registrationTime` | int64 | Yes |
| `startTime` | int64 | Yes |
| `period` | int32 | Yes |
| `cancelled` | boolean | Yes |
| `limitTiers` | integer[] | Yes |
| `capacityStatus` | CapacityEnum | Yes |

### TournamentStatusEnum

**Type**: enum string

**Values**: `PRERESUME`, `PAUSED`, `CANCELLED`, `DEFAULT`


### TracingCriticalFlowEventV1

| Field | Type | Required |
|-------|------|----------|
| `when` | uint64 | Yes |
| `eventId` | string | Yes |
| `succeeded` | boolean | Yes |
| `payloadString` | string | Yes |

### TracingEventV1

| Field | Type | Required |
|-------|------|----------|
| `when` | uint64 | Yes |
| `name` | string | Yes |
| `src` | string | Yes |
| `dest` | string | Yes |
| `tags` | string | Yes |
| `details` | string | Yes |

### TracingModuleThreadingModelV1

**Type**: enum string

**Values**: `kParallel`, `kConcurrent`, `kSequential`, `kDedicated`, `kMainThread`, `kNone`


### TracingModuleTypeV1

**Type**: enum string

**Values**: `kRemotingSource`, `kFrontEndPlugin`, `kBackendOther`, `kBackEndPlugin`, `kRemoteAppModule`, `kUnknown`


### TracingModuleV1

| Field | Type | Required |
|-------|------|----------|
| `moduleId` | uint32 | Yes |
| `name` | string | Yes |
| `type` | TracingModuleTypeV1 | Yes |
| `threadingModel` | TracingModuleThreadingModelV1 | Yes |

### TracingPhaseBeginV1

| Field | Type | Required |
|-------|------|----------|
| `when` | uint64 | Yes |
| `name` | string | Yes |
| `importance` | TracingPhaseImportanceV1 | Yes |

### TracingPhaseEndV1

| Field | Type | Required |
|-------|------|----------|
| `when` | uint64 | Yes |
| `name` | string | Yes |

### TracingPhaseImportanceV1

**Type**: enum string

**Values**: `major`, `minor`, `trivial`


### basicOperatingSystemInfo

| Field | Type | Required |
|-------|------|----------|
| `edition` | string | Yes |
| `platform` | string | Yes |
| `versionMajor` | string | Yes |
| `versionMinor` | string | Yes |
| `buildNumber` | string | Yes |

### basicSystemInfo

| Field | Type | Required |
|-------|------|----------|
| `operatingSystem` | basicOperatingSystemInfo | Yes |
| `physicalMemory` | uint64 | Yes |
| `physicalProcessorCores` | uint64 | Yes |

### cookie

| Field | Type | Required |
|-------|------|----------|
| `url` | string | Yes |
| `name` | string | Yes |
| `value` | string | Yes |
| `domain` | string | Yes |
| `path` | string | Yes |
| `secure` | boolean | Yes |
| `httponly` | boolean | Yes |
| `expires` | int64 |  |

---

## TeamBuilderDirect


### TeamBuilderDirect-BenchChampion

| Field | Type | Required |
|-------|------|----------|
| `championId` | int32 | Yes |
| `isPriority` | boolean | Yes |

### TeamBuilderDirect-ChampGridChampion

| Field | Type | Required |
|-------|------|----------|
| `id` | int32 | Yes |
| `name` | string | Yes |
| `squarePortraitPath` | string | Yes |
| `freeToPlay` | boolean | Yes |
| `loyaltyReward` | boolean | Yes |
| `xboxGPReward` | boolean | Yes |
| `freeToPlayForQueue` | boolean | Yes |
| `owned` | boolean | Yes |
| `rented` | boolean | Yes |
| `disabled` | boolean | Yes |
| `roles` | string[] | Yes |
| `masteryPoints` | int32 | Yes |
| `masteryLevel` | int32 | Yes |
| `selectionStatus` | TeamBuilderDirect-ChampionSelection | Yes |
| `positionsFavorited` | string[] | Yes |

### TeamBuilderDirect-ChampSelectAction

| Field | Type | Required |
|-------|------|----------|
| `id` | int64 | Yes |
| `actorCellId` | int64 | Yes |
| `championId` | int32 | Yes |
| `type` | string | Yes |
| `completed` | boolean | Yes |
| `isAllyAction` | boolean | Yes |
| `isInProgress` | boolean | Yes |
| `pickTurn` | int32 | Yes |
| `duration` | int64 | Yes |

### TeamBuilderDirect-ChampSelectBannedChampions

| Field | Type | Required |
|-------|------|----------|
| `myTeamBans` | integer[] | Yes |
| `theirTeamBans` | integer[] | Yes |
| `numBans` | int32 | Yes |

### TeamBuilderDirect-ChampSelectChampionSwapNotification

| Field | Type | Required |
|-------|------|----------|
| `id` | int64 | Yes |
| `responderIndex` | int64 | Yes |
| `state` | TeamBuilderDirect-ChampSelectSwapState | Yes |
| `otherSummonerIndex` | int64 | Yes |
| `responderChampionName` | string | Yes |
| `requesterChampionName` | string | Yes |
| `requesterChampionSplashPath` | string | Yes |
| `initiatedByLocalPlayer` | boolean | Yes |
| `type` | TeamBuilderDirect-ChampSelectSwapType | Yes |
| `requesterChampionId` | int32 | Yes |

### TeamBuilderDirect-ChampSelectChatRoomDetails

| Field | Type | Required |
|-------|------|----------|
| `multiUserChatId` | string | Yes |
| `multiUserChatPassword` | string | Yes |
| `mucJwtDto` | TeamBuilderDirect-MucJwtDto | Yes |

### TeamBuilderDirect-ChampSelectMySelection

| Field | Type | Required |
|-------|------|----------|
| `selectedSkinId` | int32 |  |
| `spell1Id` | uint64 |  |
| `spell2Id` | uint64 |  |

### TeamBuilderDirect-ChampSelectPickOrderSwapNotification

| Field | Type | Required |
|-------|------|----------|
| `id` | int64 | Yes |
| `requestorIndex` | int64 | Yes |
| `responderIndex` | int64 | Yes |
| `state` | TeamBuilderDirect-ChampSelectSwapState | Yes |
| `otherSummonerIndex` | int64 | Yes |
| `initiatedByLocalPlayer` | boolean | Yes |
| `type` | TeamBuilderDirect-ChampSelectSwapType | Yes |

### TeamBuilderDirect-ChampSelectPinDropNotification

| Field | Type | Required |
|-------|------|----------|
| `pinDropSummoners` | TeamBuilderDirect-ChampSelectPinDropSummoner[] | Yes |
| `mapSide` | string | Yes |

### TeamBuilderDirect-ChampSelectPinDropSummoner

| Field | Type | Required |
|-------|------|----------|
| `slotId` | uint64 | Yes |
| `position` | string | Yes |
| `lane` | string | Yes |
| `lanePosition` | uint64 | Yes |
| `isLocalSummoner` | boolean | Yes |
| `isPlaceholder` | boolean | Yes |
| `isAutofilled` | boolean | Yes |

### TeamBuilderDirect-ChampSelectPlayerSelection

| Field | Type | Required |
|-------|------|----------|
| `cellId` | int64 | Yes |
| `championId` | int32 | Yes |
| `selectedSkinId` | int32 | Yes |
| `wardSkinId` | int64 | Yes |
| `spell1Id` | uint64 | Yes |
| `spell2Id` | uint64 | Yes |
| `team` | int32 | Yes |
| `assignedPosition` | string | Yes |
| `championPickIntent` | int32 | Yes |
| `playerType` | string | Yes |
| `summonerId` | uint64 | Yes |
| `gameName` | string | Yes |
| `tagLine` | string | Yes |
| `puuid` | string | Yes |
| `isHumanoid` | boolean | Yes |
| `nameVisibilityType` | string | Yes |
| `playerAlias` | string | Yes |
| `obfuscatedSummonerId` | uint64 | Yes |
| `obfuscatedPuuid` | string | Yes |
| `isAutofilled` | boolean | Yes |
| `internalName` | string | Yes |
| `pickMode` | int32 | Yes |
| `pickTurn` | int32 | Yes |

### TeamBuilderDirect-ChampSelectPositionSwapNotification

| Field | Type | Required |
|-------|------|----------|
| `id` | int64 | Yes |
| `requestorIndex` | int64 | Yes |
| `responderIndex` | int64 | Yes |
| `requesterPosition` | string | Yes |
| `responderPosition` | string | Yes |
| `state` | TeamBuilderDirect-ChampSelectSwapState | Yes |
| `otherSummonerIndex` | int64 | Yes |
| `initiatedByLocalPlayer` | boolean | Yes |
| `type` | TeamBuilderDirect-ChampSelectSwapType | Yes |

### TeamBuilderDirect-ChampSelectSession

| Field | Type | Required |
|-------|------|----------|
| `id` | string | Yes |
| `gameId` | uint64 | Yes |
| `queueId` | int32 | Yes |
| `timer` | TeamBuilderDirect-TeambuilderDirectTypes-ChampSelectTimer | Yes |
| `chatDetails` | TeamBuilderDirect-ChampSelectChatRoomDetails | Yes |
| `myTeam` | TeamBuilderDirect-ChampSelectPlayerSelection[] | Yes |
| `theirTeam` | TeamBuilderDirect-ChampSelectPlayerSelection[] | Yes |
| `trades` | TeamBuilderDirect-ChampSelectSwapContract[] | Yes |
| `pickOrderSwaps` | TeamBuilderDirect-ChampSelectSwapContract[] | Yes |
| `positionSwaps` | TeamBuilderDirect-ChampSelectSwapContract[] | Yes |
| `actions` | object[] | Yes |
| `bans` | TeamBuilderDirect-ChampSelectBannedChampions | Yes |
| `localPlayerCellId` | int64 | Yes |
| `isSpectating` | boolean | Yes |
| `allowSkinSelection` | boolean | Yes |
| `allowSubsetChampionPicks` | boolean | Yes |
| `allowDuplicatePicks` | boolean | Yes |
| `allowPlayerPickSameChampion` | boolean | Yes |
| `disallowBanningTeammateHoveredChampions` | boolean | Yes |
| `allowBattleBoost` | boolean | Yes |
| `boostableSkinCount` | int32 | Yes |
| `allowRerolling` | boolean | Yes |
| `rerollsRemaining` | uint64 | Yes |
| `allowLockedEvents` | boolean | Yes |
| `lockedEventIndex` | int32 | Yes |
| `benchEnabled` | boolean | Yes |
| `benchChampions` | TeamBuilderDirect-BenchChampion[] | Yes |
| `counter` | int64 | Yes |
| `skipChampionSelect` | boolean | Yes |
| `hasSimultaneousBans` | boolean | Yes |
| `hasSimultaneousPicks` | boolean | Yes |
| `showQuitButton` | boolean | Yes |
| `isLegacyChampSelect` | boolean | Yes |
| `isCustomGame` | boolean | Yes |

### TeamBuilderDirect-ChampSelectSummoner

| Field | Type | Required |
|-------|------|----------|
| `cellId` | int64 | Yes |
| `slotId` | uint64 | Yes |
| `spell1IconPath` | string | Yes |
| `spell2IconPath` | string | Yes |
| `assignedPosition` | string | Yes |
| `summonerId` | uint64 | Yes |
| `gameName` | string | Yes |
| `tagLine` | string | Yes |
| `puuid` | string | Yes |
| `isHumanoid` | boolean | Yes |
| `nameVisibilityType` | string | Yes |
| `obfuscatedSummonerId` | uint64 | Yes |
| `obfuscatedPuuid` | string | Yes |
| `activeActionType` | string | Yes |
| `championIconStyle` | string | Yes |
| `skinSplashPath` | string | Yes |
| `actingBackgroundAnimationState` | string | Yes |
| `statusMessageKey` | string | Yes |
| `championId` | int32 | Yes |
| `championName` | string | Yes |
| `pickSnipedClass` | string | Yes |
| `currentChampionVotePercentInteger` | int32 | Yes |
| `skinId` | int32 | Yes |
| `banIntentChampionId` | int32 | Yes |
| `isOnPlayersTeam` | boolean | Yes |
| `shouldShowSelectedSkin` | boolean | Yes |
| `shouldShowExpanded` | boolean | Yes |
| `isActingNow` | boolean | Yes |
| `shouldShowActingBar` | boolean | Yes |
| `isSelf` | boolean | Yes |
| `shouldShowBanIntentIcon` | boolean | Yes |
| `isPickIntenting` | boolean | Yes |
| `isDonePicking` | boolean | Yes |
| `isPlaceholder` | boolean | Yes |
| `shouldShowSpells` | boolean | Yes |
| `shouldShowRingAnimations` | boolean | Yes |
| `areSummonerActionsComplete` | boolean | Yes |
| `tradeId` | int64 | Yes |
| `swapId` | int64 | Yes |
| `positionSwapId` | int64 | Yes |
| `showTrades` | boolean | Yes |
| `showSwaps` | boolean | Yes |
| `showPositionSwaps` | boolean | Yes |
| `showMuted` | boolean | Yes |
| `isAutofilled` | boolean | Yes |

### TeamBuilderDirect-ChampSelectSwapContract

| Field | Type | Required |
|-------|------|----------|
| `id` | int64 | Yes |
| `cellId` | int64 | Yes |
| `state` | TeamBuilderDirect-ChampSelectSwapState | Yes |

### TeamBuilderDirect-ChampSelectSwapState

**Type**: enum string

**Values**: `ACCEPTED`, `CANCELLED`, `DECLINED`, `SENT`, `RECEIVED`, `INVALID`, `BUSY`, `AVAILABLE`


### TeamBuilderDirect-ChampSelectSwapType

**Type**: enum string

**Values**: `POSITION`, `PICK_ORDER`, `CHAMPION`


### TeamBuilderDirect-ChampionSelectPreferences

| Field | Type | Required |
|-------|------|----------|
| `skins` | object (map) | Yes |
| `spells` | object (map) | Yes |

### TeamBuilderDirect-ChampionSelection

| Field | Type | Required |
|-------|------|----------|
| `selectedByMe` | boolean | Yes |
| `banIntentedByMe` | boolean | Yes |
| `banIntented` | boolean | Yes |
| `isBanned` | boolean | Yes |
| `pickIntented` | boolean | Yes |
| `pickIntentedByMe` | boolean | Yes |
| `pickIntentedPosition` | string | Yes |
| `pickedByOtherOrBanned` | boolean | Yes |

### TeamBuilderDirect-CollectionsChampionSkinEmblem

| Field | Type | Required |
|-------|------|----------|
| `name` | string | Yes |
| `emblemPath` | TeamBuilderDirect-CollectionsChampionSkinEmblemPath | Yes |
| `positions` | TeamBuilderDirect-CollectionsChampionSkinEmblemPosition | Yes |

### TeamBuilderDirect-CollectionsOwnership

| Field | Type | Required |
|-------|------|----------|
| `loyaltyReward` | boolean | Yes |
| `xboxGPReward` | boolean | Yes |
| `owned` | boolean | Yes |
| `rental` | TeamBuilderDirect-CollectionsRental | Yes |

### TeamBuilderDirect-MatchmakingDodgeData

| Field | Type | Required |
|-------|------|----------|
| `state` | TeamBuilderDirect-MatchmakingDodgeState | Yes |
| `dodgerId` | int64 | Yes |

### TeamBuilderDirect-MatchmakingLowPriorityData

| Field | Type | Required |
|-------|------|----------|
| `penalizedSummonerIds` | integer[] | Yes |
| `penaltyTime` | number | Yes |
| `penaltyTimeRemaining` | number | Yes |
| `bustedLeaverAccessToken` | string | Yes |

### TeamBuilderDirect-MatchmakingReadyCheckResource

| Field | Type | Required |
|-------|------|----------|
| `state` | TeamBuilderDirect-MatchmakingReadyCheckState | Yes |
| `playerResponse` | TeamBuilderDirect-MatchmakingReadyCheckResponse | Yes |
| `dodgeWarning` | TeamBuilderDirect-MatchmakingDodgeWarning | Yes |
| `timer` | number | Yes |
| `declinerIds` | integer[] | Yes |
| `autoAccept` | boolean | Yes |

### TeamBuilderDirect-MatchmakingSearchErrorResource

| Field | Type | Required |
|-------|------|----------|
| `id` | int32 | Yes |
| `errorType` | string | Yes |
| `penalizedSummonerId` | int64 | Yes |
| `penaltyTimeRemaining` | number | Yes |
| `message` | string | Yes |

### TeamBuilderDirect-MatchmakingSearchResource

| Field | Type | Required |
|-------|------|----------|
| `queueId` | int32 | Yes |
| `isCurrentlyInQueue` | boolean | Yes |
| `lobbyId` | string | Yes |
| `searchState` | TeamBuilderDirect-MatchmakingSearchState | Yes |
| `timeInQueue` | number | Yes |
| `estimatedQueueTime` | number | Yes |
| `readyCheck` | TeamBuilderDirect-MatchmakingReadyCheckResource | Yes |
| `dodgeData` | TeamBuilderDirect-MatchmakingDodgeData | Yes |
| `lowPriorityData` | TeamBuilderDirect-MatchmakingLowPriorityData | Yes |
| `errors` | TeamBuilderDirect-MatchmakingSearchErrorResource[] | Yes |

### TeamBuilderDirect-MatchmakingSearchState

**Type**: enum string

**Values**: `ServiceShutdown`, `ServiceError`, `Error`, `Found`, `Searching`, `Canceled`, `AbandonedLowPriorityQueue`, `Invalid`


### TeamBuilderDirect-MutedPlayerInfo

| Field | Type | Required |
|-------|------|----------|
| `puuid` | string | Yes |
| `summonerId` | uint64 | Yes |
| `obfuscatedPuuid` | string | Yes |
| `obfuscatedSummonerId` | uint64 | Yes |

### TeamBuilderDirect-QuestSkinProductType

**Type**: enum string

**Values**: `kTieredSkin`, `kQuestSkin`


### TeamBuilderDirect-SfxNotification

| Field | Type | Required |
|-------|------|----------|
| `delayMillis` | int64 | Yes |
| `path` | string | Yes |
| `eventType` | string | Yes |

### TeamBuilderDirect-SkinSelectorChildSkin

| Field | Type | Required |
|-------|------|----------|
| `championId` | int32 | Yes |
| `chromaPreviewPath` | string |  |
| `id` | int32 | Yes |
| `name` | string | Yes |
| `ownership` | TeamBuilderDirect-CollectionsOwnership | Yes |
| `isBase` | boolean | Yes |
| `disabled` | boolean | Yes |
| `stillObtainable` | boolean | Yes |
| `isChampionUnlocked` | boolean | Yes |
| `splashPath` | string | Yes |
| `splashVideoPath` | string |  |
| `tilePath` | string | Yes |
| `unlocked` | boolean | Yes |
| `skinAugments` | object (map) | Yes |
| `parentSkinId` | int32 | Yes |
| `colors` | string[] | Yes |
| `stage` | uint64 | Yes |
| `shortName` | string | Yes |

### TeamBuilderDirect-SkinSelectorInfo

| Field | Type | Required |
|-------|------|----------|
| `selectedSkinId` | int32 | Yes |
| `isSkinGrantedFromBoost` | boolean | Yes |
| `selectedChampionId` | int32 | Yes |
| `championName` | string | Yes |
| `skinSelectionDisabled` | boolean | Yes |
| `showSkinSelector` | boolean | Yes |

### TeamBuilderDirect-SkinSelectorSkin

| Field | Type | Required |
|-------|------|----------|
| `championId` | int32 | Yes |
| `chromaPreviewPath` | string |  |
| `id` | int32 | Yes |
| `name` | string | Yes |
| `ownership` | TeamBuilderDirect-CollectionsOwnership | Yes |
| `isBase` | boolean | Yes |
| `disabled` | boolean | Yes |
| `stillObtainable` | boolean | Yes |
| `isChampionUnlocked` | boolean | Yes |
| `splashPath` | string | Yes |
| `splashVideoPath` | string |  |
| `tilePath` | string | Yes |
| `unlocked` | boolean | Yes |
| `skinAugments` | object (map) | Yes |
| `childSkins` | TeamBuilderDirect-SkinSelectorChildSkin[] | Yes |
| `emblems` | TeamBuilderDirect-CollectionsChampionSkinEmblem[] | Yes |
| `rarityGemPath` | string | Yes |
| `groupSplash` | string | Yes |
| `productType` | TeamBuilderDirect-QuestSkinProductType |  |

### TeamBuilderDirect-TeamBoost

| Field | Type | Required |
|-------|------|----------|
| `activatorCellId` | int64 | Yes |
| `skinUnlockMode` | string | Yes |
| `price` | int64 | Yes |
| `ipReward` | int64 | Yes |
| `ipRewardForPurchaser` | int64 | Yes |
| `availableSkins` | integer[] | Yes |
| `unlocked` | boolean | Yes |

### TeamBuilderDirect-TeambuilderDirectTypes-ChampSelectTimer

| Field | Type | Required |
|-------|------|----------|
| `adjustedTimeLeftInPhase` | int64 | Yes |
| `totalTimeInPhase` | int64 | Yes |
| `phase` | string | Yes |
| `isInfinite` | boolean | Yes |
| `internalNowInEpochMs` | uint64 | Yes |