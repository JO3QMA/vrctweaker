<template>
  <div data-testid="settings-account-panel">
    <SettingsPageHead
      :title="t('settings.navAccount')"
      :description="t('settings.accountLead')"
    />
    <el-card class="section-card" shadow="never">
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
        <VtTag class="login-status-tag" variant="success" size="large">{{
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
            <LoginTwoFactorCodeField
              v-model="loginForm.twoFactorCode"
              input-id="login-2fa"
              :placeholder="t('settings.twoFactorPh')"
              :disabled="loginLoading"
              @complete="onTwoFactorCodeComplete"
            />
          </el-form-item>
          <p
            v-if="twoFactorRequired && !loginError"
            class="login-2fa-hint text-body-sm"
          >
            {{ t("settings.twoFactorRequiredHint") }}
          </p>
          <VtAlert
            v-if="loginError"
            variant="danger"
            :title="loginError"
            class="login-error"
          />
          <VtButton
            variant="primary"
            :loading="loginLoading"
            :disabled="loginSubmitDisabled"
            @click="login"
          >
            {{ loginLoading ? t("settings.loggingIn") : t("settings.login") }}
          </VtButton>
        </el-form>
        <p
          class="credential-keyring-note text-body-sm"
          data-testid="settings-credential-keyring"
        >
          {{ t("settings.credentialKeyringNote") }}
        </p>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import LoginTwoFactorCodeField from "../LoginTwoFactorCodeField.vue";
import VtAlert from "../VtAlert.vue";
import VtButton from "../VtButton.vue";
import VtTag from "../VtTag.vue";
import SettingsPageHead from "./SettingsPageHead.vue";
import { useSessionUnlock } from "../../composables/useSessionUnlock";
import { friendThumbUrl } from "../../utils/vrcUserCacheDisplay";
import { isLoginBlockedByTwoFactorCode } from "../../utils/loginTwoFactorSubmit";
import { isLoginTwoFactorRequired } from "../../utils/vrchatLoginResult";
import { App } from "../../wails/app";
import type { UserCacheDTO } from "../../wails/app";
import { formatBackendError } from "../../views/settings/formatBackendError";

const {
  state: unlockState,
  errorMessage: unlockErrorMessage,
  beginStartupUnlock,
  persistAfterLogin,
  handleLogout,
} = useSessionUnlock();

const { t } = useI18n();

const isLoggedIn = ref(false);
const selfProfile = ref<UserCacheDTO | null>(null);
const profileLoading = ref(false);
const profileError = ref("");

const avatarDisplayUrl = computed(() => {
  const u = selfProfile.value;
  if (!u) return "";
  return friendThumbUrl(u) ?? "";
});

const loginForm = reactive({
  username: "",
  password: "",
  twoFactorCode: "",
});
const loginError = ref("");
const loginLoading = ref(false);
const twoFactorRequired = ref(false);

const loginSubmitDisabled = computed(
  () =>
    loginLoading.value ||
    !loginForm.username ||
    !loginForm.password ||
    isLoginBlockedByTwoFactorCode(
      loginForm.twoFactorCode,
      twoFactorRequired.value,
    ),
);

function onTwoFactorCodeComplete(): void {
  if (loginSubmitDisabled.value) return;
  void login();
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
      twoFactorRequired.value = false;
      loginForm.username = "";
      loginForm.password = "";
      loginForm.twoFactorCode = "";
      if (result.plaintextToken) {
        await persistAfterLogin(result.plaintextToken);
      }
      await loadSelfProfileSummary();
    } else if (isLoginTwoFactorRequired(result.error)) {
      twoFactorRequired.value = true;
      loginError.value = "";
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
</script>
