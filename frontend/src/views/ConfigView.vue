<template>
  <div class="config-view">
    <h1 class="page-title">その他の設定</h1>
    <el-text
      type="info"
      size="small"
      style="display: block; margin-bottom: 1.5rem"
    >
      VRChat の config.json を編集します。 パス:
      <code>%LocalAppData%Low\VRChat\VRChat\config.json</code>
    </el-text>

    <el-card
      v-if="!configExists && !editing"
      shadow="never"
      class="config-card"
    >
      <div class="config-not-found">
        <el-text type="info">
          config.json が見つかりません。新規作成して設定を始めることができます。
        </el-text>
        <el-button
          type="primary"
          data-testid="create-config-btn"
          style="margin-top: 1rem"
          @click="createConfig"
        >
          config.json を作成
        </el-button>
      </div>
    </el-card>

    <div v-if="editing" class="config-editor">
      <el-alert
        v-if="saveError"
        :title="saveError"
        type="error"
        :closable="false"
        show-icon
        style="margin-bottom: 1rem"
      />
      <el-alert
        v-if="saveSuccess"
        title="保存しました"
        type="success"
        :closable="false"
        show-icon
        style="margin-bottom: 1rem"
      />

      <!-- Camera Resolution -->
      <el-card shadow="never" class="config-card">
        <template #header>カメラ解像度</template>
        <el-text
          type="info"
          size="small"
          style="display: block; margin-bottom: 0.75rem"
        >
          VRカメラで撮影した画像の解像度を設定します（1280px × 720px ～ 7680px ×
          4320px）。VRChat内でカスタム解像度を選択すると反映されます。
        </el-text>
        <div class="resolution-section">
          <el-radio-group
            v-model="cameraPreset"
            size="small"
            style="flex-wrap: wrap; gap: 4px"
          >
            <el-radio-button
              value="HD"
              data-testid="camera-preset-hd"
              @change="applyCameraPreset"
              >HD</el-radio-button
            >
            <el-radio-button
              value="FHD"
              data-testid="camera-preset-fhd"
              @change="applyCameraPreset"
              >FHD</el-radio-button
            >
            <el-radio-button
              value="WQHD"
              data-testid="camera-preset-wqhd"
              @change="applyCameraPreset"
              >WQHD</el-radio-button
            >
            <el-radio-button
              value="4K"
              data-testid="camera-preset-4k"
              @change="applyCameraPreset"
              >4K</el-radio-button
            >
            <el-radio-button
              value="8K"
              data-testid="camera-preset-8k"
              @change="applyCameraPreset"
              >8K</el-radio-button
            >
            <el-radio-button
              value="custom"
              data-testid="camera-preset-custom"
              @change="applyCameraPreset"
              >手動設定</el-radio-button
            >
          </el-radio-group>
          <div class="resolution-fields">
            <el-input-number
              v-model="config.cameraResWidth"
              :min="1280"
              :max="7680"
              :disabled="cameraPreset !== 'custom'"
              data-testid="camera-width-input"
              size="small"
              placeholder="幅"
              style="width: 130px"
            />
            <span class="resolution-sep">×</span>
            <el-input-number
              v-model="config.cameraResHeight"
              :min="720"
              :max="4320"
              :disabled="cameraPreset !== 'custom'"
              data-testid="camera-height-input"
              size="small"
              placeholder="高さ"
              style="width: 130px"
            />
          </div>
        </div>
      </el-card>

      <!-- Screenshot Resolution -->
      <el-card shadow="never" class="config-card">
        <template #header>スクリーンショット解像度</template>
        <el-text
          type="info"
          size="small"
          style="display: block; margin-bottom: 0.75rem"
        >
          F12キーで撮影するスクリーンショットの解像度を設定します（1280px ×
          720px ～ 3840px × 2160px）。
        </el-text>
        <div class="resolution-section">
          <el-radio-group
            v-model="screenshotPreset"
            size="small"
            style="flex-wrap: wrap; gap: 4px"
          >
            <el-radio-button
              value="HD"
              data-testid="screenshot-preset-hd"
              @change="applyScreenshotPreset"
              >HD</el-radio-button
            >
            <el-radio-button
              value="FHD"
              data-testid="screenshot-preset-fhd"
              @change="applyScreenshotPreset"
              >FHD</el-radio-button
            >
            <el-radio-button
              value="WQHD"
              data-testid="screenshot-preset-wqhd"
              @change="applyScreenshotPreset"
              >WQHD</el-radio-button
            >
            <el-radio-button
              value="4K"
              data-testid="screenshot-preset-4k"
              @change="applyScreenshotPreset"
              >4K</el-radio-button
            >
            <el-radio-button
              value="custom"
              data-testid="screenshot-preset-custom"
              @change="applyScreenshotPreset"
              >手動設定</el-radio-button
            >
          </el-radio-group>
          <div class="resolution-fields">
            <el-input-number
              v-model="config.screenshotResWidth"
              :min="1280"
              :max="3840"
              :disabled="screenshotPreset !== 'custom'"
              data-testid="screenshot-width-input"
              size="small"
              placeholder="幅"
              style="width: 130px"
            />
            <span class="resolution-sep">×</span>
            <el-input-number
              v-model="config.screenshotResHeight"
              :min="720"
              :max="2160"
              :disabled="screenshotPreset !== 'custom'"
              data-testid="screenshot-height-input"
              size="small"
              placeholder="高さ"
              style="width: 130px"
            />
          </div>
        </div>
      </el-card>

      <!-- Picture Output -->
      <el-card shadow="never" class="config-card">
        <template #header>写真出力</template>
        <el-form label-position="top" size="default">
          <el-form-item label="出力フォルダ">
            <div class="path-row">
              <el-input
                id="picture-output-folder"
                v-model="config.pictureOutputFolder"
                placeholder="デフォルト（空欄で既定パス）"
                data-testid="picture-output-folder-input"
              />
              <el-button
                data-testid="picture-output-folder-browse"
                @click="browsePictureOutputFolder"
              >
                参照
              </el-button>
            </div>
          </el-form-item>
          <el-form-item>
            <el-checkbox
              v-model="pictureOutputSplitByDate"
              data-testid="picture-split-by-date-checkbox"
            >
              日付別フォルダに分割（YYYY-MM）
            </el-checkbox>
          </el-form-item>
        </el-form>
      </el-card>

      <!-- Steadycam FOV -->
      <el-card shadow="never" class="config-card">
        <template #header>Steadycam FOV</template>
        <el-text
          type="info"
          size="small"
          style="display: block; margin-bottom: 0.75rem"
        >
          一人称視点 Steadycam の垂直 FOV を設定します（30〜100）。
        </el-text>
        <div class="fov-row">
          <el-slider
            :model-value="steadycamFovSliderValue"
            :min="STEADYCAM_FOV_MIN"
            :max="STEADYCAM_FOV_MAX"
            style="flex: 1; max-width: 240px"
            data-testid="steadycam-fov-slider"
            @input="onSteadycamFovSliderInput"
          />
          <el-input-number
            id="steadycam-fov"
            :min="STEADYCAM_FOV_MIN"
            :max="STEADYCAM_FOV_MAX"
            :model-value="config.fpvSteadycamFov || undefined"
            :placeholder="String(STEADYCAM_FOV_PLACEHOLDER)"
            data-testid="steadycam-fov-input"
            size="small"
            style="width: 100px"
            @change="onSteadycamFovChange"
            @blur="clampSteadycamFov"
          />
        </div>
      </el-card>

      <!-- Cache -->
      <el-card shadow="never" class="config-card">
        <template #header>キャッシュ設定</template>
        <el-text
          type="info"
          size="small"
          style="display: block; margin-bottom: 0.75rem"
        >
          キャッシュサイズと有効期限は30以上を指定してください。
        </el-text>
        <el-form label-position="top" size="default">
          <el-form-item label="キャッシュディレクトリ">
            <div class="path-row">
              <el-input
                id="cache-directory"
                v-model="config.cacheDirectory"
                placeholder="デフォルト（空欄で既定パス）"
                data-testid="cache-directory-input"
              />
              <el-button
                data-testid="cache-directory-browse"
                @click="browseCacheDirectory"
              >
                参照
              </el-button>
            </div>
          </el-form-item>
          <el-form-item label="キャッシュサイズ上限（GB）">
            <el-input-number
              id="cache-size"
              v-model="config.cacheSize"
              :min="30"
              :step="1"
              placeholder="30"
              data-testid="cache-size-input"
              @blur="clampCacheSize"
            />
          </el-form-item>
          <el-form-item label="キャッシュ有効期限（日）">
            <el-input-number
              id="cache-expiry"
              v-model="config.cacheExpiryDelay"
              :min="30"
              :step="1"
              placeholder="30"
              data-testid="cache-expiry-input"
              @blur="clampCacheExpiry"
            />
          </el-form-item>
        </el-form>
      </el-card>

      <!-- Rich Presence -->
      <el-card shadow="never" class="config-card">
        <template #header>その他</template>
        <el-checkbox
          v-model="disableRichPresence"
          data-testid="disable-rich-presence-checkbox"
        >
          Discord / Steam Rich Presence を無効にする
        </el-checkbox>
      </el-card>

      <!-- Actions -->
      <div class="config-actions">
        <el-button
          type="primary"
          data-testid="save-config-btn"
          @click="saveConfig"
        >
          保存
        </el-button>
        <el-button
          type="danger"
          plain
          data-testid="delete-config-btn"
          @click="deleteConfig"
        >
          config.json を削除
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { ElMessageBox } from "element-plus";
import { App } from "../wails/app";
import type { VRChatConfigDTO } from "../wails/app";

