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
      <el-select
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
      </el-select>
    </el-card>

    <!-- VRChat ログイン -->
    <el-card class="settings-card" shadow="never">
      <template #header>
        <span>{{ t("settings.loginSection") }}</span>
      </template>
      <div v-if="isLoggedIn" class="login-status">
        <div v-if="profileLoading && !currentUser" class="profile-loading">
          {{ t("settings.profileLoading") }}
        </div>
        <div v-else-if="currentUser" class="current-user-card">
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
              {{ currentUser.displayName || t("settings.noDisplayName") }}
            </p>
            <p v-if="currentUser.username" class="current-user-line">
              @{{ currentUser.username }}
            </p>
            <p v-if="currentUser.id" class="current-user-line muted">
              {{ currentUser.id }}
            </p>
            <p class="current-user-line">
              {{ t("settings.statusLine") }}
              {{ currentUser.status || t("common.dash") }} /
              {{ currentUser.state || t("common.dash") }}
            </p>
            <p
              v-if="currentUser.statusDescription"
              class="current-user-line muted"
            >
              {{ currentUser.statusDescription }}
            </p>
          </div>
        </div>
        <el-alert
          v-if="profileError"
          :title="profileError"
          type="error"
          :closable="false"
          show-icon
        />
        <el-tag type="success" size="large">{{
          t("settings.loggedInTag")
        }}</el-tag>
        <div class="login-actions">
          <el-button
            type="primary"
            :loading="profileLoading"
            @click="loadCurrentUser(true)"
          >
            {{ t("settings.refreshProfile") }}
          </el-button>
          <el-button type="primary" @click="refreshFriends">
            {{ t("settings.refreshFriends") }}
          </el-button>
          <el-button type="danger" plain @click="logout">
            {{ t("settings.logout") }}
          </el-button>
        </div>
      </div>
      <div v-else class="login-form">
        <el-alert
          v-if="unlockState === 'needs-relogin' && unlockErrorMessage"
          :title="unlockErrorMessage"
          type="warning"
          :closable="false"
          show-icon
          class="login-error"
        />
        <el-form label-position="top" size="default">
          <el-form-item :label="t('settings.username')">
            <el-input
              id="login-username"
              v-model="loginForm.username"
              :placeholder="t('settings.usernamePh')"
              autocomplete="username"
            />
          </el-form-item>
          <el-form-item :label="t('settings.password')">
            <el-input
              id="login-password"
              v-model="loginForm.password"
              type="password"
              :placeholder="t('settings.passwordPh')"
              autocomplete="current-password"
              show-password
            />
          </el-form-item>
          <el-form-item :label="t('settings.twoFactor')">
            <el-input
              id="login-2fa"
              v-model="loginForm.twoFactorCode"
              :placeholder="t('settings.twoFactorPh')"
              autocomplete="one-time-code"
            />
          </el-form-item>
          <el-alert
            v-if="loginError"
            :title="loginError"
            type="error"
            :closable="false"
            show-icon
            class="login-error"
          />
          <el-button
            type="primary"
            :loading="loginLoading"
            :disabled="
              loginLoading || !loginForm.username || !loginForm.password
            "
            @click="login"
          >
            {{ loginLoading ? t("settings.loggingIn") : t("settings.login") }}
          </el-button>
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
            <el-input
              v-model="pathSettings[field.key]"
              :placeholder="field.placeholder"
              @change="savePathSettings"
            />
            <el-button
              v-for="btn in field.buttons"
              :key="btn.label"
              :data-testid="btn.testid"
              :title="btn.title"
              @click="btn.handler"
            >
              {{ btn.label }}
            </el-button>
            <el-button
              type="primary"
              :disabled="!pathSettings[field.key]"
              @click="validatePathField(field.key)"
            >
              {{ t("settings.validateExists") }}
            </el-button>
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
        <el-switch
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
      <div style="margin-top: 0.75rem">
        <router-link class="btn-licenses" to="/licenses">
          <el-button type="primary">{{ t("settings.ossButton") }}</el-button>
        </router-link>
      </div>
    </el-card>

    <!-- DB メンテナンス -->
    <el-card class="settings-card" shadow="never">
      <template #header>
        <span>{{ t("settings.dbSection") }}</span>
      </template>
      <el-text
        type="info"
        size="small"
        class="hint"
        style="display: block; margin-bottom: 1rem"
      >
        {{ t("settings.dbHint") }}
      </el-text>
      <el-alert
        v-if="maintenanceError"
        :title="maintenanceError"
        type="error"
        :closable="false"
        show-icon
        style="margin-bottom: 0.75rem"
      />
      <div class="maintenance-actions">
        <el-button :loading="maintenanceLoading" @click="doVacuumDb">
          {{
            maintenanceLoading ? t("settings.running") : t("settings.vacuum")
          }}
        </el-button>
        <el-button
          type="danger"
          plain
          :loading="maintenanceLoading"
          @click="doClearEncounters"
        >
          {{ t("settings.clearEncounters") }}
        </el-button>
        <el-button
          type="danger"
          plain
          :loading="maintenanceLoading"
          @click="doClearScreenshots"
        >
          {{ t("settings.clearScreenshots") }}
        </el-button>
        <el-button
          type="danger"
          plain
          :loading="maintenanceLoading"
          @click="doClearFriendsCache"
        >
          {{ t("settings.clearFriends") }}
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import { ElMessageBox, ElMessage } from "element-plus";
import { App } from "../wails/app";
import type { PathSettingsDTO, VRChatCurrentUserDTO } from "../wails/app";
import { useSessionUnlock } from "../composables/useSessionUnlock";
import { isAppLocale, setLanguage } from "../i18n";

