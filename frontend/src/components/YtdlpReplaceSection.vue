<template>
  <section
    data-testid="ytdlp-replace-section"
    aria-labelledby="ytdlp-replace-heading"
  >
    <h2 id="ytdlp-replace-heading" class="video-block-title">
      {{ t("video.replaceLabel") }}
    </h2>

    <div v-if="loading" class="vt-text-muted">{{ t("common.loading") }}</div>
    <template v-else>
      <div class="video-alerts" data-testid="ytdlp-alert-area">
        <VtAlert
          v-if="topBannerError"
          variant="danger"
          class="block-hint"
          data-testid="ytdlp-error-banner"
          :title="topBannerError"
        />
        <VtAlert
          v-else-if="!status.supported"
          variant="warning"
          :title="userFacingReason(status.unsupportedReason ?? '')"
        />
        <template v-else>
          <VtAlert
            variant="warning"
            class="block-hint"
            :title="t('video.alwaysWarn')"
          />
          <VtAlert
            v-if="bannerError"
            variant="danger"
            class="block-hint"
            data-testid="ytdlp-error-banner"
            :title="bannerError"
          />
        </template>
      </div>

      <template v-if="status.supported">
        <section class="video-ops" data-testid="ytdlp-ops">
          <div class="video-switch-row">
            <el-switch
              v-model="maintainOn"
              data-testid="ytdlp-maintain-switch"
              :disabled="busy"
              @change="onMaintainChange"
            />
            <span class="switch-status" data-testid="ytdlp-effective-inline">
              {{ t("video.statusPrefix") }}{{ effectiveStatusText }}
            </span>
          </div>

          <p
            class="version-summary text-body-sm"
            data-testid="ytdlp-version-summary"
          >
            <span
              class="version-summary-segment"
              data-testid="ytdlp-version-current"
              >{{
                t("video.versionLineCurrent", {
                  version: currentVersionDisplay,
                })
              }}</span
            >
            <span class="version-sep" aria-hidden="true">・</span>
            <span class="version-summary-segment">
              <template v-if="!githubBackgroundCheckFailed">
                {{ t("video.versionLineGithubLabel") }}
              </template>
              <span
                data-testid="ytdlp-version-github"
                :class="githubVersionClass"
                >{{ githubVersionDisplay }}</span
              >
            </span>
            <span class="version-sep" aria-hidden="true">・</span>
            <span
              class="version-summary-segment"
              data-testid="ytdlp-version-last-check"
              >{{
                t("video.versionLineLastCheck", { when: lastCheckDisplay })
              }}</span
            >
          </p>

          <div class="video-action-stack">
            <div class="video-update-row" data-testid="ytdlp-update-row">
              <VtButton
                variant="primary"
                class="video-update-btn"
                data-testid="ytdlp-check-update"
                :loading="actionLoading"
                :disabled="busy"
                @click="checkAndUpdate"
              >
                {{ actionButtonLabel }}
              </VtButton>
            </div>

            <div class="video-folder-row" data-testid="ytdlp-folder-row">
              <VtButton
                variant="secondary"
                class="video-folder-btn"
                data-testid="ytdlp-open-cache-folder"
                :disabled="busy"
                @click="openCacheFolder"
              >
                <VtIcon size="default" class="btn-icon"
                  ><FolderOpened
                /></VtIcon>
                {{ t("video.openCacheFolder") }}
              </VtButton>
              <VtButton
                variant="secondary"
                class="video-folder-btn"
                data-testid="ytdlp-open-tools-folder"
                :disabled="busy"
                @click="openToolsFolder"
              >
                <VtIcon size="default" class="btn-icon"
                  ><FolderOpened
                /></VtIcon>
                {{ t("video.openToolsFolder") }}
              </VtButton>
            </div>
          </div>

          <p
            v-if="resultLine.text"
            class="flash"
            :class="resultLine.className"
            data-testid="ytdlp-result-line"
          >
            {{ resultLine.text }}
          </p>

          <div
            v-if="showErrorDetailsToggle"
            class="error-details"
            data-testid="ytdlp-error-details"
          >
            <button
              type="button"
              class="error-details-toggle"
              data-testid="ytdlp-error-details-toggle"
              :aria-expanded="errorDetailsExpanded"
              :aria-controls="errorDetailsPanelId"
              @click="errorDetailsExpanded = !errorDetailsExpanded"
            >
              <VtIcon
                size="default"
                class="error-details-toggle-icon"
                decorative
              >
                <CaretBottom v-if="errorDetailsExpanded" />
                <CaretRight v-else />
              </VtIcon>
              <span>{{ t("video.errorDetailsToggle") }}</span>
            </button>
            <div
              v-show="errorDetailsExpanded"
              :id="errorDetailsPanelId"
              class="error-details-panel"
              role="region"
            >
              <p class="detail-error-raw" data-testid="ytdlp-detail-error">
                {{ detailErrorRaw }}
              </p>
            </div>
          </div>
        </section>
      </template>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { ElMessageBox } from "element-plus";
