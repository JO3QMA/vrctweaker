<template>
  <div class="automation-view">
    <h1 class="page-title">
      オートメーション
    </h1>
    <p class="page-desc">
      IF-THEN 形式のルールで、トリガー発生時にアクションを実行します。
    </p>
    <div class="rules-section">
      <div class="rules-list">
        <button
          class="btn-add"
          @click="addNew"
        >
          + 新規ルール
        </button>
        <div
          v-for="r in rules"
          :key="r.id"
          class="rule-card"
          :class="{ active: selected?.id === r.id, disabled: !r.isEnabled }"
          @click="select(r)"
        >
          <div class="rule-header">
            <span class="rule-name">{{ r.name }}</span>
            <label
              class="toggle-wrap"
              title="有効/無効"
              @click.stop
            >
              <input
                v-model="r.isEnabled"
                type="checkbox"
                @change="toggleRule(r)"
              >
              <span class="toggle-label">{{ r.isEnabled ? 'ON' : 'OFF' }}</span>
            </label>
          </div>
          <div class="rule-summary">
            <span class="if-then">
              IF {{ triggerLabel(r.triggerType) }} → THEN {{ actionLabel(r.actionType, r.actionPayload) }}
            </span>
          </div>
        </div>
      </div>
      <div
        v-if="selected"
        class="rule-editor"
      >
        <h3 class="editor-title">
          {{ selected.id ? 'ルールを編集' : '新規ルール' }}
        </h3>
        <form
          class="editor-form"
          @submit.prevent="save"
        >
          <label>ルール名</label>
          <input
            v-model="selected.name"
            type="text"
            placeholder="例: AFK時にステータスをbusyに"
            required
          >
          <div class="if-then-block">
            <h4>IF（トリガー）</h4>
            <label>条件</label>
            <select
              v-model="selected.triggerType"
              required
            >
              <option
                v-for="opt in triggerOptions"
                :key="opt.value"
                :value="opt.value"
              >
                {{ opt.label }}
              </option>
            </select>
          </div>
          <div class="if-then-block">
            <h4>THEN（アクション）</h4>
            <label>アクション</label>
            <select
              v-model="selected.actionType"
              required
            >
              <option
                v-for="opt in actionOptions"
                :key="opt.value"
                :value="opt.value"
              >
                {{ opt.label }}
              </option>
            </select>
            <template v-if="selected.actionType === 'change_status'">
              <label>ステータス</label>
              <select v-model="statusValue">
                <option
                  v-for="opt in statusOptions"
                  :key="opt.value"
                  :value="opt.value"
                >
                  {{ opt.label }}
                </option>
              </select>
            </template>
          </div>
          <div class="editor-actions">
            <button
              type="submit"
              class="btn-save"
            >
              保存
            </button>
            <button
              v-if="selected.id"
              type="button"
              class="btn-delete"
              @click="confirmDelete"
            >
              削除
            </button>
            <button
              type="button"
              class="btn-cancel"
              @click="cancelEdit"
            >
              キャンセル
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { App, type AutomationRuleDTO } from '../wails/app'

const TRIGGER_OPTIONS = [
  { value: 'afk_detected', label: 'AFK検出時' },
  { value: 'friend_joined', label: 'フレンド参加時' },
] as const

const ACTION_OPTIONS = [
  { value: 'change_status', label: 'ステータスを変更' },
] as const

const STATUS_OPTIONS = [
  { value: 'busy', label: 'Busy' },
  { value: 'ask me', label: 'Ask Me' },
  { value: 'join me', label: 'Join Me' },
] as const

const rules = ref<AutomationRuleDTO[]>([])
const selected = ref<AutomationRuleDTO | null>(null)

const triggerOptions = TRIGGER_OPTIONS
const actionOptions = ACTION_OPTIONS
const statusOptions = STATUS_OPTIONS

const statusValue = ref<string>('busy')

function triggerLabel(triggerType: string): string {
  return triggerOptions.find(o => o.value === triggerType)?.label ?? triggerType
}

function actionLabel(actionType: string, actionPayload?: string): string {
  const base = actionOptions.find(o => o.value === actionType)?.label ?? actionType
  if (actionType === 'change_status' && actionPayload) {
    try {
      const p = JSON.parse(actionPayload) as { status?: string }
      const s = p?.status ?? ''
      if (s) {
        const lbl = statusOptions.find(o => o.value === s)?.label ?? s
        return `${base} → ${lbl}`
      }
    } catch {
      /* ignore */
    }
  }
  return base
}

function buildActionPayload(): string {
  if (selected.value?.actionType === 'change_status') {
    return JSON.stringify({ status: statusValue.value })
  }
  return ''
}

