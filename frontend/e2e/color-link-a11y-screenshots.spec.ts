import { test, expect } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";

const scheduledFor = new Date(2026, 11, 1, 10, 0, 0).toISOString();
const scheduledUntil = new Date(2026, 11, 1, 14, 0, 0).toISOString();

/** Same maintenance payload as `server-status-maintenance-screenshots.spec.ts` (#409). */
const serverStatusInit = getMockWailsInitScript({
  serverStatus: {
    fetchState: "ok",
    maintenances: [
      {
        id: "mapped",
        name: "Maintenance Window",
        status: "scheduled",
        scheduledFor,
        scheduledUntil,
        updates: ["No service impact expected during this window."],
      },
      {
        id: "unknown",
        name: "Partner API rollout",
        status: "scheduled",
        scheduledFor,
        scheduledUntil,
        updates: ["Updates stay in the tooltip."],
      },
    ],
  },
});

for (const width of [1280, 768] as const) {
  test(`server status card with link tokens @ ${width}px`, async ({
    browser,
  }) => {
    const context = await browser.newContext({
      viewport: { width, height: 900 },
      timezoneId: "Asia/Tokyo",
      locale: "ja-JP",
    });
    const page = await context.newPage();
    await page.addInitScript({ content: serverStatusInit });
    await page.goto("/");
    const lines = page.getByTestId("server-status-maintenance-line");
    await expect(lines).toHaveCount(2);
    await expect(lines.nth(0)).toContainText("(メンテ予告)");
    await page.getByTestId("server-status-section").screenshot({
      path: `/opt/cursor/artifacts/server-status-link-a11y-${width}px.png`,
    });
    await context.close();
  });

  test(`licenses page with link tokens @ ${width}px`, async ({ browser }) => {
    const context = await browser.newContext({
      viewport: { width, height: 900 },
      locale: "ja-JP",
    });
    const page = await context.newPage();
    await page.addInitScript({ content: getMockWailsInitScript() });
    await page.goto("/#/licenses");
    await page.waitForSelector(".licenses-view");
    await page.locator(".licenses-view").screenshot({
      path: `/opt/cursor/artifacts/licenses-link-a11y-${width}px.png`,
    });
    await context.close();
  });
}
