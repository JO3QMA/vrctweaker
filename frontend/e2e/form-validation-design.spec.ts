import { test, expect } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";

test.beforeEach(async ({ page }) => {
  await page.addInitScript(getMockWailsInitScript());
});

async function openLauncherResolution(page: import("@playwright/test").Page) {
  await page.goto("/#/launcher");
  await expect(page.locator("h1")).toContainText("ランチャー");
  await page.getByTestId("profile-card-profile-1").click();
  await page.locator(".args-collapse .el-collapse-item__header").click();
  await page.getByTestId("resolution-enabled-checkbox").click();
  await page.getByTestId("resolution-preset-custom").click();
  await expect(page.getByTestId("screen-width-input")).toBeVisible();
}

test.describe("Form validation layout (design QA)", () => {
  test("launcher resolution hint uses grid below inputs at 1280px", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1280, height: 800 });
    await openLauncherResolution(page);

    const content = page
      .locator(".vt-form-item-with-range-hint .el-form-item__content")
      .first();
    await expect(content).toBeVisible();
    expect(await content.evaluate((el) => getComputedStyle(el).display)).toBe(
      "grid",
    );

    const hintBelow = await page
      .locator(".resolution-form-item")
      .first()
      .evaluate((item) => {
        const input = item.querySelector(".el-input-number")!;
        const hint = item.querySelector(".form-field-range-hint")!;
        return (
          hint.getBoundingClientRect().top >=
          input.getBoundingClientRect().bottom - 1
        );
      });
    expect(hintBelow).toBe(true);
  });

  test("launcher resolution row uses flex-wrap at 600px", async ({ page }) => {
    await page.setViewportSize({ width: 600, height: 900 });
    await openLauncherResolution(page);

    const row = page.locator(".resolution-fields");
    expect(await row.evaluate((el) => getComputedStyle(el).flexWrap)).toBe(
      "wrap",
    );
    const rowWidth = await row.evaluate(
      (el) => el.getBoundingClientRect().width,
    );
    expect(rowWidth).toBeLessThanOrEqual(600);
  });

  test("automation editor scroll padding follows sticky action bar inset at 600px", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 600, height: 800 });
    await page.goto("/#/automation");
    await page.getByTestId("add-item").click();
    await page.getByTestId("add-rule").click();
    await expect(page.locator(".rule-editor")).toBeVisible();

    await expect(page.locator(".automation-editor__scroll")).toBeVisible();
    const { paddingBottom, insetPx } = await page.evaluate(() => {
      const scrollEl = document.querySelector(
        ".automation-editor__scroll",
      ) as HTMLElement | null;
      const root = document.querySelector(
        ".automation-rule-editor",
      ) as HTMLElement | null;
      if (!scrollEl || !root) {
        return { paddingBottom: 0, insetPx: 0 };
      }
      return {
        paddingBottom: parseFloat(getComputedStyle(scrollEl).paddingBottom),
        insetPx: parseFloat(
          getComputedStyle(root).getPropertyValue(
            "--automation-action-bar-inset",
          ),
        ),
      };
    });
    expect(insetPx).toBeGreaterThanOrEqual(80);
    expect(paddingBottom).toBeGreaterThanOrEqual(80);
    expect(paddingBottom).toBe(insetPx);
  });

  test("resolution width inner input receives aria-disabled when preset is not custom", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1280, height: 800 });
    await openLauncherResolution(page);
    await page.getByTestId("resolution-preset-fhd").click();

    const inner = page.getByTestId("screen-width-input").locator("input");
    await expect(inner).toBeDisabled();
    await expect(inner).toHaveAttribute("aria-disabled", "true");
  });

  test("automation schedule inline form items use 22px bottom margin at 600px", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 600, height: 800 });
    await page.goto("/#/automation");
    await page.getByTestId("add-item").click();
    await page.getByTestId("add-rule").click();
    await expect(page.locator(".rule-editor")).toBeVisible();
    await page.getByTestId("trigger-type").click();
    await page
      .locator(".el-select-dropdown__item")
      .filter({ hasText: "スケジュール" })
      .click();

    const margins = await page.evaluate(() => {
      const items = Array.from(
        document.querySelectorAll(
          ".automation-rule-editor .el-form-item.automation-inline-form-item",
        ),
      );
      return items.map((el) => getComputedStyle(el).marginBottom);
    });
    expect(margins.length).toBeGreaterThanOrEqual(2);
    for (const marginBottom of margins) {
      expect(marginBottom).toBe("22px");
    }
  });
});
