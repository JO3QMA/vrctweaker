import type { Page, Locator } from "@playwright/test";
import { getMockWailsInitScript } from "../fixtures/mock-wails";
import {
  VISUAL_FIXED_CLOCK_ISO,
  type VisualRouteName,
  VISUAL_ROUTE_MASK_SELECTORS,
} from "./visual-routes";

const VISUAL_FONT_CSS = `
:root {
  --font-family-ui: "Noto Sans CJK JP", "Noto Sans JP", "Segoe UI", system-ui, sans-serif;
}
`;

export async function installVisualTestContext(
  page: Page,
  locale: "ja" | "en",
): Promise<void> {
  await page.addInitScript(getMockWailsInitScript({ language: locale }));
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
  // ルート遷移フェードの完了を待つ
  await page.waitForTimeout(150);
}

export function maskLocatorsForRoute(
  page: Page,
  routeName: VisualRouteName,
): Locator[] {
  const selectors = VISUAL_ROUTE_MASK_SELECTORS[routeName];
  if (!selectors?.length) {
    return [];
  }
  return selectors.map((sel) => page.locator(sel));
}
