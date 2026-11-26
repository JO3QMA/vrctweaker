import { test } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";

const serverStatusInit = getMockWailsInitScript({
  serverStatus: {
    fetchState: "ok",
    summary: { indicator: "none", description: "All Systems Operational" },
    maintenances: [
      {
        id: "link-demo",
        name: "Maintenance Window",
        status: "scheduled",
        scheduledFor: new Date(2026, 11, 1, 10, 0, 0).toISOString(),
        scheduledUntil: new Date(2026, 11, 1, 14, 0, 0).toISOString(),
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
    await page.waitForSelector('[data-testid="server-status-section"]');
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