import { CaretBottom, CaretRight, FolderOpened } from "@element-plus/icons-vue";
import VtAlert from "./VtAlert.vue";
import VtButton from "./VtButton.vue";
import VtIcon from "./VtIcon.vue";
import {
  VT_BUTTON_DANGER_CONFIRM_CLASS,
  VT_BUTTON_SECONDARY_CANCEL_CLASS,
} from "./vtButtonClasses";
import {
  App,
  type YTDLPMaintainActionResultDTO,
  type YTDLPMaintainStatusDTO,
} from "../wails/app";
import { appLocaleToBcp47 } from "../i18n";
import { videoErrorI18nKey, classifyVideoError } from "../views/videoErrors";
import { isYtdlpVersionOlder } from "../utils/ytdlpVersion";
import {
  isYtdlpMaintainOutcome,
  YTDLP_MAINTAIN_OUTCOME,
} from "../utils/ytdlpMaintainOutcome";
import { showToast } from "../utils/showToast";

const { t, te, locale } = useI18n();

const emptyStatus = (): YTDLPMaintainStatusDTO => ({
  supported: false,
  unsupportedReason: "",
  maintainDesired: false,
  riskAcknowledged: false,
  effectiveOfficial: false,
  cachePresent: false,
  cacheVersion: "",
  toolsPath: "",
  cachePath: "",
  pendingError: "",
  latestVersion: "",
  latestTag: "",
  latestDownloadUrl: "",
  latestError: "",
  latestCheckedAt: "",
});

const status = ref<YTDLPMaintainStatusDTO>(emptyStatus());
const maintainOn = ref(false);
const loading = ref(true);
const busy = ref(false);
const actionPhase = ref<"idle" | "checking" | "updating">("idle");
const actionError = ref("");
const detailErrorRaw = ref("");
const resultOutcome = ref<
  "" | "firstDownload" | "updated" | "unchanged" | "error"
>("");
const resultPreviousVersion = ref("");
const resultNewVersion = ref("");
const errorDetailsExpanded = ref(false);
const errorDetailsPanelId = "ytdlp-error-details-panel";

let sectionGeneration = 0;
function isSectionStale(gen: number): boolean {
  return gen !== sectionGeneration;
}

const effectiveStatusText = computed(() =>
  status.value.effectiveOfficial
    ? t("video.effectiveOfficial")
    : t("video.effectiveBundled"),
);

const showInlineResultError = computed(
  () =>
    resultOutcome.value === YTDLP_MAINTAIN_OUTCOME.error && !!actionError.value,
);

const topBannerError = computed(() => {
  if (showInlineResultError.value) return "";
  return actionError.value;
});

const bannerError = computed(() => {
  if (topBannerError.value) return topBannerError.value;
  if (status.value.pendingError) {
    return userFacingError(status.value.pendingError);
  }
  return "";
});

