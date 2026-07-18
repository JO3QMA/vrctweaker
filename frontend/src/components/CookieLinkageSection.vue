<template>
  <el-card
    v-if="cookieSupported"
    class="cookie-card"
    shadow="never"
    data-testid="video-cookie-linkage"
  >
    <template #header>
      <span>{{ t("video.cookieLinkage.section") }}</span>
    </template>
    <el-alert
      type="warning"
      :closable="false"
      show-icon
      class="cookie-always-warn"
      :title="t('video.cookieLinkage.alwaysWarn')"
    />
    <el-alert
      v-if="!toolsEffectiveOfficial"
      type="info"
      :closable="false"
      show-icon
      class="cookie-official-hint"
      data-testid="video-cookie-official-hint"
      :title="t('video.cookieLinkage.officialHint')"
    />
    <el-alert
      v-if="cookieSourceKind === 'unsupported'"
      type="warning"
      :closable="false"
      show-icon
      :title="t('video.cookieLinkage.unsupportedForm')"
    />
    <el-alert
      v-if="cookieActionError"
      :title="cookieActionError"
      type="error"
      :closable="false"
      show-icon
      class="cookie-action-error"
    />
    <div class="cookie-switch-row">
      <div class="cookie-toggle-label">
        <span>{{ t("video.cookieLinkage.enableLabel") }}</span>
        <el-text type="info" size="small" class="hint block-hint">{{
          t("video.cookieLinkage.lockHint")
        }}</el-text>
      </div>
      <el-switch
        v-model="cookieEnabled"
        class="cookie-switch"
        data-testid="video-cookie-enable"
        :disabled="cookieBusy"
        @change="onCookieEnableChange"
      />
    </div>
    <el-form label-position="top" size="default" class="cookie-form">
      <el-form-item :label="t('video.cookieLinkage.sourceLabel')">
        <el-radio-group
          v-model="cookieDraftSource"
          data-testid="video-cookie-source"
          :disabled="cookieBusy"
          @change="onCookieSourceChange"
        >
          <el-radio-button value="browser">{{
            t("video.cookieLinkage.sourceBrowser")
          }}</el-radio-button>
          <el-radio-button value="file">{{
            t("video.cookieLinkage.sourceFile")
          }}</el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-form-item
        v-if="cookieDraftSource === 'browser'"
        :label="t('video.cookieLinkage.browserLabel')"
      >
        <el-select
          v-model="cookieDraftBrowser"
          data-testid="video-cookie-browser"
          :disabled="cookieBusy"
          @change="onCookieBrowserChange"
        >
          <el-option value="chrome" label="Chrome" />
          <el-option value="edge" label="Edge" />
          <el-option value="firefox" label="Firefox" />
        </el-select>
      </el-form-item>
      <el-form-item v-else :label="t('video.cookieLinkage.cookiesFileLabel')">
        <div class="path-input-group">
          <el-input
            v-model="cookieDraftCookiesPath"
            data-testid="video-cookie-file-path"
            :placeholder="t('video.cookieLinkage.cookiesFilePh')"
            :disabled="cookieBusy"
            @change="onCookiePathChange"
          />
          <el-button
            data-testid="video-cookie-file-browse"
            :disabled="cookieBusy"
            @click="browseCookieFile"
          >
            {{ t("video.cookieLinkage.browse") }}
          </el-button>
        </div>
      </el-form-item>
    </el-form>
  </el-card>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { ElMessageBox } from "element-plus";
import { App, type CookieLinkageStatusDTO } from "../wails/app";
import { cookieLinkageErrorI18nKey } from "../views/cookieLinkageErrors";

const { t } = useI18n();

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
    t("video.cookieLinkage.riskAckBody"),
    t("video.cookieLinkage.riskAckTitle"),
    {
      confirmButtonText: t("video.cookieLinkage.riskAckConfirm"),
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

async function onCookieEnableChange(on: boolean) {
  const desired = on;
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

function dirOfPath(p: string): string {
  if (!p) return "";
  const sep = p.includes("\\") ? "\\" : "/";
  const idx = p.lastIndexOf(sep);
  return idx >= 0 ? p.slice(0, idx) : "";
}

async function browseCookieFile() {
  const gen = ++cookieViewGen;
  const picked = await App.openFileDialog(
    t("video.cookieLinkage.browseTitle"),
    dirOfPath(cookieDraftCookiesPath.value),
    "*.txt",
  );
  if (!picked || isCookieViewStale(gen)) return;
  cookieDraftCookiesPath.value = picked;
  if (cookieEnabled.value) {
    await onCookiePathChange();
  }
}

onBeforeUnmount(() => {
  cookieViewGen++;
});

onMounted(async () => {
  try {
    await refreshCookieStatus(++cookieViewGen);
  } catch (e) {
    cookieActionError.value = userFacingCookieError(
      e instanceof Error ? e.message : String(e),
    );
  }
});
</script>

<style scoped>
.cookie-card {
  margin-top: 1rem;
  width: 100%;
}
.cookie-always-warn,
.cookie-official-hint,
.cookie-action-error {
  margin-bottom: 0.75rem;
}
.cookie-switch-row {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 0.5rem;
}
.cookie-toggle-label {
  flex: 1;
  min-width: 0;
}
.block-hint {
  display: block;
  margin-top: 0.35rem;
}
.cookie-switch {
  flex-shrink: 0;
  margin-top: 0.15rem;
}
.cookie-form {
  margin-top: 0.75rem;
}
.path-input-group {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  flex-wrap: wrap;
  width: 100%;
}
.path-input-group :deep(.el-input) {
  flex: 1;
  min-width: 0;
}
.hint {
  color: var(--text-secondary);
}
</style>
