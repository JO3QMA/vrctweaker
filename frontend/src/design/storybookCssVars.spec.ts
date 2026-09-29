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

  it("resolves multiple var() hops for alias chains", () => {
    document.documentElement.style.setProperty(
      "--storybook-test-final",
      "#123",
    );
    document.documentElement.style.setProperty(
      "--storybook-test-mid",
      "var(--storybook-test-final)",
    );
    document.documentElement.style.setProperty(
      "--storybook-test-alias",
      "var(--storybook-test-mid)",
    );
    expect(readRootCssCustomProperty("--storybook-test-alias")).toBe("#123");
    document.documentElement.style.removeProperty("--storybook-test-alias");
    document.documentElement.style.removeProperty("--storybook-test-mid");
    document.documentElement.style.removeProperty("--storybook-test-final");
  });

  it("resolves one var() hop for alias tokens", () => {
    document.documentElement.style.setProperty(
      "--storybook-test-target",
      "#def",
    );
    document.documentElement.style.setProperty(
      "--storybook-test-alias",
      "var(--storybook-test-target)",
    );
    expect(readRootCssCustomProperty("--storybook-test-alias")).toBe("#def");
    document.documentElement.style.removeProperty("--storybook-test-alias");
    document.documentElement.style.removeProperty("--storybook-test-target");
  });
});
