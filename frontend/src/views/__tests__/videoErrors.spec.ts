import { describe, expect, it } from "vitest";
import { classifyVideoError, videoErrorI18nKey } from "../videoErrors";

describe("classifyVideoError", () => {
  it("maps GitHub 403 / rate limit", () => {
    expect(
      classifyVideoError(
        'github api: 403 Forbidden: {"message":"API rate limit exceeded"}',
      ),
    ).toBe("githubRateLimit");
    expect(classifyVideoError("API rate limit exceeded for xxx")).toBe(
      "githubRateLimit",
    );
  });

  it("maps placement / developer mode errors", () => {
    expect(
      classifyVideoError(
        "symlink tools yt-dlp: enable Windows Developer Mode or run as administrator: access denied",
      ),
    ).toBe("placeOfficial");
  });

  it("maps network-ish failures", () => {
    expect(
      classifyVideoError("download request failed: dial tcp timeout"),
    ).toBe("network");
    expect(classifyVideoError("download URL is empty")).toBe("generic");
  });

  it("maps 401 separately from 403 forbidden", () => {
    expect(classifyVideoError("github api: 401 Unauthorized")).toBe(
      "githubUnauthorized",
    );
    expect(classifyVideoError("403 Forbidden: access denied")).toBe(
      "githubForbidden",
    );
  });

  it("handles null/undefined input", () => {
    expect(classifyVideoError(null)).toBe("generic");
    expect(classifyVideoError(undefined)).toBe("generic");
  });

  it("returns i18n key path", () => {
    expect(videoErrorI18nKey("API rate limit exceeded")).toBe(
      "video.errors.githubRateLimit",
    );
    expect(videoErrorI18nKey("403 Forbidden")).toBe(
      "video.errors.githubForbidden",
    );
  });
});
