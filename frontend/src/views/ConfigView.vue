<template>
  <div class="config-view">
    <h1 class="page-title">
      その他の設定
    </h1>
    <p class="config-description">
      VRChat の config.json を編集します。 パス:
      <code>%LocalAppData%Low\VRChat\VRChat\config.json</code>
    </p>

    <div
      v-if="!configExists && !editing"
      class="config-not-found"
    >
      <p>
        config.json が見つかりません。新規作成して設定を始めることができます。
      </p>
      <button
        type="button"
        class="btn-primary"
        data-testid="create-config-btn"
        @click="createConfig"
      >
        config.json を作成
      </button>
    </div>

    <div
      v-if="editing"
      class="config-editor"
    >
      <p
        v-if="saveError"
        class="error-message"
      >
        {{ saveError }}
      </p>
      <p
        v-if="saveSuccess"
        class="success-message"
      >
        保存しました
      </p>

      <!-- Camera Resolution -->
      <section class="config-section">
        <h2>カメラ解像度</h2>
        <p class="hint">
          VRカメラで撮影した画像の解像度を設定します（720〜4320px）。
          アプリ内で「Config File」を選択すると反映されます。
        </p>
        <div class="resolution-preset-section">
          <label class="block-label">プリセット</label>
          <div
            class="toggle-group"
            role="group"
            aria-label="カメラ解像度プリセット"
          >
            <label
              class="toggle-option"
              :class="{ active: cameraPreset === 'FHD' }"
            >
              <input
                v-model="cameraPreset"
                type="radio"
                value="FHD"
                data-testid="camera-preset-fhd"
                @change="applyCameraPreset"
              >
              <span>FHD</span>
            </label>
            <label
              class="toggle-option"
              :class="{ active: cameraPreset === 'WQHD' }"
            >
              <input
                v-model="cameraPreset"
                type="radio"
                value="WQHD"
                data-testid="camera-preset-wqhd"
                @change="applyCameraPreset"
              >
              <span>WQHD</span>
            </label>
            <label
              class="toggle-option"
              :class="{ active: cameraPreset === '4K' }"
            >
              <input
                v-model="cameraPreset"
                type="radio"
                value="4K"
                data-testid="camera-preset-4k"
                @change="applyCameraPreset"
              >
              <span>4K</span>
            </label>
            <label
              class="toggle-option"
              :class="{ active: cameraPreset === 'custom' }"
            >
              <input
                v-model="cameraPreset"
                type="radio"
                value="custom"
                data-testid="camera-preset-custom"
                @change="applyCameraPreset"
              >
              <span>手動設定</span>
            </label>
          </div>
          <div class="resolution-fields">
            <label class="resolution-field">
              <span class="resolution-field-label">幅</span>
              <input
                v-model.number="config.cameraResWidth"
                type="number"
                :min="1280"
                :max="7680"
                :disabled="cameraPreset !== 'custom'"
                data-testid="camera-width-input"
              >
            </label>
            <span class="resolution-sep">&times;</span>
            <label class="resolution-field">
              <span class="resolution-field-label">高さ</span>
              <input
                v-model.number="config.cameraResHeight"
                type="number"
                :min="720"
                :max="4320"
                :disabled="cameraPreset !== 'custom'"
                data-testid="camera-height-input"
              >
            </label>
          </div>
        </div>
      </section>

      <!-- Screenshot Resolution -->
      <section class="config-section">
        <h2>スクリーンショット解像度</h2>
        <p class="hint">
          F12キーで撮影するスクリーンショットの解像度を設定します（720〜2160px）。
        </p>
        <div class="resolution-preset-section">
          <label class="block-label">プリセット</label>
          <div
            class="toggle-group"
            role="group"
            aria-label="スクリーンショット解像度プリセット"
          >
            <label
              class="toggle-option"
              :class="{ active: screenshotPreset === 'FHD' }"
            >
              <input
                v-model="screenshotPreset"
                type="radio"
                value="FHD"
                data-testid="screenshot-preset-fhd"
                @change="applyScreenshotPreset"
              >
              <span>FHD</span>
            </label>
            <label
              class="toggle-option"
              :class="{ active: screenshotPreset === 'WQHD' }"
            >
              <input
                v-model="screenshotPreset"
                type="radio"
                value="WQHD"
                data-testid="screenshot-preset-wqhd"
                @change="applyScreenshotPreset"
              >
              <span>WQHD</span>
            </label>
            <label
              class="toggle-option"
              :class="{ active: screenshotPreset === '4K' }"
            >
              <input
                v-model="screenshotPreset"
                type="radio"
                value="4K"
                data-testid="screenshot-preset-4k"
                @change="applyScreenshotPreset"
              >
              <span>4K</span>
            </label>
            <label
              class="toggle-option"
              :class="{ active: screenshotPreset === 'custom' }"
            >
              <input
                v-model="screenshotPreset"
                type="radio"
                value="custom"
                data-testid="screenshot-preset-custom"
                @change="applyScreenshotPreset"
              >
              <span>手動設定</span>
            </label>
          </div>
          <div class="resolution-fields">
            <label class="resolution-field">
              <span class="resolution-field-label">幅</span>
              <input
                v-model.number="config.screenshotResWidth"
                type="number"
                :min="1280"
                :max="3840"
                :disabled="screenshotPreset !== 'custom'"
                data-testid="screenshot-width-input"
              >
            </label>
            <span class="resolution-sep">&times;</span>
            <label class="resolution-field">
              <span class="resolution-field-label">高さ</span>
              <input
                v-model.number="config.screenshotResHeight"
                type="number"
                :min="720"
                :max="2160"
                :disabled="screenshotPreset !== 'custom'"
                data-testid="screenshot-height-input"
              >
            </label>
          </div>
        </div>
      </section>

      <!-- Picture Output -->
      <section class="config-section">
        <h2>写真出力</h2>
        <div class="setting-row">
          <label for="picture-output-folder">出力フォルダ</label>
          <div class="path-input-group">
            <input
              id="picture-output-folder"
              v-model="config.pictureOutputFolder"
              type="text"
              placeholder="デフォルト（空欄で既定パス）"
              data-testid="picture-output-folder-input"
            >
            <button
              type="button"
              class="btn-browse"
              data-testid="picture-output-folder-browse"
              @click="browsePictureOutputFolder"
            >
              参照
            </button>
          </div>
        </div>
        <label class="checkbox-row">
          <input
            v-model="pictureOutputSplitByDate"
            type="checkbox"
            data-testid="picture-split-by-date-checkbox"
          >
          日付別フォルダに分割（YYYY-MM）
        </label>
      </section>

      <!-- Steadycam FOV -->
      <section class="config-section">
        <h2>Steadycam FOV</h2>
        <p class="hint">
          一人称視点 Steadycam の垂直 FOV を設定します（30〜110）。
        </p>
        <div class="setting-row">
          <label for="steadycam-fov">FOV</label>
          <input
            id="steadycam-fov"
            v-model.number="config.fpvSteadycamFov"
            type="number"
            :min="30"
            :max="110"
            placeholder="50"
            data-testid="steadycam-fov-input"
          >
        </div>
      </section>

      <!-- Cache -->
      <section class="config-section">
        <h2>キャッシュ設定</h2>
        <div class="setting-row">
          <label for="cache-directory">キャッシュディレクトリ</label>
          <div class="path-input-group">
            <input
              id="cache-directory"
              v-model="config.cacheDirectory"
              type="text"
              placeholder="デフォルト（空欄で既定パス）"
              data-testid="cache-directory-input"
            >
            <button
              type="button"
              class="btn-browse"
              data-testid="cache-directory-browse"
              @click="browseCacheDirectory"
            >
              参照
            </button>
          </div>
        </div>
        <div class="setting-row">
          <label for="cache-size">キャッシュサイズ上限（GB）</label>
          <input
            id="cache-size"
            v-model.number="config.cacheSize"
            type="number"
            :min="30"
            placeholder="30"
            data-testid="cache-size-input"
          >
        </div>
        <div class="setting-row">
          <label for="cache-expiry">キャッシュ有効期限（日）</label>
          <input
            id="cache-expiry"
            v-model.number="config.cacheExpiryDelay"
            type="number"
            :min="30"
            placeholder="30"
            data-testid="cache-expiry-input"
          >
        </div>
      </section>

      <!-- Rich Presence -->
      <section class="config-section">
        <h2>その他</h2>
        <label class="checkbox-row">
          <input
            v-model="disableRichPresence"
            type="checkbox"
            data-testid="disable-rich-presence-checkbox"
          >
          Discord / Steam Rich Presence を無効にする
        </label>
      </section>

      <!-- Actions -->
      <div class="config-actions">
        <button
          type="button"
          class="btn-primary"
          data-testid="save-config-btn"
          @click="saveConfig"
        >
          保存
        </button>
        <button
          type="button"
          class="btn-danger"
          data-testid="delete-config-btn"
          @click="deleteConfig"
        >
          config.json を削除
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { App } from "../wails/app";
import type { VRChatConfigDTO } from "../wails/app";

