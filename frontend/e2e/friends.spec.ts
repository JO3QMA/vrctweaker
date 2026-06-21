import { test, expect } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";

test.beforeEach(async ({ page }) => {
  await page.addInitScript(getMockWailsInitScript());
});

test.describe("Friends", () => {
  test("shows login hint when not logged in", async ({ page }) => {
    await page.goto("/#/friends");
    await expect(page.locator("h1")).toContainText("フレンド");

    await expect(page.locator(".login-hint")).toBeVisible();
    await expect(page.locator(".login-hint")).toContainText(
      "フレンド一覧の更新にはログインが必要です",
    );
  });

  test("shows cached online friends and empty search message when not logged in", async ({
    page,
  }) => {
    await page.goto("/#/friends");
    await expect(page.locator("h1")).toContainText("フレンド");

    await expect(page.getByText("読み込み中")).toBeHidden({ timeout: 15_000 });

    await expect(page.getByText("E2E Online Friend")).toBeVisible();
    await expect(page.getByText("E2E Offline Friend")).toBeHidden();

    const searchInput = page.getByTestId("friends-search-display-name");
    await searchInput.fill("存在しない名前");
    await expect(
      page.getByText("検索に一致するフレンドはいません"),
    ).toBeVisible();
  });
});
