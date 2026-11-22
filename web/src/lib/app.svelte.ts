import { api, ApiError, isDemo, setUnauthorizedHandler, type Device, type DeviceEvent, type Status } from './api';

type Phase = 'loading' | 'error' | 'ready';

export const app = $state({
  phase: 'loading' as Phase,
  status: { version: '', setup_required: false, authenticated: false } as Status,
  devices: [] as Device[],
  events: [] as DeviceEvent[],
  scanning: false,
  scanError: '',
  lastScan: '',
  live: false,
  loadError: '',
});

let source: EventSource | null = null;

export async function refreshStatus() {
  try {
    app.status = await api.status();
    app.phase = 'ready';
    if (app.status.scan) {
      app.scanning = app.status.scan.scanning;
      app.lastScan = app.status.scan.last_scan;
      app.scanError = app.status.scan.last_error;
    }
  } catch (e) {
    app.phase = 'error';
    app.loadError = e instanceof ApiError && e.status === 0 ? 'network' : 'generic';
  }
}

export async function loadData() {
  try {
    const [devices, events] = await Promise.all([api.devices(), api.events(300)]);
    app.devices = devices;
    app.events = events;
  } catch {
    /* a 401 is handled globally; other errors keep the last data on screen */
  }
}

export function connectLive() {
  if (isDemo) {
    app.live = true; // the demo has no server to stream from
    return;
  }
  if (source || typeof EventSource === 'undefined') return;
  source = new EventSource('/api/stream');
  source.onopen = () => (app.live = true);
  source.onerror = () => (app.live = false);
  source.addEventListener('scan_started', () => (app.scanning = true));
  source.addEventListener('scan_finished', async (e) => {
    app.scanning = false;
    app.scanError = '';
    app.lastScan = (JSON.parse((e as MessageEvent).data) as { at: string }).at;
    await loadData();
  });
  source.addEventListener('scan_failed', (e) => {
    app.scanning = false;
    app.scanError = (JSON.parse((e as MessageEvent).data) as { message?: string }).message ?? 'error';
  });
}

export function disconnectLive() {
  source?.close();
  source = null;
  app.live = false;
}

export async function startScan() {
  app.scanning = true;
  try {
    await api.scan();
  } catch {
    app.scanning = false;
  }
}

export async function signedIn() {
  await refreshStatus();
  await loadData();
  connectLive();
}

export async function signOut() {
  try {
    await api.logout();
  } catch {
    /* the cookie is gone either way */
  }
  disconnectLive();
  app.devices = [];
  app.events = [];
  await refreshStatus();
}

export async function boot() {
  setUnauthorizedHandler(() => {
    disconnectLive();
    app.status = { ...app.status, authenticated: false };
  });
  await refreshStatus();
  if (app.status.authenticated) {
    await loadData();
    connectLive();
  }
}