const {
  state: unlockState,
  errorMessage: unlockErrorMessage,
  beginStartupUnlock,
  persistAfterLogin,
  handleLogout,
} = useSessionUnlock();

const { t, locale } = useI18n();

const isLoggedIn = ref(false);
const currentUser = ref<VRChatCurrentUserDTO | null>(null);
const profileLoading = ref(false);
const profileError = ref("");

const avatarDisplayUrl = computed(() => {
  const u = currentUser.value;
  if (!u) return "";
  return (
    u.profilePicOverrideThumbnail ||
    u.currentAvatarThumbnailImageUrl ||
    u.userIcon ||
    ""
  );
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
        label: t("settings.browseFile"),
        testid: "output-log-path-browse",
        title: t("settings.titlePickOutputLog"),
        handler: browseOutputLogPath,
      },
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
    ElMessage.error(formatBackendError(e, t("settings.errLanguageSave")));
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
    await loadCurrentUser();
  }
  logRetentionDays.value = await App.getLogRetentionDays();
  suppressSleepWhileVRChat.value = await App.getSuppressSleepWhileVRChat();
  const ps = await App.getPathSettings();
  pathSettings.vrchatPathWindows = ps.vrchatPathWindows;
  pathSettings.steamPathLinux = ps.steamPathLinux;
  pathSettings.outputLogPath = ps.outputLogPath;
});

async function loadCurrentUser(forceRefresh = false) {
  profileError.value = "";
  profileLoading.value = true;
  try {
    currentUser.value = await App.getVRChatCurrentUser(forceRefresh);
  } catch (e) {
    currentUser.value = null;
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
      await loadCurrentUser();
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
  currentUser.value = null;
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
  await App.setPathSettings(pathSettings);
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

async function browseOutputLogPath() {
  const path = await App.openFileDialog(
    t("settings.titlePickOutputLog"),
    dirOfPath(pathSettings.outputLogPath),
    "*.txt",
  );
  if (path) {
    pathSettings.outputLogPath = path;
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
      ElMessage.success(msg);
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
      currentUser.value = null;
      profileError.value = "";
      return t("settings.clearFriendsDone", { n: String(n ?? 0) });
    },
  );
}
</script>

<style scoped>
.settings-card {
  margin-bottom: 1.5rem;
  background: var(--bg-secondary) !important;
  border-color: var(--border) !important;
}

.settings-card :deep(.el-card__header) {
  font-weight: 600;
  border-bottom-color: var(--border);
}

.login-status {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.profile-loading {
  font-size: 0.9rem;
  color: var(--text-secondary);
}

.current-user-card {
  display: flex;
  gap: 1rem;
  align-items: flex-start;
  padding: 0.75rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  max-width: 480px;
}

.current-user-avatar {
  flex-shrink: 0;
  border-radius: var(--radius);
  object-fit: cover;
  background: var(--bg-primary);
}

.current-user-details {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.current-user-display-name {
  margin: 0;
  font-size: 1.05rem;
  font-weight: 600;
}

.current-user-line {
  margin: 0;
  font-size: 0.88rem;
  word-break: break-all;
}

.current-user-line.muted {
  color: var(--text-secondary);
}

.login-actions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.login-form {
  max-width: 360px;
}

.login-error {
  margin-bottom: 0.75rem;
}

.hint {
  display: block;
  margin-top: 0.75rem;
}

.path-settings {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.path-row {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.path-label {
  font-size: 0.95rem;
  color: var(--text-primary);
}

.path-input-group {
  display: flex;
  gap: 0.5rem;
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
  gap: 1rem;
  margin-bottom: 0.5rem;
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
  margin-top: 0.35rem;
}

.power-switch {
  flex-shrink: 0;
  margin-top: 0.15rem;
}

.maintenance-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.language-select {
  display: block;
  margin-top: 0.65rem;
  max-width: 22rem;
  width: 100%;
}
</style>
