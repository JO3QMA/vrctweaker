<template>
  <div class="launcher-view">
    <h1 class="page-title">
      ランチャー
    </h1>
    <div class="profiles-section">
      <div class="profiles-list">
        <button
          class="btn-add"
          @click="addNew"
        >
          + 新規プロファイル
        </button>
        <div
          v-for="p in profiles"
          :key="p.id"
          class="profile-card"
          :class="{ active: selected?.id === p.id }"
          @click="select(p)"
        >
          <span class="profile-name">{{ p.name }}</span>
          <span
            v-if="p.isDefault"
            class="badge"
          >既定</span>
        </div>
      </div>
      <div
        v-if="selected"
        class="profile-editor"
      >
        <label>プロファイル名</label>
        <input
          v-model="selected.name"
          type="text"
        >
        <label>起動引数</label>
        <div class="launch-args-gui">
          <label class="checkbox-row">
            <input
              v-model="launchArgs.noVr"
              type="checkbox"
              data-testid="no-vr-checkbox"
            >
            デスクトップモードで起動（-no-vr）
          </label>
          <div class="screen-mode-section">
            <label class="block-label">表示モード</label>
            <div
              class="toggle-group"
              role="group"
              aria-label="表示モード"
            >
              <label
                class="toggle-option"
                :class="{ active: launchArgs.screenMode === 'fullscreen' }"
              >
                <input
                  v-model="launchArgs.screenMode"
                  type="radio"
                  value="fullscreen"
                  data-testid="screen-mode-fullscreen"
                >
                <span>フルスクリーン</span>
              </label>
              <label
                class="toggle-option"
                :class="{ active: launchArgs.screenMode === 'windowed' }"
              >
                <input
                  v-model="launchArgs.screenMode"
                  type="radio"
                  value="windowed"
                  data-testid="screen-mode-windowed"
                >
                <span>ウィンドウ</span>
              </label>
              <label
                class="toggle-option"
                :class="{ active: launchArgs.screenMode === 'popupwindow' }"
              >
                <input
                  v-model="launchArgs.screenMode"
                  type="radio"
                  value="popupwindow"
                  data-testid="screen-mode-popupwindow"
                >
                <span>仮想フルスクリーン</span>
              </label>
            </div>
          </div>
          <details class="details-advanced">
            <summary>詳細設定</summary>
            <div class="launch-args-advanced">
              <label class="checkbox-row">
                <input
                  v-model="valueOptionsEnabled.resolution"
                  type="checkbox"
                  data-testid="resolution-enabled-checkbox"
                  @change="onResolutionEnabledChange"
                >
                解像度を指定（-screen-width, -screen-height）
              </label>
              <div
                v-if="valueOptionsEnabled.resolution"
                class="option-value-row"
              >
                <label class="resolution-field">
                  <span class="resolution-field-label">幅</span>
                  <input
                    v-model.number="launchArgs.screenWidth"
                    type="number"
                    min="0"
                    placeholder="1920"
                    data-testid="screen-width-input"
                  >
                </label>
                <span class="resolution-sep">×</span>
                <label class="resolution-field">
                  <span class="resolution-field-label">高さ</span>
                  <input
                    v-model.number="launchArgs.screenHeight"
                    type="number"
                    min="0"
                    placeholder="1080"
                    data-testid="screen-height-input"
                  >
                </label>
              </div>
              <label class="checkbox-row">
                <input
                  v-model="valueOptionsEnabled.monitor"
                  type="checkbox"
                  data-testid="monitor-enabled-checkbox"
                  @change="onMonitorEnabledChange"
                >
                モニター指定（-monitor N）
              </label>
              <div
                v-if="valueOptionsEnabled.monitor"
                class="option-value-row"
              >
                <input
                  v-model.number="launchArgs.monitor"
                  type="number"
                  min="1"
                  placeholder="1=1番目"
                  data-testid="monitor-input"
                >
              </div>
              <label class="checkbox-row">
                <input
                  v-model="valueOptionsEnabled.fps"
                  type="checkbox"
                  data-testid="fps-enabled-checkbox"
                  @change="onFpsEnabledChange"
                >
                FPS制限（--fps=N）
              </label>
              <div
                v-if="valueOptionsEnabled.fps"
                class="option-value-row"
              >
                <input
                  v-model.number="launchArgs.fps"
                  type="number"
                  min="1"
                  placeholder="90"
                  data-testid="fps-input"
                >
              </div>
              <label class="checkbox-row">
                <input
                  v-model="launchArgs.skipRegistry"
                  type="checkbox"
                  data-testid="skip-registry-checkbox"
                >
                レジストリ登録スキップ（--skip-registry-install）
              </label>
              <label class="checkbox-row">
                <input
                  v-model="valueOptionsEnabled.processPriority"
                  type="checkbox"
                  data-testid="process-priority-enabled-checkbox"
                  @change="onProcessPriorityEnabledChange"
                >
                プロセス優先度（--process-priority=N）
              </label>
              <div
                v-if="valueOptionsEnabled.processPriority"
                class="option-value-row"
              >
                <input
                  v-model.number="launchArgs.processPriority"
                  type="number"
                  min="-2"
                  max="2"
                  placeholder="-2～2"
                  data-testid="process-priority-input"
                >
              </div>
              <label class="checkbox-row">
                <input
                  v-model="valueOptionsEnabled.mainThreadPriority"
                  type="checkbox"
                  data-testid="main-thread-priority-enabled-checkbox"
                  @change="onMainThreadPriorityEnabledChange"
                >
                メインスレッド優先度（--main-thread-priority=N）
              </label>
              <div
                v-if="valueOptionsEnabled.mainThreadPriority"
                class="option-value-row"
              >
                <input
                  v-model.number="launchArgs.mainThreadPriority"
                  type="number"
                  min="-2"
                  max="2"
                  placeholder="-2～2"
                  data-testid="main-thread-priority-input"
                >
              </div>
              <label class="checkbox-row">
                <input
                  v-model="valueOptionsEnabled.profile"
                  type="checkbox"
                  data-testid="profile-enabled-checkbox"
                  @change="onProfileEnabledChange"
                >
                プロファイル（--profile=N）
              </label>
              <div
                v-if="valueOptionsEnabled.profile"
                class="option-value-row"
              >
                <input
                  v-model.number="launchArgs.profile"
                  type="number"
                  min="0"
                  placeholder="0=既定"
                  data-testid="profile-input"
                >
              </div>
            </div>
          </details>
          <details class="details-advanced">
            <summary>クリエイター・デバッグ向け</summary>
            <div class="launch-args-advanced">
              <label class="checkbox-row">
                <input
                  v-model="launchArgs.enableDebugGui"
                  type="checkbox"
                  data-testid="enable-debug-gui-checkbox"
                >
                デバッグGUI（--enable-debug-gui）
              </label>
              <label class="checkbox-row">
                <input
                  v-model="launchArgs.enableSDKLogLevels"
                  type="checkbox"
                  data-testid="enable-sdk-log-levels-checkbox"
                >
                SDKログ拡張（--enable-sdk-log-levels）
              </label>
              <label class="checkbox-row">
                <input
                  v-model="launchArgs.enableUdonDebugLogging"
                  type="checkbox"
                  data-testid="enable-udon-debug-logging-checkbox"
                >
                Udonデバッグログ（--enable-udon-debug-logging）
              </label>
              <label class="checkbox-row">
                <input
                  v-model="launchArgs.watchWorlds"
                  type="checkbox"
                  data-testid="watch-worlds-checkbox"
                >
                ワールド監視（--watch-worlds）
              </label>
              <label class="checkbox-row">
                <input
                  v-model="launchArgs.watchAvatars"
                  type="checkbox"
                  data-testid="watch-avatars-checkbox"
                >
                アバター監視（--watch-avatars）
              </label>
              <label class="checkbox-row">
                <input
                  v-model="launchArgs.enforceWorldServerChecks"
                  type="checkbox"
                  data-testid="enforce-world-server-checks-checkbox"
                >
                ワールドサーバーチェック強制（--enforce-world-server-checks）
              </label>
              <label class="checkbox-row">
                <input
                  v-model="valueOptionsEnabled.midi"
                  type="checkbox"
                  data-testid="midi-enabled-checkbox"
                  @change="onMidiEnabledChange"
                >
                MIDIデバイス（--midi=deviceName）
              </label>
              <div
                v-if="valueOptionsEnabled.midi"
                class="option-value-row"
              >
                <input
                  v-model="launchArgs.midi"
                  type="text"
                  placeholder="デバイス名"
                  data-testid="midi-input"
                >
              </div>
              <label class="checkbox-row">
                <input
                  v-model="valueOptionsEnabled.ignoreTrackers"
                  type="checkbox"
                  data-testid="ignore-trackers-enabled-checkbox"
                  @change="onIgnoreTrackersEnabledChange"
                >
                無視トラッカー（--ignore-trackers=serial1,serial2）
              </label>
              <div
                v-if="valueOptionsEnabled.ignoreTrackers"
                class="option-value-row"
              >
                <input
                  v-model="launchArgs.ignoreTrackers"
                  type="text"
                  placeholder="serial1,serial2"
                  data-testid="ignore-trackers-input"
                >
              </div>
              <div class="render-backend-section">
                <label class="block-label">動画デコーディング</label>
                <div
                  class="toggle-group"
                  role="group"
                  aria-label="動画デコーディング"
                >
                  <label
                    class="toggle-option"
                    :class="{ active: launchArgs.videoDecoding === '' }"
                  >
                    <input
                      v-model="launchArgs.videoDecoding"
                      type="radio"
                      value=""
                      data-testid="video-decoding-default"
                    >
                    <span>既定</span>
                  </label>
                  <label
                    class="toggle-option"
                    :class="{ active: launchArgs.videoDecoding === 'software' }"
                  >
                    <input
                      v-model="launchArgs.videoDecoding"
                      type="radio"
                      value="software"
                      data-testid="video-decoding-software"
                    >
                    <span>ソフトウェア</span>
                  </label>
                  <label
                    class="toggle-option"
                    :class="{ active: launchArgs.videoDecoding === 'hardware' }"
                  >
                    <input
                      v-model="launchArgs.videoDecoding"
                      type="radio"
                      value="hardware"
                      data-testid="video-decoding-hardware"
                    >
                    <span>ハードウェア</span>
                  </label>
                </div>
              </div>
              <label class="checkbox-row">
                <input
                  v-model="launchArgs.disableAMDStutterWorkaround"
                  type="checkbox"
                  data-testid="disable-amd-stutter-workaround-checkbox"
                >
                AMDスタッター回避無効（--disable-amd-stutter-workaround）
              </label>
              <label class="checkbox-row">
                <input
                  v-model="valueOptionsEnabled.osc"
                  type="checkbox"
                  data-testid="osc-enabled-checkbox"
                  @change="onOscEnabledChange"
                >
                OSC（--osc=inPort:outIP:outPort）
              </label>
              <div
                v-if="valueOptionsEnabled.osc"
                class="option-value-row"
              >
                <input
                  v-model="launchArgs.osc"
                  type="text"
                  placeholder="例: 9000:127.0.0.1:9001"
                  data-testid="osc-input"
                >
              </div>
              <label class="checkbox-row">
                <input
                  v-model="valueOptionsEnabled.affinity"
                  type="checkbox"
                  data-testid="affinity-enabled-checkbox"
                  @change="onAffinityEnabledChange"
                >
                スレッドアフィニティ（--affinity=FFFF）
              </label>
              <div
                v-if="valueOptionsEnabled.affinity"
                class="option-value-row"
              >
                <input
                  v-model="launchArgs.affinity"
                  type="text"
                  placeholder="16進ビットマスク"
                  data-testid="affinity-input"
                >
              </div>
            </div>
          </details>
          <label>カスタム引数（上級者向け）</label>
          <input
            v-model="launchArgs.custom"
            type="text"
            placeholder="-batchmode"
            data-testid="custom-args-input"
          >
        </div>
        <label>
          <input
            v-model="selected.isDefault"
            type="checkbox"
          >
          デフォルトに設定
        </label>
        <div class="editor-actions">
          <button
            class="btn-save"
            @click="save"
          >
            保存
          </button>
          <button
            class="btn-launch"
            @click="launch"
          >
            この設定で起動
          </button>
          <button
            v-if="selected.id"
            type="button"
            class="btn-delete"
            data-testid="delete-profile-btn"
            @click="confirmDelete"
          >
            削除
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import {
  App,
  type LaunchProfileDTO,
  type LaunchArgsParsedDTO,
  PRIORITY_OMIT,
} from "../wails/app";

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
const valueOptionsEnabled = reactive<ValueOptionsEnabled>(
  defaultValueOptionsEnabled(),
);

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
    if (launchArgs.value.screenWidth <= 0 && launchArgs.value.screenHeight <= 0)
      launchArgs.value.screenWidth = 1920;
    if (launchArgs.value.screenHeight <= 0)
      launchArgs.value.screenHeight = 1080;
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
  if (!window.confirm(`「${selected.value.name}」を削除しますか？`)) return;
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
}
.btn-add {
  width: 100%;
  padding: 0.5rem;
  margin-bottom: 0.5rem;
  background: var(--bg-tertiary);
  border: 1px dashed var(--border);
  border-radius: var(--radius);
  color: var(--text-secondary);
  cursor: pointer;
}
.btn-add:hover {
  background: var(--bg-secondary);
  color: var(--accent);
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
}
.profile-card:hover,
.profile-card.active {
  background: var(--bg-tertiary);
}
.badge {
  font-size: 0.7rem;
  color: var(--accent);
}
.profile-editor {
  flex: 1;
}
.profile-editor label {
  display: block;
  margin: 0.5rem 0 0.2rem;
  font-size: 0.85rem;
}
.launch-args-gui {
  margin-top: 0.5rem;
}
.details-advanced {
  margin: 0.75rem 0;
  padding: 0.5rem;
  background: var(--bg-tertiary);
  border-radius: var(--radius);
  border: 1px solid var(--border);
}
.details-advanced summary {
  cursor: pointer;
  font-size: 0.9rem;
  color: var(--text-secondary);
}
.launch-args-advanced {
  margin-top: 0.75rem;
  padding-top: 0.5rem;
  border-top: 1px solid var(--border);
}
.screen-mode-section,
.render-backend-section {
  margin: 0.75rem 0;
}
.block-label {
  display: block;
  margin-bottom: 0.4rem;
  font-size: 0.85rem;
}
.toggle-group {
  display: flex;
  gap: 0.25rem;
  flex-wrap: wrap;
}
.toggle-option {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 0.4rem 0.6rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  cursor: pointer;
  transition:
    background 0.15s,
    border-color 0.15s;
}
.toggle-option:first-of-type {
  border-radius: var(--radius) 0 0 var(--radius);
}
.toggle-option:not(:first-of-type):not(:last-of-type) {
  border-radius: 0;
}
.toggle-option:last-of-type {
  border-radius: 0 var(--radius) var(--radius) 0;
}
.toggle-option:first-of-type:last-of-type {
  border-radius: var(--radius);
}
.toggle-option:hover {
  background: var(--bg-secondary);
}
.toggle-option.active {
  background: var(--accent);
  border-color: var(--accent);
  color: white;
}
.toggle-option input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}
.resolution-row {
  margin-top: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}
