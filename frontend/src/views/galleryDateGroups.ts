import type { ScreenshotDTO } from "../wails/app";

export const GALLERY_YEAR_HEADER_ROW_HEIGHT_PX = 32;
export const GALLERY_DAY_HEADER_ROW_HEIGHT_PX = 28;

/** Leap year used only as a calendar scaffold for month/day Intl formatting (year is omitted in labels). */
export const EXEMPLAR_LEAP_YEAR = 2024;

export type GalleryVirtualRow =
  | {
      type: "yearHeader";
      label: string;
      rowKey: string;
    }
  | {
      type: "dayHeader";
      label: string;
      rowKey: string;
      dayKey: string;
    }
  | { type: "grid"; items: ScreenshotDTO[]; rowKey: string };

function localDayParts(
  takenAt: string,
): { y: number; m: number; d: number } | null {
  const d = new Date(takenAt);
  if (Number.isNaN(d.getTime())) {
    return null;
  }
  return {
    y: d.getFullYear(),
    m: d.getMonth() + 1,
    d: d.getDate(),
  };
}

function dayKeyFromParts(y: number, m: number, d: number): string {
  return `${y}-${m}-${d}`;
}

/** Labels for gallery date group headers; default matches legacy Japanese formatting. */
export interface GalleryDateLabels {
  formatYear: (year: number) => string;
  formatDay: (month: number, day: number) => string;
  unknownDate: string;
}

export const galleryLabelsJapanese: GalleryDateLabels = {
  formatYear: (y) => `${y}年`,
  formatDay: (m, d) => `${m}月${d}日`,
  unknownDate: "日付不明",
};

export function galleryLabelsFromLocale(
  calendarLocale: string,
  unknownDate: string,
): GalleryDateLabels {
  const yearFmt = new Intl.DateTimeFormat(calendarLocale, { year: "numeric" });
  const dayFmt = new Intl.DateTimeFormat(calendarLocale, {
    month: "long",
    day: "numeric",
  });
  return {
    formatYear: (y) => yearFmt.format(new Date(y, 5, 15)),
    formatDay: (m, d) => dayFmt.format(new Date(EXEMPLAR_LEAP_YEAR, m - 1, d)),
    unknownDate,
  };
}

function compareDayKeysDesc(a: string, b: string): number {
  const pa = a.split("-").map(Number);
  const pb = b.split("-").map(Number);
  for (let i = 0; i < 3; i++) {
    const da = pa[i] ?? 0;
    const db = pb[i] ?? 0;
    if (da !== db) {
      return db - da;
    }
  }
  return 0;
}

/** Split list into per-calendar-day buckets (local time) and unknown-dated tail. */
export function partitionScreenshotsByLocalDay(list: ScreenshotDTO[]): {
  byDay: Map<string, ScreenshotDTO[]>;
  unknown: ScreenshotDTO[];
} {
  const byDay = new Map<string, ScreenshotDTO[]>();
  const unknown: ScreenshotDTO[] = [];

  for (const item of list) {
    const ta = item.takenAt;
    if (ta == null || ta === "") {
      unknown.push(item);
      continue;
    }
    const parts = localDayParts(ta);
    if (parts == null) {
      unknown.push(item);
      continue;
    }
    const k = dayKeyFromParts(parts.y, parts.m, parts.d);
    const arr = byDay.get(k);
    if (arr) {
      arr.push(item);
    } else {
      byDay.set(k, [item]);
    }
  }

  return { byDay, unknown };
}

function pushGridRows(
  rows: GalleryVirtualRow[],
  items: ScreenshotDTO[],
  cols: number,
  rowKeyPrefix: string,
): void {
  if (cols < 1 || items.length === 0) {
    return;
  }
  let chunk = 0;
  for (let i = 0; i < items.length; i += cols) {
    const slice = items.slice(i, i + cols);
    rows.push({
      type: "grid",
      items: slice,
      rowKey: `${rowKeyPrefix}:${chunk}`,
    });
    chunk++;
  }
}

/**
 * Build flat virtual rows for the gallery (newest-first days, same order as input within each day).
 * Year dividers appear only when the calendar year changes; day headers omit the year.
 */
