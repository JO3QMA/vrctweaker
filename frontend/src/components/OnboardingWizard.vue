<template>
  <div
    class="onboarding-wizard"
    data-testid="onboarding-wizard"
    role="dialog"
    aria-modal="true"
    :aria-label="t('onboarding.title')"
  >
    <header class="onboarding-header">
      <div class="progress-dots" data-testid="onboarding-progress">
        <span
          v-for="i in 3"
          :key="i"
          class="progress-dot"
          :class="{ 'progress-dot--active': step === i - 1 }"
        />
      </div>
      <button
        type="button"
        class="skip-all-link text-body-sm"
        data-testid="onboarding-skip-all"
        @click="skipAll"
      >
        {{ t("onboarding.skipAll") }}
      </button>
    </header>

    <main class="onboarding-main">
      <h1 class="onboarding-title text-h2">{{ stepTitle }}</h1>

      <section
        v-if="step === 0"
        class="onboarding-step"
        data-testid="onboarding-step-login"
      >
        <p class="onboarding-lead text-body-sm">
          {{ t("onboarding.loginLead") }}
        </p>
        <p class="onboarding-keyring-hint text-caption">
          {{ t("onboarding.loginKeyringHint") }}
        </p>
        <div v-if="loggedIn" class="login-done">
          <VtTag variant="success" size="small">{{
            t("settings.loggedInTag")
          }}</VtTag>
          <p v-if="displayName" class="login-display-name text-body">
            {{ displayName }}
          </p>
        </div>
        <el-form v-else label-position="top" size="default" class="login-form">
          <VtAlert
            v-if="unlockState === 'needs-relogin' && unlockErrorMessage"
            variant="warning"
            :title="unlockErrorMessage"
            class="login-error"
          />
          <el-form-item :label="t('settings.username')">
            <VtInput
              v-model="loginForm.username"
              data-testid="onboarding-login-username"
              :placeholder="t('settings.usernamePh')"
              autocomplete="username"
            />
          </el-form-item>
          <el-form-item :label="t('settings.password')">
            <VtInput
              v-model="loginForm.password"
              data-testid="onboarding-login-password"
              type="password"
              :placeholder="t('settings.passwordPh')"
              autocomplete="current-password"
              show-password
            />
          </el-form-item>
          <el-form-item :label="t('settings.twoFactor')">
            <VtInput
              v-model="loginForm.twoFactorCode"
              :placeholder="t('settings.twoFactorPh')"
              autocomplete="one-time-code"
            />
          </el-form-item>
          <VtAlert
            v-if="loginError"
            variant="danger"
            :title="loginError"
            class="login-error"
          />
          <VtButton
            variant="primary"
            data-testid="onboarding-login-submit"
            :loading="loginLoading"
            :disabled="
              loginLoading || !loginForm.username || !loginForm.password
            "
            @click="submitLogin"
          >
            {{ loginLoading ? t("settings.loggingIn") : t("settings.login") }}
          </VtButton>
        </el-form>
      </section>

      <section
        v-else-if="step === 1"
        class="onboarding-step"
        data-testid="onboarding-step-folders"
      >
        <p class="onboarding-lead text-body-sm">
          {{ t("onboarding.foldersLead") }}
        </p>
        <div v-if="foldersLoading" class="section-loading text-body-sm">
          {{ t("common.loading") }}
        </div>
        <ul v-else class="folder-list">
          <li
            v-for="row in folderRows"
            :key="row.key"
            class="folder-row"
            :data-testid="`onboarding-folder-row-${row.key}`"
          >
            <div class="folder-row-text">
              <span class="folder-label text-body-sm">{{ row.label }}</span>
              <span class="folder-path text-caption" :title="row.value">{{
                row.value || t("onboarding.pathUnset")
              }}</span>
            </div>
            <VtButton
              variant="secondary"
              :data-testid="`onboarding-folder-change-${row.key}`"
              @click="row.change"
            >
              {{ t("onboarding.changePath") }}
            </VtButton>
          </li>
        </ul>
        <VtAlert
          v-if="folderSaveError"
          variant="danger"
          :title="folderSaveError"
        />
      </section>

      <section
        v-else
        class="onboarding-step"
        data-testid="onboarding-step-complete"
      >
        <p class="onboarding-lead text-body-sm">
          {{ t("onboarding.completeLead") }}
        </p>
        <ul class="summary-list text-body-sm">
          <li v-for="item in summaryItems" :key="item.key">
            <span class="summary-label">{{ item.label }}</span>
            <span class="summary-value">{{ item.value }}</span>
          </li>
        </ul>
        <p class="onboarding-later-hint text-caption">
          {{ t("onboarding.settingsLaterHint") }}
        </p>
      </section>
    </main>

    <footer class="onboarding-footer">
      <VtButton
        variant="tertiary"
        data-testid="onboarding-skip-step"
        @click="skipStep"
      >
        {{ t("onboarding.skipStep") }}
      </VtButton>
      <VtButton
        variant="primary"
        data-testid="onboarding-next"
        :loading="nextLoading"
        @click="goNext"
      >
        {{ nextLabel }}
      </VtButton>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import VtAlert from "./VtAlert.vue";
