import { test } from "@playwright/test";
import {
  VISUAL_LOCALES,
  VISUAL_ROUTES,
  VISUAL_VIEWPORTS,
} from "./visual-routes";
import {
  captureAppLayoutScreenshot,
  installVisualTestContext,
  prepareVisualRoute,
} from "./visual-helpers";

for (const locale of VISUAL_LOCALES) {
  for (const viewport of VISUAL_VIEWPORTS) {
    for (const route of VISUAL_ROUTES) {
      test(`${route.name} ${locale} ${viewport.label}`, async ({ page }) => {
        await installVisualTestContext(page, locale);
        await prepareVisualRoute(page, route.path, viewport);
        const screenshotName = `route-${route.name}-${locale}-${viewport.label}.png`;
        await captureAppLayoutScreenshot(page, screenshotName, route.name);
      });
    }
  }
}
