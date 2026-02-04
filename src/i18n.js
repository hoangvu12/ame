import { fetchJson } from './api';

const FALLBACK_LOCALE = 'en_US';
const MANIFEST_URL = new URL('./locales/manifest.json', import.meta.url);

let currentLocale = FALLBACK_LOCALE;
let messages = {};
let fallbackMessages = {};

async function loadJson(url) {
  try {
    const res = await fetch(url);
    if (!res.ok) return null;
    return await res.json();
  } catch {
    return null;
  }
}

function localeUrl(locale) {
  return new URL(`./locales/${locale}.json`, import.meta.url);
}

function normalizeLocale(input) {
  if (!input || typeof input !== 'string') return '';
  const cleaned = input.trim().replace('-', '_');
  const parts = cleaned.split('_').filter(Boolean);
  if (parts.length === 0) return '';
  if (parts.length === 1) return parts[0].toLowerCase();
  return `${parts[0].toLowerCase()}_${parts[1].toUpperCase()}`;
}

function resolveLocale(preferred, available) {
  if (!available || available.length === 0) return FALLBACK_LOCALE;
  const normalized = normalizeLocale(preferred);
  if (normalized && available.includes(normalized)) return normalized;

  const lang = normalized ? normalized.split('_')[0] : '';
  if (lang) {
    const match = available.find(loc => loc.toLowerCase().startsWith(`${lang}_`));
    if (match) return match;
  }

  return available.includes(FALLBACK_LOCALE) ? FALLBACK_LOCALE : available[0];
}

async function detectPreferredLocale() {
  const lcu = await fetchJson('/riotclient/region-locale');
  if (lcu && lcu.locale) return lcu.locale;
  if (typeof navigator !== 'undefined') {
    return navigator.language || (navigator.languages && navigator.languages[0]) || '';
  }
  return FALLBACK_LOCALE;
}

async function loadMessages(locale) {
  if (!locale) return null;
  return await loadJson(localeUrl(locale));
}

function getValue(obj, path) {
  if (!obj || !path) return undefined;
  return path.split('.').reduce((acc, part) => (acc && acc[part] !== undefined ? acc[part] : undefined), obj);
}

function interpolate(template, vars) {
  if (!vars) return template;
  return template.replace(/\{(\w+)\}/g, (match, key) => {
    if (vars[key] === undefined || vars[key] === null) return match;
    return String(vars[key]);
  });
}

export async function initI18n() {
  const manifest = await loadJson(MANIFEST_URL);
  const available = Array.isArray(manifest) ? manifest : [FALLBACK_LOCALE];

  fallbackMessages = (await loadMessages(FALLBACK_LOCALE)) || {};

  const preferred = await detectPreferredLocale();
  const targetLocale = resolveLocale(preferred, available);
  const loaded = await loadMessages(targetLocale);

  messages = loaded || fallbackMessages;
  currentLocale = loaded ? targetLocale : FALLBACK_LOCALE;
}

export function t(key, vars) {
  const value = getValue(messages, key) ?? getValue(fallbackMessages, key) ?? key;
  if (typeof value !== 'string') return key;
  return interpolate(value, vars);
}

export function getLocale() {
  return currentLocale;
}
