export function el(tag, attrs, ...children) {
  const element = document.createElement(tag);

  if (attrs !== null && attrs !== undefined &&
      (typeof attrs !== 'object' || attrs instanceof Node || Array.isArray(attrs))) {
    children.unshift(attrs);
    attrs = null;
  }

  if (attrs) {
    for (const [key, value] of Object.entries(attrs)) {
      if (key === 'class') {
        element.classList.add(...value.split(' ').filter(Boolean));
      } else if (key === 'id') {
        element.id = value;
      } else if (key === 'style' && typeof value === 'object') {
        Object.assign(element.style, value);
      } else if (key === 'dataset' && typeof value === 'object') {
        Object.assign(element.dataset, value);
      } else if (key.startsWith('on') && typeof value === 'function') {
        element.addEventListener(key.slice(2).toLowerCase(), value);
      } else {
        element.setAttribute(key, value);
      }
    }
  }

  appendChildren(element, children);
  return element;
}

function appendChildren(parent, children) {
  for (const child of children) {
    if (child == null || child === false) continue;
    if (typeof child === 'string') {
      parent.appendChild(document.createTextNode(child));
    } else if (Array.isArray(child)) {
      appendChildren(parent, child);
    } else {
      parent.appendChild(child);
    }
  }
}

export function ensureElement(id, containerSelector, buildFn) {
  const existing = document.getElementById(id);
  if (existing) return existing;
  const container = document.querySelector(containerSelector);
  if (!container) return null;
  const element = buildFn(container);
  element.id = id;
  container.appendChild(element);
  return element;
}

export function removeElement(id) {
  document.getElementById(id)?.remove();
}
