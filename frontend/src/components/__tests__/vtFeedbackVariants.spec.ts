import { describe, it, expect } from "vitest";
import { vtAlertElementType } from "../vtAlertVariants";
import { vtTagElementType } from "../vtTagVariants";

describe("vtAlertElementType", () => {
  it("maps danger to error", () => {
    expect(vtAlertElementType("danger")).toBe("error");
  });
});

describe("vtTagElementType", () => {
  it("maps neutral to undefined", () => {
    expect(vtTagElementType("neutral")).toBeUndefined();
  });
});
