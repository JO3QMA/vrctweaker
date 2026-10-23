<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import SettingsRow from "../../components/settings/SettingsRow.vue";
import VtButton from "../../components/VtButton.vue";
import type { DetectedPowerPlanDTO, UserCacheDTO } from "../../wails/app";
import type { EditorState } from "../automationEditorMapping";

type SelectOption = { value: string; label: string; disabled?: boolean };
type WeekdayOption = { value: number; label: string };

const props = defineProps<{
  editor: EditorState;
  saving: boolean;
  isDirty: boolean;
  disabled: boolean;
  nameError: string;
  triggerOptions: SelectOption[];
  statusOptions: SelectOption[];
  powerPlanPresets: SelectOption[];
  weekdayOptions: WeekdayOption[];
  actionOptions: SelectOption[];
  powerPlans: DetectedPowerPlanDTO[];
  friends: UserCacheDTO[];
}>();

const emit = defineEmits<{
  save: [];
  cancel: [];
  delete: [];
  back: [];
  addAction: [];
  removeAction: [index: number];
  moveAction: [index: number, direction: -1 | 1];
}>();

const { t } = useI18n();

// The parent owns this reactive editor. Bind fields through an alias so the
// template does not assign the prop identifier directly.
const state = computed(() => props.editor);

const canDeleteItem = computed(
  () => !!props.editor.id && props.editor.isNew !== true,
);
</script>

