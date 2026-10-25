<script setup lang="ts">
import { useI18n } from "vue-i18n";
import VtButton from "../../components/VtButton.vue";
import VtTag from "../../components/VtTag.vue";
import type { AutomationItemDTO } from "../../wails/app";

const props = defineProps<{
  items: AutomationItemDTO[];
  activeId: string;
  triggerLabels: Record<string, string>;
  disabled: boolean;
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
  const triggerType = item.triggerType ?? "";
  if (!Object.prototype.hasOwnProperty.call(props.triggerLabels, triggerType)) {
    return "";
  }
  return props.triggerLabels[triggerType] ?? "";
}

function onToggle(item: AutomationItemDTO, value: string | number | boolean) {
  if (props.disabled) return;
  emit("toggle", item, value === true);
}

function onCreate(command: string) {
  if (props.disabled) return;
  if (command === "script") emit("addScript");
  else if (command === "rule") emit("addRule");
}
</script>

<template>
  <el-card
    class="automation-list section-card"
    shadow="never"
    data-testid="automation-list"
  >
    <div class="automation-list__actions">
      <el-dropdown
        trigger="click"
        popper-class="automation-dropdown"
        :disabled="disabled"
        @command="onCreate"
      >
        <VtButton
          variant="secondary"
          class="btn-add"
          data-testid="add-item"
          :disabled="disabled"
        >
          {{ t("automation.addItem") }}
        </VtButton>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="rule" data-testid="add-rule">
              {{ t("automation.addRule") }}
            </el-dropdown-item>
            <el-dropdown-item command="script" data-testid="add-script">
              {{ t("automation.addScript") }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
    <div class="automation-list__rows">
      <p v-if="items.length === 0" class="automation-list__empty text-body-sm">
        {{ t("automation.emptyList") }}
      </p>
      <div v-else role="list">
        <div
          v-for="item in items"
          :key="item.id"
          class="automation-list__item"
          role="listitem"
        >
          <div
            class="rule-card"
            :class="{
              active: activeId === item.id,
              disabled: !item.isEnabled,
            }"
          >
            <div
              class="rule-card__select"
              role="button"
              tabindex="0"
              :aria-current="activeId === item.id ? 'true' : undefined"
              @click="emit('select', item)"
              @keydown.enter.prevent="emit('select', item)"
              @keydown.space.prevent="emit('select', item)"
            >
              <div class="rule-header">
                <span class="rule-name">{{ item.name }}</span>
              </div>
              <div class="rule-summary">
                <VtTag size="small" variant="neutral">{{
                  item.kind === "script"
                    ? t("automation.kindScript")
                    : t("automation.kindRule")
                }}</VtTag>
                <span class="rule-summary__text">{{ summary(item) }}</span>
              </div>
            </div>
            <div class="toggle-wrap" @click.stop @keydown.stop>
              <el-switch
                :model-value="item.isEnabled"
                :disabled="disabled"
                size="small"
                @change="(v: string | number | boolean) => onToggle(item, v)"
              />
            </div>
          </div>
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
  padding: var(--space-form-field) var(--space-block);
  border-bottom: 1px solid var(--color-border);
}

.automation-list__actions :deep(.el-dropdown) {
  display: block;
  width: 100%;
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

.automation-list__empty {
  margin: 0;
  padding: var(--space-block);
  color: var(--color-text-secondary);
  overflow-wrap: anywhere;
}

.rule-card {
  position: relative;
  padding: var(--space-form-field) var(--space-block);
  border-left: 2px solid transparent;
  background: transparent;
  min-width: 0;
}

.rule-card__select {
  min-width: 0;
  cursor: pointer;
}

.rule-card:has(.rule-card__select:focus-visible) {
  outline: var(--focus-ring-width) solid var(--focus-ring-color);
  outline-offset: calc(-1 * var(--focus-ring-offset));
}

.automation-list__item + .automation-list__item {
  border-top: 1px solid var(--color-border);
}

.rule-card.active {
  border-left-color: var(--color-brand);
  background: var(--color-bg-muted);
}

.rule-card.disabled .rule-name,
.rule-card.disabled .rule-summary {
  color: var(--color-text-muted);
}

.rule-header {
  display: flex;
  align-items: center;
  min-width: 0;
  min-height: 24px;
  padding-right: calc(30px + var(--space-action-group));
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
  position: absolute;
  z-index: 1;
  top: var(--space-form-field);
  right: var(--space-block);
  display: flex;
  align-items: center;
  height: 24px;
}

.rule-summary {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
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
}

.rule-summary__text {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>

<style>
.automation-dropdown.el-popper,
.automation-dropdown .el-dropdown-menu {
  background-color: var(--color-bg-muted);
}
</style>
