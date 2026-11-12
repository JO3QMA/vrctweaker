import type { Page } from "@playwright/test";
import { expect } from "@playwright/test";

/**
 * メイン余白E2Eで試すビューポート幅。将来は `app-window-bounds.ts`（最小ウィンドウなど）
 * と揃える想定。現行は600、#396マージ後は768へ差し替え。
 */
export const MAIN_SCROLL_LAYOUT_VIEWPORT_WIDTHS = [1280, 600] as const;

/** main.ts の `meta.selfScroll: true` と対応（ドキュメント用） */
export const SELF_SCROLL_ROUTE_NAMES = [
  "gallery",
  "activity",
  "me",
  "friends",
  "user-profile",
  "automation",
  "settings",
  "settings-account",
  "settings-general",
  "settings-data",
  "settings-about",
  "settings-reset",
] as const;

type RouteScrollLayoutConfig = {
  /** 四辺余白を測る要素 */
  paddingRoot: string;
  /** 右端まで伸びる縦スクロール要素（未指定時は paddingRoot が scroller ならそれを使う） */
  edgeScroller?: string;
  bare?: boolean;
};

/**
 * フレーム右paddingを0にし内側スクロールだけ右端まで伸ばす画面の、
 * 固定ヘッダー／ツールバー（スクロール外）の右余白を検証する。
 */
export const ROUTE_FIXED_CHROME_SELECTORS: Partial<
  Record<string, readonly string[]>
> = {
  gallery: [
    ".gallery-view--scroll-frame .page-title",
    ".gallery-view--scroll-frame .filters",
  ],
};

const ROUTE_SCROLL_LAYOUT: Record<string, RouteScrollLayoutConfig> = {
  dashboard: { paddingRoot: ".router-outlet-host > :first-child" },
  launcher: { paddingRoot: ".router-outlet-host > :first-child" },
  gallery: {
    paddingRoot: ".gallery-view--scroll-frame",
    edgeScroller: '[data-testid="gallery-grid-scroll"]',
  },
  activity: {
    paddingRoot: ".activity-view.main-scroll-region",
  },
  "encounter-history": {
    paddingRoot: ".router-outlet-host > :first-child",
    bare: true,
  },
  me: { paddingRoot: ".self-profile-view.main-scroll-region" },
  friends: { paddingRoot: ".friends-view.main-scroll-region" },
  "user-profile": {
    paddingRoot: ".user-profile-detail-view.main-scroll-region",
  },
  automation: { paddingRoot: ".automation-view.main-scroll-region" },
  config: { paddingRoot: ".router-outlet-host > :first-child" },
  video: { paddingRoot: ".router-outlet-host > :first-child" },
  settings: {
    paddingRoot: ".settings-view.main-scroll-region",
    edgeScroller: ".settings-view.main-scroll-region",
  },
  "settings-account": {
    paddingRoot: ".settings-view.main-scroll-region",
    edgeScroller: ".settings-view.main-scroll-region",
  },
  "settings-general": {
    paddingRoot: ".settings-view.main-scroll-region",
    edgeScroller: ".settings-view.main-scroll-region",
  },
  "settings-data": {
    paddingRoot: ".settings-view.main-scroll-region",
    edgeScroller: ".settings-view.main-scroll-region",
  },
  "settings-about": {
    paddingRoot: ".settings-view.main-scroll-region",
    edgeScroller: ".settings-view.main-scroll-region",
  },
  "settings-reset": {
    paddingRoot: ".settings-view.main-scroll-region",
    edgeScroller: ".settings-view.main-scroll-region",
  },
  licenses: { paddingRoot: ".router-outlet-host > :first-child" },
  "legal-privacy": { paddingRoot: ".router-outlet-host > :first-child" },
  "legal-terms": { paddingRoot: ".router-outlet-host > :first-child" },
};

type LayoutMetrics = {
  mainPaddingTop: number;
  padTop: number;
  padRight: number;
  padBottom: number;
  padLeft: number;
  galleryGridPadRight: number | null;
  edgeScrollerRight: number;
  mainRight: number;
  edgeIsScroller: boolean;
  edgeScrollHeight: number;
  edgeClientHeight: number;
  horizontalOverflow: boolean;
  atRestTopGap: number;
  atRestBottomGap: number;
  scrolledTopGap: number;
};

