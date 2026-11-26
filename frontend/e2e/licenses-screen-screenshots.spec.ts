import { test } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";

for (const width of [1280, 768] as const) {
  test(`licenses screen ja @ ${width}x800 viewport screenshot`, async ({
    browser,
  }) => {
    const context = await browser.newContext({
      viewport: { width, height: 800 },
      locale: "ja-JP",
    });
    const page = await context.newPage();
    await page.addInitScript(getMockWailsInitScript());
    await page.goto("/#/licenses");
    await page.waitForSelector(".licenses-view");
    await page.screenshot({
      path: `/opt/cursor/artifacts/licenses-screen-ja-${width}px.png`,
    });
    await page.screenshot({
      path: `/opt/cursor/artifacts/licenses-screen-ja-${width}px-fullpage.png`,
      fullPage: true,
    });
    await context.close();
  });
}

test("licenses table scroll right edge @ 768px screenshot", async ({
  browser,
}) => {
  const context = await browser.newContext({
    viewport: { width: 768, height: 800 },
    locale: "ja-JP",
  });
  const page = await context.newPage();
  await page.addInitScript(getMockWailsInitScript());
  await page.goto("/#/licenses");
  const scroll = page.getByTestId("licenses-table-scroll-frontend");
  await scroll.evaluate((el) => {
    el.scrollLeft = el.scrollWidth;
  });
  await page.screenshot({
    path: "/opt/cursor/artifacts/licenses-screen-ja-768px-scrolled-right.png",
  });
  await context.close();
});
