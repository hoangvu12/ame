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
export const IN_GAME_PHASES = ['InProgress', 'Reconnect'];
export const IN_GAME_CONTAINER_ID = 'ame-ingame-container';
export const IN_GAME_POLL_MS = 500;
export const AUTO_ACCEPT_DELAY_MS = 2000;
export const AUTO_SELECT_DELAY_MS = 1500;
export const AUTO_SELECT_ROLES = [
  { key: 'top', label: 'Top', icon: '/fe/lol-parties/icon-position-top.png' },
  { key: 'jungle', label: 'Jungle', icon: '/fe/lol-parties/icon-position-jungle.png' },
  { key: 'middle', label: 'Mid', icon: '/fe/lol-parties/icon-position-middle.png' },
  { key: 'bottom', label: 'Bot', icon: '/fe/lol-parties/icon-position-bottom.png' },
  { key: 'utility', label: 'Sup', icon: '/fe/lol-parties/icon-position-utility.png' },
];
export const CHAT_AVAILABILITY_OPTIONS = [
  { value: '', label: 'Default' },
  { value: 'chat', label: 'Online' },
  { value: 'away', label: 'Away' },
  { value: 'dnd', label: 'Busy' },
  { value: 'mobile', label: 'Mobile' },
  { value: 'offline', label: 'Offline' },
];
export const SWIFTPLAY_BUTTON_ID = 'ame-swiftplay-apply-btn';
export const STYLE_ID = 'ame-styles';
export const ROOM_PARTY_INDICATOR_CLASS = 'ame-room-party-indicator';
