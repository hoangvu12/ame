import { STYLE_ID, BUTTON_ID, SWIFTPLAY_BUTTON_ID, CHROMA_BTN_CLASS, CHROMA_PANEL_ID, IN_GAME_CONTAINER_ID, ROOM_PARTY_INDICATOR_CLASS, CUSTOM_SKINS_MODAL_ID } from './constants';

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
    .ame-connection-banner {
      display: flex;
      align-items: center;
      gap: 8px;
      margin: 6px 0 8px;
      padding: 6px 10px;
      border-radius: 4px;
      background: rgba(30, 24, 16, 0.6);
      border: 1px solid rgba(200, 170, 110, 0.35);
      font-family: var(--font-body);
      font-size: 12px;
      color: #f0e6d2;
      letter-spacing: 0.01em;
    }
    .ame-connection-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
      background: #c89b3c;
      box-shadow: 0 0 4px rgba(200, 155, 60, 0.6);
      flex-shrink: 0;
    }
    .ame-connection-text {
      line-height: 1.2;
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
      top: 36px;
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
    .ame-ingame-unstuck {
      font-size: 11px;
      white-space: nowrap;
      padding: 0 8px !important;
      min-width: 0 !important;
      height: 24px !important;
      line-height: 24px !important;
      margin: 0 !important;
    }
    .ame-settings-panel {
      height: 100%;
    }
    .ame-settings-panel > lol-uikit-scrollable {
      height: 100%;
    }
    .ame-settings-panel-inner {
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
    .ame-section + .ame-section {
      margin-top: 4px;
    }
    .ame-section-header {
      display: flex;
      align-items: center;
      gap: 8px;
      cursor: pointer;
      padding: 10px 0 6px;
      font-family: var(--font-display);
      font-size: 12px;
      color: #f0e6d2;
      letter-spacing: 0.075em;
      text-transform: uppercase;
      user-select: none;
      border-bottom: 1px solid rgba(91, 90, 86, 0.25);
    }
    .ame-section-header:hover {
      color: #c8aa6e;
    }
    .ame-section-chevron {
      width: 0;
      height: 0;
      border-left: 4px solid transparent;
      border-right: 4px solid transparent;
      border-top: 5px solid currentColor;
      transition: transform 0.15s;
      flex-shrink: 0;
    }
    .ame-section-chevron.collapsed {
      transform: rotate(-90deg);
    }
    .ame-section-body.collapsed {
      display: none;
    }
    .ame-settings-toggle-row {
      margin-top: 12px;
      display: flex;
      align-items: center;
    }
    .ame-sub-toggle {
      padding-left: 20px;
      opacity: 0.75;
    }
    .ame-bench-mark {
      position: absolute;
      top: 4px;
      right: 4px;
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background: #c89b3c;
      box-shadow: 0 0 4px #c89b3c;
      z-index: 99999;
      pointer-events: none;
    }
    .quick-play-skin-select-component .skin-purchase-button-container {
      display: none !important;
      visibility: hidden !important;
      pointer-events: none !important;
    }
    .quick-play-skin-select-component .thumbnail-wrapper.unowned,
    .quick-play-skin-select-component .thumbnail-wrapper.unowned .skin-thumbnail-img {
      opacity: 1 !important;
      filter: none !important;
      -webkit-filter: none !important;
      pointer-events: auto !important;
    }
    .ame-role-tabs {
      display: flex;
      gap: 4px;
      margin-top: 12px;
    }
    .ame-role-tab {
      width: 36px;
      height: 36px;
      padding: 4px;
      cursor: pointer;
      border: 2px solid transparent;
      border-radius: 4px;
      background: rgba(30, 35, 40, 0.6);
      display: flex;
      align-items: center;
      justify-content: center;
      transition: border-color 0.15s, background 0.15s;
    }
    .ame-role-tab:hover {
      border-color: #5b5a56;
      background: rgba(50, 55, 60, 0.8);
    }
    .ame-role-tab.active {
      border-color: #c8aa6e;
      background: rgba(60, 50, 30, 0.6);
    }
    .ame-role-tab img {
      width: 24px;
      height: 24px;
      display: block;
    }
    .ame-role-content {
      margin-top: 12px;
    }
    .ame-champion-list {
      margin-top: 8px;
    }
    .ame-champion-entry {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 4px 6px;
      border-radius: 4px;
      background: rgba(30, 35, 40, 0.4);
      margin-bottom: 4px;
    }
    .ame-champion-entry:hover {
      background: rgba(40, 45, 50, 0.6);
    }
    .ame-champion-entry .ame-champion-number {
      font-size: 11px;
      color: #5b5a56;
      width: 16px;
      text-align: center;
      flex-shrink: 0;
    }
    .ame-champion-entry img {
      width: 24px;
      height: 24px;
      border-radius: 2px;
      flex-shrink: 0;
    }
    .ame-champion-entry .ame-champion-name {
      flex: 1;
      font-size: 12px;
      color: #a09b8c;
    }
    .ame-champion-entry .ame-champion-remove {
      width: 18px;
      height: 18px;
      cursor: pointer;
      border: none;
      background: none;
      color: #5b5a56;
      font-size: 14px;
      line-height: 18px;
      text-align: center;
      padding: 0;
      flex-shrink: 0;
      border-radius: 2px;
    }
    .ame-champion-entry .ame-champion-remove:hover {
      color: #e25151;
      background: rgba(226, 81, 81, 0.15);
    }
    .ame-champion-search {
      position: relative;
    }
    .ame-champion-search lol-uikit-flat-input {
      width: 100%;
    }
    .ame-add-champion {
      margin-top: 6px;
    }
    .ame-search-dropdown {
      position: absolute;
      left: 0;
      right: 0;
      z-index: 100;
      background: #1e2328;
      border: 1px solid #5b5a56;
      border-radius: 4px;
      margin-top: 2px;
    }
    .ame-search-results {
      max-height: 180px;
    }
    .ame-search-results > * {
      flex-shrink: 0;
    }
    .ame-search-item {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 5px 8px;
      cursor: pointer;
      font-size: 12px;
      color: #a09b8c;
    }
    .ame-search-item:hover {
      background: rgba(200, 170, 110, 0.15);
      color: #f0e6d2;
    }
    .ame-search-item img {
      width: 20px;
      height: 20px;
      border-radius: 2px;
      flex-shrink: 0;
    }
    .ame-list-label {
      font-family: var(--font-body);
      font-size: 12px;
      color: #a09b8c;
      margin-top: 14px;
      margin-bottom: 2px;
      display: block;
    }
    #${SWIFTPLAY_BUTTON_ID}[disabled] {
      opacity: 0.4 !important;
      pointer-events: none !important;
      cursor: default !important;
      filter: grayscale(0.5) !important;
    }
    .${ROOM_PARTY_INDICATOR_CLASS} {
      position: absolute;
      bottom: -2px;
      left: 50%;
      transform: translateX(-50%);
      z-index: 50;
      background: rgba(30, 35, 40, 0.9);
      border: 1px solid #c89b3c;
      border-radius: 3px;
      padding: 1px 6px;
      font-size: 10px;
      color: #c8aa6e;
      white-space: nowrap;
      pointer-events: none;
      max-width: 140px;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .ame-settings-description {
      margin-top: 8px;
      font-family: var(--font-body);
      font-size: 11px;
      color: #5b5a56;
      line-height: 1.4;
    }
    .ame-status-buttons {
      display: flex;
      flex-wrap: wrap;
      gap: 6px;
      margin-top: 8px;
    }
    .ame-status-btn {
      padding: 4px 12px;
      font-family: var(--font-body);
      font-size: 12px;
      color: #a09b8c;
      background: rgba(30, 35, 40, 0.6);
      border: 1px solid #5b5a56;
      border-radius: 4px;
      cursor: pointer;
      user-select: none;
      transition: border-color 0.15s, color 0.15s, background 0.15s;
    }
    .ame-status-btn:hover {
      border-color: #c8aa6e;
      color: #f0e6d2;
      background: rgba(50, 55, 60, 0.8);
    }
    .ame-status-btn.ame-status-active {
      border-color: #c8aa6e;
      color: #c8aa6e;
      background: rgba(60, 50, 30, 0.6);
    }

    /* Custom Skins Modal */
    #${CUSTOM_SKINS_MODAL_ID} {
      position: absolute;
      inset: 0;
      z-index: 9999;
      font-family: var(--font-body);
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-backdrop {
      position: absolute;
      inset: 0;
      background: rgba(0, 0, 0, 0.7);
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-dialog {
      position: absolute;
      inset: 0;
      display: flex;
      align-items: center;
      justify-content: center;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-frame {
      width: 720px;
      max-height: 80vh;
      display: flex;
      flex-direction: column;
      background: #010a13;
      border: 2px solid #463714;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 12px 16px;
      border-bottom: 1px solid #463714;
      flex-shrink: 0;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-title {
      font-family: var(--font-display);
      font-size: 14px;
      color: #c8aa6e;
      letter-spacing: 0.1em;
      text-transform: uppercase;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-close {
      width: 24px;
      height: 24px;
      cursor: pointer;
      border: none;
      background: none;
      color: #5b5a56;
      font-size: 18px;
      line-height: 24px;
      text-align: center;
      padding: 0;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-close:hover {
      color: #c8aa6e;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-toolbar {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 10px 16px;
      border-bottom: 1px solid rgba(70, 55, 20, 0.5);
      flex-shrink: 0;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-toolbar lol-uikit-flat-input {
      flex: 1;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-toolbar .ame-champion-search {
      width: 200px;
      flex-shrink: 0;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-body {
      flex: 1;
      overflow-y: auto;
      padding: 12px 16px;
      min-height: 200px;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-empty {
      text-align: center;
      padding: 40px 20px;
      color: #5b5a56;
      font-size: 13px;
      line-height: 1.6;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
      gap: 10px;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-card {
      background: rgba(30, 35, 40, 0.6);
      border: 1px solid #463714;
      border-radius: 4px;
      overflow: hidden;
      display: flex;
      flex-direction: column;
      transition: border-color 0.15s;
      cursor: pointer;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-card:hover {
      border-color: #c8aa6e;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-card.csm-disabled {
      opacity: 0.5;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-card-img {
      width: 100%;
      aspect-ratio: 2/3;
      object-fit: contain;
      display: block;
      background: #0a1428;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-card-placeholder {
      width: 100%;
      aspect-ratio: 2/3;
      display: flex;
      align-items: center;
      justify-content: center;
      background: #0a1428;
      color: #3c3c41;
      font-size: 28px;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-card-info {
      padding: 8px 10px;
      flex: 1;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-card-name {
      font-size: 12px;
      color: #f0e6d2;
      font-weight: 600;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-card-author {
      font-size: 11px;
      color: #5b5a56;
      margin-top: 2px;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-card-actions {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 6px 10px;
      border-top: 1px solid rgba(70, 55, 20, 0.3);
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-card-btns {
      display: flex;
      gap: 6px;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-icon-btn {
      width: 22px;
      height: 22px;
      cursor: pointer;
      border: none;
      background: none;
      color: #5b5a56;
      font-size: 13px;
      line-height: 22px;
      text-align: center;
      padding: 0;
      border-radius: 3px;
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-icon-btn:hover {
      color: #c8aa6e;
      background: rgba(200, 170, 110, 0.1);
    }
    #${CUSTOM_SKINS_MODAL_ID} .csm-icon-btn.csm-delete:hover {
      color: #e25151;
      background: rgba(226, 81, 81, 0.1);
    }

    /* Import/Edit sub-dialog */
    .csm-subdialog-overlay {
      position: absolute;
      inset: 0;
      z-index: 10000;
      display: flex;
      align-items: center;
      justify-content: center;
      background: rgba(0, 0, 0, 0.5);
    }
    .csm-subdialog {
      width: 480px;
      background: #010a13;
      border: 2px solid #463714;
    }
    .csm-subdialog .csm-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 12px 16px;
      border-bottom: 1px solid #463714;
    }
    .csm-subdialog-body {
      padding: 16px;
      display: flex;
      gap: 16px;
    }
    .csm-subdialog-img {
      width: 100px;
      aspect-ratio: 2/3;
      border: 1px solid #463714;
      border-radius: 4px;
      cursor: pointer;
      overflow: hidden;
      display: flex;
      align-items: center;
      justify-content: center;
      background: #0a1428;
      flex-shrink: 0;
      position: relative;
    }
    .csm-subdialog-img img {
      width: 100%;
      height: 100%;
      object-fit: contain;
    }
    .csm-subdialog-img .csm-img-hint {
      font-size: 11px;
      color: #5b5a56;
      text-align: center;
      padding: 8px;
    }
    .csm-subdialog-img:hover {
      border-color: #c8aa6e;
    }
    .csm-subdialog-fields {
      flex: 1;
      display: flex;
      flex-direction: column;
      gap: 10px;
    }
    .csm-field-label {
      font-size: 11px;
      color: #a09b8c;
      margin-bottom: 3px;
      display: block;
    }
    .csm-subdialog-fields lol-uikit-flat-input {
      width: 100%;
    }
    .csm-subdialog-fields .ame-champion-search {
      width: 100%;
    }
    .csm-subdialog-footer {
      display: flex;
      justify-content: flex-end;
      gap: 8px;
      padding: 12px 16px;
      border-top: 1px solid #463714;
    }
  `;
  document.head.appendChild(style);
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