const actionLoading = computed(() => actionPhase.value !== "idle");

const actionButtonLabel = computed(() => {
  if (actionPhase.value === "checking") {
    return t("video.actionChecking");
  }
  if (actionPhase.value === "updating") {
    return t("video.actionUpdating");
  }
  return status.value.cachePresent
    ? t("video.actionUpdate")
    : t("video.actionFetch");
});

const updateAvailable = computed(() => {
  if (status.value.latestError) return false;
  const latest = status.value.latestVersion;
  if (!latest) return false;
  if (!status.value.cachePresent) return true;
  return isYtdlpVersionOlder(status.value.cacheVersion, latest);
});

const githubBackgroundCheckFailed = computed(
  () => !!status.value.latestError?.trim(),
);

const githubVersionDisplay = computed(() => {
  if (githubBackgroundCheckFailed.value) {
    return t("video.githubCheckUnavailable");
  }
  return status.value.latestVersion
    ? status.value.latestVersion
    : t("video.latestUnknown");
});

const githubVersionClass = computed(() => ({
  "version-github-new": updateAvailable.value,
  "version-github-unavailable": githubBackgroundCheckFailed.value,
}));

const currentVersionDisplay = computed(() => {
  if (!status.value.cachePresent) {
    return t("video.currentNotDownloaded");
  }
  if (status.value.cacheVersion) {
    return status.value.cacheVersion;
  }
  return t("video.currentNotDownloaded");
});

const showErrorDetailsToggle = computed(
  () => resultOutcome.value === "error" && detailErrorRaw.value !== "",
);

const resultLine = computed(() => {
  switch (resultOutcome.value) {
    case "updated":
      return {
        text: t("video.resultUpdated", {
          old: resultPreviousVersion.value,
          new: resultNewVersion.value,
        }),
        className: "flash-ok",
      };
    case "firstDownload":
      return {
        text: t("video.resultFirstDownload", {
          version: resultNewVersion.value,
        }),
        className: "flash-ok",
      };
    case "unchanged":
      return {
        text: t("video.resultUpToDate", { version: resultNewVersion.value }),
        className: "flash-secondary",
      };
    case "error":
      return {
        text: actionError.value,
        className: "flash-err",
      };
    default:
      return { text: "", className: "" };
  }
});

const lastCheckDisplay = computed(() => {
  const raw = status.value.latestCheckedAt;
  if (!raw) return t("video.lastCheckNever");
  const d = new Date(raw);
  if (Number.isNaN(d.getTime())) return t("video.lastCheckNever");
  return new Intl.DateTimeFormat(appLocaleToBcp47(String(locale.value)), {
    month: "numeric",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  }).format(d);
});

function userFacingReason(code: string): string {
  if (!code) return t("video.unsupported");
  const key = `video.reason.${code}`;
  return te(key) ? t(key) : t("video.unsupported");
}

function userFacingError(raw: string): string {
  if (!raw) return "";
  if (te(`video.${raw}`)) return t(`video.${raw}`);
  return t(videoErrorI18nKey(raw));
}

function userFacingActionError(raw: string): string {
  const kind = classifyVideoError(raw);
  if (kind === "githubRateLimit" || kind === "githubForbidden") {
    return t("video.errors.tryAgainLater");
  }
  return userFacingError(raw);
}

function applyStatus(
  s: YTDLPMaintainStatusDTO,
  opts?: { syncSwitch?: boolean },
) {
  const prev = status.value;
  status.value = {
    ...s,
    cachePresent: s.cachePresent,
    cacheVersion: s.cachePresent ? s.cacheVersion || prev.cacheVersion : "",
    latestVersion: s.latestVersion || prev.latestVersion,
    latestTag: s.latestTag || prev.latestTag,
    latestDownloadUrl: s.latestDownloadUrl || prev.latestDownloadUrl,
    latestCheckedAt: s.latestCheckedAt || prev.latestCheckedAt,
  };
  if (opts?.syncSwitch !== false) {
    maintainOn.value = !!s.maintainDesired;
  }
}

