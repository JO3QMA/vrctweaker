<template>
  <div class="video-view">
    <h1 class="page-title">{{ t("video.title") }}</h1>

    <el-card
      class="section-card video-card video-history-card"
      shadow="never"
      data-testid="video-playback-history"
    >
      <template #header>
        <span>{{ t("video.historySection") }}</span>
      </template>
      <p class="history-retention-hint">
        {{ t("video.historyRetentionHint", { days: logRetentionDays }) }}
      </p>
      <VtAlert
        v-if="historyFetchError"
        variant="danger"
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
        size="small"
        class="history-table"
        data-testid="video-history-table"
      >
        <el-table-column :label="t('video.historyColTime')" min-width="140">
          <template #default="{ row }">
            {{ formatAttemptAt(row.attemptedAt) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('video.historyColUrl')" min-width="200">
          <template #default="{ row }">
            <div class="url-cell">
              <div class="url-text-wrap">
                <el-tooltip
                  :content="row.url"
                  placement="top"
                  :disabled="!row.url"
                >
                  <span class="url-text">{{ row.url }}</span>
                </el-tooltip>
              </div>
              <el-tooltip
                :content="t('video.historyCopyUrl')"
                placement="top"
                :trigger="['hover', 'focus']"
                :disabled="!row.url"
              >
                <VtButton
                  variant="primary"
                  link
                  size="small"
                  :aria-label="t('video.historyCopyUrl')"
                  :disabled="!row.url"
                  data-testid="video-history-copy-url"
                  @click="copyAttemptUrl(row.url)"
                >
                  <VtIcon size="default"><Link /></VtIcon>
                </VtButton>
              </el-tooltip>
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
      class="section-card video-card"
      shadow="never"
      data-testid="ytdlp-experimental-features"
    >
      <template #header>
        <span>{{ t("video.experimentalFeatures") }}</span>
      </template>

      <YtdlpReplaceSection />

      <CookieLinkageSection />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, onUnmounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { Link } from "@element-plus/icons-vue";
import CookieLinkageSection from "../components/CookieLinkageSection.vue";
import YtdlpReplaceSection from "../components/YtdlpReplaceSection.vue";
import VtAlert from "../components/VtAlert.vue";
import VtButton from "../components/VtButton.vue";
import VtIcon from "../components/VtIcon.vue";
import { App, type VideoPlaybackDTO } from "../wails/app";
import { getRuntime } from "../wails/runtime";
import { formatEncounteredAt } from "../utils/formatEncounteredAt";
import { copyDisplayName } from "../utils/vrcUserCacheDisplay";
import { appLocaleToBcp47 } from "../i18n";

const VIDEO_PLAYBACK_CHANGED_DEBOUNCE_MS = 400;
const HISTORY_COPY_FLASH_MS = 2000;

const { t, locale } = useI18n();

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

onBeforeUnmount(() => {
  viewGeneration++;
});

onMounted(() => {
  const gen = viewGeneration;
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
  margin-top: var(--space-block);
  margin-bottom: 0;
  width: 100%;
}
.history-retention-hint {
  margin: 0 0 var(--space-form-field);
  color: var(--color-text-secondary);
  font-size: var(--font-size-14);
}
.history-table {
  width: 100%;
}
.url-cell {
  display: flex;
  align-items: center;
  gap: var(--space-action-group);
  min-width: 0;
}
.url-text-wrap {
  flex: 1;
  min-width: 0;
  display: flex;
}
.url-text {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.flash {
  margin-top: var(--space-form-field);
}
.flash-ok {
  color: var(--color-success);
}
</style>
