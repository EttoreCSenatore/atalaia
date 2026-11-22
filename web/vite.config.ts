import { defineConfig } from 'vitest/config';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { createHash } from 'node:crypto';
import { readFileSync, readdirSync } from 'node:fs';
import type { Plugin } from 'vite';

const demo = process.env.VITE_DEMO === '1';

// Files from public/ that the service worker keeps for offline start, besides the built bundle.
function publicFiles(dir = 'public', prefix = ''): string[] {
  return readdirSync(dir, { withFileTypes: true }).flatMap((e) =>
    e.isDirectory() ? publicFiles(`${dir}/${e.name}`, `${prefix}${e.name}/`) : [`${prefix}${e.name}`],
  );
}

// Writes sw.js (from pwa/sw.js) with the exact list of built files and a version that changes
// with them. Not in the demo, which runs on GitHub Pages and must not install a worker there.
function serviceWorker(): Plugin {
  return {
    name: 'atalaia-service-worker',
    apply: 'build',
    generateBundle(_, bundle) {
      if (demo) return;
      const files = ['/index.html', ...[...Object.keys(bundle), ...publicFiles()].filter((f) => f !== 'index.html').sort().map((f) => `/${f}`)];
      const version = createHash('sha256').update(files.join('\n')).digest('hex').slice(0, 12);
      const source = readFileSync('pwa/sw.js', 'utf8')
        .replaceAll('__VERSION__', version)
        .replaceAll('__PRECACHE__', JSON.stringify(files));
      this.emitFile({ type: 'asset', fileName: 'sw.js', source });
    },
  };
}

// The app is built straight into the Go package that embeds it.
export default defineConfig({
  plugins: [svelte(), serviceWorker()],
  // a constant, so the normal build drops the demo code (and its fictitious data) entirely
  define: { 'import.meta.env.VITE_DEMO': JSON.stringify(demo ? '1' : '') },
  build: { outDir: '../internal/web/dist', emptyOutDir: false, sourcemap: false },
  server: {
    port: 5173,
    proxy: {
      // `atalaia serve` runs on 8080; the browser's Origin (5173) would not match the Host the server sees, so drop it.
      '/api': {
        target: process.env.ATALAIA_API ?? 'http://127.0.0.1:8080',
        configure: (proxy) => {
          proxy.on('proxyReq', (req) => req.removeHeader('origin'));
        },
      },
    },
  },
  test: { environment: 'jsdom', include: ['tests/**/*.test.ts'] },
});
