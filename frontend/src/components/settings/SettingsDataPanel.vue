<template>
  <div data-testid="settings-data-panel">
    <el-card class="section-card settings-card--rows" shadow="never">
      <template #header>
        <span>{{ t("settings.pathSection") }}</span>
      </template>
      <SettingsPathRow
        v-for="field in pathFields"
        :key="field.id"
        :label="field.label"
      >
        <el-input
          v-model="pathSettings[field.id]"
          :placeholder="field.placeholder"
          @change="savePathSettings"
        />
        <div class="path-input-actions">
          <VtButton
            variant="secondary"
            :data-testid="field.browseTestId"
            :title="field.browseTitle"
            @click="field.browse"
          >
            {{ t("common.browse") }}
          </VtButton>
          <VtButton
            variant="secondary"
            :data-testid="`path-validate-${field.id}`"
            :disabled="!pathSettings[field.id]"
            @click="validatePathField(field.id)"
          >
            {{ t("settings.validateExists") }}
          </VtButton>
        </div>
        <template
          v-if="
            validateResult[field.id] !== null || field.id === 'outputLogPath'
          "
          #below
        >
          <el-text
            v-if="validateResult[field.id] !== null"
            :type="validateResult[field.id] ? 'success' : 'danger'"
            size="small"
          >
            {{
              validateResult[field.id]
                ? t("settings.existsYes")
                : t("settings.existsNo")
            }}
          </el-text>
          <a
            v-if="field.id === 'outputLogPath'"
            href="#"
            class="path-open-log-link"
            data-testid="settings-open-log-folder"
            :title="t('settings.titleOpenLogFolder')"
            @click.prevent="openVRChatLogFolder"
          >
            {{ t("settings.openLogFolder") }}
          </a>
        </template>
      </SettingsPathRow>
      <p class="settings-note text-body-sm">{{ t("settings.pathHint") }}</p>
    </el-card>

    <el-card class="section-card settings-card--rows" shadow="never">
      <template #header>
        <span>{{ t("settings.dataSection") }}</span>
      </template>
      <SettingsRow
        :label="t('settings.retentionLabel')"
        :description="t('settings.retentionHint')"
        control="number"
      >
        <el-input-number
          v-model="logRetentionDays"
          :min="1"
          :max="365"
          data-testid="settings-log-retention"
          @change="saveRetention"
        />
      </SettingsRow>
      <SettingsRow
        :label="t('settings.galleryAutoEnrich')"
        :description="t('settings.galleryAutoEnrichHint')"
        control="switch"
      >
        <el-switch
          v-model="galleryAutoEnrichMetadata"
          data-testid="settings-gallery-auto-enrich"
          @change="saveGalleryAutoEnrich"
        />
      </SettingsRow>
      <p
        class="settings-note text-body-sm"
        data-testid="settings-app-data-storage"
      >
        {{ t("settings.dataStorageNote") }}
      </p>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import VtButton from "../VtButton.vue";
import { showToast } from "../../utils/showToast";
import { App } from "../../wails/app";
import type { PathSettingsDTO } from "../../wails/app";
import { formatBackendError } from "../../views/settings/formatBackendError";
import SettingsPathRow from "./SettingsPathRow.vue";
import SettingsRow from "./SettingsRow.vue";

const { t } = useI18n();

const logRetentionDays = ref(30);
const galleryAutoEnrichMetadata = ref(true);
const pathSettings = reactive<PathSettingsDTO>({
  vrchatPathWindows: "",
  steamPathLinux: "",
  outputLogPath: "",
});

const validateResult = reactive<Record<keyof PathSettingsDTO, boolean | null>>({
  vrchatPathWindows: null,
  steamPathLinux: null,
  outputLogPath: null,
});

const pathFields = computed(() => [
  {
    id: "vrchatPathWindows" as keyof PathSettingsDTO,
    label: t("settings.pathVrchatWin"),
    placeholder: t("settings.pathVrchatWinPh"),
    browseTestId: "vrchat-path-browse",
    browseTitle: t("settings.titlePickVrchatExe"),
    browse: browseVrchatPath,
  },
  {
    id: "steamPathLinux" as keyof PathSettingsDTO,
    label: t("settings.pathSteamLinux"),
    placeholder: t("settings.pathSteamLinuxPh"),
    browseTestId: "steam-path-browse",
    browseTitle: t("settings.titlePickSteam"),
    browse: browseSteamPath,
  },
  {
    id: "outputLogPath" as keyof PathSettingsDTO,
    label: t("settings.pathOutputLog"),
    placeholder: t("settings.pathOutputLogPh"),
    browseTestId: "output-log-dir-browse",
    browseTitle: t("settings.titlePickLogDir"),
    browse: browseOutputLogDirectory,
  },
]);

onMounted(async () => {
  const ps = await App.getPathSettings();
  pathSettings.vrchatPathWindows = ps.vrchatPathWindows;
  pathSettings.steamPathLinux = ps.steamPathLinux;
  pathSettings.outputLogPath = ps.outputLogPath;
  logRetentionDays.value = await App.getLogRetentionDays();
  galleryAutoEnrichMetadata.value = await App.getGalleryAutoEnrichMetadata();
});

async function saveRetention() {
  await App.setLogRetentionDays(logRetentionDays.value);
}

async function saveGalleryAutoEnrich() {
  const attempted = galleryAutoEnrichMetadata.value;
  try {
    await App.setGalleryAutoEnrichMetadata(attempted);
  } catch (e) {
    galleryAutoEnrichMetadata.value = !attempted;
    showToast.error(formatBackendError(e, t("settings.errOperation")));
  }
}

async function savePathSettings() {
  try {
    await App.setPathSettings(pathSettings);
  } catch (e) {
    showToast.error(formatBackendError(e, t("settings.errOperation")));
  }
}

function dirOfPath(p: string): string {
  if (!p) return "";
  const sep = p.includes("\\") ? "\\" : "/";
  const idx = p.lastIndexOf(sep);
  return idx >= 0 ? p.slice(0, idx) : "";
}

async function browseVrchatPath() {
  const path = await App.openFileDialog(
    t("settings.titlePickVrchatExe"),
    dirOfPath(pathSettings.vrchatPathWindows),
    "*.exe",
  );
  if (path) {
    pathSettings.vrchatPathWindows = path;
    await savePathSettings();
  }
}

async function browseSteamPath() {
  const path = await App.openFileDialog(
    t("settings.titlePickSteam"),
    dirOfPath(pathSettings.steamPathLinux),
    "",
  );
  if (path) {
    pathSettings.steamPathLinux = path;
    await savePathSettings();
  }
}

async function browseOutputLogDirectory() {
  const dir = await App.openDirectoryDialog(
    t("settings.titlePickLogDirShort"),
    dirOfPath(pathSettings.outputLogPath),
  );
  if (dir) {
    pathSettings.outputLogPath = dir;
    await savePathSettings();
  }
}

async function openVRChatLogFolder(): Promise<void> {
  try {
    await App.openVRChatLogFolder();
  } catch (err) {
    console.error(err);
  }
}

async function validatePathField(field: keyof PathSettingsDTO) {
  const path = pathSettings[field];
  if (path === "") {
    validateResult[field] = false;
    return;
  }
  if (field === "outputLogPath") {
    validateResult[field] = await App.validateOutputLogPath(path);
    return;
  }
  validateResult[field] = await App.validatePath(path);
}
</script>
