import { defineConfig } from 'vitest/config';
import { svelte } from '@sveltejs/vite-plugin-svelte';

// The app is built straight into the Go package that embeds it.
export default defineConfig({
  plugins: [svelte()],
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
