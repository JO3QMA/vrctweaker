import { test, expect } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";

test.beforeEach(async ({ page }) => {
  await page.addInitScript(getMockWailsInitScript());
});

test.describe("Automation", () => {
  test("shows seeded rule and opens new rule editor", async ({ page }) => {
    await page.goto("/#/automation");
    await expect(page.locator("h1")).toContainText("オートメーション");

    await expect(page.getByText("E2E AFK → Busy")).toBeVisible();

    await page.getByTestId("add-item").click();
    await page.getByTestId("add-rule").click();
    await expect(page.locator(".rule-editor")).toBeVisible();
    await expect(page.getByTestId("automation-section-when")).toBeVisible();
  });

  test("clicking seeded rule opens editor with rule name field", async ({
    page,
  }) => {
    await page.goto("/#/automation");
    await expect(page.locator("h1")).toContainText("オートメーション");

    await page
      .locator(".rule-card")
      .filter({ hasText: "E2E AFK → Busy" })
      .click();
    await expect(page.locator(".rule-editor")).toBeVisible();
    await expect(page.getByTestId("editor-enabled")).toBeVisible();

    const nameInput = page.getByTestId("item-name-input");
    await expect(nameInput).toHaveValue("E2E AFK → Busy");
    await expect(page.getByTestId("automation-action-bar")).toBeVisible();
    await expect(page.getByTestId("automation-section-when")).toBeVisible();
  });

  test("run log uses Japanese event labels in one line", async ({ page }) => {
    await page.goto("/#/automation");
    await page.getByRole("tab", { name: /実行ログ/ }).click();
    await expect(
      page
        .getByTestId("automation-run-log-table")
        .getByText("スケジュール（毎分評価）"),
    ).toBeVisible();
    await expect(page.getByText("失敗")).toBeVisible();
    await expect(page.getByText("schedule.tick")).toHaveCount(0);
  });

  test("does not overflow horizontally from desktop down to a narrow frame", async ({
    page,
  }) => {
    test.setTimeout(60_000);
    for (const width of [1280, 960, 720, 600]) {
      await page.setViewportSize({ width, height: 800 });
      await page.goto("/#/automation");
      await page.reload();
      await expect(page.locator("h1")).toContainText("オートメーション");
      await page
        .locator(".rule-card")
        .filter({ hasText: "E2E AFK → Busy" })
        .click();
      await expect(page.locator(".rule-editor")).toBeVisible();
      const overflow = await page.evaluate(() => {
        const view = document.querySelector(".automation-view");
        if (!view) return 1;
        return view.scrollWidth - view.clientWidth;
      });
      expect(overflow, `horizontal overflow at ${width}px`).toBeLessThanOrEqual(
        1,
      );
    }
  });
});
