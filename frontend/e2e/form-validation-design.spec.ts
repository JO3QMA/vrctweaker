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

const RESOLUTION_INPUT_WIDTH_PX = 120;
const ERROR_COLOR_RGB = "rgb(239, 138, 138)";
const SUBPIXEL_PX = 1;

type ResolutionSurface = {
  name: "launcher" | "config";
  widthTestId: string;
  heightTestId: string;
  invalidWidth: number;
  invalidHeight: number;
};

function resolutionFieldsSelector(surface: ResolutionSurface): string {
  return surface.name === "config"
    ? '.vt-config-resolution-form:has([data-testid="camera-width-input"]) .resolution-fields'
    : ".launch-args-advanced .resolution-fields";
}

function resolutionFieldsLocator(
  page: import("@playwright/test").Page,
  surface: ResolutionSurface,
) {
  return page.locator(resolutionFieldsSelector(surface));
}

const RESOLUTION_SURFACES: ResolutionSurface[] = [
  {
    name: "launcher",
    widthTestId: "screen-width-input",
    heightTestId: "screen-height-input",
    invalidWidth: 1279,
    invalidHeight: 719,
  },
  {
    name: "config",
    widthTestId: "camera-width-input",
    heightTestId: "camera-height-input",
    invalidWidth: 1279,
    invalidHeight: 719,
  },
];

async function setUiLocale(
  page: import("@playwright/test").Page,
  locale: "ja" | "en",
) {
  await page.goto("/#/settings/general");
  await page.getByTestId("settings-ui-language").click();
  const label = locale === "ja" ? "日本語" : "English";
  await page
    .locator(".el-select-dropdown__item")
    .filter({ hasText: label })
    .last()
    .click();
  await page.keyboard.press("Escape");
}

async function openConfigCameraCustom(
  page: import("@playwright/test").Page,
  locale: "ja" | "en",
) {
  await page.goto("/#/config");
  if (locale === "ja") {
    await expect(page.locator("h1")).toContainText("その他の設定");
  } else {
    await expect(page.locator("h1")).toContainText("Other settings");
  }
  await expect(page.getByTestId("save-config-btn")).toBeVisible();
  await page.getByTestId("camera-preset-custom").click();
  await expect(page.getByTestId("camera-width-input")).toBeVisible();
}

async function openResolutionSurface(
  page: import("@playwright/test").Page,
  surface: ResolutionSurface,
  locale: "ja" | "en",
) {
  if (surface.name === "launcher") {
    await page.goto("/#/launcher");
    if (locale === "ja") {
      await expect(page.locator("h1")).toContainText("ランチャー");
    } else {
      await expect(page.locator("h1")).toContainText("Launcher");
    }
    await page.getByTestId("profile-card-profile-1").click();
    const viewport = page.viewportSize();
    if (viewport && viewport.width <= 640) {
      const layout = page.locator(".launcher-layout");
      const collapsed = await layout.evaluate((el) =>
        el.classList.contains("sidebar-collapsed"),
      );
      if (!collapsed) {
        await page.getByTestId("sidebar-toggle-btn").click();
        await expect(layout).toHaveClass(/sidebar-collapsed/);
      }
    }
    const collapseHeader = page.locator(
      ".args-collapse .el-collapse-item__header",
    );
    if (!(await page.getByTestId("resolution-enabled-checkbox").isVisible())) {
      await collapseHeader.click();
    }
    const resolutionToggle = page.getByTestId("resolution-enabled-checkbox");
    await expect(resolutionToggle).toBeVisible({ timeout: 10_000 });
    await resolutionToggle.click();
    await expect(resolutionToggle).toBeChecked({ timeout: 10_000 });
    await page.getByTestId("resolution-preset-custom").click();
    await expect(page.getByTestId(surface.widthTestId)).toBeVisible({
      timeout: 10_000,
    });
    if (viewport && viewport.width >= 1280) {
      await expect(
        page.locator(".launch-args-advanced .resolution-sep"),
      ).toBeVisible({ timeout: 10_000 });
    }
    return;
  }
  await openConfigCameraCustom(page, locale);
}

