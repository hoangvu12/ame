import { wsSend, onGamePath, onAutoAccept } from './websocket';
import { setAutoAcceptEnabled } from './autoAcceptMatch';

const NAV_TITLE_CLASS = 'lol-settings-nav-title';
const AME_NAV_NAME = 'ame-settings';
const AME_PANEL_CLASS = 'ame-settings-panel';

let settingsObserver = null;
let injected = false;
let retryTimer = null;

function buildNavGroup() {
  const frag = document.createDocumentFragment();

  const title = document.createElement('div');
  title.className = NAV_TITLE_CLASS;
  title.textContent = 'Ame';
  title.dataset.ame = '1';

  const bar = document.createElement('lol-uikit-navigation-bar');
  bar.setAttribute('direction', 'down');
  bar.setAttribute('type', 'tabbed');
  bar.setAttribute('selectedindex', '-1');
  bar.dataset.ame = '1';

  const item = document.createElement('lol-uikit-navigation-item');
  item.setAttribute('name', AME_NAV_NAME);
  item.className = 'lol-settings-nav';

  const label = document.createElement('div');
  label.textContent = 'SETTINGS';
  item.appendChild(label);
  bar.appendChild(item);

  frag.appendChild(title);
  frag.appendChild(bar);
  return frag;
}

function buildPanel() {
  const panel = document.createElement('div');
  panel.className = AME_PANEL_CLASS;

  const section = document.createElement('div');
  section.className = 'lol-settings-ingame-section-title';
  section.textContent = 'Game Path';

  const row = document.createElement('div');
  row.className = 'ame-settings-row';

  const flatInput = document.createElement('lol-uikit-flat-input');
  const input = document.createElement('input');
  input.type = 'text';
  input.placeholder = 'C:\\Riot Games\\League of Legends\\Game';
  flatInput.appendChild(input);

  const btn = document.createElement('lol-uikit-flat-button');
  btn.className = 'ame-settings-save';
  btn.textContent = 'Save';
  btn.addEventListener('click', () => {
    const path = input.value.trim();
    if (!path) return;
    wsSend({ type: 'setGamePath', path });
  });

  row.appendChild(flatInput);
  row.appendChild(btn);

  panel.appendChild(section);
  panel.appendChild(row);

  // Auto Accept Match toggle
  const autoAcceptSection = document.createElement('div');
  autoAcceptSection.className = 'lol-settings-ingame-section-title ame-settings-section-gap';
  autoAcceptSection.textContent = 'Auto Accept Match';

  const toggleRow = document.createElement('div');
  toggleRow.className = 'ame-settings-toggle-row';

  const checkbox = document.createElement('lol-uikit-flat-checkbox');
  checkbox.setAttribute('for', 'ameAutoAccept');

  const cbInput = document.createElement('input');
  cbInput.setAttribute('slot', 'input');
  cbInput.setAttribute('name', 'ameAutoAccept');
  cbInput.type = 'checkbox';
  cbInput.id = 'ameAutoAccept';
  checkbox.appendChild(cbInput);

  const cbLabel = document.createElement('label');
  cbLabel.setAttribute('slot', 'label');
  cbLabel.textContent = 'Automatically accept match when found';
  checkbox.appendChild(cbLabel);

  cbInput.addEventListener('change', () => {
    const enabled = cbInput.checked;
    setAutoAcceptEnabled(enabled);
    wsSend({ type: 'setAutoAccept', enabled });
  });

  toggleRow.appendChild(checkbox);
  panel.appendChild(autoAcceptSection);
  panel.appendChild(toggleRow);

  return panel;
}

function deselectAllNav(container) {
  container.querySelectorAll('lol-uikit-navigation-item').forEach(el => {
    el.removeAttribute('active');
  });
  container.querySelectorAll('lol-uikit-navigation-bar').forEach(bar => {
    bar.setAttribute('selectedindex', '-1');
  });
}

