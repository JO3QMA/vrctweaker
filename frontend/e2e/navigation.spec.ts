import { test, expect } from "@playwright/test";
import { APP_ROUTES } from "./fixtures/app-routes";
import {
  E2E_ENCOUNTER_INSTANCE_KEY,
  E2E_TEST_USER_DISPLAY_NAME,
  E2E_TEST_USER_ID,
  E2E_WORLD_ID,
} from "./fixtures/seed-data";
import { getMockWailsInitScript } from "./fixtures/mock-wails";

/** Sidebar.vue の el-menu-item（設定除く） */
const MAIN_SIDEBAR_PATHS = [
  "/",
  "/launcher",
  "/gallery",
  "/activity",
  "/friends",
  "/me",
  "/automation",
  "/video",
  "/config",
] as const;

/**
 * navigation.spec.ts + gallery.spec.ts + self-profile.spec.ts で専用のナビ／ロードテストを持つルート名。
 * APP_ROUTES のうち 90% 以上を E2E でカバーする目標。
 */
const ROUTES_WITH_DEDICATED_E2E_TESTS = new Set([
  "dashboard",
  "launcher",
  "gallery",
  "activity",
  "friends",
  "me",
  "automation",
  "config",
  "video",
  "settings",
  "licenses",
  "encounter-history",
  "user-profile",
]);

test.beforeEach(async ({ page }) => {
  await page.addInitScript(getMockWailsInitScript());
});

test.describe("Navigation", () => {
  for (const route of APP_ROUTES.filter((r) =>
    (MAIN_SIDEBAR_PATHS as readonly string[]).includes(r.path),
  )) {
    test(`sidebar click navigates to ${route.name}`, async ({ page }) => {
      await page.goto("/");
      await page.getByRole("menuitem", { name: route.titleJa }).click();
      await expect(page.getByRole("heading", { level: 1 })).toContainText(
        route.titleJa,
      );
    });
  }

  test("footer settings navigates to settings", async ({ page }) => {
    await page.goto("/");
    await page
      .locator(".sidebar-footer")
      .getByRole("menuitem", { name: "設定" })
      .click();
    await expect(page.getByRole("heading", { level: 1 })).toContainText("設定");
  });

  test("settings .btn-licenses navigates to OSS licenses", async ({ page }) => {
    await page.goto("/#/settings");
    await page.locator(".btn-licenses").click();
    await expect(page.getByRole("heading", { level: 1 })).toContainText(
      "OSS ライセンス",
    );
  });

  test("direct goto encounter-history shows user encounter history", async ({
    page,
  }) => {
    await page.goto(
      `/#/activity/encounter-history?kind=user&vrcUserId=${E2E_TEST_USER_ID}`,
    );
    await expect(page.getByRole("heading", { level: 1 })).toContainText(
      "ユーザー別 遭遇履歴",
    );
    const instanceLink = page.getByTestId("encounter-instance-link").first();
    await expect(instanceLink).toBeVisible();
    await expect(instanceLink).toContainText("Public #88577 [JP]");
    await expect(instanceLink).toHaveAttribute(
      "href",
      `https://vrchat.com/home/launch?worldId=${E2E_WORLD_ID}&instanceId=88577%7Eregion%28jp%29`,
    );
    await expect(instanceLink).toHaveAttribute(
      "title",
      new RegExp(
        `Public #88577 \\[JP\\].*\\(${E2E_ENCOUNTER_INSTANCE_KEY.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}\\)`,
      ),
    );
  });

  test("direct goto user-profile shows title and display name", async ({
    page,
  }) => {
    const qs = new URLSearchParams({
      vrcUserId: E2E_TEST_USER_ID,
      displayName: E2E_TEST_USER_DISPLAY_NAME,
    });
    await page.goto(`/#/user-profile?${qs}`);
    await expect(page.getByRole("heading", { level: 1 })).toContainText(
      "ユーザー",
    );
    await expect(
      page.getByRole("heading", {
        name: E2E_TEST_USER_DISPLAY_NAME,
        level: 2,
      }),
    ).toBeVisible();
  });

  test("meta: >= 90% of APP_ROUTES have dedicated navigation/load E2E tests", () => {
    const covered = APP_ROUTES.filter((r) =>
      ROUTES_WITH_DEDICATED_E2E_TESTS.has(r.name),
    );
    const total = APP_ROUTES.length;
    const minCovered = Math.ceil(total * 0.9);
    expect(
      covered.length,
      `expected >= ${minCovered}/${total} routes covered, got ${covered.length}: ${covered.map((r) => r.name).join(", ")}`,
    ).toBeGreaterThanOrEqual(minCovered);
    expect(covered.length).toBeGreaterThanOrEqual(11);
  });
});
