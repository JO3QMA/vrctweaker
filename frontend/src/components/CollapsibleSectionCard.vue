<template>
  <el-card
    class="section-card collapsible-section-card"
    :class="{ 'section-card--collapsed': !expanded }"
    shadow="never"
  >
    <template #header>
      <button
        type="button"
        class="section-card__toggle"
        :aria-expanded="expanded"
        :aria-controls="panelId"
        @click="toggle"
      >
        <VtIcon
          size="default"
          class="section-card__toggle-icon"
          aria-hidden="true"
        >
          <CaretBottom v-if="expanded" />
          <CaretRight v-else />
        </VtIcon>
        <span class="section-card__toggle-label">
          <slot name="title">{{ title }}</slot>
        </span>
      </button>
    </template>
    <div :id="panelId" class="section-card__panel" role="region">
      <slot />
    </div>
  </el-card>
</template>

<script lang="ts">
let nextCollapsibleSectionId = 0;
</script>

<script setup lang="ts">
import { CaretBottom, CaretRight } from "@element-plus/icons-vue";
import VtIcon from "./VtIcon.vue";

defineOptions({ name: "CollapsibleSectionCard" });

withDefaults(
  defineProps<{
    /** `title` スロットが空のときに表示する見出し */
    title?: string;
  }>(),
  { title: "" },
);

const expanded = defineModel<boolean>({ default: true });

nextCollapsibleSectionId += 1;
const panelId = `collapsible-section-${nextCollapsibleSectionId}`;

function toggle(): void {
  expanded.value = !expanded.value;
}
</script>
