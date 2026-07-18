// Wails app bindings - calls Go backend methods
// When running in Wails, window.go.main.App is injected

import {
  activity,
  automation,
  launcher,
  main,
  usecase,
} from "../../wailsjs/go/models";
import type * as WailsApp from "../../wailsjs/go/main/App";

/** Data fields only (wailsjs model classes may include convertValues). */
type WailsDTO<T> = Omit<T, "convertValues">;

export type LaunchProfileDTO = WailsDTO<main.LaunchProfileDTO>;
export type DashboardRejoinDTO = WailsDTO<main.DashboardRejoinDTO>;
export type DashboardLaunchBlockDTO = WailsDTO<main.DashboardLaunchBlockDTO>;
export type ServerStatusDTO = WailsDTO<main.ServerStatusDTO>;
export type ServerStatusSummaryDTO = WailsDTO<main.ServerStatusSummaryDTO>;
export type ServerStatusComponentDTO = WailsDTO<main.ServerStatusComponentDTO>;
export type ServerStatusHeadlineDTO = WailsDTO<main.ServerStatusHeadlineDTO>;
export type LaunchArgsParsedDTO = WailsDTO<launcher.LaunchArgsParsed>;
export type ScreenshotDTO = WailsDTO<main.ScreenshotDTO>;
export type ScreenshotSearchDTO = WailsDTO<main.ScreenshotSearchDTO>;
export type UserEncounterDTO = WailsDTO<main.UserEncounterDTO>;
export type UserCacheDTO = WailsDTO<main.UserCacheDTO>;
export type UserProfileNavigationDTO = WailsDTO<main.UserProfileNavigationDTO>;
export type PathSettingsDTO = WailsDTO<usecase.PathSettings>;
export type YTDLPMaintainStatusDTO = WailsDTO<usecase.YTDLPMaintainStatus>;
export type LoginResultDTO = WailsDTO<main.LoginResultDTO>;
export type VRChatCurrentUserDTO = WailsDTO<main.VRChatCurrentUserDTO>;
export type DailyPlaySecondsDTO = WailsDTO<activity.DailyPlaySeconds>;
export type TopWorldDTO = WailsDTO<activity.TopWorldSummary>;
export type ActivityStatsDTO = WailsDTO<activity.ActivityStats>;
export type AutomationRuleDTO = WailsDTO<automation.AutomationRule>;
export type VRChatConfigDTO = WailsDTO<main.VRChatConfigDTO>;

/** Cookie linkage status (usecase.CookieLinkageStatus); defined locally until wails generate. */
export type CookieLinkageStatusDTO = {
  supported: boolean;
  unsupportedReason?: string;
  enabled: boolean;
  sourceKind: string;
  browser?: string;
  cookiesFilePath?: string;
  configPath?: string;
  riskAcknowledged: boolean;
};

type CookieLinkageAppBindings = {
  GetYTDLPCookieLinkageStatus: () => Promise<CookieLinkageStatusDTO>;
  AcknowledgeYTDLPCookieLinkageRisk: () => Promise<void>;
  SetYTDLPCookieLinkageBrowser: (browser: string) => Promise<void>;
  SetYTDLPCookieLinkageCookiesFile: (path: string) => Promise<void>;
  DisableYTDLPCookieLinkage: () => Promise<void>;
};

function emptyServerStatus(): ServerStatusDTO {
  return {
    fetchState: "unavailable",
    summary: { indicator: "", description: "" },
    components: [],
    incidents: [],
    maintenances: [],
  };
}

function emptyDashboardLaunchBlock(): DashboardLaunchBlockDTO {
  return {
    profiles: [],
    selectedProfileId: "",
    rejoin: undefined,
  };
}

