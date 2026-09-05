import { describe, expect, it, vi, afterEach } from "vitest";
import {
  ENCOUNTER_LOG_TIME_COL_WIDTH,
  formatEncounteredAt,
  formatEncounterLogTimestamp,
} from "./formatEncounteredAt";

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

describe("formatEncounterLogTimestamp", () => {
  it("exports encounter log column width", () => {
    expect(ENCOUNTER_LOG_TIME_COL_WIDTH).toBe(130);
  });

  it("formats valid ISO in local time as yyyy/MM/DD HH:mm:ss", () => {
    const iso = "2025-06-01T00:00:00.000Z";
    const d = new Date(iso);
    const pad = (n: number) => String(n).padStart(2, "0");
    const expected = `${d.getFullYear()}/${pad(d.getMonth() + 1)}/${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
    const result = formatEncounterLogTimestamp(iso);
    expect(result).toBe(expected);
    expect(result).toMatch(/^\d{4}\/\d{2}\/\d{2} \d{2}:\d{2}:\d{2}$/);
    expect(result!.length).toBe(19);
  });

  it("returns null for empty or invalid input", () => {
    expect(formatEncounterLogTimestamp("")).toBeNull();
    expect(formatEncounterLogTimestamp("   ")).toBeNull();
    expect(formatEncounterLogTimestamp("not-a-date")).toBeNull();
  });
});
