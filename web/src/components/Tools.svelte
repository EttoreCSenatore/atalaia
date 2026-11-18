<script lang="ts">
  import { onMount } from 'svelte';
  import { api, type AuditEntry, type DnsResult, type PingResult } from '../lib/api';
  import { toolError } from '../lib/errors';
  import { dateTime } from '../lib/format';
  import { t, i18n } from '../lib/i18n.svelte';

  let host = $state('');
  let pingBusy = $state(false);
  let ping = $state<PingResult | null>(null);
  let pingErr = $state('');

  let name = $state('');
  let type = $state('A');
  let dnsBusy = $state(false);
  let dns = $state<DnsResult | null>(null);
  let dnsErr = $state('');

  let audit = $state<AuditEntry[]>([]);
  const types = ['A', 'AAAA', 'CNAME', 'MX', 'NS', 'TXT', 'PTR'];

  const loadAudit = async () => {
    try {
      audit = await api.audit();
    } catch {
      /* the log is optional on this page */
    }
  };
  onMount(loadAudit);

  async function runPing(e: SubmitEvent) {
    e.preventDefault();
    pingBusy = true;
    pingErr = '';
    ping = null;
    try {
      ping = await api.ping(host.trim());
    } catch (err) {
      pingErr = toolError(err);
    } finally {
      pingBusy = false;
      loadAudit();
    }
  }

  async function runDns(e: SubmitEvent) {
    e.preventDefault();
    dnsBusy = true;
    dnsErr = '';
    dns = null;
    try {
      dns = await api.dns(name.trim(), type);
    } catch (err) {
      dnsErr = toolError(err);
    } finally {
      dnsBusy = false;
    }
  }
</script>

<h1>{t('tools.title')}</h1>

<div class="grid">
  <section class="card">
    <h2>{t('tools.ping')}</h2>
    <form onsubmit={runPing}>
      <label for="host">{t('tools.host')}</label>
      <input id="host" bind:value={host} placeholder="192.168.1.1" required />
      {#if pingErr}<p class="error" role="alert">{pingErr}</p>{/if}
      <button class="primary" disabled={pingBusy}>{pingBusy ? t('tools.running') : t('tools.run')}</button>
    </form>
    {#if ping}
      <p class="mono">{ping.address}: {ping.sent - ping.lost}/{ping.sent} {t('tools.replies')}</p>
      <ul class="mono">
        {#each ping.replies as r (r.seq)}<li>#{r.seq}: {r.ok ? `${r.rtt_ms.toFixed(1)} ms` : t('tools.timeout')}</li>{/each}
      </ul>
    {/if}
  </section>

  <section class="card">
    <h2>{t('tools.dns')}</h2>
    <form onsubmit={runDns}>
      <label for="dname">{t('tools.name')}</label>
      <input id="dname" bind:value={name} placeholder="example.com" required />
      <label for="dtype">{t('tools.type')}</label>
      <select id="dtype" bind:value={type}>
        {#each types as ty (ty)}<option value={ty}>{ty}</option>{/each}
      </select>
      {#if dnsErr}<p class="error" role="alert">{dnsErr}</p>{/if}
      <button class="primary" disabled={dnsBusy}>{dnsBusy ? t('tools.running') : t('tools.run')}</button>
    </form>
    {#if dns}
      {#if dns.answers.length === 0}
        <p class="muted">{t('tools.noAnswer')}</p>
      {:else}
        <ul class="mono">{#each dns.answers as a, i (i)}<li>{a}</li>{/each}</ul>
      {/if}
    {/if}
  </section>
</div>

<h2>{t('tools.auditTitle')}</h2>
<div class="card table-wrap">
  <p class="muted">{t('tools.auditHelp')}</p>
  {#if audit.length === 0}
    <p class="muted">{t('tools.auditEmpty')}</p>
  {:else}
    <table>
      <tbody>
        {#each audit as a, i (i)}
          <tr><td class="muted">{dateTime(a.at, i18n.lang)}</td><td>{t(`tools.actions.${a.action}`)}</td><td class="mono">{a.target}</td><td class="mono">{a.detail}</td></tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

<style>
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(18rem, 1fr));
    gap: 1rem;
  }
  form button {
    margin-top: 0.8rem;
  }
  ul {
    padding-left: 1.2rem;
    margin: 0.5rem 0 0;
  }
</style>
