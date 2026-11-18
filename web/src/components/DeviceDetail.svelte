<script lang="ts">
  import { api, type Exposure, type PortsResult } from '../lib/api';
  import { toolError } from '../lib/errors';
  import { app, loadData } from '../lib/app.svelte';
  import { t, i18n } from '../lib/i18n.svelte';
  import { kindIcon } from '../lib/kinds';
  import { riskGroup } from '../lib/risk';
  import { dateTime, parseTags, relativeTime } from '../lib/format';
  import StateBadge from './StateBadge.svelte';

  let { mac }: { mac: string } = $props();

  const device = $derived(app.devices.find((d) => d.mac === mac));
  const events = $derived(app.events.filter((e) => e.mac === mac).slice(0, 50));
  let label = $state('');
  let tagText = $state('');
  let known = $state(false);
  let loadedFor = '';
  let saved = $state(false);
  let failed = $state('');
  let busy = $state(false);

  // fill the form once per device, never while the user is typing
  $effect(() => {
    if (device && loadedFor !== device.mac) {
      loadedFor = device.mac;
      label = device.label;
      tagText = device.tags.join(', ');
      known = device.known;
      saved = false;
      failed = '';
    }
  });

  async function save(e: SubmitEvent) {
    e.preventDefault();
    saved = false;
    failed = '';
    busy = true;
    try {
      await api.updateDevice(mac, label, parseTags(tagText), known);
      await loadData();
      saved = true;
    } catch (err) {
      failed = err instanceof Error && err.message && err.message !== 'network' ? err.message : t('common.errorGeneric');
    } finally {
      busy = false;
    }
  }

  let ports = $state<PortsResult | null>(null);
  let confirmScan = $state(false);
  let scanBusy = $state(false);
  let toolMsg = $state<{ kind: 'ok' | 'error'; text: string } | null>(null);
  let portsFor = '';
  let exposure = $state<Exposure | null>(null);
  let exposureBusy = $state(false);

  $effect(() => {
    if (device && portsFor !== device.mac) {
      portsFor = device.mac;
      ports = null;
      exposure = null;
      confirmScan = false;
      toolMsg = null;
      api.ports(device.mac).then((p) => (ports = p), () => (ports = null));
      api.exposure(device.mac).then((x) => (exposure = x), () => (exposure = null));
    }
  });

  async function wake() {
    toolMsg = null;
    try {
      await api.wake(mac);
      toolMsg = { kind: 'ok', text: t('detail.wakeSent') };
    } catch (err) {
      toolMsg = { kind: 'error', text: toolError(err) };
    }
  }

  async function scanPorts() {
    toolMsg = null;
    scanBusy = true;
    try {
      ports = await api.portScan(mac);
      confirmScan = false;
    } catch (err) {
      toolMsg = { kind: 'error', text: toolError(err) };
    } finally {
      scanBusy = false;
    }
  }

  async function checkExposure() {
    toolMsg = null;
    exposureBusy = true;
    try {
      exposure = await api.exposureScan(mac);
    } catch (err) {
      toolMsg = { kind: 'error', text: toolError(err) };
    } finally {
      exposureBusy = false;
    }
  }

  const when = (iso: string) => dateTime(iso, i18n.lang);
  const rel = (iso: string) => relativeTime(iso, i18n.lang, new Date(), t('common.never'), t('time.justNow'));
</script>

<p><a href="/" data-link>← {t('detail.back')}</a></p>

