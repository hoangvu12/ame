import { el } from './dom';
import { createButton, createCheckbox, createChampionSearch, createInput } from './components';
import { t } from './i18n';
import { loadChampionSummary } from './api';
import { CUSTOM_SKINS_MODAL_ID, CUSTOM_SKINS_IMAGE_BASE } from './constants';
import {
  onCustomMods, getCustomModsCache, refreshCustomMods,
  pickCustomModFile, importCustomMod, deleteCustomMod,
  updateCustomMod, toggleCustomMod, pickCustomModImage,
  deleteCustomModImage,
} from './websocket';

let modalEl = null;
let championList = null; // cached champion summary
let filterChampion = 0;
let filterText = '';
let unsubMods = null;

// --- Public API ---

export function openCustomSkinsModal() {
  if (modalEl) return;
  buildModal();
}

export function closeCustomSkinsModal() {
  if (modalEl) {
    modalEl.remove();
    modalEl = null;
  }
  if (unsubMods) {
    unsubMods();
    unsubMods = null;
  }
}

// --- Modal ---

async function buildModal() {
  // Ensure champion data is loaded
  championList = await loadChampionSummary();

  const searchInput = createInput({ placeholder: t('custom_skins.filter_placeholder') });
  searchInput.input.addEventListener('input', () => {
    filterText = searchInput.input.value.toLowerCase();
    renderGrid();
  });

  const champSearch = createChampionSearch({
    champions: championList || [],
    allOption: t('custom_skins.all_champions'),
    selected: 0,
    placeholder: t('custom_skins.all_champions'),
    onSelect: (id) => {
      filterChampion = id;
      renderGrid();
    },
  });

  const addBtn = createButton(t('custom_skins.add_mod'), {
    onClick: () => handleAddMod(),
  });

  const gridContainer = el('div', { class: 'csm-body' });

  modalEl = el('div', { id: CUSTOM_SKINS_MODAL_ID },
    el('lol-uikit-full-page-backdrop', { class: 'csm-backdrop', onClick: closeCustomSkinsModal }),
    el('div', { class: 'csm-dialog' },
      el('div', { class: 'csm-frame' },
        el('div', { class: 'csm-header' },
          el('span', { class: 'csm-title' }, t('custom_skins.title')),
          el('button', { class: 'csm-close', onClick: closeCustomSkinsModal }, '\u00D7'),
        ),
        el('div', { class: 'csm-toolbar' },
          searchInput.container,
          champSearch.container,
          addBtn,
        ),
        gridContainer,
      ),
    ),
  );

  // Stop click propagation on the frame so backdrop click doesn't close
  modalEl.querySelector('.csm-frame').addEventListener('click', e => e.stopPropagation());

  document.body.appendChild(modalEl);

  // Subscribe to mod list updates
  unsubMods = onCustomMods(() => renderGrid());
  refreshCustomMods();
}

function renderGrid() {
  if (!modalEl) return;
  const body = modalEl.querySelector('.csm-body');
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

  body.innerHTML = '';

  if (filtered.length === 0) {
    body.appendChild(el('div', { class: 'csm-empty' }, t('custom_skins.no_mods')));
    return;
  }

  const grid = el('div', { class: 'csm-grid' });

  for (const mod of filtered) {
    grid.appendChild(buildCard(mod));
  }

  body.appendChild(grid);
}

function buildCard(mod) {
  const imgEl = mod.hasImage
    ? el('img', { class: 'csm-card-img', src: CUSTOM_SKINS_IMAGE_BASE + mod.id + '?t=' + Date.now() })
    : el('div', { class: 'csm-card-placeholder' }, '\u{1F3AE}');

  const champName = getChampionNameById(mod.championId) || t('custom_skins.all_champions');

  const { checkbox, input } = createCheckbox(
    'csm-toggle-' + mod.id,
    '',
    (checked) => toggleCustomMod(mod.id, checked),
  );
  input.checked = mod.enabled;

  const card = el('div', {
    class: 'csm-card' + (mod.enabled ? '' : ' csm-disabled'),
    onClick: (e) => {
      if (e.target.closest('.csm-card-btns') || e.target.closest('lol-uikit-flat-checkbox')) return;
      toggleCustomMod(mod.id, !mod.enabled);
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
