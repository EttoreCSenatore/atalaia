<script lang="ts">
  import { api, ApiError } from '../lib/api';
  import { t, i18n, setLang } from '../lib/i18n.svelte';
  import { signedIn } from '../lib/app.svelte';

  let { mode }: { mode: 'setup' | 'login' } = $props();

  let code = $state('');
  let password = $state('');
  let confirm = $state('');
  let error = $state('');
  let busy = $state(false);

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = '';
    if (mode === 'setup' && password !== confirm) {
      error = t('auth.mismatch');
      return;
    }
    busy = true;
    try {
      if (mode === 'setup') await api.setup(code, password);
      else await api.login(password);
      password = confirm = code = '';
      await signedIn();
    } catch (err) {
      error = describe(err);
    } finally {
      busy = false;
    }
  }

  function describe(err: unknown): string {
    if (!(err instanceof ApiError)) return t('common.errorGeneric');
    if (err.status === 0) return t('common.errorNetwork');
    if (err.status === 429) return t('auth.tooMany');
    if (mode === 'login' && err.status === 401) return t('auth.wrongPassword');
    if (mode === 'setup' && err.status === 403) return t('auth.wrongCode');
    if (err.status === 400) return err.message; // e.g. the password length rule, already explicit
    return t('common.errorGeneric');
  }
</script>

<div class="wrap">
  <div class="card">
    <div class="brand">
      <img src="/favicon.svg" alt="" width="40" height="40" />
      <div>
        <strong>{t('app.name')}</strong>
        <div class="muted">{t('app.tagline')}</div>
      </div>
    </div>
    <h1>{mode === 'setup' ? t('auth.setupTitle') : t('auth.loginTitle')}</h1>
    <p class="muted">{mode === 'setup' ? t('auth.setupIntro') : t('auth.loginIntro')}</p>
    <form onsubmit={submit}>
      {#if mode === 'setup'}
        <label for="code">{t('auth.code')}</label>
        <input id="code" class="mono" bind:value={code} autocomplete="off" autocapitalize="characters" spellcheck="false" required />
      {/if}
      <label for="password">{t('auth.password')}</label>
      <input id="password" type="password" bind:value={password} autocomplete={mode === 'setup' ? 'new-password' : 'current-password'} required />
      {#if mode === 'setup'}
        <p class="muted hint">{t('auth.passwordHint')}</p>
        <label for="confirm">{t('auth.confirm')}</label>
        <input id="confirm" type="password" bind:value={confirm} autocomplete="new-password" required />
      {/if}
      {#if error}<p class="error" role="alert">{error}</p>{/if}
      <button class="primary" disabled={busy}>{mode === 'setup' ? t('auth.create') : t('auth.login')}</button>
    </form>
    <p class="lang">
      <button type="button" class="link" onclick={() => setLang(i18n.lang === 'en' ? 'pt-BR' : 'en')}>
        {i18n.lang === 'en' ? 'Português' : 'English'}
      </button>
    </p>
  </div>
</div>

<style>
  .wrap {
    max-width: 26rem;
    margin: 8vh auto;
    padding: 0 1rem;
  }
  .brand {
    display: flex;
    align-items: center;
    gap: 0.7rem;
    margin-bottom: 1rem;
  }
  .hint {
    font-size: 0.85rem;
    margin: 0.3rem 0 0;
  }
  form button.primary {
    margin-top: 1rem;
    width: 100%;
  }
  .lang {
    text-align: center;
    margin: 1rem 0 0;
  }
  .link {
    border: 0;
    background: none;
    color: var(--primary);
    text-decoration: underline;
    padding: 0.2rem;
  }
</style>