function showAmePanel(settingsContainer) {
  const optionsArea = settingsContainer.querySelector('.lol-settings-options');
  if (!optionsArea) return;

  // Hide existing children
  for (const child of optionsArea.children) {
    if (!child.classList.contains(AME_PANEL_CLASS)) {
      child.dataset.ameHidden = child.style.display;
      child.style.display = 'none';
    }
  }

  // Add our panel if not already there
  let panel = optionsArea.querySelector(`.${AME_PANEL_CLASS}`);
  if (!panel) {
    panel = buildPanel();
    optionsArea.appendChild(panel);
  }
  panel.style.display = '';

  // Populate input from server
  const input = panel.querySelector('input[type="text"]');
  if (input) {
    onGamePath((path) => {
      input.value = path || '';
    });
    wsSend({ type: 'getGamePath' });
  }

  // Populate auto-accept checkbox from server
  const cbInput = panel.querySelector('lol-uikit-flat-checkbox input[type="checkbox"]');
  if (cbInput) {
    onAutoAccept((enabled) => {
      cbInput.checked = enabled;
    });
    wsSend({ type: 'getAutoAccept' });
  }
}

function hideAmePanel(settingsContainer) {
  const optionsArea = settingsContainer.querySelector('.lol-settings-options');
  if (!optionsArea) return;

  const panel = optionsArea.querySelector(`.${AME_PANEL_CLASS}`);
  if (panel) panel.style.display = 'none';

  // Restore hidden children
  for (const child of optionsArea.children) {
    if (child.dataset.ameHidden !== undefined) {
      child.style.display = child.dataset.ameHidden;
      delete child.dataset.ameHidden;
    }
  }

  // Deselect our nav item
  const ameItem = settingsContainer.querySelector(`lol-uikit-navigation-item[name="${AME_NAV_NAME}"]`);
  if (ameItem) ameItem.removeAttribute('active');
}

function inject(settingsContainer) {
  if (injected) return;

  const scrollerContent = settingsContainer.querySelector('.lol-settings-nav-scroller > div');
  if (!scrollerContent) return;

  // Don't re-inject if already present
  if (scrollerContent.querySelector(`[data-ame="1"]`)) {
    injected = true;
    return;
  }

  scrollerContent.prepend(buildNavGroup());
  injected = true;

  // Handle our nav item click
  const ameItem = scrollerContent.querySelector(`lol-uikit-navigation-item[name="${AME_NAV_NAME}"]`);
  if (ameItem) {
    ameItem.addEventListener('click', (e) => {
      e.stopPropagation();
      const navArea = settingsContainer.querySelector('.lol-settings-navs') || scrollerContent;
      deselectAllNav(navArea);
      ameItem.setAttribute('active', 'true');
      showAmePanel(settingsContainer);
    });
  }

  // When any other nav item is clicked, hide our panel
  scrollerContent.addEventListener('click', (e) => {
    const navItem = e.target.closest('lol-uikit-navigation-item');
    if (!navItem) return;
    if (navItem.getAttribute('name') === AME_NAV_NAME) return;
    hideAmePanel(settingsContainer);
  });
}

function stopRetry() {
  if (retryTimer) { clearInterval(retryTimer); retryTimer = null; }
}

function cleanup() {
  injected = false;
  stopRetry();
}

function tryInject() {
  const container = document.querySelector('.lol-settings-container');
  if (!container) {
    cleanup();
    return;
  }
  inject(container);
  if (injected) stopRetry();
}

export function initSettings() {
  if (settingsObserver) return;

  if (!document.body) {
    setTimeout(initSettings, 250);
    return;
  }

  // Listen for settings button click — starts retry immediately
  document.addEventListener('click', (e) => {
    if (e.target.closest('.app-controls-settings')) {
      if (!retryTimer) {
        retryTimer = setInterval(tryInject, 200);
      }
    }
  });

  // Fallback: MutationObserver for cases where dialog appears without button click
  settingsObserver = new MutationObserver(() => {
    const container = document.querySelector('.lol-settings-container');
    if (container) {
      if (!injected && !retryTimer) {
        retryTimer = setInterval(tryInject, 200);
      }
    } else {
      cleanup();
    }
  });

  settingsObserver.observe(document.body, { childList: true, subtree: true });

  // Check immediately in case settings is already open
  tryInject();
}
