import { describe, expect, it } from "vitest";
import {
  PLAY_TIME_ONE_HOUR_SEC,
  PLAY_TIME_TEN_MINUTES_SEC,
  formatPlayTimeAxisTickLabel,
  playTimeChartAxisTickUnit,
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

  it("snaps 11-hour data max to 12 hours for uniform tick steps", () => {
    expect(playTimeChartMaxY(11 * PLAY_TIME_ONE_HOUR_SEC)).toBe(
      12 * PLAY_TIME_ONE_HOUR_SEC,
    );
  });

  it("keeps 51–60 minute caps on the minute grid at exactly one hour", () => {
    expect(playTimeChartMaxY(55 * 60)).toBe(PLAY_TIME_ONE_HOUR_SEC);
    expect(playTimeChartMaxY(60 * 60)).toBe(PLAY_TIME_ONE_HOUR_SEC);
  });

  it("switches to the hour scale above 60 minutes", () => {
    expect(playTimeChartMaxY(61 * 60)).toBe(2 * PLAY_TIME_ONE_HOUR_SEC);
  });
});

describe("playTimeChartYAxisTicks", () => {
  it("steps by 10 minutes when max is within one hour", () => {
    expect(playTimeChartYAxisTicks(30 * 60)).toEqual([0, 600, 1200, 1800]);
  });

  it("steps by whole hours between one and four hours", () => {
    const ticks = playTimeChartYAxisTicks(2 * 3600);
    expect(ticks).toEqual([0, 3600, 7200]);
  });

  it("steps by whole hours when max is four hours or more", () => {
    const ticks = playTimeChartYAxisTicks(4 * 3600);
    expect(ticks[0]).toBe(0);
    expect(ticks[ticks.length - 1]).toBe(4 * 3600);
    for (const v of ticks) {
      expect(v % 3600).toBe(0);
    }
  });

  it("ends on snapped max for 24-hour scale", () => {
    const maxY = playTimeChartMaxY(24 * 3600);
    const ticks = playTimeChartYAxisTicks(maxY);
    expect(ticks[ticks.length - 1]).toBe(24 * 3600);
    for (const v of ticks) {
      expect(v % 3600).toBe(0);
    }
  });

  it("uses uniform two-hour steps through 12 hours for 11-hour axis max", () => {
    const maxY = playTimeChartMaxY(11 * PLAY_TIME_ONE_HOUR_SEC);
    expect(maxY).toBe(12 * PLAY_TIME_ONE_HOUR_SEC);
    expect(playTimeChartYAxisTicks(maxY)).toEqual([
      0,
      2 * 3600,
      4 * 3600,
      6 * 3600,
      8 * 3600,
      10 * 3600,
      12 * 3600,
    ]);
  });
});

describe("playTimeChartAxisTickUnit", () => {
  it("uses minutes when axis max is at most one hour", () => {
    expect(playTimeChartAxisTickUnit(30 * 60)).toBe("minute");
    expect(playTimeChartAxisTickUnit(PLAY_TIME_ONE_HOUR_SEC)).toBe("minute");
  });

  it("uses hours when axis max is above one hour", () => {
    expect(playTimeChartAxisTickUnit(PLAY_TIME_ONE_HOUR_SEC + 1)).toBe("hour");
    expect(playTimeChartAxisTickUnit(2 * PLAY_TIME_ONE_HOUR_SEC)).toBe("hour");
    expect(playTimeChartAxisTickUnit(10 * PLAY_TIME_ONE_HOUR_SEC)).toBe("hour");
  });
});

describe("formatPlayTimeAxisTickLabel", () => {
  const units = { hour: "時間", minute: "分" };

  it("uses minute unit for all ticks when axis is in minutes", () => {
    expect(formatPlayTimeAxisTickLabel(0, units, "minute")).toBe("0分");
    expect(formatPlayTimeAxisTickLabel(600, units, "minute")).toBe("10分");
    expect(formatPlayTimeAxisTickLabel(1800, units, "minute")).toBe("30分");
    expect(formatPlayTimeAxisTickLabel(3600, units, "minute")).toBe("60分");
  });

  it("uses whole-hour labels when axis is in hours", () => {
    expect(formatPlayTimeAxisTickLabel(0, units, "hour")).toBe("0時間");
    expect(formatPlayTimeAxisTickLabel(7200, units, "hour")).toBe("2時間");
    expect(formatPlayTimeAxisTickLabel(3600, units, "hour")).toBe("1時間");
  });

  it("labels 55–60 minute axis max in minutes without fractional hours", () => {
    const maxY = playTimeChartMaxY(55 * 60);
    const unit = playTimeChartAxisTickUnit(maxY);
    expect(unit).toBe("minute");
    for (const sec of playTimeChartYAxisTicks(maxY)) {
      const label = formatPlayTimeAxisTickLabel(sec, units, unit);
      expect(label).not.toMatch(/\./);
      expect(label.endsWith("分")).toBe(true);
    }
  });
});
