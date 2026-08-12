<template>
  <div class="launcher-view">
    <h1 class="page-title">{{ t("routes.launcher") }}</h1>
    <div class="launcher-layout" :class="{ 'sidebar-collapsed': !sidebarOpen }">
      <aside class="profiles-sidebar">
        <div class="sidebar-toolbar">
          <VtButton
            variant="tertiary"
            class="sidebar-toggle"
            data-testid="sidebar-toggle-btn"
            :aria-label="t('launcher.toggleSidebar')"
            @click="toggleSidebar"
          >
            <VtIcon size="default"
              ><component :is="sidebarOpen ? Fold : Expand"
            /></VtIcon>
          </VtButton>
        </div>
        <VtButton
          v-show="sidebarOpen"
          variant="secondary"
          class="btn-add"
          @click="requestAddNew"
        >
          {{ t("launcher.newProfile") }}
        </VtButton>
        <div v-show="sidebarOpen" class="profiles-list">
          <div
            v-for="p in profiles"
            :key="p.id"
            class="profile-card"
            :class="{ active: selected?.id === p.id }"
            :data-testid="`profile-card-${p.id}`"
            @click="requestSelect(p)"
          >
            <span class="profile-name">{{ p.name }}</span>
            <span
              v-if="isProfileDirtyInSidebar(p)"
              class="unsaved-dot"
              data-testid="unsaved-dot"
              :title="t('launcher.unsavedBanner')"
            />
            <VtTag v-if="p.isDefault" variant="primary" size="small">{{
              t("launcher.defaultTag")
            }}</VtTag>
          </div>
        </div>
      </aside>

      <div v-if="selected" class="profile-editor">
        <div class="editor-toolbar">
          <VtAlert
            v-if="isDirty"
            class="unsaved-banner"
            data-testid="unsaved-banner"
            variant="warning"
            :title="t('launcher.unsavedBanner')"
          />
          <div class="toolbar-actions">
            <VtButton variant="secondary" class="btn-launch" @click="launch">
              {{ t("launcher.launchWithThis") }}
            </VtButton>
            <VtButton variant="primary" class="btn-save" @click="save">
              {{ t("launcher.save") }}
            </VtButton>
            <VtButton
              variant="secondary"
              data-testid="duplicate-profile-btn"
              @click="requestDuplicate"
            >
              {{ t("launcher.duplicate") }}
            </VtButton>
            <VtButton
              variant="danger"
              plain
              data-testid="delete-profile-btn"
              @click="confirmDelete"
            >
              {{ t("launcher.delete") }}
            </VtButton>
          </div>
        </div>

        <el-form label-position="top" size="default">
          <el-form-item :label="t('launcher.profileName')">
            <VtInput v-model="selected.name" />
          </el-form-item>

          <el-form-item>
            <VtCheckbox v-model="selected.isDefault">
              {{ t("launcher.setAsDefault") }}
            </VtCheckbox>
          </el-form-item>

          <el-form-item :label="t('launcher.launchArgs')">
            <div class="launch-args-gui">
              <div class="arg-row">
                <VtCheckbox
                  v-model="launchArgs.noVr"
                  data-testid="no-vr-checkbox"
                >
                  {{ t("launcher.desktopMode") }}
                </VtCheckbox>
              </div>

              <el-form-item
                :label="t('launcher.screenMode')"
                class="nested-form-item"
              >
                <el-radio-group
                  v-model="launchArgs.screenMode"
                  :aria-label="t('launcher.screenMode')"
                  size="default"
                >
                  <el-radio-button
                    value="fullscreen"
                    data-testid="screen-mode-fullscreen"
                    >{{ t("launcher.screenModeFullscreen") }}</el-radio-button
                  >
                  <el-radio-button
                    value="windowed"
                    data-testid="screen-mode-windowed"
                    >{{ t("launcher.screenModeWindowed") }}</el-radio-button
                  >
                  <el-radio-button
                    value="popupwindow"
                    data-testid="screen-mode-popupwindow"
                    >{{ t("launcher.screenModePopup") }}</el-radio-button
                  >
                </el-radio-group>
              </el-form-item>

              <el-form-item :label="t('launcher.customArgs')">
                <VtInput
                  v-model="launchArgs.custom"
                  placeholder="-batchmode"
                  data-testid="custom-args-input"
                />
              </el-form-item>

              <el-collapse
                v-model="advancedCollapseActive"
                class="args-collapse"
              >
                <el-collapse-item
                  :title="t('launcher.allOptions')"
                  name="advanced"
                >
                  <div class="advanced-section">
                    <h3 class="advanced-section-title">
                      {{ t("launcher.advancedDisplayPerformance") }}
                    </h3>
                    <div class="launch-args-advanced">
                      <div class="arg-row">
                        <VtCheckbox
                          v-model="valueOptionsEnabled.resolution"
                          data-testid="resolution-enabled-checkbox"
                          @change="onResolutionEnabledChange"
                        >
                          {{ t("launcher.resolutionHint") }}
                        </VtCheckbox>
                      </div>
                      <div
                        v-if="valueOptionsEnabled.resolution"
                        class="sub-options"
                      >
                        <el-form-item
                          :label="t('launcher.preset')"
                          class="nested-form-item"
                        >
                          <el-radio-group
                            v-model="resolutionPreset"
                            :aria-label="t('launcher.preset')"
                            size="small"
                            @change="applyResolutionPreset"
                          >
                            <el-radio-button
                              value="HD"
                              data-testid="resolution-preset-hd"
                              >HD</el-radio-button
                            >
                            <el-radio-button
                              value="FHD"
                              data-testid="resolution-preset-fhd"
                              >FHD</el-radio-button
                            >
                            <el-radio-button
                              value="WQHD"
                              data-testid="resolution-preset-wqhd"
                              >WQHD</el-radio-button
                            >
                            <el-radio-button
                              value="4K"
                              data-testid="resolution-preset-4k"
                              >4K</el-radio-button
                            >
                            <el-radio-button
                              value="custom"
                              data-testid="resolution-preset-custom"
                              >{{ t("launcher.manual") }}</el-radio-button
                            >
                          </el-radio-group>
                        </el-form-item>
                        <div class="resolution-fields">
                          <el-input-number
                            v-model="launchArgs.screenWidth"
                            :min="1280"
                            :max="7680"
                            :disabled="resolutionPreset !== 'custom'"
                            data-testid="screen-width-input"
                            size="small"
                            :placeholder="t('launcher.widthPh')"
                            style="width: 120px"
                          />
                          <span class="resolution-sep">×</span>
                          <el-input-number
                            v-model="launchArgs.screenHeight"
                            :min="720"
                            :max="4320"
                            :disabled="resolutionPreset !== 'custom'"
                            data-testid="screen-height-input"
                            size="small"
                            :placeholder="t('launcher.heightPh')"
                            style="width: 120px"
                          />
                        </div>
                      </div>

                      <div class="arg-row">
                        <VtCheckbox
                          v-model="valueOptionsEnabled.monitor"
                          data-testid="monitor-enabled-checkbox"
                          @change="onMonitorEnabledChange"
                        >
                          {{ t("launcher.monitorHint") }}
                        </VtCheckbox>
                      </div>
                      <div
                        v-if="valueOptionsEnabled.monitor"
                        class="sub-options"
                      >
                        <el-input-number
                          v-model="launchArgs.monitor"
                          :min="1"
                          data-testid="monitor-input"
                          size="small"
                          :placeholder="t('launcher.monitorPh')"
                          style="width: 120px"
                        />
                      </div>

                      <div class="arg-row">
                        <VtCheckbox
                          v-model="valueOptionsEnabled.fps"
                          data-testid="fps-enabled-checkbox"
                          @change="onFpsEnabledChange"
                        >
                          {{ t("launcher.fpsHint") }}
                        </VtCheckbox>
                      </div>
                      <div v-if="valueOptionsEnabled.fps" class="sub-options">
                        <el-input-number
                          v-model="launchArgs.fps"
                          :min="1"
                          data-testid="fps-input"
                          size="small"
                          placeholder="90"
                          style="width: 120px"
                        />
                      </div>

                      <div class="arg-row">
                        <VtCheckbox
                          v-model="launchArgs.skipRegistry"
                          data-testid="skip-registry-checkbox"
                        >
                          {{ t("launcher.skipRegistry") }}
                        </VtCheckbox>
                      </div>

                      <div class="arg-row">
                        <VtCheckbox
                          v-model="valueOptionsEnabled.processPriority"
                          data-testid="process-priority-enabled-checkbox"
                          @change="onProcessPriorityEnabledChange"
                        >
                          {{ t("launcher.processPriority") }}
                        </VtCheckbox>
                      </div>
                      <div
                        v-if="valueOptionsEnabled.processPriority"
                        class="sub-options"
                      >
                        <el-input-number
                          v-model="launchArgs.processPriority"
                          :min="-2"
                          :max="2"
                          data-testid="process-priority-input"
                          size="small"
                          placeholder="-2～2"
                          style="width: 120px"
                        />
                      </div>

                      <div class="arg-row">
                        <VtCheckbox
                          v-model="valueOptionsEnabled.mainThreadPriority"
                          data-testid="main-thread-priority-enabled-checkbox"
                          @change="onMainThreadPriorityEnabledChange"
                        >
                          {{ t("launcher.mainThreadPriority") }}
                        </VtCheckbox>
                      </div>
                      <div
                        v-if="valueOptionsEnabled.mainThreadPriority"
                        class="sub-options"
                      >
                        <el-input-number
                          v-model="launchArgs.mainThreadPriority"
                          :min="-2"
                          :max="2"
                          data-testid="main-thread-priority-input"
                          size="small"
                          placeholder="-2～2"
                          style="width: 120px"
                        />
                      </div>

                      <div class="arg-row">
                        <VtCheckbox
                          v-model="valueOptionsEnabled.profile"
                          data-testid="profile-enabled-checkbox"
                          @change="onProfileEnabledChange"
                        >
                          {{ t("launcher.profileHint") }}
                        </VtCheckbox>
                      </div>
                      <div
                        v-if="valueOptionsEnabled.profile"
                        class="sub-options"
                      >
                        <el-input-number
                          v-model="launchArgs.profile"
                          :min="0"
                          data-testid="profile-input"
                          size="small"
                          :placeholder="t('launcher.profilePh')"
                          style="width: 120px"
                        />
                      </div>
                    </div>

                    <h3
                      class="advanced-section-title"
                      data-testid="advanced-debug-section"
                    >
                      {{ t("launcher.advancedDebugExpert") }}
                    </h3>
                    <div class="launch-args-advanced">
                      <div class="arg-row">
                        <VtCheckbox
                          v-model="launchArgs.enableDebugGui"
                          data-testid="enable-debug-gui-checkbox"
                        >
                          {{ t("launcher.debugGui") }}
                        </VtCheckbox>
                      </div>
                      <div class="arg-row">
                        <VtCheckbox
                          v-model="launchArgs.enableSDKLogLevels"
                          data-testid="enable-sdk-log-levels-checkbox"
                        >
                          {{ t("launcher.sdkLog") }}
                        </VtCheckbox>
                      </div>
                      <div class="arg-row">
                        <VtCheckbox
                          v-model="launchArgs.enableUdonDebugLogging"
                          data-testid="enable-udon-debug-logging-checkbox"
                        >
                          {{ t("launcher.udonDebug") }}
                        </VtCheckbox>
                      </div>
                      <div class="arg-row">
                        <VtCheckbox
                          v-model="launchArgs.watchWorlds"
                          data-testid="watch-worlds-checkbox"
                        >
                          {{ t("launcher.watchWorlds") }}
                        </VtCheckbox>
                      </div>
                      <div class="arg-row">
                        <VtCheckbox
                          v-model="launchArgs.watchAvatars"
                          data-testid="watch-avatars-checkbox"
                        >
                          {{ t("launcher.watchAvatars") }}
                        </VtCheckbox>
                      </div>
                      <div class="arg-row">
                        <VtCheckbox
                          v-model="launchArgs.enforceWorldServerChecks"
                          data-testid="enforce-world-server-checks-checkbox"
                        >
                          {{ t("launcher.enforceWorldServer") }}
                        </VtCheckbox>
                      </div>

                      <div class="arg-row">
                        <VtCheckbox
                          v-model="valueOptionsEnabled.midi"
                          data-testid="midi-enabled-checkbox"
                          @change="onMidiEnabledChange"
                        >
                          {{ t("launcher.midi") }}
                        </VtCheckbox>
                      </div>
                      <div v-if="valueOptionsEnabled.midi" class="sub-options">
                        <VtInput
                          v-model="launchArgs.midi"
                          :placeholder="t('launcher.midiPh')"
                          data-testid="midi-input"
                          size="small"
                          style="max-width: 240px"
                        />
                      </div>

                      <div class="arg-row">
                        <VtCheckbox
                          v-model="valueOptionsEnabled.ignoreTrackers"
                          data-testid="ignore-trackers-enabled-checkbox"
                          @change="onIgnoreTrackersEnabledChange"
                        >
                          {{ t("launcher.ignoreTrackers") }}
                        </VtCheckbox>
                      </div>
                      <div
                        v-if="valueOptionsEnabled.ignoreTrackers"
                        class="sub-options"
                      >
                        <VtInput
                          v-model="launchArgs.ignoreTrackers"
                          placeholder="serial1,serial2"
                          data-testid="ignore-trackers-input"
                          size="small"
                          style="max-width: 240px"
                        />
                      </div>

                      <el-form-item
                        :label="t('launcher.videoDecoding')"
                        class="nested-form-item"
                      >
                        <el-radio-group
                          v-model="launchArgs.videoDecoding"
                          :aria-label="t('launcher.videoDecoding')"
                          size="small"
                        >
                          <el-radio-button
                            value=""
                            data-testid="video-decoding-default"
                            >{{
                              t("launcher.videoDecDefault")
                            }}</el-radio-button
                          >
                          <el-radio-button
                            value="software"
                            data-testid="video-decoding-software"
                            >{{
                              t("launcher.videoDecSoftware")
                            }}</el-radio-button
                          >
                          <el-radio-button
                            value="hardware"
                            data-testid="video-decoding-hardware"
                            >{{
                              t("launcher.videoDecHardware")
                            }}</el-radio-button
                          >
                        </el-radio-group>
                      </el-form-item>

                      <div class="arg-row">
                        <VtCheckbox
                          v-model="launchArgs.disableAMDStutterWorkaround"
                          data-testid="disable-amd-stutter-workaround-checkbox"
                        >
                          {{ t("launcher.disableAmdStutter") }}
                        </VtCheckbox>
                      </div>

                      <div class="arg-row">
                        <VtCheckbox
                          v-model="valueOptionsEnabled.osc"
                          data-testid="osc-enabled-checkbox"
                          @change="onOscEnabledChange"
                        >
                          {{ t("launcher.osc") }}
                        </VtCheckbox>
                      </div>
                      <div v-if="valueOptionsEnabled.osc" class="sub-options">
                        <VtInput
                          v-model="launchArgs.osc"
                          :placeholder="t('launcher.oscPh')"
                          data-testid="osc-input"
                          size="small"
                          style="max-width: 240px"
                        />
                      </div>

                      <div class="arg-row">
                        <VtCheckbox
                          v-model="valueOptionsEnabled.affinity"
                          data-testid="affinity-enabled-checkbox"
                          @change="onAffinityEnabledChange"
                        >
                          {{ t("launcher.affinity") }}
                        </VtCheckbox>
                      </div>
                      <div
                        v-if="valueOptionsEnabled.affinity"
                        class="sub-options"
                      >
                        <VtInput
                          v-model="launchArgs.affinity"
                          :placeholder="t('launcher.affinityPh')"
                          data-testid="affinity-input"
                          size="small"
                          style="max-width: 200px"
                        />
                      </div>
                    </div>

                    <h3
                      class="advanced-section-title"
                      data-testid="advanced-ik-section"
                    >
                      {{ t("launcher.advancedIk20") }}
                    </h3>
                    <div class="launch-args-advanced">
                      <div class="arg-row">
                        <VtCheckbox
                          v-model="valueOptionsEnabled.customArmRatio"
                          data-testid="custom-arm-ratio-enabled-checkbox"
                          @change="onCustomArmRatioEnabledChange"
                        >
                          {{ t("launcher.customArmRatio") }}
                        </VtCheckbox>
                      </div>
                      <div
                        v-if="valueOptionsEnabled.customArmRatio"
                        class="sub-options"
                      >
                        <el-input-number
                          v-model="launchArgs.customArmRatio"
                          :min="0.0001"
                          :step="0.0001"
                          :precision="4"
                          data-testid="custom-arm-ratio-input"
                          size="small"
                          controls-position="right"
                        />
                      </div>

                      <div class="arg-row">
                        <VtCheckbox
                          v-model="launchArgs.disableShoulderTracking"
                          data-testid="disable-shoulder-tracking-checkbox"
                        >
                          {{ t("launcher.disableShoulderTracking") }}
                        </VtCheckbox>
                      </div>

                      <div class="arg-row">
                        <VtCheckbox
                          v-model="launchArgs.enableIkDebugLogging"
                          data-testid="enable-ik-debug-logging-checkbox"
                        >
                          {{ t("launcher.enableIkDebugLogging") }}
                        </VtCheckbox>
                      </div>

                      <div class="arg-row">
                        <VtCheckbox
                          v-model="valueOptionsEnabled.calibrationRange"
                          data-testid="calibration-range-enabled-checkbox"
                          @change="onCalibrationRangeEnabledChange"
                        >
                          {{ t("launcher.calibrationRange") }}
                        </VtCheckbox>
                      </div>
                      <div
                        v-if="valueOptionsEnabled.calibrationRange"
                        class="sub-options"
                      >
                        <el-input-number
                          v-model="launchArgs.calibrationRange"
                          :min="0.0001"
                          :step="0.1"
                          :precision="2"
                          data-testid="calibration-range-input"
                          size="small"
                          controls-position="right"
                        />
                      </div>

                      <div class="arg-row">
                        <VtCheckbox
                          v-model="launchArgs.freezeTrackingOnDisconnect"
                          data-testid="freeze-tracking-on-disconnect-checkbox"
                        >
                          {{ t("launcher.freezeTrackingOnDisconnect") }}
                        </VtCheckbox>
                      </div>
                    </div>
                  </div>
                </el-collapse-item>
              </el-collapse>
            </div>
          </el-form-item>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount } from "vue";
