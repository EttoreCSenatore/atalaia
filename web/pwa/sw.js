// Service worker do app instalável (PWA). / Service worker for the installable app (PWA).
//
// Guarda só o "casco" do app (HTML, JS, CSS, ícones) para ele abrir sem rede e mostrar
// "Não foi possível falar com o servidor". NUNCA guarda respostas de /api/ nem /metrics:
// os dados da sua rede não ficam no cache do navegador por causa dele.
// Caches only the app shell (HTML, JS, CSS, icons) so it opens without the server and shows
// "Could not reach the server". It NEVER caches /api/ or /metrics responses.
//
// The build (vite.config.ts) fills in the version and the file list below; each build gets a new
// cache name and old caches are deleted when it activates.
const SHELL = 'atalaia-shell-__VERSION__';
const PRECACHE = __PRECACHE__;

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches
      .open(SHELL)
      .then((cache) => cache.addAll(PRECACHE))
      .then(() => self.skipWaiting()),
  );
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches
      .keys()
      .then((keys) => Promise.all(keys.filter((k) => k.startsWith('atalaia-') && k !== SHELL).map((k) => caches.delete(k))))
      .then(() => self.clients.claim()),
  );
});

function passThrough(url) {
  return url.origin !== self.location.origin || url.pathname.startsWith('/api/') || url.pathname === '/metrics' || url.pathname === '/sw.js';
}

self.addEventListener('fetch', (event) => {
  const req = event.request;
  if (req.method !== 'GET') return;
  const url = new URL(req.url);
  if (passThrough(url)) return; // the browser handles it as if there were no service worker
  if (req.mode === 'navigate') {
    // network first, so a running server always gives the newest app; the cached shell only when it is unreachable
    event.respondWith(fetch(req).catch(() => caches.match('/index.html', { cacheName: SHELL }).then((r) => r || Response.error())));
    return;
  }
  event.respondWith(caches.match(req, { cacheName: SHELL }).then((r) => r || fetch(req)));
});
