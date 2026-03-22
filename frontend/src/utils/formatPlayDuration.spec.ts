import { describe, it, expect } from "vitest";
import { formatPlayDurationHMS } from "./formatPlayDuration";

describe("formatPlayDurationHMS", () => {
  it("formats seconds only", () => {
    expect(formatPlayDurationHMS(58)).toBe("00時間00分58秒");
  });

  it("formats minutes and seconds", () => {
    expect(formatPlayDurationHMS(1428)).toBe("00時間23分48秒");
  });

  it("formats hours from user examples", () => {
    expect(formatPlayDurationHMS(45645)).toBe("12時間40分45秒");
    expect(formatPlayDurationHMS(83512)).toBe("23時間11分52秒");
    expect(formatPlayDurationHMS(18904)).toBe("05時間15分04秒");
    expect(formatPlayDurationHMS(2385)).toBe("00時間39分45秒");
  });

  it("clamps negative to zero", () => {
    expect(formatPlayDurationHMS(-1)).toBe("00時間00分00秒");
  });

  it("floors fractional seconds", () => {
    expect(formatPlayDurationHMS(58.9)).toBe("00時間00分58秒");
  });
});