import VtButton from "./VtButton.vue";
import VtInput from "./VtInput.vue";
import VtTag from "./VtTag.vue";
import { App } from "@/wails/app";
import type { OnboardingFolderDefaultsDTO } from "@/wails/app";
import { useSessionUnlock } from "@/composables/useSessionUnlock";

const emit = defineEmits<{
  finished: [];
}>();

const { t } = useI18n();
const {
  state: unlockState,
  errorMessage: unlockErrorMessage,
  persistAfterLogin,
} = useSessionUnlock();

const step = ref(0);
const nextLoading = ref(false);

const loggedIn = ref(false);
const displayName = ref("");
const loginSkipped = ref(false);
const loginCompleted = ref(false);

const loginForm = reactive({
  username: "",
  password: "",
  twoFactorCode: "",
});
const loginError = ref("");
const loginLoading = ref(false);

const foldersLoading = ref(true);
const folderSaveError = ref("");
const foldersSaved = ref(false);
const folderDefs = reactive<OnboardingFolderDefaultsDTO>({
  vrchatPathWindows: "",
  steamPathLinux: "",
  outputLogPath: "",
  pictureOutputFolder: "",
});

const stepTitle = computed(() => {
  if (step.value === 0) return t("onboarding.stepLogin");
  if (step.value === 1) return t("onboarding.stepFolders");
  return t("onboarding.stepComplete");
});

const nextLabel = computed(() =>
  step.value === 2 ? t("onboarding.finish") : t("onboarding.next"),
);

type FolderKey = keyof OnboardingFolderDefaultsDTO;

const folderRows = computed(() => {
  const rows: {
    key: FolderKey;
    label: string;
    value: string;
    change: () => void;
  }[] = [
    {
      key: "outputLogPath",
      label: t("settings.pathOutputLog"),
      value: folderDefs.outputLogPath,
      change: () => void changeOutputLogPath(),
    },
    {
      key: "pictureOutputFolder",
      label: t("onboarding.pictureFolder"),
      value: folderDefs.pictureOutputFolder,
      change: () => void changePictureFolder(),
    },
    {
      key: "vrchatPathWindows",
      label: t("settings.pathVrchatWin"),
      value: folderDefs.vrchatPathWindows,
      change: () => void changeVrchatPath(),
    },
    {
      key: "steamPathLinux",
      label: t("settings.pathSteamLinux"),
      value: folderDefs.steamPathLinux,
      change: () => void changeSteamPath(),
    },
  ];
  return rows;
});

