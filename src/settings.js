import { wsSend, onGamePath, onSetting, refreshSettings, getAutoSelectRolesCache, onAutoSelectRoles, onChatStatus } from './websocket';
import { el } from './dom';
import { createButton, createCheckbox, createInput } from './components';
import { loadChampionSummary } from './api';
import { AUTO_SELECT_ROLES, CHAT_AVAILABILITY_OPTIONS } from './constants';
import { applyChatStatus } from './chatStatus';

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

function buildSection(title, ...children) {
  const chevron = el('span', { class: 'ame-section-chevron' });
  const body = el('div', { class: 'ame-section-body' }, ...children);
  const header = el('div', {
    class: 'ame-section-header',
    onClick: () => {
      body.classList.toggle('collapsed');
      chevron.classList.toggle('collapsed');
    },
  }, chevron, el('span', null, title));
  return el('div', { class: 'ame-section' }, header, body);
}

function buildToggle(id, labelText, settingKey) {
  const { checkbox, input } = createCheckbox(id, labelText, (checked) => {
    wsSend({ type: `set${settingKey.charAt(0).toUpperCase()}${settingKey.slice(1)}`, enabled: checked });
  });
  onSetting(settingKey, (enabled) => { input.checked = enabled; });
  return el('div', { class: 'ame-settings-toggle-row' }, checkbox);
}

let championSummary = null;
let activeRole = 'top';

function getChampionName(id) {
  if (!championSummary) return `Champion ${id}`;
  const champ = championSummary.find(c => c.id === id);
  return champ ? champ.name : `Champion ${id}`;
}

function getChampionIcon(id) {
  return `/lol-game-data/assets/v1/champion-icons/${id}.png`;
}

function sendRoleUpdate(role) {
  const config = getAutoSelectRolesCache();
  const roleConfig = config[role] || { picks: [], bans: [] };
  wsSend({ type: 'setAutoSelectRole', role, picks: roleConfig.picks, bans: roleConfig.bans });
}

function buildChampionEntry(id, index, listType) {
  const removeBtn = el('button', {
    class: 'ame-champion-remove',
    onClick: () => {
      const config = getAutoSelectRolesCache();
      const roleConfig = config[activeRole] || { picks: [], bans: [] };
      const list = [...(roleConfig[listType] || [])];
      list.splice(list.indexOf(id), 1);
      config[activeRole] = { ...roleConfig, [listType]: list };
      sendRoleUpdate(activeRole);
      renderRoleContent();
    },
  }, '\u00d7');

  return el('div', { class: 'ame-champion-entry' },
    el('span', { class: 'ame-champion-number' }, `${index + 1}.`),
    el('img', { src: getChampionIcon(id) }),
    el('span', { class: 'ame-champion-name' }, getChampionName(id)),
    removeBtn,
  );
}

function buildChampionPicker(listType) {
  const { container: flatInput, input } = createInput({ placeholder: 'Search champion...' });
  const resultsList = el('lol-uikit-scrollable', {
    class: 'ame-search-results',
    'overflow-masks': 'enabled',
  });
  const resultsContainer = el('div', { class: 'ame-search-dropdown' }, resultsList);
  resultsContainer.style.display = 'none';

  function getAvailable() {
    const config = getAutoSelectRolesCache();
    const roleConfig = config[activeRole] || { picks: [], bans: [] };
    const existing = new Set([...(roleConfig.picks || []), ...(roleConfig.bans || [])]);
    return (championSummary || [])
      .filter(c => c.id > 0 && !existing.has(c.id))
      .sort((a, b) => a.name.localeCompare(b.name));
  }

  function showResults(filter) {
    const available = getAvailable();
    const query = (filter || '').toLowerCase();
    const filtered = query ? available.filter(c => c.name.toLowerCase().includes(query)) : available;

    resultsList.innerHTML = '';
    if (filtered.length === 0) {
      resultsContainer.style.display = 'none';
      return;
    }

    for (const champ of filtered) {
      const item = el('div', {
        class: 'ame-search-item',
        onClick: () => {
          const cfg = getAutoSelectRolesCache();
          const rc = cfg[activeRole] || { picks: [], bans: [] };
          const list = [...(rc[listType] || []), champ.id];
          cfg[activeRole] = { ...rc, [listType]: list };
          sendRoleUpdate(activeRole);
          input.value = '';
          resultsContainer.style.display = 'none';
          renderRoleContent();
        },
      },
        el('img', { src: getChampionIcon(champ.id) }),
        el('span', null, champ.name),
      );
      resultsList.appendChild(item);
    }
    resultsContainer.style.display = '';
  }

  input.addEventListener('input', () => showResults(input.value));
  input.addEventListener('focus', () => showResults(input.value));
  input.addEventListener('blur', () => {
    // Delay to allow click on results
    setTimeout(() => { resultsContainer.style.display = 'none'; }, 200);
  });

  return el('div', { class: 'ame-add-champion' }, flatInput, resultsContainer);
}

