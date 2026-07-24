import { test, expect } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";

test.beforeEach(async ({ page }) => {
  await page.addInitScript(getMockWailsInitScript());
});

test.describe("Launcher redesign", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/#/launcher");
    await expect(page.locator("h1")).toContainText("ランチャー");
    await expect(page.locator(".profile-editor")).toBeVisible();
  });

  test("shows unsaved banner and dot after editing primary launch options", async ({
    page,
  }) => {
    await expect(page.getByTestId("unsaved-banner")).toHaveCount(0);

    await page.getByTestId("no-vr-checkbox").click();

    await expect(page.getByTestId("unsaved-banner")).toBeVisible();
    await expect(page.getByTestId("unsaved-dot")).toBeVisible();
  });

  test("collapses and expands profile sidebar", async ({ page }) => {
    await expect(page.locator(".profiles-list")).toBeVisible();

    await page.getByTestId("sidebar-toggle-btn").click();
    await expect(page.locator(".profiles-list")).toBeHidden();

    await page.getByTestId("sidebar-toggle-btn").click();
    await expect(page.locator(".profiles-list")).toBeVisible();
  });

  test("prompts on profile switch and cancels to keep editing", async ({
    page,
  }) => {
    await page.getByTestId("no-vr-checkbox").click();
    await expect(page.getByTestId("unsaved-banner")).toBeVisible();

    await page.getByTestId("profile-card-profile-2").click();

    const dialog = page.locator(".el-message-box");
    await expect(dialog).toBeVisible();
    await expect(dialog).toContainText("未保存の変更");

    await dialog.locator(".el-message-box__headerbtn").click();
    await expect(dialog).toBeHidden();

    await expect(page.getByTestId("unsaved-banner")).toBeVisible();
    await expect(page.getByText("デフォルトプロファイル")).toBeVisible();
  });

  test("discards unsaved edits when switching profiles", async ({ page }) => {
    await page.getByTestId("no-vr-checkbox").click();
    await page.getByTestId("profile-card-profile-2").click();

    const dialog = page.locator(".el-message-box");
    await expect(dialog).toBeVisible();
    await page.getByRole("button", { name: "破棄" }).click();
    await expect(dialog).toBeHidden();

    await expect(page.getByTestId("unsaved-banner")).toHaveCount(0);
    await expect(
      page.locator('.profile-editor input[type="text"]').first(),
    ).toHaveValue("デスクトップ用");
  });

  test("prompts before leaving Launcher when edits are unsaved", async ({
    page,
  }) => {
    await page.getByTestId("no-vr-checkbox").click();

    await page.getByRole("menuitem", { name: "ダッシュボード" }).click();

    const dialog = page.locator(".el-message-box");
    await expect(dialog).toBeVisible();
    await page.getByRole("button", { name: "破棄" }).click();
    await expect(dialog).toBeHidden();

    await expect(page.locator("h1")).toContainText("ダッシュボード");
  });

  test("creates and selects a saved profile from new profile button", async ({
    page,
  }) => {
    await page.getByRole("button", { name: "+ 新規プロファイル" }).click();

    await expect(
      page
        .locator(".profiles-list .profile-name")
        .getByText("新しいプロファイル"),
    ).toBeVisible();
    await expect(
      page.locator('.profile-editor input[type="text"]').first(),
    ).toHaveValue("新しいプロファイル");
    await expect(page.getByTestId("unsaved-banner")).toHaveCount(0);
  });

  test("deletes saved profile from overflow menu", async ({ page }) => {
    await page.getByTestId("profile-card-profile-2").click();
    await expect(
      page.locator('.profile-editor input[type="text"]').first(),
    ).toHaveValue("デスクトップ用");

    await page.getByTestId("profile-overflow-btn").click();
    await page.getByTestId("delete-profile-btn").click();

    const dialog = page.locator(".el-message-box");
    await expect(dialog).toContainText("デスクトップ用");
    await page.getByRole("button", { name: "削除" }).click();
    await expect(dialog).toBeHidden();

    await expect(page.getByTestId("profile-card-profile-2")).toHaveCount(0);
  });
});
