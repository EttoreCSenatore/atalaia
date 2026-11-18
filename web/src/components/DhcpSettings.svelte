<script lang="ts">
  import { onMount } from 'svelte';
  import { api, type Dhcp } from '../lib/api';
  import { app } from '../lib/app.svelte';
  import { toolError } from '../lib/errors';
  import { dateTime } from '../lib/format';
  import { t, i18n } from '../lib/i18n.svelte';

  let dhcp = $state<Dhcp | null>(null);
  let confirm = $state(false);
  let busy = $state(false);
  let msg = $state<{ kind: 'ok' | 'error'; text: string } | null>(null);
  const observeOnly = $derived(app.status.active_tools === false);

  onMount(async () => {
    try {
      dhcp = await api.dhcp();
    } catch {
      /* the card stays empty */
    }
  });

  async function run(fn: () => Promise<Dhcp>, ok?: string) {
    msg = null;
    busy = true;
    try {
      dhcp = await fn();
      if (ok) msg = { kind: 'ok', text: ok };
    } catch (err) {
      msg = { kind: 'error', text: toolError(err) };
    } finally {
      busy = false;
    }
  }

  const check = () => run(() => api.checkDhcp(), t('dhcp.checked'));
  const toggle = (e: Event) => run(() => api.setDhcp((e.currentTarget as HTMLInputElement).checked));
  const reset = () => run(() => api.resetDhcp(), t('dhcp.forgotten'));
</script>

<section class="card wide">
  <h2>{t('dhcp.title')}</h2>
  <p class="muted">{t('dhcp.intro')}</p>
  {#if observeOnly}
    <p class="muted">{t('detail.observeOnly')}</p>
  {:else if dhcp}
    <label class="check"><input type="checkbox" checked={dhcp.enabled} disabled={busy} onchange={toggle} /> {t('dhcp.scheduled')}</label>
    <p class="muted help">{t('dhcp.privilege')}</p>

    {#if dhcp.last?.error}<p class="error" role="alert">{dhcp.last.error}</p>{/if}
    {#if dhcp.last && dhcp.last.new.length > 0}
      <p class="error" role="alert">{t('dhcp.newServer')}: {dhcp.last.new.join(', ')}</p>
    {/if}
    {#if dhcp.last?.baseline}<p class="ok" role="status">{t('dhcp.baseline')}</p>{/if}

    {#if dhcp.servers.length === 0}
      <p class="muted">{t('dhcp.none')}</p>
    {:else}
      <table>
        <thead><tr><th>{t('dhcp.server')}</th><th>{t('dhcp.gateway')}</th><th>DNS</th><th>{t('dhcp.lastSeen')}</th></tr></thead>
        <tbody>
          {#each dhcp.servers as s (s.ip)}
            <tr><td class="mono">{s.ip}</td><td class="mono">{s.router || '-'}</td><td class="mono">{s.dns.join(', ') || '-'}</td><td class="muted">{dateTime(s.last_seen, i18n.lang)}</td></tr>
          {/each}
        </tbody>
      </table>
    {/if}

    <label class="check"><input type="checkbox" bind:checked={confirm} /> {t('dhcp.confirm')}</label>
    <div class="row">
      <button type="button" disabled={!confirm || busy} onclick={check}>{t('dhcp.checkNow')}</button>
      <button type="button" class="danger" disabled={!confirm || busy || dhcp.servers.length === 0} onclick={reset}>{t('dhcp.reset')}</button>
    </div>
    {#if msg}<p class={msg.kind} role="alert">{msg.text}</p>{/if}
  {/if}
</section>

<style>
  .wide {
    grid-column: 1 / -1;
  }
  .check {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin: 0.6rem 0;
  }
  .check input {
    width: auto;
  }
  .row {
    display: flex;
    gap: 0.6rem;
    flex-wrap: wrap;
    margin-top: 0.4rem;
  }
  .help {
    font-size: 0.85rem;
  }
</style>
