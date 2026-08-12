<template>
  <div class="settings-view">
    <h1 class="page-title">{{ t("settings.title") }}</h1>

    <el-card class="settings-card" shadow="never">
      <template #header>
        <span>{{ t("settings.language") }}</span>
      </template>
      <el-text type="info" size="small" class="hint block-hint">{{
        t("settings.languageHint")
      }}</el-text>
      <VtSelect
        class="language-select"
        :model-value="locale"
        data-testid="settings-ui-language"
        @update:model-value="onLanguageChange"
      >
        <el-option value="ja" :label="t('settingsLanguages.ja')" />
        <el-option value="en" :label="t('settingsLanguages.en')" />
        <el-option value="ko" :label="t('settingsLanguages.ko')" />
        <el-option value="zh-TW" :label="t('settingsLanguages.zhTW')" />
        <el-option value="zh-CN" :label="t('settingsLanguages.zhCN')" />
      </VtSelect>
    </el-card>

    <!-- VRChat ログイン -->
    <el-card class="settings-card" shadow="never">
      <template #header>
        <span>{{ t("settings.loginSection") }}</span>
      </template>
      <div v-if="isLoggedIn" class="login-status">
        <div v-if="profileLoading && !selfProfile" class="profile-loading">
          {{ t("settings.profileLoading") }}
        </div>
        <div v-else-if="selfProfile" class="current-user-card">
          <img
            v-if="avatarDisplayUrl"
            :src="avatarDisplayUrl"
            alt=""
            class="current-user-avatar"
            width="96"
            height="96"
          />
          <div class="current-user-details">
            <p class="current-user-display-name">
              {{ selfProfile.displayName || t("settings.noDisplayName") }}
            </p>
            <p v-if="selfProfile.username" class="current-user-line">
              @{{ selfProfile.username }}
            </p>
            <p v-if="selfProfile.vrcUserId" class="current-user-line muted">
              {{ selfProfile.vrcUserId }}
            </p>
            <p class="current-user-line">
              {{ t("settings.statusLine") }}
              {{ selfProfile.status || t("common.dash") }} /
              {{ selfProfile.state || t("common.dash") }}
            </p>
            <p
              v-if="selfProfile.statusDescription"
              class="current-user-line muted"
            >
              {{ selfProfile.statusDescription }}
            </p>
            <router-link
              :to="{ name: 'me' }"
              class="self-profile-link"
              data-testid="settings-view-self-profile"
            >
              {{ t("settings.viewSelfProfile") }}
            </router-link>
          </div>
        </div>
        <VtAlert v-if="profileError" variant="danger" :title="profileError" />
        <VtTag variant="success" size="large">{{
          t("settings.loggedInTag")
        }}</VtTag>
        <div class="login-actions">
          <VtButton
            variant="tertiary"
            :loading="profileLoading"
            @click="loadSelfProfileSummary(true)"
          >
            {{ t("settings.refreshProfile") }}
          </VtButton>
          <VtButton variant="secondary" @click="refreshFriends">
            {{ t("settings.refreshFriends") }}
          </VtButton>
          <VtButton variant="danger" plain @click="logout">
            {{ t("settings.logout") }}
          </VtButton>
        </div>
      </div>
      <div v-else class="login-form">
        <VtAlert
          v-if="unlockState === 'needs-relogin' && unlockErrorMessage"
          variant="warning"
          :title="unlockErrorMessage"
          class="login-error"
        />
        <el-form label-position="top" size="default">
          <el-form-item :label="t('settings.username')">
            <VtInput
              id="login-username"
              v-model="loginForm.username"
              :placeholder="t('settings.usernamePh')"
              autocomplete="username"
            />
          </el-form-item>
          <el-form-item :label="t('settings.password')">
            <VtInput
              id="login-password"
              v-model="loginForm.password"
              type="password"
              :placeholder="t('settings.passwordPh')"
              autocomplete="current-password"
              show-password
            />
          </el-form-item>
          <el-form-item :label="t('settings.twoFactor')">
            <VtInput
              id="login-2fa"
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
            :loading="loginLoading"
            :disabled="
              loginLoading || !loginForm.username || !loginForm.password
            "
            @click="login"
          >
            {{ loginLoading ? t("settings.loggingIn") : t("settings.login") }}
          </VtButton>
        </el-form>
      </div>
    </el-card>

    <!-- パス設定 -->
    <el-card class="settings-card" shadow="never">
      <template #header>
        <span>{{ t("settings.pathSection") }}</span>
      </template>
      <div class="path-settings">
        <div v-for="field in pathFields" :key="field.key" class="path-row">
          <label class="path-label">{{ field.label }}</label>
          <div class="path-input-group">
            <VtInput
              v-model="pathSettings[field.key]"
              :placeholder="field.placeholder"
              @change="savePathSettings"
            />
            <VtButton
              v-for="btn in field.buttons"
              :key="btn.label"
              variant="secondary"
              :data-testid="btn.testid"
              :title="btn.title"
              @click="btn.handler"
            >
              {{ btn.label }}
            </VtButton>
            <VtButton
              variant="secondary"
              :data-testid="`path-validate-${field.key}`"
              :disabled="!pathSettings[field.key]"
              @click="validatePathField(field.key)"
            >
              {{ t("settings.validateExists") }}
            </VtButton>
          </div>
          <el-text
            v-if="validateResult[field.key] !== null"
            :type="validateResult[field.key] ? 'success' : 'danger'"
            size="small"
          >
            {{
              validateResult[field.key]
                ? t("settings.existsYes")
                : t("settings.existsNo")
            }}
          </el-text>
        </div>
      </div>
      <el-text type="info" size="small" class="hint">{{
        t("settings.pathHint")
      }}</el-text>
    </el-card>

    <!-- 電源（Windows） -->
    <el-card class="settings-card" shadow="never">
      <template #header>
        <span>{{ t("settings.powerSection") }}</span>
      </template>
      <div class="setting-row power-setting-row">
        <div class="power-toggle-label">
          <span>{{ t("settings.suppressSleep") }}</span>
          <el-text type="info" size="small" class="hint block-hint">{{
            t("settings.suppressSleepHint")
          }}</el-text>
        </div>
        <VtSwitch
          v-model="suppressSleepWhileVRChat"
          class="power-switch"
          @change="saveSuppressSleepWhileVRChat"
        />
      </div>
    </el-card>

    <!-- ログ・データ管理 -->
    <el-card class="settings-card" shadow="never">
      <template #header>
        <span>{{ t("settings.dataSection") }}</span>
      </template>
      <div class="setting-row">
        <label>{{ t("settings.retentionLabel") }}</label>
        <el-input-number
          v-model="logRetentionDays"
          :min="1"
          :max="365"
          @change="saveRetention"
        />
      </div>
      <el-text type="info" size="small" class="hint">{{
        t("settings.retentionHint")
      }}</el-text>
    </el-card>

    <!-- OSS ライセンス -->
    <el-card class="settings-card" shadow="never">
      <template #header>
        <span>{{ t("settings.ossSection") }}</span>
      </template>
      <el-text type="info" size="small" class="hint">{{
        t("settings.ossHint")
      }}</el-text>
      <div class="oss-link-row">
        <router-link class="btn-licenses" to="/licenses">
          <VtButton variant="primary">{{ t("settings.ossButton") }}</VtButton>
        </router-link>
      </div>
    </el-card>

    <!-- DB メンテナンス -->
    <el-card class="settings-card" shadow="never">
      <template #header>
        <span>{{ t("settings.dbSection") }}</span>
      </template>
      <el-text type="info" size="small" class="hint db-hint">
        {{ t("settings.dbHint") }}
      </el-text>
      <VtAlert
        v-if="maintenanceError"
        variant="danger"
        :title="maintenanceError"
        class="maintenance-error-alert"
      />
      <div class="maintenance-actions">
        <VtButton
          variant="secondary"
          :loading="maintenanceLoading"
          @click="doVacuumDb"
        >
          {{
            maintenanceLoading ? t("settings.running") : t("settings.vacuum")
          }}
        </VtButton>
        <VtButton
          variant="danger"
          plain
          :loading="maintenanceLoading"
          @click="doClearEncounters"
        >
          {{ t("settings.clearEncounters") }}
        </VtButton>
        <VtButton
          variant="danger"
          plain
          :loading="maintenanceLoading"
          @click="doClearScreenshots"
        >
          {{ t("settings.clearScreenshots") }}
        </VtButton>
        <VtButton
          variant="danger"
          plain
          :loading="maintenanceLoading"
          @click="doClearFriendsCache"
        >
          {{ t("settings.clearFriends") }}
        </VtButton>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import { ElMessageBox } from "element-plus";
