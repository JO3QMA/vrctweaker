import { afterEach, describe, expect, it, vi } from "vitest";
import type { UserCacheDTO } from "../wails/app";
import {
  copyDisplayName,
  friendDetailStickyHeaderVisible,
  friendLocationLabel,
  friendProfileBannerUrl,
  friendThumbUrl,
  jsonStringArray,
  PIPELINE_LOCATION_UNKNOWN,
} from "./vrcUserCacheDisplay";

function dto(partial: Partial<UserCacheDTO> = {}): UserCacheDTO {
  return {
    vrcUserId: "u",
    displayName: "D",
    status: "",
    isFavorite: false,
    lastUpdated: "",
    ...partial,
  };
}

describe("friendLocationLabel", () => {
  it("returns empty for missing location", () => {
    expect(friendLocationLabel(undefined)).toBe("");
    expect(friendLocationLabel("")).toBe("");
  });

  it("maps pipeline unknown sentinel to 不明", () => {
    expect(friendLocationLabel(PIPELINE_LOCATION_UNKNOWN)).toBe("不明");
  });

  it("passes through concrete locations", () => {
    expect(friendLocationLabel("wrld_x:123~grp")).toBe("wrld_x:123~grp");
  });
});

describe("friendThumbUrl", () => {
  it("prefers currentAvatarThumbnailImageUrl", () => {
    const u = dto({
      currentAvatarThumbnailImageUrl: "a",
      profilePicOverrideThumbnail: "b",
      userIcon: "c",
      imageUrl: "d",
    });
    expect(friendThumbUrl(u)).toBe("a");
  });

  it("falls through profilePicOverrideThumbnail, userIcon, imageUrl", () => {
    expect(
      friendThumbUrl(
        dto({ profilePicOverrideThumbnail: "b", userIcon: "c", imageUrl: "d" }),
      ),
    ).toBe("b");
    expect(friendThumbUrl(dto({ userIcon: "c", imageUrl: "d" }))).toBe("c");
    expect(friendThumbUrl(dto({ imageUrl: "d" }))).toBe("d");
  });

  it("returns undefined when none set", () => {
    expect(friendThumbUrl(dto())).toBeUndefined();
  });
});

describe("friendProfileBannerUrl", () => {
  it("prefers profilePicOverride then avatar image then imageUrl then thumb", () => {
    expect(friendProfileBannerUrl(dto({ profilePicOverride: "p" }))).toBe("p");
    expect(friendProfileBannerUrl(dto({ currentAvatarImageUrl: "big" }))).toBe(
      "big",
    );
    expect(friendProfileBannerUrl(dto({ imageUrl: "img" }))).toBe("img");
    expect(
      friendProfileBannerUrl(dto({ currentAvatarThumbnailImageUrl: "t" })),
    ).toBe("t");
  });
});

describe("friendDetailStickyHeaderVisible", () => {
  it("is false when scrollTop is 0", () => {
    expect(
      friendDetailStickyHeaderVisible({
        scrollTop: 0,
        anchorTopViewport: 0,
        bodyTopViewport: 100,
      }),
    ).toBe(false);
  });

  it("is true when scrolled and anchor is at or above body top with default slop", () => {
    expect(
      friendDetailStickyHeaderVisible({
        scrollTop: 10,
        anchorTopViewport: 100,
        bodyTopViewport: 100,
      }),
    ).toBe(true);
    expect(
      friendDetailStickyHeaderVisible({
        scrollTop: 5,
        anchorTopViewport: 104,
        bodyTopViewport: 100,
      }),
    ).toBe(true);
  });

  it("respects edgeSlopPx", () => {
    expect(
      friendDetailStickyHeaderVisible({
        scrollTop: 1,
        anchorTopViewport: 100,
        bodyTopViewport: 100,
        edgeSlopPx: 0,
      }),
    ).toBe(true);
  });
});

describe("jsonStringArray", () => {
  it("returns empty for undefined, blank, invalid JSON, non-array", () => {
    expect(jsonStringArray(undefined)).toEqual([]);
    expect(jsonStringArray("   ")).toEqual([]);
    expect(jsonStringArray("{")).toEqual([]);
    expect(jsonStringArray("{}")).toEqual([]);
    expect(jsonStringArray("[1]")).toEqual([]);
  });

  it("returns only string elements", () => {
    expect(jsonStringArray('["a", 1, "b"]')).toEqual(["a", "b"]);
  });
});

describe("copyDisplayName", () => {
  const writeText = vi.fn().mockResolvedValue(undefined);

  afterEach(() => {
    writeText.mockClear();
    vi.unstubAllGlobals();
  });

  it("no-op for empty name", async () => {
    vi.stubGlobal("navigator", { clipboard: { writeText } });
    await copyDisplayName("");
    expect(writeText).not.toHaveBeenCalled();
  });

  it("uses clipboard when available", async () => {
    vi.stubGlobal("navigator", { clipboard: { writeText } });
    await copyDisplayName("hello");
    expect(writeText).toHaveBeenCalledWith("hello");
  });

  it("falls back to execCommand when clipboard throws", async () => {
    writeText.mockRejectedValueOnce(new Error("denied"));
    vi.stubGlobal("navigator", { clipboard: { writeText } });
    const execCommand = vi.fn().mockReturnValue(true);
    Object.defineProperty(document, "execCommand", {
      value: execCommand,
      configurable: true,
      writable: true,
    });
    try {
      await copyDisplayName("x");
      expect(execCommand).toHaveBeenCalledWith("copy");
    } finally {
      Reflect.deleteProperty(document, "execCommand");
    }
  });
});
