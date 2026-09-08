import { describe, expect, it, vi, afterEach } from "vitest";
import { formatVrcUserCacheDateTime } from "./formatVrcUserCacheDateTime";

describe("formatVrcUserCacheDateTime", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("returns null for empty or whitespace input", () => {
    expect(formatVrcUserCacheDateTime("")).toBeNull();
    expect(formatVrcUserCacheDateTime("   ")).toBeNull();
    expect(formatVrcUserCacheDateTime(undefined)).toBeNull();
    expect(formatVrcUserCacheDateTime(null)).toBeNull();
  });

  it("returns null for invalid datetime strings", () => {
    expect(formatVrcUserCacheDateTime("not-a-date")).toBeNull();
  });

  it("formats valid ISO strings with the provided locale", () => {
    const spy = vi.spyOn(Intl, "DateTimeFormat").mockImplementation(
      () =>
        ({
          format: () => "2025/1/1 9:00",
        }) as Intl.DateTimeFormat,
    );

    const result = formatVrcUserCacheDateTime(
      "2025-01-01T00:00:00.000Z",
      "ja-JP",
    );
    expect(result).toBe("2025/1/1 9:00");
    expect(spy).toHaveBeenCalledWith("ja-JP", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "numeric",
      minute: "2-digit",
    });
  });

  it("formats VRChat-style timestamps with fractional seconds", () => {
    const result = formatVrcUserCacheDateTime(
      "2026-03-27T08:10:57.917Z",
      "en-US",
    );
    expect(result).toMatch(/2026/);
    expect(result).toMatch(/27/);
    expect(result!.length).toBeGreaterThan(8);
  });

  it("returns the original string when Intl.DateTimeFormat throws", () => {
    vi.spyOn(Intl, "DateTimeFormat").mockImplementation(() => {
      throw new Error("locale failure");
    });

    expect(
      formatVrcUserCacheDateTime("2025-01-01T00:00:00.000Z", "xx-INVALID"),
    ).toBe("2025-01-01T00:00:00.000Z");
  });
});
