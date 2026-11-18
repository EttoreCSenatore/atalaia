<script lang="ts">
  import { onMount } from 'svelte';
  import { app, boot } from './lib/app.svelte';
  import { route, match } from './lib/router.svelte';
  import { t, i18n } from './lib/i18n.svelte';
  import Auth from './components/Auth.svelte';
  import Shell from './components/Shell.svelte';
  import Dashboard from './components/Dashboard.svelte';
  import DeviceDetail from './components/DeviceDetail.svelte';
  import History from './components/History.svelte';
  import Tools from './components/Tools.svelte';
  import Settings from './components/Settings.svelte';

  onMount(boot);

  const current = $derived(match(route.path));
  $effect(() => {
    document.documentElement.lang = i18n.lang;
    document.title = t('app.name');
  });
</script>

{#if app.phase === 'loading'}
  <p class="center muted">{t('common.loading')}</p>
{:else if app.phase === 'error'}
  <div class="center">
    <p class="error">{app.loadError === 'network' ? t('common.errorNetwork') : t('common.errorGeneric')}</p>
    <button onclick={boot}>{t('common.retry')}</button>
  </div>
{:else if app.status.setup_required}
  <Auth mode="setup" />
{:else if !app.status.authenticated}
  <Auth mode="login" />
{:else}
  <Shell>
    {#if current.name === 'devices'}
      <Dashboard />
    {:else if current.name === 'device'}
      <DeviceDetail mac={current.mac} />
    {:else if current.name === 'history'}
      <History />
    {:else if current.name === 'tools'}
      <Tools />
    {:else if current.name === 'settings'}
      <Settings />
    {:else}
      <p class="muted">404</p>
      <a href="/" data-link>{t('detail.back')}</a>
    {/if}
  </Shell>
{/if}

<style>
  .center {
    max-width: 28rem;
    margin: 20vh auto;
    padding: 0 1rem;
    text-align: center;
  }
</style>
