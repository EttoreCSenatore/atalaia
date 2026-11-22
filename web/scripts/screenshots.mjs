// Takes the documentation screenshots of the web app with FICTITIOUS data (API calls are answered by tests/fixtures.mjs).
// Usage: PLAYWRIGHT=/path/to/playwright/index.mjs node scripts/screenshots.mjs <output-folder>
// Writes app-<page>.png (Portuguese) and app-<page>-en.png (English). Run `npm run build` first.
import { spawn } from 'node:child_process';
import { alerts, audit, dhcp, devices, exposure, ports, events, releases, settings, status } from '../tests/fixtures.mjs';

const {chromium} = await import(process.env.PLAYWRIGHT ?? 'playwright');
const out = process.argv[2];
if (!out) throw new Error('usage: node scripts/screenshots.mjs <output-folder>');
const port = 4173;
const server = spawn('npx', ['vite', 'preview', '--host', '127.0.0.1', '--port', String(port), '--strictPort'], { stdio: 'ignore' });
const base = `http://127.0.0.1:${port}`;
for (let i = 0; i < 50; i++) {
  try { await fetch(base); break; } catch { await new Promise((r) => setTimeout(r, 200)); }
}

const json = (body, code = 200) => ({ status: code, contentType: 'application/json', body: JSON.stringify(body) });
const browser = await chromium.launch();
try {
  for (const [lang, suffix] of [['pt-BR', ''], ['en', '-en']]) {
    const shot = async (name, path, { auth = true, setup = false, width = 1100, height = 640, seen = '', disk } = {}) => {
      const ctx = await browser.newContext({ viewport: { width, height }, colorScheme: 'light', locale: lang });
      await ctx.addInitScript((l) => localStorage.setItem('atalaia.lang', l), lang);
      // the newest release this browser has seen: an older one shows the "updated" notice
      if (seen) await ctx.addInitScript((v) => localStorage.setItem('atalaia.seenRelease', v), seen);
      // a stand-in for the live stream, so the page shows "live updates on"
      await ctx.addInitScript(() => {
        window.EventSource = class { constructor() { setTimeout(() => this.onopen?.(), 0); } addEventListener() {} close() {} };
      });
      const page = await ctx.newPage();
      await page.route('**/api/**', (route) => {
        const p = new URL(route.request().url()).pathname;
        if (p === '/api/status') return route.fulfill(json(status({ authenticated: auth, setup_required: setup, ...(auth ? {} : { scan: undefined, disk: undefined }), ...(disk ? { disk } : {}) })));
        if (p === '/api/devices') return route.fulfill(json(devices));
        if (p === '/api/events') return route.fulfill(json(events));
        if (p === '/api/dhcp') return route.fulfill(json(dhcp));
        if (p === '/api/audit') return route.fulfill(json(audit));
        if (/^\/api\/devices\/[^/]+\/exposure$/.test(p)) return route.fulfill(json(exposure));
        if (/^\/api\/devices\/[^/]+\/ports$/.test(p)) return route.fulfill(json(ports));
        if (p === '/api/alerts') return route.fulfill(json(alerts));
        if (p === '/api/settings') return route.fulfill(json(settings));
        if (p === '/api/releases') return route.fulfill(json(releases));
        if (p === '/api/stream') return route.fulfill({ status: 200, contentType: 'text/event-stream', body: 'event: hello\ndata: {}\n\n' });
        return route.fulfill(json({ error: 'not mocked' }, 404));
      });
      await page.goto(base + path);
      await page.waitForSelector('h1');
      await page.waitForTimeout(300);
      await page.screenshot({ path: `${out}/app-${name}${suffix}.png` });
      await ctx.close();
    };
    await shot('painel', '/', { height: 740 });
    await shot('dispositivo', '/devices/b8:27:eb:00:00:0a', { height: 1900 });
    await shot('historico', '/history', { height: 560 });
    await shot('configuracoes', '/settings', { height: 700 });
    await shot('ferramentas', '/tools', { height: 760 });
    await shot('alertas', '/settings', { height: 1700 });
    await shot('novidades', '/whats-new', { height: 760, seen: '0.5.0' });
    await shot('aviso-novidades', '/', { height: 300, seen: '0.5.0' });
    await shot('disco-cheio', '/', { height: 300, disk: { free_bytes: 300 * 2 ** 20, total_bytes: 32 * 2 ** 30, low: true } });
    await shot('login', '/', { auth: false, width: 640, height: 460 });
    await shot('configuracao-inicial', '/', { auth: false, setup: true, width: 640, height: 640 });
  }
} finally {
  await browser.close();
  server.kill();
}
