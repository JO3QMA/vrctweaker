import { describe, it, expect } from "vitest";
import { vrcStatusElementTagType } from "../vrcStatus";

describe("vrcStatusElementTagType", () => {
  it.each([
    ["offline", "info"],
    ["Offline", "info"],
    [" join me ", "success"],
    ["JOIN ME", "success"],
    ["busy", "danger"],
    ["ask me", "warning"],
    ["active", "primary"],
    ["Active", "primary"],
    ["unknown-custom", "primary"],
  ] as const)("maps %s -> %s", (input, want) => {
    expect(vrcStatusElementTagType(input)).toBe(want);
  });

  it("maps empty / whitespace to info", () => {
    expect(vrcStatusElementTagType("")).toBe("info");
    expect(vrcStatusElementTagType("   ")).toBe("info");
    expect(vrcStatusElementTagType(undefined)).toBe("info");
    expect(vrcStatusElementTagType(null)).toBe("info");
  });
});
