import { describe, expect, it } from "vitest";
import { callApp } from "../app";

describe("callApp", () => {
  it("returns fallback when bindings are missing (Vitest skips dev wait)", async () => {
    const out = await callApp(async () => "invoked", "fallback");
    expect(out).toBe("fallback");
  });
});
