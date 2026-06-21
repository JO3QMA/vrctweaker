import { test, expect } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";

test.beforeEach(async ({ page }) => {
  await page.addInitScript(getMockWailsInitScript());
});

test.describe("Config", () => {
  test("shows editor when VRChat config exists and camera preset interaction", async ({
    page,
  }) => {
    await page.goto("/#/config");
    await expect(page.locator("h1")).toContainText("その他の設定");

    // mock: VRChatConfigExists → true → editor mode with save button
    await expect(page.getByTestId("save-config-btn")).toBeVisible();

    await page.getByTestId("camera-preset-fhd").click();
    await expect(page.getByTestId("picture-output-folder-input")).toBeVisible();
  });
});
