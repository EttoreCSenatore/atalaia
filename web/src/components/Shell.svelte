<script lang="ts">
  import type { Snippet } from 'svelte';
  import { route } from '../lib/router.svelte';
  import { t } from '../lib/i18n.svelte';
  import { onMount } from 'svelte';
  import { app, signOut } from '../lib/app.svelte';
  import { news, loadNews, hasUnseen, latest, markSeen } from '../lib/news.svelte';
  import { formatBytes } from '../lib/format';
  import { i18n } from '../lib/i18n.svelte';

  onMount(() => {
    if (!news.loaded) loadNews();
  });

  let { children }: { children: Snippet } = $props();

  const links = [
    { href: '/', key: 'nav.devices' },
    { href: '/history', key: 'nav.history' },
    { href: '/tools', key: 'nav.tools' },
    { href: '/settings', key: 'nav.settings' },
  ];
  const active = (href: string) => (href === '/' ? route.path === '/' || route.path.startsWith('/devices') : route.path.startsWith(href));
</script>

<header>
  <div class="bar">
    <a class="brand" href="/" data-link>
      <img src={`${import.meta.env.BASE_URL}favicon.svg`} alt="" width="28" height="28" />
      <span>{t('app.name')}</span>
    </a>
    <nav aria-label="Main">
      {#each links as l (l.href)}
        <a href={l.href} data-link aria-current={active(l.href) ? 'page' : undefined}>{t(l.key)}</a>
      {/each}
    </nav>
    <button class="logout" onclick={signOut}>{t('nav.logout')}</button>
  </div>
</header>
<main>
  {#if app.status.disk?.low}
    <div class="notice card danger" role="alert">
      <span><strong>{t('disk.title')}</strong> {t('disk.body')} ({formatBytes(app.status.disk.free_bytes, i18n.lang)} {t('disk.free')})</span>
    </div>
  {/if}
  {#if app.status.backup?.enabled && app.status.backup.late}
    <div class="notice card warn" role="status">
      <span><strong>{t('backup.lateTitle')}</strong> {t('backup.lateBody')}</span>
      <a href="/settings#backups" data-link>{t('backup.see')}</a>
    </div>
  {/if}
  {#if hasUnseen() && route.path !== '/whats-new'}
    <div class="notice card" role="status">
      <span>{t('news.updated')} <strong class="mono">v{latest()}</strong>.</span>
      <a href="/whats-new" data-link>{t('news.see')}</a>
      <button class="link" onclick={markSeen}>{t('news.dismiss')}</button>
    </div>
  {/if}
  {@render children()}
</main>

<footer class="source">
  <span class="mono">{app.status.version}</span>
  · <a href="/whats-new" data-link>{t('news.title')}</a>{#if hasUnseen()}<span class="dot" aria-label={t('news.new')}></span>{/if}
  ·
  <a href={`https://github.com/EttoreCSenatore/atalaia/tree/${app.status.version.startsWith('v') ? app.status.version : 'main'}`} target="_blank" rel="noopener noreferrer">{t('app.source')}</a>
  · AGPL-3.0-or-later
</footer>

<style>
  .notice {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.4rem 1rem;
    margin-bottom: 1rem;
    border-left: 4px solid var(--accent);
  }
  .notice.danger {
    border-left-color: var(--danger);
    background: var(--danger-bg);
  }
  .notice.warn {
    border-left-color: var(--warn);
    background: var(--warn-bg);
  }
  .notice .link {
    margin-left: auto;
  }
  .dot {
    display: inline-block;
    width: 0.5rem;
    height: 0.5rem;
    margin-left: 0.25rem;
    border-radius: 50%;
    background: var(--accent);
    vertical-align: super;
  }
  .source {
    max-width: 72rem;
    margin: 2rem auto 1rem;
    padding: 0 1rem;
    font-size: 0.85rem;
    color: var(--muted);
    text-align: center;
  }
  header {
    background: var(--surface);
    border-bottom: 3px solid var(--accent);
  }
  .bar {
    max-width: 68rem;
    margin: 0 auto;
    padding: 0.6rem 1rem;
    display: flex;
    align-items: center;
    gap: 1.2rem;
    flex-wrap: wrap;
  }
  .brand {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-weight: 800;
    font-size: 1.2rem;
    text-decoration: none;
    color: var(--primary);
  }
  nav {
    display: flex;
    gap: 0.3rem;
    flex: 1;
    flex-wrap: wrap;
  }
  nav a {
    text-decoration: none;
    color: var(--text);
    padding: 0.35rem 0.7rem;
    border-radius: 8px;
  }
  nav a[aria-current='page'] {
    background: var(--bg);
    font-weight: 700;
    box-shadow: inset 0 -2px 0 var(--primary);
  }
  main {
    max-width: 68rem;
    margin: 0 auto;
    padding: 1.2rem 1rem 3rem;
  }
</style>
