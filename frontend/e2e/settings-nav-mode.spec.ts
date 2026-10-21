import { test, expect, type Page } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";

const SUBPAGES = [
  "/settings/account",
  "/settings/general",
  "/settings/data",
  "/settings/about",
  "/settings/reset",
] as const;

/** Targets for the settings frame (not the window). 720px is the side-nav threshold. */
const FRAME_WIDTHS = [700, 720, 740] as const;

test.beforeEach(async ({ page }) => {
  await page.addInitScript(getMockWailsInitScript());
});

async function setSettingsFrameWidth(
  page: Page,
  target: number,
): Promise<void> {
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.goto("/#/settings/account");
  await page.locator(".settings-frame").waitFor();
  const current = await page
    .locator(".settings-frame")
    .evaluate((el) => el.clientWidth);
  const viewport = page.viewportSize();
  if (!viewport) {
    throw new Error("viewport is unset");
  }
  await page.setViewportSize({
    width: viewport.width + (target - current),
    height: 800,
  });
  await expect
    .poll(() =>
      page.locator(".settings-frame").evaluate((el) => el.clientWidth),
    )
    .toBe(target);
}

test.describe("Settings nav mode", () => {
  for (const target of FRAME_WIDTHS) {
    test(`every sub-page uses the same nav at frame width ${target}px`, async ({
      page,
    }) => {
      await setSettingsFrameWidth(page, target);
      const readings: {
        path: string;
        mode: string;
        width: number;
        scrolls: boolean;
      }[] = [];

      for (const path of SUBPAGES) {
        await page.goto(`/#${path}`);
        await expect(page.locator(".settings-nav")).toBeVisible();
        const reading = await page.locator(".settings-nav").evaluate((el) => {
          const frame = el.closest(".settings-frame");
          const scroller = el.closest(".settings-view");
          if (
            !(frame instanceof HTMLElement) ||
            !(scroller instanceof HTMLElement)
          ) {
            throw new Error("settings frame or scroller is missing");
          }
          const direction = getComputedStyle(el).flexDirection;
          return {
            mode: direction === "column" ? "side" : "tabs",
            width: frame.clientWidth,
            scrolls: scroller.scrollHeight > scroller.clientHeight + 1,
          };
        });
        readings.push({ path, ...reading });
      }

      const modes = new Set(readings.map((reading) => reading.mode));
      expect(modes, JSON.stringify(readings)).toEqual(
        new Set([target < 720 ? "tabs" : "side"]),
      );
      const widths = readings.map((reading) => reading.width);
      expect(Math.max(...widths) - Math.min(...widths)).toBeLessThanOrEqual(1);
      expect(widths.every((width) => width === target)).toBe(true);
      expect(readings.some((reading) => reading.scrolls)).toBe(true);
      expect(readings.some((reading) => !reading.scrolls)).toBe(true);
    });
  }
});
