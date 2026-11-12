import type { MockWailsOptions } from "../fixtures/mock-wails";
import type { VisualRouteName } from "./visual-routes";

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
    scroller: ".friends-view.main-scroll-region",
  },
  {
    name: "me",
    path: "/#/me",
    scroller: ".self-profile-view.main-scroll-region",
  },
  {
    name: "automation",
    path: "/#/automation",
    scroller: ".automation-view.main-scroll-region",
  },
  {
    name: "settings",
    path: "/#/settings/general",
    scroller: ".settings-view.main-scroll-region",
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