function clearFeedback() {
  actionError.value = "";
  resultOutcome.value = "";
  resultPreviousVersion.value = "";
  resultNewVersion.value = "";
  detailErrorRaw.value = "";
  errorDetailsExpanded.value = false;
}

function isMessageBoxDismiss(e: unknown): boolean {
  if (e === "cancel" || e === "close") return true;
  if (e instanceof Error) {
    const m = e.message.toLowerCase();
    return m === "cancel" || m === "close";
  }
  return false;
}

async function refreshSilent(gen: number): Promise<boolean> {
  try {
    if (isSectionStale(gen)) return false;
    applyStatus(await App.getYTDLPMaintainStatus());
    return !isSectionStale(gen);
  } catch (e) {
    if (!isSectionStale(gen)) {
      actionError.value = userFacingError(
        e instanceof Error ? e.message : String(e),
      );
    }
    return false;
  }
}

async function refreshBackgroundLatest(gen: number): Promise<void> {
  try {
    if (isSectionStale(gen)) return;
    applyStatus(await App.refreshYTDLPLatestRelease(false), {
      syncSwitch: false,
    });
  } catch {
    // Background refresh failure: show GitHub確認できませんでした in the summary; keep last successful 最終確認 time.
  }
}

async function refresh() {
  const gen = sectionGeneration;
  loading.value = true;
  clearFeedback();
  try {
    if (isSectionStale(gen)) return;
    applyStatus(await App.getYTDLPMaintainStatus());
    if (isSectionStale(gen)) return;
    void refreshBackgroundLatest(gen);
  } catch (e) {
    if (isSectionStale(gen)) return;
    actionError.value = userFacingError(
      e instanceof Error ? e.message : String(e),
    );
  } finally {
    if (!isSectionStale(gen)) loading.value = false;
  }
}

async function onMaintainChange(on: boolean) {
  const gen = sectionGeneration;
  const desired = on;
  busy.value = true;
  clearFeedback();
  try {
    if (desired && !status.value.riskAcknowledged) {
      await ElMessageBox.confirm(
        t("video.riskAckBody"),
        t("video.riskAckTitle"),
        {
          confirmButtonText: t("video.riskAckConfirm"),
          cancelButtonText: t("common.cancel"),
          type: "warning",
          confirmButtonClass: VT_BUTTON_DANGER_CONFIRM_CLASS,
          cancelButtonClass: VT_BUTTON_SECONDARY_CANCEL_CLASS,
        },
      );
      if (isSectionStale(gen)) return;
      await App.acknowledgeYTDLPToolsReplaceRisk();
    }
    if (isSectionStale(gen)) return;
    await App.setYTDLPToolsReplaceMaintain(desired);
    if (isSectionStale(gen)) return;
    try {
      applyStatus(await App.getYTDLPMaintainStatus());
    } catch {
      maintainOn.value = desired;
      status.value = {
        ...status.value,
        maintainDesired: desired,
        riskAcknowledged: desired ? true : status.value.riskAcknowledged,
        effectiveOfficial: desired,
        pendingError: desired ? "" : status.value.pendingError,
      };
      await refreshSilent(gen);
    }
    if (isSectionStale(gen)) return;
    showToast.success(
      desired ? t("video.flashEnabled") : t("video.flashDisabled"),
    );
  } catch (e) {
    if (isSectionStale(gen)) return;
    maintainOn.value = status.value.maintainDesired;
    if (isMessageBoxDismiss(e)) {
      return;
    }
    actionError.value = userFacingError(
      e instanceof Error ? e.message : String(e),
    );
  } finally {
    if (!isSectionStale(gen)) busy.value = false;
  }
}

