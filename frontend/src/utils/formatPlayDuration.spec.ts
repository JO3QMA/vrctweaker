import { describe, it, expect } from "vitest";
import { formatPlayDurationHMS } from "./formatPlayDuration";
import { createAppI18n } from "../i18n";

describe("formatPlayDurationHMS", () => {
  const jaT = createAppI18n("ja").global.t as (
    key: string,
    values?: Record<string, unknown>,
  ) => string;

  it("formats seconds only", () => {
    expect(formatPlayDurationHMS(58, jaT)).toBe("00時間00分58秒");
  });

  it("formats minutes and seconds", () => {
    expect(formatPlayDurationHMS(1428, jaT)).toBe("00時間23分48秒");
  });

  it("formats hours from user examples", () => {
    expect(formatPlayDurationHMS(45645, jaT)).toBe("12時間40分45秒");
    expect(formatPlayDurationHMS(83512, jaT)).toBe("23時間11分52秒");
    expect(formatPlayDurationHMS(18904, jaT)).toBe("05時間15分04秒");
    expect(formatPlayDurationHMS(2385, jaT)).toBe("00時間39分45秒");
  });

  it("clamps negative to zero", () => {
    expect(formatPlayDurationHMS(-1, jaT)).toBe("00時間00分00秒");
  });

  it("floors fractional seconds", () => {
    expect(formatPlayDurationHMS(58.9, jaT)).toBe("00時間00分58秒");
  });

  it("uses spaced join for English", () => {
    const enT = createAppI18n("en").global.t as (
      key: string,
      values?: Record<string, unknown>,
    ) => string;
    expect(formatPlayDurationHMS(58, enT)).toBe("00 h 00 min 58 s");
  });
});