import { onBeforeRouteLeave } from "vue-router";
import { useI18n } from "vue-i18n";
import { ElMessageBox } from "element-plus";
import { Expand, Fold } from "@element-plus/icons-vue";
import VtAlert from "../components/VtAlert.vue";
import VtButton from "../components/VtButton.vue";
import VtCheckbox from "../components/VtCheckbox.vue";
import VtIcon from "../components/VtIcon.vue";
import VtInput from "../components/VtInput.vue";
import VtTag from "../components/VtTag.vue";
import {
  App,
  type LaunchProfileDTO,
  type LaunchArgsParsedDTO,
  PRIORITY_OMIT,
} from "../wails/app";
import {
  defaultValueOptionsEnabled,
  hasAdvancedLaunchOptionsEnabled,
  isLaunchProfileEditDirty,
  nextDefaultLaunchProfileName,
  readSidebarOpenPreference,
  syncValueOptionsEnabled,
  writeSidebarOpenPreference,
  IK_CUSTOM_ARM_RATIO_DEFAULT,
  IK_CALIBRATION_RANGE_DEFAULT,
  type LaunchProfileEditSnapshot,
  type ValueOptionsEnabled,
} from "./launcher/launcherProfileEdits";
import { formatError } from "../utils/formatError";
import { showToast } from "../utils/showToast";
import {
  VT_BUTTON_DANGER_CONFIRM_CLASS,
  VT_BUTTON_SECONDARY_CANCEL_CLASS,
} from "../components/vtButtonClasses";

