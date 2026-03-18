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
          <div class="vr-mode-section">
            <label class="block-label">VRモード</label>
            <div
              class="toggle-group"
              role="group"
              aria-label="VRモード"
            >
              <label
                class="toggle-option"
                :class="{ active: launchArgs.vrMode === 'desktop' }"
              >
                <input
                  v-model="launchArgs.vrMode"
                  type="radio"
                  value="desktop"
                  data-testid="vr-mode-desktop"
                >
                <span>デスクトップモード</span>
              </label>
              <label
                class="toggle-option"
                :class="{ active: launchArgs.vrMode === '' }"
              >
                <input
                  v-model="launchArgs.vrMode"
                  type="radio"
                  value=""
                  data-testid="vr-mode-none"
                >
                <span>無設定</span>
              </label>
              <label
                class="toggle-option"
                :class="{ active: launchArgs.vrMode === 'vr' }"
              >
                <input
                  v-model="launchArgs.vrMode"
                  type="radio"
                  value="vr"
                  data-testid="vr-mode-vr"
                >
                <span>強制VRモード</span>
              </label>
            </div>
          </div>
          <label class="checkbox-row">
            <input
              v-model="launchArgs.clearCache"
              type="checkbox"
              data-testid="clear-cache-checkbox"
            >
            起動前にキャッシュをクリア
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
            <div class="resolution-row">
              <label class="resolution-label">解像度</label>
              <div class="resolution-inputs">
                <input
                  v-model.number="launchArgs.screenWidth"
                  type="number"
                  min="0"
                  placeholder="幅"
                  data-testid="screen-width-input"
                >
                <span class="resolution-sep">×</span>
                <input
                  v-model.number="launchArgs.screenHeight"
                  type="number"
                  min="0"
                  placeholder="高さ"
                  data-testid="screen-height-input"
                >
              </div>
            </div>
          </div>
          <details class="details-advanced">
            <summary>詳細設定</summary>
            <div class="launch-args-advanced">
              <label class="checkbox-row">
                <input
                  v-model="launchArgs.fpfc"
                  type="checkbox"
                  data-testid="fpfc-checkbox"
                >
                FPFC（-fpfc）ワールド制作用カメラ
              </label>
              <label>
                FPS制限（--fps=N）
                <input
                  v-model.number="launchArgs.fps"
                  type="number"
                  min="0"
                  placeholder="0=省略"
                  data-testid="fps-input"
                >
              </label>
              <label class="checkbox-row">
                <input
                  v-model="launchArgs.safe"
                  type="checkbox"
                  data-testid="safe-checkbox"
                >
                セーフモード（-safe）
              </label>
              <label class="checkbox-row">
                <input
                  v-model="launchArgs.noSplash"
                  type="checkbox"
                  data-testid="nosplash-checkbox"
                >
                スプラッシュスキップ（-nosplash）
              </label>
              <label class="checkbox-row">
                <input
                  v-model="launchArgs.noAudio"
                  type="checkbox"
                  data-testid="noaudio-checkbox"
                >
                オーディオ無効（-noaudio）
              </label>
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
                  v-model="launchArgs.forceD3d11"
                  type="checkbox"
                  data-testid="force-d3d11-checkbox"
                >
                DirectX 11強制（-force-d3d11）
              </label>
              <label class="checkbox-row">
                <input
                  v-model="launchArgs.forceVulkan"
                  type="checkbox"
                  data-testid="force-vulkan-checkbox"
                >
                Vulkan強制（-force-vulkan）
              </label>
              <label class="checkbox-row">
                <input
                  v-model="launchArgs.log"
                  type="checkbox"
                  data-testid="log-checkbox"
                >
                ログ出力（-log）
              </label>
              <label>
                プロセス優先度（--process-priority=N）
                <input
                  v-model.number="launchArgs.processPriority"
                  type="number"
                  min="0"
                  placeholder="0=省略、2=高"
                  data-testid="process-priority-input"
                >
              </label>
            </div>
          </details>
          <label>カスタム引数（上級者向け）</label>
          <input
            v-model="launchArgs.custom"
            type="text"
            placeholder="-batchmode -nographics"
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
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import {
  App,
  type LaunchProfileDTO,
  type LaunchArgsParsedDTO,
} from "../wails/app";

const defaultLaunchArgs = (): LaunchArgsParsedDTO => ({
  vrMode: "",
  clearCache: false,
  screenMode: "",
  fpfc: false,
  screenWidth: 0,
  screenHeight: 0,
  fps: 0,
  safe: false,
  noSplash: false,
  noAudio: false,
  skipRegistry: false,
  forceD3d11: false,
  forceVulkan: false,
  log: false,
  processPriority: 0,
  custom: "",
});

const profiles = ref<LaunchProfileDTO[]>([]);
const selected = ref<LaunchProfileDTO | null>(null);
const launchArgs = ref<LaunchArgsParsedDTO>(defaultLaunchArgs());

async function syncLaunchArgsFromProfile(p: LaunchProfileDTO) {
  launchArgs.value = await App.parseLaunchArgsForGUI(p.arguments);
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
}

function sanitizeLaunchArgs(a: LaunchArgsParsedDTO): LaunchArgsParsedDTO {
  return {
    ...a,
    screenWidth: Math.max(0, Number(a.screenWidth) || 0),
    screenHeight: Math.max(0, Number(a.screenHeight) || 0),
    fps: Math.max(0, Number(a.fps) || 0),
    processPriority: Math.max(0, Number(a.processPriority) || 0),
  };
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
.vr-mode-section,
.screen-mode-section {
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
.resolution-inputs input {
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
</style>