async function measureLayout(
  page: Page,
  config: RouteScrollLayoutConfig,
): Promise<LayoutMetrics> {
  return page.evaluate((cfg) => {
    const main = document.querySelector(".main-content") as HTMLElement | null;
    const paddingRoot = document.querySelector(
      cfg.paddingRoot,
    ) as HTMLElement | null;
    if (!main || !paddingRoot) {
      throw new Error("main or paddingRoot missing");
    }

    const edgeSel = cfg.edgeScroller ?? cfg.paddingRoot;
    const edgeEl = document.querySelector(edgeSel) as HTMLElement | null;
    if (!edgeEl) {
      throw new Error(`edge/padding scroller missing: ${edgeSel}`);
    }

    const mainStyle = getComputedStyle(main);
    const padStyle = getComputedStyle(paddingRoot);
    const edgeStyle = getComputedStyle(edgeEl);

    const mainRect = main.getBoundingClientRect();
    const edgeRect = edgeEl.getBoundingClientRect();

    const overflowY = edgeStyle.overflowY;
    const edgeIsScroller = overflowY === "auto" || overflowY === "scroll";

    const padFirst = paddingRoot.firstElementChild as HTMLElement | null;
    const padRect = paddingRoot.getBoundingClientRect();
    const atRestTopGap = padFirst
      ? padFirst.getBoundingClientRect().top - padRect.top
      : parseFloat(padStyle.paddingTop);

    const padLast = paddingRoot.lastElementChild as HTMLElement | null;
    let atRestBottomGap = parseFloat(padStyle.paddingBottom);
    const canScrollAtRest =
      edgeIsScroller && edgeEl.scrollHeight > edgeEl.clientHeight + 1;
    if (canScrollAtRest) {
      const prevTop = edgeEl.scrollTop;
      edgeEl.scrollTop = edgeEl.scrollHeight - edgeEl.clientHeight;
      const scrollLast = edgeEl.lastElementChild as HTMLElement | null;
      atRestBottomGap = scrollLast
        ? edgeEl.getBoundingClientRect().bottom -
          scrollLast.getBoundingClientRect().bottom
        : parseFloat(edgeStyle.paddingBottom);
      edgeEl.scrollTop = prevTop;
    } else if (padLast) {
      atRestBottomGap =
        padRect.bottom -
        parseFloat(padStyle.paddingBottom) -
        padLast.getBoundingClientRect().bottom;
    }

    let scrolledTopGap = atRestTopGap;
    if (edgeIsScroller && edgeEl.scrollHeight > edgeEl.clientHeight + 1) {
      edgeEl.scrollTop = Math.min(
        80,
        edgeEl.scrollHeight - edgeEl.clientHeight,
      );
      const firstAfterScroll = edgeEl.firstElementChild as HTMLElement | null;
      scrolledTopGap = firstAfterScroll
        ? firstAfterScroll.getBoundingClientRect().top -
          edgeEl.getBoundingClientRect().top
        : atRestTopGap;
      edgeEl.scrollTop = 0;
    }

    const grid = document.querySelector(
      '[data-testid="gallery-grid-scroll"]',
    ) as HTMLElement | null;
    const galleryGridPadRight = grid
      ? parseFloat(getComputedStyle(grid).paddingRight)
      : null;

    return {
      mainPaddingTop: parseFloat(mainStyle.paddingTop),
      padTop: parseFloat(padStyle.paddingTop),
      padRight: parseFloat(padStyle.paddingRight),
      padBottom: parseFloat(padStyle.paddingBottom),
      padLeft: parseFloat(padStyle.paddingLeft),
      galleryGridPadRight,
      edgeScrollerRight: edgeRect.right,
      mainRight: mainRect.right,
      edgeIsScroller,
      edgeScrollHeight: edgeEl.scrollHeight,
      edgeClientHeight: edgeEl.clientHeight,
      horizontalOverflow:
        main.scrollWidth > main.clientWidth + 1 ||
        document.documentElement.scrollWidth >
          document.documentElement.clientWidth + 1,
      atRestTopGap,
      atRestBottomGap,
      scrolledTopGap,
    };
  }, config);
}

