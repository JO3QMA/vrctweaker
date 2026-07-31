<template>
  <div class="video-view">
    <h1 class="page-title">{{ t("video.title") }}</h1>

    <el-card
      class="video-card video-history-card"
      shadow="never"
      data-testid="video-playback-history"
    >
      <template #header>
        <span>{{ t("video.historySection") }}</span>
      </template>
      <p class="history-retention-hint">
        {{ t("video.historyRetentionHint", { days: logRetentionDays }) }}
      </p>
      <el-alert
        v-if="historyFetchError"
        type="error"
        :closable="false"
        show-icon
        class="block-hint"
        data-testid="video-history-fetch-error"
        :title="t('video.historyFetchError')"
      />
      <div v-else-if="historyLoading" class="muted">
        {{ t("common.loading") }}
      </div>
      <div
        v-else-if="historyRows.length === 0"
        class="muted"
        data-testid="video-history-empty"
      >
        {{ t("video.historyEmpty") }}
      </div>
      <el-table
        v-else
        :data="historyRows"
        stripe
        class="history-table"
        data-testid="video-history-table"
      >
        <el-table-column :label="t('video.historyColTime')" min-width="140">
          <template #default="{ row }">
            {{ formatAttemptAt(row.attemptedAt) }}
          </template>
        </el-table-column>
        <el-table-column
          :label="t('video.historyColUrl')"
          min-width="200"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            <div class="url-cell">
              <span class="url-text">{{ row.url }}</span>
              <el-button
                link
                type="primary"
                size="small"
                data-testid="video-history-copy-url"
                @click="copyAttemptUrl(row.url)"
              >
                {{ t("video.historyCopyUrl") }}
              </el-button>
            </div>
          </template>
        </el-table-column>
        <el-table-column :label="t('video.historyColOutcome')" min-width="90">
          <template #default="{ row }">
            {{ outcomeLabel(row.outcome) }}
          </template>
        </el-table-column>
        <el-table-column
          :label="t('video.historyColFailureReason')"
          min-width="180"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.failureReason || t("common.dash") }}
          </template>
        </el-table-column>
        <el-table-column
          :label="t('video.historyColWorld')"
          min-width="120"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.worldDisplayName || t("common.dash") }}
          </template>
        </el-table-column>
      </el-table>
      <p
        v-if="historyCopyFlash"
        class="flash flash-ok"
        data-testid="video-history-copy-ok"
      >
        {{ historyCopyFlash }}
      </p>
    </el-card>

    <el-card
      class="video-card"
      shadow="never"
      data-testid="ytdlp-experimental-features"
    >
      <template #header>
        <span>{{ t("video.experimentalFeatures") }}</span>
      </template>

      <section
        data-testid="ytdlp-replace-section"
        aria-labelledby="ytdlp-replace-heading"
      >
        <h2 id="ytdlp-replace-heading" class="video-block-title">
          {{ t("video.replaceLabel") }}
        </h2>

        <div v-if="loading" class="muted">{{ t("common.loading") }}</div>
        <template v-else>
          <!-- 1. 注意・エラー（置換ブロック直下・1箇所） -->
          <div class="video-alerts" data-testid="ytdlp-alert-area">
            <el-alert
              v-if="actionError"
              type="error"
              :closable="false"
              show-icon
              class="block-hint"
              data-testid="ytdlp-error-banner"
              :title="actionError"
            />
            <el-alert
              v-else-if="!status.supported"
              type="warning"
              :closable="false"
              show-icon
              :title="userFacingReason(status.unsupportedReason ?? '')"
            />
            <template v-else>
              <el-alert
                type="warning"
                :closable="false"
                show-icon
                class="block-hint"
                :title="t('video.alwaysWarn')"
              />
              <el-alert
                v-if="bannerError"
                type="error"
                :closable="false"
                show-icon
                class="block-hint"
                data-testid="ytdlp-error-banner"
                :title="bannerError"
              />
            </template>
          </div>

          <template v-if="status.supported">
            <!-- 2. 操作エリア -->
            <section class="video-ops" data-testid="ytdlp-ops">
              <div class="video-switch-row">
                <el-switch
                  v-model="maintainOn"
                  data-testid="ytdlp-maintain-switch"
                  :disabled="busy"
                  @change="onMaintainChange"
                />
                <span
                  class="switch-status"
                  data-testid="ytdlp-effective-inline"
                >
                  {{ t("video.statusPrefix") }}{{ effectiveStatusText }}
                </span>
              </div>

              <div class="video-actions" data-testid="ytdlp-action-grid">
                <el-button
                  data-testid="ytdlp-check-latest"
                  :loading="checkLoading"
                  :disabled="busy"
                  @click="checkLatest"
                >
                  {{ t("video.checkLatest") }}
                </el-button>
                <el-button
                  type="primary"
                  data-testid="ytdlp-update-cache"
                  :loading="updateLoading"
                  :disabled="busy"
                  @click="updateCache"
                >
                  {{ t("video.updateCache") }}
                </el-button>
                <el-button
                  data-testid="ytdlp-open-cache-folder"
                  :disabled="busy"
                  @click="openCacheFolder"
                >
                  <el-icon class="btn-icon"><FolderOpened /></el-icon>
                  {{ t("video.openCacheFolder") }}
                </el-button>
                <el-button
                  data-testid="ytdlp-open-tools-folder"
                  :disabled="busy"
                  @click="openToolsFolder"
                >
                  <el-icon class="btn-icon"><FolderOpened /></el-icon>
                  {{ t("video.openToolsFolder") }}
                </el-button>
              </div>

              <p
                v-if="flashOk"
                class="flash flash-ok"
                data-testid="ytdlp-flash-ok"
              >
                {{ flashOk }}
              </p>
            </section>

            <!-- 3. 詳細（バージョンのみ・初期は折りたたみ） -->
            <section class="video-status" data-testid="ytdlp-status">
              <button
                type="button"
                class="details-toggle"
                data-testid="ytdlp-details-toggle"
                :aria-expanded="detailsExpanded"
                :aria-controls="detailsPanelId"
                @click="detailsExpanded = !detailsExpanded"
              >
                <el-icon class="details-toggle-icon" aria-hidden="true">
                  <CaretBottom v-if="detailsExpanded" />
                  <CaretRight v-else />
                </el-icon>
                <span>{{ t("video.detailsToggle") }}</span>
              </button>
              <div
                v-show="detailsExpanded"
                :id="detailsPanelId"
                class="details-panel"
                role="region"
              >
                <dl class="video-dl">
                  <dt>{{ t("video.cacheVersion") }}</dt>
                  <dd data-testid="ytdlp-cache-version">
                    {{ status.cacheVersion || t("video.cacheMissing") }}
                  </dd>
                  <dt>{{ t("video.latest") }}</dt>
                  <dd data-testid="ytdlp-latest-version">
                    {{
                      status.latestVersion
                        ? status.latestVersion
                        : t("video.latestUnchecked")
                    }}
                  </dd>
                </dl>
              </div>
            </section>
          </template>
        </template>
      </section>

      <CookieLinkageSection />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, onUnmounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { ElMessageBox } from "element-plus";
