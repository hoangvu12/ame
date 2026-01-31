import { wsSend, onGamePath, onSetting, refreshSettings } from './websocket';
import { el } from './dom';
import { createButton, createCheckbox, createInput } from './components';

const NAV_TITLE_CLASS = 'lol-settings-nav-title';
const AME_NAV_NAME = 'ame-settings';
const AME_PANEL_CLASS = 'ame-settings-panel';

let settingsObserver = null;
let injected = false;
let retryTimer = null;

function buildNavGroup() {
  const frag = document.createDocumentFragment();
  frag.appendChild(
    el('div', { class: NAV_TITLE_CLASS, dataset: { ame: '1' } }, 'Ame')
  );
  frag.appendChild(
    el('lol-uikit-navigation-bar', {
      direction: 'down', type: 'tabbed', selectedindex: '-1', dataset: { ame: '1' },
    },
      el('lol-uikit-navigation-item', { name: AME_NAV_NAME, class: 'lol-settings-nav' },
        el('div', null, 'SETTINGS')
      )
    )
  );
  return frag;
}

function buildToggle(id, labelText, settingKey) {
  const { checkbox, input } = createCheckbox(id, labelText, (checked) => {
    wsSend({ type: `set${settingKey.charAt(0).toUpperCase()}${settingKey.slice(1)}`, enabled: checked });
  });
  onSetting(settingKey, (enabled) => { input.checked = enabled; });
  return el('div', { class: 'ame-settings-toggle-row' }, checkbox);
}

function buildPanel() {
  const { container: flatInput, input } = createInput({
    placeholder: 'C:\\Riot Games\\League of Legends\\Game',
  });

  return el('div', { class: AME_PANEL_CLASS },
    el('div', { class: 'lol-settings-ingame-section-title' }, 'Game Path'),
    el('div', { class: 'ame-settings-row' },
      flatInput,
      createButton('Save', {
        class: 'ame-settings-save',
        onClick: () => {
          const path = input.value.trim();
          if (!path) return;
          wsSend({ type: 'setGamePath', path });
        },
      })
    ),
    el('div', { class: 'lol-settings-ingame-section-title ame-settings-section-gap' }, 'Auto Accept Match'),
    buildToggle('ameAutoAccept', 'Automatically accept match when found', 'autoAccept'),
    el('div', { class: 'lol-settings-ingame-section-title ame-settings-section-gap' }, 'ARAM Bench Swap'),
    el('label', { class: 'ame-settings-description' },
      'Click a champion on the bench while it\'s on cooldown to mark it. When the cooldown ends, it will automatically be swapped to you.'
    ),
    buildToggle('ameBenchSwap', 'Enable auto bench swap in ARAM', 'benchSwap')
  );
}

function deselectAllNav(container) {
  container.querySelectorAll('lol-uikit-navigation-item').forEach(item => {
    item.removeAttribute('active');
  });
  container.querySelectorAll('lol-uikit-navigation-bar').forEach(bar => {
    bar.setAttribute('selectedindex', '-1');
  });
}

function showAmePanel(settingsContainer) {
  const optionsArea = settingsContainer.querySelector('.lol-settings-options');
  if (!optionsArea) return;

  for (const child of optionsArea.children) {
    if (!child.classList.contains(AME_PANEL_CLASS)) {
      child.dataset.ameHidden = child.style.display;
      child.style.display = 'none';
    }
  }

  let panel = optionsArea.querySelector(`.${AME_PANEL_CLASS}`);
  if (!panel) {
    panel = buildPanel();
    optionsArea.appendChild(panel);
  }
  panel.style.display = '';

  const input = panel.querySelector('input[type="text"]');
  if (input) {
    onGamePath((path) => { input.value = path || ''; });
    wsSend({ type: 'getGamePath' });
  }

  refreshSettings();
}

function hideAmePanel(settingsContainer) {
  const optionsArea = settingsContainer.querySelector('.lol-settings-options');
  if (!optionsArea) return;

  const panel = optionsArea.querySelector(`.${AME_PANEL_CLASS}`);
  if (panel) panel.style.display = 'none';

  for (const child of optionsArea.children) {
    if (child.dataset.ameHidden !== undefined) {
      child.style.display = child.dataset.ameHidden;
      delete child.dataset.ameHidden;
    }
  }

  const ameItem = settingsContainer.querySelector(`lol-uikit-navigation-item[name="${AME_NAV_NAME}"]`);
  if (ameItem) ameItem.removeAttribute('active');
}

function inject(settingsContainer) {
  if (injected) return;

  const scrollerContent = settingsContainer.querySelector('.lol-settings-nav-scroller > div');
  if (!scrollerContent) return;

  if (scrollerContent.querySelector(`[data-ame="1"]`)) {
    injected = true;
    return;
  }

  scrollerContent.prepend(buildNavGroup());
  injected = true;

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

  document.addEventListener('click', (e) => {
    if (e.target.closest('.app-controls-settings')) {
      if (!retryTimer) {
        retryTimer = setInterval(tryInject, 200);
      }
    }
  });

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
  tryInject();
}
