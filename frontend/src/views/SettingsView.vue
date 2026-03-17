<template>
  <div class="settings-view">
    <h1 class="page-title">設定</h1>
    <section class="settings-section">
      <h2>VRChat ログイン</h2>
      <div v-if="isLoggedIn" class="login-status">
        <span class="logged-in-label">ログイン済み</span>
        <div class="login-actions">
          <button type="button" class="btn-refresh" @click="refreshFriends">
            フレンド一覧を更新
          </button>
          <button type="button" class="btn-logout" @click="logout">
            ログアウト
          </button>
        </div>
      </div>
      <div v-else class="login-form">
        <div class="form-row">
          <label for="login-username">ユーザー名</label>
          <input
            id="login-username"
            v-model="loginForm.username"
            type="text"
            placeholder="VRChat ユーザー名"
            autocomplete="username"
          />
        </div>
        <div class="form-row">
          <label for="login-password">パスワード</label>
          <input
            id="login-password"
            v-model="loginForm.password"
            type="password"
            placeholder="パスワード"
            autocomplete="current-password"
          />
        </div>
        <div class="form-row">
          <label for="login-2fa">2FAコード（オプション）</label>
          <input
            id="login-2fa"
            v-model="loginForm.twoFactorCode"
            type="text"
            placeholder="6桁の認証コード"
            autocomplete="one-time-code"
          />
        </div>
        <p v-if="loginError" class="login-error">
          {{ loginError }}
        </p>
        <button
          type="button"
          class="btn-login"
          :disabled="loginLoading || !loginForm.username || !loginForm.password"
          @click="login"
        >
          {{ loginLoading ? "ログイン中..." : "ログイン" }}
        </button>
      </div>
    </section>
    <section class="settings-section">
      <h2>パス設定</h2>
      <div class="path-settings">
        <div class="path-row">
          <label for="vrchat-path">VRChat実行ファイル（Windows）</label>
          <div class="path-input-group">
            <input
              id="vrchat-path"
              v-model="pathSettings.vrchatPathWindows"
              type="text"
              placeholder="例: C:\Program Files (x86)\Steam\steamapps\common\VRChat\launch.exe"
              @change="savePathSettings"
            />
            <button
              type="button"
              class="btn-validate"
              :disabled="!pathSettings.vrchatPathWindows"
              @click="validatePathField('vrchatPathWindows')"
            >
              存在確認
            </button>
          </div>
          <span
            v-if="validateResult.vrchatPathWindows !== null"
            :class="
              validateResult.vrchatPathWindows ? 'validate-ok' : 'validate-ng'
            "
          >
            {{
              validateResult.vrchatPathWindows ? "存在します" : "存在しません"
            }}
          </span>
        </div>
        <div class="path-row">
          <label for="steam-path">Steamコマンド（Linux）</label>
          <div class="path-input-group">
            <input
              id="steam-path"
              v-model="pathSettings.steamPathLinux"
              type="text"
              placeholder="例: steam または /usr/bin/steam"
              @change="savePathSettings"
            />
            <button
              type="button"
              class="btn-validate"
              :disabled="!pathSettings.steamPathLinux"
              @click="validatePathField('steamPathLinux')"
            >
              存在確認
            </button>
          </div>
          <span
            v-if="validateResult.steamPathLinux !== null"
            :class="
              validateResult.steamPathLinux ? 'validate-ok' : 'validate-ng'
            "
          >
            {{ validateResult.steamPathLinux ? "存在します" : "存在しません" }}
          </span>
        </div>
        <div class="path-row">
          <label for="output-log-path">output_log.txt</label>
          <div class="path-input-group">
            <input
              id="output-log-path"
              v-model="pathSettings.outputLogPath"
              type="text"
              placeholder="例: C:\Users\...\AppData\LocalLow\VRChat\VRChat\output_log.txt"
              @change="savePathSettings"
            />
            <button
              type="button"
              class="btn-validate"
              :disabled="!pathSettings.outputLogPath"
              @click="validatePathField('outputLogPath')"
            >
              存在確認
            </button>
          </div>
          <span
            v-if="validateResult.outputLogPath !== null"
            :class="
              validateResult.outputLogPath ? 'validate-ok' : 'validate-ng'
            "
          >
            {{ validateResult.outputLogPath ? "存在します" : "存在しません" }}
          </span>
        </div>
      </div>
      <p class="hint">
        VRChatの起動とログ監視で使用します。launch.exeを指定してください（VRChat.exe直接起動はオフラインモードになります）。空の場合はデフォルトパスを使用します。
      </p>
    </section>
    <section class="settings-section">
      <h2>ログ・データ管理</h2>
      <div class="setting-row">
        <label>遭遇記録の保存期間（日）</label>
        <input
          v-model.number="logRetentionDays"
          type="number"
          min="1"
          max="365"
          @change="saveRetention"
        />
      </div>
      <p class="hint">この日数を過ぎたuser_encountersは自動削除されます</p>
    </section>
    <section class="settings-section">
      <h2>DBメンテナンス</h2>
      <p class="hint" style="margin-bottom: 1rem">
        データベースの最適化と、各テーブルのクリア操作を行います。操作前に確認ダイアログが表示されます。
      </p>
      <p v-if="maintenanceError" class="maintenance-error">
        {{ maintenanceError }}
      </p>
      <div class="maintenance-actions">
        <button
          type="button"
          class="btn-maintenance"
          :disabled="maintenanceLoading"
          @click="doVacuumDb"
        >
          {{ maintenanceLoading ? "実行中..." : "DB最適化 (VACUUM)" }}
        </button>
        <button
          type="button"
          class="btn-maintenance btn-maintenance-danger"
          :disabled="maintenanceLoading"
          @click="doClearEncounters"
        >
          遭遇ログをクリア
        </button>
        <button
          type="button"
          class="btn-maintenance btn-maintenance-danger"
          :disabled="maintenanceLoading"
          @click="doClearScreenshots"
        >
          スクショインデックスをクリア
        </button>
        <button
          type="button"
          class="btn-maintenance btn-maintenance-danger"
          :disabled="maintenanceLoading"
          @click="doClearFriendsCache"
        >
          フレンドキャッシュをクリア
        </button>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { App } from "../wails/app";