import { CaretBottom, CaretRight, FolderOpened } from "@element-plus/icons-vue";
import CookieLinkageSection from "../components/CookieLinkageSection.vue";
import {
  App,
  type VideoPlaybackDTO,
  type YTDLPMaintainStatusDTO,
} from "../wails/app";
import { getRuntime } from "../wails/runtime";
import { formatEncounteredAt } from "../utils/formatEncounteredAt";
import { copyDisplayName } from "../utils/vrcUserCacheDisplay";
import { appLocaleToBcp47 } from "../i18n";
import { videoErrorI18nKey } from "./videoErrors";

const VIDEO_PLAYBACK_CHANGED_DEBOUNCE_MS = 400;
const HISTORY_COPY_FLASH_MS = 2000;

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
});

const status = ref<YTDLPMaintainStatusDTO>(emptyStatus());
const maintainOn = ref(false);
const loading = ref(true);
const checkLoading = ref(false);
const updateLoading = ref(false);
const busy = ref(false);
const flashOk = ref("");
const actionError = ref("");
/** 詳細アコーディオンは初期閉じ */
const detailsExpanded = ref(false);
const detailsPanelId = "ytdlp-details-panel";

const historyRows = ref<VideoPlaybackDTO[]>([]);
const historyLoading = ref(true);
const historyFetchError = ref(false);
const logRetentionDays = ref(30);
const historyCopyFlash = ref("");

let viewGeneration = 0;
function isViewStale(gen: number): boolean {
  return gen !== viewGeneration;
}

let historyChangedDebounceTimer: ReturnType<typeof setTimeout> | null = null;
let historyCopyFlashTimer: ReturnType<typeof setTimeout> | null = null;
let unsubscribeVideoPlaybackChanged: (() => void) | undefined;

function formatAttemptAt(iso: string): string {
  return formatEncounteredAt(iso, appLocaleToBcp47(String(locale.value)));
}

function outcomeLabel(outcome: string): string {
  if (outcome === "success") return t("video.historyOutcomeSuccess");
  if (outcome === "failure") return t("video.historyOutcomeFailure");
  return t("video.historyOutcomeOpen");
}

async function loadHistory(gen: number): Promise<void> {
  historyLoading.value = true;
  historyFetchError.value = false;
  try {
    if (isViewStale(gen)) return;
    const rows = await App.videoPlaybackHistory();
    if (isViewStale(gen)) return;
    historyRows.value = rows ?? [];
  } catch {
    if (isViewStale(gen)) return;
    historyFetchError.value = true;
    historyRows.value = [];
  } finally {
    if (!isViewStale(gen)) historyLoading.value = false;
  }
}