.resolution-label {
  font-size: 0.85rem;
}
.resolution-inputs {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.resolution-field {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}
.resolution-field-label {
  font-size: 0.75rem;
  color: var(--text-secondary);
}
.option-value-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0.25rem 0 0.75rem 1.5rem;
}
.option-value-row input {
  padding: 0.35rem;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
}
.option-value-row input[type="number"] {
  width: 6rem;
}
.resolution-inputs input,
.screen-mode-section input[type="number"] {
  width: 6rem;
  padding: 0.4rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
}
.resolution-sep {
  color: var(--text-secondary);
  font-size: 0.9rem;
}
.number-row {
  display: flex;
  gap: 1rem;
  margin: 0.5rem 0;
}
.number-row label {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}
.number-row input,
.launch-args-advanced input[type="number"] {
  width: 6rem;
  padding: 0.35rem;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
}
.hint {
  font-size: 0.75rem;
  color: var(--text-secondary);
}
.checkbox-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0.3rem 0;
  cursor: pointer;
}
.checkbox-row input {
  margin: 0;
}
.profile-editor input[type="text"] {
  width: 100%;
  padding: 0.5rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
}
.editor-actions {
  margin-top: 1rem;
  display: flex;
  gap: 0.5rem;
}
.btn-save,
.btn-launch {
  padding: 0.5rem 1rem;
  border-radius: var(--radius);
  border: none;
}
.btn-save {
  background: var(--bg-tertiary);
  color: var(--text-primary);
}
.btn-launch {
  background: var(--accent);
  color: white;
}
.btn-delete {
  margin-left: auto;
  padding: 0.5rem 1rem;
  border-radius: var(--radius);
  border: 1px solid var(--border);
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
}
.btn-delete:hover {
  background: rgba(220, 53, 69, 0.15);
  color: #dc3545;
  border-color: #dc3545;
}
</style>
