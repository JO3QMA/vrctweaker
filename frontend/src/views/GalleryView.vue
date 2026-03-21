<template>
  <div class="gallery-view">
    <h1 class="page-title">ギャラリー</h1>

    <!-- フィルタ（最小: worldId） -->
    <div class="filters">
      <input
        v-model="filterWorldId"
        data-testid="gallery-world-filter"
        type="search"
        placeholder="World ID で検索（入力で自動検索 / Enter）"
        class="filter-input"
        @keyup.enter="onFilterEnter"
      />
      <button
        type="button"
        class="btn-refresh"
        :disabled="loading || scanning"
        @click="load"
      >
        更新
      </button>
      <button
        type="button"
        data-testid="gallery-scan-folder"
        class="btn-scan"
        :disabled="loading || scanning"
        @click="scanFolder"
      >
        {{ scanning ? "スキャン中…" : "Scan Folder" }}
      </button>
    </div>

    <p v-if="loadError" class="banner-error" role="alert">
      {{ loadError }}
    </p>
    <p v-if="scanError" class="banner-error banner-warn" role="status">
      {{ scanError }}
    </p>

    <div class="gallery-body">
      <!-- グリッド一覧（この領域のみ縦スクロール） -->
      <div class="grid-section">
        <div v-if="scanning" class="loading">フォルダをスキャンしています…</div>
        <div v-else-if="loading" class="loading">読み込み中…</div>
        <div v-else-if="list.length === 0" class="empty">
          スクリーンショットがありません。Scan Folder
          か設定の出力フォルダを確認してください。
        </div>
        <div
          v-else
          ref="gridScrollRef"
          data-testid="gallery-grid-scroll"
          class="grid-scroll"
          @scroll.passive="onGridScroll"
        >
          <div class="grid-virtual-spacer" :style="spacerStyle">
            <div
              v-for="vr in virtualRows"
              :key="virtualRowDomKey(vr.index)"
              class="grid-virtual-row"
              :style="virtualRowStyle(vr)"
            >
              <template v-if="isGridRow(vr.index)">
                <div class="grid-row-inner" :style="gridRowInnerStyle">
                  <div
                    v-for="item in gridRowItems(vr.index)"
                    :key="item.id"
                    class="grid-item"
                    :class="{ selected: selected?.id === item.id }"
                    :style="gridItemStyle"
                    @click="select(item)"
                  >
                    <div class="thumbnail-wrap">
                      <img
                        :src="thumbnailSrc(item)"
                        :alt="fileNameFromPath(item.filePath)"
                        class="thumbnail"
                        @error="onThumbnailError"
                      />
                    </div>
                  </div>
                </div>
              </template>
              <button
                v-else-if="galleryHeaderAt(vr.index)"
                type="button"
                class="gallery-group-header"
                :class="galleryHeaderIndentClass(galleryHeaderAt(vr.index)!)"
                data-testid="gallery-group-header"
                :data-collapse-key="galleryHeaderAt(vr.index)!.collapseKey"
                :aria-expanded="galleryHeaderAt(vr.index)!.expanded"
                @click="
                  toggleGalleryCollapse(galleryHeaderAt(vr.index)!.collapseKey)
                "
              >
                <span class="gallery-group-chevron" aria-hidden="true">{{
                  galleryHeaderAt(vr.index)!.expanded ? "▼" : "▶"
                }}</span>
                <span class="gallery-group-label">{{
                  galleryHeaderAt(vr.index)!.label
                }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 詳細プレビュー（スクロールに追従しない） -->
      <aside v-if="selected" class="detail-panel">
        <h3>詳細</h3>
        <dl class="detail-list">
          <dt>ファイル名</dt>
          <dd>{{ fileNameFromPath(selected.filePath) }}</dd>
          <dt>ファイルサイズ</dt>
          <dd>{{ formatFileSize(selected.fileSizeBytes) }}</dd>
          <dt>撮影日時</dt>
          <dd>{{ formatTakenAt(selected.takenAt) }}</dd>
          <dt>ワールド名</dt>
          <dd>{{ selected.worldName || "—" }}</dd>
          <dt>World ID</dt>
          <dd>{{ selected.worldId || "—" }}</dd>
          <dt>ファイルパス</dt>
          <dd class="file-path">
            {{ selected.filePath }}
          </dd>
        </dl>
        <p v-if="joinError" class="join-error">
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
      </aside>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useVirtualizer, type VirtualItem } from "@tanstack/vue-virtual";
