import type { Decorator } from "@storybook/vue3-vite";
import type { PathSettingsDTO, UserCacheDTO } from "../../wails/app";
import { resetSessionUnlockForStorybook } from "../../composables/useSessionUnlock";
import { withWailsApp } from "../wailsDecorator";

export type SettingsViewWailsPreset =
  "needsReloginClean" | "needsReloginUnlockError" | "loggedIn";

const emptyPaths: PathSettingsDTO = {
  vrchatPathWindows: "",
  steamPathLinux: "",
  outputLogPath: "",
};

const sampleSelfProfile: UserCacheDTO = {
  vrcUserId: "usr_storybook_vrc",
  displayName: "ストーリー花子",
  username: "story_hanako",
  status: "join me",
  statusDescription: "Storybook 用の表示です",
  state: "active",
  isFavorite: false,
  lastUpdated: "2025-01-01T00:00:00Z",
  currentAvatarThumbnailImageUrl: "",
  userIcon: "",
  profilePicOverrideThumbnail: "",
};

const emptySelfProfile: UserCacheDTO = {
  vrcUserId: "",
  displayName: "",
  status: "",
  isFavorite: false,
  lastUpdated: "",
};

export function withSettingsWails(preset: SettingsViewWailsPreset): Decorator {
  const hasBlob = preset !== "needsReloginClean";
  const blob =
    preset === "needsReloginClean" ? "" : "storybook-legacy-plain-token";
  const loggedIn = preset === "loggedIn";

  const unlockVRChatSession =
    preset === "needsReloginUnlockError"
      ? (_token: string) =>
          Promise.reject(
            new Error(
              "VRCTWK_UNLOCK_NEEDS_RELOGIN: session expired: GET /auth/user",
            ),
          )
      : (_token: string) => Promise.resolve();

  return withWailsApp(
    {
      HasStoredCredential: () => Promise.resolve(hasBlob),
      GetCredentialBlob: () => Promise.resolve(blob),
      UnlockVRChatSession: unlockVRChatSession,
      PersistWrappedCredential: () => Promise.resolve(),
      ClearStoredCredential: () => Promise.resolve(),
      IsLoggedIn: () => Promise.resolve(loggedIn),
      GetSelfProfile: (_force?: boolean) =>
        Promise.resolve(loggedIn ? sampleSelfProfile : emptySelfProfile),
      GetLogRetentionDays: () => Promise.resolve(30),
      SetLogRetentionDays: (_days: number) => Promise.resolve(),
      GetPathSettings: () => Promise.resolve({ ...emptyPaths }),
      SetPathSettings: (_dto: PathSettingsDTO) => Promise.resolve(),
      GetSuppressSleepWhileVRChat: () => Promise.resolve(false),
      SetSuppressSleepWhileVRChat: (_on: boolean) => Promise.resolve(),
      ValidatePath: (_path: string) => Promise.resolve(true),
      ValidateOutputLogPath: (_path: string) => Promise.resolve(true),
      OpenVRChatLogFolder: () => Promise.resolve(),
      OpenFileDialog: () => Promise.resolve(""),
      OpenDirectoryDialog: () => Promise.resolve(""),
      GetYTDLPCookieLinkageStatus: () =>
        Promise.resolve({
          supported: true,
          enabled: false,
          sourceKind: "",
          riskAcknowledged: false,
          browser: "chrome",
        }),
      GetYTDLPMaintainStatus: () =>
        Promise.resolve({
          supported: true,
          maintainDesired: false,
          riskAcknowledged: true,
          effectiveOfficial: true,
          cachePresent: true,
          cacheVersion: "",
          toolsPath: "",
          cachePath: "",
          pendingError: "",
          latestVersion: "",
          latestTag: "",
          latestDownloadUrl: "",
          latestError: "",
        }),
      AcknowledgeYTDLPCookieLinkageRisk: () => Promise.resolve(),
      SetYTDLPCookieLinkageBrowser: () => Promise.resolve(),
      SetYTDLPCookieLinkageCookiesFile: () => Promise.resolve(),
      DisableYTDLPCookieLinkage: () => Promise.resolve(),
      Login: () =>
        Promise.resolve({
          ok: false,
          error: "Storybook ではログインできません",
        }),
      Logout: () => Promise.resolve(),
      RefreshFriends: () => Promise.resolve(),
      VacuumDb: () => Promise.resolve(),
      ClearEncounters: () => Promise.resolve(0),
      ClearScreenshots: () => Promise.resolve(0),
      ClearFriendsCache: () => Promise.resolve(0),
    },
    {
      created: () => resetSessionUnlockForStorybook(),
      beforeUnmount: () => resetSessionUnlockForStorybook(),
    },
  );
}
