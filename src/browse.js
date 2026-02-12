import { el } from './dom';
import { createButton, createInput, createCardImage } from './components';
import { t } from './i18n';
import { PROXY_IMAGE_BASE } from './constants';
import { startBrowseDownload } from './websocket';
import { loadExtensions, getLoadedSources, isLoaded, addExtension, addExtensionFromFile, removeExtension } from './extensionManager';

let browseRoot = null;
let activeSource = null;
let results = [];
let currentPage = 1;
let hasNextPage = false;
let currentQuery = '';
let currentFilters = {};
let selectedDetail = null;
let loading = false;
let browseMode = 'popular'; // 'popular' | 'latest' | 'search'
let filtersVisible = false;
let downloadProgress = {};
let gridEl = null;
let scrollableEl = null;
let loadMoreRow = null;

// --- Public API ---

export function createBrowseContent(container) {
  browseRoot = container;
  if (!isLoaded()) {
    browseRoot.innerHTML = '';
    browseRoot.appendChild(el('div', { class: 'csm-empty' }, t('ui.loading')));
    loadExtensions().then(() => renderBrowse());
  } else {
    renderBrowse();
  }
}

export function destroyBrowseContent() {
  browseRoot = null;
  selectedDetail = null;
  gridEl = null;
  scrollableEl = null;
  loadMoreRow = null;
}

// --- Render ---

function renderBrowse() {
  if (!browseRoot) return;
  browseRoot.innerHTML = '';
  gridEl = null;
  scrollableEl = null;
  loadMoreRow = null;

  const sources = getLoadedSources();

  if (sources.length === 0) {
    browseRoot.appendChild(buildNoSources());
    return;
  }

  if (!activeSource) {
    activeSource = sources[0];
    currentFilters = getDefaultFilters(activeSource);
    doFetch();
    return;
  }

  browseRoot.appendChild(buildBrowseToolbar());
  if (filtersVisible) browseRoot.appendChild(buildFilterPanel());
  const chips = buildFilterChips();
  if (chips) browseRoot.appendChild(chips);
  browseRoot.appendChild(buildResultsArea());
}

function buildNoSources() {
  return el('div', { class: 'brw-no-sources' },
    el('div', { class: 'csm-empty' }, t('browse.no_sources')),
    buildExtensionsManager(),
  );
}

// --- Toolbar ---

function buildBrowseToolbar() {
  const sources = getLoadedSources();

  const sourceSelect = el('select', { class: 'brw-source-select' });
  for (const src of sources) {
    const opt = el('option', { value: src.id }, src.name);
    if (activeSource && src.id === activeSource.id) opt.selected = true;
    sourceSelect.appendChild(opt);
  }
  sourceSelect.addEventListener('change', () => {
    const src = sources.find(s => s.id === sourceSelect.value);
    if (src && src.id !== activeSource?.id) {
      activeSource = src;
      results = [];
      currentPage = 1;
      currentQuery = '';
      currentFilters = getDefaultFilters(activeSource);
      selectedDetail = null;
      doFetch();
    }
  });

  const gearBtn = createButton('+', {
    class: 'brw-gear-btn',
    onClick: () => openExtensionsOverlay(),
  });

  const searchInput = createInput({ placeholder: t('browse.search_placeholder') });
  searchInput.input.value = currentQuery;
  searchInput.input.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') {
      currentQuery = searchInput.input.value.trim();
      results = [];
      currentPage = 1;
      browseMode = currentQuery ? 'search' : 'popular';
      doFetch();
    }
  });

  const filterBtn = el('button', {
    class: 'csm-icon-btn brw-filter-btn' + (filtersVisible ? ' active' : ''),
    title: 'Filters',
    onClick: () => {
      filtersVisible = !filtersVisible;
      renderBrowse();
    },
  }, '\u2630');

  return el('div', { class: 'csm-toolbar brw-toolbar' },
    sourceSelect,
    gearBtn,
    searchInput.container,
    filterBtn,
  );
}

