<script setup lang="ts">
import { computed } from "vue";
import { type VtAlertProps, vtAlertElementType } from "./vtAlertVariants";

const props = defineProps<VtAlertProps>();

defineOptions({
  inheritAttrs: false,
});

const alertType = computed(() => vtAlertElementType(props.variant));
</script>

<template>
  <el-alert v-bind="$attrs" :type="alertType" :closable="false" show-icon />
</template>
