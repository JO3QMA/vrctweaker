<template>
  <div class="launcher-view">
    <h1 class="page-title">ランチャー</h1>
    <div class="profiles-section">
      <!-- プロファイルリスト -->
      <div class="profiles-list">
        <el-button class="btn-add" @click="addNew"
          >+ 新規プロファイル</el-button
        >
        <div
          v-for="p in profiles"
          :key="p.id"
          class="profile-card"
          :class="{ active: selected?.id === p.id }"
          @click="select(p)"
        >
          <span class="profile-name">{{ p.name }}</span>
          <el-tag v-if="p.isDefault" size="small" type="primary">既定</el-tag>
        </div>
      </div>

      <!-- プロファイルエディタ -->
      <div v-if="selected" class="profile-editor">
        <el-form label-position="top" size="default">
          <el-form-item label="プロファイル名">
            <el-input v-model="selected.name" />
          </el-form-item>

          <el-form-item label="起動引数">
            <div class="launch-args-gui">
              <div class="arg-row">
                <el-checkbox
                  v-model="launchArgs.noVr"
                  data-testid="no-vr-checkbox"
                >
                  デスクトップモードで起動（-no-vr）
                </el-checkbox>
              </div>

              <!-- 表示モード -->
              <el-form-item label="表示モード" class="nested-form-item">
                <el-radio-group v-model="launchArgs.screenMode" size="default">
                  <el-radio-button
                    value="fullscreen"
                    data-testid="screen-mode-fullscreen"
                    >フルスクリーン</el-radio-button
                  >
                  <el-radio-button
                    value="windowed"
                    data-testid="screen-mode-windowed"
                    >ウィンドウ</el-radio-button
                  >
                  <el-radio-button
                    value="popupwindow"
                    data-testid="screen-mode-popupwindow"
                    >仮想フルスクリーン</el-radio-button
                  >
                </el-radio-group>
              </el-form-item>

              <!-- 詳細設定 -->
              <el-collapse class="args-collapse">
                <el-collapse-item title="詳細設定" name="advanced">
                  <div class="launch-args-advanced">
                    <div class="arg-row">
                      <el-checkbox
                        v-model="valueOptionsEnabled.resolution"
                        data-testid="resolution-enabled-checkbox"
                        @change="onResolutionEnabledChange"
                      >
                        解像度を指定（-screen-width, -screen-height）
                      </el-checkbox>
                    </div>
                    <div
                      v-if="valueOptionsEnabled.resolution"
                      class="sub-options"
                    >
                      <el-form-item label="プリセット" class="nested-form-item">
                        <el-radio-group v-model="resolutionPreset" size="small">
                          <el-radio-button
                            value="HD"
                            data-testid="resolution-preset-hd"
                            @change="applyResolutionPreset"
                            >HD</el-radio-button
                          >
                          <el-radio-button
                            value="FHD"
                            data-testid="resolution-preset-fhd"
                            @change="applyResolutionPreset"
                            >FHD</el-radio-button
                          >
                          <el-radio-button
                            value="WQHD"
                            data-testid="resolution-preset-wqhd"
                            @change="applyResolutionPreset"
                            >WQHD</el-radio-button
                          >
                          <el-radio-button
                            value="4K"
                            data-testid="resolution-preset-4k"
                            @change="applyResolutionPreset"
                            >4K</el-radio-button
                          >
                          <el-radio-button
                            value="custom"
                            data-testid="resolution-preset-custom"
                            @change="applyResolutionPreset"
                            >手動設定</el-radio-button
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
                          placeholder="幅"
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
                          placeholder="高さ"
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
                        モニター指定（-monitor N）
                      </el-checkbox>
                    </div>
                    <div v-if="valueOptionsEnabled.monitor" class="sub-options">
                      <el-input-number
                        v-model="launchArgs.monitor"
                        :min="1"
                        data-testid="monitor-input"
                        size="small"
                        placeholder="1=1番目"
                        style="width: 120px"
                      />
                    </div>

                    <div class="arg-row">
                      <el-checkbox
                        v-model="valueOptionsEnabled.fps"
                        data-testid="fps-enabled-checkbox"
                        @change="onFpsEnabledChange"
                      >
                        FPS制限（--fps=N）
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
                        レジストリ登録スキップ（--skip-registry-install）
                      </el-checkbox>
                    </div>

                    <div class="arg-row">
                      <el-checkbox
                        v-model="valueOptionsEnabled.processPriority"
                        data-testid="process-priority-enabled-checkbox"
                        @change="onProcessPriorityEnabledChange"
                      >
                        プロセス優先度（--process-priority=N）
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
                        メインスレッド優先度（--main-thread-priority=N）
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
                        プロファイル（--profile=N）
                      </el-checkbox>
                    </div>
                    <div v-if="valueOptionsEnabled.profile" class="sub-options">
                      <el-input-number
                        v-model="launchArgs.profile"
                        :min="0"
                        data-testid="profile-input"
                        size="small"
                        placeholder="0=既定"
                        style="width: 120px"
                      />
                    </div>
                  </div>
                </el-collapse-item>
                <el-collapse-item
                  title="クリエイター・デバッグ向け"
                  name="debug"
                >
                  <div class="launch-args-advanced">
                    <div class="arg-row">
                      <el-checkbox
                        v-model="launchArgs.enableDebugGui"
                        data-testid="enable-debug-gui-checkbox"
                      >
                        デバッグGUI（--enable-debug-gui）
                      </el-checkbox>
                    </div>
                    <div class="arg-row">
                      <el-checkbox
                        v-model="launchArgs.enableSDKLogLevels"
                        data-testid="enable-sdk-log-levels-checkbox"
                      >
                        SDKログ拡張（--enable-sdk-log-levels）
                      </el-checkbox>
                    </div>
                    <div class="arg-row">
                      <el-checkbox
                        v-model="launchArgs.enableUdonDebugLogging"
                        data-testid="enable-udon-debug-logging-checkbox"
                      >
                        Udonデバッグログ（--enable-udon-debug-logging）
                      </el-checkbox>
                    </div>
                    <div class="arg-row">
                      <el-checkbox
                        v-model="launchArgs.watchWorlds"
                        data-testid="watch-worlds-checkbox"
                      >
                        ワールド監視（--watch-worlds）
                      </el-checkbox>
                    </div>
                    <div class="arg-row">
                      <el-checkbox
                        v-model="launchArgs.watchAvatars"
                        data-testid="watch-avatars-checkbox"
                      >
                        アバター監視（--watch-avatars）
                      </el-checkbox>
                    </div>
                    <div class="arg-row">
                      <el-checkbox
                        v-model="launchArgs.enforceWorldServerChecks"
                        data-testid="enforce-world-server-checks-checkbox"
                      >
                        ワールドサーバーチェック強制（--enforce-world-server-checks）
                      </el-checkbox>
                    </div>

                    <div class="arg-row">
                      <el-checkbox
                        v-model="valueOptionsEnabled.midi"
                        data-testid="midi-enabled-checkbox"
                        @change="onMidiEnabledChange"
                      >
                        MIDIデバイス（--midi=deviceName）
                      </el-checkbox>
                    </div>
                    <div v-if="valueOptionsEnabled.midi" class="sub-options">
                      <el-input
                        v-model="launchArgs.midi"
                        placeholder="デバイス名"
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
                        無視トラッカー（--ignore-trackers=serial1,serial2）
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
                      label="動画デコーディング"
                      class="nested-form-item"
                    >
                      <el-radio-group
                        v-model="launchArgs.videoDecoding"
                        size="small"
                      >
                        <el-radio-button
                          value=""
                          data-testid="video-decoding-default"
                          >既定</el-radio-button
                        >
                        <el-radio-button
                          value="software"
                          data-testid="video-decoding-software"
                          >ソフトウェア</el-radio-button
                        >
                        <el-radio-button
                          value="hardware"
                          data-testid="video-decoding-hardware"
                          >ハードウェア</el-radio-button
                        >
                      </el-radio-group>
                    </el-form-item>

                    <div class="arg-row">
                      <el-checkbox
                        v-model="launchArgs.disableAMDStutterWorkaround"
                        data-testid="disable-amd-stutter-workaround-checkbox"
                      >
                        AMDスタッター回避無効（--disable-amd-stutter-workaround）
                      </el-checkbox>
                    </div>

                    <div class="arg-row">
                      <el-checkbox
                        v-model="valueOptionsEnabled.osc"
                        data-testid="osc-enabled-checkbox"
                        @change="onOscEnabledChange"
                      >
                        OSC（--osc=inPort:outIP:outPort）
                      </el-checkbox>
                    </div>
                    <div v-if="valueOptionsEnabled.osc" class="sub-options">
                      <el-input
                        v-model="launchArgs.osc"
                        placeholder="例: 9000:127.0.0.1:9001"
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
                        スレッドアフィニティ（--affinity=FFFF）
                      </el-checkbox>
                    </div>
                    <div
                      v-if="valueOptionsEnabled.affinity"
                      class="sub-options"
                    >
                      <el-input
                        v-model="launchArgs.affinity"
                        placeholder="16進ビットマスク"
                        data-testid="affinity-input"
                        size="small"
                        style="max-width: 200px"
                      />
                    </div>
                  </div>
                </el-collapse-item>
              </el-collapse>

              <el-form-item label="カスタム引数（上級者向け）">
                <el-input
                  v-model="launchArgs.custom"
                  placeholder="-batchmode"
                  data-testid="custom-args-input"
                />
              </el-form-item>
            </div>
          </el-form-item>

          <el-form-item>
            <el-checkbox v-model="selected.isDefault">
              デフォルトに設定
            </el-checkbox>
          </el-form-item>

          <div class="editor-actions">
            <el-button class="btn-save" type="primary" @click="save"
              >保存</el-button
            >
            <el-button class="btn-launch" type="success" @click="launch"
              >この設定で起動</el-button
            >
            <el-button
              v-if="selected.id"
              type="danger"
              plain
              data-testid="delete-profile-btn"
              style="margin-left: auto"
              @click="confirmDelete"
            >
              削除
            </el-button>
          </div>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { ElMessageBox } from "element-plus";