async function measureResolutionRow(
  page: import("@playwright/test").Page,
  surface: ResolutionSurface,
) {
  return page.evaluate(
    ({ widthTestId, heightTestId, fieldsSelector, surfaceName }) => {
      const fields = document.querySelector(fieldsSelector) as HTMLElement;
      const parent = ((surfaceName === "launcher"
        ? fields?.closest(".sub-options")
        : fields?.closest(".resolution-section")) ??
        fields?.parentElement ??
        fields) as HTMLElement;
      const sep = fields?.querySelector(
        ".resolution-sep",
      ) as HTMLElement | null;
      const widthEl = (fields.querySelector(
        `[data-testid="${widthTestId}"] .el-input-number`,
      ) ??
        fields.querySelector(`[data-testid="${widthTestId}"]`)) as HTMLElement;
      const heightEl = (fields.querySelector(
        `[data-testid="${heightTestId}"] .el-input-number`,
      ) ??
        fields.querySelector(`[data-testid="${heightTestId}"]`)) as HTMLElement;
      const errors = Array.from(
        fields.querySelectorAll(".el-form-item.is-error .el-form-item__error"),
      )
        .filter((e) => (e.textContent ?? "").trim().length > 0)
        .map((e) => e.getBoundingClientRect());
      const widthRect = widthEl.getBoundingClientRect();
      const heightRect = heightEl.getBoundingClientRect();
      const inputsRowBottom = Math.max(
        widthRect.bottom,
        heightRect.bottom,
        sep?.getBoundingClientRect().bottom ?? 0,
      );
      return {
        sepLeft: sep?.getBoundingClientRect().left ?? 0,
        heightLeft: heightRect.left,
        widthInputWidth: widthRect.width,
        heightInputWidth: heightRect.width,
        parentRect: parent.getBoundingClientRect(),
        fieldsRect: fields.getBoundingClientRect(),
        widthRect,
        heightRect,
        errors,
        inputsRowBottom,
        sepTop: sep?.getBoundingClientRect().top ?? 0,
        heightTop: heightRect.top,
      };
    },
    {
      widthTestId: surface.widthTestId,
      heightTestId: surface.heightTestId,
      fieldsSelector: resolutionFieldsSelector(surface),
      surfaceName: surface.name,
    },
  );
}