function applyActionResult(res: YTDLPMaintainActionResultDTO) {
  applyStatus(res.status, { syncSwitch: false });
  resultPreviousVersion.value = res.previousVersion ?? "";
  resultNewVersion.value = res.newVersion ?? "";
  detailErrorRaw.value = res.detailError ?? "";
  const outcome = res.outcome ?? "";
  if (!isYtdlpMaintainOutcome(outcome)) {
    resultOutcome.value = YTDLP_MAINTAIN_OUTCOME.error;
    const detail = outcome
      ? `unexpected outcome: ${outcome}`
      : res.detailError || "unexpected outcome";
    detailErrorRaw.value = detail;
    actionError.value = userFacingActionError(res.detailError || detail);
    return;
  }
  if (outcome === YTDLP_MAINTAIN_OUTCOME.error) {
    resultOutcome.value = YTDLP_MAINTAIN_OUTCOME.error;
    const raw = res.detailError || status.value.latestError;
    actionError.value = userFacingActionError(raw);
    return;
  }
  resultOutcome.value = outcome;
  actionError.value = "";
}

async function checkAndUpdate() {
  const gen = sectionGeneration;
  busy.value = true;
  actionPhase.value = "checking";
  clearFeedback();
  const phaseTimer = setTimeout(() => {
    if (!isSectionStale(gen) && actionPhase.value === "checking") {
      actionPhase.value = "updating";
    }
  }, 400);
  try {
    if (isSectionStale(gen)) return;
    const res = await App.checkAndUpdateOfficialYTDLPCache();
    if (isSectionStale(gen)) return;
    applyActionResult(res);
  } catch (e) {
    if (isSectionStale(gen)) return;
    resultOutcome.value = YTDLP_MAINTAIN_OUTCOME.error;
    actionError.value = userFacingActionError(
      e instanceof Error ? e.message : String(e),
    );
  } finally {
    clearTimeout(phaseTimer);
    if (!isSectionStale(gen)) {
      actionPhase.value = "idle";
      busy.value = false;
    }
  }
}

async function openCacheFolder() {
  const gen = sectionGeneration;
  busy.value = true;
  clearFeedback();
  try {
    if (isSectionStale(gen)) return;
    await App.openYTDLPCacheFolder();
  } catch (e) {
    if (isSectionStale(gen)) return;
    actionError.value = userFacingError(
      e instanceof Error ? e.message : String(e),
    );
  } finally {
    if (!isSectionStale(gen)) busy.value = false;
  }
}

async function openToolsFolder() {
  const gen = sectionGeneration;
  busy.value = true;
  clearFeedback();
  try {
    if (isSectionStale(gen)) return;
    await App.openYTDLPToolsFolder();
  } catch (e) {
    if (isSectionStale(gen)) return;
    actionError.value = userFacingError(
      e instanceof Error ? e.message : String(e),
    );
  } finally {
    if (!isSectionStale(gen)) busy.value = false;
  }
}

onBeforeUnmount(() => {
  sectionGeneration++;
});

onMounted(() => {
  void refresh();
});

/** Storybook / tests: inject status without mount side effects */
defineExpose({
  applyStatusForTest: applyStatus,
  setLoadingForTest: (v: boolean) => {
    loading.value = v;
  },
  setActionPhaseForTest: (p: "idle" | "checking" | "updating") => {
    actionPhase.value = p;
  },
  setResultForTest: (
    outcome: typeof resultOutcome.value,
    extra?: {
      previous?: string;
      next?: string;
      error?: string;
    },
  ) => {
    resultOutcome.value = outcome;
    resultPreviousVersion.value = extra?.previous ?? "";
    resultNewVersion.value = extra?.next ?? "";
    detailErrorRaw.value = extra?.error ?? "";
    if (outcome === YTDLP_MAINTAIN_OUTCOME.error) {
      actionError.value = userFacingActionError(extra?.error ?? "generic");
    }
  },
});
</script>

