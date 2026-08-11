import { describe, it, expect } from "vitest";
import { vtAlertElementType } from "../vtAlertVariants";
import { omitVtTagAttr, vtTagElementType } from "../vtTagVariants";

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

describe("omitVtTagAttr", () => {
  it("removes a single attribute", () => {
    expect(omitVtTagAttr({ type: "danger", class: "x" }, "type")).toEqual({
      class: "x",
    });
  });
});
