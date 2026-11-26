import { test } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";

const scheduledFor = new Date(2026, 11, 1, 10, 0, 0).toISOString();
const scheduledUntil = new Date(2026, 11, 1, 14, 0, 0).toISOString();

const initScript = getMockWailsInitScript({
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
  test(`server status maintenance card ja @ ${width}px`, async ({
    browser,
  }) => {
    const context = await browser.newContext({
      viewport: { width, height: 900 },
      timezoneId: "Asia/Tokyo",
      locale: "ja-JP",
    });
    const page = await context.newPage();
    await page.addInitScript({ content: initScript });
    await page.goto("/");
    await page.waitForSelector('[data-testid="server-status-section"]');
    await page.getByTestId("server-status-section").screenshot({
      path: `/opt/cursor/artifacts/server-status-maintenance-ja-${width}px.png`,
    });
    await context.close();
  });
}
