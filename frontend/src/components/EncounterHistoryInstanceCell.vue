<template>
  <a
    v-if="cell"
    class="encounter-instance-link"
    :href="cell.href"
    :title="cell.title"
    :aria-label="t('encounterHistory.openInVrchat', { text: cell.text })"
    target="_blank"
    rel="noopener noreferrer"
    data-testid="encounter-instance-link"
  >
    {{ cell.text }}
  </a>
  <span v-else-if="trimmedId" class="mono">{{ trimmedId }}</span>
  <span v-else>{{ t("common.dash") }}</span>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
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
.encounter-instance-link {
  color: var(--color-brand);
  text-decoration: none;
  font-size: var(--font-size-12);
}

.encounter-instance-link:hover {
  text-decoration: underline;
  color: var(--color-brand-hover);
}

.mono {
  font-family: ui-monospace, monospace;
  font-size: var(--font-size-12);
  word-break: break-all;
}
</style>
