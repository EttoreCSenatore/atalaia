import { en, type Messages } from '../messages/en';
import { ptBR } from '../messages/pt-BR';

export type Lang = 'en' | 'pt-BR';
const catalogs: Record<Lang, Messages> = { en, 'pt-BR': ptBR };
const KEY = 'atalaia.lang';

export function detectLang(stored: string | null, nav: string | undefined): Lang {
  if (stored === 'en' || stored === 'pt-BR') return stored;
  return (nav ?? '').toLowerCase().startsWith('pt') ? 'pt-BR' : 'en';
}

function initial(): Lang {
  let stored: string | null = null;
  try {
    stored = localStorage.getItem(KEY);
  } catch {
    /* storage can be blocked; the preference is only a convenience */
  }
  return detectLang(stored, typeof navigator === 'undefined' ? undefined : navigator.language);
}

export const i18n = $state<{ lang: Lang }>({ lang: initial() });

export function setLang(lang: Lang) {
  i18n.lang = lang;
  try {
    localStorage.setItem(KEY, lang);
  } catch {
    /* ignore */
  }
  document.documentElement.lang = lang;
}

/** Looks up "section.key" (reactive: reads i18n.lang). Falls back to the English text, then to the key. */
export function t(key: string): string {
  return lookup(catalogs[i18n.lang], key) ?? lookup(en, key) ?? key;
}

export function lookup(cat: unknown, key: string): string | undefined {
  let cur: unknown = cat;
  for (const part of key.split('.')) {
    if (typeof cur !== 'object' || cur === null || !(part in cur)) return undefined;
    cur = (cur as Record<string, unknown>)[part];
  }
  return typeof cur === 'string' ? cur : undefined;
}

export function flatKeys(cat: unknown, prefix = ''): string[] {
  if (typeof cat !== 'object' || cat === null) return [prefix];
  return Object.entries(cat).flatMap(([k, v]) => flatKeys(v, prefix ? `${prefix}.${k}` : k));
}

export const catalogsForTests = catalogs;
