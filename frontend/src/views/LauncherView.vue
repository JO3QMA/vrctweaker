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
          <label class="checkbox-row">
            <input
              v-model="launchArgs.clearCache"
              type="checkbox"
              data-testid="clear-cache-checkbox"
            >
            起動前にキャッシュをクリア
          </label>
          <label class="checkbox-row">
            <input
              v-model="launchArgs.fullscreen"
              type="checkbox"
              data-testid="fullscreen-checkbox"
            >
            フルスクリーン（-screen-fullscreen 1）
          </label>
          <details class="details-advanced">
            <summary>詳細設定</summary>
            <div class="launch-args-advanced">
              <label class="checkbox-row">
                <input
                  v-model="launchArgs.vr"
                  type="checkbox"
                  data-testid="vr-checkbox"
                >
                強制VRモード（-vr）
              </label>
              <label class="checkbox-row">
                <input
                  v-model="launchArgs.fpfc"
                  type="checkbox"
                  data-testid="fpfc-checkbox"
                >
                FPFC（-fpfc）ワールド制作用カメラ
              </label>
              <label class="checkbox-row">
                <input
                  v-model="launchArgs.windowed"
                  type="checkbox"
                  data-testid="windowed-checkbox"
                >
                ウィンドウモード（-windowed）
              </label>
              <div class="number-row">
                <label>
                  解像度
                  <span class="hint">幅</span>
                  <input
                    v-model.number="launchArgs.screenWidth"
                    type="number"
                    min="0"
                    placeholder="0=省略"
                    data-testid="screen-width-input"
                  >
                </label>
                <label>
                  <span class="hint">高さ</span>
                  <input
                    v-model.number="launchArgs.screenHeight"
                    type="number"
                    min="0"
                    placeholder="0=省略"
                    data-testid="screen-height-input"
                  >
                </label>
              </div>
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
  noVr: false,
  clearCache: false,
  fullscreen: false,
  vr: false,
  fpfc: false,
  windowed: false,
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
