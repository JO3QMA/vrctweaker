<script setup lang="ts">
import { computed, useAttrs } from "vue";
import { type VtTagProps, vtTagElementType } from "./vtTagVariants";

const props = defineProps<VtTagProps>();
const attrs = useAttrs();

defineOptions({
  inheritAttrs: false,
});

const tagType = computed(() => vtTagElementType(props.variant));

const tagBind = computed(() => {
  if (props.variant === "neutral") {
    const { type: _type, ...forwardedAttrs } = attrs;
    return {
      ...forwardedAttrs,
      class: ["vt-tag--neutral", attrs.class],
      effect: "plain" as const,
    };
  }
  return {
    ...attrs,
    type: tagType.value,
  };
});
</script>

<template>
  <el-tag v-bind="tagBind">
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
