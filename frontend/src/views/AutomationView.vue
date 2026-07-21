<template>
  <div class="automation-view">
    <h1 class="page-title">{{ t("automation.title") }}</h1>
    <el-alert
      v-if="!runtime.available"
      type="warning"
      :title="t('automation.unavailableTitle')"
      :description="
        t(`automation.reason.${runtime.reasonKey || 'subsystemUnavailable'}`)
      "
      show-icon
      :closable="false"
      style="margin-bottom: 1rem"
    />
    <el-text
      type="info"
      size="small"
      style="display: block; margin-bottom: 1rem"
    >
      {{ t("automation.intro") }}
    </el-text>
    <div
      v-if="isDirty"
      class="unsaved-banner"
      data-testid="unsaved-banner"
      :title="t('automation.unsavedBanner')"
    >
      {{ t("automation.unsavedBanner") }}
    </div>

    <div class="automation-layout">
      <div class="items-list">
        <el-button class="btn-add" data-testid="add-rule" @click="addRule">
          {{ t("automation.addRule") }}
        </el-button>
        <el-button class="btn-add" data-testid="add-script" @click="addScript">
          {{ t("automation.addScript") }}
        </el-button>
        <div
          v-for="item in items"
          :key="item.id"
          class="rule-card"
          :class="{
            active: editor?.id === item.id && editor?.isNew !== true,
            disabled: !item.isEnabled,
          }"
          @click="selectItem(item)"
        >
          <div class="rule-header">
            <span class="rule-name">{{ item.name }}</span>
            <div class="toggle-wrap" @click.stop>
              <el-switch
                :model-value="item.isEnabled"
                size="small"
                @change="(v: boolean) => toggleItem(item, v)"
              />
            </div>
          </div>
          <div class="rule-summary">
            <el-tag size="small" type="info">{{
              item.kind === "script"
                ? t("automation.kindScript")
                : t("automation.kindRule")
            }}</el-tag>
            <span>{{ itemSummary(item) }}</span>
          </div>
        </div>
      </div>

      <el-card v-if="editor" class="rule-editor" shadow="never">
        <template #header>
          <span v-if="editor.isNew">{{
            editor.kind === "script"
              ? t("automation.newScript")
              : t("automation.newRule")
          }}</span>
          <span v-else>{{
            editor.kind === "script"
              ? t("automation.editScript")
              : t("automation.editRule")
          }}</span>
          <span
            v-if="isDirty"
            class="unsaved-dot"
            data-testid="unsaved-dot"
            :title="t('automation.unsavedBanner')"
          />
        </template>

        <el-form label-position="top" @submit.prevent="save">
          <el-form-item :label="t('automation.itemName')">
            <el-input
              v-model="editor.name"
              :placeholder="t('automation.ruleNamePh')"
              required
            />
          </el-form-item>

          <template v-if="editor.kind === 'rule'">
            <el-divider>{{ t("automation.sectionWhen") }}</el-divider>
            <el-form-item :label="t('automation.trigger')">
              <el-select v-model="editor.triggerType" style="width: 100%">
                <el-option
                  v-for="opt in triggerOptions"
                  :key="opt.value"
                  :value="opt.value"
                  :label="opt.label"
                />
              </el-select>
            </el-form-item>
            <el-form-item
              v-if="editor.triggerType === 'schedule.tick'"
              :label="t('automation.schedule')"
            >
              <el-checkbox-group v-model="editor.scheduleWeekdays">
                <el-checkbox
                  v-for="d in weekdayOptions"
                  :key="d.value"
                  :label="d.value"
                >
                  {{ d.label }}
                </el-checkbox>
              </el-checkbox-group>
              <div class="time-row">
                <el-input-number
                  v-model="editor.scheduleHour"
                  :min="0"
                  :max="23"
                />
                <span>:</span>
                <el-input-number
                  v-model="editor.scheduleMinute"
                  :min="0"
                  :max="59"
                />
              </div>
            </el-form-item>

            <el-divider>{{ t("automation.sectionIf") }}</el-divider>
            <el-form-item>
              <el-checkbox v-model="editor.vrchatRunning">
                {{ t("automation.conditionVrchatRunning") }}
              </el-checkbox>
            </el-form-item>
            <el-form-item
              v-if="editor.triggerType === 'friend_joined'"
              :label="t('automation.conditionFriendIs')"
            >
              <el-select
                v-model="editor.friendUserId"
                clearable
                filterable
                style="width: 100%"
              >
                <el-option
                  v-for="f in friends"
                  :key="f.vrcUserId"
                  :value="f.vrcUserId"
                  :label="f.displayName"
                />
              </el-select>
            </el-form-item>

            <el-divider>{{ t("automation.sectionThen") }}</el-divider>
            <div
              v-for="(action, idx) in editor.actions"
              :key="idx"
              class="action-row"
            >
              <el-form-item :label="t('automation.action')">
                <el-select v-model="action.type" style="width: 100%">
                  <el-option
                    v-for="opt in actionOptions"
                    :key="opt.value"
                    :value="opt.value"
                    :label="opt.label"
                    :disabled="opt.disabled"
                  />
                </el-select>
              </el-form-item>
              <el-form-item
                v-if="action.type === 'change_status'"
                :label="t('automation.status')"
              >
                <el-select v-model="action.status" style="width: 100%">
                  <el-option
                    v-for="opt in statusOptions"
                    :key="opt.value"
                    :value="opt.value"
                    :label="opt.label"
                  />
                </el-select>
              </el-form-item>
              <template v-if="action.type === 'set_power_plan'">
                <el-form-item :label="t('automation.powerPlanMode')">
                  <el-radio-group v-model="action.powerPlanMode">
                    <el-radio value="preset">{{
                      t("automation.powerPlanPreset")
                    }}</el-radio>
                    <el-radio value="guid" :disabled="!powerPlans.length">{{
                      t("automation.powerPlanDetected")
                    }}</el-radio>
                  </el-radio-group>
                </el-form-item>
                <el-form-item
                  v-if="action.powerPlanMode === 'preset'"
                  :label="t('automation.powerPlanPreset')"
                >
                  <el-select
                    v-model="action.powerPlanPreset"
                    style="width: 100%"
                  >
                    <el-option
                      v-for="p in powerPlanPresets"
                      :key="p.value"
                      :value="p.value"
                      :label="p.label"
                    />
                  </el-select>
                </el-form-item>
                <el-form-item v-else :label="t('automation.powerPlanDetected')">
                  <el-select v-model="action.powerPlanGuid" style="width: 100%">
                    <el-option
                      v-for="p in powerPlans"
                      :key="p.guid"
                      :value="p.guid"
                      :label="p.name"
                    />
                  </el-select>
                </el-form-item>
              </template>
              <el-checkbox v-model="action.continueOnError">
                {{ t("automation.continueOnError") }}
              </el-checkbox>
              <el-button
                v-if="editor.actions.length > 1"
                type="danger"
                text
                @click="removeAction(idx)"
              >
                {{ t("automation.removeAction") }}
              </el-button>
            </div>
            <el-button
              v-if="editor.actions.length < 10"
              class="btn-add-action"
              @click="addAction"
            >
              {{ t("automation.addAction") }}
            </el-button>
            <el-text type="info" size="small" class="partial-hint">
              {{ t("automation.partialApplyHint") }}
            </el-text>
          </template>

          <template v-else>
            <el-divider>{{ t("automation.scriptSource") }}</el-divider>
            <el-input
              v-model="editor.scriptSource"
              type="textarea"
              :rows="14"
              class="script-editor"
              :placeholder="t('automation.scriptPlaceholder')"
            />
          </template>

          <div class="editor-actions">
            <el-button
              type="primary"
              data-testid="save-item"
              :loading="saving"
              @click="save"
            >
              {{ t("automation.save") }}
            </el-button>
            <el-button
              v-if="editor.id && !editor.isNew"
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

      <el-card class="run-log-panel" shadow="never">
        <template #header>
          <span>{{ t("automation.runLogTitle") }}</span>
          <el-button text size="small" @click="loadRunLog">
            {{ t("common.refresh") }}
          </el-button>
        </template>
        <el-table :data="runLog" size="small" stripe empty-text="—">
          <el-table-column
            prop="at"
            :label="t('automation.runLogAt')"
            width="150"
          />
          <el-table-column prop="itemName" :label="t('automation.itemName')" />
          <el-table-column
            prop="eventType"
            :label="t('automation.runLogEvent')"
            width="120"
          />
          <el-table-column :label="t('automation.runLogResult')" width="100">
            <template #default="{ row }">
              <el-tag :type="row.success ? 'success' : 'danger'" size="small">
                {{
                  row.success
                    ? t("automation.runLogSuccess")
                    : t("automation.runLogFailure")
                }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column :label="t('automation.runLogActions')" width="80">
            <template #default="{ row }">
              {{ row.actionsCompleted }}/{{ row.actionsTotal }}
            </template>
          </el-table-column>
          <el-table-column
            prop="contextLabel"
            :label="t('automation.runLogContext')"
          />
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import { onBeforeRouteLeave } from "vue-router";
import { useI18n } from "vue-i18n";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  App,
  type AutomationItemDTO,
  type AutomationRunLogEntryDTO,
  type AutomationRuntimeStatusDTO,
  type DetectedPowerPlanDTO,
  type UserCacheDTO,
} from "../wails/app";
import { getRuntime } from "../wails/runtime";
import {
  defaultAction,
  dtoToEditor,
  newAutomationId,
  type EditorState,
} from "./automationEditorMapping";

