import {
  wsSend, onGamePath, onSetting, refreshSettings, getAutoSelectRolesCache, onAutoSelectRoles, onChatStatus,
  onRandomSkin, setRandomSkinMode, onMapPrep, onConnection, setSetting, setAutoSelectRole, setChatStatus,
} from './websocket';
import { downloadDiagnosticZIP } from './diagnostics';
import { el } from './dom';
import { createButton, createCheckbox, createChampionSearch, createInput } from './components';
import { loadChampionSummary } from './api';
import { AUTO_SELECT_ROLES, CHAT_AVAILABILITY_OPTIONS } from './constants';
import { openCustomSkinsModal } from './customSkins';
import { t } from './i18n';

const NAV_TITLE_CLASS = 'lol-settings-nav-title';
const AME_NAV_NAME = 'ame-settings';
const AME_PANEL_CLASS = 'ame-settings-panel';
const UNINSTALL_DIALOG_ID = 'ame-uninstall-dialog';

let settingsObserver = null;
let injected = false;
let retryTimer = null;

function buildNavGroup() {
  const frag = document.createDocumentFragment();
  frag.appendChild(
    el('div', { class: NAV_TITLE_CLASS, dataset: { ame: '1' } }, t('settings.nav_title'))
  );
  frag.appendChild(
    el('lol-uikit-navigation-bar', {
      direction: 'down', type: 'tabbed', selectedindex: '-1', dataset: { ame: '1' },
    },
      el('lol-uikit-navigation-item', { name: AME_NAV_NAME, class: 'lol-settings-nav' },
        el('div', null, t('settings.nav_item'))
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

// Settings that only ame can carry out are greyed out while it is not running.
// The client-side ones (auto accept, bench swap, auto pick/ban, availability,
// random skin) keep working and are saved in the client.
const NEEDS_AME_CLASS = 'ame-needs-ame';
const OFFLINE_CLASS = 'ame-offline';

function buildToggle(id, labelText, settingKey, { needsAme = false } = {}) {
  const { checkbox, input } = createCheckbox(id, labelText, (checked) => {
    setSetting(settingKey, checked);
  });
  onSetting(settingKey, (enabled) => { input.checked = enabled; });
  return el('div', { class: `ame-settings-toggle-row${needsAme ? ` ${NEEDS_AME_CLASS}` : ''}` }, checkbox);
}

function trackConnection(panel) {
  // Stop listening once the panel has been shown and then removed (the
  // League settings window is rebuilt each time it opens).
  let shown = false;
  let off = null;
  off = onConnection((connected) => {
    if (panel.isConnected) shown = true;
    else if (shown) { off?.(); return; }
    panel.classList.toggle(OFFLINE_CLASS, !connected);
  });
}

export function mapPrepText(status) {
  if (!status) return t('settings.map_prep.idle');
  switch (status.state) {
    case 'preparing':
      return t('settings.map_prep.preparing', { progress: status.total ? ` (${status.current}/${status.total})` : '' });
    case 'ready':
      return t('settings.map_prep.ready');
    case 'failed':
      return t('settings.map_prep.failed');
    case 'skipped':
      return t('settings.map_prep.skipped', { reason: status.reason });
    case 'off':
      return t('settings.map_prep.off');
  }
  return t('settings.map_prep.idle');
}

function buildMapPrepSection() {
  const line = el('span', { class: 'ame-list-label ame-map-prep-status', style: 'margin-top: 4px' }, mapPrepText(null));
  onMapPrep((status) => { line.textContent = mapPrepText(status); });
  return el('div', { class: NEEDS_AME_CLASS },
    buildToggle('ameMapPrep', t('settings.general.map_prep'), 'mapPrep'),
    el('div', { class: 'ame-sub-toggle' }, line),
  );
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

function updateRoleList(role, listType, list) {
  const roleConfig = getAutoSelectRolesCache()[role] || { picks: [], bans: [] };
  const next = { picks: roleConfig.picks || [], bans: roleConfig.bans || [], [listType]: list };
  setAutoSelectRole(role, next.picks, next.bans);
  renderRoleContent();
}

function buildChampionEntry(id, index, listType) {
  const removeBtn = el('button', {
    class: 'ame-champion-remove',
    onClick: () => {
      const roleConfig = getAutoSelectRolesCache()[activeRole] || { picks: [], bans: [] };
      const list = [...(roleConfig[listType] || [])];
      list.splice(list.indexOf(id), 1);
      updateRoleList(activeRole, listType, list);
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
  const search = createChampionSearch({
    champions: championSummary || [],
    placeholder: t('settings.auto_select.search_placeholder'),
    clearOnSelect: true,
    getExclude: () => {
      const config = getAutoSelectRolesCache();
      const roleConfig = config[activeRole] || { picks: [], bans: [] };
      return new Set([...(roleConfig.picks || []), ...(roleConfig.bans || [])]);
    },
    onSelect: (id) => {
      if (id <= 0) return;
      const rc = getAutoSelectRolesCache()[activeRole] || { picks: [], bans: [] };
      updateRoleList(activeRole, listType, [...(rc[listType] || []), id]);
    },
  });
  const wrapper = el('div', { class: 'ame-add-champion' });
  wrapper.appendChild(search.container);
  return wrapper;
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
  roleContentEl.appendChild(buildChampionList('picks', t('settings.auto_select.pick_priority')));
  roleContentEl.appendChild(buildChampionList('bans', t('settings.auto_select.ban_priority')));
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
    }, el('img', { src: role.icon, alt: t(role.labelKey) }));
    return tab;
  });

  return el('div', { class: 'ame-role-tabs' }, ...tabs);
}

function buildAutoSelectSection() {
  roleContentEl = el('div', { class: 'ame-role-content' });

  const section = el('div', null,
    buildToggle('ameAutoSelect', t('settings.auto_select.enable'), 'autoSelect'),
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

  function apply() {
    updateButtons();
    // chatStatus.js listens for this and applies it to the client.
    setChatStatus(selectedAvailability, '');
  }

  const buttonRow = el('div', { class: 'ame-status-buttons' },
    ...CHAT_AVAILABILITY_OPTIONS.map(opt => {
      const btn = el('div', {
        class: 'ame-status-btn',
        onClick: () => {
          selectedAvailability = opt.value;
          apply();
        },
      }, t(opt.labelKey));
      buttons.push({ btn, value: opt.value });
      return btn;
    })
  );

  onChatStatus(({ availability }) => {
    selectedAvailability = availability || '';
    updateButtons();
  });

  return el('div', null,
    el('span', { class: 'ame-list-label', style: 'font-family: var(--font-body)' }, t('settings.chat_status.title')),
    buttonRow,
  );
}

const RANDOM_SKIN_OPTIONS = [
  { value: '', labelKey: 'random_skin.off' },
  { value: 'all', labelKey: 'random_skin.all' },
  { value: 'top3', labelKey: 'random_skin.top3' },
];

function buildRandomSkinSection() {
  let selectedMode = '';
  const buttons = [];

  function updateButtons() {
    buttons.forEach(({ btn, value }) => {
      if (value === selectedMode) {
        btn.classList.add('ame-status-active');
      } else {
        btn.classList.remove('ame-status-active');
      }
    });
  }

  const buttonRow = el('div', { class: 'ame-status-buttons' },
    ...RANDOM_SKIN_OPTIONS.map(opt => {
      const btn = el('div', {
        class: 'ame-status-btn',
        onClick: () => {
          selectedMode = opt.value;
          updateButtons();
          setRandomSkinMode(opt.value);
        },
      }, t(opt.labelKey));
      buttons.push({ btn, value: opt.value });
      return btn;
    })
  );

  onRandomSkin((mode) => {
    selectedMode = mode || '';
    updateButtons();
  });

  return el('div', null,
    el('span', { class: 'ame-list-label', style: 'font-family: var(--font-body)' }, t('random_skin.title')),
    buttonRow,
  );
}

function downloadLogs() {
  downloadDiagnosticZIP().catch(() => {});
}

function showUninstallDialog() {
  if (document.getElementById(UNINSTALL_DIALOG_ID)) return;

  const modal = el('div', { id: UNINSTALL_DIALOG_ID, class: 'modal', style: { position: 'absolute', inset: '0px' } },
    el('lol-uikit-full-page-backdrop', {
      class: 'backdrop',
      style: { display: 'flex', alignItems: 'center', justifyContent: 'center', position: 'absolute', inset: '0px' },
    }),
    el('div', {
      class: 'dialog-confirm',
      style: { display: 'flex', alignItems: 'center', justifyContent: 'center', position: 'absolute', inset: '0px' },
    },
      el('lol-uikit-dialog-frame', { class: 'dialog-frame', orientation: 'bottom' },
        el('div', { class: 'dialog-content' },
          el('lol-uikit-content-block', { class: 'app-controls-exit-dialog', type: 'dialog-small' },
            el('h4', null, t('settings.uninstall.title')),
            el('hr', { class: 'heading-spacer' }),
            el('p', null, t('settings.uninstall.body')),
          ),
        ),
        el('lol-uikit-flat-button-group', { type: 'dialog-frame' },
          el('lol-uikit-flat-button', {
            tabindex: '1',
            class: 'button-accept',
            onClick: () => {
              wsSend({ type: 'uninstall' });
              modal.remove();
            },
          }, t('settings.uninstall.confirm')),
          el('lol-uikit-flat-button', {
            tabindex: '2',
            class: 'button-decline',
            onClick: () => modal.remove(),
          }, t('settings.uninstall.cancel')),
        ),
      ),
    ),
  );

  document.body.appendChild(modal);
}

function buildPanel() {
  const { container: flatInput, input } = createInput({
    placeholder: t('settings.general.game_path_placeholder'),
  });

  const panel = el('div', { class: AME_PANEL_CLASS },
    el('lol-uikit-scrollable', {
      'overflow-masks': 'enabled',
      'scrolled-bottom': 'false',
      'scrolled-top': 'true',
    },
      el('div', { class: 'ame-settings-panel-inner' },
        el('div', { class: 'ame-offline-note' }, t('settings.offline_note')),
        buildSection(t('settings.sections.general'),
          el('div', { class: `ame-settings-row ${NEEDS_AME_CLASS}` },
            flatInput,
            createButton(t('settings.general.save'), {
              class: 'ame-settings-save',
              onClick: () => {
                const path = input.value.trim();
                if (!path) return;
                wsSend({ type: 'setGamePath', path });
              },
            })
          ),
          buildToggle('ameAutoAccept', t('settings.general.auto_accept'), 'autoAccept'),
          buildToggle('ameBenchSwap', t('settings.general.bench_swap'), 'benchSwap'),
          el('div', { class: 'ame-sub-toggle' },
            buildToggle('ameBenchSwapSkipCooldown', t('settings.general.bench_swap_skip_cooldown'), 'benchSwapSkipCooldown'),
          ),
          buildToggle('ameRoomParty', t('settings.general.room_party'), 'roomParty', { needsAme: true }),
          buildToggle('ameStartWithWindows', t('settings.general.start_with_windows'), 'startWithWindows', { needsAme: true }),
          buildToggle('ameAutoUpdate', t('settings.general.auto_update'), 'autoUpdate', { needsAme: true }),
          buildMapPrepSection(),
          buildChatStatusSection(),
          buildRandomSkinSection(),
          el('div', { class: `ame-settings-toggle-row ${NEEDS_AME_CLASS}` },
            createButton(t('custom_skins.button'), {
              onClick: () => openCustomSkinsModal(),
            }),
          ),
        ),
        buildSection(t('settings.sections.auto_select'),
          buildAutoSelectSection(),
        ),
        el('div', { class: `ame-settings-toggle-row ${NEEDS_AME_CLASS}`, style: { marginTop: '16px', gap: '8px' } },
          createButton(t('settings.download_logs.button'), {
            class: 'ame-settings-download-logs',
            onClick: () => downloadDiagnosticZIP(),
          }),
          createButton(`${t('settings.download_logs.button')} (plugin)`, {
            class: 'ame-settings-download-plugin-logs',
            onClick: () => downloadLogs(),
          }),
          createButton(t('settings.uninstall.button'), {
            class: 'ame-settings-uninstall',
            onClick: () => showUninstallDialog(),
          }),
        ),
      ),
    ),
  );
  trackConnection(panel);
  return panel;
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
