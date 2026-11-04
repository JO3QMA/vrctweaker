import { test, expect } from "@playwright/test";
import { getMockWailsInitScript } from "./fixtures/mock-wails";

const FIELD_GAP_PX = 8;
const INPUT_TO_ERROR_PX = 4;
const GAP_TOLERANCE_PX = 1;

test.beforeEach(async ({ page }) => {
  await page.addInitScript(getMockWailsInitScript());
});

function expectGapPx(actual: number, expected = FIELD_GAP_PX) {
  expect(actual).toBeGreaterThanOrEqual(expected - GAP_TOLERANCE_PX);
  expect(actual).toBeLessThanOrEqual(expected + GAP_TOLERANCE_PX);
}

function expectInputToErrorGap(actual: number) {
  expect(actual).toBeGreaterThanOrEqual(-GAP_TOLERANCE_PX);
  expect(actual).toBeLessThanOrEqual(INPUT_TO_ERROR_PX + GAP_TOLERANCE_PX);
}

async function expectInlineErrorInputSpacing(
  page: import("@playwright/test").Page,
  formItem: import("@playwright/test").Locator,
) {
  expectInputToErrorGap(
    await measureInputControlBottomToErrorTop(page, formItem),
  );
  await expect(formItem.locator(".el-form-item__error")).toHaveCSS(
    "padding-top",
    `${INPUT_TO_ERROR_PX}px`,
  );
}