async function submitResolutionValidation(
  page: import("@playwright/test").Page,
  surface: ResolutionSurface,
  mode: "width" | "both" = "width",
): Promise<void> {
  await page.evaluate(
    async ({
      surfaceName,
      m,
      fieldsSelector,
      invalidWidth,
      invalidHeight: _invalidHeight,
    }) => {
      type VNode = {
        parent?: VNode;
        setupState?: Record<string, unknown>;
      };
      const fields = document.querySelector(fieldsSelector);
      if (!fields) {
        throw new Error("resolution-fields missing");
      }
      const reactive = (refLike: unknown) => {
        const maybe = refLike as {
          value?: Record<string, number | null>;
        } & Record<string, number | null>;
        return maybe.value ?? maybe;
      };
      let vnode = (fields as HTMLElement & { __vueParentComponent?: VNode })
        .__vueParentComponent;
      let ss: Record<string, unknown> | undefined;
      while (vnode) {
        const state = vnode.setupState;
        if (
          surfaceName === "launcher" &&
          state?.launcherFormModel &&
          state?.launcherFormRef
        ) {
          ss = state;
          break;
        }
        if (
          surfaceName === "config" &&
          state?.config &&
          state?.cameraResolutionFormRef
        ) {
          ss = state;
          break;
        }
        vnode = vnode.parent;
      }
      if (!ss) {
        throw new Error(`resolution form setup not found (${surfaceName})`);
      }
      if (surfaceName === "launcher") {
        const model = reactive(ss.launcherFormModel);
        const valueEnabled = reactive(ss.valueOptionsEnabled) as Record<
          string,
          unknown
        >;
        Object.assign(valueEnabled, { resolution: true });
        const preset = ss.resolutionPreset as { value?: string } | string;
        if (typeof preset === "object" && preset && "value" in preset) {
          preset.value = "custom";
        } else if (
          typeof ss.resolutionPreset === "object" &&
          ss.resolutionPreset
        ) {
          (ss.resolutionPreset as { value: string }).value = "custom";
        }
        if (m === "width") {
          model.screenWidth = invalidWidth;
        } else if (m === "both") {
          model.screenWidth = null;
          model.screenHeight = null;
        }
      } else {
        const cfg = reactive(ss.config);
        const cameraPreset = ss.cameraPreset as { value?: string };
        if (m === "width") {
          cfg.cameraResWidth = null;
        } else if (m === "both") {
          if (cameraPreset?.value !== undefined) {
            cameraPreset.value = "custom";
          }
          cfg.cameraResWidth = null;
          cfg.cameraResHeight = null;
        }
      }
      const formRefLike =
        surfaceName === "launcher"
          ? ss.launcherFormRef
          : ss.cameraResolutionFormRef;
      const unwrapRef = (refLike: unknown) => {
        const maybe = refLike as { value?: { validate?: () => Promise<void> } };
        return maybe?.value ?? maybe;
      };
      const form = unwrapRef(formRefLike) as
        | {
            validate?: () => Promise<boolean>;
            validateField?: (props: string) => Promise<boolean>;
          }
        | undefined;
      if (form?.validate) {
        const props =
          surfaceName === "launcher"
            ? m === "both"
              ? (["screenWidth", "screenHeight"] as const)
              : (["screenWidth"] as const)
            : m === "both"
              ? (["cameraResWidth", "cameraResHeight"] as const)
              : (["cameraResWidth"] as const);
        if (form.validateField) {
          for (const prop of props) {
            try {
              await form.validateField(prop);
            } catch {
              /* expected invalid */
            }
          }
        } else {
          try {
            await form.validate();
          } catch {
            /* expected invalid */
          }
        }
        return;
      }
      const save =
        surfaceName === "launcher"
          ? document.querySelector(".btn-save")
          : document.querySelector('[data-testid="save-config-btn"]');
      (save as HTMLElement | null)?.click();
    },
    {
      surfaceName: surface.name,
      m: mode,
      fieldsSelector: resolutionFieldsSelector(surface),
      invalidWidth: surface.invalidWidth,
      invalidHeight: surface.invalidHeight,
    },
  );
}

async function ensureResolutionSurfaceReady(
  page: import("@playwright/test").Page,
  surface: ResolutionSurface,
  locale: "ja" | "en",
) {
  const fields = resolutionFieldsLocator(page, surface);
  if ((await fields.count()) === 0) {
    await openResolutionSurface(page, surface, locale);
  }
  await expect(fields).toBeVisible();
  await expect(page.getByTestId(surface.widthTestId)).toBeVisible();
}

async function triggerResolutionErrorsAndMeasure(
  page: import("@playwright/test").Page,
  surface: ResolutionSurface,
  mode: "width" | "both",
  locale: "ja" | "en",
) {
  await ensureResolutionSurfaceReady(page, surface, locale);
  const before = await measureResolutionRow(page, surface);
  await submitResolutionValidation(page, surface, mode);
  const errors = resolutionFieldsLocator(page, surface).locator(
    ".el-form-item.is-error .el-form-item__error",
  );
  await expect(errors.first()).toBeAttached();
  await expect(errors.first()).not.toBeEmpty();
  if (mode === "both") {
    await expect(errors).toHaveCount(2, { timeout: 10_000 });
  }
  const after = await measureResolutionRow(page, surface);
  return { before, after };
}

