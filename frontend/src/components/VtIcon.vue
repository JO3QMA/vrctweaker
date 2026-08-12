<script setup lang="ts">
import { computed, useAttrs } from "vue";
import type { VtIconProps } from "./vtIconVariants";

const props = withDefaults(defineProps<VtIconProps>(), {
  decorative: true,
});

const attrs = useAttrs();

defineOptions({
  inheritAttrs: false,
});

const passthroughAttrs = computed(() => {
  const result = { ...attrs };
  delete result["aria-hidden"];
  return result;
});

const sizeClass = computed(() => `vt-icon--size-${props.size}`);

const ariaHidden = computed(() => {
  if (props.decorative) return "true";
  const fromAttrs = attrs["aria-hidden"];
  if (fromAttrs !== undefined) return String(fromAttrs);
  return "false";
});
</script>

<template>
  <el-icon
    v-bind="passthroughAttrs"
    class="vt-icon"
    :class="sizeClass"
    :aria-hidden="ariaHidden"
  >
    <slot />
  </el-icon>
</template>
