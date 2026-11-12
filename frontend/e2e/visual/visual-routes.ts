import {
  E2E_TEST_USER_DISPLAY_NAME,
  E2E_TEST_USER_ID,
} from "../fixtures/seed-data";

const userProfileQuery = new URLSearchParams({
  vrcUserId: E2E_TEST_USER_ID,
  displayName: E2E_TEST_USER_DISPLAY_NAME,
}).toString();

/** Playwright ビジュアルリグレッションの対象ルート（hash ルーター） */
export const VISUAL_ROUTES = [
  { name: "dashboard", path: "/#/" },
  { name: "gallery", path: "/#/gallery" },
  { name: "activity", path: "/#/activity" },
  { name: "friends", path: "/#/friends" },
  {
    name: "user-profile",
    path: `/#/user-profile?${userProfileQuery}`,
  },
  { name: "me", path: "/#/me" },
  { name: "automation", path: "/#/automation" },
  { name: "settings", path: "/#/settings/general" },
] as const;

export type VisualRouteName = (typeof VISUAL_ROUTES)[number]["name"];

/** ルートごとに不安定になりやすい領域（`toHaveScreenshot` の `mask` 用セレクタ） */
export const VISUAL_ROUTE_MASK_SELECTORS: Partial<
  Record<VisualRouteName, string[]>
> = {
  gallery: [".thumbnail-wrap"],
  activity: [".section-card--playtime canvas"],
};

export const VISUAL_VIEWPORTS = [
  { label: "1280x800", width: 1280, height: 800 },
  { label: "768x540", width: 768, height: 540 },
] as const;

export const VISUAL_LOCALES = ["ja", "en"] as const;

/** `page.clock` で固定する日時（ギャラリー・アクティビティの日付表示を安定化） */
export const VISUAL_FIXED_CLOCK_ISO = "2024-06-15T12:00:00.000Z";
