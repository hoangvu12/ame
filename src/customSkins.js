import { el } from './dom';
import { createButton, createCheckbox, createChampionSearch, createInput, createCardImage } from './components';
import { t } from './i18n';
import { loadChampionSummary } from './api';
import { CUSTOM_SKINS_MODAL_ID, CUSTOM_SKINS_IMAGE_BASE } from './constants';
import {
  onCustomMods, getCustomModsCache, refreshCustomMods,
  pickCustomModFile, importCustomMod, deleteCustomMod,
  updateCustomMod, toggleCustomMod, pickCustomModImage,
  deleteCustomModImage,
} from './websocket';
import { createBrowseContent, destroyBrowseContent } from './browse';

let modalEl = null;
let championList = null; // cached champion summary
let filterChampion = 0;
let filterText = '';
let unsubMods = null;
let activeTab = 'mods'; // 'mods' | 'browse'
let modsContainer = null;
let browseContainer = null;

// --- Public API ---

export function openCustomSkinsModal() {
  if (modalEl) return;
  buildModal();
}

export function closeCustomSkinsModal() {
  if (modalEl) {
    destroyBrowseContent();
    modalEl.remove();
    modalEl = null;
  }
  modsContainer = null;
  browseContainer = null;
  if (unsubMods) {
    unsubMods();
    unsubMods = null;
  }
  activeTab = 'mods';
}

// --- Modal ---

async function buildModal() {
  // Ensure champion data is loaded
  championList = await loadChampionSummary();

  modalEl = el('div', { id: CUSTOM_SKINS_MODAL_ID },
    el('lol-uikit-full-page-backdrop', { class: 'csm-backdrop', onClick: closeCustomSkinsModal }),
    el('div', { class: 'csm-dialog' },
      el('div', { class: 'csm-frame' },
        el('div', { class: 'csm-header' },
          el('div', { class: 'csm-tabs' },
            el('button', {
              class: 'csm-tab active',
              dataset: { tab: 'mods' },
              onClick: () => switchTab('mods'),
            }, t('browse.tab_my_mods')),
            el('button', {
              class: 'csm-tab',
              dataset: { tab: 'browse' },
              onClick: () => switchTab('browse'),
            }, t('browse.tab_browse')),
          ),
          el('button', { class: 'csm-close', onClick: closeCustomSkinsModal }, '\u00D7'),
        ),
        el('div', { class: 'csm-tab-content' }),
      ),
    ),
  );

  // Stop click propagation on the frame so backdrop click doesn't close
  modalEl.querySelector('.csm-frame').addEventListener('click', e => e.stopPropagation());

  document.body.appendChild(modalEl);

  // Subscribe to mod list updates
  unsubMods = onCustomMods((mods, event) => {
    if (event === 'updated') {
      updateCardsInPlace(mods);
    } else {
      renderGrid();
    }
  });
  refreshCustomMods();

  // Render initial tab
  switchTab('mods');
}

function switchTab(tab) {
  activeTab = tab;
  if (!modalEl) return;

  const tabs = modalEl.querySelectorAll('.csm-tab');
  tabs.forEach(t => t.classList.toggle('active', t.dataset.tab === tab));

  const content = modalEl.querySelector('.csm-tab-content');
  if (!content) return;

  // Lazily create containers on first visit
  if (!modsContainer) {
    modsContainer = el('div', { class: 'csm-tab-panel' });
    content.appendChild(modsContainer);
    renderModsTab(modsContainer);
  }
  if (!browseContainer && tab === 'browse') {
    browseContainer = el('div', { class: 'csm-tab-panel' });
    content.appendChild(browseContainer);
    renderBrowseTab(browseContainer);
  }

  // Toggle visibility
  modsContainer.style.display = tab === 'mods' ? '' : 'none';
  if (browseContainer) browseContainer.style.display = tab === 'browse' ? '' : 'none';
}

function renderModsTab(content) {
  const searchInput = createInput({ placeholder: t('custom_skins.filter_placeholder') });
  searchInput.input.value = filterText;
  searchInput.input.addEventListener('input', () => {
    filterText = searchInput.input.value.toLowerCase();
    renderGrid();
  });

  const champSearch = createChampionSearch({
    champions: championList || [],
    allOption: t('custom_skins.all_champions'),
    selected: filterChampion,
    placeholder: t('custom_skins.all_champions'),
    onSelect: (id) => {
      filterChampion = id;
      renderGrid();
    },
  });

  const addBtn = createButton(t('custom_skins.add_mod'), {
    onClick: () => handleAddMod(),
  });

  const toolbar = el('div', { class: 'csm-toolbar' },
    searchInput.container,
    champSearch.container,
    addBtn,
  );

  const gridContainer = el('div', { class: 'csm-body' });

  content.appendChild(toolbar);
  content.appendChild(gridContainer);

  renderGrid();
}

