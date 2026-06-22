import { describe, expect, it } from "vitest";
import { addLocalDays, eachLocalDateISO, localDateISO } from "../localDate";

describe("localDate", () => {
  it("formats local calendar date", () => {
    const d = new Date(2024, 2, 5, 23, 59, 59);
    expect(localDateISO(d)).toBe("2024-03-05");
  });

  it("adds local calendar days", () => {
    const d = new Date(2024, 0, 31, 12, 0, 0);
    expect(localDateISO(addLocalDays(d, 1))).toBe("2024-02-01");
  });

  it("iterates inclusive local date range", () => {
    expect([...eachLocalDateISO("2024-01-01", "2024-01-03")]).toEqual([
      "2024-01-01",
      "2024-01-02",
      "2024-01-03",
    ]);
  });
});
