import { describe, expect, it } from "vitest";
import {
  PLAY_TIME_X_LABEL_MIN_GAP,
  playTimeChartVisibleXLabelIndices,
  playTimeChartXLabelHorizontalPad,
} from "./playTimeChartXLabels";

describe("playTimeChartXLabelHorizontalPad", () => {
  it("uses half label width plus edge margin", () => {
    expect(playTimeChartXLabelHorizontalPad(40, 4)).toBe(24);
  });
});

describe("playTimeChartVisibleXLabelIndices", () => {
  const n = 14;
  const widths = Array.from({ length: n }, () => 36);
  const plotLeft = 48;

  it("shows every day when spacing is wide enough (8px min gap)", () => {
    const plotWidth = 720;
    const indices = playTimeChartVisibleXLabelIndices({
      pointCount: n,
      plotLeft,
      plotWidth,
      labelWidths: widths,
    });
    expect(indices).toEqual(Array.from({ length: n }, (_, i) => i));
  });

  it("uses stride 2 (every other day) before falling back to first and last only", () => {
    const plotWidth = 480;
    const indices = playTimeChartVisibleXLabelIndices({
      pointCount: n,
      plotLeft,
      plotWidth,
      labelWidths: widths,
      minGap: PLAY_TIME_X_LABEL_MIN_GAP,
    });
    expect(indices[0]).toBe(0);
    expect(indices[indices.length - 1]).toBe(n - 1);
    expect(indices.length).toBeGreaterThan(2);
    for (const i of indices) {
      if (i !== 0 && i !== n - 1) {
        expect(i % 2).toBe(0);
      }
    }
    expect(indices).toEqual([0, 2, 4, 6, 8, 10, 13]);
  });

  it("drops only the neighbor before the last when stride 2 still collides at the end", () => {
    const plotWidth = 200;
    const indices = playTimeChartVisibleXLabelIndices({
      pointCount: n,
      plotLeft,
      plotWidth,
      labelWidths: widths,
    });
    expect(indices).toContain(0);
    expect(indices).toContain(n - 1);
    expect(indices).not.toContain(n - 2);
  });

  it("keeps only the last label when first and last would overlap", () => {
    const indices = playTimeChartVisibleXLabelIndices({
      pointCount: n,
      plotLeft,
      plotWidth: 0,
      labelWidths: widths,
    });
    expect(indices).toEqual([n - 1]);
  });

  it("drops the first label when only first and last remain and they overlap", () => {
    const plotWidth = 20;
    const indices = playTimeChartVisibleXLabelIndices({
      pointCount: n,
      plotLeft,
      plotWidth,
      labelWidths: widths,
      minGap: PLAY_TIME_X_LABEL_MIN_GAP,
    });
    expect(indices).toEqual([n - 1]);
  });
});
