import { test, expect } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";

test.beforeEach(async ({ page }) => {
  await page.addInitScript(getMockWailsInitScript());
});

test.describe("Gallery", () => {
  test("shows gallery page title", async ({ page }) => {
    await page.goto("/#/gallery");
    await expect(page.getByRole("heading", { level: 1 })).toContainText(
      "ギャラリー",
    );
  });

  test("seed screenshots populate grid after load", async ({ page }) => {
    await page.goto("/#/gallery");
    await expect(page.getByRole("heading", { level: 1 })).toContainText(
      "ギャラリー",
    );

    const grid = page.getByTestId("gallery-grid-scroll");
    const empty = page.getByText("スクリーンショットがありません");

    await expect(grid.or(empty)).toBeVisible({ timeout: 15_000 });
    await expect(empty).toBeHidden();
    await expect(grid).toBeVisible();
  });

  test("scan folder button completes scan via mock", async ({ page }) => {
    await page.goto("/#/gallery");
    await expect(page.getByTestId("gallery-grid-scroll")).toBeVisible({
      timeout: 15_000,
    });

    const scanBtn = page.getByTestId("gallery-scan-folder");
    await scanBtn.click();
    // E2E mock resolves ScanScreenshotDir immediately; progress may flash too fast to assert visible.
    await expect(page.getByTestId("gallery-scan-progress")).toBeHidden({
      timeout: 15_000,
    });
    await expect(scanBtn).toBeEnabled();
    await expect(page.getByTestId("gallery-grid-scroll")).toBeVisible();
  });

  test("selecting first screenshot shows detail preview", async ({ page }) => {
    await page.goto("/#/gallery");
    const grid = page.getByTestId("gallery-grid-scroll");
    await expect(grid).toBeVisible({ timeout: 15_000 });

    const firstItem = grid.locator(".grid-item").first();
    await expect(firstItem).toBeVisible();
    await firstItem.click();
    await expect(page.getByTestId("gallery-detail-preview")).toBeVisible();
  });
});
