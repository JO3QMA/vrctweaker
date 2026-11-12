import type { Page, Locator } from "@playwright/test";
import { expect } from "@playwright/test";
import type { VisualRouteName } from "./visual-routes";
import { VISUAL_ROUTE_MASK_SELECTORS } from "./visual-routes";

export type CombineMaskOptions = {
  includeSidebarIcons?: boolean;
};

/**
 * サイドバーアイコン（SVG）は環境差でアンチエイリアスが揺れるためmaskする。
 * ナビのラベル・選択状態は`.sidebar`内のテキスト／背景で検証する。
 */
export function globalVisualMasks(page: Page): Locator[] {
  return [page.locator(".sidebar-icon")];
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

export function combineVisualMasks(
  page: Page,
  routeName?: VisualRouteName,
  extra: Locator[] = [],
  options: CombineMaskOptions = {},
): Locator[] {
  const includeSidebar = options.includeSidebarIcons !== false;
  const sidebar = includeSidebar ? globalVisualMasks(page) : [];
  const routeMasks = routeName ? maskLocatorsForRoute(page, routeName) : [];
  return [...sidebar, ...routeMasks, ...extra];
}

export async function assertMaskLocatorsMatch(
  masks: Locator[],
  context: string,
): Promise<void> {
  for (let i = 0; i < masks.length; i++) {
    const count = await masks[i].count();
    expect(
      count,
      `${context}: mask locator #${i + 1} must match at least one element`,
    ).toBeGreaterThan(0);
  }
}
