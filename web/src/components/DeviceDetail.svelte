<script lang="ts">
  import { app } from '../lib/app.svelte';
  import { t, i18n } from '../lib/i18n.svelte';
  import { dateTime, relativeTime } from '../lib/format';
  import StateBadge from './StateBadge.svelte';

  let { mac }: { mac: string } = $props();

  const device = $derived(app.devices.find((d) => d.mac === mac));
  const events = $derived(app.events.filter((e) => e.mac === mac).slice(0, 50));
  const when = (iso: string) => dateTime(iso, i18n.lang);
  const rel = (iso: string) => relativeTime(iso, i18n.lang, new Date(), t('common.never'), t('time.justNow'));
</script>

<p><a href="/" data-link>← {t('detail.back')}</a></p>

{#if !device}
  <p class="muted">{t('detail.notFound')}</p>
{:else}
  <div class="title">
    <h1>{device.name || device.vendor || device.mac}</h1>
    <StateBadge online={device.online} />
  </div>

  <div class="card">
    <dl>
      <dt>{t('dashboard.cols.ip')}</dt><dd class="mono">{device.ip || t('common.none')}</dd>
      <dt>{t('dashboard.cols.mac')}</dt><dd class="mono">{device.mac}</dd>
      <dt>{t('dashboard.cols.vendor')}</dt><dd>{device.randomized_mac ? t('common.randomMac') : device.vendor || t('common.unknown')}</dd>
      <dt>{t('detail.firstSeen')}</dt><dd>{when(device.first_seen)}</dd>
      <dt>{t('detail.lastSeen')}</dt><dd>{when(device.last_seen)} <span class="muted">({rel(device.last_seen)})</span></dd>
    </dl>
    {#if device.randomized_mac}<p class="note">{t('detail.randomNote')}</p>{/if}
  </div>

  <h2>{t('detail.history')}</h2>
  <div class="card table-wrap">
    {#if events.length === 0}
      <p class="muted">{t('detail.noHistory')}</p>
    {:else}
      <table>
        <tbody>
          {#each events as e, i (i)}
            <tr><td class="muted">{when(e.at)}</td><td>{t(`history.kinds.${e.kind}`)}</td><td class="mono">{e.detail}</td></tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>
{/if}

<style>
  .title {
    display: flex;
    align-items: center;
    gap: 0.8rem;
    flex-wrap: wrap;
    margin-bottom: 1rem;
  }
  .title h1 {
    margin: 0;
  }
  dl {
    display: grid;
    grid-template-columns: max-content 1fr;
    gap: 0.5rem 1.5rem;
    margin: 0;
  }
  dt {
    color: var(--muted);
  }
  dd {
    margin: 0;
  }
  .note {
    background: var(--warn-bg);
    color: var(--warn);
    border-radius: 8px;
    padding: 0.6rem 0.8rem;
    margin: 1rem 0 0;
  }
  h2 {
    margin-top: 1.5rem;
  }
</style>