function emptyYTDLPMaintainStatus(): YTDLPMaintainStatusDTO {
  return {
    supported: false,
    unsupportedReason: "",
    maintainDesired: false,
    riskAcknowledged: false,
    effectiveOfficial: false,
    cachePresent: false,
    cacheVersion: "",
    toolsPath: "",
    cachePath: "",
    pendingError: "",
    latestVersion: "",
    latestTag: "",
    latestDownloadUrl: "",
    latestError: "",
  };
}

function emptyCookieLinkageStatus(): CookieLinkageStatusDTO {
  return {
    supported: false,
    unsupportedReason: "",
    enabled: false,
    sourceKind: "",
    browser: "",
    cookiesFilePath: "",
    configPath: "",
    riskAcknowledged: false,
  };
}

function asCookieApp(a: AppBindings): AppBindings & CookieLinkageAppBindings {
  return a as AppBindings & CookieLinkageAppBindings;
}

/** -999 = omit for process/main thread priority */
export const PRIORITY_OMIT = -999;

const EMPTY_LAUNCH_ARGS: LaunchArgsParsedDTO = {
  noVr: false,
  screenMode: "",
  screenWidth: 0,
  screenHeight: 0,
  fps: 90,
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
};

const EMPTY_PATH_SETTINGS: PathSettingsDTO = {
  vrchatPathWindows: "",
  steamPathLinux: "",
  outputLogPath: "",
};

const EMPTY_USER_CACHE: UserCacheDTO = {
  vrcUserId: "",
  displayName: "",
  status: "",
  isFavorite: false,
  lastUpdated: "",
};

const EMPTY_VRCHAT_CURRENT_USER: VRChatCurrentUserDTO = {
  id: "",
  displayName: "",
  username: "",
  status: "",
  statusDescription: "",
  state: "",
  currentAvatarThumbnailImageUrl: "",
  userIcon: "",
  profilePicOverrideThumbnail: "",
};

const EMPTY_ACTIVITY_STATS: ActivityStatsDTO = {
  dailyPlaySeconds: [],
  topWorlds: [],
};

const EMPTY_VRCHAT_CONFIG: VRChatConfigDTO = {
  cameraResWidth: 0,
  cameraResHeight: 0,
  screenshotResWidth: 0,
  screenshotResHeight: 0,
  pictureOutputFolder: "",
  pictureOutputSplitByDate: undefined,
  fpvSteadycamFov: 0,
  cacheDirectory: "",
  cacheSize: 0,
  cacheExpiryDelay: 0,
  disableRichPresence: undefined,
};

const LOGIN_UNAVAILABLE: LoginResultDTO = {
  ok: false,
  error: "App not available",
};

/** Payload for Wails event gallery:scan-progress (usecase.ScanProgress). */
export interface ScanProgressPayload {
  phase: string;
  current: number;
  total: number;
  item?: string;
}

/** Payload for Wails event gallery:scan-done (usecase.GalleryScanDone). */
export interface GalleryScanDonePayload {
  count: number;
  error?: string;
  cancelled?: boolean;
}

export type AppBindings = {
  [K in keyof typeof WailsApp]: (typeof WailsApp)[K];
};

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

/** True when index was built to load Wails IPC/runtime (Vite dev injection or Wails-served HTML). */
function pageExpectsWailsBindings(): boolean {
  if (typeof document === "undefined") {
    return false;
  }
  for (const el of document.querySelectorAll("head script[src]")) {
    const src = el.getAttribute("src") ?? "";
    if (src.includes("wails/runtime") || src.includes("wails/ipc")) {
      return true;
    }
  }
  return false;
}

let wailsDevScriptReloadTried = false;

/**
 * Vite プロキシのレースや一時的な取得失敗で /wails/*.js が実行されないことがある。
 * head から該当 script を外し、キャッシュバスト付きで公式どおり ipc → runtime の順で再挿入する（1 回だけ）。
 */