async function triggerLauncherScreenWidthRangeError(
  page: import("@playwright/test").Page,
) {
  await openLauncherResolution(page);
  await submitResolutionValidation(page, RESOLUTION_SURFACES[0]!, "width");
  const widthItem = page
    .locator(".resolution-form-item")
    .filter({ has: page.getByTestId("screen-width-input") });
  await expect(widthItem).toHaveClass(/is-error/);
  await expect(widthItem.locator(".el-form-item__error")).not.toBeEmpty();
  return widthItem;
}

function expectNoHorizontalScroll(
  docScroll: boolean,
  bodyScroll: boolean,
): void {
  expect(docScroll).toBe(true);
  expect(bodyScroll).toBe(true);
}

function rectInsideParent(
  child: DOMRect,
  parent: DOMRect,
  tolerance = SUBPIXEL_PX,
): boolean {
  return (
    child.left >= parent.left - tolerance &&
    child.right <= parent.right + tolerance &&
    child.top >= parent.top - tolerance &&
    child.bottom <= parent.bottom + tolerance
  );
}

const MIN_ERROR_READABLE_WIDTH_PX = 48;
const MAX_ERROR_STACKED_LINES = 4;
const NARROW_FIELDS_MAX_WIDTH_PX = 280;
const LAUNCHER_NARROW_FIELDS = ".launch-args-advanced .resolution-fields";

async function openLauncherProfileSidebarOpen600(
  page: import("@playwright/test").Page,
  locale: "ja" | "en",
) {
  await setUiLocale(page, locale);
  await page.setViewportSize({ width: 600, height: 900 });
  await page.goto("/#/launcher");
  if (locale === "ja") {
    await expect(page.locator("h1")).toContainText("ランチャー");
  } else {
    await expect(page.locator("h1")).toContainText("Launcher");
  }
  await page.getByTestId("profile-card-profile-1").click();
  await expect(page.locator(".launcher-layout")).not.toHaveClass(
    /sidebar-collapsed/,
  );
  const collapseHeader = page.locator(
    ".args-collapse .el-collapse-item__header",
  );
  await collapseHeader.click();
  const resolutionToggle = page.getByTestId("resolution-enabled-checkbox");
  await expect(resolutionToggle).toBeVisible({ timeout: 10_000 });
  await resolutionToggle.click();
  await expect(resolutionToggle).toBeChecked({ timeout: 10_000 });
  await page.getByTestId("resolution-preset-custom").click();
  await expect(page.getByTestId("screen-width-input")).toBeVisible({
    timeout: 10_000,
  });
}

