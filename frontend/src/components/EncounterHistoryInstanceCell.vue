<template>
  <SafeExternalUrlLink
    v-if="cell"
    :url="cell.href"
    :label="cell.text"
    :title="cell.title"
    :aria-label="t('encounterHistory.openInVrchat', { text: cell.text })"
    link-class="encounter-instance-link"
    test-id="encounter-instance-link"
  />
  <!-- Unparseable but non-empty id: raw mono text, no launch link (see formatVrcInstanceCell). -->
  <span v-else-if="trimmedId" class="mono">{{ trimmedId }}</span>
  <span v-else>{{ t("common.dash") }}</span>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import SafeExternalUrlLink from "./SafeExternalUrlLink.vue";
import { formatVrcInstanceCell } from "../utils/vrcInstanceDisplay";

const props = defineProps<{
  instanceId?: string;
}>();

const { t } = useI18n();

const trimmedId = computed(() => props.instanceId?.trim() ?? "");

const cell = computed(() => {
  const id = trimmedId.value;
  if (!id) return null;
  return formatVrcInstanceCell(id, t);
});
</script>

<style scoped>
:deep(.encounter-instance-link) {
  color: var(--color-brand);
  text-decoration: none;
  font-size: var(--font-size-12);
}

:deep(.encounter-instance-link:hover) {
  text-decoration: underline;
  color: var(--color-brand-hover);
}

.mono {
  font-family: ui-monospace, monospace;
  font-size: var(--font-size-12);
  word-break: break-all;
}
</style>
