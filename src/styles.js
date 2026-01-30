import { STYLE_ID, BUTTON_ID, CHROMA_BTN_CLASS, CHROMA_PANEL_ID, IN_GAME_CONTAINER_ID } from './constants';

export function injectStyles() {
  if (document.getElementById(STYLE_ID)) return;
  const style = document.createElement('style');
  style.id = STYLE_ID;
  style.textContent = `
    .skin-selection-item.disabled {
      filter: none !important;
      -webkit-filter: none !important;
      opacity: 1 !important;
      pointer-events: auto !important;
    }
    .champ-select-bg.locked .lol-uikit-background-switcher-image {
      filter: none !important;
      -webkit-filter: none !important;
      opacity: 1 !important;
    }
    .champ-select-bg.locked::after {
      display: none !important;
    }
    .skin-selection-item .lock-icon,
    .skin-selection-item .locked-icon,
    .skin-selection-item .skin-selection-item-unowned-overlay,
    .unlock-skin-hit-area,
    .unlock-skin-hit-area *,
    .locked-state {
      display: none !important;
      visibility: hidden !important;
      pointer-events: none !important;
      width: 0 !important;
      height: 0 !important;
      overflow: hidden !important;
    }
    .skin-selection-item:hover .skin-selection-thumbnail,
    .skin-selection-item.selected .skin-selection-thumbnail {
      border: 2px solid #c8aa6e;
      box-shadow: 0 0 6px #c8aa6e80;
    }
    #${BUTTON_ID}[disabled] {
      opacity: 0.4 !important;
      pointer-events: none !important;
      cursor: default !important;
      filter: grayscale(0.5) !important;
    }
    .toggle-ability-previews-button-container {
      justify-content: center !important;
      align-items: center !important;
      gap: 20px !important;
    }
    .toggle-ability-previews-button-container .framing-line {
      display: none !important;
    }
    .skin-selection-carousel .skin-selection-item {
      position: relative;
      overflow: visible !important;
    }
    .skin-selection-button-container,
    .skin-selection-carousel-container,
    .skin-selection-carousel {
      overflow: visible !important;
    }
    .skin-selection-carousel {
      position: relative;
    }
    .skin-selection-thumbnail {
      position: relative;
      overflow: visible !important;
      pointer-events: auto !important;
    }
    .skin-selection-carousel,
    .skin-selection-carousel-container {
      overflow: visible !important;
    }
    .${CHROMA_BTN_CLASS} {
      position: absolute;
      right: 6px;
      bottom: 6px;
      z-index: 5;
      pointer-events: auto;
      cursor: pointer;
      display: block !important;
    }
    .${CHROMA_BTN_CLASS}.hidden {
      display: none !important;
      pointer-events: none !important;
    }
    #${CHROMA_PANEL_ID} {
      position: absolute;
      z-index: 10000;
      pointer-events: auto;
    }
    #${CHROMA_PANEL_ID} lc-flyout-content {
      display: block !important;
      opacity: 1 !important;
      visibility: visible !important;
      pointer-events: auto !important;
    }
    #${IN_GAME_CONTAINER_ID} {
      position: absolute;
      top: 8px;
      right: 8px;
      z-index: 100;
      display: flex;
      align-items: center;
      gap: 6px;
    }
    .ame-ingame-dropdown {
      width: 180px;
      height: 24px !important;
      margin: 0 !important;
    }
    .ame-ingame-action {
      font-size: 11px;
      white-space: nowrap;
      padding: 0 8px !important;
      min-width: 0 !important;
      height: 24px !important;
      line-height: 24px !important;
      margin: 0 !important;
    }
    .ame-settings-panel {
      padding: 10px 12px;
    }
    .ame-settings-row {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-top: 12px;
    }
    .ame-settings-row lol-uikit-flat-input {
      flex: 1;
    }
    .ame-settings-section-gap {
      margin-top: 24px;
    }
    .ame-settings-toggle-row {
      margin-top: 12px;
      display: flex;
      align-items: center;
    }
  `;
  document.head.appendChild(style);
}

export function removeStyles() {
  const style = document.getElementById(STYLE_ID);
  if (style) style.remove();
}

export function unlockSkinCarousel() {
  const items = document.querySelectorAll('.skin-selection-item.disabled');
  for (const el of items) {
    el.classList.remove('disabled');
  }
  document.querySelectorAll('.unlock-skin-hit-area, .locked-state, .skin-selection-item-unowned-overlay').forEach(el => {
    el.style.display = 'none';
    el.style.visibility = 'hidden';
    el.style.pointerEvents = 'none';
  });
}
