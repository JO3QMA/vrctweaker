import { describe, expect, it } from "vitest";
import { App, callApp } from "../app";

describe("callApp", () => {
  it("returns fallback when bindings are missing (Vitest skips dev wait)", async () => {
    const out = await callApp(async () => "invoked", "fallback");
    expect(out).toBe("fallback");
  });

  it("App.getUILanguage falls back to ja without bindings", async () => {
    await expect(App.getUILanguage()).resolves.toBe("ja");
  });

  it("App.setUILanguage no-ops without bindings", async () => {
    await expect(App.setUILanguage("en")).resolves.toBeUndefined();
  });
});
