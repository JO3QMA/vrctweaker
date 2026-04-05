<template>
  <div class="settings-view">
    <h1 class="page-title">設定</h1>

    <!-- VRChat ログイン -->
    <el-card class="settings-card" shadow="never">
      <template #header>
        <span>VRChat ログイン</span>
      </template>
      <div v-if="isLoggedIn" class="login-status">
        <div v-if="profileLoading && !currentUser" class="profile-loading">
          プロフィールを読み込み中…
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
              {{ currentUser.displayName || "（表示名なし）" }}
            </p>
            <p v-if="currentUser.username" class="current-user-line">
              @{{ currentUser.username }}
            </p>
            <p v-if="currentUser.id" class="current-user-line muted">
              {{ currentUser.id }}
            </p>
            <p class="current-user-line">
              ステータス: {{ currentUser.status || "—" }} /
              {{ currentUser.state || "—" }}
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
        <el-tag type="success" size="large">ログイン済み</el-tag>
        <div class="login-actions">
          <el-button
            type="primary"
            :loading="profileLoading"
            @click="loadCurrentUser(true)"
          >
            プロフィール再取得
          </el-button>
          <el-button type="primary" @click="refreshFriends">
            フレンド一覧を更新
          </el-button>
          <el-button type="danger" plain @click="logout">
            ログアウト
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
          <el-form-item label="ユーザー名">
            <el-input
              id="login-username"
              v-model="loginForm.username"
              placeholder="VRChat ユーザー名"
              autocomplete="username"
            />
          </el-form-item>
          <el-form-item label="パスワード">
            <el-input
              id="login-password"
              v-model="loginForm.password"
              type="password"
              placeholder="パスワード"
              autocomplete="current-password"
              show-password
            />
          </el-form-item>
          <el-form-item label="2FAコード（オプション）">
            <el-input
              id="login-2fa"
              v-model="loginForm.twoFactorCode"
              placeholder="6桁の認証コード"
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
            {{ loginLoading ? "ログイン中..." : "ログイン" }}
          </el-button>
        </el-form>
      </div>
    </el-card>

    <!-- パス設定 -->
    <el-card class="settings-card" shadow="never">
      <template #header>
        <span>パス設定</span>
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
              存在確認
            </el-button>
          </div>
          <el-text
            v-if="validateResult[field.key] !== null"
            :type="validateResult[field.key] ? 'success' : 'danger'"
            size="small"
          >
            {{ validateResult[field.key] ? "存在します" : "存在しません" }}
          </el-text>
        </div>
      </div>
      <el-text type="info" size="small" class="hint">
        VRChatの起動とログ監視で使用します。launch.exeを指定してください（VRChat.exe直接起動はオフラインモードになります）。空の場合はデフォルトパスを使用します。
        output_log は<strong>1ファイル</strong>を指定するか、<code
          >...\VRChat\VRChat</code
        >
        の<strong>フォルダ</strong>を指定してください。フォルダのときは更新日時が最新の
        <code>output_log*.txt</code>
        を自動で選び、VRChat再起動で新しいログファイルができても追従します。
      </el-text>
    </el-card>

    <!-- ログ・データ管理 -->
    <el-card class="settings-card" shadow="never">
      <template #header>
        <span>ログ・データ管理</span>
      </template>
      <div class="setting-row">
        <label>遭遇記録の保存期間（日）</label>
        <el-input-number
          v-model="logRetentionDays"
          :min="1"
          :max="365"
          @change="saveRetention"
        />
      </div>
      <el-text type="info" size="small" class="hint">
        この日数を過ぎたuser_encountersは自動削除されます
      </el-text>
    </el-card>

    <!-- OSS ライセンス -->
    <el-card class="settings-card" shadow="never">
      <template #header>
        <span>OSS ライセンス</span>
      </template>
      <el-text type="info" size="small" class="hint">
        本アプリケーションで使用しているオープンソースソフトウェアのライセンス一覧を確認できます。
      </el-text>
      <div style="margin-top: 0.75rem">
        <router-link class="btn-licenses" to="/licenses">
          <el-button type="primary">OSS ライセンス一覧を表示</el-button>
        </router-link>
      </div>
    </el-card>

    <!-- DB メンテナンス -->
    <el-card class="settings-card" shadow="never">
      <template #header>
        <span>DBメンテナンス</span>
      </template>
      <el-text
        type="info"
        size="small"
        class="hint"
        style="display: block; margin-bottom: 1rem"
      >
        データベースの最適化と、各テーブルのクリア操作を行います。操作前に確認ダイアログが表示されます。
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
          {{ maintenanceLoading ? "実行中..." : "DB最適化 (VACUUM)" }}
        </el-button>
        <el-button
          type="danger"
          plain
          :loading="maintenanceLoading"
          @click="doClearEncounters"
        >
          遭遇ログをクリア
        </el-button>
        <el-button
          type="danger"
          plain
          :loading="maintenanceLoading"
          @click="doClearScreenshots"
        >
          スクショインデックスをクリア
        </el-button>
        <el-button
          type="danger"
          plain
          :loading="maintenanceLoading"
          @click="doClearFriendsCache"
        >
          フレンドキャッシュをクリア
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from "vue";
import { ElMessageBox, ElMessage } from "element-plus";
import { App } from "../wails/app";
import type { PathSettingsDTO, VRChatCurrentUserDTO } from "../wails/app";
import { useSessionUnlock } from "../composables/useSessionUnlock";