import VtAlert from "../components/VtAlert.vue";
import VtButton from "../components/VtButton.vue";
import VtInput from "../components/VtInput.vue";
import VtSelect from "../components/VtSelect.vue";
import VtSwitch from "../components/VtSwitch.vue";
import VtTag from "../components/VtTag.vue";
import {
  VT_BUTTON_DANGER_CONFIRM_CLASS,
  VT_BUTTON_SECONDARY_CANCEL_CLASS,
} from "../components/vtButtonClasses";
import { App } from "../wails/app";
import type { PathSettingsDTO, UserCacheDTO } from "../wails/app";
import { friendThumbUrl } from "../utils/vrcUserCacheDisplay";
import { useSessionUnlock } from "../composables/useSessionUnlock";
import { isAppLocale, setLanguage } from "../i18n";
import { showToast } from "../utils/showToast";

const {
  state: unlockState,
  errorMessage: unlockErrorMessage,
  beginStartupUnlock,
  persistAfterLogin,
  handleLogout,
} = useSessionUnlock();

const { t, locale } = useI18n();

const isLoggedIn = ref(false);
const selfProfile = ref<UserCacheDTO | null>(null);
const profileLoading = ref(false);
const profileError = ref("");

const avatarDisplayUrl = computed(() => {
  const u = selfProfile.value;
  if (!u) return "";
  return friendThumbUrl(u) ?? "";
});

