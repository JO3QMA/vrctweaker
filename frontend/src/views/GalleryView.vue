<template>
  <div class="gallery-view">
    <h1 class="page-title">
      ギャラリー
    </h1>

    <!-- フィルタ（最小: worldId） -->
    <div class="filters">
      <input
        v-model="filterWorldId"
        type="text"
        placeholder="World ID で検索"
        class="filter-input"
      >
      <button
        class="btn-refresh"
        @click="load"
      >
        更新
      </button>
    </div>

    <!-- グリッド一覧 -->
    <div
      v-if="loading"
      class="loading"
    >
      読み込み中…
    </div>
    <div
      v-else-if="list.length === 0"
      class="empty"
    >
      スクリーンショットがありません。設定からフォルダをスキャンしてください。
    </div>
    <div
      v-else
      class="grid"
    >
      <div
        v-for="item in list"
        :key="item.id"
        class="grid-item"
        :class="{ selected: selected?.id === item.id }"
        @click="select(item)"
      >
        <div class="thumbnail-wrap">
          <img
            :src="thumbnailSrc(item)"
            :alt="item.filePath"
            class="thumbnail"
            @error="onThumbnailError"
          >
        </div>
      </div>
    </div>

    <!-- 詳細プレビュー -->
    <div
      v-if="selected"
      class="detail-panel"
    >
      <h3>詳細</h3>
      <dl class="detail-list">
        <dt>撮影日時</dt>
        <dd>{{ formatTakenAt(selected.takenAt) }}</dd>
        <dt>ワールド名</dt>
        <dd>{{ selected.worldName || '—' }}</dd>
        <dt>World ID</dt>
        <dd>{{ selected.worldId || '—' }}</dd>
        <dt>ファイルパス</dt>
        <dd class="file-path">
          {{ selected.filePath }}
        </dd>
      </dl>
      <p
        v-if="joinError"
        class="join-error"
      >
        {{ joinError }}
      </p>
      <button
        class="btn-join"
        :disabled="!selected.worldId || selected.worldId.trim() === ''"
        :title="joinButtonTitle"
        @click="onJoin"
      >
        このワールドへJoin
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { App, type ScreenshotDTO, type ScreenshotSearchDTO } from '../wails/app'

const list = ref<ScreenshotDTO[]>([])
const selected = ref<ScreenshotDTO | null>(null)
const loading = ref(false)
const filterWorldId = ref('')

const joinButtonTitle = computed(() => {
  if (!selected.value?.worldId || selected.value.worldId.trim() === '') {
    return 'World ID がありません'
  }
  return 'このワールドへJoin'
})

function pathToFileUrl(path: string): string {
  const normalized = path.replace(/\\/g, '/')
  if (normalized.match(/^[a-zA-Z]:/)) {
    return 'file:///' + normalized
  }
  if (normalized.startsWith('/')) {
    return 'file://' + normalized
  }
  return 'file:///' + normalized
}

function thumbnailSrc(item: ScreenshotDTO): string {
  return pathToFileUrl(item.filePath)
}

function onThumbnailError(e: Event): void {
  const img = e.target as HTMLImageElement
  img.src = 'data:image/svg+xml,' + encodeURIComponent(
    '<svg xmlns="http://www.w3.org/2000/svg" width="120" height="90" viewBox="0 0 120 90"><rect fill="#333" width="120" height="90"/><text x="50%" y="50%" dominant-baseline="middle" text-anchor="middle" fill="#666" font-size="12">画像なし</text></svg>',
  )
}

function formatTakenAt(takenAt?: string): string {
  if (!takenAt) return '—'
  try {
    const d = new Date(takenAt)
    return d.toLocaleString('ja-JP')
  } catch {
    return takenAt
  }
}

async function load(): Promise<void> {
  loading.value = true
  try {
    const wid = filterWorldId.value.trim()
    if (wid) {
      const filter: ScreenshotSearchDTO = { worldId: wid }
      list.value = await App.searchScreenshots(filter)
    } else {
      list.value = await App.screenshots('')
    }
    if (selected.value && !list.value.find((s) => s.id === selected.value?.id)) {
      selected.value = null
    }
  } finally {
    loading.value = false
  }
}

function select(item: ScreenshotDTO): void {
  selected.value = item
  joinError.value = null
}

const joinError = ref<string | null>(null)

async function onJoin(): Promise<void> {
  if (!selected.value?.worldId || selected.value.worldId.trim() === '') return
  joinError.value = null
  try {
    await App.joinWorldFromScreenshot(selected.value.id)
  } catch (err) {
    joinError.value = err instanceof Error ? err.message : String(err)
  }
}

onMounted(load)
</script>

<style scoped>
.gallery-view {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.page-title {
  margin: 0;
  font-size: 1.5rem;
}

.filters {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.filter-input {
  width: 200px;
  padding: 0.5rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
}

.btn-refresh {
  padding: 0.5rem 1rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
  cursor: pointer;
}

.btn-refresh:hover {
  background: var(--accent);
  color: white;
  border-color: var(--accent);
}

.loading,
.empty {
  padding: 2rem;
  text-align: center;
  color: var(--text-secondary);
}

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 0.75rem;
}

.grid-item {
  aspect-ratio: 4/3;
  border-radius: var(--radius);
  overflow: hidden;
  cursor: pointer;
  border: 2px solid transparent;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.grid-item:hover,
.grid-item.selected {
  border-color: var(--accent);
  box-shadow: 0 0 0 1px var(--accent);
}

.thumbnail-wrap {
  width: 100%;
  height: 100%;
  background: var(--bg-tertiary);
}

.thumbnail {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.detail-panel {
  padding: 1rem;
  background: var(--bg-secondary);
  border-radius: var(--radius);
  border: 1px solid var(--border);
}

.detail-panel h3 {
  margin: 0 0 0.75rem;
  font-size: 1rem;
}

.detail-list {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.25rem 1rem;
  font-size: 0.9rem;
  margin: 0 0 1rem;
}

.detail-list dt {
  color: var(--text-secondary);
}

.detail-list dd {
  margin: 0;
}

.detail-list .file-path {
  word-break: break-all;
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.btn-join {
  padding: 0.5rem 1rem;
  background: var(--accent);
  color: white;
  border: none;
  border-radius: var(--radius);
  cursor: pointer;
}

.btn-join:hover:not(:disabled) {
  background: var(--accent-hover);
}

.btn-join:disabled {
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  cursor: not-allowed;
}

.join-error {
  margin: 0 0 0.5rem;
  font-size: 0.85rem;
  color: var(--accent);
}
</style>
