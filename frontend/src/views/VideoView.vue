<template>
  <div class="video-view">
    <h1 class="page-title">{{ t("video.title") }}</h1>

    <el-card class="video-card" shadow="never">
      <template #header>
        <span>{{ t("video.maintainSection") }}</span>
      </template>

      <div v-if="loading" class="muted">{{ t("common.loading") }}</div>
      <template v-else>
        <!-- 1. 注意・エラー（タイトル直下・1箇所） -->
        <div class="video-alerts" data-testid="ytdlp-alert-area">
          <el-alert
            v-if="!status.supported"
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
              <span class="switch-label">{{ t("video.replaceLabel") }}</span>
              <el-switch
                v-model="maintainOn"
                data-testid="ytdlp-maintain-switch"
                :disabled="busy"
                :active-text="t('video.switchOn')"
                :inactive-text="t('video.switchOff')"
                @change="onMaintainChange"
              />
              <span class="switch-status" data-testid="ytdlp-effective-inline">
                {{ t("video.statusPrefix") }}{{ effectiveStatusText }}
              </span>
            </div>

            <div class="video-actions">
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

          <!-- 3. ステータス・詳細（パスは出さない） -->
          <section class="video-status" data-testid="ytdlp-status">
            <h3 class="status-heading">{{ t("video.statusHeading") }}</h3>
            <dl class="video-dl">
              <dt>{{ t("video.desired") }}</dt>
              <dd>
                {{
                  status.maintainDesired
                    ? t("video.desiredOn")
                    : t("video.desiredOff")
                }}
              </dd>
              <dt>{{ t("video.effective") }}</dt>
              <dd>{{ effectiveStatusText }}</dd>
              <dt>{{ t("video.cacheVersion") }}</dt>
              <dd>{{ status.cacheVersion || t("video.cacheMissing") }}</dd>
              <dt>{{ t("video.latest") }}</dt>
              <dd>
                {{
                  status.latestVersion
                    ? status.latestVersion
                    : t("video.latestUnchecked")
                }}
              </dd>
            </dl>
          </section>
        </template>
      </template>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { ElMessageBox } from "element-plus";
import { App, type YTDLPMaintainStatusDTO } from "../wails/app";
import { videoErrorI18nKey } from "./videoErrors";

const { t, te } = useI18n();

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

function applyStatus(s: YTDLPMaintainStatusDTO) {
  status.value = s;
  maintainOn.value = !!s.maintainDesired;
}

function clearFeedback() {
  flashOk.value = "";
  actionError.value = "";
}

async function refresh() {
  loading.value = true;
  clearFeedback();
  try {
    applyStatus(await App.getYTDLPMaintainStatus());
  } catch (e) {
    actionError.value = userFacingError(
      e instanceof Error ? e.message : String(e),
    );
  } finally {
    loading.value = false;
  }
}

async function onMaintainChange(on: boolean | string | number) {
  const desired = on === true;
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
      await App.acknowledgeYTDLPToolsReplaceRisk();
    }
    await App.setYTDLPToolsReplaceMaintain(desired);
    applyStatus(await App.getYTDLPMaintainStatus());
    flashOk.value = desired
      ? t("video.flashEnabled")
      : t("video.flashDisabled");
  } catch (e) {
    maintainOn.value = status.value.maintainDesired;
    if (e === "cancel") {
      return;
    }
    actionError.value = userFacingError(
      e instanceof Error ? e.message : String(e),
    );
  } finally {
    busy.value = false;
  }
}

async function checkLatest() {
  checkLoading.value = true;
  busy.value = true;
  clearFeedback();
  try {
    applyStatus(await App.checkYTDLPLatestRelease());
    if (status.value.latestError) {
      // bannerError reads latestError — do not also set actionError (no duplicate)
      return;
    }
    flashOk.value = t("video.flashLatest", {
      version: status.value.latestVersion,
    });
  } catch (e) {
    actionError.value = userFacingError(
      e instanceof Error ? e.message : String(e),
    );
  } finally {
    checkLoading.value = false;
    busy.value = false;
  }
}

async function updateCache() {
  updateLoading.value = true;
  busy.value = true;
  clearFeedback();
  try {
    applyStatus(
      await App.updateOfficialYTDLPCache(
        status.value.latestDownloadUrl || "",
        status.value.latestTag || "",
      ),
    );
    if (status.value.pendingError || status.value.latestError) {
      return;
    }
    flashOk.value = t("video.flashUpdated", {
      version: status.value.cacheVersion,
    });
  } catch (e) {
    actionError.value = userFacingError(
      e instanceof Error ? e.message : String(e),
    );
  } finally {
    updateLoading.value = false;
    busy.value = false;
  }
}

async function openCacheFolder() {
  clearFeedback();
  try {
    await App.openYTDLPCacheFolder();
  } catch (e) {
    actionError.value = userFacingError(
      e instanceof Error ? e.message : String(e),
    );
  }
}

async function openToolsFolder() {
  clearFeedback();
  try {
    await App.openYTDLPToolsFolder();
  } catch (e) {
    actionError.value = userFacingError(
      e instanceof Error ? e.message : String(e),
    );
  }
}

onMounted(() => {
  void refresh();
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
.video-alerts {
  margin-bottom: 1rem;
}
.block-hint {
  margin-bottom: 0.75rem;
}
.video-ops {
  margin-bottom: 1.5rem;
  padding-bottom: 1.25rem;
  border-bottom: 1px solid var(--border);
}
.video-switch-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem 1rem;
  margin-bottom: 1rem;
}
.switch-label {
  font-weight: 600;
}
.switch-status {
  color: var(--text-secondary);
}
.video-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
}
.btn-icon {
  margin-right: 0.25rem;
  vertical-align: middle;
}
.status-heading {
  margin: 0 0 0.75rem;
  font-size: 1rem;
  font-weight: 600;
}
.video-dl {
  display: grid;
  grid-template-columns: 12rem 1fr;
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
