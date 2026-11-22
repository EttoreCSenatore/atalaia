import { describe, expect, it } from 'vitest';
import { codeParts, compareIp, compareVersion, formatBytes, deviceLabel, parseTags, relativeTime } from '../src/lib/format';
import { riskGroup } from '../src/lib/risk';
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
    expect(deviceLabel({ label: 'Server', name: 'pi.local', vendor: 'Raspberry', mac: 'aa' })).toBe('Server');
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
    expect(match('/tools')).toEqual({ name: 'tools' });
    expect(match('/history/')).toEqual({ name: 'history' });
    expect(match('/settings')).toEqual({ name: 'settings' });
    expect(match('/devices/B8:27:EB:00:00:01')).toEqual({ name: 'device', mac: 'b8:27:eb:00:00:01' });
    expect(match('/devices/<script>')).toEqual({ name: 'notfound' });
    expect(match('/api/devices')).toEqual({ name: 'notfound' });
  });
});

describe('parseTags', () => {
  it('trims, lower-cases and removes duplicates and empties', () => {
    expect(parseTags(' IoT, media ,iot,, ')).toEqual(['iot', 'media']);
    expect(parseTags('')).toEqual([]);
  });
});

describe('riskGroup', () => {
  it('groups services and falls back to legacy', () => {
    expect(riskGroup('telnet')).toBe('remoteAdmin');
    expect(riskGroup('redis')).toBe('database');
    expect(riskGroup('smb')).toBe('fileSharing');
    expect(riskGroup('something-new')).toBe('legacy');
  });
});

describe('compareVersion', () => {
  it('orders semver with or without v, pre-releases first', () => {
    expect(compareVersion('0.10.0', '0.9.0')).toBeGreaterThan(0);
    expect(compareVersion('v0.6.0', '0.6.0')).toBe(0);
    expect(compareVersion('1.0.0-rc.1', '1.0.0')).toBeLessThan(0);
    expect(compareVersion('0.5.0', 'v0.6.0')).toBeLessThan(0);
  });
});

describe('codeParts', () => {
  it('puts code spans at odd positions', () => {
    expect(codeParts('run `atalaia backup` now')).toEqual(['run ', 'atalaia backup', ' now']);
    expect(codeParts('plain')).toEqual(['plain']);
  });
});

describe('whats-new route', () => {
  it('matches /whats-new', () => {
    expect(match('/whats-new')).toEqual({ name: 'news' });
  });
});

describe('formatBytes', () => {
  it('uses binary units and the locale', () => {
    expect(formatBytes(512, 'en')).toBe('512 B');
    expect(formatBytes(1536 * 1024 * 1024, 'en')).toBe('1.5 GB');
    expect(formatBytes(1536 * 1024 * 1024, 'pt-BR')).toBe('1,5 GB');
  });
});
