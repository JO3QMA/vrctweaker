<script setup lang="ts">
import { useSlots } from "vue";

defineProps<{
  label: string;
  description?: string;
}>();

const slots = useSlots();
</script>

<template>
  <div class="path-row">
    <div class="path-row__text">
      <div class="path-row__label">{{ label }}</div>
      <p v-if="description" class="path-row__desc text-body-sm">
        {{ description }}
      </p>
    </div>
    <div class="path-input-group">
      <slot />
    </div>
    <div v-if="slots.below" class="path-row__below">
      <slot name="below" />
    </div>
  </div>
</template>