<template>
  <div class="automation-editor rule-editor" data-testid="rule-editor">
    <el-form class="automation-editor__scroll" @submit.prevent="emit('save')">
      <VtButton
        class="automation-back"
        variant="tertiary"
        data-testid="automation-back"
        @click="emit('back')"
      >
        {{ t("automation.backToList") }}
      </VtButton>

      <div class="automation-editor-head">
        <span
          class="automation-editor-head__name"
          :class="{ 'is-empty': !state.name }"
        >
          {{ state.name || t("automation.ruleNamePh") }}
        </span>
        <span
          v-if="isDirty"
          class="unsaved-dot"
          data-testid="unsaved-dot"
          :title="t('automation.unsavedBanner')"
        />
        <el-switch
          v-model="state.isEnabled"
          :disabled="disabled"
          data-testid="editor-enabled"
          :aria-label="t('automation.enabled')"
        />
        <el-dropdown
          v-if="canDeleteItem"
          class="automation-header-more"
          trigger="click"
          :disabled="disabled"
          @command="emit('delete')"
        >
          <VtButton
            variant="tertiary"
            data-testid="editor-more"
            :disabled="disabled"
            :aria-label="t('automation.delete')"
          >
            ⋯
          </VtButton>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="delete" data-testid="delete-item-menu">
                {{ t("automation.delete") }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>

      <div class="automation-editor__cards">
        <el-card
          class="section-card settings-card--rows automation-card"
          shadow="never"
          data-testid="automation-section-basics"
        >
          <template #header>
            <span class="automation-card__title">{{
              t("automation.sectionBasics")
            }}</span>
          </template>
          <SettingsRow :label="t('automation.itemName')" control="field">
            <div class="automation-control">
              <el-input
                v-model="state.name"
                :placeholder="t('automation.ruleNamePh')"
                :disabled="disabled"
                data-testid="item-name-input"
                :aria-invalid="nameError ? 'true' : undefined"
                :aria-describedby="
                  nameError ? 'automation-name-error' : undefined
                "
              />
              <p
                v-if="nameError"
                id="automation-name-error"
                class="automation-field-error"
                data-testid="name-error"
                role="alert"
              >
                {{ nameError }}
              </p>
            </div>
          </SettingsRow>
        </el-card>

        <template v-if="editor.kind === 'rule'">
          <el-card
            class="section-card settings-card--rows automation-card"
            shadow="never"
            data-testid="automation-section-when"
          >
            <template #header>
              <span class="automation-card__title">{{
                t("automation.sectionWhen")
              }}</span>
            </template>
            <SettingsRow :label="t('automation.trigger')" control="field">
              <div class="automation-control">
                <el-select v-model="state.triggerType" :disabled="disabled">
                  <el-option
                    v-for="opt in triggerOptions"
                    :key="opt.value"
                    :value="opt.value"
                    :label="opt.label"
                  />
                </el-select>
              </div>
            </SettingsRow>
            <template v-if="editor.triggerType === 'schedule.tick'">
              <SettingsRow :label="t('automation.schedule')" control="field">
                <div class="automation-control">
                  <el-checkbox-group
                    v-model="state.scheduleWeekdays"
                    :disabled="disabled"
                  >
                    <el-checkbox
                      v-for="d in weekdayOptions"
                      :key="d.value"
                      :value="d.value"
                    >
                      {{ d.label }}
                    </el-checkbox>
                  </el-checkbox-group>
                </div>
              </SettingsRow>
              <SettingsRow
                :label="t('automation.scheduleTime')"
                control="field"
              >
                <div class="automation-time">
                  <el-input-number
                    v-model="state.scheduleHour"
                    :min="0"
                    :max="23"
                    :disabled="disabled"
                    :aria-label="t('automation.scheduleHour')"
                  />
                  <span class="automation-time__sep" aria-hidden="true">:</span>
                  <el-input-number
                    v-model="state.scheduleMinute"
                    :min="0"
                    :max="59"
                    :disabled="disabled"
                    :aria-label="t('automation.scheduleMinute')"
                  />
                </div>
              </SettingsRow>
            </template>
          </el-card>

          <el-card
            class="section-card settings-card--rows automation-card"
            shadow="never"
            data-testid="automation-section-if"
          >
            <template #header>
              <span class="automation-card__title">{{
                t("automation.sectionIf")
              }}</span>
            </template>
            <SettingsRow
              :label="t('automation.conditionVrchatRunning')"
              control="switch"
            >
              <el-switch v-model="state.vrchatRunning" :disabled="disabled" />
            </SettingsRow>
            <SettingsRow
              v-if="
                editor.triggerType === 'friend_joined' ||
                editor.triggerType === 'friend_left'
              "
              :label="t('automation.conditionFriendIs')"
              control="field"
            >
              <div class="automation-control">
                <el-select
                  v-model="state.friendUserId"
                  clearable
                  filterable
                  :disabled="disabled"
                  :placeholder="t('automation.friendPlaceholder')"
                >
                  <el-option
                    v-for="f in friends"
                    :key="f.vrcUserId"
                    :value="f.vrcUserId"
                    :label="f.displayName"
                  />
                </el-select>
              </div>
            </SettingsRow>
          </el-card>

          <el-card
            class="section-card settings-card--rows automation-card"
            shadow="never"
            data-testid="automation-section-then"
          >
            <template #header>
              <span class="automation-card__title">{{
                t("automation.sectionThen")
              }}</span>
            </template>
            <div
              v-for="(action, idx) in editor.actions"
              :key="action.clientId"
              class="automation-action"
            >
              <div class="automation-action-head">
                <span class="automation-action-head__title">{{
                  t("automation.actionHeading", { n: idx + 1 })
                }}</span>
                <div class="automation-action-head__tools">
                  <VtButton
                    variant="tertiary"
                    size="small"
                    :disabled="disabled || idx === 0"
                    :aria-label="t('automation.moveActionUp')"
                    @click="emit('moveAction', idx, -1)"
                  >
                    ↑
                  </VtButton>
                  <VtButton
                    variant="tertiary"
                    size="small"
                    :disabled="disabled || idx === editor.actions.length - 1"
                    :aria-label="t('automation.moveActionDown')"
                    @click="emit('moveAction', idx, 1)"
                  >
                    ↓
                  </VtButton>
                  <VtButton
                    variant="tertiary"
                    size="small"
                    :disabled="disabled || editor.actions.length <= 1"
                    @click="emit('removeAction', idx)"
                  >
                    {{ t("automation.removeAction") }}
                  </VtButton>
                </div>
              </div>
              <SettingsRow :label="t('automation.action')" control="field">
                <div class="automation-control">
                  <el-select v-model="action.type" :disabled="disabled">
                    <el-option
                      v-for="opt in actionOptions"
                      :key="opt.value"
                      :value="opt.value"
                      :label="opt.label"
                      :disabled="opt.disabled"
                    />
                  </el-select>
                </div>
              </SettingsRow>
              <SettingsRow
                v-if="action.type === 'change_status'"
                :label="t('automation.status')"
                control="field"
              >
                <div class="automation-control">
                  <el-select v-model="action.status" :disabled="disabled">
                    <el-option
                      v-for="opt in statusOptions"
                      :key="opt.value"
                      :value="opt.value"
                      :label="opt.label"
                    />
                  </el-select>
                </div>
              </SettingsRow>
              <template v-if="action.type === 'set_power_plan'">
                <SettingsRow
                  :label="t('automation.powerPlanMode')"
                  control="field"
                >
                  <div class="automation-control">
                    <el-radio-group
                      v-model="action.powerPlanMode"
                      :disabled="disabled"
                    >
                      <el-radio value="preset">{{
                        t("automation.powerPlanPreset")
                      }}</el-radio>
                      <el-radio
                        value="guid"
                        :disabled="disabled || !powerPlans.length"
                        >{{ t("automation.powerPlanDetected") }}</el-radio
                      >
                    </el-radio-group>
                  </div>
                </SettingsRow>
                <SettingsRow
                  v-if="action.powerPlanMode === 'preset'"
                  :label="t('automation.powerPlanPreset')"
                  control="field"
                >
                  <div class="automation-control">
                    <el-select
                      v-model="action.powerPlanPreset"
                      :disabled="disabled"
                    >
                      <el-option
                        v-for="p in powerPlanPresets"
                        :key="p.value"
                        :value="p.value"
                        :label="p.label"
                      />
                    </el-select>
                  </div>
                </SettingsRow>
                <SettingsRow
                  v-else
                  :label="t('automation.powerPlanDetected')"
                  control="field"
                >
                  <div class="automation-control">
                    <el-select
                      v-model="action.powerPlanGuid"
                      :disabled="disabled"
                    >
                      <el-option
                        v-for="p in powerPlans"
                        :key="p.guid"
                        :value="p.guid"
                        :label="p.name"
                      />
                    </el-select>
                  </div>
                </SettingsRow>
              </template>
              <template v-if="action.type === 'set_vrchat_window_size'">
                <SettingsRow
                  :label="t('automation.windowWidth')"
                  control="number"
                >
                  <el-input-number
                    v-model="action.windowWidth"
                    :min="1"
                    :max="7680"
                    :disabled="disabled"
                    data-testid="window-width-input"
                  />
                </SettingsRow>
                <SettingsRow
                  :label="t('automation.windowHeight')"
                  control="number"
                >
                  <el-input-number
                    v-model="action.windowHeight"
                    :min="1"
                    :max="4320"
                    :disabled="disabled"
                    data-testid="window-height-input"
                  />
                </SettingsRow>
              </template>
              <SettingsRow
                :label="t('automation.continueOnError')"
                control="switch"
              >
                <el-switch
                  v-model="action.continueOnError"
                  :disabled="disabled"
                />
              </SettingsRow>
            </div>
            <div
              v-if="editor.actions.length < 10"
              class="automation-add-action"
            >
              <VtButton
                variant="secondary"
                class="btn-add-action"
                :disabled="disabled"
                @click="emit('addAction')"
              >
                {{ t("automation.addAction") }}
              </VtButton>
            </div>
            <p class="text-body-sm partial-hint">
              {{ t("automation.partialApplyHint") }}
            </p>
          </el-card>
        </template>

        <el-card
          v-else
          class="section-card settings-card--rows automation-card"
          shadow="never"
        >
          <template #header>
            <span class="automation-card__title">{{
              t("automation.scriptSource")
            }}</span>
          </template>
          <div class="automation-script">
            <SettingsRow :label="t('automation.scriptSource')" control="field">
              <el-input
                v-model="state.scriptSource"
                type="textarea"
                :rows="14"
                class="script-editor"
                :disabled="disabled"
                :placeholder="t('automation.scriptPlaceholder')"
              />
            </SettingsRow>
          </div>
        </el-card>
      </div>

      <div class="automation-action-bar" data-testid="automation-action-bar">
        <div class="automation-action-bar__start">
          <VtButton
            v-if="canDeleteItem"
            class="automation-action-bar__delete"
            variant="danger"
            plain
            data-testid="delete-item"
            :disabled="disabled"
            @click="emit('delete')"
          >
            {{ t("automation.delete") }}
          </VtButton>
        </div>
        <p class="automation-action-bar__unsaved">
          <span v-if="isDirty" data-testid="unsaved-banner">{{
            t("automation.unsavedBanner")
          }}</span>
        </p>
        <div class="automation-action-bar__primary">
          <VtButton
            variant="secondary"
            data-testid="cancel-item"
            :disabled="disabled"
            @click="emit('cancel')"
          >
            {{ t("automation.cancel") }}
          </VtButton>
          <VtButton
            variant="primary"
            data-testid="save-item"
            :loading="saving"
            :disabled="disabled"
            @click="emit('save')"
          >
            {{ t("automation.save") }}
          </VtButton>
        </div>
      </div>
    </el-form>
  </div>
</template>

<style scoped>
.automation-editor {
  display: flex;
  flex-direction: column;
  min-height: 0;
  height: 100%;
  overflow: hidden;
  container-type: inline-size;
  container-name: setting-rows;
}

.automation-editor__scroll {
  flex: 1 1 auto;
  min-height: 0;
  min-width: 0;
  overflow: auto;
  scrollbar-gutter: stable;
  display: flex;
  flex-direction: column;
}

.automation-back {
  display: none;
  margin: 0 0 var(--space-block);
  align-self: flex-start;
}

.automation-editor-head {
  display: flex;
  align-items: center;
  gap: var(--space-action-group);
  height: 48px;
  min-height: 48px;
  min-width: 0;
  margin-bottom: var(--space-block);
}

.automation-editor-head__name {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-14);
  font-weight: var(--font-weight-600);
  line-height: var(--line-height-tight);
  color: var(--color-text-primary);
}