function scheduleHistoryRefresh(): void {
  if (historyChangedDebounceTimer !== null) {
    clearTimeout(historyChangedDebounceTimer);
  }
  historyChangedDebounceTimer = setTimeout(() => {
    historyChangedDebounceTimer = null;
    void loadHistory(viewGeneration);
  }, VIDEO_PLAYBACK_CHANGED_DEBOUNCE_MS);
}

async function copyAttemptUrl(url: string): Promise<void> {
  if (!url) return;
  await copyDisplayName(url);
  historyCopyFlash.value = t("video.historyCopyOk");
  if (historyCopyFlashTimer !== null) {
    clearTimeout(historyCopyFlashTimer);
  }
  historyCopyFlashTimer = setTimeout(() => {
    historyCopyFlash.value = "";
    historyCopyFlashTimer = null;
  }, HISTORY_COPY_FLASH_MS);
}

const effectiveStatusText = computed(() =>
  status.value.effectiveOfficial
    ? t("video.effectiveOfficial")
    : t("video.effectiveBundled"),
);

const bannerError = computed(() => {
  if (actionError.value) return actionError.value;
  if (status.value.pendingError) {
    return userFacingError(status.value.pendingError);
  }
  if (status.value.latestError) {
    return userFacingError(status.value.latestError);
  }
  return "";
});

function userFacingReason(code: string): string {
  if (!code) return t("video.unsupported");
  const key = `video.reason.${code}`;
  return te(key) ? t(key) : t("video.unsupported");
}

function userFacingError(raw: string): string {
  if (!raw) return "";
  // Stable app error codes (if backend ever returns them)
  if (te(`video.${raw}`)) return t(`video.${raw}`);
  return t(videoErrorI18nKey(raw));
}

function applyStatus(
  s: YTDLPMaintainStatusDTO,
  opts?: { syncSwitch?: boolean },
) {
  const prev = status.value;
  status.value = {
    ...s,
    latestVersion: s.latestVersion || prev.latestVersion,
    latestTag: s.latestTag || prev.latestTag,
    latestDownloadUrl: s.latestDownloadUrl || prev.latestDownloadUrl,
    cacheVersion: s.cacheVersion || prev.cacheVersion,
  };
  if (opts?.syncSwitch !== false) {
    maintainOn.value = !!s.maintainDesired;
  }
}

function clearFeedback() {
  flashOk.value = "";
  actionError.value = "";
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
    if (isViewStale(gen)) return false;
    applyStatus(await App.getYTDLPMaintainStatus());
    return !isViewStale(gen);
  } catch (e) {
    if (!isViewStale(gen)) {
      actionError.value = userFacingError(
        e instanceof Error ? e.message : String(e),
      );
    }
    return false;
  }
}

async function refresh() {
  const gen = viewGeneration;
  loading.value = true;
  clearFeedback();
  try {
    if (isViewStale(gen)) return;
    applyStatus(await App.getYTDLPMaintainStatus());
  } catch (e) {
    if (isViewStale(gen)) return;
    actionError.value = userFacingError(
      e instanceof Error ? e.message : String(e),
    );
  } finally {
    if (!isViewStale(gen)) loading.value = false;
  }
}

async function onMaintainChange(on: boolean) {
  const gen = viewGeneration;
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
        },
      );
      if (isViewStale(gen)) return;
      await App.acknowledgeYTDLPToolsReplaceRisk();
    }
    if (isViewStale(gen)) return;
    await App.setYTDLPToolsReplaceMaintain(desired);
    if (isViewStale(gen)) return;
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
    if (isViewStale(gen)) return;
    flashOk.value = desired
      ? t("video.flashEnabled")
      : t("video.flashDisabled");
  } catch (e) {
    if (isViewStale(gen)) return;
    maintainOn.value = status.value.maintainDesired;
    if (isMessageBoxDismiss(e)) {
      return;
    }
    actionError.value = userFacingError(
      e instanceof Error ? e.message : String(e),
    );
  } finally {
    if (!isViewStale(gen)) busy.value = false;
  }
}

async function checkLatest() {
  const gen = viewGeneration;
  checkLoading.value = true;
  busy.value = true;
  clearFeedback();
  try {
    if (isViewStale(gen)) return;
    applyStatus(await App.checkYTDLPLatestRelease(), { syncSwitch: false });
    if (isViewStale(gen)) return;
    if (status.value.latestError) {
      return;
    }
    flashOk.value = t("video.flashLatest", {
      version: status.value.latestVersion,
    });
  } catch (e) {
    if (isViewStale(gen)) return;
    actionError.value = userFacingError(
      e instanceof Error ? e.message : String(e),
    );
  } finally {
    if (!isViewStale(gen)) {
      checkLoading.value = false;
      busy.value = false;
    }
  }
}