async function measureLauncherNarrowResolutionLayout(
  page: import("@playwright/test").Page,
) {
  return page.evaluate(
    ({ fieldsSelector, overlapTol }) => {
      const fields = document.querySelector(fieldsSelector) as HTMLElement;
      const parent = fields;
      if (!fields) {
        return {
          fieldsWidth: 0,
          overlaps: ["missing-fields"],
          errorMinWidth: 0,
          allInside: false,
        };
      }
      const pickInputNumber = (testId: string) =>
        fields.querySelector(
          `[data-testid="${testId}"].el-input-number, [data-testid="${testId}"] .el-input-number`,
        ) as HTMLElement | null;
      const width = pickInputNumber("screen-width-input");
      const height = pickInputNumber("screen-height-input");
      const pickIncrease = (inputNumber: HTMLElement | null) =>
        inputNumber?.querySelector(
          ".el-input-number__increase",
        ) as HTMLElement | null;
      const widthIncrease = pickIncrease(width);
      const heightIncrease = pickIncrease(height);
      const hints = Array.from(
        fields.querySelectorAll(".form-field-range-hint"),
      ) as HTMLElement[];
      const errors = Array.from(
        fields.querySelectorAll(".el-form-item.is-error .el-form-item__error"),
      ).filter((e) => (e.textContent ?? "").trim().length > 0) as HTMLElement[];
      const labeled: { name: string; rect: DOMRect }[] = [];
      const containmentTargets: { name: string; el: HTMLElement }[] = [];
      if (width) {
        labeled.push({ name: "width", rect: width.getBoundingClientRect() });
        containmentTargets.push({ name: "width-input", el: width });
      }
      if (height) {
        labeled.push({ name: "height", rect: height.getBoundingClientRect() });
        containmentTargets.push({ name: "height-input", el: height });
      }
      if (widthIncrease) {
        containmentTargets.push({
          name: "width-increase",
          el: widthIncrease,
        });
      }
      if (heightIncrease) {
        containmentTargets.push({
          name: "height-increase",
          el: heightIncrease,
        });
      }
      for (const [i, h] of hints.entries()) {
        labeled.push({ name: `hint-${i}`, rect: h.getBoundingClientRect() });
      }
      for (const [i, e] of errors.entries()) {
        labeled.push({ name: `error-${i}`, rect: e.getBoundingClientRect() });
      }
      const overlaps: string[] = [];
      for (let i = 0; i < labeled.length; i++) {
        for (let j = i + 1; j < labeled.length; j++) {
          const a = labeled[i]!;
          const b = labeled[j]!;
          const hOverlap =
            Math.min(a.rect.right, b.rect.right) -
            Math.max(a.rect.left, b.rect.left);
          const vOverlap =
            Math.min(a.rect.bottom, b.rect.bottom) -
            Math.max(a.rect.top, b.rect.top);
          if (hOverlap > overlapTol && vOverlap > overlapTol) {
            overlaps.push(`${a.name}×${b.name}`);
          }
        }
      }
      const parentRect = parent.getBoundingClientRect();
      const outside: string[] = [];
      const containmentOverflow: { name: string; rightOverflow: number }[] = [];
      for (const { name, el } of containmentTargets) {
        const rect = el.getBoundingClientRect();
        const rightOverflow = rect.right - parentRect.right;
        if (rightOverflow > overlapTol) {
          containmentOverflow.push({ name, rightOverflow });
          outside.push(name);
        }
        const inside =
          rect.left >= parentRect.left - overlapTol &&
          rect.right <= parentRect.right + overlapTol &&
          rect.top >= parentRect.top - overlapTol &&
          rect.bottom <= parentRect.bottom + overlapTol;
        if (!inside && !outside.includes(name)) {
          outside.push(name);
        }
      }
      const allInside =
        containmentTargets.length > 0 &&
        containmentOverflow.length === 0 &&
        containmentTargets.every(({ el }) => {
          const rect = el.getBoundingClientRect();
          return (
            rect.left >= parentRect.left - overlapTol &&
            rect.right <= parentRect.right + overlapTol &&
            rect.top >= parentRect.top - overlapTol &&
            rect.bottom <= parentRect.bottom + overlapTol
          );
        });
      const errorMinWidth =
        errors.length === 0
          ? 48
          : Math.min(...errors.map((e) => e.getBoundingClientRect().width));
      let errorMaxLines = 0;
      for (const err of errors) {
        const style = getComputedStyle(err);
        const lh = parseFloat(style.lineHeight) || 16;
        const h = err.getBoundingClientRect().height;
        errorMaxLines = Math.max(errorMaxLines, Math.ceil(h / lh));
      }
      return {
        fieldsWidth: fields.getBoundingClientRect().width,
        overlaps,
        errorMinWidth,
        allInside,
        errorMaxLines,
        outside,
        containmentOverflow,
        fieldsScrollOverflow: fields.scrollWidth - fields.clientWidth,
        fieldsClientWidth: fields.clientWidth,
      };
    },
    {
      fieldsSelector: LAUNCHER_NARROW_FIELDS,
      overlapTol: GAP_TOLERANCE_PX,
    },
  );
}

