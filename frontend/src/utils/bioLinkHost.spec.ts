import { describe, expect, it } from "vitest";
import {
  bioLinkHostname,
  detectBioLinkSite,
  parseBioLinkUrl,
} from "./bioLinkHost";

describe("parseBioLinkUrl", () => {
  it("adds https when scheme is missing", () => {
    const parsed = parseBioLinkUrl("example.com/path");
    expect(parsed?.href).toBe("https://example.com/path");
  });

  it("returns null for empty or invalid values", () => {
    expect(parseBioLinkUrl("")).toBeNull();
    expect(parseBioLinkUrl("not a url")).toBeNull();
  });
});

describe("detectBioLinkSite", () => {
  it("detects common social and VRChat hosts", () => {
    expect(detectBioLinkSite("https://twitter.com/user")).toBe("twitter");
    expect(detectBioLinkSite("https://x.com/user")).toBe("twitter");
    expect(detectBioLinkSite("https://booth.pm/items/123")).toBe("booth");
    expect(detectBioLinkSite("https://github.com/org/repo")).toBe("github");
    expect(detectBioLinkSite("https://youtu.be/abc")).toBe("youtube");
    expect(detectBioLinkSite("https://discord.gg/invite")).toBe("discord");
    expect(detectBioLinkSite("https://twitch.tv/name")).toBe("twitch");
    expect(detectBioLinkSite("https://vrchat.com/home/user/usr_x")).toBe(
      "vrchat",
    );
    expect(detectBioLinkSite("https://www.pixiv.net/users/1")).toBe("pixiv");
    expect(detectBioLinkSite("https://instagram.com/name")).toBe("instagram");
    expect(detectBioLinkSite("https://tiktok.com/@name")).toBe("tiktok");
    expect(detectBioLinkSite("https://bsky.app/profile/name")).toBe("bluesky");
  });

  it("detects misskey instances by host or fediverse path", () => {
    expect(detectBioLinkSite("https://misskey.io/@user")).toBe("misskey");
    expect(detectBioLinkSite("https://social.example.misskey.online/@u")).toBe(
      "misskey",
    );
    expect(detectBioLinkSite("https://example.mk/@user")).toBe("misskey");
  });

  it("falls back to generic link for unknown hosts", () => {
    expect(detectBioLinkSite("https://example.com/profile")).toBe("link");
    expect(detectBioLinkSite("")).toBe("link");
  });
});

describe("bioLinkHostname", () => {
  it("returns hostname without www", () => {
    expect(bioLinkHostname("https://www.example.com/a")).toBe("example.com");
  });

  it("returns trimmed raw input when parsing fails", () => {
    expect(bioLinkHostname("  not a url  ")).toBe("not a url");
  });
});
