<script setup lang="ts">
import { useI18n } from "vue-i18n";
import VtButton from "../../components/VtButton.vue";
import VtTag from "../../components/VtTag.vue";
import type { AutomationItemDTO } from "../../wails/app";

const props = defineProps<{
  items: AutomationItemDTO[];
  activeId: string;
  triggerLabels: Record<string, string>;
}>();

const emit = defineEmits<{
  select: [item: AutomationItemDTO];
  toggle: [item: AutomationItemDTO, enabled: boolean];
  addRule: [];
  addScript: [];
}>();

const { t } = useI18n();

function summary(item: AutomationItemDTO): string {
  if (item.kind === "script") return t("automation.summaryScript");
  const trigger =
    props.triggerLabels[item.triggerType ?? ""] ?? item.triggerType ?? "";
  return `${t("automation.summaryIf")} ${trigger}`;
}

function onToggle(item: AutomationItemDTO, value: string | number | boolean) {
  emit("toggle", item, value === true);
}
</script>

<template>
  <el-card
    class="automation-list section-card"
    shadow="never"
    data-testid="automation-list"
  >
    <div class="automation-list__actions">
      <VtButton
        variant="secondary"
        class="btn-add"
        data-testid="add-rule"
        @click="emit('addRule')"
      >
        {{ t("automation.addRule") }}
      </VtButton>
      <VtButton
        variant="secondary"
        class="btn-add"
        data-testid="add-script"
        @click="emit('addScript')"
      >
        {{ t("automation.addScript") }}
      </VtButton>
    </div>
    <div class="automation-list__rows" role="list">
      <div
        v-for="item in items"
        :key="item.id"
        class="rule-card"
        role="button"
        tabindex="0"
        :class="{
          active: activeId === item.id,
          disabled: !item.isEnabled,
        }"
        :aria-current="activeId === item.id ? 'true' : undefined"
        @click="emit('select', item)"
        @keydown.enter.prevent="emit('select', item)"
        @keydown.space.prevent="emit('select', item)"
      >
        <div class="rule-header">
          <span class="rule-name">{{ item.name }}</span>
          <div class="toggle-wrap" @click.stop @keydown.stop>
            <el-switch
              :model-value="item.isEnabled"
              size="small"
              @change="(v: string | number | boolean) => onToggle(item, v)"
            />
          </div>
        </div>
        <div class="rule-summary">
          <VtTag size="small" variant="info">{{
            item.kind === "script"
              ? t("automation.kindScript")
              : t("automation.kindRule")
          }}</VtTag>
          <span class="rule-summary__text">{{ summary(item) }}</span>
        </div>
      </div>
    </div>
  </el-card>
</template>

<style scoped>
.automation-list {
  height: 100%;
  min-height: 0;
  margin-bottom: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.automation-list :deep(.el-card__body) {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 0;
}

.automation-list__actions {
  display: flex;
  flex-direction: column;
  gap: var(--space-action-group);
  padding: var(--space-form-field) var(--space-block);
  border-bottom: 1px solid var(--color-border);
}

.automation-list__actions :deep(.el-button) {
  width: 100%;
  margin: 0;
  border-style: dashed !important;
  color: var(--color-text-secondary);
}

.automation-list__rows {
  flex: 1 1 auto;
  min-height: 0;
  overflow: auto;
  scrollbar-gutter: stable;
}

.rule-card {
  padding: var(--space-form-field) var(--space-block);
  border-left: 2px solid transparent;
  background: transparent;
  cursor: pointer;
  min-width: 0;
}

.rule-card + .rule-card {
  border-top: 1px solid var(--color-border);
}

.rule-card.active {
  border-left-color: var(--color-brand);
  background: var(--color-bg-muted);
}

.rule-card.disabled {
  opacity: 0.65;
}

.rule-card:focus-visible {
  outline: var(--focus-ring-width) solid var(--focus-ring-color);
  outline-offset: calc(-1 * var(--focus-ring-offset));
}

.rule-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-action-group);
  min-width: 0;
}

.rule-name {
  font-weight: var(--font-weight-500);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}

.toggle-wrap {
  flex-shrink: 0;
}

.rule-summary {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: var(--space-action-group);
  margin-top: var(--space-inline-tight);
  font-size: var(--font-size-12);
  line-height: var(--line-height-normal);
  color: var(--color-text-secondary);
  min-width: 0;
}

.rule-summary :deep(.el-tag) {
  flex: 0 0 auto;
  width: auto;
  max-width: 100%;
  height: auto;
  white-space: normal;
}

.rule-summary__text {
  flex: 1 1 auto;
  min-width: 0;
  overflow-wrap: anywhere;
}
</style>
