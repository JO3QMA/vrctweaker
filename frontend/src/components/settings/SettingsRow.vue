<script setup lang="ts">
import { computed, useId, useSlots } from "vue";

const props = withDefaults(
  defineProps<{
    label: string;
    description?: string;
    control?: "field" | "number" | "switch" | "button" | "value";
  }>(),
  {
    description: "",
    control: "field",
  },
);

const slots = useSlots();
const labelId = useId();

const showDescription = computed(
  () => Boolean(props.description) || Boolean(slots.description),
);

const rowClass = computed(() => ({
  "setting-row--fixed":
    props.control === "number" ||
    props.control === "switch" ||
    props.control === "value",
  "setting-row--number": props.control === "number",
  "setting-row--button": props.control === "button",
  "setting-row--center": !showDescription.value,
}));
</script>

<template>
  <div class="setting-row" :class="rowClass">
    <div class="setting-row__text">
      <div :id="labelId" class="setting-row__label">{{ label }}</div>
      <p v-if="showDescription" class="setting-row__desc text-body-sm">
        <slot name="description">{{ description }}</slot>
      </p>
    </div>
    <div class="setting-row__control">
      <slot :label-id="labelId" :aria-label="label" />
    </div>
  </div>
</template>