function reloadDevWailsScriptsOnce(): Promise<void> {
  return new Promise((resolve, reject) => {
    if (typeof document === "undefined") {
      resolve();
      return;
    }
    for (const el of document.querySelectorAll('head script[src*="wails/"]')) {
      el.remove();
    }
    const stamp = Date.now();
    const ipc = document.createElement("script");
    ipc.src = `/wails/ipc.js?wailsRetry=${stamp}`;
    ipc.async = false;
    ipc.onload = () => {
      const runtime = document.createElement("script");
      runtime.src = `/wails/runtime.js?wailsRetry=${stamp}`;
      runtime.async = false;
      runtime.onload = () => resolve();
      runtime.onerror = () =>
        reject(new Error("failed to load /wails/runtime.js after retry"));
      document.head.appendChild(runtime);
    };
    ipc.onerror = () =>
      reject(new Error("failed to load /wails/ipc.js after retry"));
    document.head.appendChild(ipc);
  });
}

/**
 * Wails dev over Vite can leave `/wails/*.js` in the DOM before `window.go` is
 * ready (IPC/WebSocket race). Wait with rAF (no setTimeout) before falling back.
 *
 * Worst case in `callApp`: up to 360 frames here, then (once per page load) script
 * reload plus 180 more frames — roughly ~6s at 60fps. Only affects `wails dev`
 * startup races; Vitest sets `MODE === "test"` and skips this path.
 */
function waitForAppBindings(
  maxFrames: number,
): Promise<AppBindings | undefined> {
  return new Promise((resolve) => {
    let frames = 0;
    function onFrame() {
      const app = getApp();
      if (app) {
        resolve(app);
        return;
      }
      if (++frames >= maxFrames) {
        resolve(undefined);
        return;
      }
      requestAnimationFrame(onFrame);
    }
    requestAnimationFrame(onFrame);
  });
}

/** True when running inside Wails (second windows from window.open cannot load wails.localhost). */
export function isWailsRuntime(): boolean {
  return getApp() !== undefined;
}

/**
 * Invokes a Wails `App` binding when `window.go.main.App` exists.
 *
 * `fallback` is returned **only** when that binding is missing (e.g. plain browser
 * or tests without Wails). It is **not** used when Go returns an error: in that
 * case the promise from `fn(app)` rejects and the error propagates. Callers must
 * use try/catch or `.catch()` for backend failures — do not assume errors are
 * swallowed or replaced by `fallback`.
 *
 * In DEV, when the page includes Wails script tags, this may wait many rAF ticks
 * (see `waitForAppBindings`) and optionally reload scripts once before giving up.
 */
export async function callApp<T>(
  fn: (app: AppBindings) => Promise<T>,
  fallback: T,
): Promise<T> {
  let app = getApp();
  if (
    !app &&
    import.meta.env.DEV &&
    import.meta.env.MODE !== "test" &&
    pageExpectsWailsBindings()
  ) {
    app = await waitForAppBindings(360);
  }
  if (
    !app &&
    import.meta.env.DEV &&
    import.meta.env.MODE !== "test" &&
    pageExpectsWailsBindings() &&
    !wailsDevScriptReloadTried
  ) {
    wailsDevScriptReloadTried = true;
    try {
      await reloadDevWailsScriptsOnce();
      app = await waitForAppBindings(180);
    } catch {
      /* keep app undefined */
    }
  }
  if (!app) {
    return fallback;
  }
  return fn(app);
}

/** Maps camelCase App API to a Wails binding with a static fallback. */
function bindGo<TArgs extends unknown[], TResult>(
  invoke: (app: AppBindings, ...args: TArgs) => Promise<TResult>,
  fallback: TResult,
): (...args: TArgs) => Promise<TResult> {
  return (...args) => callApp((a) => invoke(a, ...args), fallback);
}

async function nullableStringDialog(
  invoke: (app: AppBindings) => Promise<string>,
): Promise<string | null> {
  const result = await callApp(invoke, "");
  return result && result !== "" ? result : null;
}

