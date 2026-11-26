import { test, expect, type Page } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";
import {
  E2E_TEST_USER_DISPLAY_NAME,
  E2E_TEST_USER_ID,
  E2E_WORLD_ID,
} from "./fixtures/seed-data";

test.beforeEach(async ({ page }) => {
  await page.addInitScript(getMockWailsInitScript());
});

function parseRgbTriplet(cssColor: string): [number, number, number] | null {
  const match = cssColor.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)/);
  if (!match) {
    return null;
  }
  return [Number(match[1]), Number(match[2]), Number(match[3])];
}

function colorsVisiblyDiffer(foreground: string, background: string): boolean {
  const fg = parseRgbTriplet(foreground);
  const bg = parseRgbTriplet(background);
  if (!fg || !bg) {
    return foreground !== background;
  }
  return fg.some((channel, index) => Math.abs(channel - bg[index]!) > 12);
}

function isOpaqueBackground(backgroundColor: string): boolean {
  if (/rgba\(\s*0\s*,\s*0\s*,\s*0\s*,\s*0\s*\)/.test(backgroundColor)) {
    return false;
  }
  const match = backgroundColor.match(
    /rgba\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*,\s*([\d.]+)\s*\)/,
  );
  if (match) {
    return Number(match[4]) >= 0.99;
  }
  return parseRgbTriplet(backgroundColor) !== null;
}

