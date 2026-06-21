import { describe, expect, it, vi, afterEach } from "vitest";
import { formatEncounteredAt } from "./formatEncounteredAt";

describe("formatEncounteredAt", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("formats valid ISO string with ja-JP locale", () => {
    const s = formatEncounteredAt("2025-01-01T12:00:00.000Z");
    expect(s).toMatch(/2025/);
    expect(s.length).toBeGreaterThan(4);
  });

  it("uses the provided calendar locale", () => {
    const spy = vi
      .spyOn(Date.prototype, "toLocaleString")
      .mockReturnValue("localized");

    expect(formatEncounteredAt("2025-06-01T00:00:00.000Z", "en-US")).toBe(
      "localized",
    );
    expect(spy).toHaveBeenCalledWith("en-US");
  });

  it("does not throw on empty input", () => {
    expect(() => formatEncounteredAt("")).not.toThrow();
  });

  it("returns the original ISO string when formatting throws", () => {
    vi.spyOn(Date.prototype, "toLocaleString").mockImplementation(() => {
      throw new Error("locale failure");
    });

    const iso = "2025-01-01T12:00:00.000Z";
    expect(formatEncounteredAt(iso)).toBe(iso);
  });
});
