import { test, expect } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";

const FIELD_GAP_PX = 8;
const GAP_TOLERANCE_PX = 1;

test.beforeEach(async ({ page }) => {
  await page.addInitScript(getMockWailsInitScript());
});

function expectGapPx(actual: number, expected = FIELD_GAP_PX) {
  expect(actual).toBeGreaterThanOrEqual(expected - GAP_TOLERANCE_PX);
  expect(actual).toBeLessThanOrEqual(expected + GAP_TOLERANCE_PX);
}

/** After error margin, the next field should sit flush (margin supplies the 8px band). */
function expectFlushAfterErrorMargin(actual: number) {
  expect(actual).toBeGreaterThanOrEqual(-3);
  expect(actual).toBeLessThanOrEqual(3);
}

async function openLauncherResolution(page: import("@playwright/test").Page) {
  await page.goto("/#/launcher");
  await expect(page.locator("h1")).toContainText("ランチャー");
  await page.getByTestId("profile-card-profile-1").click();
  await page.locator(".args-collapse .el-collapse-item__header").click();
  await page.getByTestId("resolution-enabled-checkbox").click();
  await page.getByTestId("resolution-preset-custom").click();
  await expect(page.getByTestId("screen-width-input")).toBeVisible();
}

async function openAutomationScheduleEditor(
  page: import("@playwright/test").Page,
) {
  await page.goto("/#/automation");
  await page.getByTestId("add-item").click();
  await page.getByTestId("add-rule").click();
  await expect(page.locator(".rule-editor")).toBeVisible();
  await page.getByTestId("trigger-type").click();
  await page
    .locator(".el-select-dropdown__item")
    .filter({ hasText: "スケジュール" })
    .click();
}

async function uncheckAllScheduleWeekdays(
  page: import("@playwright/test").Page,
) {
  const group = page
    .getByTestId("automation-section-when")
    .locator(".el-checkbox-group");
  while ((await group.locator(".el-checkbox.is-checked").count()) > 0) {
    await group.locator(".el-checkbox.is-checked").first().click();
  }
}

async function measureVerticalGap(
  page: import("@playwright/test").Page,
  bottomLocator: import("@playwright/test").Locator,
  topLocator: import("@playwright/test").Locator,
): Promise<number> {
  const bottomHandle = await bottomLocator.elementHandle();
  const topHandle = await topLocator.elementHandle();
  expect(bottomHandle).toBeTruthy();
  expect(topHandle).toBeTruthy();
  return page.evaluate(
    ([bottom, top]) => {
      const b = (bottom as HTMLElement).getBoundingClientRect();
      const t = (top as HTMLElement).getBoundingClientRect();
      return t.top - b.bottom;
    },
    [bottomHandle!, topHandle!],
  );
}

