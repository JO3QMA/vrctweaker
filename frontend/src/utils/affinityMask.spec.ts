import { describe, it, expect } from "vitest";
import {
  parseAffinityHex,
  formatAffinityHex,
  maskFromCoreStates,
  coreStatesFromMask,
  hiddenMaskAbove,
  hasHiddenBitsAbove,
  allCoresAllowedMask,
  mergeVisibleWithHidden,
  isValidAffinityMask,
} from "./affinityMask";

describe("parseAffinityHex", () => {
  it("parses uppercase hex without prefix", () => {
    expect(parseAffinityHex("FF")).toEqual({ ok: true, mask: 255n });
  });

  it("parses 0x prefix case-insensitively", () => {
    expect(parseAffinityHex("0xFf")).toEqual({ ok: true, mask: 255n });
  });

  it("treats empty as zero mask", () => {
    expect(parseAffinityHex("")).toEqual({ ok: true, mask: 0n });
  });

  it("rejects non-hex", () => {
    expect(parseAffinityHex("0,1").ok).toBe(false);
    expect(parseAffinityHex("GG").ok).toBe(false);
  });
});

describe("formatAffinityHex", () => {
  it("returns shortest uppercase hex", () => {
    expect(formatAffinityHex(255n)).toBe("FF");
    expect(formatAffinityHex(0xffffffffn)).toBe("FFFFFFFF");
  });

  it("returns empty for zero mask", () => {
    expect(formatAffinityHex(0n)).toBe("");
  });
});

describe("core visibility merge", () => {
  it("preserves hidden bits above core count", () => {
    const full = (1n << 32n) | 15n;
    const hidden = hiddenMaskAbove(full, 16);
    expect(hidden).toBe(1n << 32n);
    const visible = coreStatesFromMask(full, 16);
    expect(visible[0]).toBe(true);
    expect(visible[1]).toBe(true);
    expect(visible[2]).toBe(true);
    expect(visible[3]).toBe(true);
    const merged = mergeVisibleWithHidden(visible, hidden);
    expect(merged).toBe(full);
    expect(formatAffinityHex(merged)).toBe("10000000F");
  });

  it("builds mask from toggles", () => {
    const states = Array.from({ length: 4 }, (_, i) => i < 2);
    expect(maskFromCoreStates(states, 0n)).toBe(3n);
  });
});

describe("allCoresAllowedMask", () => {
  it("sets bits 0..count-1", () => {
    expect(allCoresAllowedMask(4)).toBe(15n);
    expect(allCoresAllowedMask(16)).toBe(0xffffn);
  });
});

describe("hasHiddenBitsAbove", () => {
  it("detects bits beyond display count", () => {
    expect(hasHiddenBitsAbove(255n, 8)).toBe(false);
    expect(hasHiddenBitsAbove((1n << 16n) | 1n, 16)).toBe(true);
  });
});

describe("isValidAffinityMask", () => {
  it("requires at least one allowed bit", () => {
    expect(isValidAffinityMask(0n)).toBe(false);
    expect(isValidAffinityMask(1n)).toBe(true);
  });
});