const summaryItems = computed(() => {
  const items: { key: string; label: string; value: string }[] = [];
  if (loginCompleted.value) {
    items.push({
      key: "login",
      label: t("onboarding.summaryLogin"),
      value: displayName.value || t("settings.loggedInTag"),
    });
  } else if (!loginSkipped.value && loggedIn.value) {
    items.push({
      key: "login",
      label: t("onboarding.summaryLogin"),
      value: displayName.value || t("settings.loggedInTag"),
    });
  }
  if (foldersSaved.value) {
    if (folderDefs.outputLogPath) {
      items.push({
        key: "outputLog",
        label: t("settings.pathOutputLog"),
        value: folderDefs.outputLogPath,
      });
    }
    if (folderDefs.pictureOutputFolder) {
      items.push({
        key: "picture",
        label: t("onboarding.pictureFolder"),
        value: folderDefs.pictureOutputFolder,
      });
    }
    if (folderDefs.vrchatPathWindows) {
      items.push({
        key: "vrchat",
        label: t("settings.pathVrchatWin"),
        value: folderDefs.vrchatPathWindows,
      });
    }
    if (folderDefs.steamPathLinux) {
      items.push({
        key: "steam",
        label: t("settings.pathSteamLinux"),
        value: folderDefs.steamPathLinux,
      });
    }
  }
  if (items.length === 0) {
    items.push({
      key: "none",
      label: t("onboarding.summaryNone"),
      value: t("onboarding.summarySkippedAll"),
    });
  }
  return items;
});

function dirOfPath(p: string): string {
  if (!p) return "";
  const sep = p.includes("\\") ? "\\" : "/";
  const idx = p.lastIndexOf(sep);
  return idx >= 0 ? p.slice(0, idx) : "";
}

function formatBackendError(e: unknown, fallback: string): string {
  if (e instanceof Error && e.message) return e.message;
  if (typeof e === "string" && e) return e;
  return fallback;
}

async function refreshLoginState(): Promise<void> {
  try {
    loggedIn.value = await App.isLoggedIn();
  } catch {
    loggedIn.value = false;
  }
  if (loggedIn.value) {
    try {
      const profile = await App.getSelfProfile(false);
      displayName.value = profile.displayName?.trim() ?? "";
    } catch {
      displayName.value = "";
    }
  }
}

async function loadFolderDefaults(): Promise<void> {
  foldersLoading.value = true;
  folderSaveError.value = "";
  try {
    const defs = await App.getOnboardingFolderDefaults();
    folderDefs.vrchatPathWindows = defs.vrchatPathWindows ?? "";
    folderDefs.steamPathLinux = defs.steamPathLinux ?? "";
    folderDefs.outputLogPath = defs.outputLogPath ?? "";
    folderDefs.pictureOutputFolder = defs.pictureOutputFolder ?? "";
  } catch (e) {
    folderSaveError.value = formatBackendError(e, t("settings.errOperation"));
  } finally {
    foldersLoading.value = false;
  }
}

onMounted(() => {
  void refreshLoginState();
  void loadFolderDefaults();
});

async function submitLogin(): Promise<void> {
  loginError.value = "";
  loginLoading.value = true;
  try {
    const result = await App.login(
      loginForm.username,
      loginForm.password,
      loginForm.twoFactorCode || undefined,
    );
    if (result.ok) {
      loginCompleted.value = true;
      loginForm.username = "";
      loginForm.password = "";
      loginForm.twoFactorCode = "";
      if (result.plaintextToken) {
        await persistAfterLogin(result.plaintextToken);
      }
      await refreshLoginState();
    } else {
      loginError.value = result.error || t("settings.errLogin");
    }
  } finally {
    loginLoading.value = false;
  }
}

async function saveFolders(): Promise<boolean> {
  folderSaveError.value = "";
  nextLoading.value = true;
  try {
    await App.saveOnboardingFolderSettings({ ...folderDefs });
    foldersSaved.value = true;
    return true;
  } catch (e) {
    folderSaveError.value = formatBackendError(e, t("settings.errOperation"));
    return false;
  } finally {
    nextLoading.value = false;
  }
}

async function finishWizard(): Promise<void> {
  try {
    await App.setFirstLaunchGuideCompleted(true);
  } catch {
    // Best-effort; still close the wizard.
  }
  emit("finished");
}

async function goNext(): Promise<void> {
  if (step.value === 0) {
    if (loggedIn.value) {
      loginCompleted.value = true;
    }
    step.value = 1;
    return;
  }
  if (step.value === 1) {
    const ok = await saveFolders();
    if (!ok) return;
    step.value = 2;
    return;
  }
  await finishWizard();
}

