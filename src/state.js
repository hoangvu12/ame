// Centralized shared state — single source of truth for cross-module data.
// Avoids circular dependencies between chroma.js ↔ autoApply.js ↔ ui.js.

let lastChampionId = null;
let appliedSkinName = null;
let selectedChromaId = null;
let selectedChromaBaseSkinId = null;
let selectedChromaName = null;
let selectedBaseSkinName = null;
let appliedChromaId = null;

// --- Champion ID ---

export function getLastChampionId() {
  return lastChampionId;
}

export function setLastChampionId(id) {
  lastChampionId = id;
}

// --- Applied skin ---

export function getAppliedSkinName() {
  return appliedSkinName;
}

export function setAppliedSkinName(name) {
  appliedSkinName = name;
}

// --- Selected chroma (user's current chroma pick, not yet applied) ---

export function getSelectedChroma() {
  if (!selectedChromaId) return null;
  return {
    id: selectedChromaId,
    baseSkinId: selectedChromaBaseSkinId,
    chromaName: selectedChromaName,
    baseSkinName: selectedBaseSkinName,
  };
}

export function setSelectedChroma(chromaId, baseSkinId, chromaName = null, baseSkinName = null) {
  selectedChromaId = chromaId;
  selectedChromaBaseSkinId = baseSkinId;
  selectedChromaName = chromaName;
  selectedBaseSkinName = baseSkinName;
}

export function clearSelectedChroma() {
  selectedChromaId = null;
  selectedChromaBaseSkinId = null;
  selectedChromaName = null;
  selectedBaseSkinName = null;
}

// --- Applied chroma (tracks what was actually sent to backend) ---

export function getAppliedChromaId() {
  return appliedChromaId;
}

export function setAppliedChromaId(id) {
  appliedChromaId = id;
}
