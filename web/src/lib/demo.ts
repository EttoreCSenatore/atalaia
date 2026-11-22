// The public demo (built with VITE_DEMO=1): the API is answered in the browser from the fictitious data
// in tests/fixtures.mjs. Nothing leaves the page and nothing touches a network. Not part of the real app:
// the normal build drops this module (see `isDemo` in api.ts).
import * as fx from '../../tests/fixtures.mjs';
import { ApiError, type Alerts, type Device, type Settings } from './api';
import { i18n } from './i18n.svelte';

const state = {
  devices: structuredClone(fx.devices) as Device[],
  settings: structuredClone(fx.settings) as Settings,
  alerts: structuredClone(fx.alerts) as Alerts,
};

const refused = () =>
  new ApiError(
    501,
    i18n.lang === 'pt-BR'
      ? 'Esta é uma demonstração: nenhuma ação toca uma rede de verdade nem envia nada.'
      : 'This is a demo: no action touches a real network or sends anything.',
  );

const wait = () => new Promise((r) => setTimeout(r, 250));

export async function demoRequest<T>(method: string, path: string, body?: unknown): Promise<T> {
  await wait();
  const url = new URL(path, 'http://demo');
  const p = url.pathname;
  const dev = /^\/api\/devices\/([^/]+)(?:\/([a-z]+))?$/.exec(p);
  const mac = dev ? decodeURIComponent(dev[1]) : '';
  const out = ((): unknown => {
    switch (`${method} ${dev ? `/api/devices/:mac${dev[2] ? '/' + dev[2] : ''}` : p}`) {
      case 'GET /api/status':
        return fx.status();
      case 'POST /api/login':
      case 'POST /api/logout':
        return { ok: true };
      case 'GET /api/devices':
        return state.devices;
      case 'GET /api/events':
        return fx.events;
      case 'POST /api/scan':
        return { status: 'started' };
      case 'PUT /api/devices/:mac': {
        const d = state.devices.find((x) => x.mac === mac);
        if (!d) throw new ApiError(404, 'not found');
        Object.assign(d, body as Partial<Device>);
        return d;
      }
      case 'DELETE /api/devices/:mac':
        state.devices = state.devices.filter((x) => x.mac !== mac);
        return { ok: true };
      case 'GET /api/devices/:mac/ports':
      case 'POST /api/devices/:mac/portscan':
        return fx.ports;
      case 'GET /api/devices/:mac/exposure':
      case 'POST /api/devices/:mac/exposure':
        return fx.exposure;
      case 'GET /api/settings':
        return state.settings;
      case 'PUT /api/settings':
        Object.assign(state.settings, body);
        return state.settings;
      case 'GET /api/alerts':
        return state.alerts;
      case 'PUT /api/alerts':
        state.alerts = structuredClone(body as Alerts);
        return state.alerts;
      case 'GET /api/audit':
        return fx.audit;
      case 'GET /api/dhcp':
        return fx.dhcp;
      case 'GET /api/releases':
        return fx.releases;
      case 'POST /api/backup':
        return fx.status().backup;
    }
    // everything else would act on a network or send something: alerts test, ping, traceroute,
    // speed test, DNS, Wake-on-LAN, DHCP checks, password change
    throw refused();
  })();
  return structuredClone(out) as T;
}

/** The inventory as a file, made in the browser (the real app downloads it from the server). */
export function demoExport(format: 'csv' | 'json'): string {
  const rows = state.devices.map((d) => ({ mac: d.mac, ip: d.ip, name: d.label || d.name, vendor: d.vendor, online: d.online, known: d.known }));
  const text =
    format === 'json'
      ? JSON.stringify(rows, null, 2)
      : ['mac,ip,name,vendor,online,known', ...rows.map((r) => [r.mac, r.ip, r.name, r.vendor, r.online, r.known].map((v) => `"${String(v).replace(/"/g, '""')}"`).join(','))].join('\n');
  return URL.createObjectURL(new Blob([text], { type: format === 'json' ? 'application/json' : 'text/csv' }));
}
