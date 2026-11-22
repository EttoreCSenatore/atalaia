// The public demo is static files on GitHub Pages, which cannot answer /settings with the app:
// there the route lives in the hash (#/settings). The real app uses normal paths.
const hashMode = import.meta.env.VITE_DEMO === '1';

const fromLocation = () => (hashMode ? location.hash.slice(1).split('#')[0] || '/' : location.pathname);

export const route = $state({ path: typeof location === 'undefined' ? '/' : fromLocation() });

export function navigate(to: string) {
  if (to === route.path) return;
  if (hashMode) location.hash = to;
  else history.pushState({}, '', to);
  route.path = to;
  window.scrollTo(0, 0);
}

/** The app path a data-link points to ("/settings" or, in the demo, "#/settings"), without any fragment. */
export function linkPath(href: string): string {
  return (href.startsWith('#') ? href.slice(1) : href).split('#')[0] || '/';
}

if (typeof window !== 'undefined') {
  window.addEventListener(hashMode ? 'hashchange' : 'popstate', () => (route.path = fromLocation()));
  if (hashMode) {
    // a link opened in a new tab must point into the demo too
    document.addEventListener('mousedown', (e) => {
      const a = (e.target as Element | null)?.closest?.('a[data-link]') as HTMLAnchorElement | null;
      const href = a?.getAttribute('href');
      if (a && href?.startsWith('/')) a.setAttribute('href', '#' + linkPath(href));
    });
  }
  // Same-origin links marked data-link are handled in the page, without a reload.
  document.addEventListener('click', (e) => {
    if (e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
    const a = (e.target as Element | null)?.closest?.('a[data-link]') as HTMLAnchorElement | null;
    if (!a || a.origin !== location.origin) return;
    e.preventDefault();
    navigate(hashMode ? linkPath(a.getAttribute('href') ?? '/') : a.pathname);
  });
}

export type Match = { name: 'devices' } | { name: 'device'; mac: string } | { name: 'history' } | { name: 'settings' } | { name: 'tools' } | { name: 'news' } | { name: 'notfound' };

export function match(path: string): Match {
  const p = path.replace(/\/+$/, '') || '/';
  if (p === '/') return { name: 'devices' };
  if (p === '/history') return { name: 'history' };
  if (p === '/tools') return { name: 'tools' };
  if (p === '/settings') return { name: 'settings' };
  if (p === '/whats-new') return { name: 'news' };
  const m = /^\/devices\/([0-9a-f:]{17})$/i.exec(p);
  if (m) return { name: 'device', mac: m[1].toLowerCase() };
  return { name: 'notfound' };
}
