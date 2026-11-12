import { test, expect } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";

const BRAND_FOCUS_RGB = "rgb(92, 156, 230)";

test.describe("Element Plus stack (App.vue path)", () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(
      getMockWailsInitScript({ startupAppUpdatePrompt: true }),
    );
  });

  test("ElMessage toast loads theme styles (fixed position, non-transparent surface)", async ({
    page,
  }) => {
    await page.addInitScript(getMockWailsInitScript());
    await page.goto("/");
    await page.evaluate(async () => {
      const { showToast } = await import("/src/utils/showToast.ts");
      showToast.success("e2e style probe");
    });
    const toast = page.locator(".el-message").first();
    await expect(toast).toBeVisible();
    await expect(toast).toHaveCSS("position", "fixed");
    const bg = await toast.evaluate(
      (el) => getComputedStyle(el).backgroundColor,
    );
    expect(bg).not.toBe("rgba(0, 0, 0, 0)");
    expect(bg).not.toBe("transparent");
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

  for (const viewportWidth of [1280, 600] as const) {
    test(`all concurrent toasts stay at z-index 4000 above a later dialog (viewport ${viewportWidth}px)`, async ({
      page,
    }) => {
      await page.setViewportSize({ width: viewportWidth, height: 720 });
      await page.addInitScript(getMockWailsInitScript());
      await page.goto("/");
      await expect(page.locator("h1")).toContainText("ダッシュボード");

      await page.evaluate(async () => {
        const { showToast } = await import("/src/utils/showToast.ts");
        showToast.success("e2e toast 1");
        showToast.warning("e2e toast 2");
        showToast.error("e2e toast 3");
      });

      const toasts = page.locator(".el-message");
      await expect(toasts).toHaveCount(3);

      await expect
        .poll(async () => {
          const zIndexes = await toasts.evaluateAll((nodes) =>
            nodes.map((el) => getComputedStyle(el).zIndex),
          );
          return zIndexes.every((z) => z === "4000");
        })
        .toBe(true);

      await page.evaluate(() => {
        const emit = (
          window as Window & {
            __e2eEmitWailsEvent?: (name: string, data: unknown) => void;
          }
        ).__e2eEmitWailsEvent;
        if (!emit) {
          throw new Error("e2e wails event emitter is missing");
        }
        emit("app:update-available", {
          currentVersion: "0.1.0",
          updateAvailable: true,
          shouldPrompt: true,
          latestVersion: "0.2.0",
          latestTag: "v0.2.0",
          releaseNotes: "E2E",
          downloadUrl: "",
          zipSha256: "",
          releasePageUrl: "https://example.com/releases",
          autoApplySupported: true,
        });
      });
      await expect(page.getByTestId("app-update-dialog")).toBeVisible();

      await expect
        .poll(async () => {
          const zIndexes = await toasts.evaluateAll((nodes) =>
            nodes.map((el) => getComputedStyle(el).zIndex),
          );
          return zIndexes.every((z) => z === "4000");
        })
        .toBe(true);

      const stack = await page.evaluate(() => {
        const overlayZ = Math.max(
          0,
          ...[...document.querySelectorAll(".el-overlay")].map(
            (el) => parseInt(getComputedStyle(el).zIndex, 10) || 0,
          ),
        );
        const dialogZ = Math.max(
          0,
          ...[...document.querySelectorAll(".el-dialog")].map(
            (el) => parseInt(getComputedStyle(el).zIndex, 10) || 0,
          ),
        );
        const modalTop = Math.max(overlayZ, dialogZ);
        return {
          modalTop,
          toasts: [...document.querySelectorAll(".el-message")].map((el) => {
            const z = parseInt(getComputedStyle(el).zIndex, 10) || 0;
            return {
              zIndex: getComputedStyle(el).zIndex,
              aboveModal: z > modalTop,
            };
          }),
        };
      });
      expect(stack.toasts).toHaveLength(3);
      for (const toast of stack.toasts) {
        expect(toast.zIndex).toBe("4000");
        expect(toast.aboveModal).toBe(true);
      }
    });
  }
});
