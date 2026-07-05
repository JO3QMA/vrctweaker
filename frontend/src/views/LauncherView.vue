<template>
  <div class="launcher-view">
    <h1 class="page-title">{{ t("routes.launcher") }}</h1>
    <div class="launcher-layout" :class="{ 'sidebar-collapsed': !sidebarOpen }">
      <aside class="profiles-sidebar">
        <div class="sidebar-toolbar">
          <el-button
            class="sidebar-toggle"
            data-testid="sidebar-toggle-btn"
            :aria-label="t('launcher.toggleSidebar')"
            @click="toggleSidebar"
          >
            <el-icon><component :is="sidebarOpen ? Fold : Expand" /></el-icon>
          </el-button>
        </div>
        <el-button v-show="sidebarOpen" class="btn-add" @click="requestAddNew">
          {{ t("launcher.newProfile") }}
        </el-button>
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
            <el-tag v-if="p.isDefault" size="small" type="primary">{{
              t("launcher.defaultTag")
            }}</el-tag>
          </div>
        </div>
      </aside>

      <div v-if="selected" class="profile-editor">
        <div class="editor-toolbar">
          <el-alert
            v-if="isDirty"
            class="unsaved-banner"
            data-testid="unsaved-banner"
            type="warning"
            :title="t('launcher.unsavedBanner')"
            :closable="false"
            show-icon
          />
          <div class="toolbar-actions">
            <el-button class="btn-launch" @click="launch">
              {{ t("launcher.launchWithThis") }}
            </el-button>
            <el-button class="btn-save" type="primary" @click="save">
              {{ t("launcher.save") }}
            </el-button>
            <el-dropdown
              v-if="selected.id"
              trigger="click"
              @command="onOverflowCommand"
            >
              <el-button
                data-testid="profile-overflow-btn"
                :aria-label="t('launcher.moreActions')"
              >
                ⋯
              </el-button>
              <template #dropdown>
                <el-dropdown-item
                  command="delete"
                  data-testid="delete-profile-btn"
                >
                  {{ t("launcher.delete") }}
                </el-dropdown-item>
              </template>
            </el-dropdown>
          </div>
        </div>

        <el-form label-position="top" size="default">
          <el-form-item :label="t('launcher.profileName')">
            <el-input v-model="selected.name" />
          </el-form-item>

          <el-form-item>
            <el-checkbox v-model="selected.isDefault">
              {{ t("launcher.setAsDefault") }}
            </el-checkbox>
          </el-form-item>

          <el-form-item :label="t('launcher.launchArgs')">
            <div class="launch-args-gui">
              <div class="arg-row">
                <el-checkbox
                  v-model="launchArgs.noVr"
                  data-testid="no-vr-checkbox"
                >
                  {{ t("launcher.desktopMode") }}
                </el-checkbox>
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
                <el-input
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
                        <el-checkbox
                          v-model="valueOptionsEnabled.resolution"
                          data-testid="resolution-enabled-checkbox"
                          @change="onResolutionEnabledChange"
                        >
                          {{ t("launcher.resolutionHint") }}
                        </el-checkbox>
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
                        <el-checkbox
                          v-model="valueOptionsEnabled.monitor"
                          data-testid="monitor-enabled-checkbox"
                          @change="onMonitorEnabledChange"
                        >
                          {{ t("launcher.monitorHint") }}
                        </el-checkbox>
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
                        <el-checkbox
                          v-model="valueOptionsEnabled.fps"
                          data-testid="fps-enabled-checkbox"
                          @change="onFpsEnabledChange"
                        >
                          {{ t("launcher.fpsHint") }}
                        </el-checkbox>
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
                        <el-checkbox
                          v-model="launchArgs.skipRegistry"
                          data-testid="skip-registry-checkbox"
                        >
                          {{ t("launcher.skipRegistry") }}
                        </el-checkbox>
                      </div>

                      <div class="arg-row">
                        <el-checkbox
                          v-model="valueOptionsEnabled.processPriority"
                          data-testid="process-priority-enabled-checkbox"
                          @change="onProcessPriorityEnabledChange"
                        >
                          {{ t("launcher.processPriority") }}
                        </el-checkbox>
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
                        <el-checkbox
                          v-model="valueOptionsEnabled.mainThreadPriority"
                          data-testid="main-thread-priority-enabled-checkbox"
                          @change="onMainThreadPriorityEnabledChange"
                        >
                          {{ t("launcher.mainThreadPriority") }}
                        </el-checkbox>
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
                        <el-checkbox
                          v-model="valueOptionsEnabled.profile"
                          data-testid="profile-enabled-checkbox"
                          @change="onProfileEnabledChange"
                        >
                          {{ t("launcher.profileHint") }}
                        </el-checkbox>
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
                        <el-checkbox
                          v-model="launchArgs.enableDebugGui"
                          data-testid="enable-debug-gui-checkbox"
                        >
                          {{ t("launcher.debugGui") }}
                        </el-checkbox>
                      </div>
                      <div class="arg-row">
                        <el-checkbox
                          v-model="launchArgs.enableSDKLogLevels"
                          data-testid="enable-sdk-log-levels-checkbox"
                        >
                          {{ t("launcher.sdkLog") }}
                        </el-checkbox>
                      </div>
                      <div class="arg-row">
                        <el-checkbox
                          v-model="launchArgs.enableUdonDebugLogging"
                          data-testid="enable-udon-debug-logging-checkbox"
                        >
                          {{ t("launcher.udonDebug") }}
                        </el-checkbox>
                      </div>
                      <div class="arg-row">
                        <el-checkbox
                          v-model="launchArgs.watchWorlds"
                          data-testid="watch-worlds-checkbox"
                        >
                          {{ t("launcher.watchWorlds") }}
                        </el-checkbox>
                      </div>
                      <div class="arg-row">
                        <el-checkbox
                          v-model="launchArgs.watchAvatars"
                          data-testid="watch-avatars-checkbox"
                        >
                          {{ t("launcher.watchAvatars") }}
                        </el-checkbox>
                      </div>
                      <div class="arg-row">
                        <el-checkbox
                          v-model="launchArgs.enforceWorldServerChecks"
                          data-testid="enforce-world-server-checks-checkbox"
                        >
                          {{ t("launcher.enforceWorldServer") }}
                        </el-checkbox>
                      </div>

                      <div class="arg-row">
                        <el-checkbox
                          v-model="valueOptionsEnabled.midi"
                          data-testid="midi-enabled-checkbox"
                          @change="onMidiEnabledChange"
                        >
                          {{ t("launcher.midi") }}
                        </el-checkbox>
                      </div>
                      <div v-if="valueOptionsEnabled.midi" class="sub-options">
                        <el-input
                          v-model="launchArgs.midi"
                          :placeholder="t('launcher.midiPh')"
                          data-testid="midi-input"
                          size="small"
                          style="max-width: 240px"
                        />
                      </div>

                      <div class="arg-row">
                        <el-checkbox
                          v-model="valueOptionsEnabled.ignoreTrackers"
                          data-testid="ignore-trackers-enabled-checkbox"
                          @change="onIgnoreTrackersEnabledChange"
                        >
                          {{ t("launcher.ignoreTrackers") }}
                        </el-checkbox>
                      </div>
                      <div
                        v-if="valueOptionsEnabled.ignoreTrackers"
                        class="sub-options"
                      >
                        <el-input
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
                        <el-checkbox
                          v-model="launchArgs.disableAMDStutterWorkaround"
                          data-testid="disable-amd-stutter-workaround-checkbox"
                        >
                          {{ t("launcher.disableAmdStutter") }}
                        </el-checkbox>
                      </div>

                      <div class="arg-row">
                        <el-checkbox
                          v-model="valueOptionsEnabled.osc"
                          data-testid="osc-enabled-checkbox"
                          @change="onOscEnabledChange"
                        >
                          {{ t("launcher.osc") }}
                        </el-checkbox>
                      </div>
                      <div v-if="valueOptionsEnabled.osc" class="sub-options">
                        <el-input
                          v-model="launchArgs.osc"
                          :placeholder="t('launcher.oscPh')"
                          data-testid="osc-input"
                          size="small"
                          style="max-width: 240px"
                        />
                      </div>

                      <div class="arg-row">
                        <el-checkbox
                          v-model="valueOptionsEnabled.affinity"
                          data-testid="affinity-enabled-checkbox"
                          @change="onAffinityEnabledChange"
                        >
                          {{ t("launcher.affinity") }}
                        </el-checkbox>
                      </div>
                      <div
                        v-if="valueOptionsEnabled.affinity"
                        class="sub-options"
                      >
                        <el-input
                          v-model="launchArgs.affinity"
                          :placeholder="t('launcher.affinityPh')"
                          data-testid="affinity-input"
                          size="small"
                          style="max-width: 200px"
                        />
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
import { ref, reactive, computed, onMounted } from "vue";
import { onBeforeRouteLeave } from "vue-router";
import { useI18n } from "vue-i18n";
import { ElMessageBox } from "element-plus";
import { Expand, Fold } from "@element-plus/icons-vue";
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
  readSidebarOpenPreference,
  syncValueOptionsEnabled,
  writeSidebarOpenPreference,
  type LaunchProfileEditSnapshot,
  type ValueOptionsEnabled,
} from "./launcher/launcherProfileEdits";

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
  if (choice === "save") await save();
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

