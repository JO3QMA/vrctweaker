import { describe, expect, it } from "vitest";
import { formatEncounteredAt } from "./formatEncounteredAt";

describe("formatEncounteredAt", () => {
  it("formats valid ISO string with ja-JP locale", () => {
    const s = formatEncounteredAt("2025-01-01T12:00:00.000Z");
    expect(s).toMatch(/2025/);
    expect(s.length).toBeGreaterThan(4);
  });

  it("does not throw on empty input", () => {
    expect(() => formatEncounteredAt("")).not.toThrow();
  });
});
