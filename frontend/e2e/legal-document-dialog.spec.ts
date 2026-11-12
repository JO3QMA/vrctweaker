import { test, expect } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";

const VIEWPORTS = [
  { width: 768, height: 540 },
  { width: 1280, height: 800 },
] as const;

async function openLegalDocumentDialog(page: import("@playwright/test").Page) {
  await page.goto("/#/settings/reset");
  await expect(page.getByTestId("settings-reset-setup-section")).toBeVisible();
  await page.getByTestId("settings-open-onboarding").click();
  await expect(page.getByTestId("onboarding-wizard")).toBeVisible();
  await page.getByTestId("legal-link-privacy").click();
  await expect(page.getByTestId("legal-document-dialog")).toBeVisible();
}

test.describe("LegalDocumentDialog (append-to-body scroll)", () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(getMockWailsInitScript());
  });

  for (const viewport of VIEWPORTS) {
    test(`body scrolls and dialog fits viewport (${viewport.width}x${viewport.height})`, async ({
      page,
    }) => {
      await page.setViewportSize(viewport);
      await openLegalDocumentDialog(page);

      const dialog = page.locator(".legal-document-dialog").last();
      await expect(dialog).toBeVisible();

      const viewportSize = page.viewportSize();
      if (!viewportSize) {
        throw new Error("viewport size is missing");
      }

      const dialogBox = await dialog.boundingBox();
      expect(dialogBox).not.toBeNull();
      if (dialogBox) {
        expect(dialogBox.y).toBeGreaterThanOrEqual(-1);
        expect(dialogBox.y + dialogBox.height).toBeLessThanOrEqual(
          viewportSize.height + 1,
        );
        expect(dialogBox.x).toBeGreaterThanOrEqual(-1);
        expect(dialogBox.x + dialogBox.width).toBeLessThanOrEqual(
          viewportSize.width + 1,
        );
      }

      const bodyMetrics = await dialog
        .locator(".el-dialog__body")
        .evaluate((el) => ({
          overflowY: getComputedStyle(el).overflowY,
          scrollHeight: el.scrollHeight,
          clientHeight: el.clientHeight,
        }));

      expect(bodyMetrics.overflowY).toBe("auto");
      expect(bodyMetrics.scrollHeight).toBeGreaterThan(
        bodyMetrics.clientHeight,
      );

      await expect(dialog.locator(".el-dialog__headerbtn")).toBeVisible();
    });
  }
});
