import { test, expect } from "@playwright/test";
import { APP_ROUTES } from "./fixtures/app-routes";
import {
  expectRouteMainScrollLayout,
  MAIN_SCROLL_LAYOUT_VIEWPORT_WIDTHS,
} from "./fixtures/main-scroll-layout";
import { getMockWailsInitScript } from "./fixtures/mock-wails";

const ROUTES_FOR_LAYOUT = APP_ROUTES;

test.beforeEach(async ({ page }) => {
  await page.addInitScript(getMockWailsInitScript({ loggedIn: true }));
});

for (const route of ROUTES_FOR_LAYOUT) {
  for (const width of MAIN_SCROLL_LAYOUT_VIEWPORT_WIDTHS) {
    test(`main scroll layout ${route.name} @ ${width}px`, async ({ page }) => {
      await page.goto(`/#${route.path}`);
      await expect(page.locator("h1").first()).toBeVisible({ timeout: 20_000 });

      if (route.name === "activity") {
        const encounterCard = page.locator(".section-card--encounters");
        await expect(encounterCard.getByText("遭遇ログ")).toBeVisible({
          timeout: 15_000,
        });
      }

      if (route.name === "gallery") {
        await expect(page.getByTestId("gallery-grid-scroll")).toBeVisible({
          timeout: 20_000,
        });
      }

      if (route.name === "me") {
        await expect(page.locator(".detail-wrap")).toBeVisible({
          timeout: 15_000,
        });
      }

      await expectRouteMainScrollLayout(page, route.name, width);
    });
  }
}
