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
});