import {
  App,
  type LaunchProfileDTO,
  type LaunchArgsParsedDTO,
  PRIORITY_OMIT,
} from "../wails/app";

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

interface ValueOptionsEnabled {
  resolution: boolean;
  monitor: boolean;
  fps: boolean;
  processPriority: boolean;
  mainThreadPriority: boolean;
  profile: boolean;
  midi: boolean;
  ignoreTrackers: boolean;
  osc: boolean;
  affinity: boolean;
}

function defaultValueOptionsEnabled(): ValueOptionsEnabled {
  return {
    resolution: false,
    monitor: false,
    fps: false,
    processPriority: false,
    mainThreadPriority: false,
    profile: false,
    midi: false,
    ignoreTrackers: false,
    osc: false,
    affinity: false,
  };
}

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

function syncValueOptionsEnabled(a: LaunchArgsParsedDTO) {
  valueOptionsEnabled.resolution = a.screenWidth > 0 || a.screenHeight > 0;
  valueOptionsEnabled.monitor = a.monitor >= 1;
  valueOptionsEnabled.fps = a.fps > 0;
  valueOptionsEnabled.processPriority =
    a.processPriority !== PRIORITY_OMIT &&
    a.processPriority >= -2 &&
    a.processPriority <= 2;
  valueOptionsEnabled.mainThreadPriority =
    a.mainThreadPriority !== PRIORITY_OMIT &&
    a.mainThreadPriority >= -2 &&
    a.mainThreadPriority <= 2;
  valueOptionsEnabled.profile = a.profile >= 0;
  valueOptionsEnabled.midi = a.midi !== "";
  valueOptionsEnabled.ignoreTrackers = a.ignoreTrackers !== "";
  valueOptionsEnabled.osc = a.osc !== "";
  valueOptionsEnabled.affinity = a.affinity !== "";
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
  syncValueOptionsEnabled(launchArgs.value);
  syncResolutionPresetFromArgs();
}

