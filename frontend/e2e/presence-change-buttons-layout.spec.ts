import { test, expect } from "@playwright/test";
import { E2E_VIEWPORT_WIDTHS } from "./fixtures/app-window-bounds";
import { getMockWailsInitScript } from "./fixtures/mock-wails";
import {
  expectPresenceButtonsInsideCard,
  expectPresenceButtonsSingleLine,
  expectPresenceGridSingleRow,
  expectPresenceGridTwoByTwo,
  openDashboardPresenceLoggedIn,
  readPresenceLayoutMetrics,
  setUiLocale,
} from "./fixtures/presence-change-layout";

const LOCALES = ["ja", "en"] as const;

test.describe("Presence status buttons layout", () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(getMockWailsInitScript({ loggedIn: true }));
  });

  for (const locale of LOCALES) {
    for (const width of E2E_VIEWPORT_WIDTHS) {
      test(`dashboard ${locale} @ ${width}px: labels stay on one row without overflow`, async ({
        page,
      }) => {
        await page.setViewportSize({ width, height: 800 });
        await setUiLocale(page, locale);
        await openDashboardPresenceLoggedIn(page);

        const metrics = await readPresenceLayoutMetrics(page);
        expectPresenceButtonsSingleLine(metrics);
        expectPresenceButtonsInsideCard(metrics);
        expect(metrics.horizontalOverflow, "no horizontal scroll").toBe(false);
        if (width >= 1280) {
          expectPresenceGridSingleRow(metrics);
        }
      });
    }
  }

  for (const locale of LOCALES) {
    test(`narrow presence card ${locale}: switches to 2x2 grid`, async ({
      page,
    }) => {
      await page.setViewportSize({ width: 1280, height: 800 });
      await setUiLocale(page, locale);
      await openDashboardPresenceLoggedIn(page);

      await page.evaluate(() => {
        const section = document.querySelector(
          '[data-testid="presence-change-section"]',
        );
        if (!(section instanceof HTMLElement)) {
          throw new Error("presence-change-section missing");
        }
        section.style.width = "300px";
        section.style.maxWidth = "300px";
      });

      const metrics = await readPresenceLayoutMetrics(page);
      expectPresenceGridTwoByTwo(metrics);
      expectPresenceButtonsSingleLine(metrics);
      expectPresenceButtonsInsideCard(metrics);
      expect(metrics.horizontalOverflow, "no horizontal scroll").toBe(false);
    });
  }
});
