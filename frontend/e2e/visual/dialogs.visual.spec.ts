import { test } from "@playwright/test";
import { VISUAL_LOCALES, VISUAL_VIEWPORTS } from "./visual-routes";
import { VISUAL_DIALOG_SCENARIOS } from "./visual-scenarios";
import {
  captureAppLayoutScreenshot,
  captureViewportScreenshot,
  installVisualTestContext,
  openLegalDocumentDialog,
  prepareAppUpdateDialog,
  prepareOnboardingWizard,
} from "./visual-helpers";

const VISUAL_FONT_CSS = `
:root {
  --font-family-ui: "Noto Sans CJK JP", "Noto Sans JP", "Segoe UI", system-ui, sans-serif;
}
`;

for (const locale of VISUAL_LOCALES) {
  for (const viewport of VISUAL_VIEWPORTS) {
    for (const dialog of VISUAL_DIALOG_SCENARIOS) {
      test(`dialog ${dialog.id} ${locale} ${viewport.label}`, async ({
        page,
      }) => {
        const screenshotName = `dialog-${dialog.id}-${locale}-${viewport.label}.png`;

        if (dialog.id === "legal-document") {
          await installVisualTestContext(page, locale);
          await page.setViewportSize(viewport);
          await page.addStyleTag({ content: VISUAL_FONT_CSS });
          await openLegalDocumentDialog(page);
          await captureViewportScreenshot(page, screenshotName);
          return;
        }

        if (dialog.id === "app-update") {
          await installVisualTestContext(page, locale, {
            startupAppUpdatePrompt: true,
          });
          await prepareAppUpdateDialog(page, viewport);
          await captureViewportScreenshot(page, screenshotName);
          return;
        }

        await installVisualTestContext(page, locale, {
          firstLaunchGuideCompleted: false,
        });
        await prepareOnboardingWizard(page, viewport);
        await captureAppLayoutScreenshot(page, screenshotName);
      });
    }
  }
}
