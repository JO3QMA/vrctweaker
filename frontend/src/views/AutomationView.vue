<template>
  <div class="automation-view">
    <h1 class="page-title">{{ t("automation.title") }}</h1>
    <div
      class="automation-tabs"
      role="tablist"
      :aria-label="t('automation.title')"
    >
      <button
        ref="rulesTabEl"
        type="button"
        role="tab"
        class="automation-tab"
        :class="{ 'is-active': tab === 'rules' }"
        :aria-selected="tab === 'rules'"
        :tabindex="tab === 'rules' ? 0 : -1"
        aria-controls="automation-panel-rules"
        data-testid="automation-tab-rules"
        @click="tab = 'rules'"
        @keydown="onTabKeydown"
      >
        <span class="automation-tab__label">{{
          t("automation.tabRules")
        }}</span>
        <span
          class="automation-tab__count"
          data-testid="automation-rules-count"
          >{{ items.length }}</span
        >
      </button>
      <button
        ref="logTabEl"
        type="button"
        role="tab"
        class="automation-tab"
        :class="{ 'is-active': tab === 'log' }"
        :aria-selected="tab === 'log'"
        :tabindex="tab === 'log' ? 0 : -1"
        aria-controls="automation-panel-log"
        data-testid="automation-tab-run-log"
        @click="tab = 'log'"
        @keydown="onTabKeydown"
      >
        <span class="automation-tab__label">{{
          t("automation.runLogTitle")
        }}</span>
        <span
          class="automation-tab__count"
          data-testid="automation-log-count"
          >{{ runLog.length }}</span
        >
      </button>
    </div>

    <div class="automation-body">
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

      <div
        v-show="tab === 'rules'"
        id="automation-panel-rules"
        role="tabpanel"
        class="automation-panes"
        :class="{ 'is-editing': editor !== null && !preferList }"
      >
        <AutomationItemList
          :items="items"
          :active-id="activeId"
          :trigger-labels="triggerLabels"
          :disabled="!runtime.available"
          @select="selectItem"
          @toggle="toggleItem"
          @add-rule="addRule"
          @add-script="addScript"
        />
        <AutomationRuleEditor
          v-if="editor"
          :editor="editor"
          :saving="saving"
          :is-dirty="isDirty"
          :disabled="!runtime.available"
          :name-error="nameError"
          :trigger-options="triggerOptions"
          :status-options="statusOptions"
          :power-plan-presets="powerPlanPresets"
          :weekday-options="weekdayOptions"
          :action-options="actionOptions"
          :power-plans="powerPlans"
          :friends="friends"
          @save="save"
          @cancel="cancelEdit"
          @delete="confirmDelete"
          @back="showList"
          @add-action="addAction"
          @remove-action="removeAction"
          @move-action="moveAction"
        />
        <div v-else class="automation-editor automation-editor--empty">
          <p class="automation-empty text-body-sm">
            {{ t("automation.emptyEditor") }}
          </p>
        </div>
      </div>

      <div
        v-show="tab === 'log'"
        id="automation-panel-log"
        role="tabpanel"
        class="automation-log-host"
      >
        <AutomationRunLogPanel :entries="runLog" @refresh="refreshRunLog" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  ref,
  computed,
  watch,
  nextTick,
  onMounted,
  onBeforeUnmount,
} from "vue";
import { onBeforeRouteLeave } from "vue-router";
import { useI18n } from "vue-i18n";
import { ElMessageBox } from "element-plus";
import VtAlert from "../components/VtAlert.vue";
import AutomationItemList from "./automation/AutomationItemList.vue";
import AutomationRuleEditor from "./automation/AutomationRuleEditor.vue";
import AutomationRunLogPanel from "./automation/AutomationRunLogPanel.vue";
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
const nameError = ref("");
const tab = ref<"rules" | "log">("rules");
const preferList = ref(false);
const rulesTabEl = ref<HTMLButtonElement | null>(null);
const logTabEl = ref<HTMLButtonElement | null>(null);
let loadGen = 0;
let runLogGen = 0;
let eventsOff: (() => void) | undefined;

const isDirty = computed(
  () => !!editor.value && JSON.stringify(editor.value) !== savedSnapshot.value,
);

const controlsDisabled = computed(() => !runtime.value.available);

const activeId = computed(() =>
  editor.value && editor.value.isNew !== true ? editor.value.id : "",
);

const triggerOptions = computed(() => [
  { value: "friend_joined", label: t("automation.triggerFriendJoined") },
  { value: "friend_left", label: t("automation.triggerFriendLeft") },
  { value: "schedule.tick", label: t("automation.triggerSchedule") },
  { value: "vrchat.process", label: t("automation.triggerVrchatProcess") },
]);

const triggerLabels = computed(() => {
  const labels: Record<string, string> = {};
  for (const opt of triggerOptions.value) {
    labels[opt.value] = opt.label;
  }
  return labels;
});

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

watch(
  () => editor.value?.name,
  () => {
    if (nameError.value) nameError.value = "";
  },
);

