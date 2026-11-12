import { test, expect } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";

test.beforeEach(async ({ page }) => {
  await page.addInitScript(getMockWailsInitScript());
});

async function openLauncherAdvancedResolution(
  page: import("@playwright/test").Page,
) {
  await page.goto("/#/launcher");
  await page.getByTestId("profile-card-profile-1").click();
  const collapseHeader = page.locator(
    ".args-collapse .el-collapse-item__header",
  );
  await collapseHeader.click();
  await page.getByTestId("resolution-enabled-checkbox").click();
  await expect(page.getByTestId("resolution-enabled-checkbox")).toBeChecked();
  await page.getByTestId("resolution-preset-custom").click();
  await expect(page.getByTestId("screen-width-input")).toBeVisible();
}

async function typeResolution(
  page: import("@playwright/test").Page,
  width: number,
  height: number,
) {
  const widthInput = page.getByTestId("screen-width-input").locator("input");
  const heightInput = page.getByTestId("screen-height-input").locator("input");
  await widthInput.click();
  await widthInput.fill("");
  await widthInput.fill(String(width));
  await widthInput.blur();
  await heightInput.click();
  await heightInput.fill("");
  await heightInput.fill(String(height));
  await heightInput.blur();
}

for (const viewport of [
  { width: 1280, height: 800, label: "1280px" },
  { width: 600, height: 900, label: "600px" },
] as const) {
  test(`saves custom resolution 1920x720 and keeps UI state (${viewport.label})`, async ({
    page,
  }) => {
    await page.setViewportSize({
      width: viewport.width,
      height: viewport.height,
    });
    await openLauncherAdvancedResolution(page);
    await typeResolution(page, 1920, 720);
    await expect(page.getByTestId("unsaved-banner")).toBeVisible();
    await page.locator(".btn-save").click();
    const saved = await page.evaluate(() => {
      const row = (
        window as unknown as {
          __e2eLastSaveLaunchProfile?: { arguments?: string };
        }
      ).__e2eLastSaveLaunchProfile;
      return row ?? null;
    });
    expect(saved?.arguments ?? "").toMatch(/-screen-width\s+1920/);
    expect(saved?.arguments ?? "").toMatch(/-screen-height\s+720/);
    await expect(page.getByTestId("unsaved-banner")).toHaveCount(0);
    await expect(page.getByTestId("resolution-enabled-checkbox")).toBeChecked();
    await expect(page.getByTestId("screen-width-input")).toBeVisible();
    await expect(
      page.getByTestId("screen-width-input").locator("input"),
    ).toHaveValue("1920");
    await expect(
      page.getByTestId("screen-height-input").locator("input"),
    ).toHaveValue("720");

    await page.getByTestId("profile-card-profile-2").click();
    await page.getByTestId("profile-card-profile-1").click();
    await page.locator(".args-collapse .el-collapse-item__header").click();
    await expect(page.getByTestId("resolution-enabled-checkbox")).toBeChecked();
    await expect(
      page.getByTestId("screen-width-input").locator("input"),
    ).toHaveValue("1920");
    await expect(
      page.getByTestId("screen-height-input").locator("input"),
    ).toHaveValue("720");
  });
}
