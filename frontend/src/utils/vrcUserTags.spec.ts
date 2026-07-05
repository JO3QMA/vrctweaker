import { describe, expect, it } from "vitest";
import {
  isKnownUserTag,
  isLanguageUserTag,
  isDeprecatedUserTag,
  resolveUserTagDisplay,
  userTagElementType,
} from "./vrcUserTags";

function mockT(key: string): string {
  const map: Record<string, string> = {
    "userDetail.userTags.deprecated": "Deprecated",
    "userDetail.userTags.unknown": "Unknown tag",
    "userDetail.userTags.tag_id": "ID",
    "userDetail.userTags.system_trust_basic.label": "New User (blue)",
    "userDetail.userTags.system_trust_basic.description":
      "User is New User (blue) Trust rank",
    "userDetail.userTags.language_jpn.label": "日本語",
    "userDetail.userTags.language_jpn.description": "Japanese",
    "userDetail.userTags.show_social_rank.label": "Show social rank",
    "userDetail.userTags.show_social_rank.description":
      "Toggle whether to show the user's real social rank",
  };
  return map[key] ?? key;
}

describe("vrcUserTags", () => {
  it("detects language tags", () => {
    expect(isLanguageUserTag("language_jpn")).toBe(true);
    expect(isLanguageUserTag("system_trust_basic")).toBe(false);
  });

  it("detects known tags", () => {
    expect(isKnownUserTag("system_trust_basic")).toBe(true);
    expect(isKnownUserTag("language_kor")).toBe(true);
    expect(isKnownUserTag("system_not_real")).toBe(false);
  });

  it("detects deprecated tags", () => {
    expect(isDeprecatedUserTag("show_social_rank")).toBe(true);
    expect(isDeprecatedUserTag("system_trust_basic")).toBe(false);
  });

  it("maps trust rank tags to element types", () => {
    expect(userTagElementType("system_trust_basic")).toBe("info");
    expect(userTagElementType("system_trust_known")).toBe("success");
    expect(userTagElementType("system_trust_trusted")).toBe("warning");
    expect(userTagElementType("system_trust_veteran")).toBe("primary");
  });

  it("trims tag id before resolving element type", () => {
    expect(userTagElementType(" system_trust_known ")).toBe("success");
  });

  it("resolves known user tag display", () => {
    const d = resolveUserTagDisplay("system_trust_basic", mockT);
    expect(d.isKnown).toBe(true);
    expect(d.label).toBe("New User (blue)");
    expect(d.tooltip).toContain("New User (blue) Trust rank");
    expect(d.tooltip).toContain("ID: system_trust_basic");
  });

  it("resolves language tag display", () => {
    const d = resolveUserTagDisplay("language_jpn", mockT);
    expect(d.isKnown).toBe(true);
    expect(d.label).toBe("日本語");
    expect(d.tooltip).toContain("Japanese");
  });

  it("marks deprecated in tooltip", () => {
    const d = resolveUserTagDisplay("show_social_rank", mockT);
    expect(d.deprecated).toBe(true);
    expect(d.tooltip).toContain("(Deprecated)");
  });

  it("falls back for unknown tags", () => {
    const d = resolveUserTagDisplay("system_slug", mockT);
    expect(d.isKnown).toBe(false);
    expect(d.label).toBe("system_slug");
    expect(d.tooltip).toContain("Unknown tag");
  });
});