type ResolutionPreset = "FHD" | "WQHD" | "4K" | "custom";

interface PresetResolution {
  width: number;
  height: number;
}

const CAMERA_PRESETS: Record<string, PresetResolution> = {
  FHD: { width: 1920, height: 1080 },
  WQHD: { width: 2560, height: 1440 },
  "4K": { width: 3840, height: 2160 },
};

const SCREENSHOT_PRESETS: Record<string, PresetResolution> = {
  FHD: { width: 1920, height: 1080 },
  WQHD: { width: 2560, height: 1440 },
  "4K": { width: 3840, height: 2160 },
};

const configExists = ref(false);
const editing = ref(false);
const saveError = ref("");
const saveSuccess = ref(false);

const config = ref<VRChatConfigDTO>({
  cameraResWidth: 0,
  cameraResHeight: 0,
  screenshotResWidth: 0,
  screenshotResHeight: 0,
  pictureOutputFolder: "",
  pictureOutputSplitByDate: null,
  fpvSteadycamFov: 0,
  cacheDirectory: "",
  cacheSize: 0,
  cacheExpiryDelay: 0,
  disableRichPresence: null,
});

const cameraPreset = ref<ResolutionPreset>("custom");
const screenshotPreset = ref<ResolutionPreset>("custom");
const pictureOutputSplitByDate = ref(true);
const disableRichPresence = ref(false);

