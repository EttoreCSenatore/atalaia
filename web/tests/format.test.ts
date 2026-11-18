import { describe, expect, it } from 'vitest';
import { compareIp, deviceLabel, relativeTime } from '../src/lib/format';
import { match } from '../src/lib/router.svelte';

const now = new Date('2026-11-18T12:00:00Z');

describe('relativeTime', () => {
  it('formats minutes, hours and days', () => {
    expect(relativeTime('2026-11-18T11:55:00Z', 'en', now)).toBe('5 minutes ago');
    expect(relativeTime('2026-11-18T09:00:00Z', 'en', now)).toBe('3 hours ago');
    expect(relativeTime('2026-11-16T12:00:00Z', 'en', now)).toBe('2 days ago');
    expect(relativeTime('2026-11-18T11:55:00Z', 'pt-BR', now)).toBe('há 5 minutos');
  });
  it('says "just now" for a few seconds and "none" for empty or zero dates', () => {
    expect(relativeTime('2026-11-18T11:59:50Z', 'en', now, '-', 'just now')).toBe('just now');
    expect(relativeTime('', 'en', now, 'never')).toBe('never');
    expect(relativeTime('0001-01-01T00:00:00Z', 'en', now, 'never')).toBe('never');
    expect(relativeTime('garbage', 'en', now, 'never')).toBe('never');
  });
});

describe('device helpers', () => {
  it('prefers name, then vendor + MAC, then MAC', () => {
    expect(deviceLabel({ name: 'pi.local', vendor: 'Raspberry', mac: 'aa' })).toBe('pi.local');
    expect(deviceLabel({ name: '', vendor: 'Raspberry', mac: 'aa' })).toBe('Raspberry (aa)');
    expect(deviceLabel({ name: '', vendor: '', mac: 'aa' })).toBe('aa');
  });
  it('sorts IP addresses numerically', () => {
    expect(['192.168.1.10', '192.168.1.9', '10.0.0.2'].sort(compareIp)).toEqual(['10.0.0.2', '192.168.1.9', '192.168.1.10']);
  });
});

describe('router', () => {
  it('matches the known pages and rejects the rest', () => {
    expect(match('/')).toEqual({ name: 'devices' });
    expect(match('/history/')).toEqual({ name: 'history' });
    expect(match('/settings')).toEqual({ name: 'settings' });
    expect(match('/devices/B8:27:EB:00:00:01')).toEqual({ name: 'device', mac: 'b8:27:eb:00:00:01' });
    expect(match('/devices/<script>')).toEqual({ name: 'notfound' });
    expect(match('/api/devices')).toEqual({ name: 'notfound' });
  });
});