function formatBackendError(e: unknown, fallback: string): string {
  if (e instanceof Error && e.message) return e.message;
  if (typeof e === "string" && e) return e;
  if (e && typeof e === "object" && "message" in e) {
    const m = (e as { message: unknown }).message;
    if (typeof m === "string" && m) return m;
  }
  return fallback;
}

const loginForm = reactive({
  username: "",
  password: "",
  twoFactorCode: "",
});
const loginError = ref("");
const loginLoading = ref(false);

const logRetentionDays = ref(30);
const suppressSleepWhileVRChat = ref(false);
const maintenanceError = ref("");
const maintenanceLoading = ref(false);
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
    key: "vrchatPathWindows" as keyof PathSettingsDTO,
    label: t("settings.pathVrchatWin"),
    placeholder: t("settings.pathVrchatWinPh"),
    buttons: [
      {
        label: t("common.browse"),
        testid: "vrchat-path-browse",
        title: t("settings.titlePickVrchatExe"),
        handler: browseVrchatPath,
      },
    ],
  },
  {
    key: "steamPathLinux" as keyof PathSettingsDTO,
    label: t("settings.pathSteamLinux"),
    placeholder: t("settings.pathSteamLinuxPh"),
    buttons: [
      {
        label: t("common.browse"),
        testid: "steam-path-browse",
        title: t("settings.titlePickSteam"),
        handler: browseSteamPath,
      },
    ],
  },
  {
    key: "outputLogPath" as keyof PathSettingsDTO,
    label: t("settings.pathOutputLog"),
    placeholder: t("settings.pathOutputLogPh"),
    buttons: [
      {
        label: t("settings.browseFolder"),
        testid: "output-log-dir-browse",
        title: t("settings.titlePickLogDir"),
        handler: browseOutputLogDirectory,
      },
      {
        label: t("settings.openLogFolder"),
        testid: "",
        title: t("settings.titleOpenLogFolder"),
        handler: openVRChatLogFolder,
      },
    ],
  },
]);

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

onMounted(async () => {
  await beginStartupUnlock().catch(() => undefined);
  try {
    isLoggedIn.value = await App.isLoggedIn();
  } catch {
    isLoggedIn.value = false;
  }
  if (isLoggedIn.value) {
    await loadSelfProfileSummary();
  }
  logRetentionDays.value = await App.getLogRetentionDays();
  suppressSleepWhileVRChat.value = await App.getSuppressSleepWhileVRChat();
  const ps = await App.getPathSettings();
  pathSettings.vrchatPathWindows = ps.vrchatPathWindows;
  pathSettings.steamPathLinux = ps.steamPathLinux;
  pathSettings.outputLogPath = ps.outputLogPath;
});

async function loadSelfProfileSummary(forceRefresh = false) {
  profileError.value = "";
  profileLoading.value = true;
  try {
    selfProfile.value = await App.getSelfProfile(forceRefresh);
  } catch (e) {
    selfProfile.value = null;
    profileError.value = formatBackendError(e, t("settings.errProfile"));
  } finally {
    profileLoading.value = false;
  }
}

async function login() {
  loginError.value = "";
  loginLoading.value = true;
  try {
    const result = await App.login(
      loginForm.username,
      loginForm.password,
      loginForm.twoFactorCode || undefined,
    );
    if (result.ok) {
      isLoggedIn.value = true;
      loginForm.username = "";
      loginForm.password = "";
      loginForm.twoFactorCode = "";
      // Wrap the one-time token with Web Crypto and persist the encrypted blob.
      // This must be done immediately before the token reference is dropped.
      if (result.plaintextToken) {
        await persistAfterLogin(result.plaintextToken);
      }
      await loadSelfProfileSummary();
    } else {
      loginError.value = result.error || t("settings.errLogin");
    }
  } finally {
    loginLoading.value = false;
  }
}