function emptyUserProfileNavigation(
  vrcUserID: string,
): UserProfileNavigationDTO {
  return {
    user: { ...EMPTY_USER_CACHE, vrcUserId: vrcUserID },
    openInFriendsView: false,
    openInSelfProfile: false,
  };
}

export const App = {
  launchProfiles: bindGo((a) => a.LaunchProfiles(), []),
  launchVRChat: bindGo(
    (a, profileID: string) => a.LaunchVRChat(profileID),
    undefined,
  ),
  launchVRChatWithArgs: bindGo(
    (a, args: string, lastLaunchProfileID: string = "") =>
      a.LaunchVRChatWithArgs(args, lastLaunchProfileID),
    undefined,
  ),
  getDashboardLaunchBlock: bindGo(
    (a) => a.GetDashboardLaunchBlock(),
    emptyDashboardLaunchBlock(),
  ),
  getServerStatus: bindGo((a) => a.GetServerStatus(), emptyServerStatus()),
  instanceRejoin: bindGo(
    (a, profileID: string, playSessionID: string) =>
      a.InstanceRejoin(profileID, playSessionID),
    undefined,
  ),
  parseLaunchArgsForGUI: bindGo(
    (a, args: string) => a.ParseLaunchArgsForGUI(args),
    EMPTY_LAUNCH_ARGS,
  ),
  mergeLaunchArgsForGUI: bindGo(
    (a, dto: LaunchArgsParsedDTO) => a.MergeLaunchArgsForGUI(dto),
    "",
  ),
  joinWorld: bindGo((a, worldId: string) => a.JoinWorld(worldId), undefined),
  joinWorldFromScreenshot: bindGo(
    (a, screenshotId: string) => a.JoinWorldFromScreenshot(screenshotId),
    undefined,
  ),
  saveLaunchProfile: bindGo(
    (a, p: LaunchProfileDTO) => a.SaveLaunchProfile(p),
    undefined,
  ),
  deleteLaunchProfile: bindGo(
    (a, id: string) => a.DeleteLaunchProfile(id),
    undefined,
  ),
  getLogRetentionDays: bindGo((a) => a.GetLogRetentionDays(), 30),
  setLogRetentionDays: bindGo(
    (a, days: number) => a.SetLogRetentionDays(days),
    undefined,
  ),
  getLanguage: bindGo((a) => a.GetLanguage(), ""),
  setLanguage: bindGo((a, lang: string) => a.SetLanguage(lang), undefined),
  getSystemLocale: bindGo((a) => a.GetSystemLocale(), "en"),
  getPathSettings: bindGo((a) => a.GetPathSettings(), EMPTY_PATH_SETTINGS),
  setPathSettings: bindGo(
    (a, dto: PathSettingsDTO) => a.SetPathSettings(dto),
    undefined,
  ),
  getSuppressSleepWhileVRChat: bindGo(
    (a) => a.GetSuppressSleepWhileVRChat(),
    false,
  ),
  setSuppressSleepWhileVRChat: bindGo(
    (a, on: boolean) => a.SetSuppressSleepWhileVRChat(on),
    undefined,
  ),
  runtimeIsWindows: bindGo((a) => a.RuntimeIsWindows(), false),
  getYTDLPMaintainStatus: bindGo(
    (a) => a.GetYTDLPMaintainStatus(),
    emptyYTDLPMaintainStatus(),
  ),
  acknowledgeYTDLPToolsReplaceRisk: bindGo(
    (a) => a.AcknowledgeYTDLPToolsReplaceRisk(),
    undefined,
  ),
  setYTDLPToolsReplaceMaintain: bindGo(
    (a, on: boolean) => a.SetYTDLPToolsReplaceMaintain(on),
    undefined,
  ),
  checkYTDLPLatestRelease: bindGo(
    (a) => a.CheckYTDLPLatestRelease(),
    emptyYTDLPMaintainStatus(),
  ),
  updateOfficialYTDLPCache: bindGo(
    (a, downloadURL: string, latestTag: string) =>
      a.UpdateOfficialYTDLPCache(downloadURL, latestTag),
    emptyYTDLPMaintainStatus(),
  ),
  openYTDLPCacheFolder: bindGo((a) => a.OpenYTDLPCacheFolder(), undefined),
  openYTDLPToolsFolder: bindGo((a) => a.OpenYTDLPToolsFolder(), undefined),
  getYTDLPCookieLinkageStatus: bindGo(
    (a) => asCookieApp(a).GetYTDLPCookieLinkageStatus(),
    emptyCookieLinkageStatus(),
  ),
  acknowledgeYTDLPCookieLinkageRisk: bindGo(
    (a) => asCookieApp(a).AcknowledgeYTDLPCookieLinkageRisk(),
    undefined,
  ),
  setYTDLPCookieLinkageBrowser: bindGo(
    (a, browser: string) => asCookieApp(a).SetYTDLPCookieLinkageBrowser(browser),
    undefined,
  ),
  setYTDLPCookieLinkageCookiesFile: bindGo(
    (a, path: string) => asCookieApp(a).SetYTDLPCookieLinkageCookiesFile(path),
    undefined,
  ),
  disableYTDLPCookieLinkage: bindGo(
    (a) => asCookieApp(a).DisableYTDLPCookieLinkage(),
    undefined,
  ),
  validatePath: bindGo((a, path: string) => a.ValidatePath(path), false),
  validateOutputLogPath: bindGo(
    (a, path: string) => a.ValidateOutputLogPath(path),
    false,
  ),
  openVRChatLogFolder: bindGo((a) => a.OpenVRChatLogFolder(), undefined),
  openFileDialog: (title: string, defaultDir: string, filterPattern: string) =>
    nullableStringDialog((a) =>
      a.OpenFileDialog(title, defaultDir, filterPattern),
    ),
  openDirectoryDialog: (title: string, defaultDir: string) =>
    nullableStringDialog((a) => a.OpenDirectoryDialog(title, defaultDir)),
  screenshots: bindGo(
    (a, worldId?: string) => a.Screenshots(worldId || ""),
    [],
  ),
  searchScreenshots: bindGo(
    (a, filter: ScreenshotSearchDTO) => a.SearchScreenshots(filter),
    [],
  ),
  getScreenshot: bindGo((a, id: string) => a.GetScreenshot(id), null),
  screenshotThumbnailDataURL: bindGo(
    (a, id: string) => a.ScreenshotThumbnailDataURL(id),
    "",
  ),
  openScreenshotExternally: bindGo(
    (a, id: string) => a.OpenScreenshotExternally(id),
    undefined,
  ),
  revealScreenshotInFileManager: bindGo(
    (a, id: string) => a.RevealScreenshotInFileManager(id),
    undefined,
  ),
  scanScreenshotDir: bindGo((a, path: string) => a.ScanScreenshotDir(path), 0),
  isGalleryScanning: bindGo((a) => a.IsGalleryScanning(), false),
  reindexScreenshotDir: bindGo(
    (a, path: string) => a.ReindexScreenshotDir(path),
    0,
  ),
  friends: bindGo((a) => a.Friends(), []),
  resolveUserProfileNavigation: (vrcUserID: string) =>
    callApp<UserProfileNavigationDTO>(
      (a) =>
        a.ResolveUserProfileNavigation(
          vrcUserID,
        ) as Promise<UserProfileNavigationDTO>,
      emptyUserProfileNavigation(vrcUserID),
    ),
  getSelfProfile: bindGo(
    (a, forceRefresh?: boolean) => a.GetSelfProfile(forceRefresh ?? false),
    EMPTY_USER_CACHE,
  ),
  setFavorite: bindGo(
    (a, vrcUserId: string, favorite: boolean) =>
      a.SetFavorite(vrcUserId, favorite),
    undefined,
  ),
  setStatus: bindGo((a, status: string) => a.SetStatus(status), undefined),
  setStatusDescription: bindGo(
    (a, description: string) => a.SetStatusDescription(description),
    undefined,
  ),
  setStatusAndDescription: bindGo(
    (a, status: string, description: string) =>
      a.SetStatusAndDescription(status, description),
    undefined,
  ),
  login: bindGo(
    (a, username: string, password: string, twoFactorCode?: string) =>
      a.Login(username, password, twoFactorCode ?? ""),
    LOGIN_UNAVAILABLE,
  ),
  logout: bindGo((a) => a.Logout(), undefined),
  isLoggedIn: bindGo((a) => a.IsLoggedIn(), false),
  hasStoredCredential: bindGo((a) => a.HasStoredCredential(), false),
  getCredentialBlob: bindGo((a) => a.GetCredentialBlob(), ""),
  unlockVRChatSession: bindGo(
    (a, token: string) => a.UnlockVRChatSession(token),
    undefined,
  ),
  persistWrappedCredential: bindGo(
    (a, blob: string) => a.PersistWrappedCredential(blob),
    undefined,
  ),
  clearStoredCredential: bindGo((a) => a.ClearStoredCredential(), undefined),
  getVRChatCurrentUser: bindGo(
    (a, forceRefresh?: boolean) =>
      a.GetVRChatCurrentUser(forceRefresh ?? false),
    EMPTY_VRCHAT_CURRENT_USER,
  ),
  refreshFriends: bindGo((a) => a.RefreshFriends(), undefined),
  reconcileVRChatSocialCache: bindGo(
    (a) => a.ReconcileVRChatSocialCache(),
    undefined,
  ),
  vacuumDb: bindGo((a) => a.VacuumDb(), undefined),
  encounters: bindGo((a) => a.Encounters(), []),
  encountersByVRCUserID: bindGo(
    (a, vrcUserID: string) => a.EncountersByVRCUserID(vrcUserID),
    [],
  ),
  encountersByWorldID: bindGo(
    (a, worldID: string) => a.EncountersByWorldID(worldID),
    [],
  ),
  clearEncounters: bindGo((a) => a.ClearEncounters(), 0),
  getActivityStats: bindGo(
    (a, fromISO: string, toISO: string) =>
      a.GetActivityStats(fromISO, toISO) as Promise<ActivityStatsDTO>,
    EMPTY_ACTIVITY_STATS,
  ),
  clearScreenshots: bindGo((a) => a.ClearScreenshots(), 0),
  clearFriendsCache: bindGo((a) => a.ClearFriendsCache(), 0),
  listAutomationRules: bindGo((a) => a.ListAutomationRules(), []),
  saveAutomationRule: bindGo(
    (a, rule: AutomationRuleDTO) => a.SaveAutomationRule(rule),
    undefined,
  ),
  deleteAutomationRule: bindGo(
    (a, id: string) => a.DeleteAutomationRule(id),
    undefined,
  ),
  toggleAutomationRule: bindGo(
    (a, id: string, enabled: boolean) => a.ToggleAutomationRule(id, enabled),
    undefined,
  ),
  vrchatConfigExists: bindGo((a) => a.VRChatConfigExists(), false),
  /**
   * Reads VRChat `config.json` via the backend. Rejects if the Go method errors
   * (e.g. read/parse failure) when Wails is present. The empty DTO below is only
   * the `callApp` fallback when `App` bindings are unavailable — not a substitute
   * for successful resolution on error paths.
   */
  getVRChatConfig: bindGo((a) => a.GetVRChatConfig(), EMPTY_VRCHAT_CONFIG),
  saveVRChatConfig: bindGo(
    (a, dto: VRChatConfigDTO) => a.SaveVRChatConfig(dto),
    undefined,
  ),
  deleteVRChatConfig: bindGo((a) => a.DeleteVRChatConfig(), undefined),
  defaultVRChatPictureFolder: bindGo((a) => a.DefaultVRChatPictureFolder(), ""),
};
