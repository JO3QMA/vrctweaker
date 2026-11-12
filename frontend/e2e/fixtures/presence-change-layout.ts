import { expect, type Page } from "@playwright/test";

export const PRESENCE_COLOR_TEST_IDS = [
  "presence-change-color-join-me",
  "presence-change-color-active",
  "presence-change-color-ask-me",
  "presence-change-color-busy",
] as const;

export type PresenceLayoutMetrics = {
  buttons: Array<{
    testId: string;
    scrollWidth: number;
    clientWidth: number;
    scrollHeight: number;
    clientHeight: number;
    top: number;
    left: number;
    right: number;
    bottom: number;
  }>;
  card: { left: number; right: number; top: number; bottom: number };
  gridTemplateColumns: string;
  horizontalOverflow: boolean;
  containerWidth: number;
  dashboardWidth: number;
};

export async function readPresenceLayoutMetrics(
  page: Page,
): Promise<PresenceLayoutMetrics> {
  return page.evaluate(
    (testIds) => {
      const card = document.querySelector(
        '[data-testid="presence-change-section"]',
      );
      if (!(card instanceof HTMLElement)) {
        throw new Error("presence-change-section missing");
      }
      const cardRect = card.getBoundingClientRect();
      const grid = card.querySelector(".presence-color-buttons");
      const gridStyle =
        grid instanceof HTMLElement ? getComputedStyle(grid) : null;
      const container = card.querySelector(".presence-color-buttons-container");
      const dashboard = document.querySelector(".dashboard");

      const buttons = testIds.map((testId) => {
        const el = card.querySelector(`[data-testid="${testId}"]`);
        if (!(el instanceof HTMLElement)) {
          throw new Error(`missing presence color button: ${testId}`);
        }
        const rect = el.getBoundingClientRect();
        return {
          testId,
          scrollWidth: el.scrollWidth,
          clientWidth: el.clientWidth,
          scrollHeight: el.scrollHeight,
          clientHeight: el.clientHeight,
          top: rect.top,
          left: rect.left,
          right: rect.right,
          bottom: rect.bottom,
        };
      });

      const main = document.querySelector(
        ".main-content",
      ) as HTMLElement | null;
      const horizontalOverflow =
        (main?.scrollWidth ?? 0) > (main?.clientWidth ?? 0) + 1 ||
        document.documentElement.scrollWidth >
          document.documentElement.clientWidth + 1;

      return {
        buttons,
        card: {
          left: cardRect.left,
          right: cardRect.right,
          top: cardRect.top,
          bottom: cardRect.bottom,
        },
        gridTemplateColumns: gridStyle?.gridTemplateColumns ?? "",
        horizontalOverflow,
        containerWidth:
          container instanceof HTMLElement ? container.clientWidth : 0,
        dashboardWidth:
          dashboard instanceof HTMLElement ? dashboard.clientWidth : 0,
      };
    },
    [...PRESENCE_COLOR_TEST_IDS],
  );
}

export function expectPresenceButtonsSingleLine(
  metrics: PresenceLayoutMetrics,
): void {
  const heights = metrics.buttons.map((b) => b.clientHeight);
  const maxHeight = Math.max(...heights);
  const minHeight = Math.min(...heights);
  expect(maxHeight - minHeight, "button heights aligned").toBeLessThanOrEqual(
    2,
  );

  for (const button of metrics.buttons) {
    expect(
      button.scrollWidth,
      `${button.testId} text fits without horizontal clip`,
    ).toBeLessThanOrEqual(button.clientWidth + 1);
    expect(
      button.scrollHeight,
      `${button.testId} label on one line`,
    ).toBeLessThanOrEqual(button.clientHeight + 1);
  }
}

export function expectPresenceButtonsInsideCard(
  metrics: PresenceLayoutMetrics,
): void {
  for (const button of metrics.buttons) {
    expect(
      button.left,
      `${button.testId} left inside card`,
    ).toBeGreaterThanOrEqual(metrics.card.left - 1);
    expect(
      button.right,
      `${button.testId} right inside card`,
    ).toBeLessThanOrEqual(metrics.card.right + 1);
    expect(
      button.top,
      `${button.testId} top inside card`,
    ).toBeGreaterThanOrEqual(metrics.card.top - 1);
    expect(
      button.bottom,
      `${button.testId} bottom inside card`,
    ).toBeLessThanOrEqual(metrics.card.bottom + 1);
  }
}

function gridColumnCount(gridTemplateColumns: string): number {
  return gridTemplateColumns
    .split(" ")
    .map((part) => part.trim())
    .filter(Boolean).length;
}

export function expectPresenceGridTwoByTwo(
  metrics: PresenceLayoutMetrics,
): void {
  const [a, b, c, d] = metrics.buttons;
  expect(Math.abs(a.top - b.top), "row 1 alignment").toBeLessThanOrEqual(2);
  expect(Math.abs(c.top - d.top), "row 2 alignment").toBeLessThanOrEqual(2);
  expect(c.top - a.top, "two grid rows").toBeGreaterThan(12);
  expect(
    gridColumnCount(metrics.gridTemplateColumns),
    "grid uses two columns",
  ).toBe(2);
}

export function expectPresenceGridSingleRow(
  metrics: PresenceLayoutMetrics,
): void {
  const tops = metrics.buttons.map((b) => b.top);
  const maxDelta = Math.max(...tops) - Math.min(...tops);
  expect(maxDelta, "four buttons on one row").toBeLessThanOrEqual(2);
  expect(
    gridColumnCount(metrics.gridTemplateColumns),
    "grid uses four columns",
  ).toBe(4);
}

export async function setUiLocale(
  page: Page,
  locale: "ja" | "en",
): Promise<void> {
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

export async function openDashboardPresenceLoggedIn(page: Page): Promise<void> {
  await page.goto("/#/");
  await expect(page.getByTestId("presence-change-section")).toBeVisible({
    timeout: 20_000,
  });
  await expect(page.getByTestId("presence-change-loading")).toBeHidden({
    timeout: 20_000,
  });
  await expect(page.getByTestId("presence-change-color-join-me")).toBeVisible();
}
