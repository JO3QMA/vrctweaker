import { describe, it, expect, vi, beforeEach } from "vitest";
import { createRouter, createWebHashHistory } from "vue-router";
import { navigateToUserProfile } from "../userProfileNavigation";
import { App } from "../../wails/app";

describe("navigateToUserProfile", () => {
  let router: ReturnType<typeof createRouter>;

  beforeEach(async () => {
    router = createRouter({
      history: createWebHashHistory(),
      routes: [
        { path: "/me", name: "me", component: { template: "<div/>" } },
        {
          path: "/friends",
          name: "friends",
          component: { template: "<div/>" },
        },
        {
          path: "/user-profile",
          name: "user-profile",
          component: { template: "<div/>" },
        },
      ],
    });
    await router.push("/");
    await router.isReady();
    vi.spyOn(router, "push");
  });

  it("navigates to me when openInSelfProfile", async () => {
    vi.spyOn(App, "resolveUserProfileNavigation").mockResolvedValue({
      openInFriendsView: false,
      openInSelfProfile: true,
      user: {
        vrcUserId: "usr_me",
        displayName: "Me",
        status: "",
        isFavorite: false,
        lastUpdated: "",
      },
    });

    await navigateToUserProfile(router, "usr_me", "Me");
    expect(router.push).toHaveBeenCalledWith({ name: "me" });
  });

  it("navigates to friends when openInFriendsView", async () => {
    vi.spyOn(App, "resolveUserProfileNavigation").mockResolvedValue({
      openInFriendsView: true,
      openInSelfProfile: false,
      user: {
        vrcUserId: "u1",
        displayName: "Friend",
        status: "",
        isFavorite: false,
        lastUpdated: "",
      },
    });

    await navigateToUserProfile(router, "u1");
    expect(router.push).toHaveBeenCalledWith({
      name: "friends",
      query: { vrcUserId: "u1" },
    });
  });
});
