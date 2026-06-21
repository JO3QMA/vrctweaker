/**
 * E2E 用 Wails モックのシードデータ。
 * Playwright spec からも参照可能。mock-wails.ts は JSON 化してブラウザへ注入する。
 */

export interface SeedLaunchProfile {
  id: string;
  name: string;
  arguments: string;
  isDefault: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface SeedPathSettings {
  vrchatPathWindows: string;
  steamPathLinux: string;
  outputLogPath: string;
}

export interface SeedScreenshot {
  id: string;
  filePath: string;
  worldId: string;
  worldName: string;
  takenAt: string;
  fileSizeBytes: number;
}

export interface SeedEncounter {
  id: string;
  vrcUserId: string;
  displayName: string;
  worldId: string;
  worldDisplayName: string;
  instanceId: string;
  joinedAt: string;
  leftAt: string;
}

export interface SeedFriend {
  vrcUserId: string;
  displayName: string;
  status: string;
  isFavorite: boolean;
  lastUpdated: string;
  location?: string;
  statusDescription?: string;
}

export interface SeedDailyPlaySeconds {
  date: string;
  seconds: number;
}

export interface SeedTopWorld {
  worldId: string;
  worldName: string;
  seconds: number;
  sessions: number;
}

export interface SeedActivityStats {
  dailyPlaySeconds: SeedDailyPlaySeconds[];
  topWorlds: SeedTopWorld[];
}

export interface SeedAutomationRule {
  id: string;
  name: string;
  triggerType: string;
  conditionJson: string;
  actionType: string;
  actionPayload: string;
  isEnabled: boolean;
}

export interface SeedUserProfileNavigation {
  user: SeedFriend;
  openInFriendsView: boolean;
}

/** ResolveUserProfileNavigation / user-profile ルート用の代表ユーザー ID */
export const E2E_TEST_USER_ID = "usr_e2e_test";

/** ギャラリー・遭遇履歴（ワールド別）で共有するワールド ID */
export const E2E_WORLD_ID = "wrld_e2e_gallery";

export const E2E_TEST_USER_DISPLAY_NAME = "E2E Test User";

export const SEED_LAUNCH_PROFILES: SeedLaunchProfile[] = [
  {
    id: "profile-1",
    name: "デフォルトプロファイル",
    arguments: "",
    isDefault: true,
    createdAt: "2025-01-01T00:00:00Z",
    updatedAt: "2025-01-01T00:00:00Z",
  },
];

export const SEED_PATH_SETTINGS: SeedPathSettings = {
  vrchatPathWindows: "",
  steamPathLinux: "",
  outputLogPath: "",
};

export const SEED_SCREENSHOTS: SeedScreenshot[] = [
  {
    id: "ss_e2e_001",
    filePath: "C:/VRChat/2025-06-01/VRChat_2025-06-01_12-00-00.png",
    worldId: E2E_WORLD_ID,
    worldName: "E2E Gallery World",
    takenAt: "2025-06-01T12:00:00Z",
    fileSizeBytes: 245_760,
  },
  {
    id: "ss_e2e_002",
    filePath: "C:/VRChat/2025-06-02/VRChat_2025-06-02_18-30-00.png",
    worldId: "wrld_e2e_other",
    worldName: "E2E Other World",
    takenAt: "2025-06-02T18:30:00Z",
    fileSizeBytes: 512_000,
  },
];

export const SEED_ENCOUNTERS: SeedEncounter[] = [
  {
    id: "enc_e2e_001",
    vrcUserId: E2E_TEST_USER_ID,
    displayName: E2E_TEST_USER_DISPLAY_NAME,
    worldId: E2E_WORLD_ID,
    worldDisplayName: "E2E Gallery World",
    instanceId: "inst_e2e_001",
    joinedAt: "2025-06-01T10:00:00Z",
    leftAt: "2025-06-01T11:30:00Z",
  },
  {
    id: "enc_e2e_002",
    vrcUserId: "usr_e2e_visitor",
    displayName: "E2E Visitor",
    worldId: E2E_WORLD_ID,
    worldDisplayName: "E2E Gallery World",
    instanceId: "inst_e2e_001",
    joinedAt: "2025-06-01T10:15:00Z",
    leftAt: "2025-06-01T10:45:00Z",
  },
  {
    id: "enc_e2e_003",
    vrcUserId: E2E_TEST_USER_ID,
    displayName: E2E_TEST_USER_DISPLAY_NAME,
    worldId: "wrld_e2e_other",
    worldDisplayName: "E2E Other World",
    instanceId: "inst_e2e_002",
    joinedAt: "2025-06-02T14:00:00Z",
    leftAt: "2025-06-02T15:00:00Z",
  },
];

export const SEED_FRIENDS: SeedFriend[] = [
  {
    vrcUserId: "usr_e2e_online",
    displayName: "E2E Online Friend",
    status: "active",
    isFavorite: true,
    lastUpdated: "2025-06-01T09:00:00Z",
    location: "wrld_e2e_gallery:12345",
    statusDescription: "E2E オンライン",
  },
  {
    vrcUserId: "usr_e2e_offline",
    displayName: "E2E Offline Friend",
    status: "offline",
    isFavorite: false,
    lastUpdated: "2025-05-30T20:00:00Z",
    location: "offline",
    statusDescription: "",
  },
];

export const SEED_ACTIVITY_STATS: SeedActivityStats = {
  dailyPlaySeconds: [
    { date: "2025-05-28", seconds: 3600 },
    { date: "2025-05-29", seconds: 5400 },
    { date: "2025-05-30", seconds: 1800 },
    { date: "2025-05-31", seconds: 7200 },
    { date: "2025-06-01", seconds: 4500 },
    { date: "2025-06-02", seconds: 2700 },
    { date: "2025-06-03", seconds: 6300 },
  ],
  topWorlds: [
    {
      worldId: E2E_WORLD_ID,
      worldName: "E2E Gallery World",
      seconds: 12_600,
      sessions: 8,
    },
    {
      worldId: "wrld_e2e_other",
      worldName: "E2E Other World",
      seconds: 5400,
      sessions: 3,
    },
  ],
};

export const SEED_AUTOMATION_RULES: SeedAutomationRule[] = [
  {
    id: "rule_e2e_001",
    name: "E2E AFK → Busy",
    triggerType: "afk_detected",
    conditionJson: "{}",
    actionType: "change_status",
    actionPayload: "busy",
    isEnabled: true,
  },
];

export const SEED_VRCHAT_CONFIG = {
  cameraResWidth: 1920,
  cameraResHeight: 1080,
  screenshotResWidth: 1920,
  screenshotResHeight: 1080,
  pictureOutputFolder: "",
  pictureOutputSplitByDate: null,
  fpvSteadycamFov: 0,
  cacheDirectory: "",
  cacheSize: 0,
  cacheExpiryDelay: 0,
  disableRichPresence: null,
};

/** usr_e2e_test 向け ResolveUserProfileNavigation の戻り値 */
export function seedUserProfileNavigation(
  vrcUserId: string,
): SeedUserProfileNavigation {
  if (vrcUserId === E2E_TEST_USER_ID) {
    return {
      user: {
        vrcUserId: E2E_TEST_USER_ID,
        displayName: E2E_TEST_USER_DISPLAY_NAME,
        status: "active",
        isFavorite: false,
        lastUpdated: "2025-06-01T09:00:00Z",
        statusDescription: "E2E プロフィール",
        bio: "E2E テスト用ユーザーです。",
      },
      openInFriendsView: false,
    };
  }
  return {
    user: {
      vrcUserId,
      displayName: "",
      status: "",
      isFavorite: false,
      lastUpdated: "",
    },
    openInFriendsView: false,
  };
}

export function filterScreenshotsByWorldId(
  list: SeedScreenshot[],
  worldId?: string,
): SeedScreenshot[] {
  if (!worldId?.trim()) return list;
  return list.filter((s) => s.worldId === worldId);
}

export function searchSeedScreenshots(
  list: SeedScreenshot[],
  filter: {
    worldId?: string;
    worldName?: string;
    dateFrom?: string;
    dateTo?: string;
  },
): SeedScreenshot[] {
  let result = list;
  if (filter.worldId?.trim()) {
    result = result.filter((s) => s.worldId === filter.worldId);
  }
  if (filter.worldName?.trim()) {
    const q = filter.worldName.trim().toLowerCase();
    result = result.filter((s) => s.worldName.toLowerCase().includes(q));
  }
  if (filter.dateFrom?.trim()) {
    const from = filter.dateFrom.trim();
    result = result.filter((s) => !s.takenAt || s.takenAt.slice(0, 10) >= from);
  }
  if (filter.dateTo?.trim()) {
    const to = filter.dateTo.trim();
    result = result.filter((s) => !s.takenAt || s.takenAt.slice(0, 10) <= to);
  }
  return result;
}

export function encountersByVrcUserId(
  list: SeedEncounter[],
  vrcUserId: string,
): SeedEncounter[] {
  return list.filter((e) => e.vrcUserId === vrcUserId);
}

export function encountersByWorldId(
  list: SeedEncounter[],
  worldId: string,
): SeedEncounter[] {
  return list.filter((e) => e.worldId === worldId);
}