function captureSnapshot() {
  savedSnapshot.value = editor.value ? JSON.stringify(editor.value) : "";
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

async function refreshRunLog() {
  try {
    await loadRunLog();
  } catch {
    showToast.error(t("automation.runLogLoadError"));
  }
}

function onTabKeydown(event: KeyboardEvent) {
  if (event.key !== "ArrowLeft" && event.key !== "ArrowRight") return;
  event.preventDefault();
  const next = tab.value === "rules" ? "log" : "rules";
  tab.value = next;
  void nextTick(() => {
    const el = next === "rules" ? rulesTabEl.value : logTabEl.value;
    el?.focus();
  });
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

function showList() {
  void (async () => {
    if (isDirty.value) {
      const choice = await promptUnsavedChoice();
      if (choice === "cancel") return;
      if (choice === "save") {
        if (!(await save())) return;
      } else if (savedSnapshot.value) {
        try {
          editor.value = JSON.parse(savedSnapshot.value) as EditorState;
        } catch {
          editor.value = null;
        }
        nameError.value = "";
      }
    }
    preferList.value = true;
  })();
}

async function selectItem(item: AutomationItemDTO) {
  if (!(await guardUnsaved())) return;
  nameError.value = "";
  try {
    editor.value = dtoToEditor(item);
    preferList.value = false;
    captureSnapshot();
  } catch {
    editor.value = null;
    savedSnapshot.value = "";
    showToast.error(t("automation.itemParseError"));
  }
}

function addRule() {
  void (async () => {
    if (controlsDisabled.value) return;
    if (!(await guardUnsaved())) return;
    nameError.value = "";
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
    preferList.value = false;
    captureSnapshot();
  })();
}

function addScript() {
  void (async () => {
    if (controlsDisabled.value) return;
    if (!(await guardUnsaved())) return;
    nameError.value = "";
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
    preferList.value = false;
    captureSnapshot();
  })();
}

function cancelEdit() {
  void (async () => {
    if (!(await guardUnsaved())) return;
    editor.value = null;
    savedSnapshot.value = "";
    nameError.value = "";
    preferList.value = true;
  })();
}

function addAction() {
  if (!editor.value || controlsDisabled.value) return;
  editor.value.actions.push(defaultAction());
}

function removeAction(idx: number) {
  if (!editor.value || controlsDisabled.value) return;
  if (editor.value.actions.length <= 1) return;
  editor.value.actions.splice(idx, 1);
}

function moveAction(idx: number, direction: -1 | 1) {
  if (!editor.value || controlsDisabled.value) return;
  const next = idx + direction;
  const actions = editor.value.actions;
  const moving = actions[idx];
  if (!moving || next < 0 || next >= actions.length) return;
  actions.splice(idx, 1);
  actions.splice(next, 0, moving);
}

async function save(): Promise<boolean> {
  if (!editor.value || saving.value || controlsDisabled.value) return false;
  if (editor.value.name.trim() === "") {
    nameError.value = t("automation.nameRequired");
    return false;
  }
  nameError.value = "";
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
    preferList.value = false;
    return true;
  } catch {
    showToast.error(t("automation.saveError"));
    return false;
  } finally {
    saving.value = false;
  }
}

async function toggleItem(item: AutomationItemDTO, enabled: boolean) {
  if (controlsDisabled.value) return;
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
  if (!editor.value?.id || controlsDisabled.value) return;
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
  nameError.value = "";
  preferList.value = true;
  await loadItems();
}
</script>

<style scoped>
.automation-view {
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
  container-type: inline-size;
  container-name: automation-frame;
}

.automation-view .page-title {
  margin-bottom: var(--space-block);
}

.automation-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-4);
  min-width: 0;
}

.automation-tab {
  display: inline-flex;
  align-items: center;
  gap: var(--space-action-group);
  flex: 0 1 auto;
  min-width: 0;
  max-width: 100%;
  padding: var(--space-8) var(--space-12);
  border: none;
  border-bottom: 2px solid transparent;
  border-radius: var(--radius);
  background: transparent;
  color: var(--color-text-secondary);
  font-size: var(--font-size-14);
  line-height: var(--line-height-normal);
  font-weight: var(--font-weight-400);
  font-family: inherit;
}

.automation-tab:hover {
  color: var(--color-text-primary);
  background: var(--color-bg-muted);
}

.automation-tab.is-active {
  border-bottom-color: var(--color-brand);
  color: var(--color-text-primary);
  background: var(--color-bg-elevated);
  font-weight: var(--font-weight-600);
}

.automation-tab__label {
  min-width: 0;
  overflow-wrap: anywhere;
}

.automation-tab__count {
  color: var(--color-text-secondary);
  font-weight: var(--font-weight-400);
}

.automation-tab:focus-visible {
  outline: var(--focus-ring-width) solid var(--focus-ring-color);
  outline-offset: var(--focus-ring-offset);
}

.automation-body {
  margin-top: var(--space-block);
  flex: 1 1 auto;
  min-height: 0;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-block);
  overflow: hidden;
}

.automation-intro,
.runtime-alert {
  margin: 0;
  flex-shrink: 0;
}

.automation-intro {
  color: var(--color-text-secondary);
  overflow-wrap: anywhere;
}

.automation-panes,
.automation-log-host {
  display: flex;
  align-items: stretch;
  gap: var(--space-block);
  flex: 1 1 auto;
  min-height: 0;
  min-width: 0;
}

.automation-list {
  flex: 0 0 280px;
  width: 280px;
  max-width: 280px;
  min-width: 0;
  min-height: 0;
}

.automation-editor {
  flex: 1 1 auto;
  max-width: 800px;
  min-width: 0;
  min-height: 0;
}

.automation-editor--empty {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-block);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  background: var(--color-bg-elevated);
  overflow: auto;
  scrollbar-gutter: stable;
}

.automation-empty {
  margin: 0;
  color: var(--color-text-secondary);
  overflow-wrap: anywhere;
  text-align: center;
}

.automation-log-host :deep(.automation-log) {
  width: 100%;
}

@container automation-frame (max-width: 719px) {
  .automation-list,
  .automation-editor {
    flex: 1 1 auto;
    width: 100%;
    max-width: none;
  }

  .automation-panes.is-editing .automation-list {
    display: none;
  }

  .automation-panes:not(.is-editing) .automation-editor {
    display: none;
  }

  .automation-editor :deep(.automation-back) {
    display: inline-flex;
  }
}
</style>