async function measureLicensesTableGeometry(page: Page) {
  return page.evaluate(() => {
    const wrapper = document.querySelector(
      '[data-testid="licenses-table-scroll-frontend"]',
    ) as HTMLElement | null;
    const table = wrapper?.querySelector(
      "table.licenses-table",
    ) as HTMLTableElement | null;
    const pageOverflows =
      document.documentElement.scrollWidth >
      document.documentElement.clientWidth + 1;

    if (!wrapper || !table) {
      return { ok: false as const, reason: "missing-table" };
    }

    const headerRow = table.querySelector("thead tr");
    const bodyRows = Array.from(table.querySelectorAll("tbody tr")).filter(
      (row) => !row.querySelector(".licenses-table__empty"),
    );
    const headerCells = headerRow
      ? Array.from(headerRow.querySelectorAll("th"))
      : [];
    const sampleRow =
      bodyRows.find((row) =>
        (row.querySelector(".package-name")?.textContent ?? "").includes(
          "@eslint-community/eslint-utils",
        ),
      ) ?? bodyRows[0];

    if (!sampleRow || headerCells.length === 0) {
      return { ok: false as const, reason: "missing-rows" };
    }

    const bodyCells = Array.from(sampleRow.querySelectorAll("td"));
    const headerWidths = headerCells.map(
      (cell) => cell.getBoundingClientRect().width,
    );
    const bodyWidths = bodyCells.map(
      (cell) => cell.getBoundingClientRect().width,
    );

    const columnWidthDelta = headerWidths.map((w, i) =>
      Math.abs(w - (bodyWidths[i] ?? 0)),
    );

    const packageName = sampleRow.querySelector(
      ".package-name",
    ) as HTMLElement | null;
    const versionCell = sampleRow.querySelector(
      ".licenses-table__version-cell",
    ) as HTMLElement | null;
    const repoLink = sampleRow.querySelector(
      ".repo-link",
    ) as HTMLElement | null;
    const licenseCell = sampleRow.querySelector(
      ".licenses-table__license-cell",
    ) as HTMLElement | null;
    const licenseTag = licenseCell?.querySelector(
      ".licenses-license-tag.el-tag",
    ) as HTMLElement | null;
    const licenseText = (licenseCell?.textContent ?? "").trim();
    const tagStyle = licenseTag ? getComputedStyle(licenseTag) : null;

    const packageRight = packageName?.getBoundingClientRect().right ?? 0;
    const versionLeft = versionCell?.getBoundingClientRect().left ?? 0;

    return {
      ok: true as const,
      pageOverflows,
      wrappers: {
        overflowX: getComputedStyle(wrapper).overflowX,
        scrollable: wrapper.scrollWidth > wrapper.clientWidth,
        tableWidth: table.getBoundingClientRect().width,
        wrapperClientWidth: wrapper.clientWidth,
      },
      columnWidthDelta,
      packageOverlap: packageRight > versionLeft + 0.5,
      textChecks: {
        packageText: packageName?.textContent ?? "",
        packageWhiteSpace: packageName
          ? getComputedStyle(packageName).whiteSpace
          : "",
        repoWhiteSpace: repoLink ? getComputedStyle(repoLink).whiteSpace : "",
        licenseText,
        licenseTagWidth: licenseTag?.getBoundingClientRect().width ?? 0,
        licenseVisible: licenseText.length > 0,
        badge: tagStyle
          ? {
              innerText: (licenseTag?.textContent ?? "").trim(),
              color: tagStyle.color,
              backgroundColor: tagStyle.backgroundColor,
              opacity: tagStyle.opacity,
            }
          : null,
      },
    };
  });
}

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

  test("licenses table geometry at 768px", async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 800 });
    await page.goto("/#/licenses");
    await expect(
      page.getByTestId("licenses-table-scroll-frontend"),
    ).toBeVisible();

    const metrics = await measureLicensesTableGeometry(page);
    expect(metrics.ok, metrics.ok ? "" : metrics.reason).toBe(true);
    if (!metrics.ok) {
      return;
    }

    expect(metrics.pageOverflows).toBe(false);
    expect(metrics.wrappers.overflowX).toBe("auto");
    expect(metrics.wrappers.scrollable).toBe(true);
    expect(Math.max(...metrics.columnWidthDelta)).toBeLessThan(2);
    expect(metrics.packageOverlap).toBe(false);
    expect(metrics.textChecks.packageText).toContain(
      "@eslint-community/eslint-utils",
    );
    expect(metrics.textChecks.packageWhiteSpace).toBe("nowrap");
    expect(metrics.textChecks.repoWhiteSpace).toBe("nowrap");
    expect(metrics.textChecks.licenseVisible).toBe(true);
    expect(metrics.textChecks.licenseText).toContain("MIT");
    expect(metrics.textChecks.badge).not.toBeNull();
    expect(metrics.textChecks.badge!.innerText.length).toBeGreaterThan(0);
    expect(Number(metrics.textChecks.badge!.opacity)).toBe(1);
    expect(isOpaqueBackground(metrics.textChecks.badge!.backgroundColor)).toBe(
      true,
    );
    expect(
      colorsVisiblyDiffer(
        metrics.textChecks.badge!.color,
        metrics.textChecks.badge!.backgroundColor,
      ),
    ).toBe(true);
  });

  test("licenses table geometry at 1280px", async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 800 });
    await page.goto("/#/licenses");
    await expect(
      page.getByTestId("licenses-table-scroll-frontend"),
    ).toBeVisible();

    const metrics = await measureLicensesTableGeometry(page);
    expect(metrics.ok, metrics.ok ? "" : metrics.reason).toBe(true);
    if (!metrics.ok) {
      return;
    }

    expect(metrics.pageOverflows).toBe(false);
    expect(Math.max(...metrics.columnWidthDelta)).toBeLessThan(2);
    expect(metrics.packageOverlap).toBe(false);
    expect(metrics.textChecks.packageWhiteSpace).toBe("nowrap");
    expect(metrics.textChecks.repoWhiteSpace).toBe("nowrap");
    expect(metrics.textChecks.licenseVisible).toBe(true);
    expect(metrics.textChecks.badge).not.toBeNull();
    expect(metrics.textChecks.badge!.innerText.length).toBeGreaterThan(0);
    expect(Number(metrics.textChecks.badge!.opacity)).toBe(1);
    expect(isOpaqueBackground(metrics.textChecks.badge!.backgroundColor)).toBe(
      true,
    );
    expect(
      colorsVisiblyDiffer(
        metrics.textChecks.badge!.color,
        metrics.textChecks.badge!.backgroundColor,
      ),
    ).toBe(true);
    expect(metrics.wrappers.tableWidth).toBeGreaterThanOrEqual(
      metrics.wrappers.wrapperClientWidth - 1,
    );
  });
});