import {
  ref,
  onMounted,
  onBeforeUnmount,
  computed,
  watch,
  watchEffect,
  nextTick,
} from "vue";
import {
  App,
  type ScreenshotDTO,
  type ScreenshotSearchDTO,
} from "../wails/app";
import {
  buildGalleryVirtualRows,
  galleryRowHeight,
  type GalleryVirtualRow,
} from "./galleryDateGroups";
import { pruneThumbnailUrlMap } from "./galleryThumbnailCache";

const FILTER_DEBOUNCE_MS = 400;
/** Debounced prune after scroll so off-screen Data URLs are released without thrashing. */
const THUMBNAIL_PRUNE_SCROLL_DEBOUNCE_MS = 150;
const THUMBNAIL_FETCH_CONCURRENCY = 4;
const GRID_GAP_PX = 12;
const MIN_CELL_WIDTH = 140;

const missingThumbDataUrl =
  "data:image/svg+xml," +
  encodeURIComponent(
    '<svg xmlns="http://www.w3.org/2000/svg" width="120" height="90" viewBox="0 0 120 90"><rect fill="#333" width="120" height="90"/><text x="50%" y="50%" dominant-baseline="middle" text-anchor="middle" fill="#666" font-size="12">画像なし</text></svg>',
  );

/** Placeholder while backend thumbnail is loading (avoids file:// in WebView). */
const transparentPixelDataUrl =
  "data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7";

const list = ref<ScreenshotDTO[]>([]);
const selected = ref<ScreenshotDTO | null>(null);
const loading = ref(false);
const scanning = ref(false);
const loadError = ref<string | null>(null);
const scanError = ref<string | null>(null);
const filterWorldId = ref("");
const thumbnailUrls = ref<Record<string, string>>({});

/** Collapsed section keys (`y:…`, `m:…`, `d:…`). */
const collapsed = ref(new Set<string>());

const gridScrollRef = ref<HTMLElement | null>(null);
/** Content width inside .grid-scroll (from ResizeObserver). */
const gridInnerWidth = ref(0);
/** Bumps when the grid scrolls so virtual rows / thumb prefetch stay in sync. */
const scrollSync = ref(0);

let filterDebounceTimer: ReturnType<typeof setTimeout> | null = null;
let thumbnailPruneScrollTimer: ReturnType<typeof setTimeout> | null = null;
let thumbnailFetchGeneration = 0;

const columnCount = computed(() => {
  const w = gridInnerWidth.value;
  if (w <= 0) {
    return 1;
  }
  return Math.max(
    1,
    Math.floor((w + GRID_GAP_PX) / (MIN_CELL_WIDTH + GRID_GAP_PX)),
  );
});

const cellWidthPx = computed(() => {
  const cols = columnCount.value;
  const w = gridInnerWidth.value;
  if (cols <= 0 || w <= 0) {
    return MIN_CELL_WIDTH;
  }
  return (w - GRID_GAP_PX * (cols - 1)) / cols;
});

const cellHeightPx = computed(() => (cellWidthPx.value * 3) / 4);

const rowHeightPx = computed(() => cellHeightPx.value + GRID_GAP_PX);

const flatGalleryRows = computed(() =>
  buildGalleryVirtualRows(list.value, columnCount.value, collapsed.value),
);

