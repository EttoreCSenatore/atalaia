/** "5 minutes ago", "in 2 hours", localised; `none` is returned for an empty/zero/invalid date. */
export function relativeTime(iso: string | undefined, lang: string, now: Date = new Date(), none = '-', justNow = 'just now'): string {
  if (!iso) return none;
  const then = new Date(iso);
  if (Number.isNaN(then.getTime()) || then.getUTCFullYear() < 2000) return none;
  const seconds = Math.round((then.getTime() - now.getTime()) / 1000);
  const abs = Math.abs(seconds);
  if (abs < 45) return justNow;
  const rtf = new Intl.RelativeTimeFormat(lang, { numeric: 'auto' });
  if (abs < 3600) return rtf.format(Math.round(seconds / 60), 'minute');
  if (abs < 86400) return rtf.format(Math.round(seconds / 3600), 'hour');
  return rtf.format(Math.round(seconds / 86400), 'day');
}

export function dateTime(iso: string, lang: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return new Intl.DateTimeFormat(lang, { dateStyle: 'medium', timeStyle: 'short' }).format(d);
}

/** Best label for a device: its name, else vendor + MAC, else the MAC. */
export function deviceLabel(d: { name: string; vendor: string; mac: string }): string {
  if (d.name) return d.name;
  return d.vendor ? `${d.vendor} (${d.mac})` : d.mac;
}

/** Sort key so "192.168.1.9" comes before "192.168.1.10". */
export function ipSortKey(ip: string): number[] {
  return ip.split('.').map((n) => Number.parseInt(n, 10) || 0);
}

export function compareIp(a: string, b: string): number {
  const x = ipSortKey(a);
  const y = ipSortKey(b);
  for (let i = 0; i < 4; i++) if ((x[i] ?? 0) !== (y[i] ?? 0)) return (x[i] ?? 0) - (y[i] ?? 0);
  return 0;
}
