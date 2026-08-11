<script setup lang="ts">
import { computed, useAttrs } from "vue";
import {
  type VtTagProps,
  omitVtTagAttr,
  vtTagElementType,
} from "./vtTagVariants";

const props = defineProps<VtTagProps>();
const attrs = useAttrs();

defineOptions({
  inheritAttrs: false,
});

const tagType = computed(() => vtTagElementType(props.variant));

const tagBind = computed(() => {
  // `type` is owned by vtTagElementType; strip caller attrs to avoid el-tag conflicts.
  const forwarded = omitVtTagAttr(attrs as Record<string, unknown>, "type");
  if (tagType.value === undefined) {
    return {
      ...forwarded,
      type: "info" as const,
      effect: "plain" as const,
      class: attrs.class
        ? ["vt-tag--neutral", attrs.class]
        : ["vt-tag--neutral"],
    };
  }
  return {
    ...forwarded,
    type: tagType.value,
  };
});
</script>

<template>
  <el-tag v-bind="tagBind" :size="size">
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
