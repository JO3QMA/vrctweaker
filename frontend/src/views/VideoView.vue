<template>
  <div class="video-view">
    <h1 class="page-title">{{ t("video.title") }}</h1>

    <el-card class="video-card" shadow="never">
      <template #header>
        <span>{{ t("video.maintainSection") }}</span>
      </template>

      <div v-if="loading" class="muted">{{ t("common.loading") }}</div>
      <template v-else>
        <el-alert
          v-if="!status.supported"
          type="warning"
          :closable="false"
          show-icon
          :title="status.unsupportedReason || t('video.unsupported')"
        />
        <template v-else>
          <el-alert
            type="warning"
            :closable="false"
            show-icon
            class="block-hint"
            :title="t('video.alwaysWarn')"
          />

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
            <dd>
              {{
                status.effectiveOfficial
                  ? t("video.effectiveOfficial")
                  : t("video.effectiveBundled")
              }}
            </dd>
            <dt>{{ t("video.cacheVersion") }}</dt>
            <dd>{{ status.cacheVersion || t("video.cacheMissing") }}</dd>
            <dt>{{ t("video.cachePath") }}</dt>
            <dd>
              <code class="path-code">{{ status.cachePath || "—" }}</code>
            </dd>
            <dt>{{ t("video.toolsPath") }}</dt>
            <dd>
              <code class="path-code">{{ status.toolsPath || "—" }}</code>
            </dd>
            <dt>{{ t("video.latest") }}</dt>
            <dd>
              {{
                status.latestVersion
                  ? status.latestVersion
                  : status.latestError
                    ? "—"
                    : t("video.latestUnchecked")
              }}
            </dd>
          </dl>

          <el-alert
            v-if="status.pendingError"
            type="error"
            :closable="false"
            show-icon
            class="block-hint"
            :title="t('video.pendingError', { msg: status.pendingError })"
          />
          <el-alert
            v-if="status.latestError"
            type="error"
            :closable="false"
            show-icon
            class="block-hint"
            :title="t('video.latestError', { msg: status.latestError })"
          />

          <div class="video-actions">
            <el-switch
              v-model="maintainOn"
              data-testid="ytdlp-maintain-switch"
              :disabled="busy"
              :active-text="t('video.maintainOn')"
              :inactive-text="t('video.maintainOff')"
              @change="onMaintainChange"
            />
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
          </div>

          <p v-if="flash" class="flash" :class="flashClass">{{ flash }}</p>
        </template>
      </template>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { ElMessageBox } from "element-plus";
import { App, type YTDLPMaintainStatusDTO } from "../wails/app";

const { t } = useI18n();

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
const flash = ref("");
const flashClass = ref("");

function applyStatus(s: YTDLPMaintainStatusDTO) {
  status.value = s;
  maintainOn.value = !!s.maintainDesired;
}

async function refresh() {
  loading.value = true;
  try {
    applyStatus(await App.getYTDLPMaintainStatus());
  } finally {
    loading.value = false;
  }
}

async function onMaintainChange(on: boolean | string | number) {
  const desired = on === true;
  busy.value = true;
  flash.value = "";
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
    flash.value = desired ? t("video.flashEnabled") : t("video.flashDisabled");
    flashClass.value = "flash-ok";
  } catch (e) {
    maintainOn.value = status.value.maintainDesired;
    if (e === "cancel" || (e && typeof e === "object" && "action" in e)) {
      return;
    }
    flash.value = e instanceof Error ? e.message : String(e);
    flashClass.value = "flash-err";
  } finally {
    busy.value = false;
  }
}

async function checkLatest() {
  checkLoading.value = true;
  busy.value = true;
  flash.value = "";
  try {
    applyStatus(await App.checkYTDLPLatestRelease());
    if (status.value.latestError) {
      flash.value = status.value.latestError;
      flashClass.value = "flash-err";
    } else {
      flash.value = t("video.flashLatest", {
        version: status.value.latestVersion,
      });
      flashClass.value = "flash-ok";
    }
  } catch (e) {
    flash.value = e instanceof Error ? e.message : String(e);
    flashClass.value = "flash-err";
  } finally {
    checkLoading.value = false;
    busy.value = false;
  }
}

async function updateCache() {
  updateLoading.value = true;
  busy.value = true;
  flash.value = "";
  try {
    applyStatus(
      await App.updateOfficialYTDLPCache(
        status.value.latestDownloadUrl || "",
        status.value.latestTag || "",
      ),
    );
    flash.value = t("video.flashUpdated", {
      version: status.value.cacheVersion,
    });
    flashClass.value = "flash-ok";
  } catch (e) {
    flash.value = e instanceof Error ? e.message : String(e);
    flashClass.value = "flash-err";
  } finally {
    updateLoading.value = false;
    busy.value = false;
  }
}

onMounted(() => {
  void refresh();
});
</script>

<style scoped>
.video-view {
  max-width: 52rem;
}
.video-card {
  margin-top: 1rem;
}
.block-hint {
  margin-bottom: 1rem;
}
.video-dl {
  display: grid;
  grid-template-columns: 10rem 1fr;
  gap: 0.4rem 1rem;
  margin: 0 0 1rem;
}
.video-dl dt {
  color: var(--text-secondary);
}
.video-dl dd {
  margin: 0;
  word-break: break-all;
}
.path-code {
  font-size: 0.85em;
}
.video-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
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
.flash-err {
  color: var(--el-color-danger);
}
</style>
