<script lang="ts">
  import type { Snippet } from 'svelte';
  import { route } from '../lib/router.svelte';
  import { t } from '../lib/i18n.svelte';
  import { signOut } from '../lib/app.svelte';

  let { children }: { children: Snippet } = $props();

  const links = [
    { href: '/', key: 'nav.devices' },
    { href: '/history', key: 'nav.history' },
    { href: '/settings', key: 'nav.settings' },
  ];
  const active = (href: string) => (href === '/' ? route.path === '/' || route.path.startsWith('/devices') : route.path.startsWith(href));
</script>

<header>
  <div class="bar">
    <a class="brand" href="/" data-link>
      <img src="/favicon.svg" alt="" width="28" height="28" />
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
<main>{@render children()}</main>

<style>
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
