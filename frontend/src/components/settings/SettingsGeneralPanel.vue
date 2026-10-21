<template>
  <div data-testid="settings-general-panel">
    <el-card
      class="section-card settings-card--rows"
      shadow="never"
      data-testid="settings-display-section"
    >
      <template #header>
        <span>{{ t("settings.displaySection") }}</span>
      </template>
      <SettingsRow
        :label="t('settings.language')"
        :description="t('settings.languageHint')"
        control="field"
      >
        <el-select
          :model-value="locale"
          data-testid="settings-ui-language"
          @update:model-value="onLanguageChange"
        >
          <el-option value="ja" :label="t('settingsLanguages.ja')" />
          <el-option value="en" :label="t('settingsLanguages.en')" />
          <el-option value="ko" :label="t('settingsLanguages.ko')" />
          <el-option value="zh-TW" :label="t('settingsLanguages.zhTW')" />
          <el-option value="zh-CN" :label="t('settingsLanguages.zhCN')" />
        </el-select>
      </SettingsRow>
    </el-card>

    <el-card
      class="section-card settings-card--rows"
      shadow="never"
      data-testid="settings-window-section"
    >
      <template #header>
        <span>{{ t("settings.windowSection") }}</span>
      </template>
      <SettingsRow
        :label="t('settings.closeToTray')"
        :description="t('settings.closeToTrayHint')"
        control="switch"
      >
        <el-switch
          v-model="closeToTray"
          data-testid="settings-close-to-tray"
          @change="saveCloseToTray"
        />
      </SettingsRow>
    </el-card>

    <el-card
      class="section-card settings-card--rows"
      shadow="never"
      data-testid="settings-power-section"
    >
      <template #header>
        <span>{{ t("settings.powerSection") }}</span>
      </template>
      <SettingsRow
        :label="t('settings.suppressSleep')"
        :description="t('settings.suppressSleepHint')"
        control="switch"
      >
        <el-switch
          v-model="suppressSleepWhileVRChat"
          class="power-switch"
          @change="saveSuppressSleepWhileVRChat"
        />
      </SettingsRow>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { isAppLocale, setLanguage } from "../../i18n";
import { showToast } from "../../utils/showToast";
import { App } from "../../wails/app";
import { formatBackendError } from "../../views/settings/formatBackendError";
import SettingsRow from "./SettingsRow.vue";

const { t, locale } = useI18n();

const suppressSleepWhileVRChat = ref(false);
const closeToTray = ref(true);

onMounted(async () => {
  suppressSleepWhileVRChat.value = await App.getSuppressSleepWhileVRChat();
  closeToTray.value = await App.getCloseToTray();
});

async function onLanguageChange(v: string) {
  if (!isAppLocale(v)) return;
  try {
    await App.setLanguage(v);
  } catch (e) {
    showToast.error(formatBackendError(e, t("settings.errLanguageSave")));
    return;
  }
  setLanguage(v);
}

async function saveSuppressSleepWhileVRChat() {
  await App.setSuppressSleepWhileVRChat(suppressSleepWhileVRChat.value);
}

async function saveCloseToTray() {
  try {
    await App.setCloseToTray(closeToTray.value);
  } catch (e) {
    showToast.error(formatBackendError(e, t("settings.errSaveCloseToTray")));
  }
}
</script>