const { t } = useI18n();

type ResolutionPreset = "HD" | "FHD" | "WQHD" | "4K" | "custom";

interface PresetResolution {
  width: number;
  height: number;
}

const LAUNCHER_RESOLUTION_PRESETS: Record<string, PresetResolution> = {
  HD: { width: 1280, height: 720 },
  FHD: { width: 1920, height: 1080 },
  WQHD: { width: 2560, height: 1440 },
  "4K": { width: 3840, height: 2160 },
};

const defaultLaunchArgs = (): LaunchArgsParsedDTO => ({
  noVr: false,
  screenMode: "",
  screenWidth: 0,
  screenHeight: 0,
  fps: 90,
  skipRegistry: false,
  processPriority: PRIORITY_OMIT,
  mainThreadPriority: PRIORITY_OMIT,
  monitor: 0,
  profile: -1,
  enableDebugGui: false,
  enableSDKLogLevels: false,
  enableUdonDebugLogging: false,
  midi: "",
  watchWorlds: false,
  watchAvatars: false,
  ignoreTrackers: "",
  videoDecoding: "",
  disableAMDStutterWorkaround: false,
  osc: "",
  affinity: "",
  enforceWorldServerChecks: false,
  customArmRatio: 0,
  disableShoulderTracking: false,
  enableIkDebugLogging: false,
  calibrationRange: 0,
  freezeTrackingOnDisconnect: false,
  custom: "",
});

