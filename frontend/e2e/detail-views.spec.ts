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
      "ユーザーIDが指定されていません",
    );
  });

  test("user-profile avatars tab shows log-derived avatar table", async ({
    page,
  }) => {
    const query = new URLSearchParams({
      vrcUserId: E2E_TEST_USER_ID,
      displayName: E2E_TEST_USER_DISPLAY_NAME,
    }).toString();
    await page.goto(`/#/user-profile?${query}`);
    await page.getByRole("tab", { name: "アバター" }).click();
    const table = page.getByTestId("friend-avatar-list").locator(".el-table");
    await expect(table).toBeVisible();
    await expect(table.locator(".el-table__row")).toHaveCount(2);
    await expect(table.getByText("E2E Avatar Alpha")).toBeVisible();
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
    await expect(page.locator("h1")).toContainText("OSSライセンス");
    const rowCount = await page.locator(".licenses-table tbody tr").count();
    expect(rowCount).toBeGreaterThan(0);
  });

  test("licenses table scrolls horizontally at 768px without clipping cells", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 768, height: 800 });
    await page.goto("/#/licenses");
    await expect(
      page.getByTestId("licenses-table-scroll-frontend"),
    ).toBeVisible();

    const metrics = await page.evaluate(() => {
      const wrapper = document.querySelector(
        '[data-testid="licenses-table-scroll-frontend"]',
      ) as HTMLElement | null;
      const pageOverflows =
        document.documentElement.scrollWidth >
        document.documentElement.clientWidth + 1;

      const inspectCell = (cell: HTMLElement | null) => {
        if (!cell) {
          return null;
        }
        const inner = cell.querySelector(".cell") as HTMLElement | null;
        if (!inner) {
          return null;
        }
        const content =
          (inner.querySelector(".package-name, .repo-link") as
            HTMLElement | undefined) ?? inner;
        const style = getComputedStyle(inner);
        const text = content.textContent ?? "";
        return {
          textOverflow: style.textOverflow,
          overflow: style.overflow,
          whiteSpace: style.whiteSpace,
          scrollWidth: content.scrollWidth,
          clientWidth: content.clientWidth,
          text,
          hasEllipsisChar: text.includes("…"),
        };
      };

      const longPackageRow = Array.from(
        document.querySelectorAll(".licenses-table tbody tr"),
      ).find((row) =>
        (row.querySelector(".package-name")?.textContent ?? "").includes(
          "@element-plus/icons",
        ),
      ) as HTMLElement | undefined;

      const repoRow = Array.from(
        document.querySelectorAll(".licenses-table tbody tr"),
      ).find((row) => row.querySelector(".repo-link")) as
        HTMLElement | undefined;

      return {
        pageOverflows,
        wrapperOverflowX: wrapper ? getComputedStyle(wrapper).overflowX : "",
        wrapperScrollable: wrapper
          ? wrapper.scrollWidth > wrapper.clientWidth
          : false,
        packageCell: inspectCell(
          longPackageRow?.querySelector(
            "td.licenses-table__package-cell",
          ) as HTMLElement | null,
        ),
        repoCell: inspectCell(
          repoRow?.querySelector(
            "td.licenses-table__repo-cell",
          ) as HTMLElement | null,
        ),
      };
    });

    expect(metrics.pageOverflows).toBe(false);
    expect(metrics.wrapperOverflowX).toBe("auto");
    expect(metrics.wrapperScrollable).toBe(true);

    expect(metrics.packageCell).not.toBeNull();
    expect(metrics.packageCell!.textOverflow).not.toBe("ellipsis");
    expect(metrics.packageCell!.overflow).toBe("visible");
    expect(metrics.packageCell!.hasEllipsisChar).toBe(false);
    expect(metrics.packageCell!.scrollWidth).toBeLessThanOrEqual(
      metrics.packageCell!.clientWidth + 1,
    );
    expect(metrics.packageCell!.text).toContain("@element-plus/icons-vue");

    expect(metrics.repoCell).not.toBeNull();
    expect(metrics.repoCell!.textOverflow).not.toBe("ellipsis");
    expect(metrics.repoCell!.overflow).toBe("visible");
    expect(metrics.repoCell!.hasEllipsisChar).toBe(false);
    expect(metrics.repoCell!.scrollWidth).toBeLessThanOrEqual(
      metrics.repoCell!.clientWidth + 1,
    );
    expect(metrics.repoCell!.text).toMatch(/^https?:\/\//);
  });
});
