<template>
  <div class="automation-view">
    <h1 class="page-title">オートメーション</h1>
    <el-text
      type="info"
      size="small"
      style="display: block; margin-bottom: 1rem"
    >
      IF-THEN 形式のルールで、トリガー発生時にアクションを実行します。
    </el-text>
    <div class="rules-section">
      <!-- ルールリスト -->
      <div class="rules-list">
        <el-button class="btn-add" @click="addNew">+ 新規ルール</el-button>
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
              >IF {{ triggerLabel(r.triggerType) }} → THEN
              {{ actionLabel(r.actionType, r.actionPayload) }}</span
            >
          </div>
        </div>
      </div>

      <!-- ルールエディタ -->
      <el-card v-if="selected" class="rule-editor" shadow="never">
        <template #header>
          {{ selected.id ? "ルールを編集" : "新規ルール" }}
        </template>
        <el-form label-position="top" @submit.prevent="save">
          <el-form-item label="ルール名">
            <el-input
              v-model="selected.name"
              placeholder="例: AFK時にステータスをbusyに"
              required
            />
          </el-form-item>

          <el-divider>IF（トリガー）</el-divider>
          <el-form-item label="条件">
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

          <el-divider>THEN（アクション）</el-divider>
          <el-form-item label="アクション">
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
            label="ステータス"
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
            <el-button type="primary" @click="save">保存</el-button>
            <el-button
              v-if="selected.id"
              type="danger"
              plain
              @click="confirmDelete"
            >
              削除
            </el-button>
            <el-button @click="cancelEdit">キャンセル</el-button>
          </div>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { ElMessageBox } from "element-plus";
import { App, type AutomationRuleDTO } from "../wails/app";

const TRIGGER_OPTIONS = [
  { value: "afk_detected", label: "AFK検出時" },
  { value: "friend_joined", label: "フレンド参加時" },
] as const;

const ACTION_OPTIONS = [
  { value: "change_status", label: "ステータスを変更" },
] as const;

const STATUS_OPTIONS = [
  { value: "busy", label: "Busy" },
  { value: "ask me", label: "Ask Me" },
  { value: "join me", label: "Join Me" },
] as const;

const rules = ref<AutomationRuleDTO[]>([]);
const selected = ref<AutomationRuleDTO | null>(null);

const triggerOptions = TRIGGER_OPTIONS;
const actionOptions = ACTION_OPTIONS;
const statusOptions = STATUS_OPTIONS;

const statusValue = ref<string>("busy");

function triggerLabel(triggerType: string): string {
  return (
    triggerOptions.find((o) => o.value === triggerType)?.label ?? triggerType
  );
}

function actionLabel(actionType: string, actionPayload?: string): string {
  const base =
    actionOptions.find((o) => o.value === actionType)?.label ?? actionType;
  if (actionType === "change_status" && actionPayload) {
    try {
      const p = JSON.parse(actionPayload) as { status?: string };
      const s = p?.status ?? "";
      if (s) {
        const lbl = statusOptions.find((o) => o.value === s)?.label ?? s;
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
      `「${selected.value.name}」を削除しますか？`,
      "確認",
      {
        confirmButtonText: "削除",
        cancelButtonText: "キャンセル",
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
