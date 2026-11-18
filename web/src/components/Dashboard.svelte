<script lang="ts">
  import { app, startScan } from '../lib/app.svelte';
  import { t, i18n } from '../lib/i18n.svelte';
  import { compareIp, deviceLabel, relativeTime } from '../lib/format';
  import StateBadge from './StateBadge.svelte';

  let query = $state('');
  let filter = $state<'all' | 'online' | 'offline'>('all');

  const online = $derived(app.devices.filter((d) => d.online).length);
  const newToday = $derived(app.devices.filter((d) => Date.now() - new Date(d.first_seen).getTime() < 86_400_000).length);
  const shown = $derived(
    app.devices
      .filter((d) => (filter === 'all' ? true : filter === 'online' ? d.online : !d.online))
      .filter((d) => {
        const q = query.trim().toLowerCase();
        return !q || [d.name, d.vendor, d.ip, d.mac].some((v) => v.toLowerCase().includes(q));
      })
      .toSorted((a, b) => Number(b.online) - Number(a.online) || compareIp(a.ip, b.ip)),
  );
  const rel = (iso: string) => relativeTime(iso, i18n.lang, new Date(), t('common.never'), t('time.justNow'));
</script>

<div class="head">
  <h1>{t('dashboard.title')}</h1>
  <div class="actions">
    <span class="muted live" class:off={!app.live}>{app.live ? t('dashboard.live') : t('dashboard.liveOff')}</span>
    <button class="primary" onclick={startScan} disabled={app.scanning}>{app.scanning ? t('dashboard.scanning') : t('dashboard.scanNow')}</button>
  </div>
</div>

<p class="muted last">
  {t('dashboard.lastScan')}: {rel(app.lastScan)}
</p>
{#if app.scanError}<p class="error" role="alert">{t('dashboard.scanFailed')}: {app.scanError}</p>{/if}

<div class="stats">
  <div class="card"><div class="num">{app.devices.length}</div><div class="muted">{t('dashboard.total')}</div></div>
  <div class="card"><div class="num ok-text">{online}</div><div class="muted">{t('dashboard.onlineCount')}</div></div>
  <div class="card"><div class="num">{app.devices.length - online}</div><div class="muted">{t('dashboard.offlineCount')}</div></div>
  <div class="card"><div class="num warn-text">{newToday}</div><div class="muted">{t('dashboard.newToday')}</div></div>
</div>

<div class="tools">
  <label class="sr-only" for="q">{t('dashboard.search')}</label>
  <input id="q" type="search" placeholder={t('dashboard.search')} bind:value={query} />
  <div class="seg" role="group">
    <button class:sel={filter === 'all'} onclick={() => (filter = 'all')}>{t('dashboard.all')}</button>
    <button class:sel={filter === 'online'} onclick={() => (filter = 'online')}>{t('common.online')}</button>
    <button class:sel={filter === 'offline'} onclick={() => (filter = 'offline')}>{t('common.offline')}</button>
  </div>
</div>

<div class="card table-wrap">
  {#if app.devices.length === 0}
    <p class="muted">{t('dashboard.empty')}</p>
  {:else if shown.length === 0}
    <p class="muted">{t('dashboard.noMatch')}</p>
  {:else}
    <table>
      <thead>
        <tr>
          <th>{t('dashboard.cols.state')}</th><th>{t('dashboard.cols.ip')}</th><th>{t('dashboard.cols.name')}</th>
          <th>{t('dashboard.cols.vendor')}</th><th>{t('dashboard.cols.mac')}</th><th>{t('dashboard.cols.lastSeen')}</th>
        </tr>
      </thead>
      <tbody>
        {#each shown as d (d.mac)}
          <tr>
            <td><StateBadge online={d.online} /></td>
            <td class="mono">{d.ip || t('common.none')}</td>
            <td><a href={`/devices/${d.mac}`} data-link>{d.name || (d.randomized_mac ? t('common.randomMac') : d.vendor) || d.mac}</a></td>
            <td>{d.randomized_mac ? t('common.randomMac') : d.vendor || t('common.none')}</td>
            <td class="mono">{d.mac}</td>
            <td class="muted">{rel(d.last_seen)}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

<style>
  .head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    flex-wrap: wrap;
  }
  .head h1 {
    margin: 0;
  }
  .actions {
    display: flex;
    align-items: center;
    gap: 0.8rem;
  }
  .live {
    font-size: 0.85rem;
  }
  .live.off {
    color: var(--warn);
  }
  .last {
    margin: 0.3rem 0 1rem;
  }
  .stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(9rem, 1fr));
    gap: 0.8rem;
    margin-bottom: 1rem;
  }
  .num {
    font-size: 1.8rem;
    font-weight: 800;
  }
  .ok-text {
    color: var(--ok);
  }
  .warn-text {
    color: var(--warn);
  }
  .tools {
    display: flex;
    gap: 0.8rem;
    margin-bottom: 0.8rem;
    flex-wrap: wrap;
  }
  .tools input {
    flex: 1;
    min-width: 12rem;
  }
  .seg {
    display: flex;
  }
  .seg button {
    border-radius: 0;
  }
  .seg button:first-child {
    border-radius: 8px 0 0 8px;
  }
  .seg button:last-child {
    border-radius: 0 8px 8px 0;
  }
  .seg .sel {
    background: var(--primary);
    color: var(--primary-text);
    border-color: var(--primary);
  }
</style>