function detectPreset(
  width: number,
  height: number,
  presets: Record<string, PresetResolution>,
): ResolutionPreset {
  for (const [key, val] of Object.entries(presets)) {
    if (val.width === width && val.height === height) {
      return key as ResolutionPreset;
    }
  }
  return "custom";
}

function syncFromConfig(cfg: VRChatConfigDTO) {
  config.value = { ...cfg };
  cameraPreset.value = detectPreset(
    cfg.cameraResWidth,
    cfg.cameraResHeight,
    CAMERA_PRESETS,
  );
  screenshotPreset.value = detectPreset(
    cfg.screenshotResWidth,
    cfg.screenshotResHeight,
    SCREENSHOT_PRESETS,
  );
  pictureOutputSplitByDate.value =
    cfg.pictureOutputSplitByDate === null ? true : cfg.pictureOutputSplitByDate;
  disableRichPresence.value =
    cfg.disableRichPresence === null ? false : cfg.disableRichPresence;
}

function applyCameraPreset() {
  const preset = CAMERA_PRESETS[cameraPreset.value];
  if (preset) {
    config.value.cameraResWidth = preset.width;
    config.value.cameraResHeight = preset.height;
  }
}

function applyScreenshotPreset() {
  const preset = SCREENSHOT_PRESETS[screenshotPreset.value];
  if (preset) {
    config.value.screenshotResWidth = preset.width;
    config.value.screenshotResHeight = preset.height;
  }
}

onMounted(async () => {
  configExists.value = await App.vrchatConfigExists();
  if (configExists.value) {
    try {
      const cfg = await App.getVRChatConfig();
      syncFromConfig(cfg);
      editing.value = true;
    } catch {
      configExists.value = false;
    }
  }
});

async function createConfig() {
  saveError.value = "";
  try {
    await App.saveVRChatConfig(config.value);
    configExists.value = true;
    editing.value = true;
  } catch (e) {
    saveError.value =
      e instanceof Error ? e.message : "config.json の作成に失敗しました";
  }
}

async function saveConfig() {
  saveError.value = "";
  saveSuccess.value = false;
  const dto: VRChatConfigDTO = {
    ...config.value,
    pictureOutputSplitByDate: pictureOutputSplitByDate.value,
    disableRichPresence: disableRichPresence.value,
  };
  try {
    await App.saveVRChatConfig(dto);
    saveSuccess.value = true;
    setTimeout(() => {
      saveSuccess.value = false;
    }, 3000);
  } catch (e) {
    saveError.value = e instanceof Error ? e.message : "保存に失敗しました";
  }
}