type ResolutionPreset = "HD" | "FHD" | "WQHD" | "4K" | "8K" | "custom";

interface PresetResolution {
  width: number;
  height: number;
}

const CAMERA_PRESETS: Record<string, PresetResolution> = {
  HD: { width: 1280, height: 720 },
  FHD: { width: 1920, height: 1080 },
  WQHD: { width: 2560, height: 1440 },
  "4K": { width: 3840, height: 2160 },
  "8K": { width: 7680, height: 4320 },
};

const SCREENSHOT_PRESETS: Record<string, PresetResolution> = {
  HD: { width: 1280, height: 720 },
  FHD: { width: 1920, height: 1080 },
  WQHD: { width: 2560, height: 1440 },
  "4K": { width: 3840, height: 2160 },
};

const configExists = ref(false);
const editing = ref(false);
const saveError = ref("");
const saveSuccess = ref(false);

const CACHE_MIN = 30;
const STEADYCAM_FOV_MIN = 30;
const STEADYCAM_FOV_MAX = 100;
const STEADYCAM_FOV_PLACEHOLDER = 50;

const config = ref<VRChatConfigDTO>({
  cameraResWidth: 0,
  cameraResHeight: 0,
  screenshotResWidth: 0,
  screenshotResHeight: 0,
  pictureOutputFolder: "",
  pictureOutputSplitByDate: null,
  fpvSteadycamFov: 0,
  cacheDirectory: "",
  cacheSize: CACHE_MIN,
  cacheExpiryDelay: CACHE_MIN,
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
  const fpvFov = cfg.fpvSteadycamFov;
  const fpvFovNorm =
    fpvFov >= STEADYCAM_FOV_MIN && fpvFov <= STEADYCAM_FOV_MAX ? fpvFov : 0;

  config.value = {
    ...cfg,
    fpvSteadycamFov: fpvFovNorm,
    cacheSize: cfg.cacheSize < CACHE_MIN ? CACHE_MIN : cfg.cacheSize,
    cacheExpiryDelay:
      cfg.cacheExpiryDelay < CACHE_MIN ? CACHE_MIN : cfg.cacheExpiryDelay,
  };
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

function clampCacheSize() {
  if (config.value.cacheSize < CACHE_MIN) {
    config.value.cacheSize = CACHE_MIN;
  }
}

function clampCacheExpiry() {
  if (config.value.cacheExpiryDelay < CACHE_MIN) {
    config.value.cacheExpiryDelay = CACHE_MIN;
  }
}

const steadycamFovSliderValue = computed(() => {
  const v = config.value.fpvSteadycamFov;
  if (v === 0) return STEADYCAM_FOV_PLACEHOLDER;
  return Math.max(STEADYCAM_FOV_MIN, Math.min(STEADYCAM_FOV_MAX, v));
});

function onSteadycamFovSliderInput(val: number) {
  const n = Math.round(val);
  config.value.fpvSteadycamFov = Math.max(
    STEADYCAM_FOV_MIN,
    Math.min(STEADYCAM_FOV_MAX, n),
  );
}

function onSteadycamFovChange(val: number | undefined) {
  if (val === undefined || val === null) {
    config.value.fpvSteadycamFov = 0;
    return;
  }
  config.value.fpvSteadycamFov = Math.round(val);
}

function clampSteadycamFov() {
  const v = config.value.fpvSteadycamFov;
  if (v > 0 && v < STEADYCAM_FOV_MIN) {
    config.value.fpvSteadycamFov = STEADYCAM_FOV_MIN;
  } else if (v > STEADYCAM_FOV_MAX) {
    config.value.fpvSteadycamFov = STEADYCAM_FOV_MAX;
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
  const cacheSize = Math.max(CACHE_MIN, config.value.cacheSize);
  const cacheExpiryDelay = Math.max(CACHE_MIN, config.value.cacheExpiryDelay);
  const dto: VRChatConfigDTO = {
    ...config.value,
    cacheSize,
    cacheExpiryDelay,
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
  try {
    await ElMessageBox.confirm(
      "config.json を削除します。よろしいですか？",
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
      cacheSize: CACHE_MIN,
      cacheExpiryDelay: CACHE_MIN,
      disableRichPresence: null,
    };
  } catch (e) {
    saveError.value = e instanceof Error ? e.message : "削除に失敗しました";
  }
}
</script>

<style scoped>
.config-card {
  margin-bottom: 1.25rem;
  background: var(--bg-secondary) !important;
  border-color: var(--border) !important;
}

.config-card :deep(.el-card__header) {
  font-weight: 600;
  border-bottom-color: var(--border);
}

.config-not-found {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 1.5rem;
}

.resolution-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.resolution-fields {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.resolution-sep {
  color: var(--text-secondary);
}

.path-row {
  display: flex;
  gap: 0.5rem;
  width: 100%;
  max-width: 480px;
}

.path-row :deep(.el-input) {
  flex: 1;
}

.fov-row {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.config-actions {
  display: flex;
  gap: 0.75rem;
  margin-top: 0.5rem;
}
</style>
