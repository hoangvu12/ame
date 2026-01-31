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
