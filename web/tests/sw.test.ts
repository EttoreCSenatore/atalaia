import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';

// Runs pwa/sw.js with a fake `self` and checks which requests it answers.
function load() {
  const listeners: Record<string, (e: unknown) => void> = {};
  const self = { location: { origin: 'https://atalaia.lan' }, addEventListener: (n: string, f: (e: unknown) => void) => (listeners[n] = f) };
  const src = readFileSync('pwa/sw.js', 'utf8').replaceAll('__VERSION__', 'test').replaceAll('__PRECACHE__', '["/index.html"]');
  const never = () => new Promise(() => {}); // the test only checks whether it answers
  new Function('self', 'caches', 'fetch', src)(self, { match: never }, never);
  return (url: string, method = 'GET', mode = 'cors') => {
    let answered = false;
    listeners.fetch({ request: { url, method, mode }, respondWith: () => (answered = true) });
    return answered;
  };
}

describe('service worker', () => {
  const answers = load();
  it('never answers API, metrics, other sites or non-GET requests', () => {
    expect(answers('https://atalaia.lan/api/devices')).toBe(false);
    expect(answers('https://atalaia.lan/api/me', 'GET', 'navigate')).toBe(false);
    expect(answers('https://atalaia.lan/metrics')).toBe(false);
    expect(answers('https://atalaia.lan/sw.js')).toBe(false);
    expect(answers('https://example.com/assets/x.js')).toBe(false);
    expect(answers('https://atalaia.lan/assets/x.js', 'POST')).toBe(false);
  });
  it('answers the app shell and page navigations', () => {
    expect(answers('https://atalaia.lan/assets/index-abc.js')).toBe(true);
    expect(answers('https://atalaia.lan/devices/aa', 'GET', 'navigate')).toBe(true);
  });
});