.automation-editor-head__name.is-empty {
  color: var(--color-text-muted);
  font-weight: var(--font-weight-400);
}

.automation-editor__cards {
  display: flex;
  flex-direction: column;
  gap: var(--space-block);
  flex: 1 0 auto;
}

.automation-card {
  margin-bottom: 0;
}

.automation-card :deep(.el-card__header) {
  height: 48px;
  min-height: 48px;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  gap: var(--space-action-group);
  padding: 0 var(--space-block);
  overflow: hidden;
}

.automation-card :deep(.el-card__body) {
  padding: 0;
}

.automation-card__title {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.automation-header-more {
  display: none;
  flex-shrink: 0;
}

.unsaved-dot {
  width: var(--space-action-group);
  height: var(--space-action-group);
  border-radius: 50%;
  background: var(--color-warning);
  flex-shrink: 0;
}

.automation-control {
  width: 100%;
  min-width: 0;
  display: flex;
  flex-direction: column;
  align-items: stretch;
}

.automation-field-error {
  margin: var(--space-4) 0 0;
  color: var(--color-danger);
  font-size: var(--font-size-12);
  line-height: var(--line-height-normal);
  overflow-wrap: anywhere;
  white-space: normal;
}

.automation-editor :deep(.el-checkbox-group),
.automation-editor :deep(.el-radio-group) {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-action-group);
  width: 100%;
  max-width: 100%;
  justify-content: flex-start;
}

