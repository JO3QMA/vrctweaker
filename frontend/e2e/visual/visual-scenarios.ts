import type { MockWailsOptions } from "../fixtures/mock-wails";
import type { VisualRouteName } from "./visual-routes";
import { VISUAL_VIEWPORTS } from "./visual-routes";

/** スクロール末尾まで撮るルート（主要 selfScroll 画面） */
export const VISUAL_SCROLL_END_SCENARIOS: {
  name: VisualRouteName;
  path: string;
  scroller: string;
}[] = [
  {
    name: "gallery",
    path: "/#/gallery",
    scroller: '[data-testid="gallery-grid-scroll"]',
  },
  {
    name: "activity",
    path: "/#/activity",
    scroller: ".activity-view.main-scroll-region",
  },
  {
    name: "friends",
    path: "/#/friends",
    scroller: ".friends-pane--left .friends-list",
  },
  {
    name: "me",
    path: "/#/me",
    scroller: ".self-profile-view .friend-detail-card .el-card__body",
  },
  {
    name: "automation",
    path: "/#/automation",
    scroller: ".automation-list__rows",
  },
];

/** ビューポート上でスクロール余白が発生しない組み合わせ（撮影対象外） */
export const VISUAL_SCROLL_END_SKIP: {
  name: (typeof VISUAL_SCROLL_END_SCENARIOS)[number]["name"];
  viewportLabel: (typeof VISUAL_VIEWPORTS)[number]["label"];
  reason: string;
}[] = [
  {
    name: "activity",
    viewportLabel: "1280x800",
    reason:
      "プレイ時間展開後も activity-view 全体が収まり、main-scroll-region がオーバーフローしない",
  },
];

export const VISUAL_EMPTY_SCENARIOS: {
  name: string;
  routeName?: VisualRouteName;
  path: string;
  mock: MockWailsOptions;
  readyTestId?: string;
}[] = [
  {
    name: "gallery",
    routeName: "gallery",
    path: "/#/gallery",
    mock: { emptyGallery: true },
    readyTestId: "gallery-empty",
  },
  {
    name: "friends",
    routeName: "friends",
    path: "/#/friends",
    mock: { emptyFriends: true },
  },
  {
    name: "automation",
    routeName: "automation",
    path: "/#/automation",
    mock: { emptyAutomation: true },
  },
  {
    name: "activity",
    routeName: "activity",
    path: "/#/activity",
    mock: { emptyActivity: true },
  },
];

export const VISUAL_DIALOG_SCENARIOS = [
  {
    id: "legal-document",
    labelJa: "法務ダイアログ",
    labelEn: "Legal document dialog",
  },
  {
    id: "app-update",
    labelJa: "アップデートダイアログ",
    labelEn: "App update dialog",
  },
  {
    id: "onboarding",
    labelJa: "オンボーディング",
    labelEn: "Onboarding wizard",
  },
] as const;

export type VisualDialogId = (typeof VISUAL_DIALOG_SCENARIOS)[number]["id"];