import type { PathSettingsDTO } from "../wails/app";

const isLoggedIn = ref(false);
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

onMounted(async () => {
  try {
    isLoggedIn.value = await App.isLoggedIn();
  } catch {
    isLoggedIn.value = false;
  }
  logRetentionDays.value = await App.getLogRetentionDays();
  const ps = await App.getPathSettings();
  pathSettings.vrchatPathWindows = ps.vrchatPathWindows;
  pathSettings.steamPathLinux = ps.steamPathLinux;
  pathSettings.outputLogPath = ps.outputLogPath;
});

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
    } else {
      loginError.value = result.error || "ログインに失敗しました";
    }
  } finally {
    loginLoading.value = false;
  }
}

async function logout() {
  loginError.value = "";
  try {
    await App.logout();
    isLoggedIn.value = false;
  } catch (e) {
    loginError.value =
      e instanceof Error ? e.message : "ログアウトに失敗しました";
  }
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

async function validatePathField(field: keyof PathSettingsDTO) {
  const path = pathSettings[field];
  if (path === "") {
    validateResult[field] = false;
    return;
  }
  validateResult[field] = await App.validatePath(path);
}

function runWithConfirm(
  message: string,
  fn: () => Promise<number | void>,
  successMessage?: (result?: number) => string,
) {
  if (!window.confirm(message)) {
    return;
  }
  maintenanceError.value = "";
  maintenanceLoading.value = true;
  fn()
    .then((result) => {
      const msg = successMessage
        ? successMessage(typeof result === "number" ? result : undefined)
        : "完了しました";
      if (msg) {
        maintenanceError.value = "";
        window.alert(msg);
      }
    })
    .catch((e) => {
      maintenanceError.value =
        e instanceof Error ? e.message : "操作に失敗しました";
    })
    .finally(() => {
      maintenanceLoading.value = false;
    });
}

function doVacuumDb() {
  runWithConfirm(
    "データベースを最適化（VACUUM）します。よろしいですか？",
    async () => {
      await App.vacuumDb();
    },
    () => "DBの最適化が完了しました",
  );
}

function doClearEncounters() {
  runWithConfirm(
    "遭遇ログ（user_encounters）をすべて削除します。よろしいですか？",
    async () => App.clearEncounters(),
    (n) => `${n}件の遭遇ログを削除しました`,
  );
}

function doClearScreenshots() {
  runWithConfirm(
    "スクリーンショットインデックス（screenshots）をすべて削除します。よろしいですか？",
    async () => App.clearScreenshots(),
    (n) => `${n}件のスクショインデックスを削除しました`,
  );
}

function doClearFriendsCache() {
  runWithConfirm(
    "フレンドキャッシュ（friends_cache）をすべて削除します。よろしいですか？",
    async () => App.clearFriendsCache(),
    (n) => `${n}件のフレンドキャッシュを削除しました`,
  );
}
</script>

<style scoped>
.settings-section {
  margin-bottom: 2rem;
}
.settings-section h2 {
  font-size: 1.1rem;
  margin: 0 0 1rem;
}
.setting-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 0.5rem;
}
.setting-row input {
  width: 80px;
  padding: 0.4rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
}