const { t } = useI18n();

const items = ref<AutomationItemDTO[]>([]);
const editor = ref<EditorState | null>(null);
const savedSnapshot = ref("");
const runLog = ref<AutomationRunLogEntryDTO[]>([]);
const runtime = ref<AutomationRuntimeStatusDTO>({
  available: false,
  reasonKey: "subsystemUnavailable",
});
const powerPlans = ref<DetectedPowerPlanDTO[]>([]);
const friends = ref<UserCacheDTO[]>([]);
const saving = ref(false);
let loadGen = 0;
let runLogGen = 0;
let eventsOff: (() => void) | undefined;

const isDirty = computed(
  () => !!editor.value && JSON.stringify(editor.value) !== savedSnapshot.value,
);

const triggerOptions = computed(() => [
  { value: "friend_joined", label: t("automation.triggerFriendJoined") },
  { value: "schedule.tick", label: t("automation.triggerSchedule") },
  { value: "vrchat.process", label: t("automation.triggerVrchatProcess") },
]);

const statusOptions = computed(() => [
  { value: "busy", label: t("automation.statusBusy") },
  { value: "ask me", label: t("automation.statusAskMe") },
  { value: "join me", label: t("automation.statusJoinMe") },
]);

const powerPlanPresets = computed(() => [
  { value: "balanced", label: t("automation.powerPlanBalanced") },
  { value: "high_performance", label: t("automation.powerPlanHigh") },
  { value: "power_saver", label: t("automation.powerPlanSaver") },
]);

