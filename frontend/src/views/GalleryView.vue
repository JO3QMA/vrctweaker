<template>
  <div class="gallery-view">
    <h1 class="page-title">{{ t("routes.gallery") }}</h1>

    <div class="filters">
      <VtInput
        v-model="filterWorldSearch"
        data-testid="gallery-world-filter"
        type="search"
        :placeholder="t('gallery.searchPlaceholder')"
        clearable
        class="gallery-world-filter"
        @keyup.enter="onFilterEnter"
      >
        <template #prefix>
          <VtIcon size="default"><Search /></VtIcon>
        </template>
      </VtInput>
      <el-date-picker
        v-model="filterDateRange"
        data-testid="gallery-date-range"
        type="daterange"
        :start-placeholder="t('gallery.dateRangeStart')"
        :end-placeholder="t('gallery.dateRangeEnd')"
        format="YYYY-MM-DD"
        value-format="YYYY-MM-DD"
        clearable
        class="gallery-date-range"
        @change="onFilterEnter"
      />
      <VtButton
        variant="tertiary"
        :disabled="loading || scanning"
        @click="onRefreshClick"
      >
        {{ t("common.refresh") }}
      </VtButton>
      <VtButton
        variant="secondary"
        data-testid="gallery-scan-folder"
        :disabled="loading || scanning"
        :loading="scanning"
        @click="scanFolder"
      >
        {{ scanning ? t("gallery.scanning") : t("gallery.scanFolder") }}
      </VtButton>
    </div>

    <VtAlert v-if="loadError" variant="danger" :title="loadError" />
    <VtAlert v-if="scanError" variant="warning" :title="scanError" />

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
            :striped="true"
            :striped-flow="true"
            :duration="10"
          />
        </div>
        <div v-else-if="loading" class="loading">{{ t("common.loading") }}</div>
        <div v-else-if="list.length === 0" class="empty">
          {{ t("gallery.empty") }}
        </div>
        <div
          v-else
          ref="gridScrollRef"
          data-testid="gallery-grid-scroll"
          class="grid-scroll"
          @scroll.passive="onGridScroll"
        >
          <!-- Visual sticky duplicate; in-list role="heading" rows are canonical. -->
          <div class="gallery-sticky-anchor">
            <div
              v-show="stickyOverlay.show"
              class="gallery-sticky-header gallery-sticky-header--overlay"
              :class="stickyOverlay.headerClass"
              data-testid="gallery-sticky-header"
              aria-hidden="true"
            >
              <span
                v-if="stickyOverlay.showYearContext"
                class="gallery-sticky-year"
              >
                {{ stickyOverlay.yearLabel }}
              </span>
              <span class="gallery-sticky-label">{{
                stickyOverlay.label
              }}</span>
            </div>
          </div>
          <div class="grid-scroll-inner">
            <div class="grid-virtual-spacer" :style="spacerStyle">
              <div
                v-for="rowView in virtualRowViews"
                :key="virtualRowDomKey(rowView.vr.index)"
                class="grid-virtual-row"
                :style="virtualRowStyle(rowView.vr)"
              >
                <template v-if="rowView.isGrid">
                  <div class="grid-row-inner" :style="gridRowInnerStyle">
                    <div
                      v-for="item in rowView.gridItems"
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
                <div
                  v-else-if="rowView.header"
                  class="gallery-section-header"
                  :class="galleryHeaderClass(rowView.header)"
                  data-testid="gallery-group-header"
                  role="heading"
                  :aria-level="rowView.header.type === 'yearHeader' ? 2 : 3"
                >
                  <span class="gallery-section-label">{{
                    rowView.header.label
                  }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 詳細プレビュー -->
      <el-card v-if="selected" class="section-card detail-panel" shadow="never">
        <template #header>{{ t("gallery.detail") }}</template>
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
          <el-descriptions-item :label="t('gallery.fileName')">
            {{ fileNameFromPath(selected.filePath) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('gallery.fileSize')">
            {{ formatFileSize(selected.fileSizeBytes) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('gallery.takenAt')">
            {{ formatTakenAt(selected.takenAt) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('gallery.worldName')">
            {{ selected.worldName || t("common.dash") }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('gallery.participants')">
            <ul
              v-if="selectedParticipants.length > 0"
              class="gallery-participants-list"
              data-testid="gallery-detail-participants"
            >
              <li
                v-for="(participant, index) in selectedParticipants"
                :key="participantKey(participant, index)"
                class="gallery-participants-item"
              >
                <VtButton
                  v-if="participant.vrcUserId"
                  variant="primary"
                  link
                  class="gallery-participant-link"
                  :data-testid="`gallery-participant-link-${index}`"
                  @click="openParticipantProfile(participant)"
                >
                  {{ participantLabel(participant) }}
                </VtButton>
                <span v-else class="gallery-participant-name">{{
                  participantLabel(participant)
                }}</span>
              </li>
            </ul>
            <span v-else>{{ t("common.dash") }}</span>
          </el-descriptions-item>
          <el-descriptions-item :label="t('gallery.authorDisplayName')">
            {{ selected.authorDisplayName || t("common.dash") }}
          </el-descriptions-item>
          <el-descriptions-item
            v-if="selected.enrichmentInstanceId"
            :label="t('gallery.instanceId')"
          >
            {{ selected.enrichmentInstanceId }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('gallery.filePath')">
            <VtButton
              variant="primary"
              link
              data-testid="gallery-detail-open-file"
              :title="t('gallery.openWithDefaultApp')"
              class="file-path-btn"
              @click="openSelectedFileExternally"
            >
              {{ selected.filePath }}
            </VtButton>
          </el-descriptions-item>
        </el-descriptions>
        <VtAlert
          v-if="detailActionError"
          variant="danger"
          class="detail-action-alert"
          :title="detailActionError"
        />
        <VtButton
          variant="secondary"
          class="detail-panel-action-btn"
          data-testid="gallery-detail-open-folder"
          :title="openFolderButtonTitle"
          @click="revealSelectedInFolder"
        >
          {{ t("gallery.openFolder") }}
        </VtButton>
        <VtAlert
          v-if="joinError"
          variant="danger"
          class="join-error-alert"
          :title="joinError"
        />
        <VtButton
          variant="primary"
          class="detail-panel-action-btn detail-panel-action-btn--last"
          :disabled="!selected.worldId || selected.worldId.trim() === ''"
          :title="joinButtonTitle"
          @click="onJoin"
        >
          {{ t("gallery.joinWorld") }}
        </VtButton>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Search } from "@element-plus/icons-vue";
import { useVirtualizer, type VirtualItem } from "@tanstack/vue-virtual";
import VtAlert from "../components/VtAlert.vue";
import VtButton from "../components/VtButton.vue";
import VtIcon from "../components/VtIcon.vue";
import VtInput from "../components/VtInput.vue";
import {
  ref,
  onMounted,
  onBeforeUnmount,
  computed,
  watch,
  watchEffect,
  nextTick,
} from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import {
  App,
  type ScreenshotDTO,
  type ScanProgressPayload,
  type GalleryScanDonePayload,
} from "../wails/app";
import { getRuntime } from "../wails/runtime";
import {
  buildGalleryHeaderIndices,
  buildGalleryVirtualRows,
  galleryLabelsFromLocale,
  galleryRowHeight,
  stickySectionForIndex,
  type GalleryStickySection,
  type GalleryVirtualRow,
} from "./galleryDateGroups";
import { pruneThumbnailUrlMap } from "./galleryThumbnailCache";
import {
  buildGallerySearchFilter,
  type GalleryDateRangeFilter,
} from "./gallerySearchFilter";
import { formatEncounteredAt } from "../utils/formatEncounteredAt";
import { navigateToUserProfile } from "../utils/userProfileNavigation";
import { appLocaleToBcp47 } from "../i18n";

const { t, locale } = useI18n();
const router = useRouter();

const FILTER_DEBOUNCE_MS = 400;
const GALLERY_SCREENSHOTS_CHANGED_DEBOUNCE_MS = 400;
const THUMBNAIL_PRUNE_SCROLL_DEBOUNCE_MS = 150;
const THUMBNAIL_FETCH_CONCURRENCY = 4;
const GRID_GAP_PX = 12;
const MIN_CELL_WIDTH = 140;

function escapeSvgText(s: string): string {
  return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}

const missingThumbDataUrl = computed(() => {
  const label = escapeSvgText(t("gallery.noImage"));
  return (
    "data:image/svg+xml," +
    encodeURIComponent(
      `<svg xmlns="http://www.w3.org/2000/svg" width="120" height="90" viewBox="0 0 120 90"><rect fill="#333" width="120" height="90"/><text x="50%" y="50%" dominant-baseline="middle" text-anchor="middle" fill="#666" font-size="12">${label}</text></svg>`,
    )
  );
});

const transparentPixelDataUrl =
  "data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7";

const list = ref<ScreenshotDTO[]>([]);
const selected = ref<ScreenshotDTO | null>(null);
const loading = ref(false);
const scanning = ref(false);
const scanProgress = ref<ScanProgressPayload | null>(null);
const loadError = ref<string | null>(null);
const scanError = ref<string | null>(null);
const filterWorldSearch = ref("");
const filterDateRange = ref<GalleryDateRangeFilter | null>(null);
const thumbnailUrls = ref<Record<string, string>>({});
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
  if (!p) return t("gallery.scanDefault");
  if (p.phase === "listing")
    return t("gallery.scanListing", { n: String(p.current) });
  if (p.phase === "importing") {
    if (p.total === 0) return t("gallery.scanNoFiles");
    if (p.current === 0)
      return t("gallery.scanImportPrepare", { n: String(p.total) });
    if (p.item)
      return t("gallery.scanImportItem", {
        item: p.item,
        current: String(p.current),
        total: String(p.total),
      });
    return t("gallery.scanImporting", {
      current: String(p.current),
      total: String(p.total),
    });
  }
  return t("gallery.scanDefault");
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

const galleryDateLabels = computed(() =>
  galleryLabelsFromLocale(
    appLocaleToBcp47(String(locale.value)),
    t("gallery.unknownDate"),
  ),
);

const flatGalleryRows = computed(() =>
  buildGalleryVirtualRows(
    list.value,
    columnCount.value,
    galleryDateLabels.value,
  ),
);

const galleryHeaderIndices = computed(() =>
  buildGalleryHeaderIndices(flatGalleryRows.value),
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
  void scrollSync.value;
  return rowVirtualizer.value.getVirtualItems();
});

type StickyOverlayState = {
  show: boolean;
  label: string;
  headerClass: string;
  yearLabel: string;
  showYearContext: boolean;
};

const emptyStickyOverlay: StickyOverlayState = {
  show: false,
  label: "",
  headerClass: "",
  yearLabel: "",
  showYearContext: false,
};

function anchorVirtualItemForSticky(
  vItems: VirtualItem[],
  scrollTop: number,
): VirtualItem | undefined {
  const firstVisible = vItems.find((v) => v.start + v.size > scrollTop);
  if (firstVisible) {
    return firstVisible;
  }
  // All rendered rows are above scrollTop (e.g. after a scroll jump).
  return vItems[vItems.length - 1];
}

function isHeaderPinnedAtScrollTop(v: VirtualItem, scrollTop: number): boolean {
  return v.start <= scrollTop && v.start + v.size > scrollTop;
}

function stickyHeaderClass(section: GalleryStickySection): string {
  return section.headerKind === "yearHeader"
    ? "gallery-sticky-header--year"
    : "gallery-sticky-header--day";
}

function isMatchingHeaderPinnedAtTop(
  vItems: VirtualItem[],
  rows: GalleryVirtualRow[],
  scrollTop: number,
  rowKey: string,
): boolean {
  for (const v of vItems) {
    const row = rows[v.index];
    if (
      (row?.type === "yearHeader" || row?.type === "dayHeader") &&
      row.rowKey === rowKey &&
      isHeaderPinnedAtScrollTop(v, scrollTop)
    ) {
      return true;
    }
  }
  return false;
}

const stickyOverlay = computed((): StickyOverlayState => {
  void scrollSync.value;
  const vItems = virtualRows.value;
  const rows = flatGalleryRows.value;
  const scrollEl = gridScrollRef.value;
  if (vItems.length === 0 || rows.length === 0 || !scrollEl) {
    return emptyStickyOverlay;
  }

  const scrollTop = scrollEl.scrollTop;
  const clientHeight = scrollEl.clientHeight;
  const firstVisible = anchorVirtualItemForSticky(vItems, scrollTop);
  if (!firstVisible) {
    return emptyStickyOverlay;
  }
  const firstIdx = firstVisible.index;
  const section = stickySectionForIndex(firstIdx, galleryHeaderIndices.value);
  if (!section) {
    return emptyStickyOverlay;
  }

  const headerClass = stickyHeaderClass(section);
  const hideWhenHeaderVisible = (rowKey: string): boolean => {
    if (clientHeight <= 0) {
      // Viewport unknown (init / display:none) — prefer showing sticky.
      return false;
    }
    return isMatchingHeaderPinnedAtTop(vItems, rows, scrollTop, rowKey);
  };

  if (hideWhenHeaderVisible(section.rowKey)) {
    return {
      show: false,
      label: section.label,
      headerClass,
      yearLabel: section.yearLabel ?? "",
      showYearContext: false,
    };
  }

  let showYearContext = false;
  if (
    section.headerKind === "dayHeader" &&
    section.yearLabel &&
    section.yearRowKey &&
    !hideWhenHeaderVisible(section.yearRowKey)
  ) {
    showYearContext = true;
  }

  return {
    show: true,
    label: section.label,
    headerClass,
    yearLabel: section.yearLabel ?? "",
    showYearContext,
  };
});

const totalVirtualHeight = computed(() => {
  void scrollSync.value;
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
  Extract<GalleryVirtualRow, { type: "yearHeader" | "dayHeader" }> | undefined {
  const row = galleryRowAt(index);
  if (row?.type === "yearHeader" || row?.type === "dayHeader") {
    return row;
  }
  return undefined;
}

function galleryHeaderClass(
  row: NonNullable<ReturnType<typeof galleryHeaderAt>>,
): string {
  return row.type === "yearHeader"
    ? "gallery-section-header--year"
    : "gallery-section-header--day";
}

const virtualRowViews = computed(() => {
  void scrollSync.value;
  return virtualRows.value.map((vr) => {
    const header = galleryHeaderAt(vr.index);
    return {
      vr,
      header,
      isGrid: header === undefined && isGridRow(vr.index),
      gridItems: gridRowItems(vr.index),
    };
  });
});

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
    return t("gallery.joinNoWorldId");
  }
  return t("gallery.joinWorld");
});

const openFolderButtonTitle = computed(() => t("gallery.openFolderHint"));

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
  img.src = missingThumbDataUrl.value;
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
  const vItems = virtualRows.value;
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
          [id]: url && url.length > 0 ? url : missingThumbDataUrl.value,
        };
      } catch {
        if (gen !== thumbnailFetchGeneration) return;
        thumbnailUrls.value = {
          ...thumbnailUrls.value,
          [id]: missingThumbDataUrl.value,
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
  if (!takenAt) return t("common.dash");
  return formatEncounteredAt(takenAt, appLocaleToBcp47(String(locale.value)));
}

function formatFileSize(bytes?: number): string {
  if (bytes == null || bytes < 0 || !Number.isFinite(bytes))
    return t("common.dash");
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

let loadGeneration = 0;
let loadForegroundGeneration = 0;

async function load(background = false): Promise<void> {
  loadError.value = null;
  const gen = ++loadGeneration;
  if (!background) {
    loading.value = true;
    loadForegroundGeneration = gen;
  }
  try {
    const filter = buildGallerySearchFilter(
      filterWorldSearch.value,
      filterDateRange.value,
    );
    let next: ScreenshotDTO[];
    if (filter) {
      next = await App.searchScreenshots(filter);
    } else {
      next = await App.screenshots("");
    }
    if (gen !== loadGeneration) return;
    list.value = next;
    if (
      selected.value &&
      !list.value.find((s) => s.id === selected.value?.id)
    ) {
      selected.value = null;
    }
  } catch (err) {
    if (gen !== loadGeneration) return;
    loadError.value = err instanceof Error ? err.message : String(err);
    // バックグラウンド更新の失敗では表示中のグリッドを維持する（#27）
    if (!background) {
      list.value = [];
    }
  } finally {
    if (!background && gen === loadForegroundGeneration) {
      loading.value = false;
    }
  }
}

function scheduleLoadFromPictureWatcher(): void {
  if (screenshotsChangedDebounceTimer !== null) {
    clearTimeout(screenshotsChangedDebounceTimer);
  }
  screenshotsChangedDebounceTimer = setTimeout(() => {
    screenshotsChangedDebounceTimer = null;
    // バックグラウンド更新: loading を立てずにグリッドを差し替えることで
    // スクロール位置を保持する（#27）。
    void load(true);
  }, GALLERY_SCREENSHOTS_CHANGED_DEBOUNCE_MS);
}

function onFilterEnter(): void {
  if (filterDebounceTimer !== null) {
    clearTimeout(filterDebounceTimer);
    filterDebounceTimer = null;
  }
  void load();
}

watch(filterWorldSearch, scheduleDebouncedLoad);
watch(filterDateRange, scheduleDebouncedLoad);

function scheduleDebouncedLoad(): void {
  if (filterDebounceTimer !== null) {
    clearTimeout(filterDebounceTimer);
  }
  filterDebounceTimer = setTimeout(() => {
    filterDebounceTimer = null;
    void load();
  }, FILTER_DEBOUNCE_MS);
}

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

type GalleryParticipant = NonNullable<
  ScreenshotDTO["enrichmentParticipants"]
>[number];

const selectedParticipants = computed((): GalleryParticipant[] => {
  const list = selected.value?.enrichmentParticipants;
  return list?.length ? list : [];
});

function participantLabel(participant: GalleryParticipant): string {
  const name = participant.displayName?.trim();
  return name || t("common.dash");
}

function participantKey(
  participant: GalleryParticipant,
  index: number,
): string {
  const id = participant.vrcUserId?.trim();
  return id || `participant-${index}`;
}

async function openParticipantProfile(
  participant: GalleryParticipant,
): Promise<void> {
  const vrcUserId = participant.vrcUserId?.trim();
  if (!vrcUserId) return;
  await navigateToUserProfile(
    router,
    vrcUserId,
    participant.displayName?.trim() ?? "",
  );
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
  gap: var(--space-block);
  overflow: hidden;
}

.filters {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-action-group);
  align-items: center;
  flex-shrink: 0;
}

.gallery-world-filter {
  flex: 1;
  min-width: 12rem;
  max-width: 400px;
}

.gallery-date-range {
  flex: 0 1 auto;
  min-width: 16rem;
  max-width: 320px;
}

.gallery-body {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: var(--space-block);
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
  padding: var(--space-page);
  text-align: center;
  color: var(--color-text-secondary);
}

.gallery-scan-progress {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-form-field);
  padding: var(--space-page);
}

.gallery-scan-status {
  margin: 0;
  font-size: var(--font-size-14);
  color: var(--color-text-secondary);
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

.grid-scroll-inner {
  position: relative;
  min-height: 100%;
}

.gallery-sticky-anchor {
  position: sticky;
  top: 0;
  z-index: 2;
  height: 0;
  margin: 0;
  padding: 0;
  pointer-events: none;
}

.gallery-sticky-header--overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
}

.gallery-sticky-header {
  display: flex;
  align-items: flex-end;
  box-sizing: border-box;
  padding: 0 var(--space-inline-tight);
  background: color-mix(in srgb, var(--color-bg-base) 90%, transparent);
  backdrop-filter: blur(6px);
  border-bottom: 1px solid var(--color-border);
  line-height: var(--line-height-tight);
}

.gallery-sticky-header--year {
  min-height: 32px;
  padding-top: var(--space-form-field);
  color: var(--color-text-primary);
  font-size: var(--font-size-14);
  font-weight: var(--font-weight-600);
}

.gallery-sticky-header--day {
  min-height: 28px;
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  font-weight: var(--font-weight-500);
}

.gallery-sticky-year {
  margin-right: var(--space-inline-tight);
  color: var(--color-text-primary);
  font-size: var(--font-size-14);
  font-weight: var(--font-weight-600);
}

.gallery-sticky-label {
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
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
  border-color: var(--color-brand);
  box-shadow: 0 0 0 1px var(--color-brand);
}

.thumbnail-wrap {
  position: relative;
  width: 100%;
  height: 100%;
  background: var(--color-bg-muted);
}

.thumbnail {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.detail-panel {
  flex-shrink: 0;
  margin-bottom: 0;
}

.detail-preview {
  margin: 0 0 var(--space-block);
  border-radius: var(--radius);
  overflow: hidden;
  background: var(--color-bg-muted);
  border: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: var(--space-64);
}

.detail-action-alert {
  margin: var(--space-form-field) 0;
}

.gallery-participants-list {
  margin: 0;
  padding: 0;
  list-style: none;
}

.gallery-participants-item + .gallery-participants-item {
  margin-top: var(--space-inline-tight);
}

.gallery-participant-link {
  padding: 0;
  height: auto;
}

.detail-panel-action-btn {
  width: 100%;
  margin: var(--space-form-field) 0 var(--space-action-group);
}

.detail-panel-action-btn--last {
  margin-bottom: 0;
}

.join-error-alert {
  margin-bottom: var(--space-action-group);
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

.gallery-section-header {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: flex-end;
  padding: 0 var(--space-inline-tight);
  margin: 0;
  box-sizing: border-box;
}

.gallery-section-header--year {
  padding-top: var(--space-form-field);
  color: var(--color-text-primary);
  font-size: var(--font-size-14);
  font-weight: var(--font-weight-600);
  line-height: var(--line-height-tight);
}

.gallery-section-header--day {
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  font-weight: var(--font-weight-500);
  line-height: var(--line-height-tight);
}

.gallery-section-label {
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
</style>
