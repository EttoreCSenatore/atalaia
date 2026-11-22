<script lang="ts">
  import { api } from '../lib/api';
  import { app, refreshStatus } from '../lib/app.svelte';
  import { t, i18n } from '../lib/i18n.svelte';
  import { dateTime, relativeTime } from '../lib/format';
  import { toolError } from '../lib/errors';

  let busy = $state(false);
  let msg = $state<{ kind: 'ok' | 'error'; text: string } | null>(null);
  const b = $derived(app.status.backup);

  async function backupNow() {
    busy = true;
    msg = null;
    try {
      await api.backupNow();
      await refreshStatus();
      msg = { kind: 'ok', text: t('backup.done') };
    } catch (e) {
      msg = { kind: 'error', text: toolError(e) };
    } finally {
      busy = false;
    }
  }
</script>

<section class="card" id="backups">
  <h2>{t('backup.title')}</h2>
  {#if !b?.enabled}
    <p class="muted">{t('backup.off')}</p>
  {:else}
    <dl>
      <dt>{t('backup.last')}</dt>
      <dd>
        {#if b.last}<span title={dateTime(b.last, i18n.lang)}>{relativeTime(b.last, i18n.lang, new Date(), '-', t('time.justNow'))}</span>{:else}{t('common.never')}{/if}
        {#if b.late}<span class="badge warn">{t('backup.lateBadge')}</span>{/if}
      </dd>
      <dt>{t('backup.restore')}</dt>
      <dd>
        {#if b.last_restore_test}
          {dateTime(b.last_restore_test, i18n.lang)} ·
          {#if b.restore_ok}<span class="pass">{t('backup.restoreOk')}</span>{:else}<span class="fail">{t('backup.restoreFailed')}</span>{/if}
        {:else}{t('common.never')}{/if}
      </dd>
      <dt>{t('backup.where')}</dt>
      <dd><span class="mono">{b.dir}</span> · {t('backup.keep')} {b.keep}</dd>
    </dl>
    <p class="muted help">{t('backup.help')}</p>
    {#if msg}<p class={msg.kind} role="status">{msg.text}</p>{/if}
    <button class="primary" disabled={busy} onclick={backupNow}>{t('backup.now')}</button>
  {/if}
</section>

<style>
  dl {
    margin: 0 0 0.8rem;
  }
  dt {
    color: var(--muted);
    font-size: 0.85rem;
    margin-top: 0.5rem;
  }
  dd {
    margin: 0;
    overflow-wrap: anywhere;
  }
  .badge.warn {
    margin-left: 0.4rem;
    font-size: 0.75rem;
    font-weight: 700;
    padding: 0.1rem 0.5rem;
    border-radius: 999px;
    background: var(--warn-bg);
    color: var(--warn);
  }
  .pass {
    color: var(--ok);
    font-weight: 600;
  }
  .fail {
    color: var(--danger);
    font-weight: 600;
  }
</style>
