export const SKIN_SELECTORS = [
  '.skin-name-text', // Classic Champ Select
  '.skin-name',      // Swiftplay lobby
];
export const POLL_INTERVAL_MS = 300;
export const PREFETCH_DEBOUNCE_MS = 2000;
export const OWNED_SELECTION_DELAY_MS = 150;
export const CHAMP_SELECT_PHASES = ['ChampSelect'];
// The LCU registers its routes asynchronously, so the initial phase read can
// fail for a few seconds after the plugin loads.
export const PHASE_SEED_BASE_MS = 300;
export const PHASE_SEED_MAX_MS = 2000;
export const PHASE_SEED_MAX_RETRIES = 8;
export const POST_GAME_PHASES = ['None', 'Lobby', 'EndOfGame', 'PreEndOfGame', 'Matchmaking', 'ReadyCheck'];
export const WS_URL = 'ws://localhost:18765';
export const WS_RECONNECT_BASE_MS = 1000;
// Capped at 5s, not 30s. This is a loopback connect, so retrying costs
// essentially nothing, and a 30s cap meant ame could be running for half a
// champ select before the plugin noticed it.
export const WS_RECONNECT_MAX_MS = 5000;
// Spread reconnects so several tabs/reloads don't retry in lockstep.
export const WS_RECONNECT_JITTER_MS = 400;
export const BUTTON_ID = 'ame-apply-btn';
export const CHROMA_BTN_CLASS = 'ame-chroma-button';
export const CHROMA_PANEL_ID = 'ame-chroma-panel-container';
export const CONNECTION_BANNER_ID = 'ame-connection-banner';
export const IN_GAME_PHASES = ['InProgress', 'Reconnect'];
export const IN_GAME_CONTAINER_ID = 'ame-ingame-container';
export const IN_GAME_POLL_MS = 500;
export const AUTO_ACCEPT_DELAY_MS = 2000;
export const AUTO_SELECT_DELAY_MS = 1500;
export const AUTO_SELECT_ROLES = [
  { key: 'top', labelKey: 'roles.top', icon: '/fe/lol-parties/icon-position-top.png' },
  { key: 'jungle', labelKey: 'roles.jungle', icon: '/fe/lol-parties/icon-position-jungle.png' },
  { key: 'middle', labelKey: 'roles.middle', icon: '/fe/lol-parties/icon-position-middle.png' },
  { key: 'bottom', labelKey: 'roles.bottom', icon: '/fe/lol-parties/icon-position-bottom.png' },
  { key: 'utility', labelKey: 'roles.utility', icon: '/fe/lol-parties/icon-position-utility.png' },
];
export const CHAT_AVAILABILITY_OPTIONS = [
  { value: '', labelKey: 'chat_status.default' },
  { value: 'chat', labelKey: 'chat_status.online' },
  { value: 'away', labelKey: 'chat_status.away' },
  { value: 'dnd', labelKey: 'chat_status.busy' },
  { value: 'mobile', labelKey: 'chat_status.mobile' },
  { value: 'offline', labelKey: 'chat_status.offline' },
];
export const SWIFTPLAY_BUTTON_ID = 'ame-swiftplay-apply-btn';
export const STYLE_ID = 'ame-styles';
export const ROOM_PARTY_INDICATOR_CLASS = 'ame-room-party-indicator';
export const CUSTOM_SKINS_BTN_ID = 'ame-custom-skins-btn';
export const CUSTOM_SKINS_MODAL_ID = 'ame-custom-skins-modal';
export const CUSTOM_SKINS_IMAGE_BASE = 'http://localhost:18765/custom-mod-image/';
export const PROXY_IMAGE_BASE = 'http://localhost:18765/proxy-image?url=';