const weekdayOptions = computed(() =>
  [0, 1, 2, 3, 4, 5, 6].map((value) => ({
    value,
    label: t(`automation.weekday${value}`),
  })),
);

const actionOptions = computed(() => [
  { value: "change_status", label: t("automation.actionChangeStatus") },
  {
    value: "set_power_plan",
    label: t("automation.actionSetPowerPlan"),
    disabled: powerPlans.value.length === 0,
  },
]);

function editorToDto(state: EditorState): AutomationItemDTO {
  const dto: AutomationItemDTO = {
    id: state.id,
    name: state.name,
    kind: state.kind,
    isEnabled: state.isEnabled,
  };
  if (state.kind === "script") {
    dto.scriptSource = state.scriptSource;
    return dto;
  }
  dto.triggerType = state.triggerType;
  if (state.triggerType === "schedule.tick") {
    dto.scheduleJson = JSON.stringify({
      weekdays: state.scheduleWeekdays,
      hour: state.scheduleHour,
      minute: state.scheduleMinute,
    });
  }
  const conds: Array<{ type: string; vrcUserId?: string }> = [];
  if (state.vrchatRunning) conds.push({ type: "vrchat_running" });
  if (state.friendUserId) {
    conds.push({ type: "friend_is", vrcUserId: state.friendUserId });
  }
  dto.conditionsJson = JSON.stringify(conds);
  dto.actionsJson = JSON.stringify(
    state.actions.map((a) => {
      const payload: Record<string, string> = {};
      if (a.type === "change_status") payload.status = a.status;
      if (a.type === "set_power_plan") {
        if (a.powerPlanMode === "guid" && a.powerPlanGuid) {
          payload.guid = a.powerPlanGuid;
        } else {
          payload.preset = a.powerPlanPreset;
        }
      }
      return {
        type: a.type,
        payload,
        continueOnError: a.continueOnError || undefined,
      };
    }),
  );
  return dto;
}

