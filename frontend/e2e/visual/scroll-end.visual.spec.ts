import { test } from "@playwright/test";
import { VISUAL_LOCALES, VISUAL_VIEWPORTS } from "./visual-routes";
import {
  VISUAL_SCROLL_END_SCENARIOS,
  VISUAL_SCROLL_END_SKIP,
} from "./visual-scenarios";
import {
  assertScrollerAtEnd,
  captureAppLayoutScreenshot,
  installVisualTestContext,
  prepareVisualRoute,
  scrollRegionToEnd,
} from "./visual-helpers";

for (const locale of VISUAL_LOCALES) {
  for (const viewport of VISUAL_VIEWPORTS) {
    for (const scenario of VISUAL_SCROLL_END_SCENARIOS) {
      const skipped = VISUAL_SCROLL_END_SKIP.find(
        (s) => s.name === scenario.name && s.viewportLabel === viewport.label,
      );
      if (skipped) {
        test.skip(`scroll-end ${scenario.name} ${locale} ${viewport.label}`, () => {
          /* ${skipped.reason} */
        });
        continue;
      }
      test(`scroll-end ${scenario.name} ${locale} ${viewport.label}`, async ({
        page,
      }) => {
        await installVisualTestContext(page, locale);
        await prepareVisualRoute(page, scenario.path, viewport);
        await scrollRegionToEnd(page, scenario.scroller);
        await assertScrollerAtEnd(page, scenario.scroller);
        const screenshotName = `scroll-end-${scenario.name}-${locale}-${viewport.label}.png`;
        await captureAppLayoutScreenshot(page, screenshotName, scenario.name);
      });
    }
  }
}