function skipStep(): void {
  if (step.value === 0) {
    loginSkipped.value = true;
    step.value = 1;
    return;
  }
  if (step.value === 1) {
    step.value = 2;
    return;
  }
  void finishWizard();
}

function skipAll(): void {
  void finishWizard();
}

async function changeOutputLogPath(): Promise<void> {
  const dir = await App.openDirectoryDialog(
    t("settings.titlePickLogDirShort"),
    dirOfPath(folderDefs.outputLogPath),
  );
  if (dir) folderDefs.outputLogPath = dir;
}

async function changePictureFolder(): Promise<void> {
  const dir = await App.openDirectoryDialog(
    t("onboarding.titlePickPictureFolder"),
    dirOfPath(folderDefs.pictureOutputFolder),
  );
  if (dir) folderDefs.pictureOutputFolder = dir;
}

async function changeVrchatPath(): Promise<void> {
  const path = await App.openFileDialog(
    t("settings.titlePickVrchatExe"),
    dirOfPath(folderDefs.vrchatPathWindows),
    "*.exe",
  );
  if (path) folderDefs.vrchatPathWindows = path;
}

async function changeSteamPath(): Promise<void> {
  const path = await App.openFileDialog(
    t("settings.titlePickSteam"),
    dirOfPath(folderDefs.steamPathLinux),
    "",
  );
  if (path) folderDefs.steamPathLinux = path;
}
</script>

<style scoped>
.onboarding-wizard {
  position: fixed;
  inset: 0;
  z-index: 2000;
  display: flex;
  flex-direction: column;
  background: var(--color-bg-base);
  color: var(--color-text-primary);
}

.onboarding-header {
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  padding: var(--space-section) var(--space-page) var(--space-block);
}

.progress-dots {
  display: flex;
  gap: var(--space-inline-tight);
  align-items: center;
}

.progress-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--color-border);
  transition: background 0.15s ease;
}

.progress-dot--active {
  background: var(--color-brand);
}

.skip-all-link {
  position: absolute;
  right: var(--space-page);
  top: 50%;
  transform: translateY(-50%);
  border: none;
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  padding: var(--space-inline-tight);
}

.skip-all-link:hover {
  color: var(--color-text-primary);
}

.onboarding-main {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 0 var(--space-page) var(--space-section);
  max-width: 560px;
  margin: 0 auto;
  width: 100%;
}

.onboarding-title {
  margin: 0 0 var(--space-block);
  text-align: center;
}

.onboarding-lead {
  color: var(--color-text-secondary);
  margin: 0 0 var(--space-form-field);
}

.onboarding-keyring-hint {
  color: var(--color-text-muted);
  margin: 0 0 var(--space-section);
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-form-field);
}

.login-done {
  display: flex;
  flex-direction: column;
  gap: var(--space-action-group);
  align-items: flex-start;
}

.login-display-name {
  margin: 0;
}

.folder-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-form-field);
}

.folder-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-block);
  padding: var(--space-form-field);
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md, 8px);
}

.folder-row-text {
  min-width: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--space-inline-tight);
}

.folder-path {
  color: var(--color-text-secondary);
  word-break: break-all;
}

.summary-list {
  list-style: none;
  margin: 0 0 var(--space-section);
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-form-field);
}

.summary-list li {
  display: flex;
  flex-direction: column;
  gap: var(--space-inline-tight);
  padding: var(--space-form-field);
  background: var(--color-bg-elevated);
  border-radius: var(--radius-md, 8px);
}

.summary-label {
  font-weight: var(--font-weight-600);
}

.summary-value {
  color: var(--color-text-secondary);
  word-break: break-all;
}

.onboarding-later-hint {
  color: var(--color-text-muted);
  margin: 0;
}

.onboarding-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-section) var(--space-page);
  border-top: 1px solid var(--color-border);
  background: var(--color-bg-elevated);
}

.section-loading {
  color: var(--color-text-secondary);
}
</style>