.path-settings {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.path-row {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}
.path-row label {
  font-size: 0.95rem;
}
.path-input-group {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}
.path-input-group input {
  flex: 1;
  min-width: 0;
  padding: 0.4rem 0.6rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
}
.btn-validate {
  flex-shrink: 0;
  padding: 0.4rem 0.75rem;
  background: var(--accent);
  color: var(--bg-primary);
  border: none;
  border-radius: var(--radius);
  cursor: pointer;
}
.btn-validate:hover:not(:disabled) {
  opacity: 0.9;
}
.btn-validate:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.validate-ok {
  font-size: 0.85rem;
  color: var(--success, #22c55e);
}
.validate-ng {
  font-size: 0.85rem;
  color: var(--error, #ef4444);
}

.hint {
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin-top: 0.5rem;
}

/* Login section */
.login-status {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.logged-in-label {
  font-size: 0.95rem;
  color: var(--success, #22c55e);
}
.login-actions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}
.btn-login,
.btn-logout,
.btn-refresh {
  padding: 0.5rem 1rem;
  border-radius: var(--radius);
  cursor: pointer;
  border: none;
  font-size: 0.9rem;
}
.btn-login,
.btn-refresh {
  background: var(--accent);
  color: var(--bg-primary);
}
.btn-logout {
  background: var(--bg-tertiary);
  color: var(--text-primary);
  border: 1px solid var(--border);
}
.btn-login:hover:not(:disabled),
.btn-refresh:hover,
.btn-logout:hover {
  opacity: 0.9;
}
.btn-login:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  max-width: 360px;
}
.login-form .form-row {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}
.login-form .form-row label {
  font-size: 0.95rem;
}
.login-form .form-row input {
  padding: 0.5rem 0.6rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
}
.login-error {
  font-size: 0.9rem;
  color: var(--error, #ef4444);
  margin: 0;
}

/* DB Maintenance */
.maintenance-error {
  font-size: 0.9rem;
  color: var(--error, #ef4444);
  margin: 0 0 0.5rem;
}
.maintenance-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}
.btn-maintenance {
  padding: 0.5rem 1rem;
  border-radius: var(--radius);
  cursor: pointer;
  border: 1px solid var(--border);
  font-size: 0.9rem;
  background: var(--bg-tertiary);
  color: var(--text-primary);
}
.btn-maintenance:hover:not(:disabled) {
  opacity: 0.9;
}
.btn-maintenance:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.btn-maintenance-danger {
  border-color: var(--error, #ef4444);
  color: var(--error, #ef4444);
}
</style>