export async function expectRouteMainScrollLayout(
  page: Page,
  routeName: string,
  viewportWidth: number,
): Promise<void> {
  const config = ROUTE_SCROLL_LAYOUT[routeName];
  if (!config) {
    throw new Error(`No scroll layout config for route: ${routeName}`);
  }

  await page.setViewportSize({ width: viewportWidth, height: 800 });

  const expectedPad = config.bare ? 16 : 24;
  const metrics = await measureLayout(page, config);

  expect(metrics.mainPaddingTop, "main-content should not pad").toBe(0);
  expect(metrics.padTop, "padding top").toBeCloseTo(expectedPad, 0);
  expect(metrics.padLeft, "padding left").toBeCloseTo(expectedPad, 0);
  expect(metrics.padBottom, "padding bottom").toBeCloseTo(expectedPad, 0);

  if (routeName === "gallery") {
    expect(metrics.padRight, "gallery frame right pad").toBe(0);
    expect(metrics.galleryGridPadRight, "gallery grid right pad").toBeCloseTo(
      expectedPad,
      0,
    );
    const chromeSelectors = ROUTE_FIXED_CHROME_SELECTORS.gallery;
    if (chromeSelectors?.length) {
      const chrome = await page.evaluate(
        ({ selectors }) => {
          const main = document.querySelector(
            ".main-content",
          ) as HTMLElement | null;
          if (!main) throw new Error("main missing");
          const mainRight = main.getBoundingClientRect().right;
          const pads = selectors.map((sel) => {
            const el = document.querySelector(sel) as HTMLElement | null;
            if (!el) throw new Error(`chrome missing: ${sel}`);
            return parseFloat(getComputedStyle(el).paddingRight);
          });
          const filters = document.querySelector(
            ".gallery-view--scroll-frame .filters",
          ) as HTMLElement | null;
          const lastControl = filters?.querySelector(
            ".filters-actions > *:last-child",
          ) as HTMLElement | null;
          const toolbarVisualGap = lastControl
            ? mainRight - lastControl.getBoundingClientRect().right
            : null;
          return { pads, toolbarVisualGap };
        },
        { selectors: [...chromeSelectors] },
      );
      for (const pad of chrome.pads) {
        expect(pad, "fixed chrome padding-right").toBeCloseTo(expectedPad, 0);
      }
      expect(
        chrome.toolbarVisualGap,
        "toolbar rightmost control inset",
      ).toBeCloseTo(expectedPad, 0);
    }
  } else {
    expect(metrics.padRight, "padding right").toBeCloseTo(expectedPad, 0);
    const chromeSelectors = ROUTE_FIXED_CHROME_SELECTORS[routeName];
    if (chromeSelectors?.length) {
      const pads = await page.evaluate(
        ({ selectors }) =>
          selectors.map((sel) => {
            const el = document.querySelector(sel) as HTMLElement | null;
            if (!el) throw new Error(`chrome missing: ${sel}`);
            return parseFloat(getComputedStyle(el).paddingRight);
          }),
        { selectors: [...chromeSelectors] },
      );
      for (const pad of pads) {
        expect(pad, "fixed chrome padding-right").toBeCloseTo(expectedPad, 0);
      }
    }
  }

  expect(metrics.horizontalOverflow, "no horizontal overflow").toBe(false);

  const edgeFlush = Math.abs(metrics.edgeScrollerRight - metrics.mainRight) < 2;
  const canScroll =
    metrics.edgeIsScroller &&
    metrics.edgeScrollHeight > metrics.edgeClientHeight + 1;

  const usesOutletPadding =
    config.paddingRoot === ".router-outlet-host > :first-child";
  const edgeAlign =
    usesOutletPadding ||
    config.edgeScroller === ".settings-view.main-scroll-region" ||
    config.edgeScroller === '[data-testid="gallery-grid-scroll"]';

  if (canScroll && edgeAlign) {
    expect(edgeFlush, "scrollbar flush to main right").toBe(true);
    expect(
      metrics.scrolledTopGap,
      "scrolled content nears scrollport top",
    ).toBeLessThan(2);
  }

  expect(metrics.atRestTopGap, "at-rest top inset").toBeCloseTo(expectedPad, 0);

  if (
    metrics.edgeIsScroller &&
    metrics.edgeScrollHeight > metrics.edgeClientHeight + 1 &&
    edgeAlign
  ) {
    expect(
      metrics.atRestBottomGap,
      "at-rest bottom inset",
    ).toBeGreaterThanOrEqual(expectedPad - 2);
  }
}
