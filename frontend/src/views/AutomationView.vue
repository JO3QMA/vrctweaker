<template>
  <div class="automation-view">
    <h1 class="page-title">{{ t("automation.title") }}</h1>
    <VtAlert
      v-if="!runtime.available"
      class="runtime-alert"
      variant="warning"
      :title="t('automation.unavailableTitle')"
      :description="
        t(`automation.reason.${runtime.reasonKey || 'subsystemUnavailable'}`)
      "
    />
    <p class="text-body-sm automation-intro">
      {{ t("automation.intro") }}
    </p>
    <VtAlert
      v-if="isDirty"
      class="unsaved-banner"
      data-testid="unsaved-banner"
      variant="warning"
      :title="t('automation.unsavedBanner')"
    />

    <div class="automation-layout">
      <div class="items-list">
        <VtButton
          variant="secondary"
          class="btn-add"
          data-testid="add-rule"
          @click="addRule"
        >
          {{ t("automation.addRule") }}
        </VtButton>
        <VtButton
          variant="secondary"
          class="btn-add"
          data-testid="add-script"
          @click="addScript"
        >
          {{ t("automation.addScript") }}
        </VtButton>
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
              <VtSwitch
                :model-value="item.isEnabled"
                size="small"
                @change="(v: boolean) => toggleItem(item, v)"
              />
            </div>
          </div>
          <div class="rule-summary">
            <VtTag size="small" variant="info">{{
              item.kind === "script"
                ? t("automation.kindScript")
                : t("automation.kindRule")
            }}</VtTag>
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
            <VtInput
              v-model="editor.name"
              :placeholder="t('automation.ruleNamePh')"
              required
            />
          </el-form-item>

          <template v-if="editor.kind === 'rule'">
            <el-divider>{{ t("automation.sectionWhen") }}</el-divider>
            <el-form-item :label="t('automation.trigger')">
              <VtSelect v-model="editor.triggerType" class="field-full-width">
                <el-option
                  v-for="opt in triggerOptions"
                  :key="opt.value"
                  :value="opt.value"
                  :label="opt.label"
                />
              </VtSelect>
            </el-form-item>
            <el-form-item
              v-if="editor.triggerType === 'schedule.tick'"
              :label="t('automation.schedule')"
            >
              <el-checkbox-group v-model="editor.scheduleWeekdays">
                <VtCheckbox
                  v-for="d in weekdayOptions"
                  :key="d.value"
                  :value="d.value"
                >
                  {{ d.label }}
                </VtCheckbox>
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
              <VtCheckbox v-model="editor.vrchatRunning">
                {{ t("automation.conditionVrchatRunning") }}
              </VtCheckbox>
            </el-form-item>
            <el-form-item
              v-if="editor.triggerType === 'friend_joined'"
              :label="t('automation.conditionFriendIs')"
            >
              <VtSelect
                v-model="editor.friendUserId"
                clearable
                filterable
                class="field-full-width"
              >
                <el-option
                  v-for="f in friends"
                  :key="f.vrcUserId"
                  :value="f.vrcUserId"
                  :label="f.displayName"
                />
              </VtSelect>
            </el-form-item>

            <el-divider>{{ t("automation.sectionThen") }}</el-divider>
            <div
              v-for="(action, idx) in editor.actions"
              :key="idx"
              class="action-row"
            >
              <el-form-item :label="t('automation.action')">
                <VtSelect v-model="action.type" class="field-full-width">
                  <el-option
                    v-for="opt in actionOptions"
                    :key="opt.value"
                    :value="opt.value"
                    :label="opt.label"
                    :disabled="opt.disabled"
                  />
                </VtSelect>
              </el-form-item>
              <el-form-item
                v-if="action.type === 'change_status'"
                :label="t('automation.status')"
              >
                <VtSelect v-model="action.status" class="field-full-width">
                  <el-option
                    v-for="opt in statusOptions"
                    :key="opt.value"
                    :value="opt.value"
                    :label="opt.label"
                  />
                </VtSelect>
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
                  <VtSelect
                    v-model="action.powerPlanPreset"
                    class="field-full-width"
                  >
                    <el-option
                      v-for="p in powerPlanPresets"
                      :key="p.value"
                      :value="p.value"
                      :label="p.label"
                    />
                  </VtSelect>
                </el-form-item>
                <el-form-item v-else :label="t('automation.powerPlanDetected')">
                  <VtSelect
                    v-model="action.powerPlanGuid"
                    class="field-full-width"
                  >
                    <el-option
                      v-for="p in powerPlans"
                      :key="p.guid"
                      :value="p.guid"
                      :label="p.name"
                    />
                  </VtSelect>
                </el-form-item>
              </template>
              <template v-if="action.type === 'set_vrchat_window_size'">
                <el-form-item :label="t('automation.windowWidth')">
                  <el-input-number
                    v-model="action.windowWidth"
                    :min="1"
                    :max="7680"
                    data-testid="window-width-input"
                  />
                </el-form-item>
                <el-form-item :label="t('automation.windowHeight')">
                  <el-input-number
                    v-model="action.windowHeight"
                    :min="1"
                    :max="4320"
                    data-testid="window-height-input"
                  />
                </el-form-item>
              </template>
              <VtCheckbox v-model="action.continueOnError">
                {{ t("automation.continueOnError") }}
              </VtCheckbox>
              <VtButton
                v-if="editor.actions.length > 1"
                variant="tertiary"
                @click="removeAction(idx)"
              >
                {{ t("automation.removeAction") }}
              </VtButton>
            </div>
            <VtButton
              v-if="editor.actions.length < 10"
              variant="secondary"
              class="btn-add-action"
              @click="addAction"
            >
              {{ t("automation.addAction") }}
            </VtButton>
            <p class="text-body-sm partial-hint">
              {{ t("automation.partialApplyHint") }}
            </p>
          </template>

          <template v-else>
            <el-divider>{{ t("automation.scriptSource") }}</el-divider>
            <VtInput
              v-model="editor.scriptSource"
              type="textarea"
              :rows="14"
              class="script-editor"
              :placeholder="t('automation.scriptPlaceholder')"
            />
          </template>

          <div class="editor-actions">
            <VtButton
              variant="primary"
              data-testid="save-item"
              :loading="saving"
              @click="save"
            >
              {{ t("automation.save") }}
            </VtButton>
            <VtButton
              v-if="editor.id && !editor.isNew"
              variant="danger"
              plain
              @click="confirmDelete"
            >
              {{ t("automation.delete") }}
            </VtButton>
            <VtButton variant="secondary" @click="cancelEdit">{{
              t("automation.cancel")
            }}</VtButton>
          </div>
        </el-form>
      </el-card>

      <el-card class="run-log-panel" shadow="never">
        <template #header>
          <span>{{ t("automation.runLogTitle") }}</span>
          <VtButton variant="tertiary" size="small" @click="loadRunLog">
            {{ t("common.refresh") }}
          </VtButton>
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
              <VtTag :variant="row.success ? 'success' : 'danger'" size="small">
                {{
                  row.success
                    ? t("automation.runLogSuccess")
                    : t("automation.runLogFailure")
                }}
              </VtTag>
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
import { ElMessageBox } from "element-plus";
import VtAlert from "../components/VtAlert.vue";
import VtButton from "../components/VtButton.vue";
import VtCheckbox from "../components/VtCheckbox.vue";
import VtInput from "../components/VtInput.vue";
import VtSelect from "../components/VtSelect.vue";
import VtSwitch from "../components/VtSwitch.vue";
import VtTag from "../components/VtTag.vue";
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
  editorToDto,
  newAutomationId,
  type EditorState,
} from "./automationEditorMapping";
import {
  VT_BUTTON_DANGER_CONFIRM_CLASS,
  VT_BUTTON_SECONDARY_CANCEL_CLASS,
} from "../components/vtButtonClasses";
import { showToast } from "../utils/showToast";

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
  {
    value: "set_vrchat_window_size",
    label: t("automation.actionSetVRChatWindowSize"),
  },
]);

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
    editor.value = null;
    savedSnapshot.value = "";
    showToast.error(t("automation.itemParseError"));
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
    }
    await App.saveAutomationItem(dto);
    // Commit id only after successful persist (avoid retry-as-update on failure).
    editor.value.id = dto.id;
    try {
      await loadItems();
      const match = items.value.find((it) => it.id === dto.id);
      if (match) {
        try {
          editor.value = dtoToEditor(match);
          editor.value.isNew = false;
        } catch {
          editor.value.isNew = false;
          showToast.error(t("automation.itemParseError"));
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
    showToast.error(t("automation.saveError"));
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
    showToast.error(t("automation.toggleError"));
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
        confirmButtonClass: VT_BUTTON_DANGER_CONFIRM_CLASS,
        cancelButtonClass: VT_BUTTON_SECONDARY_CANCEL_CLASS,
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
.automation-intro {
  display: block;
  margin: 0 0 var(--space-block);
  color: var(--color-text-secondary);
}

.runtime-alert {
  margin-bottom: var(--space-block);
}

.automation-layout {
  display: grid;
  grid-template-columns: 280px 1fr 320px;
  gap: var(--space-block);
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
  margin-bottom: var(--space-action-group);
  border-style: dashed !important;
  color: var(--color-text-secondary) !important;
}

.rule-card {
  padding: var(--space-form-field);
  margin-bottom: var(--space-action-group);
  background: var(--color-bg-elevated);
  border-radius: var(--radius);
  cursor: pointer;
}

.rule-card.active {
  background: var(--color-bg-muted);
}

.rule-card.disabled {
  opacity: 0.65;
}

.rule-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-action-group);
  margin-bottom: var(--space-inline-tight);
}

.rule-name {
  font-weight: var(--font-weight-500);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.rule-summary {
  display: flex;
  flex-direction: column;
  gap: var(--space-inline-tight);
  font-size: var(--font-size-12);
  color: var(--color-text-secondary);
}

.rule-editor,
.run-log-panel {
  background: var(--color-bg-elevated) !important;
  border-color: var(--color-border) !important;
}

.rule-editor :deep(.el-card__header) {
  display: flex;
  align-items: center;
  gap: var(--space-action-group);
  font-weight: var(--font-weight-600);
}

.field-full-width {
  width: 100%;
}

.editor-actions {
  margin-top: var(--space-block);
  display: flex;
  gap: var(--space-action-group);
  flex-wrap: wrap;
}

.weekday-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-action-group);
  margin-bottom: var(--space-action-group);
}

.time-row {
  display: flex;
  align-items: center;
  gap: var(--space-action-group);
}

.action-row {
  padding: var(--space-form-field);
  margin-bottom: var(--space-form-field);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
}

.btn-add-action {
  width: 100%;
  margin-bottom: var(--space-action-group);
  border-style: dashed !important;
}

.partial-hint {
  display: block;
  margin: 0 0 var(--space-block);
  color: var(--color-text-secondary);
}

.script-editor :deep(textarea) {
  font-family: ui-monospace, monospace;
  font-size: var(--font-size-12);
}

.unsaved-banner {
  margin-bottom: var(--space-form-field);
}

.unsaved-dot {
  width: var(--space-action-group);
  height: var(--space-action-group);
  border-radius: 50%;
  background: var(--color-warning);
  flex-shrink: 0;
}

.run-log-panel :deep(.el-card__header) {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
</style>
