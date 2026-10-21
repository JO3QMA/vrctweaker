import { test, expect, type Page } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";

const LONG_UPDATE_ERROR = "https://updates.example.invalid/" + "x".repeat(180);

const FRAME_WIDTHS = [740, 1100] as const;

test.beforeEach(async ({ page }) => {
  await page.addInitScript(getMockWailsInitScript());
});

async function setSettingsFrameWidth(
  page: Page,
  target: number,
): Promise<void> {
  await page.setViewportSize({ width: 1400, height: 900 });
  await page.goto("/#/settings/about");
  await page.locator(".settings-frame").waitFor();
  const current = await page
    .locator(".settings-frame")
    .evaluate((el) => el.clientWidth);
  const viewport = page.viewportSize();
  if (!viewport) {
    throw new Error("viewport is unset");
  }
  await page.setViewportSize({
    width: viewport.width + (target - current),
    height: 900,
  });
}

async function showLongUpdateError(page: Page): Promise<void> {
  await page.evaluate((message) => {
    const app = window.go?.wailsapp?.App as
      { CheckForAppUpdate?: () => Promise<unknown> } | undefined;
    if (!app) {
      throw new Error("wails app mock is missing");
    }
    app.CheckForAppUpdate = () => Promise.reject(new Error(message));
  }, LONG_UPDATE_ERROR);
  await page.getByTestId("settings-check-app-update").click();
  await expect(page.getByTestId("settings-app-update-error")).toBeVisible();
  await expect(page.getByTestId("settings-app-update-error")).toContainText(
    "更新を確認できませんでした",
  );
  await expect(page.getByTestId("settings-app-update-error")).toContainText(
    LONG_UPDATE_ERROR,
  );
}

test.describe("Settings update error", () => {
  for (const target of FRAME_WIDTHS) {
    test(`about alert stays inside the card at frame width ${target}px`, async ({
      page,
    }) => {
      await setSettingsFrameWidth(page, target);
      await page.goto("/#/settings/about");
      await showLongUpdateError(page);

      const body = page.locator(
        '[data-testid="settings-about-app-section"] .el-card__body',
      );
      const fit = await body.evaluate((el) => ({
        scrollWidth: el.scrollWidth,
        clientWidth: el.clientWidth,
      }));
      expect(fit.scrollWidth).toBeLessThanOrEqual(fit.clientWidth);

      const alertBox = await page
        .getByTestId("settings-app-update-error")
        .evaluate((el) => {
          const cardEl = el.closest(".section-card");
          if (!(cardEl instanceof HTMLElement)) {
            throw new Error("alert is outside the about card");
          }
          const alertRect = el.getBoundingClientRect();
          const cardRect = cardEl.getBoundingClientRect();
          return {
            withinCard:
              alertRect.left >= cardRect.left - 1 &&
              alertRect.right <= cardRect.right + 1,
          };
        });
      expect(alertBox.withinCard).toBe(true);
    });
  }
});