const profiles = ref<LaunchProfileDTO[]>([]);
const selected = ref<LaunchProfileDTO | null>(null);
const launchArgs = ref<LaunchArgsParsedDTO>(defaultLaunchArgs());
const resolutionPreset = ref<ResolutionPreset>("FHD");
const valueOptionsEnabled = reactive<ValueOptionsEnabled>(
  defaultValueOptionsEnabled(),
);
const savedSnapshot = ref<LaunchProfileEditSnapshot | null>(null);
const sidebarOpen = ref(readSidebarOpenPreference());
const advancedCollapseActive = ref<string[]>([]);
/** Bumped on state-changing save/create; skip post-await updates after unmount. */
let profileSaveGen = 0;

function showSaveError(e: unknown) {
  showToast.error(formatError(e, t("launcher.errSave")));
}
const currentSnapshot = computed((): LaunchProfileEditSnapshot | null => {
  if (!selected.value) return null;
  return {
    profileId: selected.value.id,
    name: selected.value.name,
    isDefault: selected.value.isDefault,
    launchArgs: { ...launchArgs.value },
    valueOptionsEnabled: { ...valueOptionsEnabled },
  };
});

const isDirty = computed(() =>
  isLaunchProfileEditDirty(savedSnapshot.value, currentSnapshot.value),
);