const rowVirtualizer = useVirtualizer(
  computed(() => ({
    count: flatGalleryRows.value.length,
    getScrollElement: () => gridScrollRef.value,
    estimateSize: (index: number) => {
      const row = flatGalleryRows.value[index];
      if (!row) {
        return rowHeightPx.value;
      }
      return galleryRowHeight(row, rowHeightPx.value);
    },
    overscan: 3,
  })),
);

const virtualRows = computed(() => {
  scrollSync.value;
  return rowVirtualizer.value.getVirtualItems();
});

const totalVirtualHeight = computed(() => {
  scrollSync.value;
  return rowVirtualizer.value.getTotalSize();
});

const spacerStyle = computed(() => ({
  height: `${totalVirtualHeight.value}px`,
  position: "relative" as const,
  width: "100%",
}));

const gridRowInnerStyle = computed(() => ({
  display: "flex" as const,
  flexDirection: "row" as const,
  gap: `${GRID_GAP_PX}px`,
  width: "100%",
}));

const gridItemStyle = computed(() => ({
  width: `${cellWidthPx.value}px`,
  height: `${cellHeightPx.value}px`,
  flexShrink: 0,
}));

function virtualRowStyle(vr: VirtualItem) {
  return {
    position: "absolute" as const,
    top: 0,
    left: 0,
    width: "100%",
    height: `${vr.size}px`,
    transform: `translateY(${vr.start}px)`,
  };
}

function virtualRowDomKey(index: number): string {
  return flatGalleryRows.value[index]?.rowKey ?? `row-${index}`;
}

function galleryRowAt(index: number): GalleryVirtualRow | undefined {
  return flatGalleryRows.value[index];
}

function isGridRow(index: number): boolean {
  return galleryRowAt(index)?.type === "grid";
}

function gridRowItems(index: number): ScreenshotDTO[] {
  const row = galleryRowAt(index);
  return row?.type === "grid" ? row.items : [];
}

function galleryHeaderAt(
  index: number,
):
  | Extract<
      GalleryVirtualRow,
      { type: "yearHeader" | "monthHeader" | "dayHeader" }
    >
  | undefined {
  const row = galleryRowAt(index);
  if (
    row?.type === "yearHeader" ||
    row?.type === "monthHeader" ||
    row?.type === "dayHeader"
  ) {
    return row;
  }
  return undefined;
}

function galleryHeaderIndentClass(
  row: NonNullable<ReturnType<typeof galleryHeaderAt>>,
): string {
  if (row.type === "yearHeader") {
    return "gallery-group-h-year";
  }
  if (row.type === "monthHeader") {
    return "gallery-group-h-month";
  }
  return "gallery-group-h-day";
}

function toggleGalleryCollapse(key: string): void {
  const next = new Set(collapsed.value);
  if (next.has(key)) {
    next.delete(key);
  } else {
    next.add(key);
  }
  collapsed.value = next;
  void nextTick(() => {
    scrollSync.value++;
    rowVirtualizer.value.measure();
    thumbnailFetchGeneration++;
    pruneThumbnailsToRetained();
    void syncThumbnailsForVisible();
  });
}

watchEffect((onCleanup) => {
  const el = gridScrollRef.value;
  if (!el || list.value.length === 0) {
    return;
  }
  const ro = new ResizeObserver((entries) => {
    const w = entries[0]?.contentRect.width ?? 0;
    gridInnerWidth.value = Math.floor(w);
    scrollSync.value++;
    void nextTick(() => {
      rowVirtualizer.value.measure();
      thumbnailFetchGeneration++;
      pruneThumbnailsToRetained();
      void syncThumbnailsForVisible();
    });
  });
  ro.observe(el);
  gridInnerWidth.value = Math.floor(el.getBoundingClientRect().width);
  void nextTick(() => {
    rowVirtualizer.value.measure();
    thumbnailFetchGeneration++;
    pruneThumbnailsToRetained();
    void syncThumbnailsForVisible();
  });
  onCleanup(() => ro.disconnect());
});

