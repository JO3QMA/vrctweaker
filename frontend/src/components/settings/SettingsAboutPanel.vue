<template>
  <div data-testid="settings-about-app-section">
    <el-card class="section-card settings-card--rows" shadow="never">
      <template #header>
        <span>{{ t("settings.aboutAppSection") }}</span>
      </template>
      <p
        class="settings-version text-body-sm"
        data-testid="settings-app-version"
      >
        {{ t("settings.appVersion", { version: appVersion }) }}
      </p>
      <SettingsRow :label="t('settings.updateCheckLabel')" control="button">
        <VtButton
          variant="secondary"
          :loading="appUpdateCheckLoading"
          data-testid="settings-check-app-update"
          @click="checkAppUpdate"
        >
          {{ t("settings.checkAppUpdate") }}
        </VtButton>
      </SettingsRow>
      <p
        v-if="appUpdateStatusLine"
        class="settings-note text-body-sm"
        data-testid="settings-app-update-status"
      >
        {{ appUpdateStatusLine }}
      </p>
      <VtAlert
        v-if="appUpdateCheckError"
        variant="danger"
        :title="appUpdateCheckError"
        class="settings-inline-alert"
      />
      <SettingsRow
        :label="t('settings.updateAutoCheck')"
        :description="t('settings.updateAutoCheckHint')"
        control="switch"
      >
        <el-switch
          v-model="appUpdateAutoCheck"
          data-testid="settings-app-update-auto-check"
          @change="saveAppUpdateAutoCheck"
        />
      </SettingsRow>
      <div class="settings-about-footer">
        <p
          class="settings-note text-body-sm"
          data-testid="settings-app-disclaimer"
        >
          {{ t("common.unofficialDisclaimer") }}
        </p>
        <LegalDocLinks stacked show-oss-licenses-link />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import LegalDocLinks from "../LegalDocLinks.vue";
import VtAlert from "../VtAlert.vue";
import VtButton from "../VtButton.vue";
import { useAppUpdate } from "../../composables/useAppUpdate";
import { showToast } from "../../utils/showToast";
import { App } from "../../wails/app";
import { formatBackendError } from "../../views/settings/formatBackendError";
import SettingsRow from "./SettingsRow.vue";

const { t } = useI18n();
const { checkManual } = useAppUpdate();

const appVersion = ref("");
const appUpdateAutoCheck = ref(true);
const appUpdateCheckLoading = ref(false);
const appUpdateCheckError = ref("");
const appUpdateStatusLine = ref("");

onMounted(async () => {
  try {
    appUpdateAutoCheck.value = await App.getAppUpdateAutoCheck();
  } catch {
    appUpdateAutoCheck.value = true;
  }
  try {
    appVersion.value = await App.getAppVersion();
  } catch {
    appVersion.value = "dev";
  }
});

async function saveAppUpdateAutoCheck() {
  const attempted = appUpdateAutoCheck.value;
  try {
    await App.setAppUpdateAutoCheck(attempted);
  } catch (e) {
    appUpdateAutoCheck.value = !attempted;
    showToast.error(formatBackendError(e, t("settings.errUpdateAutoCheck")));
  }
}

async function checkAppUpdate() {
  appUpdateCheckError.value = "";
  appUpdateStatusLine.value = "";
  appUpdateCheckLoading.value = true;
  try {
    const st = await checkManual();
    if (st.updateAvailable) {
      appUpdateStatusLine.value = t("settings.appUpdateAvailable", {
        version: st.latestVersion,
      });
    } else {
      appUpdateStatusLine.value = t("settings.appUpdateUpToDate");
    }
  } catch {
    appUpdateCheckError.value = t("settings.errAppUpdateCheck");
  } finally {
    appUpdateCheckLoading.value = false;
  }
}
</script>
