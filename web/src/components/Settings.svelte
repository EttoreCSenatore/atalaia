<script lang="ts">
  import { onMount } from 'svelte';
  import { api, ApiError, type Settings } from '../lib/api';
  import AlertsSettings from './AlertsSettings.svelte';
  import { app } from '../lib/app.svelte';
  import { t, i18n, setLang, type Lang } from '../lib/i18n.svelte';

  let settings = $state<Settings | null>(null);
  let days = $state(90);
  let saved = $state(false);
  let current = $state('');
  let next = $state('');
  let passwordMsg = $state<{ kind: 'ok' | 'error'; text: string } | null>(null);
  let busy = $state(false);

  onMount(async () => {
    try {
      settings = await api.settings();
      days = settings.retention_days;
    } catch {
      /* the page still shows the language and password sections */
    }
  });

  async function saveRetention(e: SubmitEvent) {
    e.preventDefault();
    saved = false;
    try {
      settings = await api.saveSettings(Number(days));
      saved = true;
    } catch {
      saved = false;
    }
  }

  async function changePassword(e: SubmitEvent) {
    e.preventDefault();
    passwordMsg = null;
    busy = true;
    try {
      await api.changePassword(current, next);
      current = next = '';
      passwordMsg = { kind: 'ok', text: t('settings.changed') };
    } catch (err) {
      const text =
        err instanceof ApiError
          ? err.status === 403
            ? t('settings.wrongCurrent')
            : err.status === 429
              ? t('auth.tooMany')
              : err.status === 400
                ? err.message
                : t('common.errorGeneric')
          : t('common.errorGeneric');
      passwordMsg = { kind: 'error', text };
    } finally {
      busy = false;
    }
  }
</script>

<h1>{t('settings.title')}</h1>

<div class="grid">
  <section class="card">
    <h2>{t('settings.retention')}</h2>
    <form onsubmit={saveRetention}>
      <label for="days">{t('settings.days')}</label>
      <input id="days" type="number" min="0" max="3650" bind:value={days} />
      <p class="muted help">{t('settings.retentionHelp')}</p>
      {#if saved}<p class="ok" role="status">{t('settings.saved')}</p>{/if}
      <button class="primary">{t('common.save')}</button>
    </form>
  </section>

  <section class="card">
    <h2>{t('settings.passwordTitle')}</h2>
    <form onsubmit={changePassword}>
      <label for="cur">{t('settings.current')}</label>
      <input id="cur" type="password" bind:value={current} autocomplete="current-password" required />
      <label for="new">{t('settings.newPassword')}</label>
      <input id="new" type="password" bind:value={next} autocomplete="new-password" required />
      <p class="muted help">{t('auth.passwordHint')}</p>
      {#if passwordMsg}<p class={passwordMsg.kind} role="alert">{passwordMsg.text}</p>{/if}
      <button class="primary" disabled={busy}>{t('settings.change')}</button>
    </form>
  </section>

  <AlertsSettings />

  <section class="card">
    <h2>{t('settings.language')}</h2>
    <select value={i18n.lang} onchange={(e) => setLang(e.currentTarget.value as Lang)} aria-label={t('settings.language')}>
      <option value="en">English</option>
      <option value="pt-BR">Português (Brasil)</option>
    </select>
  </section>

  <section class="card">
    <h2>{t('settings.about')}</h2>
    <p>{t('settings.version')}: <span class="mono">{app.status.version}</span></p>
    {#if settings}<p>{t('settings.scanInterval')}: <span class="mono">{settings.scan_interval}</span></p>{/if}
    <p class="muted">{t('settings.privacy')}</p>
  </section>
</div>

<style>
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(18rem, 1fr));
    gap: 1rem;
  }
  .help {
    font-size: 0.85rem;
    margin: 0.4rem 0 0.8rem;
  }
  form button {
    margin-top: 0.4rem;
  }
</style>
