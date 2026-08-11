<script setup lang="ts">
import { computed, useAttrs, useSlots } from "vue";
import { type VtAlertProps, vtAlertElementType } from "./vtAlertVariants";

const props = defineProps<VtAlertProps>();
const attrs = useAttrs();
const slots = useSlots();

defineOptions({
  inheritAttrs: false,
});

const alertType = computed(() => vtAlertElementType(props.variant));
const hasDefaultSlot = computed(() => !!slots.default);
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
    <template v-if="hasDefaultSlot" #default>
      <slot />
    </template>
  </el-alert>
</template>