async function requestSelect(p: LaunchProfileDTO) {
  if (selected.value?.id === p.id && selected.value.name === p.name) return;
  const ok = await guardUnsavedEdits();
  if (!ok) return;
  await openProfile(p);
}

async function requestAddNew() {
  const ok = await guardUnsavedEdits();
  if (!ok) return;
  selected.value = {
    id: "",
    name: t("launcher.newProfileDefaultName"),
    arguments: "",
    isDefault: profiles.value.length === 0,
  };
  launchArgs.value = defaultLaunchArgs();
  Object.assign(valueOptionsEnabled, defaultValueOptionsEnabled());
  resolutionPreset.value = "FHD";
  syncAdvancedCollapseOpenState();
  captureSnapshot();
}

function sanitizeLaunchArgs(a: LaunchArgsParsedDTO): LaunchArgsParsedDTO {
  const pp = Number(a.processPriority);
  const mtp = Number(a.mainThreadPriority);
  const profile = Number(a.profile);
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

async function save() {
  if (!selected.value) return;
  const argsStr = await App.mergeLaunchArgsForGUI(
    sanitizeLaunchArgs(launchArgs.value),
  );
  selected.value.arguments = argsStr;
  await App.saveLaunchProfile(selected.value);
  profiles.value = await App.launchProfiles();
  const id = selected.value.id;
  const refreshed =
    profiles.value.find((p) =>
      id ? p.id === id : p.name === selected.value!.name,
    ) ?? selected.value;
  await openProfile(refreshed);
}

async function launch() {
  if (!selected.value) return;
  const argsStr = await App.mergeLaunchArgsForGUI(
    sanitizeLaunchArgs(launchArgs.value),
  );
  await App.launchVRChatWithArgs(argsStr);
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
        confirmButtonClass: "el-button--danger",
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

function onOverflowCommand(command: string) {
  if (command === "delete") void confirmDelete();
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
    try {
      await save();
    } catch {
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
  gap: 1rem;
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
  margin-bottom: 0.5rem;
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
  margin-bottom: 0.5rem;
  border-style: dashed !important;
  color: var(--text-secondary) !important;
}

.btn-add:hover {
  color: var(--accent) !important;
}

.profile-card {
  padding: 0.75rem;
  margin-bottom: 0.5rem;
  background: var(--bg-secondary);
  border-radius: var(--radius);
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: background 0.15s;
}

.profile-card:hover,
.profile-card.active {
  background: var(--bg-tertiary);
}

.profile-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.unsaved-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--el-color-warning);
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
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.unsaved-banner {
  width: 100%;
}

.toolbar-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
  flex-wrap: wrap;
  align-items: center;
}

.launch-args-gui {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  width: 100%;
}

.arg-row {
  display: flex;
  align-items: center;
}

.sub-options {
  margin: 0.25rem 0 0.5rem 1.5rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.nested-form-item {
  margin-bottom: 0.5rem !important;
}

.nested-form-item :deep(.el-form-item__label) {
  font-size: 0.85rem;
  color: var(--text-secondary);
  padding-bottom: 0.25rem !important;
}

.resolution-fields {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.resolution-sep {
  color: var(--text-secondary);
}

.args-collapse {
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: var(--bg-tertiary);
  margin: 0.5rem 0;
}

.args-collapse :deep(.el-collapse-item__header) {
  background: transparent;
  border-bottom-color: var(--border);
  color: var(--text-secondary);
  font-size: 0.9rem;
  padding: 0 0.75rem;
  height: 40px;
}

.args-collapse :deep(.el-collapse-item__content) {
  padding: 0.75rem;
}

.args-collapse :deep(.el-collapse-item__wrap) {
  border-bottom-color: var(--border);
  background: transparent;
}

.advanced-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.advanced-section-title {
  margin: 0;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-secondary);
}

.launch-args-advanced {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}
</style>
