import { describe, it, expect } from "vitest";
import { shouldShowLicensesTableScrollFade } from "../licensesTableScrollFade";

describe("shouldShowLicensesTableScrollFade", () => {
  it("hides when there is no overflow", () => {
    expect(shouldShowLicensesTableScrollFade(0, 400, 400)).toBe(false);
    expect(shouldShowLicensesTableScrollFade(0, 399, 400)).toBe(false);
  });

  it("shows when overflow and not at the right end", () => {
    expect(shouldShowLicensesTableScrollFade(0, 800, 400)).toBe(true);
    expect(shouldShowLicensesTableScrollFade(100, 800, 400)).toBe(true);
  });

  it("hides when scrolled to the right end", () => {
    expect(shouldShowLicensesTableScrollFade(400, 800, 400)).toBe(false);
    expect(shouldShowLicensesTableScrollFade(399, 800, 400)).toBe(false);
  });
});
