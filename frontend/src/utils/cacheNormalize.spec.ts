import { describe, it, expect } from "vitest";
import { clampCacheNumeric } from "./cacheNormalize";

describe("clampCacheNumeric", () => {
  const min = 30;

  it("returns min when value is undefined", () => {
    expect(clampCacheNumeric(undefined, min)).toBe(min);
  });

  it("returns min when value is NaN", () => {
    expect(clampCacheNumeric(Number.NaN, min)).toBe(min);
  });

  it("returns min when value is below min", () => {
    expect(clampCacheNumeric(10, min)).toBe(min);
  });

  it("returns value when value is finite and at or above min", () => {
    expect(clampCacheNumeric(30, min)).toBe(30);
    expect(clampCacheNumeric(100, min)).toBe(100);
  });
});
