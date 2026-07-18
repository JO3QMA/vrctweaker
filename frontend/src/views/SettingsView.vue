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
            @click="loadSelfProfileSummary(true)"
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

    <!-- yt-dlp Cookie linkage（Windows） -->
    <el-card
      v-if="cookieSupported"
      class="settings-card"
      shadow="never"
      data-testid="settings-cookie-linkage"
    >
      <template #header>
        <span>{{ t("settings.cookieLinkage.section") }}</span>
      </template>
      <el-alert
        type="warning"
        :closable="false"
        show-icon
        class="cookie-always-warn"
        :title="t('settings.cookieLinkage.alwaysWarn')"
      />
      <el-alert
        v-if="!toolsEffectiveOfficial"
        type="info"
        :closable="false"
        show-icon
        class="cookie-official-hint"
        data-testid="settings-cookie-official-hint"
      >
        <template #title>
          {{ t("settings.cookieLinkage.officialHint") }}
          <router-link :to="{ name: 'video' }" class="cookie-video-link">
            {{ t("settings.cookieLinkage.openVideoTab") }}
          </router-link>
        </template>
      </el-alert>
      <el-alert
        v-if="cookieSourceKind === 'unsupported'"
        type="warning"
        :closable="false"
        show-icon
        :title="t('settings.cookieLinkage.unsupportedForm')"
      />
      <el-alert
        v-if="cookieActionError"
        :title="cookieActionError"
        type="error"
        :closable="false"
        show-icon
        class="cookie-action-error"
      />
      <div class="setting-row power-setting-row">
        <div class="power-toggle-label">
          <span>{{ t("settings.cookieLinkage.enableLabel") }}</span>
          <el-text type="info" size="small" class="hint block-hint">{{
            t("settings.cookieLinkage.lockHint")
          }}</el-text>
        </div>
        <el-switch
          v-model="cookieEnabled"
          class="power-switch"
          data-testid="settings-cookie-enable"
          :disabled="cookieBusy"
          @change="onCookieEnableChange"
        />
      </div>
      <el-form label-position="top" size="default" class="cookie-form">
        <el-form-item :label="t('settings.cookieLinkage.sourceLabel')">
          <el-radio-group
            v-model="cookieDraftSource"
            data-testid="settings-cookie-source"
            :disabled="cookieBusy"
            @change="onCookieSourceChange"
          >
            <el-radio-button value="browser">{{
              t("settings.cookieLinkage.sourceBrowser")
            }}</el-radio-button>
            <el-radio-button value="file">{{
              t("settings.cookieLinkage.sourceFile")
            }}</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item
          v-if="cookieDraftSource === 'browser'"
          :label="t('settings.cookieLinkage.browserLabel')"
        >
          <el-select
            v-model="cookieDraftBrowser"
            data-testid="settings-cookie-browser"
            :disabled="cookieBusy"
            @change="onCookieBrowserChange"
          >
            <el-option value="chrome" label="Chrome" />
            <el-option value="edge" label="Edge" />
            <el-option value="firefox" label="Firefox" />
          </el-select>
        </el-form-item>
        <el-form-item
          v-else
          :label="t('settings.cookieLinkage.cookiesFileLabel')"
        >
          <div class="path-input-group">
            <el-input
              v-model="cookieDraftCookiesPath"
              data-testid="settings-cookie-file-path"
              :placeholder="t('settings.cookieLinkage.cookiesFilePh')"
              :disabled="cookieBusy"
              @change="onCookiePathChange"
            />
            <el-button
              data-testid="settings-cookie-file-browse"
              :disabled="cookieBusy"
              @click="browseCookieFile"
            >
              {{ t("settings.cookieLinkage.browse") }}
            </el-button>
          </div>
        </el-form-item>
      </el-form>
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
import type {
  CookieLinkageStatusDTO,
  PathSettingsDTO,
  UserCacheDTO,
} from "../wails/app";
import { friendThumbUrl } from "../utils/vrcUserCacheDisplay";
import { useSessionUnlock } from "../composables/useSessionUnlock";
import { isAppLocale, setLanguage } from "../i18n";
import { cookieLinkageErrorI18nKey } from "./cookieLinkageErrors";

const {
  state: unlockState,
  errorMessage: unlockErrorMessage,
  beginStartupUnlock,
  persistAfterLogin,
  handleLogout,
} = useSessionUnlock();

const { t, locale, te } = useI18n();

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
    ElMessage.error(formatBackendError(e, t("settings.errLanguageSave")));
    return;
  }
  setLanguage(v);
}

const cookieSupported = ref(false);
const cookieEnabled = ref(false);
const cookieBusy = ref(false);
const cookieActionError = ref("");
const cookieRiskAcknowledged = ref(false);
const cookieSourceKind = ref("");
const cookieDraftSource = ref<"browser" | "file">("browser");
const cookieDraftBrowser = ref("chrome");
const cookieDraftCookiesPath = ref("");
const toolsEffectiveOfficial = ref(true);
let cookieViewGen = 0;

function isCookieViewStale(gen: number): boolean {
  return gen !== cookieViewGen;
}

function isMessageBoxDismiss(e: unknown): boolean {
  return e === "cancel" || e === "close" || e === "backdrop";
}

function userFacingCookieError(raw: string): string {
  if (!raw) return "";
  if (te(`settings.cookieLinkage.${raw}`)) {
    return t(`settings.cookieLinkage.${raw}`);
  }
  return t(cookieLinkageErrorI18nKey(raw));
}

