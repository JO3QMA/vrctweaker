import type { PathSettingsDTO, VRChatCurrentUserDTO } from "../wails/app";
import { resetSessionUnlockForStorybook } from "../composables/useSessionUnlock";

export type SettingsViewWailsPreset =
  /** No stored blob: needs-relogin, no unlock warning */
  | "needsReloginClean"
  /** Legacy blob + UnlockVRChatSession auth error: warning el-alert */
  | "needsReloginUnlockError"
  /** Blob unlock succeeds + IsLoggedIn: profile card */
  | "loggedIn";

const emptyPaths: PathSettingsDTO = {
  vrchatPathWindows: "",
  steamPathLinux: "",
  outputLogPath: "",
};

const sampleCurrentUser: VRChatCurrentUserDTO = {
  id: "usr_storybook_vrc",
  displayName: "ストーリー花子",
  username: "story_hanako",
  status: "join me",
  statusDescription: "Storybook 用の表示です",
  state: "active",
  currentAvatarThumbnailImageUrl: "",
  userIcon: "",
  profilePicOverrideThumbnail: "",
};

const emptyCurrentUser: VRChatCurrentUserDTO = {
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

/**
 * SettingsView / useSessionUnlock が onMounted で呼ぶ Wails を Storybook 用に差し替える。
 * 資格情報まわり（HasStoredCredential, UnlockVRChatSession 等）をプリセットで切り替える。
 */
export function settingsViewWailsDecorator(preset: SettingsViewWailsPreset) {
  return (story: () => unknown) => {
    let prevGo: typeof window.go;

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

    return {
      components: { story },
      template: "<story />",
      created() {
        resetSessionUnlockForStorybook();
        prevGo = window.go;
        window.go = {
          main: {
            App: {
              HasStoredCredential: () => Promise.resolve(hasBlob),
              GetCredentialBlob: () => Promise.resolve(blob),
              UnlockVRChatSession: unlockVRChatSession,
              PersistWrappedCredential: () => Promise.resolve(),
              ClearStoredCredential: () => Promise.resolve(),
              IsLoggedIn: () => Promise.resolve(loggedIn),
              GetVRChatCurrentUser: (_force?: boolean) =>
                Promise.resolve(
                  loggedIn ? sampleCurrentUser : emptyCurrentUser,
                ),
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
            } as unknown as NonNullable<
              NonNullable<typeof window.go>["main"]
            >["App"],
          },
        };
      },
      beforeUnmount() {
        resetSessionUnlockForStorybook();
        window.go = prevGo;
      },
    };
  };
}