// --- Filters ---

function getDefaultFilters(source) {
  if (!source || !source.getFilters) return {};
  const filters = {};
  try {
    const defs = source.getFilters();
    for (const f of defs) {
      filters[f.key] = f.default;
    }
  } catch { /* ignore */ }
  return filters;
}

function hasActiveFilters() {
  if (!activeSource) return false;
  const defaults = getDefaultFilters(activeSource);
  for (const key of Object.keys(currentFilters)) {
    if (key === 'sort') continue; // sort doesn't affect getPopular behavior
    if (currentFilters[key] !== defaults[key]) return true;
  }
  return false;
}

function buildFilterPanel() {
  if (!activeSource || !activeSource.getFilters) return el('div');
  let filterDefs;
  try { filterDefs = activeSource.getFilters(); } catch { return el('div'); }
  if (!filterDefs || filterDefs.length === 0) return el('div');

  const panel = el('div', { class: 'brw-filter-panel' });
  for (const f of filterDefs) {
    panel.appendChild(buildFilterControl(f));
  }
  const applyBtn = createButton(t('browse.search_placeholder').replace('...', ''), {
    onClick: () => {
      results = [];
      currentPage = 1;
      browseMode = currentQuery ? 'search' : 'popular';
      doFetch();
    },
  });
  panel.appendChild(applyBtn);
  return panel;
}

function buildFilterControl(filter) {
  const wrapper = el('div', { class: 'brw-filter-control' });
  const label = el('span', { class: 'csm-field-label' }, filter.name);
  wrapper.appendChild(label);

  if (filter.type === 'select' || filter.type === 'sort') {
    const select = el('select', { class: 'brw-filter-select' });
    for (const opt of filter.options) {
      const o = el('option', { value: opt.value }, opt.label);
      if (currentFilters[filter.key] === opt.value) o.selected = true;
      select.appendChild(o);
    }
    select.addEventListener('change', () => { currentFilters[filter.key] = select.value; });
    wrapper.appendChild(select);
  } else if (filter.type === 'text') {
    const inp = createInput({ placeholder: filter.name });
    inp.input.value = currentFilters[filter.key] || '';
    inp.input.addEventListener('input', () => { currentFilters[filter.key] = inp.input.value; });
    wrapper.appendChild(inp.container);
  } else if (filter.type === 'multiselect') {
    const selected = (currentFilters[filter.key] || '').split(',').filter(Boolean);
    const container = el('div', { class: 'brw-multiselect' });
    for (const opt of filter.options) {
      const checked = selected.includes(opt.value);
      const cb = el('input', { type: 'checkbox', class: 'brw-ms-cb' });
      cb.checked = checked;
      cb.addEventListener('change', () => {
        const cur = (currentFilters[filter.key] || '').split(',').filter(Boolean);
        if (cb.checked) {
          if (!cur.includes(opt.value)) cur.push(opt.value);
        } else {
          const idx = cur.indexOf(opt.value);
          if (idx >= 0) cur.splice(idx, 1);
        }
        currentFilters[filter.key] = cur.join(',');
      });
      container.appendChild(el('label', { class: 'brw-ms-label' }, cb, ' ' + opt.label));
    }
    wrapper.appendChild(container);
  } else if (filter.type === 'checkbox') {
    const cb = el('input', { type: 'checkbox' });
    cb.checked = !!currentFilters[filter.key];
    cb.addEventListener('change', () => { currentFilters[filter.key] = cb.checked; });
    wrapper.appendChild(cb);
  }

  return wrapper;
}