{#if !device}
  <p class="muted">{t('detail.notFound')}</p>
{:else}
  <div class="title">
    <h1>{device.label || device.name || device.vendor || device.mac}</h1>
    <StateBadge online={device.online} />
  </div>

  <div class="card">
    <dl>
      <dt>{t('dashboard.cols.ip')}</dt><dd class="mono">{device.ip || t('common.none')}</dd>
      <dt>{t('detail.type')}</dt><dd>{kindIcon(device.kind)} {device.kind ? t(`kinds.${device.kind}`) : t('kinds.unknown')}{#if device.model} · {device.model}{/if}</dd>
      <dt>{t('dashboard.cols.mac')}</dt><dd class="mono">{device.mac}</dd>
      <dt>{t('dashboard.cols.vendor')}</dt><dd>{device.randomized_mac ? t('common.randomMac') : device.vendor || t('common.unknown')}</dd>
      <dt>{t('detail.firstSeen')}</dt><dd>{when(device.first_seen)}</dd>
      <dt>{t('detail.lastSeen')}</dt><dd>{when(device.last_seen)} <span class="muted">({rel(device.last_seen)})</span></dd>
    </dl>
    {#if device.services.length > 0}<p class="muted">{t('detail.services')}: <span class="mono">{device.services.join(', ')}</span></p>{/if}
    {#if device.randomized_mac}<p class="note">{t('detail.randomNote')}</p>{/if}
  </div>

  <h2>{t('detail.edit')}</h2>
  <form class="card" onsubmit={save}>
    <label for="label">{t('detail.label')}</label>
    <input id="label" bind:value={label} maxlength="64" placeholder={device.name || ''} />
    <label for="tags">{t('detail.tags')}</label>
    <input id="tags" bind:value={tagText} placeholder="iot, media" />
    <p class="muted help">{t('detail.tagsHelp')}</p>
    <label class="check"><input type="checkbox" bind:checked={known} /> {t('detail.known')}</label>
    <p class="muted help">{t('detail.knownHelp')}</p>
    {#if saved}<p class="ok" role="status">{t('settings.saved')}</p>{/if}
    {#if failed}<p class="error" role="alert">{failed}</p>{/if}
    <button class="primary" disabled={busy}>{t('common.save')}</button>
  </form>

  {#if app.status.active_tools === false}
    <p class="muted">{t('detail.observeOnly')}</p>
  {:else}
  <h2>{t('detail.tools')}</h2>
  <div class="card">
    <div class="row">
      <button type="button" onclick={wake}>{t('detail.wake')}</button>
      <span class="muted">{t('detail.wakeHelp')}</span>
    </div>
    <hr />
    <label class="check"><input type="checkbox" bind:checked={confirmScan} /> {t('detail.confirmScan')}</label>
    <button type="button" disabled={!confirmScan || scanBusy} onclick={scanPorts}>{scanBusy ? t('tools.running') : t('detail.scanPorts')}</button>
    <p class="muted help">{t('detail.scanHelp')}</p>
    <button type="button" disabled={!confirmScan || exposureBusy || !ports} onclick={checkExposure}>{exposureBusy ? t('tools.running') : t('detail.checkExposure')}</button>
    <p class="muted help">{t('detail.exposureHelp')}</p>
    {#if toolMsg}<p class={toolMsg.kind} role="alert">{toolMsg.text}</p>{/if}
    {#if ports}
      <p class="muted">{t('detail.scannedAt')}: {when(ports.at)}</p>
      {#if ports.opened?.length}<p class="warn" role="status">{t('detail.opened')}: {ports.opened.join(', ')}</p>{/if}
      {#if ports.closed?.length}<p class="muted">{t('detail.closed')}: {ports.closed.join(', ')}</p>{/if}
      {#if ports.ports.length === 0}
        <p>{t('detail.noPorts')}</p>
      {:else}
        <table>
          <tbody>
            {#each ports.ports as p (p.port)}<tr><td class="mono">{p.port}/tcp</td><td>{p.service}</td></tr>{/each}
          </tbody>
        </table>
      {/if}
    {/if}
  </div>

  {/if}

  {#if exposure}
    <h2>{t('detail.exposure')}</h2>
    <div class="card">
      <p class="muted">{t('detail.scannedAt')}: {when(exposure.at)}</p>
      {#if exposure.risks.length === 0}
        <p class="ok">{t('detail.noRisks')}</p>
      {:else}
        <ul class="risks">
          {#each exposure.risks as r (r.port)}
            <li class={r.severity}>
              <strong>{t(`detail.severity.${r.severity}`)}</strong> · <span class="mono">{r.port}/tcp {r.service}</span>
              <br /><span class="muted">{t(`detail.risk.${riskGroup(r.service)}`)}</span>
            </li>
          {/each}
        </ul>
      {/if}
      {#if exposure.web.length > 0}
        <h3>{t('detail.web')}</h3>
        <table>
          <tbody>
            {#each exposure.web as w (w.port)}
              <tr>
                <td class="mono">{w.tls ? 'https' : 'http'}://:{w.port}</td>
                <td>{w.status}{w.auth ? ` · ${t('detail.login')}` : ''}</td>
                <td>{w.title || '-'}</td>
                <td class="muted">{w.server}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      {/if}
      {#if exposure.tls.length > 0}
        <h3>TLS</h3>
        <table>
          <tbody>
            {#each exposure.tls as x (x.port)}
              <tr>
                <td class="mono">:{x.port}</td>
                <td>{x.version}</td>
                <td class="muted">{x.subject}</td>
                <td>
                  {#if x.expired}<span class="bad">{t('detail.tlsExpired')}</span>{:else}{x.days_left} {t('detail.tlsDays')}{/if}
                  {#if x.self_signed} · {t('detail.tlsSelf')}{/if}
                  {#if x.legacy_tls} · <span class="bad">{t('detail.tlsLegacy')}</span>{/if}
                  {#if x.weak_key_bits} · <span class="bad">{t('detail.tlsWeak')} ({x.weak_key_bits})</span>{/if}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      {/if}
    </div>
  {/if}

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
  .help {
    font-size: 0.85rem;
    margin: 0.3rem 0 0.8rem;
  }
  .row {
    display: flex;
    align-items: center;
    gap: 0.8rem;
    flex-wrap: wrap;
  }
  hr {
    border: 0;
    border-top: 1px solid var(--line);
    margin: 1rem 0;
  }
  .risks {
    list-style: none;
    padding: 0;
    margin: 0.5rem 0;
  }
  .risks li {
    border-left: 4px solid var(--warn);
    padding: 0.3rem 0.8rem;
    margin: 0.5rem 0;
  }
  .risks li.critical,
  .risks li.high {
    border-color: var(--danger);
  }
  .bad {
    color: var(--danger);
    font-weight: 600;
  }
  h3 {
    margin: 1.2rem 0 0.4rem;
  }
  .warn {
    color: var(--warn);
    font-weight: 600;
  }
  .check {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
  .check input {
    width: auto;
  }
</style>
