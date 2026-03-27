<template>
  <div class="gallery-view">
    <h1 class="page-title">ギャラリー</h1>

    <div class="filters">
      <el-input
        v-model="filterWorldId"
        data-testid="gallery-world-filter"
        type="search"
        placeholder="World ID で検索（入力で自動検索 / Enter）"
        clearable
        style="flex: 1; max-width: 400px"
        @keyup.enter="onFilterEnter"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-button :disabled="loading || scanning" @click="onRefreshClick">
        更新
      </el-button>
      <el-button
        data-testid="gallery-scan-folder"
        :disabled="loading || scanning"
        :loading="scanning"
        @click="scanFolder"
      >
        {{ scanning ? "スキャン中…" : "Scan Folder" }}
      </el-button>
    </div>

    <el-alert
      v-if="loadError"
      :title="loadError"
      type="error"
      :closable="false"
      show-icon
    />
    <el-alert
      v-if="scanError"
      :title="scanError"
      type="warning"
      :closable="false"
      show-icon
    />

    <div class="gallery-body">
      <!-- グリッド一覧 -->
      <div class="grid-section">
        <div
          v-if="scanning"
          class="loading gallery-scan-progress"
          data-testid="gallery-scan-progress"
        >
          <p class="gallery-scan-status">{{ scanStatusText }}</p>
          <el-progress
            v-if="scanProgressDeterminate"
            :percentage="
              Math.round(
                ((scanProgress?.current ?? 0) /
                  Math.max(1, scanProgress?.total ?? 1)) *
                  100,
              )
            "
            :striped="true"
            :striped-flow="true"
          />
          <el-progress
            v-else
            :percentage="100"
            status="striped"
            :striped="true"
            :striped-flow="true"
            :duration="10"
          />
        </div>
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

      <!-- 詳細プレビュー -->
      <el-card v-if="selected" class="detail-panel" shadow="never">
        <template #header>詳細</template>
        <div class="detail-preview">
          <img
            data-testid="gallery-detail-preview"
            :src="thumbnailSrc(selected)"
            :alt="fileNameFromPath(selected.filePath)"
            class="detail-preview-img"
            @error="onThumbnailError"
          />
        </div>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="ファイル名">
            {{ fileNameFromPath(selected.filePath) }}
          </el-descriptions-item>
          <el-descriptions-item label="ファイルサイズ">
            {{ formatFileSize(selected.fileSizeBytes) }}
          </el-descriptions-item>
          <el-descriptions-item label="撮影日時">
            {{ formatTakenAt(selected.takenAt) }}
          </el-descriptions-item>
          <el-descriptions-item label="ワールド名">
            {{ selected.worldName || "—" }}
          </el-descriptions-item>
          <el-descriptions-item label="作者表示名">
            {{ selected.authorDisplayName || "—" }}
          </el-descriptions-item>
          <el-descriptions-item label="ファイルパス">
            <el-button
              link
              type="primary"
              data-testid="gallery-detail-open-file"
              title="既定のアプリで画像を開く"
              class="file-path-btn"
              @click="openSelectedFileExternally"
            >
              {{ selected.filePath }}
            </el-button>
          </el-descriptions-item>
        </el-descriptions>
        <el-alert
          v-if="detailActionError"
          :title="detailActionError"
          type="error"
          :closable="false"
          show-icon
          style="margin: 0.75rem 0"
        />
        <el-button
          style="width: 100%; margin: 0.75rem 0 0.5rem"
          data-testid="gallery-detail-open-folder"
          :title="openFolderButtonTitle"
          @click="revealSelectedInFolder"
        >
          フォルダを開く
        </el-button>
        <el-alert
          v-if="joinError"
          :title="joinError"
          type="error"
          :closable="false"
          show-icon
          style="margin-bottom: 0.5rem"
        />
        <el-button
          type="primary"
          style="width: 100%"
          :disabled="!selected.worldId || selected.worldId.trim() === ''"
          :title="joinButtonTitle"
          @click="onJoin"
        >
          このワールドへJoin
        </el-button>
      </el-card>
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
  type ScanProgressPayload,
  type GalleryScanDonePayload,
} from "../wails/app";
import { getRuntime } from "../wails/runtime";
import {
  buildGalleryVirtualRows,
  galleryRowHeight,
  type GalleryVirtualRow,
} from "./galleryDateGroups";
import { pruneThumbnailUrlMap } from "./galleryThumbnailCache";

const FILTER_DEBOUNCE_MS = 400;
const GALLERY_SCREENSHOTS_CHANGED_DEBOUNCE_MS = 400;
const THUMBNAIL_PRUNE_SCROLL_DEBOUNCE_MS = 150;
const THUMBNAIL_FETCH_CONCURRENCY = 4;
const GRID_GAP_PX = 12;
const MIN_CELL_WIDTH = 140;

