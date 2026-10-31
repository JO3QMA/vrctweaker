import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  openExternalHttpUrl,
  openValidatedExternalHttpUrl,
} from "../openExternalHttpUrl";
import { parseSafeHttpUrl } from "../safeHttpUrl";

vi.mock("@/wails/app", () => ({
  isWailsRuntime: vi.fn(() => false),
  App: {
    openExternalURL: vi.fn(),
  },
}));

describe("openExternalHttpUrl", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("opens safe URLs in a new tab when not in Wails", async () => {
    const open = vi.spyOn(window, "open").mockImplementation(() => null);
    const ok = await openExternalHttpUrl("https://example.com/path");
    expect(ok).toBe(true);
    expect(open).toHaveBeenCalledWith(
      "https://example.com/path",
      "_blank",
      "noopener,noreferrer",
    );
  });

  it("returns false for unsafe URLs", async () => {
    const open = vi.spyOn(window, "open").mockImplementation(() => null);
    expect(await openExternalHttpUrl("javascript:1")).toBe(false);
    expect(open).not.toHaveBeenCalled();
  });
});

describe("openValidatedExternalHttpUrl", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("opens a pre-validated href in a new tab", async () => {
    const parsed = parseSafeHttpUrl("https://example.com");
    if (!parsed.ok) {
      throw new Error("expected a safe URL");
    }
    const open = vi.spyOn(window, "open").mockImplementation(() => null);
    const ok = await openValidatedExternalHttpUrl(parsed.href);
    expect(ok).toBe(true);
    expect(open).toHaveBeenCalledWith(
      "https://example.com/",
      "_blank",
      "noopener,noreferrer",
    );
  });
});
