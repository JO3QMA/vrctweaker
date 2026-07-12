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
        "公式 yt-dlp の配置に失敗しました（Windows の開発者モードを有効にするか、管理者として実行してください）: access denied",
      ),
    ).toBe("placeOfficial");
  });

  it("maps network-ish failures", () => {
    expect(
      classifyVideoError("download request failed: dial tcp timeout"),
    ).toBe("network");
    expect(classifyVideoError("download URL is empty")).toBe("generic");
  });

  it("returns i18n key path", () => {
    expect(videoErrorI18nKey("403 Forbidden")).toBe(
      "video.errors.githubRateLimit",
    );
  });
});
