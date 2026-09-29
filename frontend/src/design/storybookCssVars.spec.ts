import { afterEach, describe, expect, it } from "vitest";
import { readRootCssCustomProperty } from "./storybookCssVars";

describe("storybookCssVars", () => {
  afterEach(() => {
    document.documentElement.style.removeProperty("--storybook-test-token");
  });

  it("reads custom properties with the -- prefix", () => {
    document.documentElement.style.setProperty(
      "--storybook-test-token",
      "#abc",
    );
    expect(readRootCssCustomProperty("--storybook-test-token")).toBe("#abc");
  });
});
