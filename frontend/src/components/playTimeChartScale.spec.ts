import { describe, expect, it } from "vitest";
import {
  PLAY_TIME_ONE_HOUR_SEC,
  PLAY_TIME_TEN_MINUTES_SEC,
  formatPlayTimeAxisTickLabel,
  playTimeChartMaxY,
  playTimeChartYAxisTicks,
} from "./playTimeChartScale";

describe("playTimeChartMaxY", () => {
  it("uses 10-minute ceiling below one hour", () => {
    expect(playTimeChartMaxY(0)).toBe(PLAY_TIME_TEN_MINUTES_SEC);
    expect(playTimeChartMaxY(1)).toBe(PLAY_TIME_TEN_MINUTES_SEC);
    expect(playTimeChartMaxY(600)).toBe(PLAY_TIME_TEN_MINUTES_SEC);
    expect(playTimeChartMaxY(601)).toBe(2 * PLAY_TIME_TEN_MINUTES_SEC);
    expect(playTimeChartMaxY(3540)).toBe(PLAY_TIME_ONE_HOUR_SEC);
  });

  it("uses whole-hour ceiling from one hour up to 24 hours", () => {
    expect(playTimeChartMaxY(3600)).toBe(PLAY_TIME_ONE_HOUR_SEC);
    expect(playTimeChartMaxY(3601)).toBe(2 * PLAY_TIME_ONE_HOUR_SEC);
    expect(playTimeChartMaxY(9000)).toBe(3 * PLAY_TIME_ONE_HOUR_SEC);
    expect(playTimeChartMaxY(24 * 3600)).toBe(24 * PLAY_TIME_ONE_HOUR_SEC);
    expect(playTimeChartMaxY(25 * 3600)).toBe(24 * PLAY_TIME_ONE_HOUR_SEC);
  });
});

describe("playTimeChartYAxisTicks", () => {
  it("steps by 10 minutes when max is within one hour", () => {
    expect(playTimeChartYAxisTicks(30 * 60)).toEqual([0, 600, 1200, 1800]);
  });

  it("steps by 30 minutes between one and four hours", () => {
    const ticks = playTimeChartYAxisTicks(2 * 3600);
    expect(ticks).toEqual([0, 1800, 3600, 5400, 7200]);
  });

  it("steps by whole hours when max is four hours or more", () => {
    const ticks = playTimeChartYAxisTicks(4 * 3600);
    expect(ticks[0]).toBe(0);
    expect(ticks[ticks.length - 1]).toBe(4 * 3600);
    for (const v of ticks) {
      expect(v % 3600).toBe(0);
    }
  });

  it("ends on maxY for 24-hour scale", () => {
    const ticks = playTimeChartYAxisTicks(24 * 3600);
    expect(ticks[ticks.length - 1]).toBe(24 * 3600);
    for (const v of ticks) {
      expect(v % 3600).toBe(0);
    }
  });
});

describe("formatPlayTimeAxisTickLabel", () => {
  const units = { hour: "時間", minute: "分", second: "秒" };

  it("labels exact minute grid values without rounding", () => {
    expect(formatPlayTimeAxisTickLabel(0, units)).toBe("0分");
    expect(formatPlayTimeAxisTickLabel(600, units)).toBe("10分");
    expect(formatPlayTimeAxisTickLabel(3600, units)).toBe("1時間");
    expect(formatPlayTimeAxisTickLabel(7200, units)).toBe("2時間");
  });
});