function renderBrowseTab(content) {
  const browseContainer = el('div', { class: 'csm-body brw-container' });
  content.appendChild(browseContainer);
  createBrowseContent(browseContainer);
}

function renderGrid() {
  if (!modalEl || !modsContainer) return;
  const body = modsContainer.querySelector('.csm-body');
  if (!body) return;

  const mods = getCustomModsCache();
  const filtered = mods.filter(m => {
    if (filterChampion > 0 && m.championId !== 0 && m.championId !== filterChampion) return false;
    if (filterText) {
      const haystack = (m.name + ' ' + m.author).toLowerCase();
      if (!haystack.includes(filterText)) return false;
    }
    return true;
  });

  let grid = body.querySelector('.csm-grid');
  const empty = body.querySelector('.csm-empty');

  if (filtered.length === 0) {
    if (grid) grid.remove();
    if (!empty) body.appendChild(el('div', { class: 'csm-empty' }, t('custom_skins.no_mods')));
    return;
  }

  if (empty) empty.remove();
  if (!grid) {
    grid = el('div', { class: 'csm-grid' });
    body.appendChild(grid);
  }

  // Map existing cards by mod ID
  const existing = new Map();
  for (const card of grid.querySelectorAll('.csm-card[data-mod-id]')) {
    existing.set(card.dataset.modId, card);
  }

  // Remove cards no longer in filtered set
  const desiredIds = new Set(filtered.map(m => m.id));
  for (const [id, card] of existing) {
    if (!desiredIds.has(id)) {
      card.remove();
      existing.delete(id);
    }
  }

  // Append in order — appendChild moves existing nodes to correct position
  for (const mod of filtered) {
    grid.appendChild(existing.get(mod.id) || buildCard(mod));
  }
}

function updateCardsInPlace(mods) {
  if (!modalEl) return;
  const cards = modalEl.querySelectorAll('.csm-card[data-mod-id]');
  for (const card of cards) {
    const mod = mods.find(m => m.id === card.dataset.modId);
    if (!mod) continue;
    card.classList.toggle('csm-disabled', !mod.enabled);
    const cb = card.querySelector('input[type="checkbox"]');
    if (cb) cb.checked = mod.enabled;
  }
}

function buildCard(mod) {
  const imgSrc = mod.hasImage
    ? CUSTOM_SKINS_IMAGE_BASE + mod.id + '?t=' + Date.now()
    : '';
  const imgEl = createCardImage(imgSrc);

  const champName = getChampionNameById(mod.championId) || t('custom_skins.all_champions');

  const { checkbox, input } = createCheckbox(
    'csm-toggle-' + mod.id,
    '',
    (checked) => toggleCustomMod(mod.id, checked),
  );
  input.checked = mod.enabled;

  const card = el('div', {
    class: 'csm-card' + (mod.enabled ? '' : ' csm-disabled'),
    dataset: { modId: mod.id },
    onClick: (e) => {
      if (e.target.closest('.csm-card-btns') || e.target.closest('lol-uikit-flat-checkbox')) return;
      toggleCustomMod(mod.id, card.classList.contains('csm-disabled'));
    },
  },
    imgEl,
    el('div', { class: 'csm-card-info' },
      el('div', { class: 'csm-card-name', title: mod.name }, mod.name),
      el('div', { class: 'csm-card-author' }, mod.author ? `by ${mod.author}` : champName),
    ),
    el('div', { class: 'csm-card-actions' },
      checkbox,
      el('div', { class: 'csm-card-btns' },
        el('button', { class: 'csm-icon-btn', title: 'Edit', onClick: () => openEditDialog(mod) }, '\u270E'),
        el('button', { class: 'csm-icon-btn csm-delete', title: 'Delete', onClick: () => openDeleteDialog(mod) }, '\u2716'),
      ),
    ),
  );

  return card;
}

// --- Add/Import Flow ---

function handleAddMod() {
  pickCustomModFile((result) => {
    if (!result) return;
    openImportDialog(result);
  });
}

function openImportDialog(picked) {
  const state = {
    name: picked.suggestedName || '',
    author: picked.suggestedAuthor || '',
    championId: picked.suggestedChampionId || 0,
    imageBase64: '',
    filePath: picked.filePath,
  };

  showSubDialog({
    title: t('custom_skins.import_title'),
    state,
    confirmText: t('custom_skins.import_btn'),
    onConfirm: () => {
      importCustomMod({
        name: state.name,
        author: state.author,
        championId: state.championId,
        filePath: state.filePath,
        imageBase64: state.imageBase64,
      });
    },
  });
}

// --- Edit Flow ---

function openEditDialog(mod) {
  const state = {
    id: mod.id,
    name: mod.name,
    author: mod.author,
    championId: mod.championId,
    hasImage: mod.hasImage,
    imageBase64: '',
    imageChanged: false,
  };

  showSubDialog({
    title: t('custom_skins.edit_title'),
    state,
    isEdit: true,
    confirmText: t('custom_skins.save_btn'),
    onConfirm: () => {
      updateCustomMod({ id: state.id, name: state.name, author: state.author, championId: state.championId });
      if (state.imageChanged && !state.imageBase64 && !state.hasImage) {
        deleteCustomModImage(state.id);
      }
    },
  });
}

