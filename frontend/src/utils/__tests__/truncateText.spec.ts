import { describe, it, expect } from "vitest";
import {
  REJOIN_WORLD_NAME_MAX_LEN,
  truncateRejoinWorldName,
  truncateText,
} from "../truncateText";

describe("truncateText", () => {
  it("returns text unchanged when within max length", () => {
    expect(truncateText("short", 12)).toBe("short");
  });

  it("appends ellipsis when text exceeds max length", () => {
    expect(truncateText("abcdefghijklmnop", 12)).toBe("abcdefghi...");
  });

  it("returns empty string for non-positive maxLength", () => {
    expect(truncateText("hello", 0)).toBe("");
    expect(truncateText("hello", -1)).toBe("");
  });
});

describe("truncateRejoinWorldName", () => {
  it("uses rejoin world name max length", () => {
    const long = "あ".repeat(REJOIN_WORLD_NAME_MAX_LEN + 5);
    expect(truncateRejoinWorldName(long)).toBe(
      `${"あ".repeat(REJOIN_WORLD_NAME_MAX_LEN - 3)}...`,
    );
  });
});