test.describe("Form validation layout (design QA)", () => {
  test("launcher resolution hint sits below inputs in resolution row at 1280px", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1280, height: 800 });
    await openLauncherResolution(page);

    const hintBelow = await page.evaluate(() => {
      const fields = document.querySelector(".resolution-fields")!;
      const input = fields.querySelector('[data-testid="screen-width-input"]')!;
      const hint = fields.querySelector(".form-field-range-hint")!;
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
      await triggerLauncherScreenWidthRangeError(page);

      const metrics = await page.evaluate(() => {
        const widthItem = document
          .querySelector('[data-testid="screen-width-input"]')!
          .closest(".el-form-item")!;
        const err = widthItem.querySelector(".el-form-item__error")!;
        const height = document.querySelector(
          '[data-testid="screen-height-input"]',
        )!;
        const errRect = err.getBoundingClientRect();
        const heightRect = height.getBoundingClientRect();
        const sepRect = document
          .querySelector(".resolution-fields .resolution-sep")
          ?.getBoundingClientRect();
        const fields = document.querySelector(
          ".launch-args-advanced .resolution-fields",
        );
        const fieldsRect = fields?.getBoundingClientRect();
        const horizontalOverlap =
          Math.min(errRect.right, heightRect.right) -
          Math.max(errRect.left, heightRect.left);
        return {
          horizontalOverlap,
          sepTop: sepRect?.top ?? 0,
          heightTop: heightRect.top,
          fieldsWidth: fieldsRect?.width ?? 0,
          inputWidth: (
            widthItem.querySelector('[data-testid="screen-width-input"]') ??
            widthItem.querySelector(".el-input-number")
          )?.getBoundingClientRect().width,
        };
      });
      if (metrics.fieldsWidth >= 279 || viewport.width >= 1280) {
        expect(metrics.horizontalOverlap).toBeLessThanOrEqual(GAP_TOLERANCE_PX);
      }
      if (viewport.width >= 1280) {
        expect(
          Math.abs(metrics.sepTop - metrics.heightTop),
        ).toBeLessThanOrEqual(4);
      }
      if (metrics.inputWidth != null && metrics.inputWidth > 0) {
        expect(metrics.inputWidth).toBeLessThanOrEqual(200);
        if (
          viewport.width >= 1280 ||
          metrics.fieldsWidth >= 279 - SUBPIXEL_PX
        ) {
          expect(metrics.inputWidth).toBeGreaterThanOrEqual(110);
        } else if (metrics.fieldsWidth > 0) {
          expect(metrics.inputWidth).toBeLessThanOrEqual(
            metrics.fieldsWidth + GAP_TOLERANCE_PX,
          );
        }
      }
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

      const hourControlBox = await page
        .getByTestId("schedule-hour-input")
        .boundingBox();
      expect(hourControlBox?.height ?? 999).toBeLessThanOrEqual(40);
      expect(hourControlBox?.height ?? 0).toBeGreaterThanOrEqual(28);
      expect(hourControlBox?.width ?? 999).toBeLessThanOrEqual(140);

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

test.describe("Resolution layout acceptance (design QA)", () => {
  test.setTimeout(120_000);
  for (const locale of ["ja", "en"] as const) {
    for (const viewport of [
      { width: 1280, height: 600 },
      { width: 600, height: 900 },
    ] as const) {
      for (const surface of RESOLUTION_SURFACES) {
        test(`${surface.name} ${locale} ${viewport.width}px: row stable and errors below row`, async ({
          page,
        }) => {
          await setUiLocale(page, locale);
          await page.setViewportSize(viewport);
          await openResolutionSurface(page, surface, locale);

          const widthError = await triggerResolutionErrorsAndMeasure(
            page,
            surface,
            "width",
            locale,
          );
          const { before: baseline, after: afterWidthError } = widthError;

          if (viewport.width >= 1280) {
            expect(
              Math.abs(afterWidthError.heightLeft - baseline.heightLeft),
            ).toBeLessThanOrEqual(SUBPIXEL_PX);
            expect(
              Math.abs(afterWidthError.sepLeft - baseline.sepLeft),
            ).toBeLessThanOrEqual(SUBPIXEL_PX);
            expect(
              Math.abs(afterWidthError.sepTop - afterWidthError.heightTop),
            ).toBeLessThanOrEqual(4);
          }

          if (viewport.width >= 1280) {
            expect(afterWidthError.widthInputWidth).toBeGreaterThanOrEqual(
              RESOLUTION_INPUT_WIDTH_PX - SUBPIXEL_PX,
            );
            expect(afterWidthError.widthInputWidth).toBeLessThanOrEqual(
              RESOLUTION_INPUT_WIDTH_PX + SUBPIXEL_PX,
            );
            expect(afterWidthError.heightInputWidth).toBeGreaterThanOrEqual(
              RESOLUTION_INPUT_WIDTH_PX - SUBPIXEL_PX,
            );
            expect(afterWidthError.heightInputWidth).toBeLessThanOrEqual(
              RESOLUTION_INPUT_WIDTH_PX + SUBPIXEL_PX,
            );
          } else if (afterWidthError.fieldsRect.width >= 279 - SUBPIXEL_PX) {
            expect(afterWidthError.widthInputWidth).toBeLessThanOrEqual(
              afterWidthError.fieldsRect.width + SUBPIXEL_PX,
            );
            expect(afterWidthError.heightInputWidth).toBeLessThanOrEqual(
              afterWidthError.fieldsRect.width + SUBPIXEL_PX,
            );
          }

          expect(afterWidthError.errors.length).toBeGreaterThanOrEqual(1);
          const gapToError =
            afterWidthError.errors[0]!.top - afterWidthError.widthRect.bottom;
          expect(gapToError).toBeGreaterThanOrEqual(
            INPUT_TO_ERROR_PX - GAP_TOLERANCE_PX,
          );
          expect(gapToError).toBeLessThanOrEqual(
            INPUT_TO_ERROR_PX + GAP_TOLERANCE_PX,
          );

          const errorColor = await page.evaluate((fieldsSelector) => {
            const el = document.querySelector(
              `${fieldsSelector} .el-form-item__error`,
            );
            return el ? getComputedStyle(el).color : "";
          }, resolutionFieldsSelector(surface));
          expect(errorColor).toBe(ERROR_COLOR_RGB);

          const bothError = await triggerResolutionErrorsAndMeasure(
            page,
            surface,
            "both",
            locale,
          );
          const afterBoth = bothError.after;
          expect(afterBoth.errors.length).toBeGreaterThanOrEqual(2);
          const sortedErrors = [...afterBoth.errors].sort(
            (a, b) => a.top - b.top || a.left - b.left,
          );
          const errA = sortedErrors[0]!;
          const errB = sortedErrors[1]!;
          if (Math.abs(errA.left - errB.left) > SUBPIXEL_PX) {
            expect(Math.abs(errA.top - errB.top)).toBeLessThanOrEqual(4);
          } else {
            expect(errB.top).toBeGreaterThanOrEqual(errA.bottom - SUBPIXEL_PX);
          }

          const scroll = await page.evaluate(() => ({
            doc:
              document.documentElement.scrollWidth <=
              document.documentElement.clientWidth + 1,
            body: document.body.scrollWidth <= document.body.clientWidth + 1,
          }));
          expectNoHorizontalScroll(scroll.doc, scroll.body);

          expect(
            rectInsideParent(afterBoth.widthRect, afterBoth.parentRect),
          ).toBe(true);
          expect(
            rectInsideParent(afterBoth.heightRect, afterBoth.parentRect),
          ).toBe(true);
          expect(
            rectInsideParent(afterBoth.fieldsRect, afterBoth.parentRect),
          ).toBe(true);
        });
      }
    }
  }
});

test.describe("on-demand Element Plus form chrome", () => {
  const VALIDATED_FIELD_SPACING_PX = 22;
  const CONFIG_SURFACE = RESOLUTION_SURFACES.find((s) => s.name === "config")!;

  test("vt-validated-form applies 22px direct child item spacing on launcher", async ({
    page,
  }) => {
    await page.goto("/#/launcher");
    await expect(page.locator("h1")).toContainText("ランチャー");
    await page.getByTestId("profile-card-profile-1").click();
    const form = page.locator(".profile-editor .vt-validated-form.el-form");
    await expect(form).toBeVisible();
    const firstItem = form.locator(":scope > .el-form-item").first();
    await expect(firstItem).toBeVisible();
    await expect(firstItem).toHaveCSS(
      "margin-bottom",
      `${VALIDATED_FIELD_SPACING_PX}px`,
    );
  });

  test("config resolution error uses themed form-item error chrome", async ({
    page,
  }) => {
    await triggerResolutionErrorsAndMeasure(
      page,
      CONFIG_SURFACE,
      "width",
      "ja",
    );
    const err = resolutionFieldsLocator(page, CONFIG_SURFACE)
      .locator(".el-form-item.is-error .el-form-item__error")
      .first();
    await expect(err).toBeVisible();
    await expect(err).toHaveAttribute("role", "alert");
    await expect(err).toHaveCSS("line-height", "16px");
    const [errColor, tokenColor] = await page.evaluate(() => {
      const errEl = document.querySelector(
        ".vt-config-resolution-form .resolution-fields .el-form-item__error",
      ) as HTMLElement | null;
      const errRgb = errEl
        ? getComputedStyle(errEl).color
        : "";
      const token = getComputedStyle(document.documentElement)
        .getPropertyValue("--color-error")
        .trim();
      const probe = document.createElement("div");
      probe.style.color = token.startsWith("var(") ? token : `var(--color-error)`;
      document.body.appendChild(probe);
      const tokenRgb = getComputedStyle(probe).color;
      probe.remove();
      return [errRgb, tokenRgb] as const;
    });
    expect(errColor).toBe(tokenColor);
    expect(errColor).not.toBe("rgb(0, 0, 0)");
  });
});

test.describe("Launcher 600 narrow column (design QA)", () => {
  test.setTimeout(120_000);
  const LAUNCHER_SURFACE = RESOLUTION_SURFACES[0]!;
  for (const locale of ["ja", "en"] as const) {
    for (const errorMode of ["none", "width", "both"] as const) {
      test(`launcher ${locale} 600px sidebar open (${errorMode} error): no overlap`, async ({
        page,
      }) => {
        await openLauncherProfileSidebarOpen600(page, locale);
        if (errorMode !== "none") {
          await submitResolutionValidation(page, LAUNCHER_SURFACE, errorMode);
          const fieldsLoc = page.locator(LAUNCHER_NARROW_FIELDS);
          await expect(fieldsLoc).toBeVisible({ timeout: 10_000 });
          const errors = fieldsLoc.locator(
            ".el-form-item.is-error .el-form-item__error",
          );
          await expect(errors.first()).not.toBeEmpty({ timeout: 10_000 });
          if (errorMode === "both") {
            await expect(errors).toHaveCount(2, { timeout: 10_000 });
          }
        }
        const layout = await measureLauncherNarrowResolutionLayout(page);
        expect(layout.fieldsWidth).toBeGreaterThan(0);
        expect(layout.fieldsWidth).toBeLessThanOrEqual(
          NARROW_FIELDS_MAX_WIDTH_PX,
        );
        expect(
          layout.allInside,
          JSON.stringify(layout.containmentOverflow),
        ).toBe(true);
        expect(layout.outside).toEqual([]);
        expect(layout.overlaps).toEqual([]);
        expect(layout.fieldsScrollOverflow).toBeLessThanOrEqual(1);
        expect(layout.fieldsClientWidth).toBeGreaterThan(0);
        if (errorMode !== "none") {
          expect(layout.errorMinWidth).toBeGreaterThanOrEqual(
            MIN_ERROR_READABLE_WIDTH_PX,
          );
          expect(layout.errorMaxLines).toBeLessThanOrEqual(
            MAX_ERROR_STACKED_LINES,
          );
        }
        const scroll = await page.evaluate(() => ({
          doc:
            document.documentElement.scrollWidth <=
            document.documentElement.clientWidth + 1,
          body: document.body.scrollWidth <= document.body.clientWidth + 1,
        }));
        expectNoHorizontalScroll(scroll.doc, scroll.body);
      });
    }
  }
});
