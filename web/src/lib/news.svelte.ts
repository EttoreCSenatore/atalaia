import { api, type Release } from './api';
import { compareVersion } from './format';

// The newest release this browser has been shown, kept per browser (a convenience, not shared state).
const KEY = 'atalaia.seenRelease';

export const news = $state({ releases: [] as Release[], seen: '', loaded: false, failed: false });

function readSeen(): string {
  try {
    return localStorage.getItem(KEY) ?? '';
  } catch {
    return '';
  }
}

function writeSeen(v: string) {
  try {
    localStorage.setItem(KEY, v);
  } catch {
    /* private window or blocked storage: the notice just shows again next time */
  }
}

export const latest = () => news.releases[0]?.version ?? '';

/** Loads the release notes. The first time a browser opens the app, nothing is "new": it starts from the running version. */
export async function loadNews() {
  try {
    const r = await api.releases();
    news.releases = r.releases;
    let seen = readSeen();
    if (!seen && latest()) {
      seen = latest();
      writeSeen(seen);
    }
    news.seen = seen;
    news.failed = false;
  } catch {
    news.failed = true;
  }
  news.loaded = true;
}

/** True when a version newer than the last one shown in this browser is running. */
export const hasUnseen = () => !!latest() && !!news.seen && compareVersion(latest(), news.seen) > 0;

export function markSeen() {
  if (latest()) {
    writeSeen(latest());
    news.seen = latest();
  }
}
