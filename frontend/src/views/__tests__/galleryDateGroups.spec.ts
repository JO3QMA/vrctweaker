import { describe, it, expect } from "vitest";
import type { ScreenshotDTO } from "../../wails/app";
import {
  buildGalleryVirtualRows,
  dayCollapseKey,
  galleryLabelsFromLocale,
  monthCollapseKey,
  partitionScreenshotsByLocalDay,
  yearCollapseKey,
} from "../galleryDateGroups";

function shot(id: string, takenAt?: string): ScreenshotDTO {
  return {
    id,
    filePath: `C:/x/${id}.png`,
    worldId: "wrld_x",
    worldName: "W",
    takenAt,
    fileSizeBytes: 1,
  };
}

describe("galleryLabelsFromLocale", () => {
  it("formats year, month, and day for a fixed locale", () => {
    const labels = galleryLabelsFromLocale("en-US", "Unknown date");
    expect(labels.formatYear(2024)).toMatch(/2024/);
    expect(labels.formatMonth(2024, 3)).toMatch(/2024/);
    expect(labels.formatMonth(2024, 3)).toMatch(/March/i);
    expect(labels.formatDay(2024, 3, 15)).toMatch(/2024/);
    expect(labels.formatDay(2024, 3, 15)).toMatch(/15/);
    expect(labels.unknownDate).toBe("Unknown date");
  });
});

describe("partitionScreenshotsByLocalDay", () => {
  it("puts items without takenAt into unknown", () => {
    const { byDay, unknown } = partitionScreenshotsByLocalDay([
      shot("a", "2024-06-01T12:00:00Z"),
      shot("b"),
    ]);
    expect(unknown.map((x) => x.id)).toEqual(["b"]);
    expect(byDay.size).toBe(1);
  });

  it("groups by local calendar day", () => {
    const list = [shot("a", "2024-06-01T12:00:00Z")];
    const { byDay } = partitionScreenshotsByLocalDay(list);
    const keys = [...byDay.keys()];
    expect(keys.length).toBe(1);
    expect(byDay.get(keys[0]!)?.map((x) => x.id)).toEqual(["a"]);
  });
});

describe("buildGalleryVirtualRows", () => {
  it("returns empty for empty list", () => {
    expect(buildGalleryVirtualRows([], 3, new Set())).toEqual([]);
  });

  it("emits year, month, day headers and one grid row for a single item", () => {
    const rows = buildGalleryVirtualRows(
      [shot("a", "2024-03-15T08:00:00Z")],
      3,
      new Set(),
    );
    expect(rows.map((r) => r.type)).toEqual([
      "yearHeader",
      "monthHeader",
      "dayHeader",
      "grid",
    ]);
    expect(rows[0]).toMatchObject({ type: "yearHeader", label: "2024年" });
    expect(rows[1]).toMatchObject({ type: "monthHeader", label: "2024年03月" });
    expect(rows[2]).toMatchObject({
      type: "dayHeader",
      label: "2024年03月15日",
    });
    const g = rows[3];
    expect(g?.type).toBe("grid");
    if (g?.type === "grid") {
      expect(g.items.map((x) => x.id)).toEqual(["a"]);
    }
  });

  it("chunks same-day items by column count", () => {
    const rows = buildGalleryVirtualRows(
      [
        shot("a", "2024-01-10T10:00:00Z"),
        shot("b", "2024-01-10T11:00:00Z"),
        shot("c", "2024-01-10T12:00:00Z"),
      ],
      2,
      new Set(),
    );
    const grids = rows.filter((r) => r.type === "grid");
    expect(grids.length).toBe(2);
    if (grids[0]?.type === "grid" && grids[1]?.type === "grid") {
      expect(grids[0].items.map((x) => x.id)).toEqual(["a", "b"]);
      expect(grids[1].items.map((x) => x.id)).toEqual(["c"]);
    }
  });

  it("orders multiple days descending within same month", () => {
    const rows = buildGalleryVirtualRows(
      [
        shot("newer", "2024-05-20T10:00:00Z"),
        shot("older", "2024-05-05T10:00:00Z"),
      ],
      3,
      new Set(),
    );
    const dayLabels = rows
      .filter((r) => r.type === "dayHeader")
      .map((r) => (r.type === "dayHeader" ? r.label : ""));
    expect(dayLabels).toEqual(["2024年05月20日", "2024年05月05日"]);
  });

  it("places unknown-dated section at end after dated groups", () => {
    const rows = buildGalleryVirtualRows(
      [shot("u", undefined), shot("d", "2023-12-01T00:00:00Z")],
      2,
      new Set(),
    );
    const years = rows
      .filter((r) => r.type === "yearHeader")
      .map((r) => (r.type === "yearHeader" ? r.label : ""));
    expect(years[0]).toBe("2023年");
    expect(years[years.length - 1]).toBe("日付不明");
    const lastGrids = rows.filter((r) => r.type === "grid");
    const last = lastGrids[lastGrids.length - 1];
    expect(last?.type).toBe("grid");
    if (last?.type === "grid") {
      expect(last.items.map((x) => x.id)).toEqual(["u"]);
    }
  });

  it("omits grids under collapsed year", () => {
    const collapsed = new Set([yearCollapseKey(2024)]);
    const rows = buildGalleryVirtualRows(
      [shot("a", "2024-01-01T00:00:00Z")],
      2,
      collapsed,
    );
    expect(rows.some((r) => r.type === "grid")).toBe(false);
    expect(rows[0]).toMatchObject({ type: "yearHeader", expanded: false });
  });

  it("omits grids under collapsed month", () => {
    const collapsed = new Set([monthCollapseKey(2024, 6)]);
    const rows = buildGalleryVirtualRows(
      [shot("a", "2024-06-15T00:00:00Z")],
      2,
      collapsed,
    );
    expect(rows.some((r) => r.type === "grid")).toBe(false);
    const mh = rows.find((r) => r.type === "monthHeader");
    expect(mh).toMatchObject({ expanded: false });
  });

  it("omits grids under collapsed day", () => {
    const collapsed = new Set([dayCollapseKey(2024, 6, 10)]);
    const rows = buildGalleryVirtualRows(
      [shot("a", "2024-06-10T00:00:00Z")],
      2,
      collapsed,
    );
    expect(rows.some((r) => r.type === "grid")).toBe(false);
    const dh = rows.find((r) => r.type === "dayHeader");
    expect(dh).toMatchObject({ expanded: false });
  });
});
