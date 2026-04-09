import { describe, expect, it } from "vitest";
import { formatEncounteredAt } from "./formatEncounteredAt";

describe("formatEncounteredAt", () => {
  it("formats valid ISO string with explicit locale", () => {
    const s = formatEncounteredAt("2025-01-01T12:00:00.000Z", "ja-JP");
    expect(s).toMatch(/2025/);
    expect(s.length).toBeGreaterThan(4);
  });

  it("formats with runtime default locale when locale omitted", () => {
    const s = formatEncounteredAt("2025-01-01T12:00:00.000Z");
    expect(s).toMatch(/2025/);
    expect(s.length).toBeGreaterThan(4);
  });

  it("does not throw on empty input", () => {
    expect(() => formatEncounteredAt("")).not.toThrow();
  });
});