async function logout() {
  loginError.value = "";
  profileError.value = "";
  selfProfile.value = null;
  try {
    await App.logout();
  } catch (e) {
    loginError.value = e instanceof Error ? e.message : t("settings.errLogout");
  }
  // Always clean up frontend-side state (IDB wrapping key + blob)
  // even if the backend logout partially failed.
  await handleLogout();
  isLoggedIn.value = false;
}

async function refreshFriends() {
  loginError.value = "";
  try {
    await App.refreshFriends();
  } catch (e) {
    loginError.value =
      e instanceof Error ? e.message : t("settings.errFriends");
  }
}

async function saveRetention() {
  await App.setLogRetentionDays(logRetentionDays.value);
}

async function saveSuppressSleepWhileVRChat() {
  await App.setSuppressSleepWhileVRChat(suppressSleepWhileVRChat.value);
}

async function savePathSettings() {
  try {
    await App.setPathSettings(pathSettings);
  } catch (e) {
    showToast.error(formatBackendError(e, t("settings.errOperation")));
    return;
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

function doClearFriendsCache() {
  void runWithConfirm(
    t("settings.clearFriendsConfirm"),
    async () => App.clearFriendsCache(),
    (n) => {
      selfProfile.value = null;
      profileError.value = "";
      return t("settings.clearFriendsDone", { n: String(n ?? 0) });
    },
  );
}
</script>

<style scoped>
.settings-card {
  margin-bottom: var(--space-section);
  background: var(--color-bg-elevated) !important;
  border-color: var(--color-border) !important;
}

.settings-card :deep(.el-card__header) {
  font-weight: var(--font-weight-600);
  border-bottom-color: var(--color-border);
}

.login-status {
  display: flex;
  flex-direction: column;
  gap: var(--space-form-field);
}

.profile-loading {
  font-size: var(--font-size-12);
  color: var(--color-text-secondary);
}

.current-user-card {
  display: flex;
  gap: var(--space-block);
  align-items: flex-start;
  padding: var(--space-form-field);
  background: var(--color-bg-muted);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  max-width: 480px;
}

.current-user-avatar {
  flex-shrink: 0;
  border-radius: var(--radius);
  object-fit: cover;
  background: var(--color-bg-base);
}

.current-user-details {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-inline-tight);
}

.current-user-display-name {
  margin: 0;
  font-size: var(--font-size-14);
  font-weight: var(--font-weight-600);
}

.current-user-line {
  margin: 0;
  font-size: var(--font-size-12);
  word-break: break-all;
}

.current-user-line.muted {
  color: var(--color-text-secondary);
}

.self-profile-link {
  display: inline-block;
  margin-top: var(--space-action-group);
  color: var(--el-color-primary);
  text-decoration: none;
  font-size: var(--font-size-12);
}

.self-profile-link:hover {
  text-decoration: underline;
}

.login-actions {
  display: flex;
  gap: var(--space-action-group);
  flex-wrap: wrap;
}

.login-form {
  max-width: 360px;
}

.login-error {
  margin-bottom: var(--space-form-field);
}

.hint {
  display: block;
  margin-top: var(--space-form-field);
}

.db-hint {
  display: block;
  margin-bottom: var(--space-block);
}

.maintenance-error-alert {
  margin-bottom: var(--space-form-field);
}

.oss-link-row {
  margin-top: var(--space-form-field);
}

.path-settings {
  display: flex;
  flex-direction: column;
  gap: var(--space-block);
}

.path-row {
  display: flex;
  flex-direction: column;
  gap: var(--space-inline-tight);
}

.path-label {
  font-size: var(--font-size-14);
  color: var(--color-text-primary);
}

.path-input-group {
  display: flex;
  gap: var(--space-action-group);
  align-items: center;
  flex-wrap: wrap;
}

.path-input-group :deep(.el-input) {
  flex: 1;
  min-width: 0;
}

.setting-row {
  display: flex;
  align-items: center;
  gap: var(--space-block);
  margin-bottom: var(--space-action-group);
}

.power-setting-row {
  align-items: flex-start;
}

.power-toggle-label {
  flex: 1;
  min-width: 0;
}

.block-hint {
  display: block;
  margin-top: var(--space-inline-tight);
}

.power-switch {
  flex-shrink: 0;
  margin-top: var(--space-inline-tight);
}

.maintenance-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-action-group);
}

.language-select {
  display: block;
  margin-top: var(--space-action-group);
  max-width: 22rem;
  width: 100%;
}
</style>
