<template>
  <span
    v-if="label"
    class="vrc-instance-key-label"
    :title="`${label} (${trimmedKey})`"
    >{{ label }}</span
  >
  <span v-else-if="trimmedKey" class="vrc-instance-key-label mono">{{
    trimmedKey
  }}</span>
  <span v-else>{{ t("common.dash") }}</span>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { formatVrcInstanceLabel } from "../utils/vrcInstanceDisplay";

const props = defineProps<{
  instanceKey?: string;
}>();

const { t } = useI18n();

const trimmedKey = computed(() => props.instanceKey?.trim() ?? "");

const label = computed(() => {
  const key = trimmedKey.value;
  if (!key) return null;
  return formatVrcInstanceLabel(key, t);
});
</script>

<style scoped>
.vrc-instance-key-label {
  font-size: var(--font-size-12);
  word-break: break-all;
}

.mono {
  font-family: ui-monospace, monospace;
}
</style>
