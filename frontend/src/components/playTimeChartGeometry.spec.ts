import { describe, it, expect } from "vitest";
import {
  clampCenteredTipX,
  clampedPlotSize,
  syncCanvasBuffer,
} from "./playTimeChartGeometry";

describe("clampedPlotSize", () => {
  const pad = { top: 36, left: 48, right: 14, bottom: 44 };

  it("subtracts padding for normal sizes", () => {
    expect(clampedPlotSize(300, 280, pad)).toEqual({
      plotW: 238,
      plotH: 200,
    });
  });

  it("clamps to zero when container is smaller than padding", () => {
    expect(clampedPlotSize(50, 70, pad)).toEqual({ plotW: 0, plotH: 0 });
  });
});

describe("clampCenteredTipX", () => {
  const padLeft = 48;
  const padRight = 14;
  const tipWidth = 120;

  it("keeps center x when tip fits inside the container", () => {
    expect(clampCenteredTipX(150, 300, tipWidth, padLeft, padRight)).toBe(150);
  });

  it("clamps near the left edge so tip stays inside", () => {
    const min = padLeft + tipWidth / 2;
    expect(clampCenteredTipX(10, 300, tipWidth, padLeft, padRight)).toBe(min);
  });

  it("clamps near the right edge so tip stays inside", () => {
    const max = 300 - padRight - tipWidth / 2;
    expect(clampCenteredTipX(290, 300, tipWidth, padLeft, padRight)).toBe(max);
  });
});

describe("syncCanvasBuffer", () => {
  it("assigns width/height only when the buffer size changes", () => {
    const canvas = { width: 600, height: 560 };
    const first = syncCanvasBuffer(canvas, 300, 280, 2);
    expect(first.resized).toBe(false);
    expect(canvas.width).toBe(600);

    const second = syncCanvasBuffer(canvas, 400, 280, 2);
    expect(second.resized).toBe(true);
    expect(canvas.width).toBe(800);
    expect(canvas.height).toBe(560);
  });
});
