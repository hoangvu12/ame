import { el } from './dom.js';

export function createButton(text, opts = {}) {
  const attrs = {};
  if (opts.id) attrs.id = opts.id;
  if (opts.class) attrs.class = opts.class;
  if (opts.onClick) attrs.onClick = opts.onClick;
  const btn = el('lol-uikit-flat-button', attrs, text);
  if (opts.disabled) btn.setAttribute('disabled', '');
  return btn;
}

export function createCheckbox(id, labelText, onChange) {
  const input = el('input', { slot: 'input', name: id, type: 'checkbox', id });
  if (onChange) input.addEventListener('change', () => onChange(input.checked));
  const checkbox = el('lol-uikit-flat-checkbox', { for: id },
    input,
    el('label', { slot: 'label' }, labelText),
  );
  return { checkbox, input };
}

/**
 * Creates a card image element with loading spinner.
 * Returns the wrapper element (or a placeholder if no src).
 */
export function createCardImage(src, placeholder = '\u{1F3AE}') {
  if (!src) {
    return el('div', { class: 'csm-card-placeholder' }, placeholder);
  }
  const wrap = el('div', { class: 'csm-card-img-wrap' }, el('div', { class: 'csm-card-spinner' }));
  const img = el('img', { class: 'csm-card-img', src, loading: 'lazy' });
  img.addEventListener('load', () => wrap.classList.add('loaded'));
  img.addEventListener('error', () => wrap.classList.add('loaded'));
  wrap.appendChild(img);
  return wrap;
}

export function createDropdown(options, opts = {}) {
  const attrs = {};
  if (opts.class) attrs.class = opts.class;
  const dropdown = el('lol-uikit-framed-dropdown', attrs,
    ...options.map(opt =>
      el('lol-uikit-dropdown-option', {
        slot: 'lol-uikit-dropdown-option',
        ...(opt.attrs || {}),
        ...(opt.selected ? { selected: '' } : {}),
      }, opt.text)
    )
  );
  if (opts.onChange) {
    dropdown.addEventListener('change', () => {
      const selected = dropdown.querySelector('lol-uikit-dropdown-option[selected]');
      if (selected) opts.onChange(selected);
    });
  }
  return dropdown;
}

export function createInput(opts = {}) {
  const input = el('input', {
    type: opts.type || 'text',
    ...(opts.placeholder ? { placeholder: opts.placeholder } : {}),
  });
  const container = el('lol-uikit-flat-input', null, input);
  return { container, input };
}

/**
 * Searchable champion selector.
 * @param {Object} opts
 * @param {Array}  opts.champions     - Array of { id, name }
 * @param {string} opts.placeholder   - Input placeholder
 * @param {Function} opts.onSelect    - (id, name) => void
 * @param {Function} [opts.getExclude]- () => Set<id> – IDs to hide
 * @param {string}  [opts.allOption]  - Label for "All" entry (id=0)
 * @param {number}  [opts.selected]   - Initial champion ID (0 = all)
 * @param {boolean} [opts.clearOnSelect] - Clear input after selection
 */
export function createChampionSearch(opts = {}) {
  const { champions = [], placeholder = '', onSelect, getExclude, allOption, clearOnSelect = false } = opts;
  let selectedId = opts.selected || 0;

  const { container: flatInput, input } = createInput({ placeholder });

  if (!clearOnSelect) {
    if (selectedId > 0) {
      const c = champions.find(ch => ch.id === selectedId);
      if (c) input.value = c.name;
    } else if (allOption) {
      input.value = allOption;
    }
  }

  const resultsList = el('lol-uikit-scrollable', {
    class: 'ame-search-results',
    'overflow-masks': 'enabled',
  });
  const dropdown = el('div', { class: 'ame-search-dropdown' }, resultsList);
  dropdown.style.display = 'none';

  function showResults(filter) {
    const exclude = getExclude ? getExclude() : new Set();
    const q = (filter || '').toLowerCase();
    let items = champions.filter(c => c.id > 0 && !exclude.has(c.id));
    if (q) items = items.filter(c =>
      c.name.toLowerCase().includes(q) ||
      (c.alias && c.alias.toLowerCase().includes(q)) ||
      (c.englishName && c.englishName.toLowerCase().includes(q))
    );
    items.sort((a, b) => a.name.localeCompare(b.name));

    resultsList.innerHTML = '';

    if (allOption && (!q || allOption.toLowerCase().includes(q))) {
      resultsList.appendChild(el('div', {
        class: 'ame-search-item',
        onClick: () => pick(0, allOption),
      }, el('span', null, allOption)));
    }

    for (const champ of items) {
      resultsList.appendChild(el('div', {
        class: 'ame-search-item',
        onClick: () => pick(champ.id, champ.name),
      },
        el('img', { src: `/lol-game-data/assets/v1/champion-icons/${champ.id}.png` }),
        el('span', null, champ.name),
      ));
    }

    dropdown.style.display = resultsList.childElementCount > 0 ? '' : 'none';
  }

  function pick(id, name) {
    selectedId = id;
    if (clearOnSelect) {
      input.value = '';
    } else {
      input.value = name;
    }
    dropdown.style.display = 'none';
    if (onSelect) onSelect(id, name);
  }

  input.addEventListener('input', () => showResults(input.value));
  input.addEventListener('focus', () => {
    if (!clearOnSelect) input.select();
    showResults('');
  });
  input.addEventListener('blur', () => {
    setTimeout(() => {
      dropdown.style.display = 'none';
      if (!clearOnSelect) {
        if (selectedId > 0) {
          const c = champions.find(ch => ch.id === selectedId);
          input.value = c ? c.name : '';
        } else if (allOption) {
          input.value = allOption;
        }
      }
    }, 200);
  });

  const container = el('div', { class: 'ame-champion-search' }, flatInput, dropdown);

  return {
    container,
    getSelectedId() { return selectedId; },
    setSelectedId(id) {
      selectedId = id;
      if (!clearOnSelect) {
        if (id > 0) {
          const c = champions.find(ch => ch.id === id);
          input.value = c ? c.name : '';
        } else if (allOption) {
          input.value = allOption;
        } else {
          input.value = '';
        }
      }
    },
  };
}