async function updateCache() {
  const gen = viewGeneration;
  updateLoading.value = true;
  busy.value = true;
  clearFeedback();
  try {
    if (isViewStale(gen)) return;
    applyStatus(
      await App.updateOfficialYTDLPCache(
        status.value.latestDownloadUrl || "",
        status.value.latestTag || "",
      ),
      { syncSwitch: false },
    );
    if (isViewStale(gen)) return;
    if (status.value.pendingError || status.value.latestError) {
      return;
    }
    flashOk.value = t("video.flashUpdated", {
      version: status.value.cacheVersion,
    });
  } catch (e) {
    if (isViewStale(gen)) return;
    actionError.value = userFacingError(
      e instanceof Error ? e.message : String(e),
    );
  } finally {
    if (!isViewStale(gen)) {
      updateLoading.value = false;
      busy.value = false;
    }
  }
}

async function openCacheFolder() {
  const gen = viewGeneration;
  busy.value = true;
  clearFeedback();
  try {
    if (isViewStale(gen)) return;
    await App.openYTDLPCacheFolder();
  } catch (e) {
    if (isViewStale(gen)) return;
    actionError.value = userFacingError(
      e instanceof Error ? e.message : String(e),
    );
  } finally {
    if (!isViewStale(gen)) busy.value = false;
  }
}

async function openToolsFolder() {
  const gen = viewGeneration;
  busy.value = true;
  clearFeedback();
  try {
    if (isViewStale(gen)) return;
    await App.openYTDLPToolsFolder();
  } catch (e) {
    if (isViewStale(gen)) return;
    actionError.value = userFacingError(
      e instanceof Error ? e.message : String(e),
    );
  } finally {
    if (!isViewStale(gen)) busy.value = false;
  }
}

onBeforeUnmount(() => {
  viewGeneration++;
});

onMounted(() => {
  const gen = viewGeneration;
  void refresh();
  void loadHistory(gen);
  void App.getLogRetentionDays().then((days) => {
    if (!isViewStale(gen)) logRetentionDays.value = days;
  });
  const rt = getRuntime();
  const off = rt?.EventsOn?.("activity:video-playback-changed", () => {
    scheduleHistoryRefresh();
  });
  if (typeof off === "function") {
    unsubscribeVideoPlaybackChanged = off;
  }
});

onUnmounted(() => {
  if (historyChangedDebounceTimer !== null) {
    clearTimeout(historyChangedDebounceTimer);
    historyChangedDebounceTimer = null;
  }
  if (historyCopyFlashTimer !== null) {
    clearTimeout(historyCopyFlashTimer);
    historyCopyFlashTimer = null;
  }
  unsubscribeVideoPlaybackChanged?.();
});
</script>

<style scoped>
.video-view {
  width: 100%;
  box-sizing: border-box;
}
.video-card {
  margin-top: 1rem;
  width: 100%;
}
.history-retention-hint {
  margin: 0 0 0.75rem;
  color: var(--text-secondary);
  font-size: 0.875rem;
}
.history-table {
  width: 100%;
}
.url-cell {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
}
.url-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}
.video-block-title {
  margin: 0 0 0.75rem;
  font-size: 1rem;
  font-weight: 600;
}
.video-alerts {
  margin-bottom: 1rem;
}
.block-hint {
  margin-bottom: 0.75rem;
}
.video-ops {
  margin-bottom: 1.25rem;
}
.video-switch-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem 1rem;
  margin-bottom: 1rem;
}
.switch-status {
  color: var(--text-secondary);
}
.video-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
  width: 100%;
  max-width: 40rem;
}
.video-actions :deep(.el-button) {
  width: 100%;
  margin: 0;
  justify-content: center;
}
.btn-icon {
  margin-right: 0.25rem;
  vertical-align: middle;
}
.video-status {
  margin-top: 0.25rem;
  padding-top: 1rem;
  border-top: 1px solid var(--border);
}
.details-toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  font: inherit;
  cursor: pointer;
}
.details-toggle:hover {
  color: var(--text-primary);
}
.details-toggle-icon {
  font-size: 0.9rem;
}
.details-panel {
  margin-top: 0.75rem;
}
.video-dl {
  display: grid;
  grid-template-columns: minmax(10rem, 14rem) 1fr;
  gap: 0.4rem 1rem;
  margin: 0;
}
.video-dl dt {
  color: var(--text-secondary);
}
.video-dl dd {
  margin: 0;
}
.muted {
  color: var(--text-secondary);
}
.flash {
  margin-top: 0.75rem;
}
.flash-ok {
  color: var(--el-color-success);
}
</style>