<style scoped>
.video-block-title {
  margin: 0 0 var(--space-block);
  font-size: var(--font-size-16);
  font-weight: var(--font-weight-600);
}
.video-alerts {
  margin-bottom: var(--space-block);
}
.block-hint {
  margin-bottom: var(--space-action-group);
}
.video-ops {
  margin-bottom: var(--space-section);
}
.video-switch-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-action-group) var(--space-block);
  margin-bottom: var(--space-action-group);
}
.switch-status {
  color: var(--color-text-secondary);
}
.version-summary {
  margin: 0 0 var(--space-action-group);
  color: var(--color-text-secondary);
}
.version-summary-segment {
  white-space: nowrap;
}
.version-sep {
  color: var(--color-text-secondary);
}
.version-github-new {
  color: var(--color-brand);
  font-weight: var(--font-weight-500);
}
.version-github-unavailable {
  color: var(--color-text-secondary);
}
.video-action-stack {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: var(--space-action-group);
  margin-bottom: 0;
}
.video-update-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-action-group) var(--space-block);
}
.video-update-row :deep(.video-update-btn.el-button--primary:focus) {
  outline: none;
}
.video-update-row
  :deep(.video-update-btn.el-button--primary:focus:not(:focus-visible)) {
  box-shadow: none;
}
.video-update-row :deep(.video-update-btn.el-button--primary:focus-visible) {
  outline: 2px solid var(--color-brand);
  outline-offset: 2px;
}
.video-folder-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-action-group);
}
.video-folder-row :deep(.el-button + .el-button) {
  margin-left: 0;
}
.video-folder-row :deep(.video-folder-btn.el-button) {
  --el-button-hover-bg-color: #36495f;
  --el-button-hover-border-color: #4a5f7a;
  --el-button-hover-text-color: #e8eaed;
  --el-button-active-bg-color: #26374a;
  --el-button-active-border-color: #4a5f7a;
  --el-button-active-text-color: #e8eaed;
  border-color: var(--color-border);
  color: #e8eaed;
  background-color: var(--el-button-bg-color);
}
.video-folder-row :deep(.video-folder-btn.el-button:hover:not(.is-disabled)) {
  color: #e8eaed !important;
  background-color: #36495f !important;
  border-color: #4a5f7a !important;
}
.video-folder-row :deep(.video-folder-btn.el-button:active:not(.is-disabled)) {
  color: #e8eaed !important;
  background-color: #26374a !important;
  border-color: #4a5f7a !important;
}
.video-folder-row :deep(.video-folder-btn.el-button:focus:not(:focus-visible)) {
  outline: none;
  box-shadow: none;
  color: #e8eaed !important;
  background-color: var(--el-button-bg-color) !important;
  border-color: var(--color-border) !important;
}
.video-folder-row :deep(.video-folder-btn.el-button:focus-visible) {
  color: #e8eaed !important;
  background-color: var(--el-button-bg-color) !important;
  border-color: var(--color-border) !important;
  outline: 2px solid var(--color-brand);
  outline-offset: 2px;
}
.btn-icon {
  margin-right: var(--space-inline-tight);
  vertical-align: middle;
}
.error-details {
  margin-top: var(--space-action-group);
}
.error-details-toggle {
  display: inline-flex;
  align-items: center;
  gap: var(--space-inline-tight);
  padding: 0;
  border: none;
  background: transparent;
  color: var(--color-text-secondary);
  font: inherit;
  font-size: var(--font-size-12);
  cursor: pointer;
}
.error-details-toggle:hover {
  color: var(--color-text-primary);
}
.error-details-toggle-icon {
  font-size: var(--font-size-12);
}
.error-details-panel {
  margin-top: var(--space-action-group);
}
.detail-error-raw {
  margin: 0;
  word-break: break-word;
  color: var(--color-text-muted);
  font-size: var(--font-size-12);
}
.flash {
  margin: var(--space-action-group) 0 0;
}
.flash-ok {
  color: var(--color-success);
}
.flash-secondary {
  color: var(--color-text-secondary);
}
.flash-err {
  color: var(--color-danger);
}
</style>
