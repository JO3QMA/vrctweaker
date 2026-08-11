<script setup lang="ts">
import { computed, useAttrs } from "vue";
import { type VtAlertProps, vtAlertElementType } from "./vtAlertVariants";

const props = defineProps<VtAlertProps>();
const attrs = useAttrs();

defineOptions({
  inheritAttrs: false,
});

const alertType = computed(() => vtAlertElementType(props.variant));
</script>

<template>
  <el-alert
    v-bind="attrs"
    :type="alertType"
    :title="title"
    :description="description"
    :closable="false"
    show-icon
  >
    <slot />
  </el-alert>
</template>