watch(
  () => selected.value?.actionPayload,
  (payload) => {
    if (selected.value?.actionType === 'change_status' && payload) {
      try {
        const p = JSON.parse(payload) as { status?: string }
        statusValue.value = p?.status ?? 'busy'
      } catch {
        statusValue.value = 'busy'
      }
    }
  },
  { immediate: true },
)

onMounted(loadRules)

async function loadRules() {
  rules.value = await App.listAutomationRules()
}

function select(r: AutomationRuleDTO) {
  selected.value = { ...r }
  if (r.actionType === 'change_status' && r.actionPayload) {
    try {
      const p = JSON.parse(r.actionPayload) as { status?: string }
      statusValue.value = p?.status ?? 'busy'
    } catch {
      statusValue.value = 'busy'
    }
  }
}

function addNew() {
  selected.value = {
    id: '',
    name: '',
    triggerType: 'afk_detected',
    conditionJson: '',
    actionType: 'change_status',
    actionPayload: JSON.stringify({ status: 'busy' }),
    isEnabled: true,
  }
  statusValue.value = 'busy'
}

function cancelEdit() {
  selected.value = null
}

async function save() {
  if (!selected.value) return
  const rule: AutomationRuleDTO = {
    ...selected.value,
    actionPayload: buildActionPayload(),
  }
  await App.saveAutomationRule(rule)
  await loadRules()
  if (rule.id) {
    selected.value = rules.value.find(r => r.id === rule.id) ?? null
  } else {
    const match = rules.value.find(
      r =>
        r.name === rule.name &&
        r.triggerType === rule.triggerType &&
        r.actionType === rule.actionType &&
        r.actionPayload === rule.actionPayload,
    )
    selected.value = match ?? null
  }
}

async function toggleRule(r: AutomationRuleDTO) {
  await App.toggleAutomationRule(r.id, r.isEnabled)
  await loadRules()
}

async function confirmDelete() {
  if (!selected.value?.id) return
  if (!window.confirm(`「${selected.value.name}」を削除しますか？`)) return
  await App.deleteAutomationRule(selected.value.id)
  selected.value = null
  await loadRules()
}
</script>

<style scoped>
.page-title { margin: 0 0 0.25rem; font-size: 1.5rem; }
.page-desc { margin: 0 0 1rem; font-size: 0.9rem; color: var(--text-secondary); }
.rules-section { display: flex; gap: 1.5rem; }
.rules-list { width: 280px; flex-shrink: 0; }
.btn-add {
  width: 100%; padding: 0.5rem; margin-bottom: 0.5rem;
  background: var(--bg-tertiary); border: 1px dashed var(--border);
  border-radius: var(--radius); color: var(--text-secondary); cursor: pointer;
}
.btn-add:hover { background: var(--bg-secondary); color: var(--accent); }
.rule-card {
  padding: 0.75rem; margin-bottom: 0.5rem; background: var(--bg-secondary);
  border-radius: var(--radius); cursor: pointer; transition: opacity 0.15s;
}
.rule-card:hover, .rule-card.active { background: var(--bg-tertiary); }
.rule-card.disabled { opacity: 0.65; }
.rule-header { display: flex; align-items: center; justify-content: space-between; gap: 0.5rem; margin-bottom: 0.35rem; }
.rule-name { font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.toggle-wrap { display: flex; align-items: center; gap: 0.35rem; cursor: pointer; flex-shrink: 0; }
.toggle-label { font-size: 0.7rem; color: var(--text-secondary); }
.rule-summary { font-size: 0.8rem; color: var(--text-secondary); }
.if-then { font-style: italic; }
.rule-editor { flex: 1; min-width: 0; }
.editor-title { margin: 0 0 1rem; font-size: 1.1rem; }
.editor-form { display: flex; flex-direction: column; gap: 0.5rem; }
.editor-form label { font-size: 0.85rem; margin-top: 0.5rem; }
.editor-form label:first-of-type { margin-top: 0; }
.editor-form input[type="text"],
.editor-form select {
  width: 100%; max-width: 320px; padding: 0.5rem;
  background: var(--bg-tertiary); border: 1px solid var(--border);
  border-radius: var(--radius); color: var(--text-primary);
}
.if-then-block { margin-top: 1rem; padding-top: 1rem; border-top: 1px solid var(--border); }
.if-then-block h4 { margin: 0 0 0.5rem; font-size: 1rem; }
.editor-actions { margin-top: 1rem; display: flex; gap: 0.5rem; flex-wrap: wrap; }
.btn-save, .btn-delete, .btn-cancel {
  padding: 0.5rem 1rem; border-radius: var(--radius); border: none; cursor: pointer;
}
.btn-save { background: var(--accent); color: white; }
.btn-delete { background: #8b2635; color: white; }
.btn-cancel { background: var(--bg-tertiary); color: var(--text-primary); }
</style>
