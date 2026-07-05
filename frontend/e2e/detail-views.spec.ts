import { test, expect } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";
import {
  E2E_TEST_USER_DISPLAY_NAME,
  E2E_TEST_USER_ID,
  E2E_WORLD_ID,
} from "./fixtures/seed-data";

test.beforeEach(async ({ page }) => {
  await page.addInitScript(getMockWailsInitScript());
});

test.describe("Detail views", () => {
  test("encounter-history shows warning for invalid query", async ({
    page,
  }) => {
    await page.goto("/#/activity/encounter-history");
    await expect(page.locator(".el-alert--warning")).toContainText(
      "表示できません",
    );
  });

  test("encounter-history shows world title for kind=world", async ({
    page,
  }) => {
    await page.goto(
      `/#/activity/encounter-history?kind=world&worldId=${E2E_WORLD_ID}`,
    );
    await expect(page.locator("h1")).toContainText("ワールド別");
  });

  test("user-profile shows error when user id is missing", async ({ page }) => {
    await page.goto("/#/user-profile");
    await expect(page.locator(".el-alert--warning")).toContainText(
      "ユーザー ID が指定されていません",
    );
  });

  test("user-profile shows display name from ResolveUserProfileNavigation", async ({
    page,
  }) => {
    const query = new URLSearchParams({
      vrcUserId: E2E_TEST_USER_ID,
      displayName: E2E_TEST_USER_DISPLAY_NAME,
    }).toString();
    await page.goto(`/#/user-profile?${query}`);
    await expect(page.locator(".profile-display-name")).toHaveText(
      E2E_TEST_USER_DISPLAY_NAME,
    );
  });

  test("licenses page shows OSS licenses table with npm rows", async ({
    page,
  }) => {
    await page.goto("/#/licenses");
    await expect(page.locator("h1")).toContainText("OSS ライセンス");
    const rowCount = await page.locator(".licenses-table tbody tr").count();
    expect(rowCount).toBeGreaterThan(0);
  });
});
