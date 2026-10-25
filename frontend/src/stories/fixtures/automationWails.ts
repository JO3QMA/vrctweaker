import type { Decorator } from "@storybook/vue3-vite";
import type {
  AutomationItemDTO,
  AutomationRunLogEntryDTO,
  DetectedPowerPlanDTO,
  UserCacheDTO,
} from "../../wails/app";
import { withWailsApp } from "../wailsDecorator";

export const AUTOMATION_STORY_FRIEND_RULE_NAME = "フレンド参加で取り込み中";

export const AUTOMATION_STORY_RUN_LOG: AutomationRunLogEntryDTO[] = [
  {
    at: "2026-10-21 21:04",
    itemId: "rule-story-2",
    itemName: "平日の朝にフレンド参加でステータスを変えるとても長いルール名",
    eventType: "schedule.tick",
    success: false,
    actionsCompleted: 1,
    actionsTotal: 2,
    contextLabel:
      "とても長いコンテキストラベルで、狭い幅でもセル内で折り返すことを確認する文章です",
  },
];

const items: AutomationItemDTO[] = [
  {
    id: "rule-story-1",
    name: AUTOMATION_STORY_FRIEND_RULE_NAME,
    kind: "rule",
    isEnabled: true,
    triggerType: "friend_joined",
    conditionsJson: "[]",
    actionsJson: JSON.stringify([
      { type: "change_status", payload: { status: "busy" } },
    ]),
  },
  {
    id: "rule-story-2",
    name: "平日の朝にフレンド参加でステータスを変えるとても長いルール名",
    kind: "rule",
    isEnabled: false,
    triggerType: "schedule.tick",
    scheduleJson: JSON.stringify({
      weekdays: [1, 2, 3, 4, 5],
      hour: 8,
      minute: 30,
    }),
    conditionsJson: JSON.stringify([{ type: "vrchat_running" }]),
    actionsJson: JSON.stringify([
      { type: "change_status", payload: { status: "ask me" } },
      {
        type: "set_vrchat_window_size",
        payload: { width: 1920, height: 1080 },
        continueOnError: true,
      },
    ]),
  },
  {
    id: "script-story-1",
    name: "参加時のメモ",
    kind: "script",
    isEnabled: true,
    scriptSource:
      'tweaker.subscribe("friend_joined", function(ev, payload)\n  -- story\nend)\n',
  },
];

const runLog = AUTOMATION_STORY_RUN_LOG;

const friends: UserCacheDTO[] = [
  {
    vrcUserId: "story-friend-1",
    displayName: "サンプルフレンド",
    status: "active",
    isFavorite: false,
    lastUpdated: "2026-10-21T00:00:00Z",
  },
];

const powerPlans: DetectedPowerPlanDTO[] = [
  { guid: "plan-balanced", name: "バランス" },
  { guid: "plan-high", name: "高パフォーマンス" },
];

export function withAutomationWails(): Decorator {
  return withWailsApp({
    ListAutomationItems: () =>
      Promise.resolve(items.map((item) => ({ ...item }))),
    SaveAutomationItem: () => Promise.resolve(),
    DeleteAutomationItem: () => Promise.resolve(),
    ToggleAutomationItem: () => Promise.resolve(),
    GetAutomationRunLog: () =>
      Promise.resolve(runLog.map((entry) => ({ ...entry }))),
    GetAutomationRuntimeStatus: () =>
      Promise.resolve({ available: true, reasonKey: "" }),
    ListDetectedPowerPlans: () =>
      Promise.resolve(powerPlans.map((plan) => ({ ...plan }))),
    Friends: () => Promise.resolve(friends.map((friend) => ({ ...friend }))),
  });
}
