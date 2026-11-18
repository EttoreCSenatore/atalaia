import { afterEach, describe, expect, it, vi } from 'vitest';
import { api, ApiError, setUnauthorizedHandler } from '../src/lib/api';

function mockFetch(status: number, body: unknown) {
  const fn = vi.fn(async () => new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } }));
  vi.stubGlobal('fetch', fn);
  return fn;
}

afterEach(() => vi.unstubAllGlobals());

describe('api client', () => {
  it('sends the CSRF header only on state-changing requests', async () => {
    const fn = mockFetch(200, { ok: true });
    await api.login('a password');
    const post = fn.mock.calls[0] as unknown as [string, RequestInit];
    expect((post[1].headers as Record<string, string>)['X-Atalaia']).toBe('1');
    expect(post[1].credentials).toBe('same-origin');
    expect(post[1].body).toBe(JSON.stringify({ password: 'a password' }));

    const get = mockFetch(200, []);
    await api.devices();
    expect(((get.mock.calls[0] as unknown as [string, RequestInit])[1].headers as Record<string, string>)['X-Atalaia']).toBeUndefined();
  });

  it('turns error responses into ApiError with the server message', async () => {
    mockFetch(429, { error: 'too many attempts' });
    await expect(api.login('x')).rejects.toMatchObject({ status: 429, message: 'too many attempts' });
  });

  it('reports network failures as status 0', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('failed')));
    await expect(api.status()).rejects.toBeInstanceOf(ApiError);
    await expect(api.status()).rejects.toMatchObject({ status: 0 });
  });

  it('calls the unauthorized handler on 401, except for the login attempt itself', async () => {
    const handler = vi.fn();
    setUnauthorizedHandler(handler);
    mockFetch(401, { error: 'login required' });
    await expect(api.devices()).rejects.toBeInstanceOf(ApiError);
    expect(handler).toHaveBeenCalledOnce();
    await expect(api.login('wrong')).rejects.toBeInstanceOf(ApiError);
    expect(handler).toHaveBeenCalledOnce();
  });

  it('edits a device with PUT and encodes the MAC in the path', async () => {
    const fn = mockFetch(200, {});
    await api.updateDevice('aa:bb:cc:00:00:01', 'TV', ['media'], true);
    const [url, init] = fn.mock.calls[0] as unknown as [string, RequestInit];
    expect(url).toBe('/api/devices/aa%3Abb%3Acc%3A00%3A00%3A01');
    expect(init.method).toBe('PUT');
    expect(init.body).toBe(JSON.stringify({ label: 'TV', tags: ['media'], known: true }));
  });

  it('tests a channel by id', async () => {
    const fn = mockFetch(200, { ok: true });
    await api.testAlert('abc');
    const [url, init] = fn.mock.calls[0] as unknown as [string, RequestInit];
    expect(url).toBe('/api/alerts/test');
    expect(init.body).toBe(JSON.stringify({ id: 'abc' }));
  });

  it('port scan sends an explicit confirmation', async () => {
    const fn = mockFetch(200, { at: '', ports: [] });
    await api.portScan('aa:bb:cc:00:00:01');
    const [url, init] = fn.mock.calls[0] as unknown as [string, RequestInit];
    expect(url).toBe('/api/devices/aa%3Abb%3Acc%3A00%3A00%3A01/portscan');
    expect(init.body).toBe(JSON.stringify({ confirm: true }));
  });

  it('ping and DNS post their parameters', async () => {
    const fn = mockFetch(200, {});
    await api.ping('192.168.1.1', 3);
    await api.dns('example.com', 'MX');
    const bodies = fn.mock.calls.map((c) => (c as unknown as [string, RequestInit])[1].body);
    expect(bodies).toEqual([JSON.stringify({ host: '192.168.1.1', count: 3 }), JSON.stringify({ name: 'example.com', type: 'MX' })]);
  });

  it('exposure check sends an explicit confirmation', async () => {
    const fn = mockFetch(200, { at: '', risks: [], web: [], tls: [] });
    await api.exposureScan('aa:bb:cc:00:00:01');
    const [url, init] = fn.mock.calls[0] as unknown as [string, RequestInit];
    expect(url).toBe('/api/devices/aa%3Abb%3Acc%3A00%3A00%3A01/exposure');
    expect(init.body).toBe(JSON.stringify({ confirm: true }));
  });
});
