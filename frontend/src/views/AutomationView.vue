<template>
  <div class="automation-view">
    <h1 class="page-title">{{ t("automation.title") }}</h1>
    <el-text
      type="info"
      size="small"
      style="display: block; margin-bottom: 1rem"
    >
      {{ t("automation.intro") }}
    </el-text>
    <div class="rules-section">
      <!-- ルールリスト -->
      <div class="rules-list">
        <el-button class="btn-add" @click="addNew">{{
          t("automation.addRule")
        }}</el-button>
        <div
          v-for="r in rules"
          :key="r.id"
          class="rule-card"
          :class="{ active: selected?.id === r.id, disabled: !r.isEnabled }"
          @click="select(r)"
        >
          <div class="rule-header">
            <span class="rule-name">{{ r.name }}</span>
            <div class="toggle-wrap" @click.stop>
              <el-switch
                v-model="r.isEnabled"
                size="small"
                @change="toggleRule(r)"
              />
            </div>
          </div>
          <div class="rule-summary">
            <span
              >{{ t("automation.summaryIf") }}
              {{ triggerLabel(r.triggerType) }} →
              {{ t("automation.summaryThen") }}
              {{ actionLabel(r.actionType, r.actionPayload) }}</span
            >
          </div>
        </div>
      </div>

      <!-- ルールエディタ -->
      <el-card v-if="selected" class="rule-editor" shadow="never">
        <template #header>
          {{ selected.id ? t("automation.editRule") : t("automation.newRule") }}
        </template>
        <el-form label-position="top" @submit.prevent="save">
          <el-form-item :label="t('automation.ruleName')">
            <el-input
              v-model="selected.name"
              :placeholder="t('automation.ruleNamePh')"
              required
            />
          </el-form-item>

          <el-divider>{{ t("automation.ifTrigger") }}</el-divider>
          <el-form-item :label="t('automation.condition')">
            <el-select
              v-model="selected.triggerType"
              required
              style="width: 100%; max-width: 320px"
            >
              <el-option
                v-for="opt in triggerOptions"
                :key="opt.value"
                :value="opt.value"
                :label="opt.label"
              />
            </el-select>
          </el-form-item>

          <el-divider>{{ t("automation.thenAction") }}</el-divider>
          <el-form-item :label="t('automation.action')">
            <el-select
              v-model="selected.actionType"
              required
              style="width: 100%; max-width: 320px"
            >
              <el-option
                v-for="opt in actionOptions"
                :key="opt.value"
                :value="opt.value"
                :label="opt.label"
              />
            </el-select>
          </el-form-item>
          <el-form-item
            v-if="selected.actionType === 'change_status'"
            :label="t('automation.status')"
          >
            <el-select
              v-model="statusValue"
              style="width: 100%; max-width: 320px"
            >
              <el-option
                v-for="opt in statusOptions"
                :key="opt.value"
                :value="opt.value"
                :label="opt.label"
              />
            </el-select>
          </el-form-item>

          <div class="editor-actions">
            <el-button type="primary" @click="save">{{
              t("automation.save")
            }}</el-button>
            <el-button
              v-if="selected.id"
              type="danger"
              plain
              @click="confirmDelete"
            >
              {{ t("automation.delete") }}
            </el-button>
            <el-button @click="cancelEdit">{{
              t("automation.cancel")
            }}</el-button>
          </div>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed } from "vue";
import { useI18n } from "vue-i18n";
import { ElMessageBox } from "element-plus";
import { App, type AutomationRuleDTO } from "../wails/app";

const { t } = useI18n();

const rules = ref<AutomationRuleDTO[]>([]);
const selected = ref<AutomationRuleDTO | null>(null);

const triggerOptions = computed(() => [
  { value: "afk_detected", label: t("automation.triggerAfk") },
  { value: "friend_joined", label: t("automation.triggerFriendJoined") },
]);

const actionOptions = computed(() => [
  { value: "change_status", label: t("automation.actionChangeStatus") },
]);

const statusOptions = computed(() => [
  { value: "busy", label: t("automation.statusBusy") },
  { value: "ask me", label: t("automation.statusAskMe") },
  { value: "join me", label: t("automation.statusJoinMe") },
]);

const statusValue = ref<string>("busy");

function triggerLabel(triggerType: string): string {
  return (
    triggerOptions.value.find((o) => o.value === triggerType)?.label ??
    triggerType
  );
}

