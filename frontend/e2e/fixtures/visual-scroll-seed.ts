import type {
  SeedAutomationItem,
  SeedEncounter,
  SeedFriend,
  SeedScreenshot,
} from "./seed-data";
import {
  E2E_ENCOUNTER_INSTANCE_KEY,
  E2E_SELF_USER_ID,
  E2E_TEST_USER_DISPLAY_NAME,
  E2E_TEST_USER_ID,
  E2E_WORLD_ID,
  SEED_SELF_PROFILE,
} from "./seed-data";

const SCREENSHOT_COUNT = 48;
const FRIEND_COUNT = 48;
const AUTOMATION_COUNT = 32;

export const VISUAL_RICH_SCREENSHOTS: SeedScreenshot[] = Array.from(
  { length: SCREENSHOT_COUNT },
  (_, i) => ({
    id: `ss_visual_${String(i + 1).padStart(3, "0")}`,
    filePath: `C:/VRChat/2024/ss_visual_${i + 1}.png`,
    worldId: i % 3 === 0 ? E2E_WORLD_ID : "wrld_e2e_other",
    worldName: i % 3 === 0 ? "E2E Gallery World" : `E2E World ${i + 1}`,
    takenAt: `2024-06-${String((i % 14) + 1).padStart(2, "0")}T${String(10 + (i % 8)).padStart(2, "0")}:00:00Z`,
    fileSizeBytes: 180_000 + i * 1024,
  }),
);

export const VISUAL_RICH_FRIENDS: SeedFriend[] = Array.from(
  { length: FRIEND_COUNT },
  (_, i) => ({
    vrcUserId: `usr_e2e_visual_friend_${String(i + 1).padStart(2, "0")}`,
    displayName: `E2E Friend ${i + 1}`,
    status: i % 4 === 0 ? "active" : "offline",
    isFavorite: i % 5 === 0,
    lastUpdated: "2024-06-10T12:00:00Z",
    location: i % 4 === 0 ? "private" : "offline",
    statusDescription: i % 7 === 0 ? "Visual seed friend" : "",
  }),
);

export const VISUAL_RICH_ENCOUNTERS: SeedEncounter[] = Array.from(
  { length: 32 },
  (_, i) => ({
    id: `enc_visual_${String(i + 1).padStart(3, "0")}`,
    vrcUserId: i % 3 === 0 ? E2E_TEST_USER_ID : `usr_e2e_visitor_${i}`,
    displayName:
      i % 3 === 0 ? E2E_TEST_USER_DISPLAY_NAME : `E2E Visitor ${i + 1}`,
    worldId: i % 2 === 0 ? E2E_WORLD_ID : "wrld_e2e_other",
    worldDisplayName: i % 2 === 0 ? "E2E Gallery World" : "E2E Other World",
    instanceId: E2E_ENCOUNTER_INSTANCE_KEY,
    joinedAt: `2024-06-${String((i % 14) + 1).padStart(2, "0")}T10:00:00Z`,
    leftAt: `2024-06-${String((i % 14) + 1).padStart(2, "0")}T11:00:00Z`,
  }),
);

export const VISUAL_RICH_AUTOMATION_ITEMS: SeedAutomationItem[] = Array.from(
  { length: AUTOMATION_COUNT },
  (_, i) => ({
    id: `rule_visual_${String(i + 1).padStart(3, "0")}`,
    name: `E2E Visual Rule ${i + 1}`,
    kind: "rule",
    isEnabled: i % 2 === 0,
    triggerType: i % 2 === 0 ? "friend_joined" : "schedule.tick",
    scheduleJson:
      i % 2 === 1
        ? JSON.stringify({ weekdays: [1, 2, 3], hour: 8, minute: 0 })
        : undefined,
    conditionsJson: "[]",
    actionsJson: JSON.stringify([
      { type: "change_status", payload: { status: "join me" } },
    ]),
  }),
);

export const VISUAL_RICH_SELF_PROFILE = {
  ...SEED_SELF_PROFILE,
  bio: Array.from(
    { length: 40 },
    (_, i) => `E2E自己プロフィールの長文行${i + 1}。視覚テスト用シードです。`,
  ).join("\n"),
  statusDescription: "E2E visual rich self profile",
};

export const VISUAL_RICH_SELF_USER_ID = E2E_SELF_USER_ID;
