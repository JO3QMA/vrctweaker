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
export type LaunchArgsParsedDTO = WailsDTO<launcher.LaunchArgsParsed>;
export type ScreenshotDTO = WailsDTO<main.ScreenshotDTO>;
export type ScreenshotSearchDTO = WailsDTO<main.ScreenshotSearchDTO>;
export type UserEncounterDTO = WailsDTO<main.UserEncounterDTO>;
export type UserCacheDTO = WailsDTO<main.UserCacheDTO>;
export type UserProfileNavigationDTO = WailsDTO<main.UserProfileNavigationDTO>;
export type PathSettingsDTO = WailsDTO<usecase.PathSettings>;
export type LoginResultDTO = WailsDTO<main.LoginResultDTO>;
export type VRChatCurrentUserDTO = WailsDTO<main.VRChatCurrentUserDTO>;
export type DailyPlaySecondsDTO = WailsDTO<activity.DailyPlaySeconds>;
export type TopWorldDTO = WailsDTO<activity.TopWorldSummary>;
export type ActivityStatsDTO = WailsDTO<activity.ActivityStats>;
export type AutomationRuleDTO = WailsDTO<automation.AutomationRule>;
export type VRChatConfigDTO = WailsDTO<main.VRChatConfigDTO>;

/** -999 = omit for process/main thread priority */
export const PRIORITY_OMIT = -999;

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
  async getLanguage(): Promise<string> {
    return callApp((a) => a.GetLanguage(), "");
  },
  async setLanguage(lang: string): Promise<void> {
    return callApp((a) => a.SetLanguage(lang), undefined);
  },
  async getSystemLocale(): Promise<string> {
    return callApp((a) => a.GetSystemLocale(), "en");
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
  async getSuppressSleepWhileVRChat(): Promise<boolean> {
    return callApp((a) => a.GetSuppressSleepWhileVRChat(), false);
  },
  async setSuppressSleepWhileVRChat(on: boolean): Promise<void> {
    return callApp((a) => a.SetSuppressSleepWhileVRChat(on), undefined);
  },
  async validatePath(path: string): Promise<boolean> {
    return callApp((a) => a.ValidatePath(path), false);
  },
  async validateOutputLogPath(path: string): Promise<boolean> {
    return callApp((a) => a.ValidateOutputLogPath(path), false);
  },
  async openVRChatLogFolder(): Promise<void> {
    return callApp((a) => a.OpenVRChatLogFolder(), undefined);
  },
  async openFileDialog(
    title: string,
    defaultDir: string,
    filterPattern: string,
  ): Promise<string | null> {
    const result = await callApp(
      (a) => a.OpenFileDialog(title, defaultDir, filterPattern),
      "",
    );
    return result && result !== "" ? result : null;
  },
  async openDirectoryDialog(
    title: string,
    defaultDir: string,
  ): Promise<string | null> {
    const result = await callApp(
      (a) => a.OpenDirectoryDialog(title, defaultDir),
      "",
    );
    return result && result !== "" ? result : null;
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
  async screenshotThumbnailDataURL(id: string): Promise<string> {
    return callApp((a) => a.ScreenshotThumbnailDataURL(id), "");
  },
  async openScreenshotExternally(id: string): Promise<void> {
    return callApp((a) => a.OpenScreenshotExternally(id), undefined);
  },
  async revealScreenshotInFileManager(id: string): Promise<void> {
    return callApp((a) => a.RevealScreenshotInFileManager(id), undefined);
  },
  async scanScreenshotDir(path: string): Promise<number> {
    return callApp((a) => a.ScanScreenshotDir(path), 0);
  },
  async isGalleryScanning(): Promise<boolean> {
    return callApp((a) => a.IsGalleryScanning(), false);
  },
  async reindexScreenshotDir(path: string): Promise<number> {
    return callApp((a) => a.ReindexScreenshotDir(path), 0);
  },
  async friends(): Promise<UserCacheDTO[]> {
    return callApp((a) => a.Friends(), []);
  },
  async resolveUserProfileNavigation(
    vrcUserID: string,
  ): Promise<UserProfileNavigationDTO> {
    return callApp<UserProfileNavigationDTO>(
      (a) =>
        a.ResolveUserProfileNavigation(
          vrcUserID,
        ) as Promise<UserProfileNavigationDTO>,
      {
        user: {
          vrcUserId: vrcUserID,
          displayName: "",
          status: "",
          isFavorite: false,
          lastUpdated: "",
        },
        openInFriendsView: false,
        openInSelfProfile: false,
      },
    );
  },
  async getSelfProfile(forceRefresh?: boolean): Promise<UserCacheDTO> {
    return callApp((a) => a.GetSelfProfile(forceRefresh ?? false), {
      vrcUserId: "",
      displayName: "",
      status: "",
      isFavorite: false,
      lastUpdated: "",
    });
  },
  async setFavorite(vrcUserId: string, favorite: boolean): Promise<void> {
    return callApp((a) => a.SetFavorite(vrcUserId, favorite), undefined);
  },
  async setStatus(status: string): Promise<void> {
    return callApp((a) => a.SetStatus(status), undefined);
  },
  async setStatusDescription(description: string): Promise<void> {
    return callApp((a) => a.SetStatusDescription(description), undefined);
  },
  async setStatusAndDescription(
    status: string,
    description: string,
  ): Promise<void> {
    return callApp(
      (a) => a.SetStatusAndDescription(status, description),
      undefined,
    );
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
  async hasStoredCredential(): Promise<boolean> {
    return callApp((a) => a.HasStoredCredential(), false);
  },
  async getCredentialBlob(): Promise<string> {
    return callApp((a) => a.GetCredentialBlob(), "");
  },
  async unlockVRChatSession(token: string): Promise<void> {
    return callApp((a) => a.UnlockVRChatSession(token), undefined);
  },
  async persistWrappedCredential(blob: string): Promise<void> {
    return callApp((a) => a.PersistWrappedCredential(blob), undefined);
  },
  async clearStoredCredential(): Promise<void> {
    return callApp((a) => a.ClearStoredCredential(), undefined);
  },
  async getVRChatCurrentUser(
    forceRefresh?: boolean,
  ): Promise<VRChatCurrentUserDTO> {
    return callApp((a) => a.GetVRChatCurrentUser(forceRefresh ?? false), {
      id: "",
      displayName: "",
      username: "",
      status: "",
      statusDescription: "",
      state: "",
      currentAvatarThumbnailImageUrl: "",
      userIcon: "",
      profilePicOverrideThumbnail: "",
    });
  },
  async refreshFriends(): Promise<void> {
    return callApp((a) => a.RefreshFriends(), undefined);
  },
  async reconcileVRChatSocialCache(): Promise<void> {
    return callApp((a) => a.ReconcileVRChatSocialCache(), undefined);
  },
  async vacuumDb(): Promise<void> {
    return callApp((a) => a.VacuumDb(), undefined);
  },
  async encounters(): Promise<UserEncounterDTO[]> {
    return callApp((a) => a.Encounters(), []);
  },
  async encountersByVRCUserID(vrcUserID: string): Promise<UserEncounterDTO[]> {
    return callApp((a) => a.EncountersByVRCUserID(vrcUserID), []);
  },
  async encountersByWorldID(worldID: string): Promise<UserEncounterDTO[]> {
    return callApp((a) => a.EncountersByWorldID(worldID), []);
  },
  async clearEncounters(): Promise<number> {
    return callApp((a) => a.ClearEncounters(), 0);
  },
  async getActivityStats(
    fromISO: string,
    toISO: string,
  ): Promise<ActivityStatsDTO> {
    return callApp<ActivityStatsDTO>(
      (a) => a.GetActivityStats(fromISO, toISO) as Promise<ActivityStatsDTO>,
      {
        dailyPlaySeconds: [],
        topWorlds: [],
      },
    );
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
  async vrchatConfigExists(): Promise<boolean> {
    return callApp((a) => a.VRChatConfigExists(), false);
  },
  /**
   * Reads VRChat `config.json` via the backend. Rejects if the Go method errors
   * (e.g. read/parse failure) when Wails is present. The empty DTO below is only
   * the `callApp` fallback when `App` bindings are unavailable — not a substitute
   * for successful resolution on error paths.
   */
  async getVRChatConfig(): Promise<VRChatConfigDTO> {
    return callApp((a) => a.GetVRChatConfig(), {
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
    });
  },
  async saveVRChatConfig(dto: VRChatConfigDTO): Promise<void> {
    return callApp((a) => a.SaveVRChatConfig(dto), undefined);
  },
  async deleteVRChatConfig(): Promise<void> {
    return callApp((a) => a.DeleteVRChatConfig(), undefined);
  },
  async defaultVRChatPictureFolder(): Promise<string> {
    return callApp((a) => a.DefaultVRChatPictureFolder(), "");
  },
};
