// Wails app bindings - calls Go backend methods
// When running in Wails, window.go.main.App is injected

export interface LaunchProfileDTO {
  id: string;
  name: string;
  arguments: string;
  isDefault: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface LaunchArgsParsedDTO {
  noVr: boolean; // -no-vr (デスクトップモード)
  screenMode: "" | "fullscreen" | "windowed" | "popupwindow";
  screenWidth: number;
  screenHeight: number;
  fps: number;
  skipRegistry: boolean;
  processPriority: number; // -2..2, -999=omit
  mainThreadPriority: number; // -2..2, -999=omit
  monitor: number; // 1-based, 0=omit
  profile: number; // --profile=X, -1=omit
  enableDebugGui: boolean;
  enableSDKLogLevels: boolean;
  enableUdonDebugLogging: boolean;
  midi: string;
  watchWorlds: boolean;
  watchAvatars: boolean;
  ignoreTrackers: string;
  videoDecoding: "" | "software" | "hardware";
  disableAMDStutterWorkaround: boolean;
  osc: string;
  affinity: string;
  enforceWorldServerChecks: boolean;
  custom: string;
}

/** -999 = omit for process/main thread priority */
export const PRIORITY_OMIT = -999;

export interface ScreenshotDTO {
  id: string;
  filePath: string;
  worldId: string;
  worldName: string;
  takenAt?: string;
}

export interface ScreenshotSearchDTO {
  worldId?: string;
  worldName?: string;
  dateFrom?: string;
  dateTo?: string;
}

export interface UserEncounterDTO {
  id: string;
  vrcUserId: string;
  displayName: string;
  action: string;
  instanceId: string;
  encounteredAt: string;
}

export interface FriendCacheDTO {
  vrcUserId: string;
  displayName: string;
  status: string;
  isFavorite: boolean;
  lastUpdated: string;
}

export interface PathSettingsDTO {
  vrchatPathWindows: string;
  steamPathLinux: string;
  outputLogPath: string;
}

export interface LoginResultDTO {
  ok: boolean;
  error?: string;
}

export interface DailyPlaySecondsDTO {
  date: string;
  seconds: number;
}

export interface TopWorldDTO {
  worldId: string;
  worldName?: string;
  seconds: number;
  sessions: number;
}

export interface ActivityStatsDTO {
  dailyPlaySeconds: DailyPlaySecondsDTO[];
  topWorlds: TopWorldDTO[];
}

export interface AutomationRuleDTO {
  id: string;
  name: string;
  triggerType: string;
  conditionJson: string;
  actionType: string;
  actionPayload: string;
  isEnabled: boolean;
}

interface AppBindings {
  Greet(name: string): Promise<string>;
  LaunchProfiles(): Promise<LaunchProfileDTO[]>;
  LaunchVRChat(profileID: string): Promise<void>;
  LaunchVRChatWithArgs(args: string): Promise<void>;
  ParseLaunchArgsForGUI(args: string): Promise<LaunchArgsParsedDTO>;
  MergeLaunchArgsForGUI(dto: LaunchArgsParsedDTO): Promise<string>;
  JoinWorld(worldId: string): Promise<void>;
  JoinWorldFromScreenshot(screenshotId: string): Promise<void>;
  SaveLaunchProfile(p: LaunchProfileDTO): Promise<void>;
  DeleteLaunchProfile(id: string): Promise<void>;
  GetLogRetentionDays(): Promise<number>;
  SetLogRetentionDays(days: number): Promise<void>;
  GetPathSettings(): Promise<PathSettingsDTO>;
  SetPathSettings(dto: PathSettingsDTO): Promise<void>;
  ValidatePath(path: string): Promise<boolean>;
  Screenshots(worldId?: string): Promise<ScreenshotDTO[]>;
  SearchScreenshots(filter: ScreenshotSearchDTO): Promise<ScreenshotDTO[]>;
  GetScreenshot(id: string): Promise<ScreenshotDTO | null>;
  ScanScreenshotDir(path: string): Promise<number>;
  ReindexScreenshotDir(path: string): Promise<number>;
  Encounters(): Promise<UserEncounterDTO[]>;
  RotateEncounters(): Promise<number>;
  GetActivityStats(fromISO: string, toISO: string): Promise<ActivityStatsDTO>;
  Friends(): Promise<FriendCacheDTO[]>;
  SetFavorite(vrcUserId: string, favorite: boolean): Promise<void>;
  SetStatus(status: string): Promise<void>;
  Login(
    username: string,
    password: string,
    twoFactorCode?: string,
  ): Promise<LoginResultDTO>;
  Logout(): Promise<void>;
  IsLoggedIn(): Promise<boolean>;
  RefreshFriends(): Promise<void>;
  VacuumDb(): Promise<void>;
  ClearEncounters(): Promise<number>;
  ClearScreenshots(): Promise<number>;
  ClearFriendsCache(): Promise<number>;
  ListAutomationRules(): Promise<AutomationRuleDTO[]>;
  SaveAutomationRule(rule: AutomationRuleDTO): Promise<void>;
  DeleteAutomationRule(id: string): Promise<void>;
  ToggleAutomationRule(id: string, enabled: boolean): Promise<void>;
}

declare global {
  interface Window {
    go?: {
      main?: {
        App?: AppBindings;
      };
    };
  }
}

function getApp(): AppBindings | undefined {
  return typeof window !== "undefined" ? window.go?.main?.App : undefined;
}

export async function callApp<T>(
  fn: (app: AppBindings) => Promise<T>,
  fallback: T,
): Promise<T> {
  const app = getApp();
  if (!app) {
    return fallback;
  }
  return fn(app);
}

export const App = {
  async greet(name: string): Promise<string> {
    return callApp((a) => a.Greet(name), `Hello ${name}, Welcome!`);
  },
  async launchProfiles(): Promise<LaunchProfileDTO[]> {
    return callApp((a) => a.LaunchProfiles(), []);
  },
  async launchVRChat(profileID: string): Promise<void> {
    return callApp((a) => a.LaunchVRChat(profileID), undefined);
  },
  async launchVRChatWithArgs(args: string): Promise<void> {
    return callApp((a) => a.LaunchVRChatWithArgs(args), undefined);
  },
  async parseLaunchArgsForGUI(args: string): Promise<LaunchArgsParsedDTO> {
    return callApp((a) => a.ParseLaunchArgsForGUI(args), {
      noVr: false,
      screenMode: "",
      screenWidth: 0,
      screenHeight: 0,
      fps: 0,
      skipRegistry: false,
      processPriority: PRIORITY_OMIT,
      mainThreadPriority: PRIORITY_OMIT,
      monitor: 0,
      profile: -1,
      enableDebugGui: false,
      enableSDKLogLevels: false,
      enableUdonDebugLogging: false,
      midi: "",
      watchWorlds: false,
      watchAvatars: false,
      ignoreTrackers: "",
      videoDecoding: "",
      disableAMDStutterWorkaround: false,
      osc: "",
      affinity: "",
      enforceWorldServerChecks: false,
      custom: "",
    });
  },
  async mergeLaunchArgsForGUI(dto: LaunchArgsParsedDTO): Promise<string> {
    return callApp((a) => a.MergeLaunchArgsForGUI(dto), "");
  },
  async joinWorld(worldId: string): Promise<void> {
    return callApp((a) => a.JoinWorld(worldId), undefined);
  },
  async joinWorldFromScreenshot(screenshotId: string): Promise<void> {
    return callApp((a) => a.JoinWorldFromScreenshot(screenshotId), undefined);
  },
  async saveLaunchProfile(p: LaunchProfileDTO): Promise<void> {
    return callApp((a) => a.SaveLaunchProfile(p), undefined);
  },
  async deleteLaunchProfile(id: string): Promise<void> {
    return callApp((a) => a.DeleteLaunchProfile(id), undefined);
  },
  async getLogRetentionDays(): Promise<number> {
    return callApp((a) => a.GetLogRetentionDays(), 30);
  },
  async setLogRetentionDays(days: number): Promise<void> {
    return callApp((a) => a.SetLogRetentionDays(days), undefined);
  },
  async getPathSettings(): Promise<PathSettingsDTO> {
    return callApp((a) => a.GetPathSettings(), {
      vrchatPathWindows: "",
      steamPathLinux: "",
      outputLogPath: "",
    });
  },
  async setPathSettings(dto: PathSettingsDTO): Promise<void> {
    return callApp((a) => a.SetPathSettings(dto), undefined);
  },
  async validatePath(path: string): Promise<boolean> {
    return callApp((a) => a.ValidatePath(path), false);
  },
  async screenshots(worldId?: string): Promise<ScreenshotDTO[]> {
    return callApp((a) => a.Screenshots(worldId || ""), []);
  },
  async searchScreenshots(
    filter: ScreenshotSearchDTO,
  ): Promise<ScreenshotDTO[]> {
    return callApp((a) => a.SearchScreenshots(filter), []);
  },
  async getScreenshot(id: string): Promise<ScreenshotDTO | null> {
    return callApp((a) => a.GetScreenshot(id), null);
  },
  async scanScreenshotDir(path: string): Promise<number> {
    return callApp((a) => a.ScanScreenshotDir(path), 0);
  },
  async reindexScreenshotDir(path: string): Promise<number> {
    return callApp((a) => a.ReindexScreenshotDir(path), 0);
  },
  async friends(): Promise<FriendCacheDTO[]> {
    return callApp((a) => a.Friends(), []);
  },
  async setFavorite(vrcUserId: string, favorite: boolean): Promise<void> {
    return callApp((a) => a.SetFavorite(vrcUserId, favorite), undefined);
  },
  async setStatus(status: string): Promise<void> {
    return callApp((a) => a.SetStatus(status), undefined);
  },
  async login(
    username: string,
    password: string,
    twoFactorCode?: string,
  ): Promise<LoginResultDTO> {
    return callApp((a) => a.Login(username, password, twoFactorCode ?? ""), {
      ok: false,
      error: "App not available",
    });
  },
  async logout(): Promise<void> {
    return callApp((a) => a.Logout(), undefined);
  },
  async isLoggedIn(): Promise<boolean> {
    return callApp((a) => a.IsLoggedIn(), false);
  },
  async refreshFriends(): Promise<void> {
    return callApp((a) => a.RefreshFriends(), undefined);
  },
  async vacuumDb(): Promise<void> {
    return callApp((a) => a.VacuumDb(), undefined);
  },
  async encounters(): Promise<UserEncounterDTO[]> {
    return callApp((a) => a.Encounters(), []);
  },
  async clearEncounters(): Promise<number> {
    return callApp((a) => a.ClearEncounters(), 0);
  },
  async getActivityStats(
    fromISO: string,
    toISO: string,
  ): Promise<ActivityStatsDTO> {
    return callApp((a) => a.GetActivityStats(fromISO, toISO), {
      dailyPlaySeconds: [],
      topWorlds: [],
    });
  },
  async clearScreenshots(): Promise<number> {
    return callApp((a) => a.ClearScreenshots(), 0);
  },
  async clearFriendsCache(): Promise<number> {
    return callApp((a) => a.ClearFriendsCache(), 0);
  },
  async listAutomationRules(): Promise<AutomationRuleDTO[]> {
    return callApp((a) => a.ListAutomationRules(), []);
  },
  async saveAutomationRule(rule: AutomationRuleDTO): Promise<void> {
    return callApp((a) => a.SaveAutomationRule(rule), undefined);
  },
  async deleteAutomationRule(id: string): Promise<void> {
    return callApp((a) => a.DeleteAutomationRule(id), undefined);
  },
  async toggleAutomationRule(id: string, enabled: boolean): Promise<void> {
    return callApp((a) => a.ToggleAutomationRule(id, enabled), undefined);
  },
};