function buildFilterChips() {
  if (!activeSource || !activeSource.getFilters) return null;
  let filterDefs;
  try { filterDefs = activeSource.getFilters(); } catch { return null; }
  if (!filterDefs || filterDefs.length === 0) return null;

  const defaults = getDefaultFilters(activeSource);
  const chips = [];

  for (const f of filterDefs) {
    if (currentFilters[f.key] === defaults[f.key]) continue;
    if (f.key === 'sort') continue; // don't chip sort

    if (f.type === 'multiselect') {
      const vals = (currentFilters[f.key] || '').split(',').filter(Boolean);
      for (const v of vals) {
        const opt = f.options.find(o => o.value === v);
        const label = f.name + ': ' + (opt ? opt.label : v);
        chips.push({ key: f.key, value: v, label, multi: true });
      }
    } else if (f.type === 'select') {
      const opt = f.options.find(o => o.value === currentFilters[f.key]);
      const label = f.name + ': ' + (opt ? opt.label : currentFilters[f.key]);
      chips.push({ key: f.key, value: '', label, multi: false });
    } else if (f.type === 'text' && currentFilters[f.key]) {
      chips.push({ key: f.key, value: '', label: f.name + ': ' + currentFilters[f.key], multi: false });
    } else if (f.type === 'checkbox' && currentFilters[f.key]) {
      chips.push({ key: f.key, value: '', label: f.name, multi: false });
    }
  }

  if (chips.length === 0) return null;

  const row = el('div', { class: 'brw-chips' });
  for (const chip of chips) {
    const removeBtn = el('span', {
      class: 'brw-chip-remove',
      onClick: (e) => {
        e.stopPropagation();
        if (chip.multi) {
          const cur = (currentFilters[chip.key] || '').split(',').filter(Boolean);
          const idx = cur.indexOf(chip.value);
          if (idx >= 0) cur.splice(idx, 1);
          currentFilters[chip.key] = cur.join(',');
        } else {
          currentFilters[chip.key] = defaults[chip.key];
        }
        results = [];
        currentPage = 1;
        doFetch();
      },
    }, '\u00D7');
    row.appendChild(el('span', { class: 'brw-chip' }, chip.label, removeBtn));
  }
  return row;
}

// --- Results ---

function buildResultsArea() {
  if (selectedDetail) return buildDetailView();

  scrollableEl = el('lol-uikit-scrollable', {
    class: 'brw-results',
    'overflow-masks': 'enabled',
  });

  if (loading && results.length === 0) {
    scrollableEl.appendChild(el('div', { class: 'brw-results-inner' },
      el('div', { class: 'csm-empty' }, t('ui.loading')),
    ));
    return scrollableEl;
  }

  if (results.length === 0 && !loading) {
    scrollableEl.appendChild(el('div', { class: 'brw-results-inner' },
      el('div', { class: 'csm-empty' }, t('browse.no_results')),
    ));
    return scrollableEl;
  }

  gridEl = el('div', { class: 'csm-grid' });
  for (const item of results) {
    gridEl.appendChild(buildBrowseCard(item));
  }

  const inner = el('div', { class: 'brw-results-inner' }, gridEl);

  if (hasNextPage) {
    loadMoreRow = buildLoadMoreRow();
    inner.appendChild(loadMoreRow);
  }

  scrollableEl.appendChild(inner);
  return scrollableEl;
}

function buildLoadMoreRow() {
  const btn = createButton(t('browse.load_more'), {
    onClick: () => {
      currentPage++;
      doFetch(true);
    },
  });
  if (loading) btn.setAttribute('disabled', '');
  return el('div', { class: 'brw-load-more-row' }, btn);
}

function appendResults(newItems) {
  if (!gridEl) return;

  for (const item of newItems) {
    gridEl.appendChild(buildBrowseCard(item));
  }

  if (loadMoreRow) {
    loadMoreRow.remove();
    loadMoreRow = null;
  }
  if (hasNextPage) {
    loadMoreRow = buildLoadMoreRow();
    gridEl.parentElement.appendChild(loadMoreRow);
  }
}

