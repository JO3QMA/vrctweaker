import { test, expect } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";
import { E2E_TEST_USER_DISPLAY_NAME } from "./fixtures/seed-data";

test.beforeEach(async ({ page }) => {
  await page.addInitScript(getMockWailsInitScript());
});

test.describe("Activity", () => {
  test("shows playtime and encounter sections with seeded data", async ({
    page,
  }) => {
    await page.goto("/#/activity");
    await expect(page.locator("h1")).toContainText("アクティビティ");

    const encounterCard = page.locator(".section-card--encounters");
    await expect(encounterCard.getByText("読み込み中")).toBeHidden({
      timeout: 15_000,
    });
    await expect(encounterCard.getByText("遭遇ログ")).toBeVisible();
    await expect(page.locator(".page-retention-hint")).toContainText("30");

    const playtimeCard = page.locator(".section-card--playtime");
    await expect(
      playtimeCard.getByText("プレイ時間（直近14日）"),
    ).toBeVisible();
    await expect(playtimeCard).toHaveClass(/section-card--collapsed/);
    await expect(playtimeCard.locator(".section-card__toggle")).toHaveAttribute(
      "aria-expanded",
      "false",
    );

    const encounterTable = encounterCard.locator(".el-table");
    await expect(encounterTable).toBeVisible();
    await expect(
      encounterTable
        .getByRole("button", { name: E2E_TEST_USER_DISPLAY_NAME })
        .first(),
    ).toBeVisible();
    await expect(encounterTable.locator(".el-table__row")).toHaveCount(3);
  });

  test("filters encounters by display name and refresh works", async ({
    page,
  }) => {
    await page.goto("/#/activity");
    await expect(page.locator("h1")).toContainText("アクティビティ");

    const encounterCard = page.locator(".section-card--encounters");
    await expect(encounterCard.getByText("読み込み中")).toBeHidden({
      timeout: 15_000,
    });

    const displayNameInput = encounterCard.getByPlaceholder("表示名で検索");
    await displayNameInput.fill("E2E Test");
    await expect(encounterCard.locator(".el-table__row")).toHaveCount(2);

    await displayNameInput.fill("E2E Visitor");
    await expect(encounterCard.locator(".el-table__row")).toHaveCount(1);
    await expect(
      encounterCard.getByRole("button", { name: "E2E Visitor" }),
    ).toBeVisible();

    await displayNameInput.clear();
    await expect(encounterCard.locator(".el-table__row")).toHaveCount(3);

    await encounterCard.getByRole("button", { name: "更新" }).click();
    await expect(encounterCard.getByText("読み込み中")).toBeHidden({
      timeout: 15_000,
    });
    await expect(encounterCard.locator(".el-table__row")).toHaveCount(3);
  });
});
