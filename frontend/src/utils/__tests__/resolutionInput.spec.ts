import { describe, expect, it } from "vitest";
import { normalizeResolutionDimension } from "../resolutionInput";

describe("normalizeResolutionDimension", () => {
  it("rounds fractional values to integers", () => {
    expect(normalizeResolutionDimension(1920.5)).toBe(1920);
    expect(normalizeResolutionDimension(1919.6)).toBe(1919);
  });

  it("returns 0 for empty input", () => {
    expect(normalizeResolutionDimension(undefined)).toBe(0);
    expect(normalizeResolutionDimension(null)).toBe(0);
  });
});
