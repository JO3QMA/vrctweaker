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

    await page.getByRole("button", { name: "+ 新規ルール" }).click();
    await expect(page.locator(".rule-editor")).toBeVisible();
    await expect(page.locator(".rule-editor")).toContainText("新規ルール");
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
    await expect(page.locator(".rule-editor")).toContainText("ルールを編集");

    const nameInput = page
      .locator(".rule-editor .el-form-item")
      .filter({ hasText: "ルール名" })
      .locator("input")
      .first();
    await expect(nameInput).toHaveValue("E2E AFK → Busy");
  });
});
