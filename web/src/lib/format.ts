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

/** Best label for a device: the user's name, else the network name, else vendor + MAC, else the MAC. */
export function deviceLabel(d: { name: string; vendor: string; mac: string; label?: string }): string {
  if (d.label) return d.label;
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

/** "iot, Media ,iot" -> ["iot", "media"]: what the server will keep. */
export function parseTags(input: string): string[] {
  const out: string[] = [];
  for (const raw of input.split(',')) {
    const t = raw.trim().toLowerCase();
    if (t && !out.includes(t)) out.push(t);
  }
  return out;
}

/** Orders two semver versions ("v" prefix optional): negative, 0 or positive. A pre-release sorts before its release. */
export function compareVersion(a: string, b: string): number {
  const parse = (v: string) => {
    const m = /^v?(\d+)\.(\d+)\.(\d+)(?:-(.+))?$/.exec(v.trim());
    return m ? { n: [Number(m[1]), Number(m[2]), Number(m[3])], pre: m[4] ?? '' } : null;
  };
  const x = parse(a);
  const y = parse(b);
  if (!x || !y) return a.localeCompare(b);
  for (let i = 0; i < 3; i++) if (x.n[i] !== y.n[i]) return x.n[i] - y.n[i];
  if (x.pre === y.pre) return 0;
  if (!x.pre) return 1;
  if (!y.pre) return -1;
  return x.pre < y.pre ? -1 : 1;
}

/** Splits text on `backticks` so code spans can be rendered without HTML: odd parts are code. */
export function codeParts(text: string): string[] {
  return text.split('`');
}

/** Human-readable size in binary units (1.5 GB = 1.5 GiB), with the locale's decimal separator. */
export function formatBytes(n: number, lang: string): string {
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let i = 0;
  while (n >= 1024 && i < units.length - 1) {
    n /= 1024;
    i++;
  }
  return `${n.toLocaleString(lang, { maximumFractionDigits: i === 0 ? 0 : 1 })} ${units[i]}`;
}
