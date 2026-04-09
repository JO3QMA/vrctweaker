import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { syncDocumentTitle } from "./syncDocumentTitle";

describe("syncDocumentTitle", () => {
  let prev: string;

  beforeEach(() => {
    prev = document.title;
  });

  afterEach(() => {
    document.title = prev;
  });

  it("sets title when titleKey is a non-empty string", () => {
    const t = (key: string) => (key === "meta.settings" ? "Settings" : "VRCT");
    syncDocumentTitle(t, { titleKey: "meta.settings" });
    expect(document.title).toBe("Settings - VRCT");
  });

  it("does not change title when titleKey is missing", () => {
    document.title = "unchanged";
    const t = () => "x";
    syncDocumentTitle(t, {});
    expect(document.title).toBe("unchanged");
  });

  it("does not change title when titleKey is empty string", () => {
    document.title = "unchanged";
    syncDocumentTitle(() => "", { titleKey: "" });
    expect(document.title).toBe("unchanged");
  });
});