const missingThumbDataUrl =
  "data:image/svg+xml," +
  encodeURIComponent(
    '<svg xmlns="http://www.w3.org/2000/svg" width="120" height="90" viewBox="0 0 120 90"><rect fill="#333" width="120" height="90"/><text x="50%" y="50%" dominant-baseline="middle" text-anchor="middle" fill="#666" font-size="12">画像なし</text></svg>',
  );

const transparentPixelDataUrl =
  "data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7";

const list = ref<ScreenshotDTO[]>([]);
const selected = ref<ScreenshotDTO | null>(null);
const loading = ref(false);
const scanning = ref(false);
const scanProgress = ref<ScanProgressPayload | null>(null);
const loadError = ref<string | null>(null);
const scanError = ref<string | null>(null);
const filterWorldId = ref("");
const thumbnailUrls = ref<Record<string, string>>({});
const collapsed = ref(new Set<string>());
const gridScrollRef = ref<HTMLElement | null>(null);
const gridInnerWidth = ref(0);
const scrollSync = ref(0);

let filterDebounceTimer: ReturnType<typeof setTimeout> | null = null;
let thumbnailPruneScrollTimer: ReturnType<typeof setTimeout> | null = null;
let thumbnailFetchGeneration = 0;
let unsubscribeScanProgress: (() => void) | undefined;
let unsubscribeScanDone: (() => void) | undefined;
let unsubscribeScreenshotsChanged: (() => void) | undefined;
let screenshotsChangedDebounceTimer: ReturnType<typeof setTimeout> | null =
  null;

const scanProgressDeterminate = computed(() => {
  const p = scanProgress.value;
  return p?.phase === "importing" && p.total > 0;
});

const scanStatusText = computed(() => {
  const p = scanProgress.value;
  if (!p) return "フォルダをスキャンしています…";
  if (p.phase === "listing")
    return `画像ファイルを検索しています…（${p.current} 件）`;
  if (p.phase === "importing") {
    if (p.total === 0) return "画像ファイルは見つかりませんでした";
    if (p.current === 0) return `画像 ${p.total} 件を取り込みます…`;
    if (p.item) return `取り込み中: ${p.item}（${p.current} / ${p.total}）`;
    return `取り込み中（${p.current} / ${p.total}）`;
  }
  return "フォルダをスキャンしています…";
});

function applyScanProgressPayload(data: unknown): void {
  if (typeof data !== "object" || data === null) return;
  const o = data as Record<string, unknown>;
  if (typeof o.phase !== "string") return;
  if (typeof o.current !== "number" || typeof o.total !== "number") return;
  const item = o.item;
  scanProgress.value = {
    phase: o.phase,
    current: o.current,
    total: o.total,
    item: typeof item === "string" ? item : "",
  };
}

function applyGalleryScanDonePayload(data: unknown): void {
  let payload: GalleryScanDonePayload = { count: 0 };
  if (typeof data === "object" && data !== null) {
    const o = data as Record<string, unknown>;
    if (typeof o.count === "number") {
      payload = {
        count: o.count,
        error: typeof o.error === "string" ? o.error : undefined,
        cancelled: o.cancelled === true,
      };
    }
  }
  scanning.value = false;
  scanProgress.value = null;
  if (payload.cancelled) {
    scanError.value = null;
  } else if (payload.error) {
    scanError.value = payload.error;
  } else {
    scanError.value = null;
  }
  void load().then(() => {
    void nextTick(() => {
      scrollSync.value++;
      void syncThumbnailsForVisible();
    });
  });
}

const columnCount = computed(() => {
  const w = gridInnerWidth.value;
  if (w <= 0) return 1;
  return Math.max(
    1,
    Math.floor((w + GRID_GAP_PX) / (MIN_CELL_WIDTH + GRID_GAP_PX)),
  );
});