watch([rowHeightPx, flatGalleryRows], () => {
  void nextTick(() => {
    rowVirtualizer.value.measure();
    scrollSync.value++;
    thumbnailFetchGeneration++;
    pruneThumbnailsToRetained();
    void syncThumbnailsForVisible();
  });
});

watch(
  () => selected.value?.id,
  () => {
    thumbnailFetchGeneration++;
    pruneThumbnailsToRetained();
    void syncThumbnailsForVisible();
  },
);

const joinButtonTitle = computed(() => {
  if (!selected.value?.worldId || selected.value.worldId.trim() === "") {
    return "World ID がありません";
  }
  return "このワールドへJoin";
});

function thumbnailSrc(item: ScreenshotDTO): string {
  const u = thumbnailUrls.value[item.id];
  if (u) return u;
  return transparentPixelDataUrl;
}

function onThumbnailError(e: Event): void {
  const img = e.target as HTMLImageElement;
  img.src = missingThumbDataUrl;
}

function onGridScroll(): void {
  scrollSync.value++;
  void syncThumbnailsForVisible();
  if (thumbnailPruneScrollTimer !== null) {
    clearTimeout(thumbnailPruneScrollTimer);
  }
  thumbnailPruneScrollTimer = setTimeout(() => {
    thumbnailPruneScrollTimer = null;
    thumbnailFetchGeneration++;
    pruneThumbnailsToRetained();
    void syncThumbnailsForVisible();
  }, THUMBNAIL_PRUNE_SCROLL_DEBOUNCE_MS);
}

/** IDs in virtualizer-rendered rows (includes overscan) plus the selected screenshot. */
function visibleScreenshotIds(): string[] {
  if (list.value.length === 0) {
    return [];
  }
  const vItems = rowVirtualizer.value.getVirtualItems();
  const idSet = new Set<string>();
  for (const v of vItems) {
    const row = flatGalleryRows.value[v.index];
    if (row?.type === "grid") {
      for (const it of row.items) {
        idSet.add(it.id);
      }
    }
  }
  const sel = selected.value?.id;
  if (sel) {
    idSet.add(sel);
  }
  return [...idSet];
}

function pruneThumbnailsToRetained(): void {
  const listIds = new Set(list.value.map((i) => i.id));
  const retained = new Set(visibleScreenshotIds());
  thumbnailUrls.value = pruneThumbnailUrlMap(
    thumbnailUrls.value,
    listIds,
    retained,
  );
}

async function syncThumbnailsForVisible(): Promise<void> {
  const gen = thumbnailFetchGeneration;
  const ids = visibleScreenshotIds();
  const toFetch = ids.filter((id) => thumbnailUrls.value[id] === undefined);
  if (toFetch.length === 0) {
    return;
  }

  let cursor = 0;
  async function worker(): Promise<void> {
    while (gen === thumbnailFetchGeneration) {
      const i = cursor++;
      if (i >= toFetch.length) {
        return;
      }
      const id = toFetch[i]!;
      try {
        const url = await App.screenshotThumbnailDataURL(id);
        if (gen !== thumbnailFetchGeneration) {
          return;
        }
        thumbnailUrls.value = {
          ...thumbnailUrls.value,
          [id]: url && url.length > 0 ? url : missingThumbDataUrl,
        };
      } catch {
        if (gen !== thumbnailFetchGeneration) {
          return;
        }
        thumbnailUrls.value = {
          ...thumbnailUrls.value,
          [id]: missingThumbDataUrl,
        };
      }
    }
  }

  await Promise.all(
    Array.from({ length: THUMBNAIL_FETCH_CONCURRENCY }, () => worker()),
  );
}

