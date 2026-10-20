import { describe, expect, it } from "vitest";
import { isSafeHttpUrl, parseSafeHttpUrl } from "../safeHttpUrl";

describe("parseSafeHttpUrl", () => {
  it("accepts https and http URLs", () => {
    expect(parseSafeHttpUrl("https://example.com/x")).toEqual({
      ok: true,
      href: "https://example.com/x",
    });
    expect(parseSafeHttpUrl("http://example.com")).toEqual({
      ok: true,
      href: "http://example.com/",
    });
  });

  it("trims whitespace", () => {
    expect(parseSafeHttpUrl("  https://a.test  ")).toEqual({
      ok: true,
      href: "https://a.test/",
    });
  });

  it("rejects javascript and data URLs", () => {
    expect(parseSafeHttpUrl("javascript:alert(1)").ok).toBe(false);
    expect(parseSafeHttpUrl("data:text/html,x").ok).toBe(false);
  });

  it("rejects empty and relative URLs", () => {
    expect(parseSafeHttpUrl("").ok).toBe(false);
    expect(parseSafeHttpUrl("/only-path").ok).toBe(false);
    expect(parseSafeHttpUrl("http://:80").ok).toBe(false);
  });
});

describe("isSafeHttpUrl", () => {
  it("mirrors parseSafeHttpUrl", () => {
    expect(isSafeHttpUrl("https://ok")).toBe(true);
    expect(isSafeHttpUrl("vbscript:foo")).toBe(false);
  });
});
