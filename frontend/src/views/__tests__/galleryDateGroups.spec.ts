import { describe, it, expect } from "vitest";
import type { ScreenshotDTO } from "../../wails/app";
import {
  buildGalleryHeaderIndices,
  buildGalleryVirtualRows,
  EXEMPLAR_LEAP_YEAR,
  galleryLabelsFromLocale,
  partitionScreenshotsByLocalDay,
  stickySectionForIndex,
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
  it("formats year and compact day for a fixed locale", () => {
    const labels = galleryLabelsFromLocale("en-US", "Unknown date");
    expect(labels.formatYear(2024)).toMatch(/2024/);
    expect(labels.formatDay(3, 15)).toMatch(/March/i);
    expect(labels.formatDay(3, 15)).toMatch(/15/);
    expect(labels.formatDay(3, 15)).not.toMatch(/2024/);
    expect(labels.unknownDate).toBe("Unknown date");
  });

  it("formats Feb 29 using the exemplar leap year scaffold", () => {
    const labels = galleryLabelsFromLocale("en-US", "Unknown date");
    expect(labels.formatDay(2, 29)).toMatch(/February/i);
    expect(labels.formatDay(2, 29)).toMatch(/29/);
    expect(EXEMPLAR_LEAP_YEAR % 4).toBe(0);
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
    expect(buildGalleryVirtualRows([], 3)).toEqual([]);
  });

  it("emits year and day headers and one grid row for a single item", () => {
    const rows = buildGalleryVirtualRows(
      [shot("a", "2024-03-15T08:00:00Z")],
      3,
    );
    expect(rows.map((r) => r.type)).toEqual([
      "yearHeader",
      "dayHeader",
      "grid",
    ]);
    expect(rows[0]).toMatchObject({ type: "yearHeader", label: "2024年" });
    expect(rows[1]).toMatchObject({ type: "dayHeader", label: "3月15日" });
    const g = rows[2];
    expect(g?.type).toBe("grid");
    if (g?.type === "grid") {
      expect(g.items.map((x) => x.id)).toEqual(["a"]);
    }
  });

  it("does not repeat year header within the same calendar year", () => {
    const rows = buildGalleryVirtualRows(
      [shot("a", "2024-05-20T10:00:00Z"), shot("b", "2024-05-05T10:00:00Z")],
      3,
    );
    const yearHeaders = rows.filter((r) => r.type === "yearHeader");
    expect(yearHeaders).toHaveLength(1);
    expect(yearHeaders[0]).toMatchObject({ label: "2024年" });
  });

  it("chunks same-day items by column count", () => {
    const rows = buildGalleryVirtualRows(
      [
        shot("a", "2024-01-10T10:00:00Z"),
        shot("b", "2024-01-10T11:00:00Z"),
        shot("c", "2024-01-10T12:00:00Z"),
      ],
      2,
    );
    const grids = rows.filter((r) => r.type === "grid");
    expect(grids.length).toBe(2);
    if (grids[0]?.type === "grid" && grids[1]?.type === "grid") {
      expect(grids[0].items.map((x) => x.id)).toEqual(["a", "b"]);
      expect(grids[1].items.map((x) => x.id)).toEqual(["c"]);
    }
  });

  it("orders multiple days descending within same year", () => {
    const rows = buildGalleryVirtualRows(
      [
        shot("newer", "2024-05-20T10:00:00Z"),
        shot("older", "2024-05-05T10:00:00Z"),
      ],
      3,
    );
    const dayLabels = rows
      .filter((r) => r.type === "dayHeader")
      .map((r) => (r.type === "dayHeader" ? r.label : ""));
    expect(dayLabels).toEqual(["5月20日", "5月5日"]);
  });

  it("inserts a new year header when the calendar year changes", () => {
    const rows = buildGalleryVirtualRows(
      [
        shot("newer", "2025-01-02T10:00:00Z"),
        shot("older", "2024-12-31T10:00:00Z"),
      ],
      3,
    );
    const yearLabels = rows
      .filter((r) => r.type === "yearHeader")
      .map((r) => (r.type === "yearHeader" ? r.label : ""));
    expect(yearLabels).toEqual(["2025年", "2024年"]);
  });

  it("places unknown-dated section at end after dated groups", () => {
    const rows = buildGalleryVirtualRows(
      [shot("u", undefined), shot("d", "2023-12-01T00:00:00Z")],
      2,
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
});

describe("buildGalleryHeaderIndices", () => {
  it("lists header rows in ascending index order", () => {
    const rows = buildGalleryVirtualRows(
      [shot("a", "2025-03-01T10:00:00Z"), shot("b", "2024-03-01T10:00:00Z")],
      2,
    );
    const indices = buildGalleryHeaderIndices(rows);
    expect(indices.map((e) => e.type)).toEqual([
      "yearHeader",
      "dayHeader",
      "yearHeader",
      "dayHeader",
    ]);
    expect(indices[1]).toMatchObject({
      type: "dayHeader",
      dayKey: expect.stringMatching(/^\d+-\d+-\d+$/),
      rowKey: expect.stringMatching(/^hdr-d-/),
    });
  });
});

describe("stickySectionForIndex", () => {
  it("resolves day sections by rowKey even when labels collide across years", () => {
    const rows = buildGalleryVirtualRows(
      [shot("a", "2024-09-04T10:00:00Z"), shot("b", "2023-09-04T10:00:00Z")],
      2,
    );
    const indices = buildGalleryHeaderIndices(rows);
    const dayHeaders = indices.filter((e) => e.type === "dayHeader");
    expect(dayHeaders[0]?.label).toBe(dayHeaders[1]?.label);
    expect(dayHeaders[0]?.dayKey).not.toBe(dayHeaders[1]?.dayKey);

    const newerDayIdx = rows.findIndex(
      (r) => r.type === "dayHeader" && r.dayKey === dayHeaders[0]?.dayKey,
    );
    const section = stickySectionForIndex(newerDayIdx, indices);
    expect(section?.rowKey).toBe(dayHeaders[0]?.rowKey);
    expect(section?.dayKey).toBe(dayHeaders[0]?.dayKey);
    expect(section?.headerKind).toBe("dayHeader");
    expect(section?.yearLabel).toBe("2024年");
  });

  it("returns year header immediately when firstIdx is on a year divider", () => {
    const rows = buildGalleryVirtualRows(
      [shot("a", "2024-01-02T10:00:00Z"), shot("b", "2023-12-31T10:00:00Z")],
      2,
    );
    const indices = buildGalleryHeaderIndices(rows);
    const year2024Idx = rows.findIndex(
      (r) => r.type === "yearHeader" && r.label === "2024年",
    );
    const section = stickySectionForIndex(year2024Idx, indices);
    expect(section?.headerKind).toBe("yearHeader");
    expect(section?.label).toBe("2024年");
    expect(section?.dayKey).toBeUndefined();
  });

  it("returns unknown year header for unknown-dated grid rows", () => {
    const rows = buildGalleryVirtualRows([shot("u", undefined)], 2);
    const indices = buildGalleryHeaderIndices(rows);
    const gridIdx = rows.findIndex((r) => r.type === "grid");
    const section = stickySectionForIndex(gridIdx, indices);
    expect(section?.headerKind).toBe("yearHeader");
    expect(section?.label).toBe("日付不明");
  });

  it("includes yearLabel on day sections for sticky year context", () => {
    const rows = buildGalleryVirtualRows(
      [shot("a", "2024-09-04T10:00:00Z")],
      2,
    );
    const indices = buildGalleryHeaderIndices(rows);
    const gridIdx = rows.findIndex((r) => r.type === "grid");
    const section = stickySectionForIndex(gridIdx, indices);
    expect(section?.headerKind).toBe("dayHeader");
    expect(section?.label).toBe("9月4日");
    expect(section?.yearLabel).toBe("2024年");
    expect(section?.yearRowKey).toBe("hdr-y-2024");
  });
});
