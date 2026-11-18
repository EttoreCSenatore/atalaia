export const route = $state({ path: typeof location === 'undefined' ? '/' : location.pathname });

export function navigate(to: string) {
  if (to === route.path) return;
  history.pushState({}, '', to);
  route.path = to;
  window.scrollTo(0, 0);
}

if (typeof window !== 'undefined') {
  window.addEventListener('popstate', () => (route.path = location.pathname));
  // Same-origin links marked data-link are handled in the page, without a reload.
  document.addEventListener('click', (e) => {
    if (e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
    const a = (e.target as Element | null)?.closest?.('a[data-link]') as HTMLAnchorElement | null;
    if (!a || a.origin !== location.origin) return;
    e.preventDefault();
    navigate(a.pathname);
  });
}

export type Match = { name: 'devices' } | { name: 'device'; mac: string } | { name: 'history' } | { name: 'settings' } | { name: 'notfound' };

export function match(path: string): Match {
  const p = path.replace(/\/+$/, '') || '/';
  if (p === '/') return { name: 'devices' };
  if (p === '/history') return { name: 'history' };
  if (p === '/settings') return { name: 'settings' };
  const m = /^\/devices\/([0-9a-f:]{17})$/i.exec(p);
  if (m) return { name: 'device', mac: m[1].toLowerCase() };
  return { name: 'notfound' };
}
