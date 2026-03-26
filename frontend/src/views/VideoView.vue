<template>
  <div class="video-view">
    <h1 class="page-title">動画</h1>

    <section class="settings-section ytdlp-section">
      <h2>yt-dlp のバージョン管理</h2>
      <p class="section-lead">
        VRChat が参照する公式
        <code>yt-dlp.exe</code>
        を GitHub の最新リリースで置き換えます。YouTube
        再生不具合の対策に使えます。
      </p>

      <div v-if="basicsLoading" class="muted">読み込み中…</div>
      <template v-else>
        <p v-if="!status.supported" class="banner-warn" role="status">
          {{ status.unsupportedReason || "この環境では利用できません。" }}
        </p>
        <template v-else>
          <dl class="ytdlp-dl">
            <dt>配置先</dt>
            <dd>
              <code class="path-code">{{ status.targetPath || "—" }}</code>
            </dd>
            <dt>現在のバージョン</dt>
            <dd>{{ status.localVersion || "（未取得・未配置）" }}</dd>
            <dt>GitHub 最新</dt>
            <dd>
              {{
                status.latestVersion
                  ? status.latestVersion
                  : status.latestError
                    ? "—"
                    : "未確認（下のボタンで取得）"
              }}
            </dd>
          </dl>
          <p v-if="status.latestError" class="banner-error" role="alert">
            最新版の取得に失敗しました: {{ status.latestError }}
          </p>
          <div class="ytdlp-actions">
            <button
              type="button"
              class="btn-refresh"
              data-testid="ytdlp-check-latest"
              :disabled="latestCheckLoading || applyLoading"
              @click="checkLatest"
            >
              {{ latestCheckLoading ? "確認中…" : "GitHub で最新版を確認" }}
            </button>
            <button
              type="button"
              class="btn-apply"
              data-testid="ytdlp-apply"
              :disabled="
                applyLoading ||
                latestCheckLoading ||
                !status.latestDownloadUrl ||
                !!status.latestError
              "
              @click="applyLatest"
            >
              {{
                applyLoading
                  ? "適用中…"
                  : "最新版をダウンロードして VRChat に適用"
              }}
            </button>
          </div>
          <p v-if="applyFlash" class="apply-flash" :class="applyFlashClass">
            {{ applyFlash }}
          </p>
        </template>
      </template>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { App, type YTDLPUpdateStatusDTO } from "../wails/app";

const status = ref<YTDLPUpdateStatusDTO>({
  supported: false,
  targetPath: "",
  localVersion: "",
  latestVersion: "",
  latestTag: "",
  latestDownloadUrl: "",
  latestError: "",
});
const basicsLoading = ref(true);
const latestCheckLoading = ref(false);
const applyLoading = ref(false);
const applyFlash = ref("");
const applyFlashClass = ref("");

async function loadBasics() {
  basicsLoading.value = true;
  try {
    status.value = await App.getYTDLPBasics();
  } finally {
    basicsLoading.value = false;
  }
}

async function checkLatest() {
  latestCheckLoading.value = true;
  applyFlash.value = "";
  try {
    status.value = await App.getYTDLPUpdateStatus();
  } finally {
    latestCheckLoading.value = false;
  }
}

async function applyLatest() {
  if (!status.value.latestDownloadUrl) return;
  applyLoading.value = true;
  applyFlash.value = "";
  try {
    const r = await App.applyYTDLP(
      status.value.latestDownloadUrl,
      status.value.latestTag,
    );
    if (r.ok) {
      applyFlashClass.value = "apply-flash--ok";
      applyFlash.value = r.message || "適用しました。";
      status.value = await App.getYTDLPBasics();
    } else {
      applyFlashClass.value = "apply-flash--err";
      applyFlash.value = r.error || "適用に失敗しました。";
    }
  } finally {
    applyLoading.value = false;
  }
}

onMounted(() => {
  void loadBasics();
});
</script>

<style scoped>
.video-view {
  max-width: 720px;
  margin: 0 auto;
}

.page-title {
  margin: 0 0 1rem;
  font-size: 1.5rem;
}

.settings-section {
  padding: 1rem 1.25rem;
  background: var(--bg-secondary);
  border-radius: var(--radius);
  border: 1px solid var(--border);
}

.settings-section h2 {
  margin: 0 0 0.5rem;
  font-size: 1.1rem;
}

.section-lead {
  margin: 0 0 1rem;
  color: var(--text-secondary);
  font-size: 0.9rem;
  line-height: 1.5;
}

.section-lead code {
  font-size: 0.85em;
  padding: 0.1em 0.35em;
  background: var(--bg-tertiary);
  border-radius: 4px;
}

.muted {
  color: var(--text-secondary);
}

.banner-warn {
  margin: 0 0 1rem;
  padding: 0.75rem 1rem;
  background: var(--bg-tertiary);
  border-radius: var(--radius);
  border-left: 3px solid var(--accent);
  color: var(--text-secondary);
}

.banner-error {
  margin: 0.75rem 0;
  padding: 0.75rem 1rem;
  background: rgba(229, 115, 115, 0.12);
  border-radius: var(--radius);
  color: var(--danger);
}

.ytdlp-dl {
  display: grid;
  grid-template-columns: 10rem 1fr;
  gap: 0.35rem 1rem;
  margin: 0 0 1rem;
  font-size: 0.9rem;
}

.ytdlp-dl dt {
  margin: 0;
  color: var(--text-secondary);
}

.ytdlp-dl dd {
  margin: 0;
  word-break: break-all;
}

.path-code {
  font-size: 0.8rem;
}

.ytdlp-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.btn-refresh {
  padding: 0.45rem 0.9rem;
  border-radius: var(--radius);
  border: 1px solid var(--border);
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.btn-refresh:hover:not(:disabled) {
  background: var(--bg-primary);
}

.btn-refresh:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.btn-apply {
  padding: 0.45rem 0.9rem;
  border-radius: var(--radius);
  border: none;
  background: var(--accent);
  color: white;
}

.btn-apply:hover:not(:disabled) {
  background: var(--accent-hover);
}

.btn-apply:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.apply-flash {
  margin: 0.75rem 0 0;
  font-size: 0.9rem;
}

.apply-flash--ok {
  color: var(--success);
}

.apply-flash--err {
  color: var(--danger);
}
</style>