async function measureAutomationHourToMinuteSpacing(
  page: import("@playwright/test").Page,
): Promise<{ axis: "x" | "y"; gap: number }> {
  return page.evaluate(() => {
    const hour = document
      .querySelector('[data-testid="schedule-hour-input"]')!
      .closest(".el-form-item")!;
    const minute = document
      .querySelector('[data-testid="schedule-minute-input"]')!
      .closest(".el-form-item")!;
    const h = hour.getBoundingClientRect();
    const m = minute.getBoundingClientRect();
    if (m.left > h.right + 1) {
      return { axis: "x" as const, gap: m.left - h.right };
    }
    return { axis: "y" as const, gap: m.top - h.bottom };
  });
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

async function measureInputControlBottomToErrorTop(
  page: import("@playwright/test").Page,
  formItem: import("@playwright/test").Locator,
): Promise<number> {
  const handle = await formItem.elementHandle();
  expect(handle).toBeTruthy();
  return page.evaluate((item) => {
    const root = item as HTMLElement;
    const control =
      root.querySelector(".el-input-number") ??
      root.querySelector(".el-checkbox-group");
    const err = root.querySelector(".el-form-item__error");
    if (!control || !err) {
      return -1;
    }
    return (
      err.getBoundingClientRect().top - control.getBoundingClientRect().bottom
    );
  }, handle!);
}

async function expectErrorBelowControl(
  page: import("@playwright/test").Page,
  formItem: import("@playwright/test").Locator,
) {
  const handle = await formItem.elementHandle();
  expect(handle).toBeTruthy();
  const layout = await page.evaluate((item) => {
    const root = item as HTMLElement;
    const control = root.querySelector(".el-input-number")!;
    const err = root.querySelector(".el-form-item__error")!;
    const c = control.getBoundingClientRect();
    const e = err.getBoundingClientRect();
    return {
      below: e.top >= c.bottom - 1,
      alignedLeft: Math.abs(e.left - c.left) <= 4,
    };
  }, handle!);
  expect(layout.below).toBe(true);
  expect(layout.alignedLeft).toBe(true);
}

async function triggerLauncherScreenWidthRangeError(
  page: import("@playwright/test").Page,
) {
  await openLauncherResolution(page);
  await page.evaluate(async () => {
    type VueNode = {
      setupState?: Record<string, unknown>;
      parent?: VueNode;
    };
    const findLauncherStateInDom = (): Record<string, unknown> | null => {
      for (const el of document.querySelectorAll("*")) {
        let vnode = (el as HTMLElement & { __vueParentComponent?: VueNode })
          .__vueParentComponent;
        let depth = 0;
        while (vnode && depth < 32) {
          const state = vnode.setupState;
          if (state?.launcherFormRef && state?.launcherFormModel) {
            return state;
          }
          vnode = vnode.parent;
          depth += 1;
        }
      }
      return null;
    };
    const launcherState = findLauncherStateInDom();
    if (!launcherState) {
      throw new Error("launcher form state not found");
    }
    const model = launcherState.launcherFormModel as { screenWidth: number };
    model.screenWidth = 1279;
    const rawRef = launcherState.launcherFormRef as
      | {
          value?: {
            validateField: (p: string) => Promise<void>;
            validate: () => Promise<boolean>;
          };
        }
      | {
          validateField?: (p: string) => Promise<void>;
          validate?: () => Promise<boolean>;
        };
    const form =
      rawRef &&
      "value" in rawRef &&
      rawRef.value &&
      typeof rawRef.value.validateField === "function"
        ? rawRef.value
        : rawRef;
    if (!form || typeof form.validateField !== "function") {
      throw new Error("launcher form ref missing validateField");
    }
    await form.validateField("screenWidth").catch(() => {});
    if (typeof form.validate === "function") {
      await form.validate().catch(() => {});
    }
  });
  const widthItem = page
    .locator(".resolution-form-item")
    .filter({ has: page.getByTestId("screen-width-input") });
  await expect(widthItem).toHaveClass(/is-error/);
  await expect(widthItem.locator(".el-form-item__error")).not.toBeEmpty();
  return widthItem;
}

test.describe("Form validation layout (design QA)", () => {
  test("launcher resolution hint uses flex column below inputs at 1280px", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1280, height: 800 });
    await openLauncherResolution(page);

    const content = page
      .locator(".vt-form-item-with-range-hint .el-form-item__content")
      .first();
    await expect(content).toBeVisible();
    expect(await content.evaluate((el) => getComputedStyle(el).display)).toBe(
      "flex",
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

    test(`launcher resolution width error spacing at ${viewport.name}`, async ({
      page,
    }) => {
      await page.setViewportSize({
        width: viewport.width,
        height: viewport.height,
      });
      const widthItem = await triggerLauncherScreenWidthRangeError(page);

      expectInputToErrorGap(
        await measureInputControlBottomToErrorTop(page, widthItem),
      );
      await expectErrorBelowControl(page, widthItem);

      const metrics = await page.evaluate(() => {
        const widthItem = document
          .querySelector('[data-testid="screen-width-input"]')!
          .closest(".el-form-item")!;
        const err = widthItem.querySelector(".el-form-item__error")!;
        const height = document.querySelector(
          '[data-testid="screen-height-input"]',
        )!;
        const errBottom = err.getBoundingClientRect().bottom;
        const heightTop = height.getBoundingClientRect().top;
        const itemBottom = widthItem.getBoundingClientRect().bottom;
        return {
          errorToHeight: heightTop - errBottom,
          itemToHeight: heightTop - itemBottom,
          overlapPx: errBottom - heightTop,
        };
      });
      expect(metrics.overlapPx).toBeLessThanOrEqual(GAP_TOLERANCE_PX);
      expectGapPx(metrics.itemToHeight);
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

      await expectInlineErrorInputSpacing(page, weekdayItem);

      const timeRow = page
        .getByTestId("automation-section-when")
        .locator(".setting-row")
        .filter({ has: page.getByTestId("schedule-hour-input") });

      const gap = await measureVerticalGap(
        page,
        weekdayItem.locator(".el-form-item__error"),
        timeRow,
      );
      expect(gap).toBeGreaterThanOrEqual(FIELD_GAP_PX - GAP_TOLERANCE_PX);
      expect(gap).toBeLessThanOrEqual(FIELD_GAP_PX + 3);
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

    test(`automation schedule hour error layout at ${viewport.name}`, async ({
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

      await expectInlineErrorInputSpacing(page, hourItem);
      await expectErrorBelowControl(page, hourItem);

      const timeFlexGap = await page
        .locator(".automation-time")
        .evaluate((el) => {
          return parseFloat(
            getComputedStyle(el).columnGap || getComputedStyle(el).gap,
          );
        });
      expectGapPx(timeFlexGap);

      const spacing = await measureAutomationHourToMinuteSpacing(page);
      expect(spacing.gap).toBeGreaterThanOrEqual(
        FIELD_GAP_PX - GAP_TOLERANCE_PX,
      );
      if (spacing.axis === "x") {
        expect(spacing.gap).toBeGreaterThan(FIELD_GAP_PX - GAP_TOLERANCE_PX);
      } else {
        expectGapPx(spacing.gap);
      }
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

      const spacing = await measureAutomationHourToMinuteSpacing(page);
      expect(spacing.gap).toBeGreaterThanOrEqual(
        FIELD_GAP_PX - GAP_TOLERANCE_PX,
      );
      if (spacing.axis === "y") {
        expectGapPx(spacing.gap);
      }
    });

    test(`automation schedule time row keeps row padding and divider at ${viewport.name}`, async ({
      page,
    }) => {
      await page.setViewportSize({
        width: viewport.width,
        height: viewport.height,
      });
      await openAutomationScheduleEditor(page);

      const whenSection = page.getByTestId("automation-section-when");
      const triggerLabel = whenSection
        .locator(".setting-row")
        .first()
        .locator(".setting-row__label");
      const timeRow = whenSection
        .locator(".setting-row")
        .filter({ has: page.getByTestId("schedule-hour-input") });
      const timeLabel = timeRow.locator(".setting-row__label");

      const timeRowStyle = await timeRow.evaluate((el) => {
        const style = getComputedStyle(el);
        return {
          paddingTop: parseFloat(style.paddingTop),
          borderTopWidth: parseFloat(style.borderTopWidth),
        };
      });
      expect(timeRowStyle.paddingTop).toBe(16);
      expect(timeRowStyle.borderTopWidth).toBe(1);

      const fonts = await page.evaluate(
        ([trigger, time]) => {
          const t = getComputedStyle(trigger as Element);
          const s = getComputedStyle(time as Element);
          return {
            trigger: {
              size: t.fontSize,
              weight: t.fontWeight,
              lineHeight: t.lineHeight,
            },
            time: {
              size: s.fontSize,
              weight: s.fontWeight,
              lineHeight: s.lineHeight,
            },
          };
        },
        [await triggerLabel.elementHandle(), await timeLabel.elementHandle()],
      );
      expect(fonts.time.size).toBe(fonts.trigger.size);
      expect(fonts.time.weight).toBe(fonts.trigger.weight);
      expect(fonts.time.lineHeight).toBe(fonts.trigger.lineHeight);
    });
  }
});