function captureSnapshot() {
  savedSnapshot.value = editor.value ? JSON.stringify(editor.value) : "";
}

function itemSummary(item: AutomationItemDTO): string {
  if (item.kind === "script") return t("automation.summaryScript");
  const trigger =
    triggerOptions.value.find((o) => o.value === item.triggerType)?.label ??
    item.triggerType;
  return `${t("automation.summaryIf")} ${trigger}`;
}

async function loadItems() {
  const gen = ++loadGen;
  const list = await App.listAutomationItems();
  if (gen !== loadGen) return;
  items.value = list;
}

async function loadRunLog() {
  const gen = ++runLogGen;
  const log = await App.getAutomationRunLog();
  if (gen !== runLogGen) return;
  runLog.value = log;
}

async function loadRuntime() {
  runtime.value = await App.getAutomationRuntimeStatus();
}

async function loadPowerPlans() {
  powerPlans.value = await App.listDetectedPowerPlans();
}

async function loadFriends() {
  friends.value = await App.friends();
}

onMounted(async () => {
  try {
    await Promise.all([
      loadItems(),
      loadRunLog(),
      loadRuntime(),
      loadPowerPlans(),
      loadFriends(),
    ]);
  } catch {
    // Keep whatever succeeded; runtime defaults to unavailable until loadRuntime wins.
  }
  const rt = getRuntime();
  eventsOff = rt?.EventsOn?.("automation:run-log-changed", () => {
    void loadRunLog().catch(() => {});
  });
});

onBeforeUnmount(() => {
  eventsOff?.();
});

type UnsavedChoice = "save" | "discard" | "cancel";

async function promptUnsavedChoice(): Promise<UnsavedChoice> {
  try {
    await ElMessageBox.confirm(
      t("automation.unsavedSwitchMessage"),
      t("automation.unsavedSwitchTitle"),
      {
        confirmButtonText: t("automation.saveAndContinue"),
        cancelButtonText: t("automation.discardAndContinue"),
        distinguishCancelAndClose: true,
        type: "warning",
      },
    );
    return "save";
  } catch (action) {
    if (action === "cancel") return "discard";
    return "cancel";
  }
}

async function guardUnsaved(): Promise<boolean> {
  if (!isDirty.value) return true;
  const choice = await promptUnsavedChoice();
  if (choice === "cancel") return false;
  if (choice === "save") return await save();
  return true;
}

onBeforeRouteLeave(async () => {
  return guardUnsaved();
});

async function selectItem(item: AutomationItemDTO) {
  if (!(await guardUnsaved())) return;
  try {
    editor.value = dtoToEditor(item);
    captureSnapshot();
  } catch {
    ElMessage.error(t("automation.itemParseError"));
  }
}

function addRule() {
  void (async () => {
    if (!(await guardUnsaved())) return;
    editor.value = {
      id: "",
      name: "",
      kind: "rule",
      isEnabled: true,
      isNew: true,
      triggerType: "friend_joined",
      scheduleWeekdays: [1, 2, 3, 4, 5],
      scheduleHour: 0,
      scheduleMinute: 0,
      vrchatRunning: false,
      friendUserId: "",
      actions: [defaultAction()],
      scriptSource: "",
    };
    captureSnapshot();
  })();
}

