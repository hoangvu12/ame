export const SKIN_SELECTORS = [
  '.skin-name-text', // Classic Champ Select
  '.skin-name',      // Swiftplay lobby
];
export const POLL_INTERVAL_MS = 300;
export const PREFETCH_DEBOUNCE_MS = 2000;
export const CHAMP_SELECT_PHASES = ['ChampSelect'];
export const POST_GAME_PHASES = ['None', 'Lobby', 'EndOfGame', 'PreEndOfGame', 'Matchmaking', 'ReadyCheck'];
export const WS_URL = 'ws://localhost:18765';
export const WS_RECONNECT_BASE_MS = 1000;
export const WS_RECONNECT_MAX_MS = 30000;
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
