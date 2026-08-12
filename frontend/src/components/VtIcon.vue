<script setup lang="ts">
import { computed } from "vue";
import type { VtIconProps } from "./vtIconVariants";

const props = withDefaults(defineProps<VtIconProps>(), {
  decorative: true,
});

defineOptions({
  inheritAttrs: false,
});

const sizeClass = computed(() => `vt-icon--size-${props.size}`);
</script>

<template>
  <el-icon
    v-bind="$attrs"
    class="vt-icon"
    :class="sizeClass"
    :aria-hidden="decorative ? 'true' : undefined"
  >
    <slot />
  </el-icon>
</template>
