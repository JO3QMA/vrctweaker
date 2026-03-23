/**
 * E2E テスト専用: window.go.main.App のモック
 *
 * 【import 条件】
 * 本ファイルは frontend/e2e/app.spec.ts からのみ import し、Playwright の addInitScript で
 * 実行時にブラウザへ注入されます。src/ 以下のアプリケーションコードからは一切参照しません。
 *
 * 本番ビルドに混入しない理由:
 * - frontend/e2e は Vite の build 対象外（index.html と src/main.ts から到達しない）
 * - mock-wails.ts は app.spec.ts 経由でのみ参照され、app.spec.ts は Playwright テストランナー専用
 */

/** ページ読み込み前に注入する window.go スタブの初期化スクリプトを返す */
export function getMockWailsInitScript(): string {
  const seedProfiles = [
    {
      id: "profile-1",
      name: "デフォルトプロファイル",
      arguments: "",
      isDefault: true,
      createdAt: "2025-01-01T00:00:00Z",
      updatedAt: "2025-01-01T00:00:00Z",
    },
  ];

  const seedPathSettings = {
    vrchatPathWindows: "",
    steamPathLinux: "",
    outputLogPath: "",
  };

  // JSON として埋め込み、ブラウザ側でパースして使用
  const profilesJson = JSON.stringify(seedProfiles);
  const pathSettingsJson = JSON.stringify(seedPathSettings);

  return `
    (function() {
      if (typeof window === 'undefined') return;
      const profiles = ${profilesJson};
      const pathSettings = ${pathSettingsJson};
      window.go = window.go || {};
      window.go.main = window.go.main || {};
      window.go.main.App = {
        Greet: () => Promise.resolve('Hello, Welcome!'),
        LaunchProfiles: () => Promise.resolve(profiles),
        SaveLaunchProfile: () => Promise.resolve(),
        DeleteLaunchProfile: () => Promise.resolve(),
        LaunchVRChat: () => Promise.resolve(),
        JoinWorld: () => Promise.resolve(),
        JoinWorldFromScreenshot: () => Promise.resolve(),
        GetLogRetentionDays: () => Promise.resolve(30),
        SetLogRetentionDays: () => Promise.resolve(),
        GetPathSettings: () => Promise.resolve(pathSettings),
        SetPathSettings: () => Promise.resolve(),
        ValidatePath: () => Promise.resolve(true),
        ValidateOutputLogPath: () => Promise.resolve(true),
        OpenVRChatLogFolder: () => Promise.resolve(),
        OpenFileDialog: () => Promise.resolve(''),
        OpenDirectoryDialog: () => Promise.resolve(''),
        Screenshots: () => Promise.resolve([]),
        SearchScreenshots: () => Promise.resolve([]),
        GetScreenshot: () => Promise.resolve(null),
        ScreenshotThumbnailDataURL: () =>
          Promise.resolve(
            'data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7',
          ),
        OpenScreenshotExternally: () => Promise.resolve(),
        RevealScreenshotInFileManager: () => Promise.resolve(),
        ScanScreenshotDir: () => Promise.resolve(0),
        IsGalleryScanning: () => Promise.resolve(false),
        ReindexScreenshotDir: () => Promise.resolve(0),
        Encounters: () => Promise.resolve([]),
        EncountersByVRCUserID: () => Promise.resolve([]),
        EncountersByWorldID: () => Promise.resolve([]),
        RotateEncounters: () => Promise.resolve(0),
        GetActivityStats: () => Promise.resolve({ dailyPlaySeconds: [], topWorlds: [] }),
        Friends: () => Promise.resolve([]),
        SetFavorite: () => Promise.resolve(),
        SetStatus: () => Promise.resolve(),
        Login: () => Promise.resolve({ ok: false, error: 'E2E mock' }),
        Logout: () => Promise.resolve(),
        IsLoggedIn: () => Promise.resolve(false),
        GetVRChatCurrentUser: () =>
          Promise.reject(new Error('E2E mock: not logged in')),
        RefreshFriends: () => Promise.resolve(),
        VacuumDb: () => Promise.resolve(),
        ClearEncounters: () => Promise.resolve(0),
        ClearScreenshots: () => Promise.resolve(0),
        ClearFriendsCache: () => Promise.resolve(0),
        ListAutomationRules: () => Promise.resolve([]),
        SaveAutomationRule: () => Promise.resolve(),
        DeleteAutomationRule: () => Promise.resolve(),
        ToggleAutomationRule: () => Promise.resolve(),
        VRChatConfigExists: () => Promise.resolve(false),
        GetVRChatConfig: () => Promise.resolve({
          cameraResWidth: 1920,
          cameraResHeight: 1080,
          screenshotResWidth: 1920,
          screenshotResHeight: 1080,
          pictureOutputFolder: '',
          pictureOutputSplitByDate: null,
          fpvSteadycamFov: 0,
          cacheDirectory: '',
          cacheSize: 0,
          cacheExpiryDelay: 0,
          disableRichPresence: null,
        }),
        SaveVRChatConfig: () => Promise.resolve(),
        DeleteVRChatConfig: () => Promise.resolve(),
        DefaultVRChatPictureFolder: () =>
          Promise.resolve('C:\\\\Temp\\\\VRChatTweakerE2E\\\\Pictures\\\\VRChat'),
      };
    })();
  `.trim();
}
