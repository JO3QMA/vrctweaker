import type { Page, Locator } from "@playwright/test";
import { expect } from "@playwright/test";
import {
  getMockWailsInitScript,
  type MockWailsOptions,
} from "../fixtures/mock-wails";
import { VISUAL_FIXED_CLOCK_ISO, type VisualRouteName } from "./visual-routes";
import {
  assertMaskLocatorsMatch,
  combineVisualMasks,
  type CombineMaskOptions,
} from "./visual-masks";

const VISUAL_FONT_CSS = `
:root {
  --font-family-ui: "Noto Sans CJK JP", "Noto Sans JP", "Segoe UI", system-ui, sans-serif;
}
`;

export type VisualMockOptions = Omit<MockWailsOptions, "language">;

/**
 * PlayTimeChartの軸・凡例はcanvas内描画のためmaskしない。
 * devicePixelRatio=1・stableActivityChart・セクション展開後に描画完了を待つ。
 */
async function prepareActivityPlaytimeForVisual(page: Page): Promise<void> {
  const playtime = page.locator(".section-card--playtime");
  await playtime.waitFor({ state: "visible" });
  await expect(playtime.locator(".loading")).toBeHidden({ timeout: 15_000 });
  if (await playtime.locator(".empty-stats").isVisible()) {
    return;
  }
  const toggle = playtime.locator(".section-card__toggle");
  if ((await toggle.getAttribute("aria-expanded")) === "false") {
    await toggle.click();
  }
  const canvas = playtime.locator("canvas");
  await expect(canvas).toBeVisible();
  await page.waitForFunction(() => {
    const c = document.querySelector(
      ".section-card--playtime canvas",
    ) as HTMLCanvasElement | null;
    return Boolean(c && c.clientWidth > 0 && c.height > 0);
  });
  await page.waitForTimeout(200);
  const main = page.locator(".activity-view.main-scroll-region");
  await main.evaluate((el) => {
    el.scrollTop = 0;
  });
  await page.waitForTimeout(100);
}

async function waitForRouteReady(page: Page, routePath: string): Promise<void> {
  if (routePath.includes("/activity")) {
    await prepareActivityPlaytimeForVisual(page);
  }
  if (routePath.includes("/gallery")) {
    const empty = page.getByTestId("gallery-empty");
    const grid = page.getByTestId("gallery-grid-scroll");
    await expect(grid.or(empty)).toBeVisible({ timeout: 15_000 });
    if (await empty.isVisible()) {
      return;
    }
    await page.waitForFunction(() => {
      const thumbs = document.querySelectorAll(".thumbnail");
      if (thumbs.length === 0) {
        return false;
      }
      return Array.from(thumbs).every(
        (img) => (img as HTMLImageElement).complete,
      );
    });
  }
  if (routePath.includes("/automation")) {
    await page.locator(".automation-list__rows").waitFor({ state: "visible" });
  }
  if (routePath.includes("/me")) {
    await page
      .locator(".self-profile-view .friend-detail-card .el-card__body")
      .waitFor({ state: "visible" });
  }
  if (routePath.includes("/friends")) {
    await page.locator(".friends-pane--left .friends-list").waitFor({
      state: "visible",
    });
  }
}

export async function installVisualTestContext(
  page: Page,
  locale: "ja" | "en",
  extra: VisualMockOptions = {},
): Promise<void> {
  await page.addInitScript(() => {
    Object.defineProperty(window, "devicePixelRatio", {
      get: () => 1,
      configurable: true,
    });
  });
  await page.addInitScript(
    getMockWailsInitScript({
      language: locale,
      stableActivityChart: true,
      visualRichScrollContent: true,
      loggedIn: true,
      ...extra,
    }),
  );
  await page.clock.install({ time: new Date(VISUAL_FIXED_CLOCK_ISO) });
}

export async function prepareVisualRoute(
  page: Page,
  routePath: string,
  viewport: { width: number; height: number },
): Promise<void> {
  await page.setViewportSize(viewport);
  await page.addStyleTag({ content: VISUAL_FONT_CSS });
  await page.goto(routePath);
  await page.locator(".app-layout").waitFor({ state: "visible" });
  await page.locator("h1.page-title, h1").first().waitFor({ state: "visible" });
  await waitForRouteReady(page, routePath);
  await page.waitForTimeout(150);
}

export async function scrollRegionToEnd(
  page: Page,
  scrollerSelector: string,
): Promise<void> {
  const scroller = page.locator(scrollerSelector);
  await scroller.waitFor({ state: "visible" });
  await scroller.evaluate((el) => {
    el.scrollTop = Math.max(0, el.scrollHeight - el.clientHeight);
  });
  await page.waitForTimeout(150);
}

export async function assertScrollerAtEnd(
  page: Page,
  scrollerSelector: string,
): Promise<void> {
  const metrics = await page.locator(scrollerSelector).evaluate((el) => ({
    scrollHeight: el.scrollHeight,
    clientHeight: el.clientHeight,
    scrollTop: el.scrollTop,
  }));
  expect(metrics.scrollHeight).toBeGreaterThan(metrics.clientHeight);
  expect(metrics.scrollTop).toBeGreaterThan(0);
}

export async function openLegalDocumentDialog(page: Page): Promise<void> {
  await page.goto("/#/settings/reset");
  await page.getByTestId("settings-reset-setup-section").waitFor({
    state: "visible",
  });
  await page.getByTestId("settings-open-onboarding").click();
  await page.getByTestId("onboarding-wizard").waitFor({ state: "visible" });
  await page.getByTestId("legal-link-privacy").click();
  await page.getByTestId("legal-document-dialog").waitFor({ state: "visible" });
}

export async function prepareOnboardingWizard(
  page: Page,
  viewport: { width: number; height: number },
): Promise<void> {
  await page.setViewportSize(viewport);
  await page.addStyleTag({ content: VISUAL_FONT_CSS });
  await page.goto("/#/");
  await page.getByTestId("onboarding-wizard").waitFor({ state: "visible" });
  await page.waitForTimeout(150);
}

export async function prepareAppUpdateDialog(
  page: Page,
  viewport: { width: number; height: number },
): Promise<void> {
  await page.setViewportSize(viewport);
  await page.addStyleTag({ content: VISUAL_FONT_CSS });
  await page.goto("/#/");
  await page.locator(".app-layout").waitFor({ state: "visible" });
  await page.getByTestId("app-update-dialog").waitFor({ state: "visible" });
  await page.waitForTimeout(150);
}

export async function captureAppLayoutScreenshot(
  page: Page,
  screenshotName: string,
  routeName?: VisualRouteName,
  extraMasks: Locator[] = [],
  maskOptions?: CombineMaskOptions,
): Promise<void> {
  const masks = combineVisualMasks(page, routeName, extraMasks, maskOptions);
  await assertMaskLocatorsMatch(masks, screenshotName);
  await expect(page.locator(".app-layout")).toHaveScreenshot(screenshotName, {
    mask: masks,
  });
}

/** append-to-body の `el-dialog` を含める */
export async function captureViewportScreenshot(
  page: Page,
  screenshotName: string,
  extraMasks: Locator[] = [],
  maskOptions?: CombineMaskOptions,
): Promise<void> {
  const masks = combineVisualMasks(page, undefined, extraMasks, maskOptions);
  if (masks.length > 0) {
    await assertMaskLocatorsMatch(masks, screenshotName);
  }
  await expect(page).toHaveScreenshot(screenshotName, {
    mask: masks,
  });
}
