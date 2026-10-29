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

  it("shows every day when spacing is wide enough", () => {
    const plotWidth = 700;
    const indices = playTimeChartVisibleXLabelIndices({
      pointCount: n,
      plotLeft,
      plotWidth,
      labelWidths: widths,
    });
    expect(indices).toEqual(Array.from({ length: n }, (_, i) => i));
  });

  it("thins labels when cramped but keeps first and last", () => {
    const plotWidth = 200;
    const indices = playTimeChartVisibleXLabelIndices({
      pointCount: n,
      plotLeft,
      plotWidth,
      labelWidths: widths,
    });
    expect(indices[0]).toBe(0);
    expect(indices[indices.length - 1]).toBe(n - 1);
    expect(indices.length).toBeLessThan(n);
  });

  it("drops the neighbor before the last label when they would overlap", () => {
    const plotWidth = 120;
    const indices = playTimeChartVisibleXLabelIndices({
      pointCount: n,
      plotLeft,
      plotWidth,
      labelWidths: widths,
      minGap: PLAY_TIME_X_LABEL_MIN_GAP,
    });
    expect(indices).toContain(0);
    expect(indices).toContain(n - 1);
    if (indices.includes(n - 2) && indices.includes(n - 1)) {
      const pos = (i: number) => plotLeft + (i / (n - 1)) * plotWidth;
      const need = widths[0]! + PLAY_TIME_X_LABEL_MIN_GAP;
      expect(Math.abs(pos(n - 2) - pos(n - 1))).toBeGreaterThanOrEqual(need);
    }
  });
});