const cellWidthPx = computed(() => {
  const cols = columnCount.value;
  const w = gridInnerWidth.value;
  if (cols <= 0 || w <= 0) return MIN_CELL_WIDTH;
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
      if (!row) return rowHeightPx.value;
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
  if (row.type === "yearHeader") return "gallery-group-h-year";
  if (row.type === "monthHeader") return "gallery-group-h-month";
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
  if (!el || list.value.length === 0) return;
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

const openFolderButtonTitle =
  "画像があるフォルダをファイルマネージャで開きます（環境によってはフォルダのみの場合があります）";

const detailActionError = ref<string | null>(null);

async function openSelectedFileExternally(): Promise<void> {
  if (!selected.value) return;
  detailActionError.value = null;
  try {
    await App.openScreenshotExternally(selected.value.id);
  } catch (err) {
    detailActionError.value = err instanceof Error ? err.message : String(err);
  }
}

async function revealSelectedInFolder(): Promise<void> {
  if (!selected.value) return;
  detailActionError.value = null;
  try {
    await App.revealScreenshotInFileManager(selected.value.id);
  } catch (err) {
    detailActionError.value = err instanceof Error ? err.message : String(err);
  }
}

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

function visibleScreenshotIds(): string[] {
  if (list.value.length === 0) return [];
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
  if (sel) idSet.add(sel);
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
  if (toFetch.length === 0) return;

  let cursor = 0;
  async function worker(): Promise<void> {
    while (gen === thumbnailFetchGeneration) {
      const i = cursor++;
      if (i >= toFetch.length) return;
      const id = toFetch[i]!;
      try {
        const url = await App.screenshotThumbnailDataURL(id);
        if (gen !== thumbnailFetchGeneration) return;
        thumbnailUrls.value = {
          ...thumbnailUrls.value,
          [id]: url && url.length > 0 ? url : missingThumbDataUrl,
        };
      } catch {
        if (gen !== thumbnailFetchGeneration) return;
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

watch(list, () => {
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
});

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
  if (bytes == null || bytes < 0 || !Number.isFinite(bytes)) return "—";
  if (bytes === 0) return "0 B";
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

function onRefreshClick(): void {
  if (filterDebounceTimer !== null) {
    clearTimeout(filterDebounceTimer);
    filterDebounceTimer = null;
  }
  void load();
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

function scheduleLoadFromPictureWatcher(): void {
  if (screenshotsChangedDebounceTimer !== null) {
    clearTimeout(screenshotsChangedDebounceTimer);
  }
  screenshotsChangedDebounceTimer = setTimeout(() => {
    screenshotsChangedDebounceTimer = null;
    void load();
  }, GALLERY_SCREENSHOTS_CHANGED_DEBOUNCE_MS);
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
  unsubscribeScanProgress?.();
  unsubscribeScanProgress = undefined;
  unsubscribeScanDone?.();
  unsubscribeScanDone = undefined;
  unsubscribeScreenshotsChanged?.();
  unsubscribeScreenshotsChanged = undefined;
  if (screenshotsChangedDebounceTimer !== null) {
    clearTimeout(screenshotsChangedDebounceTimer);
    screenshotsChangedDebounceTimer = null;
  }
  if (filterDebounceTimer !== null) {
    clearTimeout(filterDebounceTimer);
  }
  if (thumbnailPruneScrollTimer !== null) {
    clearTimeout(thumbnailPruneScrollTimer);
    thumbnailPruneScrollTimer = null;
  }
});

async function scanFolder(): Promise<void> {
  scanError.value = null;
  loadError.value = null;
  scanProgress.value = null;
  scanning.value = true;
  let goScanStarted = false;
  try {
    let path = "";
    try {
      const cfg = await App.getVRChatConfig();
      path = (cfg.pictureOutputFolder ?? "").trim();
    } catch {
      path = "";
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
    goScanStarted = true;
    await App.scanScreenshotDir(path);
  } catch (err) {
    if (scanning.value) {
      scanning.value = false;
      scanProgress.value = null;
      scanError.value = err instanceof Error ? err.message : String(err);
    }
  } finally {
    if (!goScanStarted) {
      scanning.value = false;
      scanProgress.value = null;
    }
  }
}

function select(item: ScreenshotDTO): void {
  selected.value = item;
  joinError.value = null;
  detailActionError.value = null;
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
  const rt = getRuntime();
  const offProgress = rt?.EventsOn?.(
    "gallery:scan-progress",
    (data?: unknown) => {
      applyScanProgressPayload(data);
    },
  );
  if (typeof offProgress === "function") {
    unsubscribeScanProgress = offProgress;
  }
  const offDone = rt?.EventsOn?.("gallery:scan-done", (data?: unknown) => {
    applyGalleryScanDonePayload(data);
  });
  if (typeof offDone === "function") {
    unsubscribeScanDone = offDone;
  }
  const offChanged = rt?.EventsOn?.("gallery:screenshots-changed", () => {
    scheduleLoadFromPictureWatcher();
  });
  if (typeof offChanged === "function") {
    unsubscribeScreenshotsChanged = offChanged;
  }

  void (async () => {
    try {
      if (await App.isGalleryScanning()) {
        scanning.value = true;
      }
    } catch {
      /* ignore */
    }
    await load();
    await nextTick();
    scrollSync.value++;
    void syncThumbnailsForVisible();
  })();
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

.filters {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
  flex-shrink: 0;
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

.gallery-scan-progress {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  padding: 2rem;
}

.gallery-scan-status {
  margin: 0;
  font-size: 0.95rem;
  color: var(--text-secondary);
  max-width: 28rem;
  word-break: break-all;
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
  flex-shrink: 0;
  background: var(--bg-secondary) !important;
  border-color: var(--border) !important;
}

.detail-panel :deep(.el-card__header) {
  font-weight: 600;
  border-bottom-color: var(--border);
}

.detail-preview {
  margin: 0 0 1rem;
  border-radius: var(--radius);
  overflow: hidden;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 4rem;
}

.detail-preview-img {
  display: block;
  width: 100%;
  max-height: 260px;
  object-fit: contain;
}

.file-path-btn {
  word-break: break-all;
  white-space: normal;
  text-align: left;
  height: auto !important;
  line-height: 1.4 !important;
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