function buildBrowseCard(item) {
  const imgSrc = item.thumbnailUrl
    ? PROXY_IMAGE_BASE + encodeURIComponent(item.thumbnailUrl)
    : '';

  const imgEl = createCardImage(imgSrc);

  return el('div', {
    class: 'csm-card brw-card',
    onClick: () => openDetail(item),
  },
    imgEl,
    el('div', { class: 'csm-card-info' },
      el('div', { class: 'csm-card-name', title: item.title }, item.title),
      el('div', { class: 'csm-card-author' }, item.author ? `by ${item.author}` : ''),
    ),
  );
}

// --- Detail View ---

async function openDetail(item) {
  if (!activeSource) return;
  selectedDetail = { ...item, loading: true };
  renderBrowse();

  try {
    const detail = await activeSource.getDetails(item);
    selectedDetail = { ...detail, loading: false };
  } catch (err) {
    console.error('[browse] getDetails error:', err);
    selectedDetail = { ...item, loading: false, description: 'Failed to load details.' };
  }
  renderBrowse();
}

function buildDetailView() {
  const d = selectedDetail;
  const area = el('lol-uikit-scrollable', {
    class: 'brw-detail',
    'overflow-masks': 'enabled',
  });

  const inner = el('div', { class: 'brw-detail-inner' });

  // Back button
  inner.appendChild(el('button', {
    class: 'brw-back-btn',
    onClick: () => { selectedDetail = null; renderBrowse(); },
  }, '\u2190 ' + t('browse.back')));

  if (d.loading) {
    inner.appendChild(el('div', { class: 'csm-empty' }, t('ui.loading')));
    area.appendChild(inner);
    return area;
  }

  // Images — hero with spinner + optional thumbnails
  const images = d.images || [];
  if (images.length > 0) {
    const heroWrap = el('div', { class: 'brw-hero-wrap' },
      el('div', { class: 'brw-img-spinner' }),
    );
    const heroImg = el('img', {
      class: 'brw-hero-img',
      src: PROXY_IMAGE_BASE + encodeURIComponent(images[0]),
    });
    heroImg.addEventListener('load', () => heroWrap.classList.add('loaded'));
    heroImg.addEventListener('error', () => heroWrap.classList.add('loaded'));
    heroWrap.appendChild(heroImg);
    inner.appendChild(heroWrap);

    if (images.length > 1) {
      const thumbRow = el('div', { class: 'brw-thumb-row' });
      for (let i = 0; i < images.length; i++) {
        const imgUrl = images[i];
        const thumb = el('img', {
          class: 'brw-thumb' + (i === 0 ? ' active' : ''),
          src: PROXY_IMAGE_BASE + encodeURIComponent(imgUrl),
          onClick: () => {
            heroWrap.classList.remove('loaded');
            heroImg.src = PROXY_IMAGE_BASE + encodeURIComponent(imgUrl);
            thumbRow.querySelectorAll('.brw-thumb').forEach(t => t.classList.remove('active'));
            thumb.classList.add('active');
          },
        });
        thumbRow.appendChild(thumb);
      }
      inner.appendChild(thumbRow);
    }
  }

  // Title + author
  inner.appendChild(el('div', { class: 'brw-detail-title' }, d.title || ''));
  if (d.author) {
    inner.appendChild(el('div', { class: 'brw-detail-author' }, t('browse.detail_by', { author: d.author })));
  }

  // Description
  if (d.description) {
    inner.appendChild(el('div', { class: 'brw-detail-desc' }, d.description));
  }

  // Downloads
  const downloads = d.downloads || [];
  if (downloads.length > 0) {
    const dlRow = el('div', { class: 'brw-download-row' });
    for (const dl of downloads) {
      const dlId = d.id + '-' + dl.label;
      const prog = downloadProgress[dlId];
      const dlContainer = el('div', { class: 'brw-dl-item', dataset: { dlId: dlId } });

      if (prog && prog.phase === 'done') {
        const btn = createButton('\u2713 ' + t('browse.download_complete'), {});
        btn.setAttribute('disabled', '');
        dlContainer.appendChild(btn);
      } else if (prog && prog.phase === 'error') {
        const btn = createButton(t('browse.retry'), {
          onClick: () => startDownload(dl, d, dlId),
        });
        btn.classList.add('brw-dl-error-btn');
        dlContainer.appendChild(btn);
      } else if (prog) {
        const pct = prog.progress >= 0 ? Math.round(prog.progress) + '%' : '';
        const phaseText = prog.phase === 'downloading' ? t('browse.downloading') : t('browse.importing');
        dlContainer.appendChild(el('div', { class: 'brw-progress' },
          el('div', { class: 'brw-progress-bar', style: { width: prog.progress >= 0 ? prog.progress + '%' : '100%' } }),
          el('span', { class: 'brw-progress-text' }, phaseText + (pct ? ' ' + pct : '')),
        ));
      } else {
        const label = dl.label || t('browse.detail_download');
        dlContainer.appendChild(createButton(label, {
          onClick: () => startDownload(dl, d, dlId),
        }));
      }
      dlRow.appendChild(dlContainer);
    }
    inner.appendChild(dlRow);
  }

  area.appendChild(inner);
  return area;
}