export function buildGalleryVirtualRows(
  list: ScreenshotDTO[],
  cols: number,
  labels: GalleryDateLabels = galleryLabelsJapanese,
): GalleryVirtualRow[] {
  if (list.length === 0 || cols < 1) {
    return [];
  }

  const { byDay, unknown } = partitionScreenshotsByLocalDay(list);
  const sortedDayKeys = [...byDay.keys()].sort(compareDayKeysDesc);
  const rows: GalleryVirtualRow[] = [];
  let lastYear: number | null = null;

  for (const dayKey of sortedDayKeys) {
    const items = byDay.get(dayKey);
    if (!items?.length) {
      continue;
    }
    const [ys, ms, ds] = dayKey.split("-");
    const y = Number(ys);
    const m = Number(ms);
    const day = Number(ds);

    if (lastYear !== y) {
      rows.push({
        type: "yearHeader",
        label: labels.formatYear(y),
        rowKey: `hdr-y-${y}`,
      });
      lastYear = y;
    }

    rows.push({
      type: "dayHeader",
      label: labels.formatDay(m, day),
      rowKey: `hdr-d-${dayKey}`,
      dayKey,
    });
    pushGridRows(rows, items, cols, `grid-${dayKey}`);
  }

  if (unknown.length > 0) {
    rows.push({
      type: "yearHeader",
      label: labels.unknownDate,
      rowKey: "hdr-y-unknown",
    });
    pushGridRows(rows, unknown, cols, "grid-unknown");
  }

  return rows;
}

export function galleryRowHeight(
  row: GalleryVirtualRow,
  gridRowHeightPx: number,
): number {
  if (row.type === "grid") {
    return gridRowHeightPx;
  }
  if (row.type === "yearHeader") {
    return GALLERY_YEAR_HEADER_ROW_HEIGHT_PX;
  }
  if (row.type === "dayHeader") {
    return GALLERY_DAY_HEADER_ROW_HEIGHT_PX;
  }
  const _exhaustive: never = row;
  throw new Error(
    `galleryRowHeight: unknown row type ${JSON.stringify(_exhaustive)}`,
  );
}

export interface GalleryHeaderIndexEntry {
  index: number;
  rowKey: string;
  label: string;
  type: "yearHeader" | "dayHeader";
  dayKey?: string;
  yearLabel?: string;
  yearRowKey?: string;
}

/** Sorted header row indices for O(log n) sticky section lookup. */
export function buildGalleryHeaderIndices(
  rows: GalleryVirtualRow[],
): GalleryHeaderIndexEntry[] {
  const indices: GalleryHeaderIndexEntry[] = [];
  let currentYearLabel: string | undefined;
  let currentYearRowKey: string | undefined;
  for (let i = 0; i < rows.length; i++) {
    const row = rows[i];
    if (row?.type === "yearHeader") {
      currentYearLabel = row.label;
      currentYearRowKey = row.rowKey;
      indices.push({
        index: i,
        rowKey: row.rowKey,
        label: row.label,
        type: "yearHeader",
      });
    } else if (row?.type === "dayHeader") {
      indices.push({
        index: i,
        rowKey: row.rowKey,
        label: row.label,
        type: "dayHeader",
        dayKey: row.dayKey,
        yearLabel: currentYearLabel,
        yearRowKey: currentYearRowKey,
      });
    }
  }
  return indices;
}

export interface GalleryStickySection {
  label: string;
  rowKey: string;
  headerKind: "yearHeader" | "dayHeader";
  dayKey?: string;
  yearLabel?: string;
  yearRowKey?: string;
}

function toStickySection(entry: GalleryHeaderIndexEntry): GalleryStickySection {
  if (entry.type === "yearHeader") {
    return {
      label: entry.label,
      rowKey: entry.rowKey,
      headerKind: "yearHeader",
    };
  }
  return {
    label: entry.label,
    rowKey: entry.rowKey,
    headerKind: "dayHeader",
    dayKey: entry.dayKey,
    yearLabel: entry.yearLabel,
    yearRowKey: entry.yearRowKey,
  };
}

/** Resolve the sticky section for the first visible virtual row index. */
export function stickySectionForIndex(
  firstIdx: number,
  headerIndices: readonly GalleryHeaderIndexEntry[],
): GalleryStickySection | null {
  if (headerIndices.length === 0) {
    return null;
  }

  let lo = 0;
  let hi = headerIndices.length;
  while (lo < hi) {
    const mid = (lo + hi) >> 1;
    if (headerIndices[mid]!.index <= firstIdx) {
      lo = mid + 1;
    } else {
      hi = mid;
    }
  }
  const end = lo - 1;
  if (end < 0) {
    return null;
  }

  const nearest = headerIndices[end]!;
  return toStickySection(nearest);
}