function isProfileDirtyInSidebar(p: LaunchProfileDTO): boolean {
  if (!isDirty.value || !selected.value) return false;
  return p.id === selected.value.id;
}

function syncAdvancedCollapseOpenState() {
  advancedCollapseActive.value = hasAdvancedLaunchOptionsEnabled(
    launchArgs.value,
    valueOptionsEnabled,
  )
    ? ["advanced"]
    : [];
}

function captureSnapshot() {
  if (!currentSnapshot.value) {
    savedSnapshot.value = null;
    return;
  }
  savedSnapshot.value = {
    ...currentSnapshot.value,
    launchArgs: { ...currentSnapshot.value.launchArgs },
    valueOptionsEnabled: { ...currentSnapshot.value.valueOptionsEnabled },
  };
}

function toggleSidebar() {
  sidebarOpen.value = !sidebarOpen.value;
  writeSidebarOpenPreference(sidebarOpen.value);
}

type UnsavedChoice = "save" | "discard" | "cancel";

async function promptUnsavedChoice(): Promise<UnsavedChoice> {
  try {
    await ElMessageBox.confirm(
      t("launcher.unsavedSwitchMessage"),
      t("launcher.unsavedSwitchTitle"),
      {
        confirmButtonText: t("launcher.saveAndContinue"),
        cancelButtonText: t("launcher.discardAndContinue"),
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

async function guardUnsavedEdits(): Promise<boolean> {
  if (!isDirty.value) return true;
  const choice = await promptUnsavedChoice();
  if (choice === "cancel") return false;
  if (choice === "save") return await save();
  return true;
}

function detectResolutionPreset(
  width: number,
  height: number,
): ResolutionPreset {
  for (const [key, val] of Object.entries(LAUNCHER_RESOLUTION_PRESETS)) {
    if (val.width === width && val.height === height) {
      return key as ResolutionPreset;
    }
  }
  return "custom";
}

function syncResolutionPresetFromArgs() {
  if (!valueOptionsEnabled.resolution) return;
  resolutionPreset.value = detectResolutionPreset(
    launchArgs.value.screenWidth,
    launchArgs.value.screenHeight,
  );
}

function applyResolutionPreset() {
  const preset = LAUNCHER_RESOLUTION_PRESETS[resolutionPreset.value];
  if (preset) {
    launchArgs.value.screenWidth = preset.width;
    launchArgs.value.screenHeight = preset.height;
  }
}

function onResolutionEnabledChange() {
  if (valueOptionsEnabled.resolution) {
    if (
      launchArgs.value.screenWidth <= 0 &&
      launchArgs.value.screenHeight <= 0
    ) {
      launchArgs.value.screenWidth = 1920;
      launchArgs.value.screenHeight = 1080;
    }
    if (launchArgs.value.screenHeight <= 0) {
      launchArgs.value.screenHeight = 1080;
    }
    syncResolutionPresetFromArgs();
  } else {
    launchArgs.value.screenWidth = 0;
    launchArgs.value.screenHeight = 0;
  }
}

function onMonitorEnabledChange() {
  if (!valueOptionsEnabled.monitor) launchArgs.value.monitor = 0;
  else if (launchArgs.value.monitor < 1) launchArgs.value.monitor = 1;
}

function onFpsEnabledChange() {
  if (!valueOptionsEnabled.fps) launchArgs.value.fps = 0;
  else if (launchArgs.value.fps <= 0) launchArgs.value.fps = 90;
}

function onProcessPriorityEnabledChange() {
  if (!valueOptionsEnabled.processPriority)
    launchArgs.value.processPriority = PRIORITY_OMIT;
  else if (launchArgs.value.processPriority === PRIORITY_OMIT)
    launchArgs.value.processPriority = 0;
}

function onMainThreadPriorityEnabledChange() {
  if (!valueOptionsEnabled.mainThreadPriority)
    launchArgs.value.mainThreadPriority = PRIORITY_OMIT;
  else if (launchArgs.value.mainThreadPriority === PRIORITY_OMIT)
    launchArgs.value.mainThreadPriority = 0;
}

function onProfileEnabledChange() {
  if (!valueOptionsEnabled.profile) launchArgs.value.profile = -1;
  else if (launchArgs.value.profile < 0) launchArgs.value.profile = 0;
}

function onMidiEnabledChange() {
  if (!valueOptionsEnabled.midi) launchArgs.value.midi = "";
}

function onIgnoreTrackersEnabledChange() {
  if (!valueOptionsEnabled.ignoreTrackers) launchArgs.value.ignoreTrackers = "";
}

function onOscEnabledChange() {
  if (!valueOptionsEnabled.osc) launchArgs.value.osc = "";
}

function onAffinityEnabledChange() {
  if (!valueOptionsEnabled.affinity) launchArgs.value.affinity = "";
}

function onCustomArmRatioEnabledChange() {
  if (!valueOptionsEnabled.customArmRatio) launchArgs.value.customArmRatio = 0;
  else if (
    !Number.isFinite(launchArgs.value.customArmRatio) ||
    !(launchArgs.value.customArmRatio > 0)
  )
    launchArgs.value.customArmRatio = IK_CUSTOM_ARM_RATIO_DEFAULT;
}

function onCalibrationRangeEnabledChange() {
  if (!valueOptionsEnabled.calibrationRange)
    launchArgs.value.calibrationRange = 0;
  else if (
    !Number.isFinite(launchArgs.value.calibrationRange) ||
    !(launchArgs.value.calibrationRange > 0)
  )
    launchArgs.value.calibrationRange = IK_CALIBRATION_RANGE_DEFAULT;
}

async function syncLaunchArgsFromProfile(p: LaunchProfileDTO) {
  launchArgs.value = await App.parseLaunchArgsForGUI(p.arguments);
  Object.assign(valueOptionsEnabled, syncValueOptionsEnabled(launchArgs.value));
  syncResolutionPresetFromArgs();
  syncAdvancedCollapseOpenState();
}

async function openProfile(p: LaunchProfileDTO) {
  savedSnapshot.value = null;
  selected.value = { ...p };
  await syncLaunchArgsFromProfile(p);
  captureSnapshot();
}

onMounted(async () => {
  profiles.value = await App.launchProfiles();
  if (profiles.value.length > 0 && !selected.value) {
    const p = profiles.value.find((p) => p.isDefault) ?? profiles.value[0];
    await openProfile(p);
  }
});

onBeforeUnmount(() => {
  profileSaveGen += 1;
});

async function requestSelect(p: LaunchProfileDTO) {
  if (selected.value?.id === p.id && selected.value?.name === p.name) return;
  const ok = await guardUnsavedEdits();
  if (!ok) return;
  await openProfile(p);
}

async function requestAddNew() {
  const previousId = selected.value?.id ?? "";
  const ok = await guardUnsavedEdits();
  if (!ok) return;

  const beforeIds = new Set(profiles.value.map((p) => p.id));
  const name = nextDefaultLaunchProfileName(
    t("launcher.newProfileDefaultName"),
    profiles.value.map((p) => p.name),
  );
  const gen = ++profileSaveGen;
  try {
    await App.saveLaunchProfile({
      id: "",
      name,
      arguments: "",
      isDefault: profiles.value.length === 0,
    });
    if (gen !== profileSaveGen) return;
    profiles.value = await App.launchProfiles();
    if (gen !== profileSaveGen) return;
    const created =
      profiles.value.find((p) => !beforeIds.has(p.id)) ??
      profiles.value.find((p) => p.name === name);
    if (created) {
      await openProfile(created);
      if (gen !== profileSaveGen) return;
    } else if (profiles.value.length > 0) {
      const fallback =
        profiles.value.find((p) => p.isDefault) ?? profiles.value[0];
      await openProfile(fallback);
      if (gen !== profileSaveGen) return;
    }
  } catch (e) {
    if (gen !== profileSaveGen) return;
    showSaveError(e);
    try {
      profiles.value = await App.launchProfiles();
    } catch (listErr) {
      if (gen !== profileSaveGen) return;
      showSaveError(listErr);
      return;
    }
    if (gen !== profileSaveGen) return;
    if (!previousId) return;
    const prev = profiles.value.find((p) => p.id === previousId);
    if (prev) {
      await openProfile(prev);
      if (gen !== profileSaveGen) return;
    }
  }
}

function sanitizeLaunchArgs(a: LaunchArgsParsedDTO): LaunchArgsParsedDTO {
  const pp = Number(a.processPriority);
  const mtp = Number(a.mainThreadPriority);
  const profile = Number(a.profile);
  let customArmRatio = Number(a.customArmRatio);
  let calibrationRange = Number(a.calibrationRange);
  if (valueOptionsEnabled.customArmRatio) {
    if (!Number.isFinite(customArmRatio) || !(customArmRatio > 0))
      customArmRatio = IK_CUSTOM_ARM_RATIO_DEFAULT;
  } else {
    customArmRatio = 0;
  }
  if (valueOptionsEnabled.calibrationRange) {
    if (!Number.isFinite(calibrationRange) || !(calibrationRange > 0))
      calibrationRange = IK_CALIBRATION_RANGE_DEFAULT;
  } else {
    calibrationRange = 0;
  }
  const base = {
    ...a,
    screenWidth: Math.max(0, Number(a.screenWidth) || 0),
    screenHeight: Math.max(0, Number(a.screenHeight) || 0),
    fps: Math.max(0, Number(a.fps) || 0),
    processPriority:
      Number.isInteger(pp) && pp >= -2 && pp <= 2 ? pp : PRIORITY_OMIT,
    mainThreadPriority:
      Number.isInteger(mtp) && mtp >= -2 && mtp <= 2 ? mtp : PRIORITY_OMIT,
    monitor: Math.max(0, Math.floor(Number(a.monitor) || 0)),
    profile: Number.isInteger(profile) && profile >= 0 ? profile : -1,
    customArmRatio,
    calibrationRange,
    disableShoulderTracking: !!a.disableShoulderTracking,
    enableIkDebugLogging: !!a.enableIkDebugLogging,
    freezeTrackingOnDisconnect: !!a.freezeTrackingOnDisconnect,
  };
  if (!valueOptionsEnabled.resolution) {
    base.screenWidth = 0;
    base.screenHeight = 0;
  }
  if (!valueOptionsEnabled.monitor) base.monitor = 0;
  if (!valueOptionsEnabled.fps) base.fps = 0;
  if (!valueOptionsEnabled.processPriority)
    base.processPriority = PRIORITY_OMIT;
  if (!valueOptionsEnabled.mainThreadPriority)
    base.mainThreadPriority = PRIORITY_OMIT;
  if (!valueOptionsEnabled.profile) base.profile = -1;
  if (!valueOptionsEnabled.midi) base.midi = "";
  if (!valueOptionsEnabled.ignoreTrackers) base.ignoreTrackers = "";
  if (!valueOptionsEnabled.osc) base.osc = "";
  if (!valueOptionsEnabled.affinity) base.affinity = "";
  return base;
}

async function save(): Promise<boolean> {
  if (!selected.value) return false;
  const gen = ++profileSaveGen;
  try {
    const argsStr = await App.mergeLaunchArgsForGUI(
      sanitizeLaunchArgs(launchArgs.value),
    );
    await App.saveLaunchProfile({
      ...selected.value,
      arguments: argsStr,
    });
    if (gen !== profileSaveGen) return false;
    profiles.value = await App.launchProfiles();
    if (gen !== profileSaveGen) return false;
    const id = selected.value.id;
    const refreshed = profiles.value.find((p) =>
      id ? p.id === id : p.name === selected.value!.name,
    );
    if (!refreshed) {
      showSaveError(new Error(t("launcher.errProfileNotFound")));
      return false;
    }
    await openProfile(refreshed);
    if (gen !== profileSaveGen) return false;
    return true;
  } catch (e) {
    if (gen !== profileSaveGen) return false;
    showSaveError(e);
    return false;
  }
}

async function launch() {
  if (!selected.value) return;
  const argsStr = await App.mergeLaunchArgsForGUI(
    sanitizeLaunchArgs(launchArgs.value),
  );
  await App.launchVRChatWithArgs(argsStr, selected.value.id ?? "");
}

async function requestDuplicate() {
  if (!selected.value?.id) return;
  const sourceId = selected.value.id;
  const ok = await guardUnsavedEdits();
  if (!ok) return;

  const source = profiles.value.find((p) => p.id === sourceId);
  if (!source) {
    showSaveError(new Error(t("launcher.errProfileNotFound")));
    return;
  }

  const beforeIds = new Set(profiles.value.map((p) => p.id));
  const name = nextDefaultLaunchProfileName(
    source.name,
    profiles.value.map((p) => p.name),
  );
  const gen = ++profileSaveGen;
  try {
    await App.saveLaunchProfile({
      id: "",
      name,
      arguments: source.arguments,
      isDefault: false,
    });
    if (gen !== profileSaveGen) return;
    profiles.value = await App.launchProfiles();
    if (gen !== profileSaveGen) return;
    const created =
      profiles.value.find((p) => !beforeIds.has(p.id)) ??
      profiles.value.find((p) => p.name === name);
    if (created) {
      await openProfile(created);
      if (gen !== profileSaveGen) return;
    } else if (profiles.value.length > 0) {
      const fallback =
        profiles.value.find((p) => p.isDefault) ?? profiles.value[0];
      await openProfile(fallback);
      if (gen !== profileSaveGen) return;
    }
  } catch (e) {
    if (gen !== profileSaveGen) return;
    showSaveError(e);
    try {
      profiles.value = await App.launchProfiles();
    } catch (listErr) {
      if (gen !== profileSaveGen) return;
      showSaveError(listErr);
      return;
    }
    if (gen !== profileSaveGen) return;
    const prev = profiles.value.find((p) => p.id === sourceId);
    if (prev) {
      await openProfile(prev);
      if (gen !== profileSaveGen) return;
    } else if (profiles.value.length > 0) {
      const fallback =
        profiles.value.find((p) => p.isDefault) ?? profiles.value[0];
      await openProfile(fallback);
      if (gen !== profileSaveGen) return;
    }
  }
}

async function confirmDelete() {
  if (!selected.value?.id) return;
  try {
    await ElMessageBox.confirm(
      t("launcher.deleteConfirm", { name: selected.value.name }),
      t("common.confirm"),
      {
        confirmButtonText: t("launcher.deleteOk"),
        cancelButtonText: t("common.cancel"),
        type: "warning",
        confirmButtonClass: VT_BUTTON_DANGER_CONFIRM_CLASS,
        cancelButtonClass: VT_BUTTON_SECONDARY_CANCEL_CLASS,
      },
    );
  } catch {
    return;
  }
  await App.deleteLaunchProfile(selected.value.id);
  selected.value = null;
  savedSnapshot.value = null;
  launchArgs.value = defaultLaunchArgs();
  Object.assign(valueOptionsEnabled, defaultValueOptionsEnabled());
  profiles.value = await App.launchProfiles();
  if (profiles.value.length > 0) {
    const p = profiles.value.find((pr) => pr.isDefault) ?? profiles.value[0];
    await openProfile(p);
  }
}

onBeforeRouteLeave(async (_to, _from, next) => {
  if (!isDirty.value) {
    next();
    return;
  }
  const choice = await promptUnsavedChoice();
  if (choice === "cancel") {
    next(false);
    return;
  }
  if (choice === "save") {
    const saved = await save();
    if (!saved) {
      next(false);
      return;
    }
  }
  next();
});
</script>

<style scoped>
.launcher-layout {
  display: flex;
  gap: var(--space-block);
  align-items: flex-start;
}

.launcher-layout.sidebar-collapsed .profiles-sidebar {
  width: auto;
  min-width: 0;
}

.profiles-sidebar {
  width: 200px;
  flex-shrink: 0;
  transition: width 0.15s ease;
}

.sidebar-toolbar {
  display: flex;
  margin-bottom: var(--space-action-group);
  align-items: center;
}

.sidebar-toggle {
  flex-shrink: 0;
}

.profiles-list {
  width: 100%;
}

.btn-add {
  width: 100%;
  margin-bottom: var(--space-action-group);
  border-style: dashed !important;
  color: var(--color-text-secondary) !important;
}

.btn-add:hover {
  color: var(--color-brand) !important;
}

.profile-card {
  padding: var(--space-form-field);
  margin-bottom: var(--space-action-group);
  background: var(--color-bg-elevated);
  border-radius: var(--radius);
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: var(--space-action-group);
  transition: background 0.15s;
}

.profile-card:hover,
.profile-card.active {
  background: var(--color-bg-muted);
}

.profile-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.unsaved-dot {
  width: var(--space-action-group);
  height: var(--space-action-group);
  border-radius: 50%;
  background: var(--color-warning);
  flex-shrink: 0;
}

.profile-editor {
  flex: 1;
  min-width: 0;
  overflow-y: auto;
}

.editor-toolbar {
  display: flex;
  flex-direction: column;
  gap: var(--space-form-field);
  margin-bottom: var(--space-block);
}

.unsaved-banner {
  width: 100%;
}

.toolbar-actions {
  display: flex;
  gap: var(--space-action-group);
  justify-content: flex-end;
  flex-wrap: wrap;
  align-items: center;
}

.launch-args-gui {
  display: flex;
  flex-direction: column;
  gap: var(--space-action-group);
  width: 100%;
}

.arg-row {
  display: flex;
  align-items: center;
}

.sub-options {
  margin: var(--space-inline-tight) 0 var(--space-action-group)
    var(--space-section);
  display: flex;
  align-items: center;
  gap: var(--space-action-group);
  flex-wrap: wrap;
}

.nested-form-item {
  margin-bottom: var(--space-action-group) !important;
}

.nested-form-item :deep(.el-form-item__label) {
  font-size: var(--font-size-14);
  color: var(--color-text-secondary);
  padding-bottom: var(--space-inline-tight) !important;
}

.resolution-fields {
  display: flex;
  align-items: center;
  gap: var(--space-action-group);
}

.resolution-sep {
  color: var(--color-text-secondary);
}

.args-collapse {
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  background: var(--color-bg-muted);
  margin: var(--space-action-group) 0;
}

.args-collapse :deep(.el-collapse-item__header) {
  background: transparent;
  border-bottom-color: var(--color-border);
  color: var(--color-text-secondary);
  font-size: var(--font-size-14);
  padding: 0 var(--space-form-field);
  height: 40px;
}

.args-collapse :deep(.el-collapse-item__content) {
  padding: var(--space-form-field);
}

.args-collapse :deep(.el-collapse-item__wrap) {
  border-bottom-color: var(--color-border);
  background: transparent;
}

.advanced-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-form-field);
}

.advanced-section-title {
  margin: 0;
  font-size: var(--font-size-14);
  font-weight: var(--font-weight-600);
  color: var(--color-text-secondary);
}

.launch-args-advanced {
  display: flex;
  flex-direction: column;
  gap: var(--space-inline-tight);
}
</style>
