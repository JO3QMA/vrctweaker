import { test, expect, type Page } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";
import {
  E2E_TEST_USER_DISPLAY_NAME,
  E2E_TEST_USER_ID,
} from "./fixtures/seed-data";

const FRIEND_AVATAR_LAYOUT_VIEWPORTS = [
  { name: "768x540", width: 768, height: 540 },
  { name: "1280x800", width: 1280, height: 800 },
] as const;

const E2E_USER_NO_AVATAR_LOGS = "usr_e2e_visitor";

type FriendAvatarTableLayoutMetrics = {
  containerClientWidth: number;
  tableScrollWidth: number;
  columnWidths: number[];
  headersInside: boolean;
  countCellsInside: boolean;
  lastColumnRight: number;
  containerRight: number;
};

async function setUiLocale(page: Page, locale: "ja" | "en") {
  await page.goto("/#/settings/general");
  await page.getByTestId("settings-ui-language").click();
  const label = locale === "ja" ? "日本語" : "English";
  await page
    .locator(".el-select-dropdown__item")
    .filter({ hasText: label })
    .last()
    .click();
  await page.keyboard.press("Escape");
}

async function openFriendAvatarTab(
  page: Page,
  locale: "ja" | "en",
  vrcUserId: string,
  displayName: string,
) {
  await setUiLocale(page, locale);
  const tabName = locale === "ja" ? "アバター" : "Avatars";
  const query = new URLSearchParams({ vrcUserId, displayName }).toString();
  await page.goto(`/#/user-profile?${query}`);
  await page.getByRole("tab", { name: tabName }).click();
  await expect(
    page.getByTestId("friend-avatar-list").locator(".el-table"),
  ).toBeVisible();
}

async function measureFriendAvatarTableLayout(
  page: Page,
): Promise<FriendAvatarTableLayoutMetrics> {
  return page.evaluate(() => {
    const root = document.querySelector(
      '[data-testid="friend-avatar-list"]',
    ) as HTMLElement | null;
    if (!root) {
      throw new Error("friend-avatar-list missing");
    }
    const table = root.querySelector(".el-table") as HTMLElement | null;
    if (!table) {
      throw new Error("el-table missing");
    }
    const inner = table.querySelector(
      ".el-table__inner-wrapper",
    ) as HTMLElement | null;
    const headerCells = Array.from(
      table.querySelectorAll(".el-table__header-wrapper th.el-table__cell"),
    ) as HTMLElement[];
    const scrollHost = inner ?? table;
    const containerRect = root.getBoundingClientRect();
    const lastHeader = headerCells[headerCells.length - 1];
    const lastHeaderRect = lastHeader?.getBoundingClientRect();
    const countCells = Array.from(
      table.querySelectorAll(
        ".el-table__body-wrapper td.el-table__cell:last-child",
      ),
    ) as HTMLElement[];

    let lastBodyCellRight = lastHeaderRect?.right ?? 0;
    if (countCells.length > 0) {
      const lastCount = countCells[countCells.length - 1];
      lastBodyCellRight = Math.max(
        lastBodyCellRight,
        lastCount.getBoundingClientRect().right,
      );
    }

    const headersInside = headerCells.every((th) => {
      const r = th.getBoundingClientRect();
      return (
        r.left >= containerRect.left - 0.5 &&
        r.right <= containerRect.right + 0.5
      );
    });
    const countCellsInside =
      countCells.length === 0
        ? true
        : countCells.every((cell) => {
            const r = cell.getBoundingClientRect();
            return r.right <= containerRect.right + 0.5;
          });

    return {
      containerClientWidth: root.clientWidth,
      tableScrollWidth: scrollHost.scrollWidth,
      columnWidths: headerCells.map((th) => th.offsetWidth),
      headersInside,
      countCellsInside,
      lastColumnRight: Math.max(lastHeaderRect?.right ?? 0, lastBodyCellRight),
      containerRight: containerRect.right,
    };
  });
}

async function assertFriendAvatarTableFits(page: Page) {
  const metrics = await measureFriendAvatarTableLayout(page);
  expect(
    metrics.tableScrollWidth,
    `table scrollWidth ${metrics.tableScrollWidth} > container ${metrics.containerClientWidth}`,
  ).toBeLessThanOrEqual(metrics.containerClientWidth + 1);
  expect(
    metrics.lastColumnRight,
    `last column right ${metrics.lastColumnRight} > container ${metrics.containerRight}`,
  ).toBeLessThanOrEqual(metrics.containerRight + 1);
  expect(metrics.headersInside).toBe(true);
  expect(metrics.countCellsInside).toBe(true);
  return metrics;
}

test.beforeEach(async ({ page }) => {
  await page.addInitScript(getMockWailsInitScript());
});

for (const locale of ["ja", "en"] as const) {
  for (const viewport of FRIEND_AVATAR_LAYOUT_VIEWPORTS) {
    test(`friend avatar table fits with rows (${locale}, ${viewport.name})`, async ({
      page,
    }) => {
      await page.setViewportSize({
        width: viewport.width,
        height: viewport.height,
      });
      await openFriendAvatarTab(
        page,
        locale,
        E2E_TEST_USER_ID,
        E2E_TEST_USER_DISPLAY_NAME,
      );
      await expect(
        page.getByTestId("friend-avatar-list").locator(".el-table__row"),
      ).toHaveCount(2);
      const metrics = await assertFriendAvatarTableFits(page);
      const countColWidth =
        metrics.columnWidths[metrics.columnWidths.length - 1] ?? 0;
      expect(countColWidth).toBeGreaterThan(60);
      console.log(
        `friend-avatar-layout:with-rows:${locale}:${viewport.name}:${JSON.stringify(metrics)}`,
      );
    });

    test(`friend avatar table fits when empty (${locale}, ${viewport.name})`, async ({
      page,
    }) => {
      await page.setViewportSize({
        width: viewport.width,
        height: viewport.height,
      });
      await openFriendAvatarTab(
        page,
        locale,
        E2E_USER_NO_AVATAR_LOGS,
        "E2E Visitor",
      );
      const emptyText =
        locale === "ja" ? "アバターはまだありません" : "No avatars yet";
      await expect(
        page.getByTestId("friend-avatar-list").locator(".el-table__empty-text"),
      ).toContainText(emptyText);
      const metrics = await assertFriendAvatarTableFits(page);
      console.log(
        `friend-avatar-layout:empty:${locale}:${viewport.name}:${JSON.stringify(metrics)}`,
      );
    });
  }
}
