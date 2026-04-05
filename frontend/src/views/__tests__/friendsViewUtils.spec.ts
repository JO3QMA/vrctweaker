import { describe, it, expect } from "vitest";
import {
  friendDetailStickyHeaderVisible,
  friendIsOffline,
  friendProfileBannerUrl,
  friendThumbUrl,
  jsonStringArray,
} from "../friends/friendsViewUtils";
import type { UserCacheDTO } from "../../wails/app";

function u(
  partial: Partial<UserCacheDTO> &
    Pick<UserCacheDTO, "vrcUserId" | "displayName" | "status">,
): UserCacheDTO {
  return {
    isFavorite: false,
    lastUpdated: "",
    ...partial,
  } as UserCacheDTO;
}

describe("friendDetailStickyHeaderVisible", () => {
  it("is false when not scrolled yet even if name is near top", () => {
    expect(
      friendDetailStickyHeaderVisible({
        scrollTop: 0,
        anchorTopViewport: 100,
        bodyTopViewport: 100,
      }),
    ).toBe(false);
  });

  it("is false when scrolled but name is still below header band", () => {
    expect(
      friendDetailStickyHeaderVisible({
        scrollTop: 40,
        anchorTopViewport: 200,
        bodyTopViewport: 100,
      }),
    ).toBe(false);
  });

  it("is true when scrolled and name row reaches body top band", () => {
    expect(
      friendDetailStickyHeaderVisible({
        scrollTop: 80,
        anchorTopViewport: 100,
        bodyTopViewport: 100,
      }),
    ).toBe(true);
  });
});

describe("friendIsOffline", () => {
  it("treats empty and offline as offline", () => {
    expect(friendIsOffline("")).toBe(true);
    expect(friendIsOffline("offline")).toBe(true);
    expect(friendIsOffline("Offline")).toBe(true);
  });

  it("treats other statuses as online", () => {
    expect(friendIsOffline("active")).toBe(false);
    expect(friendIsOffline("join me")).toBe(false);
  });
});

describe("friendThumbUrl", () => {
  it("prefers avatar thumbnail", () => {
    expect(
      friendThumbUrl(
        u({
          vrcUserId: "1",
          displayName: "A",
          status: "active",
          currentAvatarThumbnailImageUrl: "https://a",
          userIcon: "https://u",
        }),
      ),
    ).toBe("https://a");
  });
});

describe("friendProfileBannerUrl", () => {
  it("prefers profile pic override", () => {
    expect(
      friendProfileBannerUrl(
        u({
          vrcUserId: "1",
          displayName: "A",
          status: "active",
          profilePicOverride: "https://ppo",
          currentAvatarImageUrl: "https://av",
          currentAvatarThumbnailImageUrl: "https://th",
        }),
      ),
    ).toBe("https://ppo");
  });

  it("falls back to current avatar image then thumb chain", () => {
    expect(
      friendProfileBannerUrl(
        u({
          vrcUserId: "1",
          displayName: "A",
          status: "active",
          currentAvatarImageUrl: "https://av",
          currentAvatarThumbnailImageUrl: "https://th",
        }),
      ),
    ).toBe("https://av");
    expect(
      friendProfileBannerUrl(
        u({
          vrcUserId: "1",
          displayName: "A",
          status: "active",
          currentAvatarThumbnailImageUrl: "https://th",
        }),
      ),
    ).toBe("https://th");
  });
});

describe("jsonStringArray", () => {
  it("parses JSON string array", () => {
    expect(jsonStringArray('["a","b"]')).toEqual(["a", "b"]);
  });

  it("returns empty for invalid input", () => {
    expect(jsonStringArray("")).toEqual([]);
    expect(jsonStringArray("not json")).toEqual([]);
    expect(jsonStringArray("{}")).toEqual([]);
  });
});
