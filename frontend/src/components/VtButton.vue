<script setup lang="ts">
import { computed } from "vue";
import type { VtButtonProps } from "./vtButtonVariants";

const props = defineProps<VtButtonProps>();

defineOptions({
  inheritAttrs: false,
});

const buttonType = computed(() => {
  if (props.variant === "primary") return "primary";
  if (props.variant === "danger") return "danger";
  return undefined;
});

const isText = computed(() => props.variant === "tertiary");

const isPlain = computed(
  () => props.variant === "danger" && props.plain === true,
);
</script>

<template>
  <el-button v-bind="$attrs" :type="buttonType" :text="isText" :plain="isPlain">
    <slot />
  </el-button>
</template>
