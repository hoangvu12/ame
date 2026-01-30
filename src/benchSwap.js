import { wsSend, onBenchSwap } from './websocket';

let enabled = false;
let markedEl = null;
let markedId = null;
let markLabel = null;
let swapObserver = null;
let attached = false;

export function setBenchSwapEnabled(v) {
  enabled = v;
  if (!v) clearMark();
}

export function loadBenchSwapSetting() {
  onBenchSwap((v) => { enabled = v; });
  wsSend({ type: 'getBenchSwap' });
}

function champIdFrom(el) {
  const bg = el.querySelector('.bench-champion-background');
  if (!bg) return null;
  const m = bg.style.backgroundImage.match(/champion-icons\/(\d+)\.png/);
  return m ? Number(m[1]) : null;
}

function hasCooldown(el) {
  for (const c of el.classList) {
    if (c.startsWith('on-cooldown')) return true;
  }
  return false;
}

function clearMark() {
  if (markLabel && markLabel.parentNode) {
    markLabel.remove();
  }
  markLabel = null;
  markedEl = null;
  markedId = null;
  if (swapObserver) {
    swapObserver.disconnect();
    swapObserver = null;
  }
}

function doSwap(championId) {
  fetch(`/lol-champ-select/v1/session/bench/swap/${championId}`, { method: 'POST' })
    .catch(() => {});
}

function mark(el, id) {
  clearMark();
  markedEl = el;
  markedId = id;

  // Inject small gold dot in the top-right corner
  el.style.position = 'relative';
  const dot = document.createElement('div');
  Object.assign(dot.style, {
    position: 'absolute',
    top: '4px',
    right: '4px',
    width: '8px',
    height: '8px',
    borderRadius: '50%',
    background: '#c89b3c',
    boxShadow: '0 0 4px #c89b3c',
    zIndex: '99999',
    pointerEvents: 'none',
  });
  el.appendChild(dot);
  markLabel = dot;

  console.log('[ame] Bench marked champion', id);

  // Already off cooldown — swap immediately
  if (!hasCooldown(el)) {
    const swapId = markedId;
    clearMark();
    doSwap(swapId);
    return;
  }

  swapObserver = new MutationObserver(() => {
    if (!markedEl) return;

    // Element removed from DOM
    if (!markedEl.isConnected) {
      clearMark();
      return;
    }

    // Became empty (teammate took it)
    if (markedEl.classList.contains('empty-bench-item')) {
      clearMark();
      return;
    }

    // Cooldown finished — swap now
    if (!hasCooldown(markedEl)) {
      const swapId = markedId;
      clearMark();
      console.log('[ame] Cooldown ended, swapping', swapId);
      doSwap(swapId);
    }
  });

  swapObserver.observe(el, { attributes: true, attributeFilter: ['class'] });
}

function onClick(e) {
  if (!enabled) return;
  const item = e.target.closest('.champion-bench-item');
  if (!item) return;
  if (item.classList.contains('empty-bench-item')) return;
  if (!hasCooldown(item)) return;

  const id = champIdFrom(item);
  if (!id) return;

  // Toggle off if clicking same one
  if (markedEl === item) {
    clearMark();
    return;
  }

  mark(item, id);
}

export function ensureBenchSwap() {
  if (attached) return;
  const container = document.querySelector('.bench-container');
  if (!container) return;
  attached = true;
  container.addEventListener('click', onClick);
  console.log('[ame] Bench swap ready');
}

export function cleanupBenchSwap() {
  clearMark();
  attached = false;
}