function actionLabel(actionType: string, actionPayload?: string): string {
  const base =
    actionOptions.value.find((o) => o.value === actionType)?.label ??
    actionType;
  if (actionType === "change_status" && actionPayload) {
    try {
      const p = JSON.parse(actionPayload) as { status?: string };
      const s = p?.status ?? "";
      if (s) {
        const lbl = statusOptions.value.find((o) => o.value === s)?.label ?? s;
        return `${base} → ${lbl}`;
      }
    } catch {
      /* ignore */
    }
  }
  return base;
}

function buildActionPayload(): string {
  if (selected.value?.actionType === "change_status") {
    return JSON.stringify({ status: statusValue.value });
  }
  return "";
}

watch(
  () => selected.value?.actionPayload,
  (payload) => {
    if (selected.value?.actionType === "change_status" && payload) {
      try {
        const p = JSON.parse(payload) as { status?: string };
        statusValue.value = p?.status ?? "busy";
      } catch {
        statusValue.value = "busy";
      }
    }
  },
  { immediate: true },
);

onMounted(loadRules);

async function loadRules() {
  rules.value = await App.listAutomationRules();
}

function select(r: AutomationRuleDTO) {
  selected.value = { ...r };
  if (r.actionType === "change_status" && r.actionPayload) {
    try {
      const p = JSON.parse(r.actionPayload) as { status?: string };
      statusValue.value = p?.status ?? "busy";
    } catch {
      statusValue.value = "busy";
    }
  }
}

function addNew() {
  selected.value = {
    id: "",
    name: "",
    triggerType: "afk_detected",
    conditionJson: "",
    actionType: "change_status",
    actionPayload: JSON.stringify({ status: "busy" }),
    isEnabled: true,
  };
  statusValue.value = "busy";
}

function cancelEdit() {
  selected.value = null;
}

async function save() {
  if (!selected.value) return;
  const rule: AutomationRuleDTO = {
    ...selected.value,
    actionPayload: buildActionPayload(),
  };
  await App.saveAutomationRule(rule);
  await loadRules();
  if (rule.id) {
    selected.value = rules.value.find((r) => r.id === rule.id) ?? null;
  } else {
    const match = rules.value.find(
      (r) =>
        r.name === rule.name &&
        r.triggerType === rule.triggerType &&
        r.actionType === rule.actionType &&
        r.actionPayload === rule.actionPayload,
    );
    selected.value = match ?? null;
  }
}

async function toggleRule(r: AutomationRuleDTO) {
  await App.toggleAutomationRule(r.id, r.isEnabled);
  await loadRules();
}

async function confirmDelete() {
  if (!selected.value?.id) return;
  try {
    await ElMessageBox.confirm(
      t("automation.deleteConfirm", { name: selected.value.name }),
      t("common.confirm"),
      {
        confirmButtonText: t("common.delete"),
        cancelButtonText: t("common.cancel"),
        type: "warning",
        confirmButtonClass: "el-button--danger",
      },
    );
  } catch {
    return;
  }
  await App.deleteAutomationRule(selected.value.id);
  selected.value = null;
  await loadRules();
}
</script>

<style scoped>
.rules-section {
  display: flex;
  gap: 1.5rem;
}

.rules-list {
  width: 280px;
  flex-shrink: 0;
}

.btn-add {
  width: 100%;
  margin-bottom: 0.5rem;
  border-style: dashed !important;
  color: var(--text-secondary) !important;
}

.btn-add:hover {
  color: var(--accent) !important;
}

.rule-card {
  padding: 0.75rem;
  margin-bottom: 0.5rem;
  background: var(--bg-secondary);
  border-radius: var(--radius);
  cursor: pointer;
  transition:
    background 0.15s,
    opacity 0.15s;
}

.rule-card:hover,
.rule-card.active {
  background: var(--bg-tertiary);
}

.rule-card.disabled {
  opacity: 0.65;
}

.rule-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  margin-bottom: 0.35rem;
}

.rule-name {
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.toggle-wrap {
  flex-shrink: 0;
}

.rule-summary {
  font-size: 0.8rem;
  color: var(--text-secondary);
  font-style: italic;
}

.rule-editor {
  flex: 1;
  min-width: 0;
  background: var(--bg-secondary) !important;
  border-color: var(--border) !important;
}

.rule-editor :deep(.el-card__header) {
  font-weight: 600;
  border-bottom-color: var(--border);
}

.editor-actions {
  margin-top: 1rem;
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}
</style>
