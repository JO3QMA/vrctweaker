<template>
  <div data-testid="settings-reset-setup-section">
    <SettingsPageHead
      :title="t('settings.resetSetupSection')"
      :description="t('settings.resetLead')"
    />
    <el-card class="section-card settings-card--rows" shadow="never">
      <SettingsRow
        :label="t('settings.openFirstLaunchGuide')"
        :description="t('settings.openFirstLaunchGuideHint')"
        control="button"
      >
        <VtButton
          variant="secondary"
          data-testid="settings-open-onboarding"
          @click="openFirstLaunchGuide"
        >
          {{ t("settings.openFirstLaunchGuide") }}
        </VtButton>
      </SettingsRow>
      <SettingsRow
        :label="t('settings.vacuum')"
        :description="t('settings.vacuumHint')"
        control="button"
      >
        <VtButton
          variant="secondary"
          :loading="maintenanceLoading"
          data-testid="settings-vacuum"
          @click="doVacuumDb"
        >
          {{
            maintenanceLoading ? t("settings.running") : t("settings.vacuum")
          }}
        </VtButton>
      </SettingsRow>
    </el-card>

    <el-card
      class="section-card settings-card--rows settings-danger-card"
      shadow="never"
      data-testid="settings-destructive-section"
    >
      <template #header>
        <span>{{ t("settings.destructiveSection") }}</span>
      </template>
      <p class="settings-note text-body-sm settings-danger-intro">
        {{ t("settings.resetSetupDestructiveHint") }}
      </p>
      <div v-if="maintenanceError" class="settings-alert-inset">
        <VtAlert
          variant="danger"
          :title="maintenanceError"
          class="settings-inline-alert"
        />
      </div>
      <SettingsRow
        :label="t('settings.clearEncountersItem')"
        :description="t('settings.clearEncountersHint')"
        control="button"
      >
        <VtButton
          variant="danger"
          plain
          :loading="maintenanceLoading"
          data-testid="settings-clear-encounters"
          @click="doClearEncounters"
        >
          {{ t("settings.clearEncounters") }}
        </VtButton>
      </SettingsRow>
      <SettingsRow
        :label="t('settings.clearScreenshotsItem')"
        :description="t('settings.clearScreenshotsHint')"
        control="button"
      >
        <VtButton
          variant="danger"
          plain
          :loading="maintenanceLoading"
          data-testid="settings-clear-screenshots"
          @click="doClearScreenshots"
        >
          {{ t("settings.clearScreenshots") }}
        </VtButton>
      </SettingsRow>
      <SettingsRow
        :label="t('settings.clearFriendsItem')"
        :description="t('settings.clearFriendsHint')"
        control="button"
      >
        <VtButton
          variant="danger"
          plain
          :loading="maintenanceLoading"
          data-testid="settings-clear-friends"
          @click="doClearFriends"
        >
          {{ t("settings.clearFriends") }}
        </VtButton>
      </SettingsRow>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { ElMessageBox } from "element-plus";
import VtAlert from "../VtAlert.vue";
import VtButton from "../VtButton.vue";
import {
  VT_BUTTON_DANGER_CONFIRM_CLASS,
  VT_BUTTON_SECONDARY_CANCEL_CLASS,
} from "../vtButtonClasses";
import { useFirstLaunchGuide } from "../../composables/useFirstLaunchGuide";
import { showToast } from "../../utils/showToast";
import { App } from "../../wails/app";
import SettingsPageHead from "./SettingsPageHead.vue";
import SettingsRow from "./SettingsRow.vue";

const { t } = useI18n();
const { requestFirstLaunchGuide } = useFirstLaunchGuide();

const maintenanceError = ref("");
const maintenanceLoading = ref(false);

function openFirstLaunchGuide(): void {
  requestFirstLaunchGuide();
}

async function runWithConfirm(
  message: string,
  fn: () => Promise<number | void>,
  successMessage?: (result?: number) => string,
) {
  try {
    await ElMessageBox.confirm(message, t("settings.confirmTitle"), {
      confirmButtonText: t("common.execute"),
      cancelButtonText: t("common.cancel"),
      type: "warning",
      confirmButtonClass: VT_BUTTON_DANGER_CONFIRM_CLASS,
      cancelButtonClass: VT_BUTTON_SECONDARY_CANCEL_CLASS,
    });
  } catch {
    return;
  }
  maintenanceError.value = "";
  maintenanceLoading.value = true;
  try {
    const result = await fn();
    const msg = successMessage
      ? successMessage(typeof result === "number" ? result : undefined)
      : t("settings.complete");
    if (msg) {
      showToast.success(msg);
    }
  } catch (e) {
    maintenanceError.value =
      e instanceof Error ? e.message : t("settings.errOperation");
  } finally {
    maintenanceLoading.value = false;
  }
}

function doVacuumDb() {
  void runWithConfirm(
    t("settings.vacuumConfirm"),
    async () => {
      await App.vacuumDb();
    },
    () => t("settings.vacuumDone"),
  );
}

function doClearEncounters() {
  void runWithConfirm(
    t("settings.clearEncountersConfirm"),
    async () => App.clearEncounters(),
    (n) => t("settings.clearEncountersDone", { n: String(n ?? 0) }),
  );
}

function doClearScreenshots() {
  void runWithConfirm(
    t("settings.clearScreenshotsConfirm"),
    async () => App.clearScreenshots(),
    (n) => t("settings.clearScreenshotsDone", { n: String(n ?? 0) }),
  );
}

function doClearFriends() {
  void runWithConfirm(
    t("settings.clearFriendsConfirm"),
    async () => App.clearFriendsCache(),
    (n) => t("settings.clearFriendsDone", { n: String(n ?? 0) }),
  );
}
</script>