const {
  state: unlockState,
  errorMessage: unlockErrorMessage,
  beginStartupUnlock,
  persistAfterLogin,
  handleLogout,
} = useSessionUnlock();

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
    label: "VRChat実行ファイル（Windows）",
    placeholder:
      "例: C:\\Program Files (x86)\\Steam\\steamapps\\common\\VRChat\\launch.exe",
    buttons: [
      {
        label: "参照",
        testid: "vrchat-path-browse",
        title: "ファイルを選択",
        handler: browseVrchatPath,
      },
    ],
  },
  {
    key: "steamPathLinux" as keyof PathSettingsDTO,
    label: "Steamコマンド（Linux）",
    placeholder: "例: steam または /usr/bin/steam",
    buttons: [
      {
        label: "参照",
        testid: "steam-path-browse",
        title: "ファイルを選択",
        handler: browseSteamPath,
      },
    ],
  },
  {
    key: "outputLogPath" as keyof PathSettingsDTO,
    label: "output_log（ファイルまたはフォルダ）",
    placeholder:
      "例: ...\\VRChat\\VRChat\\output_log_....txt または ...\\VRChat\\VRChat フォルダ",
    buttons: [
      {
        label: "ファイル",
        testid: "output-log-path-browse",
        title: "ログファイルを選択",
        handler: browseOutputLogPath,
      },
      {
        label: "フォルダ",
        testid: "output-log-dir-browse",
        title: "VRChat ログフォルダを選択（最新 output_log*.txt に追従）",
        handler: browseOutputLogDirectory,
      },
      {
        label: "ログフォルダを開く",
        testid: "",
        title: "VRChat のログフォルダをファイルマネージャで開く",
        handler: openVRChatLogFolder,
      },
    ],
  },
]);

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
    profileError.value = formatBackendError(
      e,
      "プロフィールの取得に失敗しました",
    );
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
      loginError.value = result.error || "ログインに失敗しました";
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
      loginError.value =
        e instanceof Error ? e.message : "ログアウトに失敗しました";
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
      e instanceof Error ? e.message : "フレンド一覧の更新に失敗しました";
  }
}

async function saveRetention() {
  await App.setLogRetentionDays(logRetentionDays.value);
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
    "VRChat実行ファイルを選択",
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
    "Steam実行ファイルを選択",
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
    "output_log.txt を選択",
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
    "VRChat ログフォルダを選択（output_log*.txt があるフォルダ）",
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
    await ElMessageBox.confirm(message, "確認", {
      confirmButtonText: "実行",
      cancelButtonText: "キャンセル",
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
      : "完了しました";
    if (msg) {
      ElMessage.success(msg);
    }
  } catch (e) {
    maintenanceError.value =
      e instanceof Error ? e.message : "操作に失敗しました";
  } finally {
    maintenanceLoading.value = false;
  }
}

function doVacuumDb() {
  void runWithConfirm(
    "データベースを最適化（VACUUM）します。よろしいですか？",
    async () => {
      await App.vacuumDb();
    },
    () => "DBの最適化が完了しました",
  );
}

function doClearEncounters() {
  void runWithConfirm(
    "遭遇ログ（user_encounters）をすべて削除します。よろしいですか？",
    async () => App.clearEncounters(),
    (n) => `${n}件の遭遇ログを削除しました`,
  );
}

function doClearScreenshots() {
  void runWithConfirm(
    "スクリーンショットインデックス（screenshots）をすべて削除します。よろしいですか？",
    async () => App.clearScreenshots(),
    (n) => `${n}件のスクショインデックスを削除しました`,
  );
}

function doClearFriendsCache() {
  void runWithConfirm(
    "ユーザーキャッシュ（users_cache）の全行（自分・フレンド・遭遇ログ由来のユーザー）を削除します。よろしいですか？",
    async () => App.clearFriendsCache(),
    (n) => {
      currentUser.value = null;
      profileError.value = "";
      return `${n}件のフレンドキャッシュを削除しました`;
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

.maintenance-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}
</style>
