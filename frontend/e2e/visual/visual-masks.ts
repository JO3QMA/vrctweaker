import type { Page, Locator } from "@playwright/test";
import type { VisualRouteName } from "./visual-routes";
import { VISUAL_ROUTE_MASK_SELECTORS } from "./visual-routes";

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
): Locator[] {
  const routeMasks = routeName ? maskLocatorsForRoute(page, routeName) : [];
  return [...globalVisualMasks(page), ...routeMasks, ...extra];
}
