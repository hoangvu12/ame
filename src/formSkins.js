import { PROXY_IMAGE_BASE } from './constants';

const LSKINS = 'https://raw.githubusercontent.com/Alban1911/LeagueSkins/main/skins';
const FORM_ICON = 'http://localhost:18765/form-icon/';

function githubPreview(champion, skinName, formN) {
  const form = `${skinName} (Form ${formN})`;
  return PROXY_IMAGE_BASE + encodeURIComponent(`${LSKINS}/${champion}/${skinName}/${form}/${form}.png`);
}

function nestedPreview(champion, parentSkin, subName) {
  return PROXY_IMAGE_BASE + encodeURIComponent(`${LSKINS}/${champion}/${parentSkin}/${subName}/${subName}.png`);
}

// Elementalist Lux — form ID → element name (from LeagueSkins mod metadata)
const LUX_ELEMENTS = [
  'fire', 'water', 'air', 'nature', 'ice', 'dark', 'mystic', 'magma', 'storm',
];

const FORM_SKINS = {
  // Elementalist Lux — 9 forms (Form 2–10)
  99007: () => LUX_ELEMENTS.map((element, i) => {
    const id = 99991 + i;
    return {
      id,
      name: `Elementalist Lux (${element[0].toUpperCase() + element.slice(1)})`,
      colors: [],
      chromaPreviewPath: githubPreview('Lux', 'Elementalist Lux', i + 2),
      iconUrl: `${FORM_ICON}${id}.png`,
    };
  }),

  // DJ Sona — 2 forms (Form 2–3)
  37006: () => [37998, 37999].map((id, i) => ({
    id,
    name: `DJ Sona (Form ${i + 2})`,
    colors: [],
    chromaPreviewPath: githubPreview('Sona', 'DJ Sona', i + 2),
    iconUrl: `${FORM_ICON}${id}.png`,
  })),

  // Spirit Blossom Morgana — 1 form (Form 2)
  25080: () => [{
    id: 25999,
    name: 'Spirit Blossom Morgana (Form 2)',
    colors: [],
    chromaPreviewPath: githubPreview('Morgana', 'Spirit Blossom Morgana', 2),
    iconUrl: `${FORM_ICON}25999.png`,
  }],

  // Sahn-Uzal Mordekaiser — 2 forms (Form 2–3)
  82054: () => [82998, 82999].map((id, i) => ({
    id,
    name: `Sahn-Uzal Mordekaiser (Form ${i + 2})`,
    colors: [],
    chromaPreviewPath: githubPreview('Mordekaiser', 'Sahn-Uzal Mordekaiser', i + 2),
    iconUrl: `${FORM_ICON}${id}.png`,
  })),

  // Risen Legend Ahri — 2 forms (Immortalized + Form 2)
  103085: () => [
    { id: 103086, name: 'Immortalized Legend Ahri', sub: 'Immortalized Legend Ahri' },
    { id: 103087, name: 'Immortalized Legend Ahri (Form 2)', sub: 'Immortalized Legend Ahri (Form 2)' },
  ].map(({ id, name, sub }) => ({
    id,
    name,
    colors: [],
    chromaPreviewPath: nestedPreview('Ahri', 'Risen Legend Ahri', sub),
    iconUrl: `${FORM_ICON}${id}.png`,
  })),

  // Risen Legend Kai'Sa — 2 forms (Immortalized + Form 2)
  145070: () => [
    { id: 145071, name: 'Immortalized Legend Kai\'Sa', sub: 'Immortalized Legend Kai\'Sa' },
    { id: 145999, name: 'Immortalized Legend Kai\'Sa (Form 2)', sub: 'Immortalized Legend Kai\'Sa (Form 2)' },
  ].map(({ id, name, sub }) => ({
    id,
    name,
    colors: [],
    chromaPreviewPath: nestedPreview("Kai'Sa", "Risen Legend Kai'Sa", sub),
    iconUrl: `${FORM_ICON}${id}.png`,
  })),

  // K/DA ALL OUT Seraphine — 2 forms (Rising Star + Superstar)
  147001: () => [
    { id: 147002, name: 'K/DA ALL OUT Seraphine Rising Star', sub: 'KDA ALL OUT Seraphine Rising Star' },
    { id: 147003, name: 'K/DA ALL OUT Seraphine Superstar', sub: 'KDA ALL OUT Seraphine Superstar' },
  ].map(({ id, name, sub }) => ({
    id,
    name,
    colors: [],
    chromaPreviewPath: nestedPreview('Seraphine', 'KDA ALL OUT Seraphine Indie', sub),
    iconUrl: `${FORM_ICON}${id}.png`,
  })),

  // Arcane Fractured Jinx — 2 forms (Form 2–3)
  222060: () => [222998, 222999].map((id, i) => ({
    id,
    name: `Arcane Fractured Jinx (Form ${i + 2})`,
    colors: [],
    chromaPreviewPath: githubPreview('Jinx', 'Arcane Fractured Jinx', i + 2),
    iconUrl: `${FORM_ICON}${id}.png`,
  })),

  // Revenant Reign Viego — 6 forms (Form 2–7)
  234043: () => [234994, 234995, 234996, 234997, 234998, 234999].map((id, i) => ({
    id,
    name: `Revenant Reign Viego (Form ${i + 2})`,
    colors: [],
    chromaPreviewPath: githubPreview('Viego', 'Revenant Reign Viego', i + 2),
    iconUrl: `${FORM_ICON}${id}.png`,
  })),

  // Radiant Serpent Sett — 2 forms (Form 2–3)
  875066: () => [875998, 875999].map((id, i) => ({
    id,
    name: `Radiant Serpent Sett (Form ${i + 2})`,
    colors: [],
    chromaPreviewPath: githubPreview('Sett', 'Radiant Serpent Sett', i + 2),
    iconUrl: `${FORM_ICON}${id}.png`,
  })),
};

export function getFormChromas(skinId) {
  const builder = FORM_SKINS[skinId];
  return builder ? builder() : null;
}
