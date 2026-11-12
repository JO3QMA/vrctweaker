import { describe, it, expect } from "vitest";
import { formatNumericRangeHint } from "../formNumericRangeHint";

describe("formatNumericRangeHint", () => {
  it("passes min and max into i18n", () => {
    const t = (key: string, params?: Record<string, string>) =>
      `${key}:${params?.min}-${params?.max}`;
    expect(formatNumericRangeHint(t, 1280, 7680)).toBe(
      "formValidation.numericRangeHint:1280-7680",
    );
  });
});