function startDownload(dl, detail, dlId) {
  downloadProgress[dlId] = { phase: 'downloading', progress: 0 };
  updateDownloadUI(dlId);

  startBrowseDownload(
    dl.url,
    detail.title || 'Unknown',
    detail.author || '',
    detail.championId || 0,
    detail.images?.[0] || '',
    (msg) => {
      downloadProgress[dlId] = msg;
      updateDownloadUI(dlId);
    },
  );
}

function updateDownloadUI(dlId) {
  if (!browseRoot) return;
  const container = browseRoot.querySelector(`[data-dl-id="${dlId}"]`);
  if (!container) return;

  const prog = downloadProgress[dlId];
  if (!prog) return;

  container.innerHTML = '';

  if (prog.phase === 'done') {
    const btn = createButton('\u2713 ' + t('browse.download_complete'), {});
    btn.setAttribute('disabled', '');
    container.appendChild(btn);
  } else if (prog.phase === 'error') {
    container.appendChild(el('span', { class: 'brw-dl-error' }, prog.error || t('browse.download_failed')));
  } else {
    const pct = prog.progress >= 0 ? Math.round(prog.progress) + '%' : '';
    const phaseText = prog.phase === 'downloading' ? t('browse.downloading') : t('browse.importing');
    container.appendChild(el('div', { class: 'brw-progress' },
      el('div', { class: 'brw-progress-bar', style: { width: prog.progress >= 0 ? prog.progress + '%' : '100%' } }),
      el('span', { class: 'brw-progress-text' }, phaseText + (pct ? ' ' + pct : '')),
    ));
  }
}


// --- Extensions Manager Overlay ---