watch(
  list,
  () => {
    thumbnailFetchGeneration++;
    const listIds = new Set(list.value.map((i) => i.id));
    thumbnailUrls.value = pruneThumbnailUrlMap(
      thumbnailUrls.value,
      listIds,
      listIds,
    );
    void nextTick(() => {
      scrollSync.value++;
      rowVirtualizer.value.measure();
      void nextTick(() => {
        pruneThumbnailsToRetained();
        void syncThumbnailsForVisible();
      });
    });
  },
  { deep: true },
);

function formatTakenAt(takenAt?: string): string {
  if (!takenAt) return "—";
  try {
    const d = new Date(takenAt);
    return d.toLocaleString("ja-JP");
  } catch {
    return takenAt;
  }
}

function formatFileSize(bytes?: number): string {
  if (bytes == null || bytes < 0 || !Number.isFinite(bytes)) {
    return "—";
  }
  if (bytes === 0) {
    return "0 B";
  }
  const units = ["B", "KB", "MB", "GB", "TB"];
  let v = bytes;
  let u = 0;
  while (v >= 1024 && u < units.length - 1) {
    v /= 1024;
    u++;
  }
  const rounded = u === 0 || v >= 10 ? Math.round(v).toString() : v.toFixed(1);
  return `${rounded} ${units[u]}`;
}

function fileNameFromPath(path: string): string {
  const norm = path.replace(/\\/g, "/");
  const i = norm.lastIndexOf("/");
  return i >= 0 ? norm.slice(i + 1) : norm;
}

async function load(): Promise<void> {
  loadError.value = null;
  loading.value = true;
  try {
    const wid = filterWorldId.value.trim();
    if (wid) {
      const filter: ScreenshotSearchDTO = { worldId: wid };
      list.value = await App.searchScreenshots(filter);
    } else {
      list.value = await App.screenshots("");
    }
    if (
      selected.value &&
      !list.value.find((s) => s.id === selected.value?.id)
    ) {
      selected.value = null;
    }
  } catch (err) {
    loadError.value = err instanceof Error ? err.message : String(err);
    list.value = [];
  } finally {
    loading.value = false;
  }
}

function onFilterEnter(): void {
  if (filterDebounceTimer !== null) {
    clearTimeout(filterDebounceTimer);
    filterDebounceTimer = null;
  }
  void load();
}

watch(filterWorldId, () => {
  if (filterDebounceTimer !== null) {
    clearTimeout(filterDebounceTimer);
  }
  filterDebounceTimer = setTimeout(() => {
    filterDebounceTimer = null;
    void load();
  }, FILTER_DEBOUNCE_MS);
});

onBeforeUnmount(() => {
  thumbnailFetchGeneration++;
  if (filterDebounceTimer !== null) {
    clearTimeout(filterDebounceTimer);
  }
  if (thumbnailPruneScrollTimer !== null) {
    clearTimeout(thumbnailPruneScrollTimer);
    thumbnailPruneScrollTimer = null;
  }
});

/** True when Go returns missing config.json (VRChatConfigFileRepository.Read). */
function isMissingVRChatConfigError(err: unknown): boolean {
  const msg = err instanceof Error ? err.message : String(err);
  return msg.includes("config.json does not exist");
}

async function scanFolder(): Promise<void> {
  scanError.value = null;
  loadError.value = null;
  scanning.value = true;
  try {
    let path = "";
    try {
      const cfg = await App.getVRChatConfig();
      path = (cfg.pictureOutputFolder ?? "").trim();
    } catch (err) {
      if (!isMissingVRChatConfigError(err)) {
        scanError.value = err instanceof Error ? err.message : String(err);
        return;
      }
    }
    if (!path) {
      try {
        path = (await App.defaultVRChatPictureFolder()).trim();
      } catch {
        path = "";
      }
      if (!path) {
        scanError.value =
          "デフォルトの保存先（ユーザーフォルダー内の「ピクチャ」／「マイ ピクチャ」にある VRChat フォルダ）を解決できませんでした。";
        return;
      }
    }
    try {
      await App.scanScreenshotDir(path);
    } catch (err) {
      scanError.value = err instanceof Error ? err.message : String(err);
      return;
    }
    await load();
  } finally {
    scanning.value = false;
  }
}

