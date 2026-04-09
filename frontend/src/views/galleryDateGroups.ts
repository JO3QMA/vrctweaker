import type { ScreenshotDTO } from "../wails/app";

export const GALLERY_HEADER_ROW_HEIGHT_PX = 42;

export type GalleryVirtualRow =
  | {
      type: "yearHeader";
      collapseKey: string;
      label: string;
      rowKey: string;
      expanded: boolean;
    }
  | {
      type: "monthHeader";
      collapseKey: string;
      label: string;
      rowKey: string;
      expanded: boolean;
    }
  | {
      type: "dayHeader";
      collapseKey: string;
      label: string;
      rowKey: string;
      expanded: boolean;
    }
  | { type: "grid"; items: ScreenshotDTO[]; rowKey: string };

export function yearCollapseKey(year: number | "unknown"): string {
  return year === "unknown" ? "y:unknown" : `y:${year}`;
}

export function monthCollapseKey(year: number, month: number): string {
  return `m:${year}-${month}`;
}

export function dayCollapseKey(
  year: number,
  month: number,
  day: number,
): string {
  return `d:${year}-${month}-${day}`;
}

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

function pad2(n: number): string {
  return String(n).padStart(2, "0");
}

/** Labels for gallery date group headers; default matches legacy Japanese formatting. */
export interface GalleryDateLabels {
  formatYear: (year: number) => string;
  formatMonth: (year: number, month: number) => string;
  formatDay: (year: number, month: number, day: number) => string;
  unknownDate: string;
}

export const galleryLabelsJapanese: GalleryDateLabels = {
  formatYear: (y) => `${y}年`,
  formatMonth: (y, m) => `${y}年${pad2(m)}月`,
  formatDay: (y, m, d) => `${y}年${pad2(m)}月${pad2(d)}日`,
  unknownDate: "日付不明",
};

export function galleryLabelsFromLocale(
  calendarLocale: string,
  unknownDate: string,
): GalleryDateLabels {
  const yearFmt = new Intl.DateTimeFormat(calendarLocale, { year: "numeric" });
  const monthFmt = new Intl.DateTimeFormat(calendarLocale, {
    year: "numeric",
    month: "long",
  });
  const dayFmt = new Intl.DateTimeFormat(calendarLocale, {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
  return {
    formatYear: (y) => yearFmt.format(new Date(y, 5, 15)),
    formatMonth: (y, m) => monthFmt.format(new Date(y, m - 1, 1)),
    formatDay: (y, m, d) => dayFmt.format(new Date(y, m - 1, d)),
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

type YearTree = Map<
  number,
  Map<number, Array<{ day: number; items: ScreenshotDTO[]; dayKey: string }>>
>;

function buildYearTree(
  sortedDayKeys: string[],
  byDay: Map<string, ScreenshotDTO[]>,
): YearTree {
  const tree: YearTree = new Map();
  for (const dayKey of sortedDayKeys) {
    const items = byDay.get(dayKey);
    if (!items?.length) {
      continue;
    }
    const [ys, ms, ds] = dayKey.split("-");
    const y = Number(ys);
    const m = Number(ms);
    const day = Number(ds);
    let months = tree.get(y);
    if (!months) {
      months = new Map();
      tree.set(y, months);
    }
    let days = months.get(m);
    if (!days) {
      days = [];
      months.set(m, days);
    }
    days.push({ day, items, dayKey });
  }
  return tree;
}

function sortedYearsDesc(tree: YearTree): number[] {
  return [...tree.keys()].sort((a, b) => b - a);
}

function sortedMonthsDesc(months: Map<number, unknown>): number[] {
  return [...months.keys()].sort((a, b) => b - a);
}

function sortedDayEntries(
  days: Array<{ day: number; items: ScreenshotDTO[]; dayKey: string }>,
): Array<{ day: number; items: ScreenshotDTO[]; dayKey: string }> {
  return [...days].sort((a, b) => b.day - a.day);
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
 * @param collapsed — keys from yearCollapseKey / monthCollapseKey / dayCollapseKey; when present, that node's children are omitted.
 */
export function buildGalleryVirtualRows(
  list: ScreenshotDTO[],
  cols: number,
  collapsed: ReadonlySet<string>,
  labels: GalleryDateLabels = galleryLabelsJapanese,
): GalleryVirtualRow[] {
  if (list.length === 0 || cols < 1) {
    return [];
  }

  const { byDay, unknown } = partitionScreenshotsByLocalDay(list);
  const sortedDayKeys = [...byDay.keys()].sort(compareDayKeysDesc);
  const tree = buildYearTree(sortedDayKeys, byDay);
  const rows: GalleryVirtualRow[] = [];

  for (const y of sortedYearsDesc(tree)) {
    const yKey = yearCollapseKey(y);
    const yCollapsed = collapsed.has(yKey);
    rows.push({
      type: "yearHeader",
      collapseKey: yKey,
      label: labels.formatYear(y),
      rowKey: `hdr-y-${y}`,
      expanded: !yCollapsed,
    });
    if (yCollapsed) {
      continue;
    }

    const months = tree.get(y);
    if (!months) {
      continue;
    }

    for (const m of sortedMonthsDesc(months)) {
      const mKey = monthCollapseKey(y, m);
      const mCollapsed = collapsed.has(mKey);
      rows.push({
        type: "monthHeader",
        collapseKey: mKey,
        label: labels.formatMonth(y, m),
        rowKey: `hdr-m-${y}-${m}`,
        expanded: !mCollapsed,
      });
      if (mCollapsed) {
        continue;
      }

      const days = months.get(m);
      if (!days?.length) {
        continue;
      }

      for (const { day, items, dayKey } of sortedDayEntries(days)) {
        const dKey = dayCollapseKey(y, m, day);
        const dCollapsed = collapsed.has(dKey);
        rows.push({
          type: "dayHeader",
          collapseKey: dKey,
          label: labels.formatDay(y, m, day),
          rowKey: `hdr-d-${dayKey}`,
          expanded: !dCollapsed,
        });
        if (dCollapsed) {
          continue;
        }
        pushGridRows(rows, items, cols, `grid-${dayKey}`);
      }
    }
  }

  if (unknown.length > 0) {
    const uk = yearCollapseKey("unknown");
    const uCollapsed = collapsed.has(uk);
    rows.push({
      type: "yearHeader",
      collapseKey: uk,
      label: labels.unknownDate,
      rowKey: "hdr-y-unknown",
      expanded: !uCollapsed,
    });
    if (!uCollapsed) {
      pushGridRows(rows, unknown, cols, "grid-unknown");
    }
  }

  return rows;
}

export function galleryRowHeight(
  row: GalleryVirtualRow,
  gridRowHeightPx: number,
): number {
  return row.type === "grid" ? gridRowHeightPx : GALLERY_HEADER_ROW_HEIGHT_PX;
}
