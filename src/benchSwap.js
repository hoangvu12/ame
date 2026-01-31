import { onSetting } from './websocket';
import { el } from './dom';

let enabled = false;
let skipCooldown = false;
let markedEl = null;
let markedId = null;
let markLabel = null;
let swapObserver = null;
let attached = false;

export function loadBenchSwapSetting() {
  onSetting('benchSwap', (v) => { enabled = v; if (!v) clearMark(); });
  onSetting('benchSwapSkipCooldown', (v) => { skipCooldown = v; });
}

function champIdFrom(item) {
  const bg = item.querySelector('.bench-champion-background');
  if (!bg) return null;
  const m = bg.style.backgroundImage.match(/champion-icons\/(\d+)\.png/);
  return m ? Number(m[1]) : null;
}

function hasCooldown(item) {
  for (const c of item.classList) {
    if (c.startsWith('on-cooldown')) return true;
  }
  return false;
}

function clearMark() {
  if (markLabel?.parentNode) markLabel.remove();
  markLabel = null;
  markedEl = null;
  markedId = null;
  if (swapObserver) {
    swapObserver.disconnect();
    swapObserver = null;
  }
}

function doSwap(championId) {
  return fetch(`/lol-champ-select/v1/session/bench/swap/${championId}`, { method: 'POST' })
    .then(r => { console.log('[ame] bench swap', championId, r.status); return r.ok; })
    .catch(() => false);
}

async function mark(item, id) {
  clearMark();
  markedEl = item;
  markedId = id;

  item.style.position = 'relative';
  const dot = el('div', { class: 'ame-bench-mark' });
  item.appendChild(dot);
  markLabel = dot;

  if (skipCooldown) {
    // Try swapping immediately, ignoring cooldown
    if (await doSwap(id)) { clearMark(); return; }

    // Swap rejected — if no longer on cooldown or detached, give up
    if (!markedEl || !markedEl.isConnected || !hasCooldown(markedEl)) {
      clearMark();
      return;
    }
  }

  // Already off cooldown — swap now
  if (!hasCooldown(item)) {
    const swapId = markedId;
    clearMark();
    doSwap(swapId);
    return;
  }

  // Wait for cooldown to end, then swap
  swapObserver = new MutationObserver(() => {
    if (!markedEl) return;

    if (!markedEl.isConnected || markedEl.classList.contains('empty-bench-item')) {
      clearMark();
      return;
    }

    if (!hasCooldown(markedEl)) {
      const swapId = markedId;
      clearMark();
      doSwap(swapId);
    }
  });

  swapObserver.observe(item, { attributes: true, attributeFilter: ['class'] });
}

function onClick(e) {
  if (!enabled) return;
  const item = e.target.closest('.champion-bench-item');
  if (!item) return;
  if (item.classList.contains('empty-bench-item')) return;
  if (!hasCooldown(item)) return;

  const id = champIdFrom(item);
  if (!id) return;

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
}

export function cleanupBenchSwap() {
  clearMark();
  attached = false;
}
