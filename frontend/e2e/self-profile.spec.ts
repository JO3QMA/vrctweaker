import { test, expect } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";
import { E2E_SELF_DISPLAY_NAME, E2E_SELF_USER_ID } from "./fixtures/seed-data";

test.describe("Self profile (not logged in)", () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(getMockWailsInitScript());
  });

  test("shows login hint at /me", async ({ page }) => {
    await page.goto("/#/me");
    await expect(page.getByRole("heading", { level: 1 })).toContainText("自分");
    await expect(page.locator(".login-hint")).toContainText(
      "ログインが必要です",
    );
  });

  test("login hint links to settings", async ({ page }) => {
    await page.goto("/#/me");
    await page.locator(".settings-link").click();
    await expect(page.getByRole("heading", { level: 1 })).toContainText("設定");
  });

  test("sidebar navigates to /me", async ({ page }) => {
    await page.goto("/");
    await page.getByRole("menuitem", { name: "自分" }).click();
    await expect(page.getByRole("heading", { level: 1 })).toContainText("自分");
    await expect(page.locator(".login-hint")).toBeVisible();
  });
});

test.describe("Self profile (logged in)", () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(getMockWailsInitScript({ loggedIn: true }));
  });

  test("shows self profile display name", async ({ page }) => {
    await page.goto("/#/me");
    await expect(page.getByText("読み込み中")).toBeHidden({ timeout: 15_000 });
    await expect(page.locator(".profile-display-name")).toHaveText(
      E2E_SELF_DISPLAY_NAME,
    );
  });

  test("hides favorite and encounters tab in self variant", async ({
    page,
  }) => {
    await page.goto("/#/me");
    await expect(page.getByText("読み込み中")).toBeHidden({ timeout: 15_000 });
    await expect(page.getByText("お気に入り")).toBeHidden();
    await expect(page.getByRole("tab", { name: "遭遇履歴" })).toBeHidden();
    await expect(page.getByTestId("self-profile-refresh")).toBeVisible();
  });

  test("refresh button updates profile", async ({ page }) => {
    await page.goto("/#/me");
    await expect(page.getByText("読み込み中")).toBeHidden({ timeout: 15_000 });
    await expect(page.locator(".profile-status-desc")).toContainText(
      "E2E 自己プロフィール",
    );
    await page.getByTestId("self-profile-refresh").click();
    await expect(page.locator(".profile-status-desc")).toContainText(
      "E2E refreshed 1",
    );
  });

  test("settings view details navigates to /me", async ({ page }) => {
    await page.goto("/#/settings");
    await expect(page.getByTestId("settings-view-self-profile")).toBeVisible({
      timeout: 15_000,
    });
    await page.getByTestId("settings-view-self-profile").click();
    await expect(page.getByRole("heading", { level: 1 })).toContainText("自分");
    await expect(page.locator(".profile-display-name")).toHaveText(
      E2E_SELF_DISPLAY_NAME,
    );
  });

  test("user-profile redirects to /me for self vrcUserId", async ({ page }) => {
    const qs = new URLSearchParams({
      vrcUserId: E2E_SELF_USER_ID,
      displayName: E2E_SELF_DISPLAY_NAME,
    });
    await page.goto(`/#/user-profile?${qs}`);
    await expect(page.getByRole("heading", { level: 1 })).toContainText("自分");
    await expect(page.locator(".profile-display-name")).toHaveText(
      E2E_SELF_DISPLAY_NAME,
    );
  });
});
