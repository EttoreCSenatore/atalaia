<script lang="ts">
  import { onMount } from 'svelte';
  import { app } from '../lib/app.svelte';
  import { api, type AuditEntry, type DnsResult, type PingResult, type SpeedResult, type TraceResult } from '../lib/api';
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

  let traceHost = $state('');
  let traceBusy = $state(false);
  let trace = $state<TraceResult | null>(null);
  let traceErr = $state('');

  let speedOk = $state(false);
  let speedBusy = $state(false);
  let speed = $state<SpeedResult | null>(null);
  let speedErr = $state('');

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

  async function runTrace(e: SubmitEvent) {
    e.preventDefault();
    traceBusy = true;
    traceErr = '';
    trace = null;
    try {
      trace = await api.traceroute(traceHost.trim());
    } catch (err) {
      traceErr = toolError(err);
    } finally {
      traceBusy = false;
      loadAudit();
    }
  }

  async function runSpeed() {
    speedBusy = true;
    speedErr = '';
    speed = null;
    try {
      speed = await api.speedtest();
      speedOk = false;
    } catch (err) {
      speedErr = toolError(err);
    } finally {
      speedBusy = false;
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
{#if app.status.active_tools === false}<p class="muted">{t('tools.observeOnly')}</p>{/if}

<div class="grid">
  {#if app.status.active_tools !== false}
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
    <h2>{t('tools.trace')}</h2>
    <form onsubmit={runTrace}>
      <label for="thost">{t('tools.host')}</label>
      <input id="thost" bind:value={traceHost} placeholder="1.1.1.1" required />
      {#if traceErr}<p class="error" role="alert">{traceErr}</p>{/if}
      <button class="primary" disabled={traceBusy}>{traceBusy ? t('tools.running') : t('tools.run')}</button>
    </form>
    {#if trace}
      <ol class="mono hops">
        {#each trace.hops as h (h.ttl)}
          <li>{h.ttl}. {h.addr || '*'}{h.addr ? ` · ${h.rtt_ms.toFixed(1)} ms` : ` · ${t('tools.timeout')}`}{h.reached ? ` · ${t('tools.arrived')}` : ''}</li>
        {/each}
      </ol>
    {/if}
  </section>

  <section class="card">
    <h2>{t('tools.speed')}</h2>
    <p class="muted">{t('tools.speedNotice')}</p>
    <label class="check"><input type="checkbox" bind:checked={speedOk} /> {t('tools.speedConfirm')}</label>
    <button type="button" class="primary" disabled={!speedOk || speedBusy} onclick={runSpeed}>{speedBusy ? t('tools.running') : t('tools.run')}</button>
    {#if speedErr}<p class="error" role="alert">{speedErr}</p>{/if}
    {#if speed}
      <p class="mono"><strong>{speed.mbps.toFixed(1)} Mbit/s</strong> · {(speed.bytes / 1e6).toFixed(1)} MB · {speed.seconds.toFixed(1)} s · {speed.latency_ms.toFixed(0)} ms</p>
      <p class="muted">{speed.server}</p>
    {/if}
  </section>
  {/if}

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
  .check {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin: 0.6rem 0;
  }
  .check input {
    width: auto;
  }
  .hops {
    padding-left: 1.6rem;
    margin: 0.5rem 0 0;
  }
  ul {
    padding-left: 1.2rem;
    margin: 0.5rem 0 0;
  }
</style>
