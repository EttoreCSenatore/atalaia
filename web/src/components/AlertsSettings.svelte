<script lang="ts">
  import { onMount } from 'svelte';
  import { api, ApiError, type Alerts, type Channel, type ChannelType } from '../lib/api';
  import { t } from '../lib/i18n.svelte';

  let data = $state<Alerts | null>(null);
  let adding = $state<ChannelType>('ntfy');
  let msg = $state<{ kind: 'ok' | 'error'; text: string } | null>(null);
  let results = $state<Record<string, { kind: 'ok' | 'error'; text: string }>>({});
  let busy = $state(false);

  const types: ChannelType[] = ['ntfy', 'telegram', 'webhook', 'email'];

  onMount(async () => {
    try {
      data = await api.alerts();
    } catch {
      msg = { kind: 'error', text: t('common.errorGeneric') };
    }
  });

  function add() {
    if (!data) return;
    data.channels.push({ id: '', type: adding, name: t(`alerts.types.${adding}`), enabled: true });
  }

  function remove(i: number) {
    data?.channels.splice(i, 1);
  }

  const errText = (e: unknown) => (e instanceof ApiError && e.status === 400 ? e.message : t('common.errorGeneric'));

  async function save(e: SubmitEvent) {
    e.preventDefault();
    if (!data) return;
    msg = null;
    busy = true;
    try {
      data = await api.saveAlerts(data);
      msg = { kind: 'ok', text: t('settings.saved') };
    } catch (err) {
      msg = { kind: 'error', text: errText(err) };
    } finally {
      busy = false;
    }
  }

  async function test(c: Channel) {
    if (!c.id) {
      results[c.id] = { kind: 'error', text: t('alerts.saveFirst') };
      return;
    }
    results[c.id] = { kind: 'ok', text: '…' };
    try {
      await api.testAlert(c.id);
      results[c.id] = { kind: 'ok', text: t('alerts.testOk') };
    } catch (err) {
      results[c.id] = { kind: 'error', text: err instanceof ApiError && err.status === 502 ? err.message : t('common.errorGeneric') };
    }
  }
</script>

<section class="card wide">
  <h2>{t('alerts.title')}</h2>
  <p class="muted">{t('alerts.intro')}</p>
  {#if data}
    <form onsubmit={save}>
      <label class="check"><input type="checkbox" bind:checked={data.rules.new_device} /> {t('alerts.ruleNew')}</label>
      <label class="check"><input type="checkbox" bind:checked={data.rules.known_offline} /> {t('alerts.ruleOffline')}</label>
      <label class="check"><input type="checkbox" bind:checked={data.rules.arp_anomaly} /> {t('alerts.ruleArp')}</label>
      <label class="check"><input type="checkbox" bind:checked={data.rules.rogue_dhcp} /> {t('alerts.ruleDhcp')}</label>
      <label class="check"><input type="checkbox" bind:checked={data.rules.new_ports} /> {t('alerts.rulePorts')}</label>
      <p class="muted help">{t('alerts.firstScan')}</p>

      {#each data.channels as c, i (c.id || i)}
        <fieldset>
          <legend>{t(`alerts.types.${c.type}`)}</legend>
          <label for={`n${i}`}>{t('alerts.name')}</label>
          <input id={`n${i}`} bind:value={c.name} maxlength="64" />
          {#if c.type === 'ntfy' || c.type === 'webhook'}
            <label for={`u${i}`}>{c.type === 'ntfy' ? t('alerts.ntfyUrl') : t('alerts.webhookUrl')}</label>
            <input id={`u${i}`} type="url" bind:value={c.url} placeholder={c.type === 'ntfy' ? 'https://ntfy.sh/my-secret-topic' : 'http://homeassistant.local:8123/api/webhook/…'} required />
          {/if}
          {#if c.type === 'ntfy'}
            <label for={`t${i}`}>{t('alerts.ntfyToken')}</label>
            <input id={`t${i}`} type="password" bind:value={c.token} autocomplete="off" placeholder={c.has_token ? t('alerts.secretKept') : ''} />
          {/if}
          {#if c.type === 'telegram'}
            <label for={`t${i}`}>{t('alerts.botToken')}</label>
            <input id={`t${i}`} type="password" bind:value={c.token} autocomplete="off" placeholder={c.has_token ? t('alerts.secretKept') : '123456:ABC…'} required={!c.has_token} />
            <label for={`c${i}`}>{t('alerts.chatId')}</label>
            <input id={`c${i}`} bind:value={c.chat_id} required />
          {/if}
          {#if c.type === 'email'}
            <label for={`h${i}`}>{t('alerts.smtpHost')}</label>
            <input id={`h${i}`} bind:value={c.host} placeholder="smtp.example.com:587" required />
            <label for={`us${i}`}>{t('alerts.smtpUser')}</label>
            <input id={`us${i}`} bind:value={c.username} autocomplete="off" />
            <label for={`pw${i}`}>{t('alerts.smtpPassword')}</label>
            <input id={`pw${i}`} type="password" bind:value={c.password} autocomplete="new-password" placeholder={c.has_password ? t('alerts.secretKept') : ''} />
            <label for={`f${i}`}>{t('alerts.from')}</label>
            <input id={`f${i}`} type="email" bind:value={c.from} required />
            <label for={`to${i}`}>{t('alerts.to')}</label>
            <input id={`to${i}`} type="email" bind:value={c.to} required />
          {/if}
          <label class="check"><input type="checkbox" bind:checked={c.enabled} /> {t('alerts.enabled')}</label>
          <div class="row">
            <button type="button" onclick={() => test(c)}>{t('alerts.test')}</button>
            <button type="button" class="danger" onclick={() => remove(i)}>{t('alerts.remove')}</button>
            {#if results[c.id]}<span class={results[c.id].kind} role="status">{results[c.id].text}</span>{/if}
          </div>
        </fieldset>
      {/each}

      {#if data.channels.length < 8}
        <div class="row">
          <select bind:value={adding} aria-label={t('alerts.addType')}>
            {#each types as ty (ty)}<option value={ty}>{t(`alerts.types.${ty}`)}</option>{/each}
          </select>
          <button type="button" onclick={add}>{t('alerts.add')}</button>
        </div>
      {/if}
      <p class="muted help">{t('alerts.secretsNote')}</p>
      {#if msg}<p class={msg.kind} role="alert">{msg.text}</p>{/if}
      <button class="primary" disabled={busy}>{t('common.save')}</button>
    </form>
  {:else if msg}
    <p class={msg.kind} role="alert">{msg.text}</p>
  {/if}
</section>

<style>
  .wide {
    grid-column: 1 / -1;
  }
  fieldset {
    border: 1px solid var(--line);
    border-radius: 8px;
    margin: 1rem 0;
    padding: 0.5rem 1rem 1rem;
  }
  .row {
    display: flex;
    gap: 0.6rem;
    align-items: center;
    flex-wrap: wrap;
    margin: 0.8rem 0 0;
  }
  .check {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin: 0.5rem 0;
  }
  .check input {
    width: auto;
  }
  .help {
    font-size: 0.85rem;
    margin: 0.4rem 0 0.8rem;
  }
</style>
