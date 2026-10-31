import { describe, it, expect } from "vitest";
import { isYtdlpVersionOlder } from "../ytdlpVersion";

describe("isYtdlpVersionOlder", () => {
  it("detects newer GitHub releases", () => {
    expect(isYtdlpVersionOlder("2026.01.01", "2026.02.02")).toBe(true);
    expect(isYtdlpVersionOlder("2026.02.02", "2026.02.02")).toBe(false);
    expect(isYtdlpVersionOlder("", "2026.02.02")).toBe(true);
  });

  it("treats empty latest as not older", () => {
    expect(isYtdlpVersionOlder("2026.01.01", "")).toBe(false);
    expect(isYtdlpVersionOlder("", "")).toBe(false);
  });

  it("ignores fourth segment and compares first three only", () => {
    expect(isYtdlpVersionOlder("2026.02.02", "2026.02.02.1")).toBe(false);
    expect(isYtdlpVersionOlder("2026.02.02.1", "2026.02.03")).toBe(true);
    expect(isYtdlpVersionOlder("2026.02.02.9", "2026.02.02.1")).toBe(false);
  });

  it("strips v prefix and trims whitespace", () => {
    expect(isYtdlpVersionOlder(" v2026.01.01 ", "2026.02.02")).toBe(true);
  });
});
