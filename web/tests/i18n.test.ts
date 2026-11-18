import { describe, expect, it } from 'vitest';
import { catalogsForTests, detectLang, flatKeys, lookup } from '../src/lib/i18n.svelte';

describe('i18n', () => {
  it('has the same keys in English and Portuguese', () => {
    const en = flatKeys(catalogsForTests.en).sort();
    const pt = flatKeys(catalogsForTests['pt-BR']).sort();
    expect(pt).toEqual(en);
  });

  it('has no empty texts', () => {
    for (const cat of Object.values(catalogsForTests)) {
      for (const key of flatKeys(cat)) expect(lookup(cat, key), key).toBeTruthy();
    }
  });

  it('detects the language from the stored choice, then the browser', () => {
    expect(detectLang('pt-BR', 'en-US')).toBe('pt-BR');
    expect(detectLang('en', 'pt-BR')).toBe('en');
    expect(detectLang(null, 'pt-BR')).toBe('pt-BR');
    expect(detectLang(null, 'pt-PT')).toBe('pt-BR');
    expect(detectLang(null, 'fr-FR')).toBe('en');
    expect(detectLang('klingon', undefined)).toBe('en');
  });

  it('looks up nested keys and ignores partial paths', () => {
    expect(lookup(catalogsForTests.en, 'history.kinds.joined')).toBe('Joined the network');
    expect(lookup(catalogsForTests.en, 'history.kinds')).toBeUndefined();
    expect(lookup(catalogsForTests.en, 'nope.nope')).toBeUndefined();
  });
});
