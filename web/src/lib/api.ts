export type Device = {
  mac: string;
  ip: string;
  name: string;
  vendor: string;
  randomized_mac: boolean;
  online: boolean;
  first_seen: string;
  last_seen: string;
  label: string;
  tags: string[];
  known: boolean;
};
export type DeviceEvent = { at: string; kind: string; detail: string; mac: string; name: string; vendor: string };
export type ScanStatus = { scanning: boolean; last_scan: string; last_error: string; interval: string };
export type Status = { version: string; setup_required: boolean; authenticated: boolean; scan?: ScanStatus };
export type ChannelType = 'ntfy' | 'telegram' | 'webhook' | 'email';
export type Channel = {
  id: string;
  type: ChannelType;
  name: string;
  enabled: boolean;
  url?: string;
  token?: string;
  chat_id?: string;
  host?: string;
  username?: string;
  password?: string;
  from?: string;
  to?: string;
  has_token?: boolean;
  has_password?: boolean;
};
export type AlertRules = { new_device: boolean; known_offline: boolean };
export type Alerts = { rules: AlertRules; channels: Channel[] };
export type Settings = { retention_days: number; scan_interval: string };

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
  }
}

/** Called when the server says the session is gone, so the UI can show the login page. */
export let onUnauthorized: () => void = () => {};
export function setUnauthorizedHandler(fn: () => void) {
  onUnauthorized = fn;
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = { Accept: 'application/json' };
  if (method !== 'GET') headers['X-Atalaia'] = '1'; // the server refuses state changes without it (CSRF)
  if (body !== undefined) headers['Content-Type'] = 'application/json';
  let res: Response;
  try {
    res = await fetch(path, { method, headers, body: body === undefined ? undefined : JSON.stringify(body), credentials: 'same-origin' });
  } catch {
    throw new ApiError(0, 'network');
  }
  let data: unknown = null;
  try {
    data = await res.json();
  } catch {
    /* empty body */
  }
  if (!res.ok) {
    const msg = (data as { error?: string } | null)?.error ?? res.statusText;
    if (res.status === 401 && path !== '/api/login' && path !== '/api/password') onUnauthorized();
    throw new ApiError(res.status, msg);
  }
  return data as T;
}

export const api = {
  status: () => request<Status>('GET', '/api/status'),
  setup: (code: string, password: string) => request<{ ok: boolean }>('POST', '/api/setup', { code, password }),
  login: (password: string) => request<{ ok: boolean }>('POST', '/api/login', { password }),
  logout: () => request<{ ok: boolean }>('POST', '/api/logout'),
  devices: () => request<Device[]>('GET', '/api/devices'),
  events: (limit = 200) => request<DeviceEvent[]>('GET', `/api/events?limit=${limit}`),
  scan: () => request<{ status: string }>('POST', '/api/scan'),
  settings: () => request<Settings>('GET', '/api/settings'),
  saveSettings: (retention_days: number) => request<Settings>('PUT', '/api/settings', { retention_days }),
  updateDevice: (mac: string, label: string, tags: string[], known: boolean) =>
    request<Device>('PUT', `/api/devices/${encodeURIComponent(mac)}`, { label, tags, known }),
  alerts: () => request<Alerts>('GET', '/api/alerts'),
  saveAlerts: (a: Alerts) => request<Alerts>('PUT', '/api/alerts', a),
  testAlert: (id: string) => request<{ ok: boolean }>('POST', '/api/alerts/test', { id }),
  changePassword: (current: string, next: string) => request<{ ok: boolean }>('POST', '/api/password', { current, new: next }),
};
