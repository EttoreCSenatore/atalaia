import { ApiError } from './api';
import { t } from './i18n.svelte';

/** The server's own message for client errors it explains (guards, limits); a generic text otherwise. */
export function toolError(e: unknown): string {
  if (e instanceof ApiError && [400, 403, 404, 429, 501, 502, 504].includes(e.status) && e.message) return e.message;
  return t('common.errorGeneric');
}
