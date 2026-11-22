import { mount } from 'svelte';
import './app.css';
import App from './App.svelte';
import { isDemo } from './lib/api';

export default mount(App, { target: document.getElementById('app')! });

// Installable app (PWA): the service worker only keeps the app shell (see pwa/sw.js). Browsers allow it
// only in a secure context (HTTPS with a trusted certificate, or localhost); elsewhere the app works
// as before, just not installable. Not in the demo.
if (!isDemo && 'serviceWorker' in navigator && window.isSecureContext) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/sw.js', { scope: '/' }).catch(() => {});
  });
}
