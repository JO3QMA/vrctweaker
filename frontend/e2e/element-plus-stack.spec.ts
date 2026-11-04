import { test, expect } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";

const BRAND_FOCUS_RGB = "rgb(92, 156, 230)";

test.describe("Element Plus stack (App.vue path)", () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(
      getMockWailsInitScript({ startupAppUpdatePrompt: true }),
    );
  });

  test("AppUpdateDialog later button uses brand focus ring (computed outline)", async ({
    page,
  }) => {
    await page.goto("/");
    await expect(page.getByTestId("app-update-dialog")).toBeVisible();

    const later = page.getByTestId("app-update-later");
    const outline = await later.evaluate((el) => {
      const button = el as HTMLButtonElement;
      button.focus({ focusVisible: true });
      const cs = getComputedStyle(button);
      return {
        outlineWidth: cs.outlineWidth,
        outlineStyle: cs.outlineStyle,
        outlineColor: cs.outlineColor,
        outlineOffset: cs.outlineOffset,
      };
    });

    expect(outline.outlineWidth).toBe("2px");
    expect(outline.outlineStyle).toBe("solid");
    expect(outline.outlineColor).toBe(BRAND_FOCUS_RGB);
    expect(outline.outlineOffset).toBe("2px");
  });

  test("showToast renders ElMessage with z-index 4000 under el-config-provider", async ({
    page,
  }) => {
    await page.goto("/");
    await expect(page.getByTestId("app-update-dialog")).toBeVisible();

    await page.evaluate(() => {
      const app = window.go?.wailsapp?.App as
        { DismissAppUpdate?: (_version: string) => Promise<void> } | undefined;
      if (!app?.DismissAppUpdate) {
        throw new Error("wails app mock is missing DismissAppUpdate");
      }
      app.DismissAppUpdate = () =>
        Promise.reject(new Error("e2e dismiss failed"));
    });

    await page.getByTestId("app-update-later").click();
    const toast = page.locator(".el-message").last();
    await expect(toast).toBeVisible();

    await expect
      .poll(async () => toast.evaluate((el) => getComputedStyle(el).zIndex))
      .toBe("4000");
  });
});