/** Error line box includes extra descent; include the error's margin in the gap. */
async function measureVerticalGapAfterFormError(
  page: import("@playwright/test").Page,
  formItem: import("@playwright/test").Locator,
  topLocator: import("@playwright/test").Locator,
): Promise<number> {
  const itemHandle = await formItem.elementHandle();
  const topHandle = await topLocator.elementHandle();
  expect(itemHandle).toBeTruthy();
  expect(topHandle).toBeTruthy();
  return page.evaluate(
    ([item, top]) => {
      const err = (item as HTMLElement).querySelector(
        ".el-form-item__error",
      ) as HTMLElement | null;
      if (!err) {
        return -1;
      }
      const marginBottom = parseFloat(getComputedStyle(err).marginBottom) || 0;
      const from = err.getBoundingClientRect().bottom + marginBottom;
      return (top as HTMLElement).getBoundingClientRect().top - from;
    },
    [itemHandle!, topHandle!],
  );
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

  for (const viewport of [
    { width: 1280, height: 800, name: "1280px" },
    { width: 600, height: 900, name: "600px" },
  ] as const) {
    test(`launcher leaves ${FIELD_GAP_PX}px below resolution hint before next row at ${viewport.name}`, async ({
      page,
    }) => {
      await page.setViewportSize({
        width: viewport.width,
        height: viewport.height,
      });
      await openLauncherResolution(page);

      const gap = await page.evaluate(() => {
        const hints = Array.from(
          document.querySelectorAll(
            ".launch-args-advanced .sub-options .form-field-range-hint",
          ),
        );
        const monitorRow = document
          .querySelector('[data-testid="monitor-enabled-checkbox"]')
          ?.closest(".arg-row");
        if (hints.length === 0 || !monitorRow) return -1;
        const bottom = Math.max(
          ...hints.map((h) => h.getBoundingClientRect().bottom),
        );
        return monitorRow.getBoundingClientRect().top - bottom;
      });
      expect(gap).toBeGreaterThanOrEqual(0);
      expectGapPx(gap);
    });
  }

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

  for (const viewport of [
    { width: 1280, height: 800, name: "1280px" },
    { width: 600, height: 900, name: "600px" },
  ] as const) {
    test(`automation schedule weekday error to time row is ${FIELD_GAP_PX}px at ${viewport.name}`, async ({
      page,
    }) => {
      await page.setViewportSize({
        width: viewport.width,
        height: viewport.height,
      });
      await openAutomationScheduleEditor(page);
      await page.getByTestId("item-name-input").fill("E2E spacing");
      await uncheckAllScheduleWeekdays(page);
      await page.getByTestId("save-item").click();

      const weekdayItem = page
        .locator(".automation-inline-form-item")
        .filter({ has: page.locator(".el-checkbox-group") });
      await expect(weekdayItem).toHaveClass(/is-error/);
      await expect(weekdayItem.locator(".el-form-item__error")).toBeVisible();

      const timeRow = page
        .getByTestId("automation-section-when")
        .locator(".setting-row")
        .filter({ has: page.getByTestId("schedule-hour-input") });

      const gap = await measureVerticalGapAfterFormError(
        page,
        weekdayItem,
        timeRow,
      );
      expectFlushAfterErrorMargin(gap);
      await expect(weekdayItem.locator(".el-form-item__error")).toHaveCSS(
        "margin-bottom",
        `${FIELD_GAP_PX}px`,
      );
    });

    test(`automation schedule weekday control to time row is ${FIELD_GAP_PX}px without error at ${viewport.name}`, async ({
      page,
    }) => {
      await page.setViewportSize({
        width: viewport.width,
        height: viewport.height,
      });
      await openAutomationScheduleEditor(page);
      await page
        .getByTestId("automation-section-when")
        .locator(".el-checkbox")
        .first()
        .click();

      const weekdayRow = page
        .getByTestId("automation-section-when")
        .locator(".setting-row")
        .filter({ has: page.locator(".el-checkbox-group") });

      const timeRow = page
        .getByTestId("automation-section-when")
        .locator(".setting-row")
        .filter({ has: page.getByTestId("schedule-hour-input") });

      const gap = await measureVerticalGap(
        page,
        weekdayRow.locator(".el-checkbox-group"),
        timeRow,
      );
      expectGapPx(gap);
    });

    test(`automation schedule hour error to minute field is ${FIELD_GAP_PX}px at ${viewport.name}`, async ({
      page,
    }) => {
      await page.setViewportSize({
        width: viewport.width,
        height: viewport.height,
      });
      await openAutomationScheduleEditor(page);
      await page.getByTestId("item-name-input").fill("E2E spacing");
      await page
        .getByTestId("automation-section-when")
        .locator(".el-checkbox")
        .first()
        .click();
      const hourInput = page
        .getByTestId("schedule-hour-input")
        .locator("input");
      await hourInput.click();
      await hourInput.fill("");
      await hourInput.blur();
      await page.getByTestId("save-item").click();

      const hourItem = page
        .locator(".automation-inline-form-item")
        .filter({ has: page.getByTestId("schedule-hour-input") });
      await expect(hourItem).toHaveClass(/is-error/);

      const minuteItem = page
        .locator(".automation-inline-form-item")
        .filter({ has: page.getByTestId("schedule-minute-input") });

      const gap = await measureVerticalGapAfterFormError(
        page,
        hourItem,
        minuteItem,
      );
      expectFlushAfterErrorMargin(gap);
      await expect(hourItem.locator(".el-form-item__error")).toHaveCSS(
        "margin-bottom",
        `${FIELD_GAP_PX}px`,
      );
    });

    test(`automation schedule hour control to minute field is ${FIELD_GAP_PX}px without error at ${viewport.name}`, async ({
      page,
    }) => {
      await page.setViewportSize({
        width: viewport.width,
        height: viewport.height,
      });
      await openAutomationScheduleEditor(page);
      await page
        .getByTestId("automation-section-when")
        .locator(".el-checkbox")
        .first()
        .click();

      const hourItem = page
        .locator(".automation-inline-form-item")
        .filter({ has: page.getByTestId("schedule-hour-input") });
      const minuteItem = page
        .locator(".automation-inline-form-item")
        .filter({ has: page.getByTestId("schedule-minute-input") });

      const gap = await measureVerticalGap(page, hourItem, minuteItem);
      expectGapPx(gap);
    });
  }
});
