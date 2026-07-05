import type { ScreenshotSearchDTO } from "../wails/app";

/** Gallery date range from el-date-picker (value-format YYYY-MM-DD). */
export type GalleryDateRangeFilter = [string, string];

export function resolveWorldSearchInput(
  input: string,
): Pick<ScreenshotSearchDTO, "worldId" | "worldName"> {
  const trimmed = input.trim();
  if (!trimmed) {
    return {};
  }
  if (trimmed.toLowerCase().startsWith("wrld_")) {
    return { worldId: trimmed };
  }
  return { worldName: trimmed };
}

export function buildGallerySearchFilter(
  worldInput: string,
  dateRange: GalleryDateRangeFilter | null | undefined,
): ScreenshotSearchDTO | null {
  const filter: ScreenshotSearchDTO = {
    ...resolveWorldSearchInput(worldInput),
  };
  if (dateRange && dateRange.length === 2) {
    const [from, to] = dateRange;
    if (from) {
      filter.dateFrom = from;
    }
    if (to) {
      filter.dateTo = to;
    }
  }
  if (filter.worldId || filter.worldName || filter.dateFrom || filter.dateTo) {
    return filter;
  }
  return null;
}