// --- Delete Flow ---

function openDeleteDialog(mod) {
  const overlay = el('div', { class: 'csm-subdialog-overlay' },
    el('div', { class: 'csm-subdialog' },
      el('div', { class: 'csm-header' },
        el('span', { class: 'csm-title' }, t('custom_skins.delete_title')),
      ),
      el('div', { style: { padding: '16px', color: '#a09b8c', fontSize: '13px' } },
        t('custom_skins.delete_body', { name: mod.name }),
      ),
      el('div', { class: 'csm-subdialog-footer' },
        createButton(t('custom_skins.delete_cancel'), {
          onClick: () => overlay.remove(),
        }),
        createButton(t('custom_skins.delete_confirm'), {
          onClick: () => {
            deleteCustomMod(mod.id);
            overlay.remove();
          },
        }),
      ),
    ),
  );

  overlay.querySelector('.csm-subdialog').addEventListener('click', e => e.stopPropagation());
  overlay.addEventListener('click', () => overlay.remove());

  (modalEl || document.body).appendChild(overlay);
}

// --- Sub-dialog (Import/Edit) ---

function showSubDialog({ title, state, isEdit, confirmText, onConfirm }) {
  const nameInput = createInput({ placeholder: t('custom_skins.name') });
  nameInput.input.value = state.name;
  nameInput.input.addEventListener('input', () => { state.name = nameInput.input.value; });

  const authorInput = createInput({ placeholder: t('custom_skins.author') });
  authorInput.input.value = state.author;
  authorInput.input.addEventListener('input', () => { state.author = authorInput.input.value; });

  const champSearch = createChampionSearch({
    champions: championList || [],
    allOption: t('custom_skins.all_champions'),
    selected: state.championId,
    placeholder: t('custom_skins.champion'),
    onSelect: (id) => {
      state.championId = id;
    },
  });

  // Image area
  const imgPreview = el('div', { class: 'csm-subdialog-img' });
  const fileInput = el('input', { type: 'file', accept: 'image/*', style: { display: 'none' } });

  function updateImagePreview() {
    imgPreview.innerHTML = '';
    if (state.imageBase64) {
      imgPreview.appendChild(el('img', { src: state.imageBase64 }));
    } else if (isEdit && state.hasImage) {
      imgPreview.appendChild(el('img', { src: CUSTOM_SKINS_IMAGE_BASE + state.id + '?t=' + Date.now() }));
    } else {
      imgPreview.appendChild(el('div', { class: 'csm-img-hint' }, t('custom_skins.image_placeholder')));
    }
  }
  updateImagePreview();

  imgPreview.addEventListener('click', () => {
    if (isEdit) {
      // For edit mode, use backend file picker
      pickCustomModImage(state.id);
      state.hasImage = true;
      state.imageChanged = true;
      // Image will be updated via customModUpdated message
    } else {
      fileInput.click();
    }
  });

  fileInput.addEventListener('change', () => {
    const file = fileInput.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => {
      state.imageBase64 = reader.result;
      updateImagePreview();
    };
    reader.readAsDataURL(file);
  });

  const overlay = el('div', { class: 'csm-subdialog-overlay' },
    el('div', { class: 'csm-subdialog' },
      el('div', { class: 'csm-header' },
        el('span', { class: 'csm-title' }, title),
        el('button', { class: 'csm-close', onClick: () => overlay.remove() }, '\u00D7'),
      ),
      el('div', { class: 'csm-subdialog-body' },
        imgPreview,
        fileInput,
        el('div', { class: 'csm-subdialog-fields' },
          el('div', null,
            el('span', { class: 'csm-field-label' }, t('custom_skins.name')),
            nameInput.container,
          ),
          el('div', null,
            el('span', { class: 'csm-field-label' }, t('custom_skins.author')),
            authorInput.container,
          ),
          el('div', null,
            el('span', { class: 'csm-field-label' }, t('custom_skins.champion')),
            champSearch.container,
          ),
        ),
      ),
      el('div', { class: 'csm-subdialog-footer' },
        createButton(t('custom_skins.cancel_btn'), {
          onClick: () => overlay.remove(),
        }),
        createButton(confirmText, {
          onClick: () => {
            onConfirm();
            overlay.remove();
          },
        }),
      ),
    ),
  );

  overlay.querySelector('.csm-subdialog').addEventListener('click', e => e.stopPropagation());
  overlay.addEventListener('click', () => overlay.remove());

  (modalEl || document.body).appendChild(overlay);
}

// --- Helpers ---

function getChampionNameById(id) {
  if (!id || id === 0 || !championList) return null;
  const entry = championList.find(c => c.id === id);
  return entry ? entry.name : null;
}