function applyCookieStatus(st: CookieLinkageStatusDTO) {
  cookieSupported.value = !!st.supported;
  cookieEnabled.value = !!st.enabled;
  cookieRiskAcknowledged.value = !!st.riskAcknowledged;
  cookieSourceKind.value = st.sourceKind || "";
  if (st.sourceKind === "browser" && st.browser) {
    cookieDraftSource.value = "browser";
    cookieDraftBrowser.value = st.browser;
  } else if (st.sourceKind === "file" && st.cookiesFilePath) {
    cookieDraftSource.value = "file";
    cookieDraftCookiesPath.value = st.cookiesFilePath;
  } else if (st.sourceKind === "unsupported") {
    // keep draft; user picks v1 form to replace
  }
}

async function refreshCookieStatus(gen: number) {
  const st = await App.getYTDLPCookieLinkageStatus();
  if (isCookieViewStale(gen)) return;
  applyCookieStatus(st);
  try {
    const maintain = await App.getYTDLPMaintainStatus();
    if (isCookieViewStale(gen)) return;
    toolsEffectiveOfficial.value = !!maintain.effectiveOfficial;
  } catch {
    if (!isCookieViewStale(gen)) toolsEffectiveOfficial.value = true;
  }
}

async function ensureCookieRiskAck(gen: number): Promise<boolean> {
  if (cookieRiskAcknowledged.value) return true;
  await ElMessageBox.confirm(
    t("settings.cookieLinkage.riskAckBody"),
    t("settings.cookieLinkage.riskAckTitle"),
    {
      confirmButtonText: t("settings.cookieLinkage.riskAckConfirm"),
      cancelButtonText: t("common.cancel"),
      type: "warning",
    },
  );
  if (isCookieViewStale(gen)) return false;
  await App.acknowledgeYTDLPCookieLinkageRisk();
  if (isCookieViewStale(gen)) return false;
  cookieRiskAcknowledged.value = true;
  return true;
}

async function writeCookieFromDraft(gen: number) {
  if (cookieDraftSource.value === "browser") {
    await App.setYTDLPCookieLinkageBrowser(cookieDraftBrowser.value);
  } else {
    await App.setYTDLPCookieLinkageCookiesFile(cookieDraftCookiesPath.value);
  }
  if (isCookieViewStale(gen)) return;
  await refreshCookieStatus(gen);
}

async function onCookieEnableChange(on: boolean | string | number) {
  const desired = on === true || on === "true";
  const gen = ++cookieViewGen;
  cookieBusy.value = true;
  cookieActionError.value = "";
  try {
    await ensureCookieRiskAck(gen);
    if (isCookieViewStale(gen)) return;
    if (desired) {
      await writeCookieFromDraft(gen);
    } else {
      await App.disableYTDLPCookieLinkage();
      if (isCookieViewStale(gen)) return;
      await refreshCookieStatus(gen);
    }
  } catch (e) {
    if (isCookieViewStale(gen)) return;
    cookieEnabled.value = !desired;
    if (isMessageBoxDismiss(e)) return;
    cookieActionError.value = userFacingCookieError(
      e instanceof Error ? e.message : String(e),
    );
  } finally {
    if (!isCookieViewStale(gen)) cookieBusy.value = false;
  }
}

async function onCookieSourceChange() {
  if (!cookieEnabled.value) return;
  const gen = ++cookieViewGen;
  cookieBusy.value = true;
  cookieActionError.value = "";
  try {
    await ensureCookieRiskAck(gen);
    if (isCookieViewStale(gen)) return;
    await writeCookieFromDraft(gen);
  } catch (e) {
    if (isCookieViewStale(gen)) return;
    if (isMessageBoxDismiss(e)) return;
    cookieActionError.value = userFacingCookieError(
      e instanceof Error ? e.message : String(e),
    );
    await refreshCookieStatus(gen);
  } finally {
    if (!isCookieViewStale(gen)) cookieBusy.value = false;
  }
}

async function onCookieBrowserChange() {
  if (!cookieEnabled.value || cookieDraftSource.value !== "browser") return;
  await onCookieSourceChange();
}

async function onCookiePathChange() {
  if (!cookieEnabled.value || cookieDraftSource.value !== "file") return;
  await onCookieSourceChange();
}

async function browseCookieFile() {
  const gen = cookieViewGen;
  const picked = await App.openFileDialog(
    t("settings.cookieLinkage.browseTitle"),
    dirOfPath(cookieDraftCookiesPath.value),
    "*.txt",
  );
  if (!picked || isCookieViewStale(gen)) return;
  cookieDraftCookiesPath.value = picked;
  if (cookieEnabled.value) {
    await onCookiePathChange();
  }
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
  try {
    await refreshCookieStatus(++cookieViewGen);
  } catch (e) {
    cookieActionError.value = userFacingCookieError(
      e instanceof Error ? e.message : String(e),
    );
  }
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
    ElMessage.error(formatBackendError(e, t("settings.errOperation")));
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
      selfProfile.value = null;
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

.self-profile-link {
  display: inline-block;
  margin-top: 0.65rem;
  color: var(--el-color-primary);
  text-decoration: none;
  font-size: 0.9rem;
}

.self-profile-link:hover {
  text-decoration: underline;
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

.cookie-always-warn,
.cookie-official-hint,
.cookie-action-error {
  margin-bottom: 0.75rem;
}

.cookie-video-link {
  margin-left: 0.35rem;
}

.cookie-form {
  margin-top: 0.75rem;
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
