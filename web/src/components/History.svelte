<script lang="ts">
  import { app } from '../lib/app.svelte';
  import { t, i18n } from '../lib/i18n.svelte';
  import { dateTime, deviceLabel } from '../lib/format';

  const when = (iso: string) => dateTime(iso, i18n.lang);
</script>

<h1>{t('history.title')}</h1>
<div class="card table-wrap">
  {#if app.events.length === 0}
    <p class="muted">{t('history.empty')}</p>
  {:else}
    <table>
      <thead>
        <tr><th>{t('history.when')}</th><th>{t('history.event')}</th><th>{t('history.device')}</th><th>{t('history.detail')}</th></tr>
      </thead>
      <tbody>
        {#each app.events as e, i (i)}
          <tr>
            <td class="muted">{when(e.at)}</td>
            <td><span class="kind {e.kind}">{t(`history.kinds.${e.kind}`)}</span></td>
            <td><a href={`/devices/${e.mac}`} data-link>{deviceLabel(e)}</a></td>
            <td class="mono">{e.detail}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

<style>
  .kind {
    padding: 0.1rem 0.55rem;
    border-radius: 999px;
    font-size: 0.85rem;
    background: var(--off-bg);
    color: var(--off);
  }
  .kind.joined {
    background: var(--warn-bg);
    color: var(--warn);
  }
  .kind.online {
    background: var(--ok-bg);
    color: var(--ok);
  }
  .kind.ip_changed {
    background: var(--warn-bg);
    color: var(--warn);
  }
</style>
