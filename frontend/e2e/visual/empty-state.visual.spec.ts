import { test, expect } from "@playwright/test";
import { VISUAL_LOCALES, VISUAL_VIEWPORTS } from "./visual-routes";
import { VISUAL_EMPTY_SCENARIOS } from "./visual-scenarios";
import {
  captureAppLayoutScreenshot,
  installVisualTestContext,
  prepareVisualRoute,
} from "./visual-helpers";

for (const locale of VISUAL_LOCALES) {
  for (const viewport of VISUAL_VIEWPORTS) {
    for (const scenario of VISUAL_EMPTY_SCENARIOS) {
      test(`empty ${scenario.name} ${locale} ${viewport.label}`, async ({
        page,
      }) => {
        await installVisualTestContext(page, locale, scenario.mock);
        await prepareVisualRoute(page, scenario.path, viewport);
        if (scenario.readyTestId) {
          await expect(page.getByTestId(scenario.readyTestId)).toBeVisible();
        }
        const screenshotName = `empty-${scenario.name}-${locale}-${viewport.label}.png`;
        const maskRoute =
          scenario.name === "gallery" ? undefined : scenario.routeName;
        await captureAppLayoutScreenshot(page, screenshotName, maskRoute);
      });
    }
  }
}