function addScript() {
  void (async () => {
    if (!(await guardUnsaved())) return;
    editor.value = {
      id: "",
      name: "",
      kind: "script",
      isEnabled: true,
      isNew: true,
      triggerType: "",
      scheduleWeekdays: [],
      scheduleHour: 0,
      scheduleMinute: 0,
      vrchatRunning: false,
      friendUserId: "",
      actions: [defaultAction()],
      scriptSource:
        'tweaker.subscribe("friend_joined", function(ev, payload)\n  -- your logic\nend)\n',
    };
    captureSnapshot();
  })();
}

function cancelEdit() {
  void (async () => {
    if (!(await guardUnsaved())) return;
    editor.value = null;
    savedSnapshot.value = "";
  })();
}

function addAction() {
  if (!editor.value) return;
  editor.value.actions.push(defaultAction());
}

function removeAction(idx: number) {
  editor.value?.actions.splice(idx, 1);
}

async function save(): Promise<boolean> {
  if (!editor.value || saving.value) return false;
  saving.value = true;
  try {
    const dto = editorToDto(editor.value);
    if (!dto.id) {
      dto.id = newAutomationId();
      editor.value.id = dto.id;
    }
    await App.saveAutomationItem(dto);
    try {
      await loadItems();
      const match = items.value.find((it) => it.id === dto.id);
      if (match) {
        try {
          editor.value = dtoToEditor(match);
          editor.value.isNew = false;
        } catch {
          editor.value.isNew = false;
          ElMessage.error(t("automation.itemParseError"));
        }
      } else {
        editor.value.isNew = false;
      }
    } catch {
      // Persist succeeded; list refresh is best-effort.
      editor.value.isNew = false;
    }
    captureSnapshot();
    return true;
  } catch {
    ElMessage.error(t("automation.saveError"));
    return false;
  } finally {
    saving.value = false;
  }
}

async function toggleItem(item: AutomationItemDTO, enabled: boolean) {
  const previous = item.isEnabled;
  try {
    await App.toggleAutomationItem(item.id, enabled);
    item.isEnabled = enabled;
    await loadItems();
  } catch {
    item.isEnabled = previous;
    ElMessage.error(t("automation.toggleError"));
  }
}

async function confirmDelete() {
  if (!editor.value?.id) return;
  try {
    await ElMessageBox.confirm(
      t("automation.deleteConfirm", { name: editor.value.name }),
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
  await App.deleteAutomationItem(editor.value.id);
  editor.value = null;
  savedSnapshot.value = "";
  await loadItems();
}
</script>

<style scoped>
.automation-layout {
  display: grid;
  grid-template-columns: 280px 1fr 320px;
  gap: 1rem;
  align-items: start;
}

@media (max-width: 1100px) {
  .automation-layout {
    grid-template-columns: 1fr;
  }
}

.items-list {
  width: 100%;
}

.btn-add {
  width: 100%;
  margin-bottom: 0.5rem;
  border-style: dashed !important;
  color: var(--text-secondary) !important;
}

.rule-card {
  padding: 0.75rem;
  margin-bottom: 0.5rem;
  background: var(--bg-secondary);
  border-radius: var(--radius);
  cursor: pointer;
}

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

.rule-summary {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.rule-editor,
.run-log-panel {
  background: var(--bg-secondary) !important;
  border-color: var(--border) !important;
}

.rule-editor :deep(.el-card__header) {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 600;
}

.editor-actions {
  margin-top: 1rem;
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.weekday-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.time-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.action-row {
  padding: 0.75rem;
  margin-bottom: 0.75rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
}

.btn-add-action {
  width: 100%;
  margin-bottom: 0.5rem;
  border-style: dashed !important;
}

.partial-hint {
  display: block;
  margin-bottom: 1rem;
}

.script-editor :deep(textarea) {
  font-family: ui-monospace, monospace;
  font-size: 0.85rem;
}

.unsaved-banner {
  margin-bottom: 0.75rem;
  padding: 0.5rem 0.75rem;
  border-radius: var(--radius);
  background: color-mix(in srgb, var(--el-color-warning) 18%, transparent);
  color: var(--el-color-warning);
  font-size: 0.85rem;
}

.unsaved-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--el-color-warning);
  flex-shrink: 0;
}

.run-log-panel :deep(.el-card__header) {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
</style>
