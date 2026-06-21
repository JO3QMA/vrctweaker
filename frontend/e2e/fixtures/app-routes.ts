import {
  E2E_TEST_USER_DISPLAY_NAME,
  E2E_TEST_USER_ID,
  E2E_WORLD_ID,
} from "./seed-data";

const userProfileQuery = new URLSearchParams({
  vrcUserId: E2E_TEST_USER_ID,
  displayName: E2E_TEST_USER_DISPLAY_NAME,
}).toString();

const encounterHistoryUserQuery = new URLSearchParams({
  kind: "user",
  vrcUserId: E2E_TEST_USER_ID,
}).toString();

/** frontend/src/main.ts の全 11 ルート（hash ルーター。Playwright では /#/path 形式で遷移） */
export const APP_ROUTES = [
  { path: "/", name: "dashboard", titleJa: "ダッシュボード" },
  { path: "/launcher", name: "launcher", titleJa: "ランチャー" },
  { path: "/gallery", name: "gallery", titleJa: "ギャラリー" },
  { path: "/activity", name: "activity", titleJa: "アクティビティ" },
  {
    path: `/activity/encounter-history?${encounterHistoryUserQuery}`,
    name: "encounter-history",
    titleJa: "遭遇履歴",
  },
  { path: "/friends", name: "friends", titleJa: "フレンド" },
  {
    path: `/user-profile?${userProfileQuery}`,
    name: "user-profile",
    titleJa: "ユーザー",
  },
  { path: "/automation", name: "automation", titleJa: "オートメーション" },
  { path: "/config", name: "config", titleJa: "その他の設定" },
  { path: "/settings", name: "settings", titleJa: "設定" },
  { path: "/licenses", name: "licenses", titleJa: "OSSライセンス" },
] as const;

/** 遭遇履歴（ワールド別）の例。APP_ROUTES 外の補助定数 */
export const ENCOUNTER_HISTORY_WORLD_PATH = `/activity/encounter-history?${new URLSearchParams(
  {
    kind: "world",
    worldId: E2E_WORLD_ID,
  },
).toString()}`;
