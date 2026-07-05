import type { Router } from "vue-router";
import { App } from "../wails/app";

export async function navigateToUserProfile(
  router: Router,
  vrcUserId: string,
  displayName = "",
): Promise<void> {
  if (!vrcUserId) return;
  try {
    const nav = await App.resolveUserProfileNavigation(vrcUserId);
    if (nav.openInSelfProfile) {
      await router.push({ name: "me" });
      return;
    }
    if (nav.openInFriendsView) {
      await router.push({ name: "friends", query: { vrcUserId } });
      return;
    }
    await router.push({
      name: "user-profile",
      query: { vrcUserId, displayName },
    });
  } catch {
    await router.push({
      name: "user-profile",
      query: { vrcUserId, displayName },
    });
  }
}
