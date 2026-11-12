import { test, expect } from "@playwright/test";
import {
  MAIN_SCROLL_LAYOUT_VIEWPORT_WIDTHS,
  expectStableTransitionInsets,
  sampleRouteTransitionInsets,
} from "./fixtures/main-scroll-layout";
import { getMockWailsInitScript } from "./fixtures/mock-wails";

const TRANSITION_PAIRS = [
  {
    name: "video→activity",
    from: "/#/video",
    targetMenu: "アクティビティ",
  },
  {
    name: "activity→video",
    from: "/#/activity",
    targetMenu: "動画",
  },
  {
    name: "dashboard→gallery",
    from: "/#/",
    targetMenu: "ギャラリー",
  },
  {
    name: "gallery→config",
    from: "/#/gallery",
    targetMenu: "その他の設定",
  },
  {
    name: "config→settings",
    from: "/#/config",
    targetMenu: "設定",
    footerNav: true,
  },
  {
    name: "settings→config",
    from: "/#/settings/account",
    targetMenu: "その他の設定",
  },
] as const;

test.beforeEach(async ({ page }) => {
  await page.addInitScript(getMockWailsInitScript({ loggedIn: true }));
});

for (const pair of TRANSITION_PAIRS) {
  for (const width of MAIN_SCROLL_LAYOUT_VIEWPORT_WIDTHS) {
    test(`route transition keeps 24px insets ${pair.name} @ ${width}px`, async ({
      page,
    }) => {
      await page.setViewportSize({ width, height: 800 });
      await page.goto(pair.from);
      await expect(page.locator("h1").first()).toBeVisible({ timeout: 20_000 });

      if ("footerNav" in pair && pair.footerNav) {
        await page
          .locator(".sidebar-footer")
          .getByRole("menuitem", { name: pair.targetMenu })
          .click();
      } else {
        await page.getByRole("menuitem", { name: pair.targetMenu }).click();
      }

      const samples = await sampleRouteTransitionInsets(page);
      expectStableTransitionInsets(samples, 24);

      await expect(page.locator("h1").first()).toBeVisible({ timeout: 20_000 });
    });
  }
}