.automation-action-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: var(--space-action-group);
  min-height: 40px;
  box-sizing: border-box;
  padding: 0 var(--space-block);
  background: var(--color-bg-muted);
  border-bottom: 1px solid var(--color-border);
  min-width: 0;
}

.automation-action-head__title {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-14);
  font-weight: var(--font-weight-600);
  line-height: var(--line-height-tight);
}

.automation-action-head__tools {
  display: flex;
  align-items: center;
  gap: var(--space-inline-tight);
  flex-shrink: 0;
}

.automation-action-head__tools :deep(.el-button) {
  min-width: 0;
  margin-left: 0;
  color: var(--color-text-primary);
}

.automation-time {
  display: flex;
  align-items: center;
  flex-wrap: nowrap;
  gap: var(--space-action-group);
  max-width: 100%;
}

.automation-time :deep(.el-input-number) {
  width: 120px;
  flex: 0 0 120px;
  max-width: 120px;
}

.automation-time__sep {
  color: var(--color-text-secondary);
}

.automation-add-action {
  padding: var(--space-block);
}

.btn-add-action {
  width: 100%;
  border-style: dashed !important;
}

.partial-hint {
  display: block;
  margin: 0;
  padding: 0 var(--space-block) var(--space-block);
  color: var(--color-text-secondary);
  overflow-wrap: anywhere;
}

.automation-script :deep(.setting-row) {
  flex-direction: column;
  align-items: stretch;
}

.automation-script :deep(.setting-row__control) {
  flex: 1 1 auto;
  width: 100%;
  max-width: none;
  justify-content: stretch;
}

.script-editor :deep(textarea) {
  font-family: ui-monospace, monospace;
  font-size: var(--font-size-12);
  overflow-wrap: anywhere;
}

.automation-action-bar {
  position: sticky;
  bottom: 0;
  z-index: 2;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
  align-items: center;
  gap: var(--space-action-group);
  margin-top: auto;
  padding: var(--space-form-field) var(--space-block);
  background: var(--color-bg-elevated);
  border-top: 1px solid var(--color-border);
}

.automation-action-bar__delete {
  justify-self: start;
}

.automation-action-bar__unsaved {
  margin: 0;
  justify-self: center;
  text-align: center;
  color: var(--color-warning);
  font-size: var(--font-size-12);
  line-height: var(--line-height-normal);
  overflow-wrap: anywhere;
}

.automation-action-bar__primary {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-action-group);
  justify-self: end;
}

.automation-action-bar :deep(.el-button) {
  min-width: var(--button-min-width);
  margin-left: 0;
}

.automation-action-bar :deep(.el-button + .el-button) {
  margin-left: 0;
}

@container setting-rows (max-width: 559px) {
  .automation-action-bar__delete {
    display: none;
  }

  .automation-header-more {
    display: inline-flex;
  }
}
</style>