function buildChampionList(listType, label) {
  const config = getAutoSelectRolesCache();
  const roleConfig = config[activeRole] || { picks: [], bans: [] };
  const list = roleConfig[listType] || [];

  const entries = list.map((id, i) => buildChampionEntry(id, i, listType));

  return el('div', null,
    el('span', { class: 'ame-list-label' }, label),
    el('div', { class: 'ame-champion-list' }, ...entries),
    buildChampionPicker(listType),
  );
}

let roleContentEl = null;

function renderRoleContent() {
  if (!roleContentEl) return;
  roleContentEl.innerHTML = '';
  roleContentEl.appendChild(buildChampionList('picks', 'Pick priority:'));
  roleContentEl.appendChild(buildChampionList('bans', 'Ban priority:'));
}

function buildRoleTabs() {
  const tabs = AUTO_SELECT_ROLES.map(role => {
    const tab = el('div', {
      class: `ame-role-tab${role.key === activeRole ? ' active' : ''}`,
      onClick: () => {
        activeRole = role.key;
        tabs.forEach(t => t.classList.remove('active'));
        tab.classList.add('active');
        renderRoleContent();
      },
    }, el('img', { src: role.icon, alt: role.label }));
    return tab;
  });

  return el('div', { class: 'ame-role-tabs' }, ...tabs);
}

function buildAutoSelectSection() {
  roleContentEl = el('div', { class: 'ame-role-content' });

  const section = el('div', null,
    buildToggle('ameAutoSelect', 'Enable auto champion select', 'autoSelect'),
    buildRoleTabs(),
    roleContentEl,
  );

  // Load champion data then render the role content
  loadChampionSummary().then(data => {
    championSummary = data;
    renderRoleContent();
  });

  // Re-render when roles config arrives/updates from server
  onAutoSelectRoles(() => {
    if (championSummary) renderRoleContent();
  });

  return section;
}

function buildChatStatusSection() {
  let selectedAvailability = '';
  const buttons = [];

  function updateButtons() {
    buttons.forEach(({ btn, value }) => {
      if (value === selectedAvailability) {
        btn.classList.add('ame-status-active');
      } else {
        btn.classList.remove('ame-status-active');
      }
    });
  }

  async function apply() {
    const label = CHAT_AVAILABILITY_OPTIONS.find(o => o.value === selectedAvailability)?.label || 'Default';
    updateButtons();
    applyChatStatus(selectedAvailability);
    wsSend({ type: 'setChatStatus', availability: selectedAvailability, statusMessage: '' });
  }

  const buttonRow = el('div', { class: 'ame-status-buttons' },
    ...CHAT_AVAILABILITY_OPTIONS.map(opt => {
      const btn = el('div', {
        class: 'ame-status-btn',
        onClick: () => {
          selectedAvailability = opt.value;
          apply();
        },
      }, opt.label);
      buttons.push({ btn, value: opt.value });
      return btn;
    })
  );

  onChatStatus(({ availability }) => {
    selectedAvailability = availability || '';
    updateButtons();
  });

  return el('div', null,
    el('span', { class: 'ame-list-label', style: 'font-family: var(--font-body)' }, 'Availability'),
    buttonRow,
  );
}

function buildPanel() {
  const { container: flatInput, input } = createInput({
    placeholder: 'C:\\Riot Games\\League of Legends\\Game',
  });

  return el('div', { class: AME_PANEL_CLASS },
    el('lol-uikit-scrollable', {
      'overflow-masks': 'enabled',
      'scrolled-bottom': 'false',
      'scrolled-top': 'true',
    },
      el('div', { class: 'ame-settings-panel-inner' },
        buildSection('General',
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
          buildToggle('ameAutoAccept', 'Automatically accept match when found', 'autoAccept'),
          buildToggle('ameBenchSwap', 'Automatically swap ARAM bench for preferred champion', 'benchSwap'),
          el('div', { class: 'ame-sub-toggle' },
            buildToggle('ameBenchSwapSkipCooldown', 'Skip bench swap cooldown (experimental)', 'benchSwapSkipCooldown'),
          ),
          buildToggle('ameRoomParty', 'Share skins with teammates using Ame', 'roomParty'),
          buildToggle('ameStartWithWindows', 'Start ame with Windows', 'startWithWindows'),
          buildToggle('ameAutoUpdate', 'Automatically install updates', 'autoUpdate'),
          buildChatStatusSection(),
        ),
        buildSection('Auto Champion Select',
          buildAutoSelectSection(),
        ),
      ),
    ),
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
