<script lang="ts">
  import { t } from '../lib/i18n.svelte';

  let { online }: { online: boolean } = $props();
</script>

<span class="badge" class:on={online}>
  <span class="dot" aria-hidden="true"></span>{online ? t('common.online') : t('common.offline')}
</span>

<style>
  .badge {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    border-radius: 999px;
    padding: 0.1rem 0.6rem 0.1rem 0.5rem;
    font-size: 0.85rem;
    background: var(--off-bg);
    color: var(--off);
  }
  .badge.on {
    background: var(--ok-bg);
    color: var(--ok);
  }
  .dot {
    width: 0.55rem;
    height: 0.55rem;
    border-radius: 50%;
    background: currentColor;
  }
</style>
