import type { Page, Locator } from "@playwright/test";
import { expect } from "@playwright/test";
import {
  getMockWailsInitScript,
  type MockWailsOptions,
} from "../fixtures/mock-wails";
import { VISUAL_FIXED_CLOCK_ISO, type VisualRouteName } from "./visual-routes";
import { combineVisualMasks } from "./visual-masks";

const VISUAL_FONT_CSS = `
:root {
  --font-family-ui: "Noto Sans CJK JP", "Noto Sans JP", "Segoe UI", system-ui, sans-serif;
}
`;

export type VisualMockOptions = Omit<MockWailsOptions, "language">;

async function waitForRouteReady(page: Page, routePath: string): Promise<void> {
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
}

export async function installVisualTestContext(
  page: Page,
  locale: "ja" | "en",
  extra: VisualMockOptions = {},
): Promise<void> {
  await page.addInitScript(
    getMockWailsInitScript({ language: locale, ...extra }),
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
): Promise<void> {
  const masks = combineVisualMasks(page, routeName, extraMasks);
  await expect(page.locator(".app-layout")).toHaveScreenshot(screenshotName, {
    mask: masks,
  });
}

/** append-to-body の `el-dialog` を含める */
export async function captureViewportScreenshot(
  page: Page,
  screenshotName: string,
  extraMasks: Locator[] = [],
): Promise<void> {
  const masks = combineVisualMasks(page, undefined, extraMasks);
  await expect(page).toHaveScreenshot(screenshotName, {
    mask: masks,
  });
}