function openExtensionsOverlay() {
  const modal = document.getElementById('ame-custom-skins-modal');
  if (!modal) return;

  const overlay = el('div', { class: 'csm-subdialog-overlay' });
  const content = el('div', { class: 'csm-subdialog brw-ext-dialog' });

  content.appendChild(el('div', { class: 'csm-header' },
    el('span', { class: 'csm-title' }, t('browse.extensions_title')),
    el('button', { class: 'csm-close', onClick: () => overlay.remove() }, '\u00D7'),
  ));

  const body = el('div', { class: 'brw-ext-body' });
  content.appendChild(body);

  function renderExtList() {
    body.innerHTML = '';
    const sources = getLoadedSources();

    if (sources.length > 0) {
      for (const src of sources) {
        const row = el('div', { class: 'brw-ext-row' },
          el('span', { class: 'brw-ext-name' }, src.name),
          el('button', {
            class: 'csm-icon-btn csm-delete',
            onClick: async () => {
              await removeExtension(src._filename);
              if (activeSource?.id === src.id) {
                activeSource = null;
                results = [];
              }
              renderExtList();
              renderBrowse();
            },
          }, '\u2716'),
        );
        body.appendChild(row);
      }
    } else {
      body.appendChild(el('div', { class: 'brw-ext-empty' }, t('browse.no_sources')));
    }

    const urlInput = createInput({ placeholder: t('browse.add_placeholder') });
    const addBtn = createButton(t('browse.add_extension'), {
      onClick: async () => {
        const url = urlInput.input.value.trim();
        if (!url) return;
        addBtn.setAttribute('disabled', '');
        await addExtension(url);
        urlInput.input.value = '';
        addBtn.removeAttribute('disabled');
        renderExtList();
        renderBrowse();
      },
    });

    const fileBtn = createButton(t('browse.add_from_file'), {
      onClick: async () => {
        fileBtn.setAttribute('disabled', '');
        await addExtensionFromFile();
        fileBtn.removeAttribute('disabled');
        renderExtList();
        renderBrowse();
      },
    });

    body.appendChild(el('div', { class: 'brw-ext-add-row' }, urlInput.container, addBtn));
    body.appendChild(el('div', { class: 'brw-ext-add-row' }, fileBtn));
  }

  renderExtList();
  overlay.appendChild(content);
  content.addEventListener('click', e => e.stopPropagation());
  overlay.addEventListener('click', () => overlay.remove());
  modal.appendChild(overlay);
}

function buildExtensionsManager() {
  const container = el('div', { class: 'brw-ext-inline' });
  const sources = getLoadedSources();

  if (sources.length > 0) {
    for (const src of sources) {
      container.appendChild(el('div', { class: 'brw-ext-row' },
        el('span', { class: 'brw-ext-name' }, src.name),
      ));
    }
  }

  const urlInput = createInput({ placeholder: t('browse.add_placeholder') });
  const addBtn = createButton(t('browse.add_extension'), {
    onClick: async () => {
      const url = urlInput.input.value.trim();
      if (!url) return;
      addBtn.setAttribute('disabled', '');
      await addExtension(url);
      urlInput.input.value = '';
      addBtn.removeAttribute('disabled');
      renderBrowse();
    },
  });

  const fileBtn = createButton(t('browse.add_from_file'), {
    onClick: async () => {
      fileBtn.setAttribute('disabled', '');
      await addExtensionFromFile();
      fileBtn.removeAttribute('disabled');
      renderBrowse();
    },
  });

  container.appendChild(el('div', { class: 'brw-ext-add-row' }, urlInput.container, addBtn));
  container.appendChild(el('div', { class: 'brw-ext-add-row' }, fileBtn));
  return container;
}

// --- Fetch ---

async function doFetch(append = false) {
  if (!activeSource) return;
  loading = true;

  if (append) {
    // Disable load more button without full re-render
    if (loadMoreRow) {
      const btn = loadMoreRow.querySelector('lol-uikit-flat-button');
      if (btn) btn.setAttribute('disabled', '');
    }
  } else {
    renderBrowse();
  }

  try {
    let page;
    if (currentQuery || hasActiveFilters()) {
      page = await activeSource.search(currentQuery, currentPage, currentFilters);
    } else if (browseMode === 'latest' && activeSource.getLatest) {
      page = await activeSource.getLatest(currentPage);
    } else {
      page = await activeSource.getPopular(currentPage);
    }

    if (append) {
      const newItems = page.items || [];
      results = [...results, ...newItems];
      hasNextPage = !!page.hasNextPage;
      loading = false;
      appendResults(newItems);
    } else {
      results = page.items || [];
      hasNextPage = !!page.hasNextPage;
      loading = false;
      renderBrowse();
    }
  } catch (err) {
    console.error('[browse] fetch error:', err);
    if (!append) results = [];
    hasNextPage = false;
    loading = false;
    if (!append) renderBrowse();
  }
}
