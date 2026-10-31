import { describe, it, expect } from "vitest";
import { isYtdlpVersionOlder } from "../ytdlpVersion";

describe("isYtdlpVersionOlder", () => {
  it("detects newer GitHub releases", () => {
    expect(isYtdlpVersionOlder("2026.01.01", "2026.02.02")).toBe(true);
    expect(isYtdlpVersionOlder("2026.02.02", "2026.02.02")).toBe(false);
    expect(isYtdlpVersionOlder("", "2026.02.02")).toBe(true);
  });
});
