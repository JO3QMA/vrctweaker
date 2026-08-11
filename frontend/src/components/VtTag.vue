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
  const rawAttrs = attrs as Record<string, unknown>;
  if (props.variant === "neutral") {
    const withoutType = omitVtTagAttr(rawAttrs, "type");
    return {
      ...omitVtTagAttr(withoutType, "size"),
      class: ["vt-tag--neutral", attrs.class],
      effect: "plain" as const,
    };
  }
  return {
    ...omitVtTagAttr(rawAttrs, "size"),
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
