<script lang="ts">
  import { onMount } from 'svelte';
  import { t, i18n } from '../lib/i18n.svelte';
  import { app } from '../lib/app.svelte';
  import { news, loadNews, markSeen } from '../lib/news.svelte';
  import { compareVersion, codeParts } from '../lib/format';

  // Versions newer than what this browser saw before opening the page stay marked as new while it is open.
  let before = $state('');

  onMount(async () => {
    if (!news.loaded) await loadNews();
    before = news.seen;
    markSeen();
  });

  const running = (v: string) => app.status.version.replace(/^v/, '') === v;
  const isNew = (v: string) => !!before && compareVersion(v, before) > 0;
  const date = (d: string) => new Date(d + 'T12:00:00Z').toLocaleDateString(i18n.lang, { year: 'numeric', month: 'long', day: 'numeric', timeZone: 'UTC' });
</script>

<h1>{t('news.title')}</h1>
<p class="muted">{t('news.intro')}</p>

{#if news.failed}
  <p class="error">{t('news.error')}</p>
{:else if !news.loaded}
  <p class="muted">{t('common.loading')}</p>
{:else}
  {#each news.releases as r (r.version)}
    {@const n = r[i18n.lang] ?? r.en}
    <section class="card release" aria-labelledby={`v-${r.version}`}>
      <h2 id={`v-${r.version}`}>
        <span class="mono">v{r.version}</span> · {n.title}
        {#if isNew(r.version)}<span class="badge new">{t('news.new')}</span>{/if}
        {#if running(r.version)}<span class="badge">{t('news.running')}</span>{/if}
      </h2>
      <p class="muted"><time datetime={r.date}>{date(r.date)}</time></p>
      <ul>
        {#each n.items as item, i (i)}
          <li>
            {#each codeParts(item) as part, j (j)}{#if j % 2 === 1}<code>{part}</code>{:else}{part}{/if}{/each}
          </li>
        {/each}
      </ul>
    </section>
  {/each}
{/if}

<style>
  .release {
    margin-bottom: 1rem;
  }
  h2 {
    font-size: 1.1rem;
    margin: 0 0 0.2rem;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.4rem;
  }
  ul {
    margin: 0.5rem 0 0;
    padding-left: 1.2rem;
  }
  li {
    margin: 0.25rem 0;
  }
  .badge {
    font-size: 0.75rem;
    font-weight: 700;
    padding: 0.1rem 0.5rem;
    border-radius: 999px;
    background: var(--bg);
    border: 1px solid var(--line);
  }
  .badge.new {
    background: var(--warn-bg);
    border-color: var(--warn);
    color: var(--warn);
  }
</style>
