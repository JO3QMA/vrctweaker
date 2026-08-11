<script setup lang="ts">
import { computed, useAttrs } from "vue";
import { type VtTagProps, vtTagElementType } from "./vtTagVariants";

const props = defineProps<VtTagProps>();
const attrs = useAttrs();

defineOptions({
  inheritAttrs: false,
});

const tagType = computed(() => vtTagElementType(props.variant));
const isNeutral = computed(() => tagType.value === undefined);
</script>

<template>
  <el-tag
    v-bind="attrs"
    :type="isNeutral ? 'info' : tagType"
    :effect="isNeutral ? 'plain' : undefined"
    :class="
      isNeutral
        ? attrs.class
          ? ['vt-tag--neutral', attrs.class]
          : 'vt-tag--neutral'
        : undefined
    "
    :size="props.size"
  >
    <slot />
  </el-tag>
</template>

<style scoped>
.vt-tag--neutral {
  --el-tag-bg-color: var(--color-bg-muted);
  --el-tag-border-color: var(--color-border);
  --el-tag-text-color: var(--color-text-secondary);
}
</style>
