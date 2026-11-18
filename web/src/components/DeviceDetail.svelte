<script lang="ts">
  import { api, type PortsResult } from '../lib/api';
  import { toolError } from '../lib/errors';
  import { app, loadData } from '../lib/app.svelte';
  import { t, i18n } from '../lib/i18n.svelte';
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

  $effect(() => {
    if (device && portsFor !== device.mac) {
      portsFor = device.mac;
      ports = null;
      confirmScan = false;
      toolMsg = null;
      api.ports(device.mac).then((p) => (ports = p), () => (ports = null));
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
      <dt>{t('dashboard.cols.mac')}</dt><dd class="mono">{device.mac}</dd>
      <dt>{t('dashboard.cols.vendor')}</dt><dd>{device.randomized_mac ? t('common.randomMac') : device.vendor || t('common.unknown')}</dd>
      <dt>{t('detail.firstSeen')}</dt><dd>{when(device.first_seen)}</dd>
      <dt>{t('detail.lastSeen')}</dt><dd>{when(device.last_seen)} <span class="muted">({rel(device.last_seen)})</span></dd>
    </dl>
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