async function browsePictureOutputFolder() {
  const dir = await App.openDirectoryDialog(
    "写真の出力フォルダを選択",
    config.value.pictureOutputFolder || "",
  );
  if (dir) {
    config.value.pictureOutputFolder = dir;
  }
}

async function browseCacheDirectory() {
  const dir = await App.openDirectoryDialog(
    "キャッシュディレクトリを選択",
    config.value.cacheDirectory || "",
  );
  if (dir) {
    config.value.cacheDirectory = dir;
  }
}

async function deleteConfig() {
  if (!window.confirm("config.json を削除します。よろしいですか？")) {
    return;
  }
  saveError.value = "";
  saveSuccess.value = false;
  try {
    await App.deleteVRChatConfig();
    configExists.value = false;
    editing.value = false;
    config.value = {
      cameraResWidth: 0,
      cameraResHeight: 0,
      screenshotResWidth: 0,
      screenshotResHeight: 0,
      pictureOutputFolder: "",
      pictureOutputSplitByDate: null,
      fpvSteadycamFov: 0,
      cacheDirectory: "",
      cacheSize: 0,
      cacheExpiryDelay: 0,
      disableRichPresence: null,
    };
  } catch (e) {
    saveError.value = e instanceof Error ? e.message : "削除に失敗しました";
  }
}
</script>

<style scoped>
.config-description {
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin-bottom: 1.5rem;
}
.config-description code {
  background: var(--bg-tertiary);
  padding: 0.15rem 0.4rem;
  border-radius: var(--radius);
  font-size: 0.85rem;
}

.config-not-found {
  padding: 2rem;
  text-align: center;
  background: var(--bg-secondary);
  border-radius: var(--radius);
  border: 1px dashed var(--border);
}
.config-not-found p {
  margin-bottom: 1rem;
  color: var(--text-secondary);
}

.config-section {
  margin-bottom: 2rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid var(--border);
}
.config-section:last-of-type {
  border-bottom: none;
}
.config-section h2 {
  font-size: 1.1rem;
  margin: 0 0 0.5rem;
}

.hint {
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin-bottom: 0.75rem;
}

.resolution-preset-section {
  margin-top: 0.5rem;
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
  margin-bottom: 0.75rem;
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

.resolution-fields {
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
.resolution-field input {
  width: 7rem;
  padding: 0.4rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
}
.resolution-field input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.resolution-sep {
  color: var(--text-secondary);
  font-size: 0.9rem;
  margin-top: 1rem;
}

.setting-row {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  margin-bottom: 0.75rem;
}
.setting-row label {
  font-size: 0.9rem;
}
.path-input-group {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  max-width: 480px;
}
.path-input-group input {
  flex: 1;
  min-width: 0;
}
.btn-browse {
  flex-shrink: 0;
  padding: 0.4rem 0.75rem;
  background: var(--accent);
  color: var(--bg-primary);
  border: none;
  border-radius: var(--radius);
  cursor: pointer;
  font-size: 0.9rem;
}
.btn-browse:hover {
  opacity: 0.9;
}
.setting-row input[type="text"] {
  width: 100%;
  max-width: 480px;
  padding: 0.4rem 0.6rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
}
.setting-row input[type="number"] {
  width: 7rem;
  padding: 0.4rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
}

.checkbox-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0.3rem 0;
  cursor: pointer;
  font-size: 0.9rem;
}
.checkbox-row input {
  margin: 0;
}

.config-actions {
  display: flex;
  gap: 0.75rem;
  margin-top: 1.5rem;
}

.btn-primary {
  padding: 0.5rem 1.25rem;
  background: var(--accent);
  color: var(--bg-primary);
  border: none;
  border-radius: var(--radius);
  cursor: pointer;
  font-size: 0.9rem;
}
.btn-primary:hover {
  opacity: 0.9;
}

.btn-danger {
  padding: 0.5rem 1.25rem;
  background: transparent;
  color: var(--error, #ef4444);
  border: 1px solid var(--error, #ef4444);
  border-radius: var(--radius);
  cursor: pointer;
  font-size: 0.9rem;
}
.btn-danger:hover {
  background: rgba(239, 68, 68, 0.1);
}

.error-message {
  font-size: 0.9rem;
  color: var(--error, #ef4444);
  margin: 0 0 1rem;
}
.success-message {
  font-size: 0.9rem;
  color: var(--success, #22c55e);
  margin: 0 0 1rem;
}
</style>