function select(item: ScreenshotDTO): void {
  selected.value = item;
  joinError.value = null;
}

const joinError = ref<string | null>(null);

async function onJoin(): Promise<void> {
  if (!selected.value?.worldId || selected.value.worldId.trim() === "") return;
  joinError.value = null;
  try {
    await App.joinWorldFromScreenshot(selected.value.id);
  } catch (err) {
    joinError.value = err instanceof Error ? err.message : String(err);
  }
}

onMounted(() => {
  void load().then(() => {
    void nextTick(() => {
      scrollSync.value++;
      void syncThumbnailsForVisible();
    });
  });
});
</script>

<style scoped>
.gallery-view {
  flex: 1;
  min-height: 0;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  overflow: hidden;
}

.page-title {
  margin: 0;
  font-size: 1.5rem;
  flex-shrink: 0;
}

.filters {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
  flex-shrink: 0;
}

.filter-input {
  flex: 1;
  min-width: 12rem;
  max-width: 24rem;
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

.btn-refresh:hover:not(:disabled) {
  background: var(--accent);
  color: white;
  border-color: var(--accent);
}

.btn-scan {
  padding: 0.5rem 1rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
  cursor: pointer;
}

.btn-scan:hover:not(:disabled) {
  background: var(--accent);
  color: white;
  border-color: var(--accent);
}

.btn-refresh:disabled,
.btn-scan:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.banner-error {
  margin: 0;
  padding: 0.5rem 0.75rem;
  border-radius: var(--radius);
  font-size: 0.9rem;
  background: color-mix(in srgb, var(--accent) 15%, transparent);
  color: var(--text-primary);
  border: 1px solid color-mix(in srgb, var(--accent) 40%, transparent);
  flex-shrink: 0;
}

.banner-warn {
  background: color-mix(in srgb, var(--text-secondary) 12%, transparent);
  border-color: var(--border);
}

.gallery-body {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 1rem;
  align-items: stretch;
  min-height: 0;
  min-width: 0;
}

@media (min-width: 960px) {
  .gallery-body {
    flex-direction: row;
    align-items: stretch;
  }

  .grid-section {
    flex: 1;
    min-width: 0;
    min-height: 0;
  }

  .detail-panel {
    width: min(320px, 100%);
    flex-shrink: 0;
    align-self: stretch;
    overflow-y: auto;
  }
}

.loading,
.empty {
  padding: 2rem;
  text-align: center;
  color: var(--text-secondary);
}

.grid-section {
  flex: 1;
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.grid-scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
}

.grid-item {
  border-radius: var(--radius);
  overflow: hidden;
  cursor: pointer;
  border: 2px solid transparent;
  transition:
    border-color 0.15s,
    box-shadow 0.15s;
  box-sizing: border-box;
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
  flex-shrink: 0;
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

.gallery-group-header {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0 0.35rem;
  margin: 0;
  background: color-mix(in srgb, var(--bg-secondary) 88%, transparent);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
  font-size: 0.9rem;
  cursor: pointer;
  text-align: left;
  box-sizing: border-box;
}

.gallery-group-header:hover {
  background: var(--bg-tertiary);
}

.gallery-group-h-year {
  font-weight: 600;
}

.gallery-group-h-month {
  padding-left: 1.1rem;
  font-weight: 550;
}

.gallery-group-h-day {
  padding-left: 2rem;
  font-weight: 450;
}

.gallery-group-chevron {
  flex-shrink: 0;
  width: 0.85rem;
  font-size: 0.6rem;
  opacity: 0.85;
  line-height: 1;
}

.gallery-group-label {
  flex: 1;
  min-width: 0;
}
</style>