onMounted(async () => {
  profiles.value = await App.launchProfiles();
  if (profiles.value.length > 0 && !selected.value) {
    const p = profiles.value.find((p) => p.isDefault) ?? profiles.value[0];
    selected.value = { ...p };
    await syncLaunchArgsFromProfile(p);
  }
});

async function select(p: LaunchProfileDTO) {
  selected.value = { ...p };
  await syncLaunchArgsFromProfile(p);
}

function addNew() {
  selected.value = {
    id: "",
    name: "新しいプロファイル",
    arguments: "",
    isDefault: profiles.value.length === 0,
  };
  launchArgs.value = defaultLaunchArgs();
  Object.assign(valueOptionsEnabled, defaultValueOptionsEnabled());
  resolutionPreset.value = "FHD";
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
  await App.deleteLaunchProfile(selected.value.id);
  selected.value = null;
  launchArgs.value = defaultLaunchArgs();
  Object.assign(valueOptionsEnabled, defaultValueOptionsEnabled());
  profiles.value = await App.launchProfiles();
  if (profiles.value.length > 0) {
    const p = profiles.value.find((pr) => pr.isDefault) ?? profiles.value[0];
    selected.value = { ...p };
    await syncLaunchArgsFromProfile(p);
  }
}
</script>

<style scoped>
.profiles-section {
  display: flex;
  gap: 1.5rem;
}

.profiles-list {
  width: 240px;
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

.profile-editor {
  flex: 1;
  min-width: 0;
  overflow-y: auto;
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

.launch-args-advanced {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.editor-actions {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.5rem;
  flex-wrap: wrap;
}
</style>
