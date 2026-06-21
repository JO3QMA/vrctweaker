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

import {
  E2E_TEST_USER_ID,
  SEED_ACTIVITY_STATS,
  SEED_AUTOMATION_RULES,
  SEED_ENCOUNTERS,
  SEED_FRIENDS,
  SEED_LAUNCH_PROFILES,
  SEED_PATH_SETTINGS,
  SEED_SCREENSHOTS,
  SEED_VRCHAT_CONFIG,
  seedUserProfileNavigation,
} from "./seed-data";

/** ページ読み込み前に注入する window.go スタブの初期化スクリプトを返す（中身はブラウザでそのまま実行されるため TypeScript 構文不可） */
export function getMockWailsInitScript(): string {
  const profilesJson = JSON.stringify(SEED_LAUNCH_PROFILES);
  const pathSettingsJson = JSON.stringify(SEED_PATH_SETTINGS);
  const screenshotsJson = JSON.stringify(SEED_SCREENSHOTS);
  const encountersJson = JSON.stringify(SEED_ENCOUNTERS);
  const friendsJson = JSON.stringify(SEED_FRIENDS);
  const activityStatsJson = JSON.stringify(SEED_ACTIVITY_STATS);
  const automationRulesJson = JSON.stringify(SEED_AUTOMATION_RULES);
  const vrchatConfigJson = JSON.stringify(SEED_VRCHAT_CONFIG);
  const e2eTestUserIdJson = JSON.stringify(E2E_TEST_USER_ID);
  const resolveUserProfileSeedJson = JSON.stringify(
    seedUserProfileNavigation(E2E_TEST_USER_ID),
  );

  return `
    (function() {
      if (typeof window === 'undefined') return;
      const profiles = ${profilesJson};
      const pathSettings = ${pathSettingsJson};
      const screenshots = ${screenshotsJson};
      const encounters = ${encountersJson};
      const friends = ${friendsJson};
      const activityStats = ${activityStatsJson};
      const automationRules = ${automationRulesJson};
      const vrchatConfig = ${vrchatConfigJson};
      const e2eTestUserId = ${e2eTestUserIdJson};
      const resolveUserProfileSeed = ${resolveUserProfileSeedJson};

      function filterScreenshotsByWorldId(list, worldId) {
        if (!worldId || !String(worldId).trim()) return list;
        return list.filter(function(s) { return s.worldId === worldId; });
      }

      function searchScreenshots(list, filter) {
        var result = list;
        if (!filter) return result;
        if (filter.worldId && String(filter.worldId).trim()) {
          result = result.filter(function(s) { return s.worldId === filter.worldId; });
        }
        if (filter.worldName && String(filter.worldName).trim()) {
          var q = String(filter.worldName).trim().toLowerCase();
          result = result.filter(function(s) {
            return s.worldName && s.worldName.toLowerCase().indexOf(q) !== -1;
          });
        }
        if (filter.dateFrom && String(filter.dateFrom).trim()) {
          var from = String(filter.dateFrom).trim();
          result = result.filter(function(s) {
            return !s.takenAt || s.takenAt.slice(0, 10) >= from;
          });
        }
        if (filter.dateTo && String(filter.dateTo).trim()) {
          var to = String(filter.dateTo).trim();
          result = result.filter(function(s) {
            return !s.takenAt || s.takenAt.slice(0, 10) <= to;
          });
        }
        return result;
      }

      function encountersByVrcUserId(list, vrcUserId) {
        return list.filter(function(e) { return e.vrcUserId === vrcUserId; });
      }

      function encountersByWorldId(list, worldId) {
        return list.filter(function(e) { return e.worldId === worldId; });
      }

      function resolveUserProfileNavigation(vrcUserId) {
        if (vrcUserId === e2eTestUserId) {
          return {
            user: resolveUserProfileSeed.user,
            openInFriendsView: resolveUserProfileSeed.openInFriendsView,
          };
        }
        return {
          user: {
            vrcUserId: vrcUserId,
            displayName: '',
            status: '',
            isFavorite: false,
            lastUpdated: '',
          },
          openInFriendsView: false,
        };
      }

      var wailsEventHandlers = {};
      function emitWailsEvent(eventName, data) {
        var handlers = wailsEventHandlers[eventName] || [];
        handlers.forEach(function(h) { h(data); });
      }

      window.runtime = {
        EventsOn: function(eventName, cb) {
          if (!wailsEventHandlers[eventName]) wailsEventHandlers[eventName] = [];
          wailsEventHandlers[eventName].push(cb);
          return function() {
            wailsEventHandlers[eventName] = (wailsEventHandlers[eventName] || []).filter(function(h) {
              return h !== cb;
            });
          };
        },
      };

      window.go = window.go || {};
      window.go.main = window.go.main || {};
      window.go.main.App = {
        Greet: () => Promise.resolve('Hello, Welcome!'),
        LaunchProfiles: () => Promise.resolve(profiles),
        SaveLaunchProfile: () => Promise.resolve(),
        DeleteLaunchProfile: () => Promise.resolve(),
        LaunchVRChat: () => Promise.resolve(),
        LaunchVRChatWithArgs: () => Promise.resolve(),
        ParseLaunchArgsForGUI: () =>
          Promise.resolve({
            noVr: false,
            screenMode: '',
            screenWidth: 0,
            screenHeight: 0,
            fps: 90,
            skipRegistry: false,
            processPriority: -999,
            mainThreadPriority: -999,
            monitor: 0,
            profile: -1,
            enableDebugGui: false,
            enableSDKLogLevels: false,
            enableUdonDebugLogging: false,
            midi: '',
            watchWorlds: false,
            watchAvatars: false,
            ignoreTrackers: '',
            videoDecoding: '',
            disableAMDStutterWorkaround: false,
            osc: '',
            affinity: '',
            enforceWorldServerChecks: false,
            custom: '',
          }),
        MergeLaunchArgsForGUI: () => Promise.resolve(''),
        JoinWorld: () => Promise.resolve(),
        JoinWorldFromScreenshot: () => Promise.resolve(),
        GetLogRetentionDays: () => Promise.resolve(30),
        SetLogRetentionDays: () => Promise.resolve(),
        GetLanguage: () => Promise.resolve('ja'),
        SetLanguage: () => Promise.resolve(),
        GetSystemLocale: () => Promise.resolve('ja'),
        GetPathSettings: () => Promise.resolve(pathSettings),
        SetPathSettings: () => Promise.resolve(),
        GetSuppressSleepWhileVRChat: () => Promise.resolve(false),
        SetSuppressSleepWhileVRChat: () => Promise.resolve(),
        ValidatePath: () => Promise.resolve(true),
        ValidateOutputLogPath: () => Promise.resolve(true),
        OpenVRChatLogFolder: () => Promise.resolve(),
        OpenFileDialog: () => Promise.resolve(''),
        OpenDirectoryDialog: () => Promise.resolve(''),
        Screenshots: (worldId) =>
          Promise.resolve(filterScreenshotsByWorldId(screenshots, worldId)),
        SearchScreenshots: (filter) =>
          Promise.resolve(searchScreenshots(screenshots, filter)),
        GetScreenshot: (id) =>
          Promise.resolve(
            screenshots.find(function(s) { return s.id === id; }) || null,
          ),
        ScreenshotThumbnailDataURL: () =>
          Promise.resolve(
            'data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7',
          ),
        OpenScreenshotExternally: () => Promise.resolve(),
        RevealScreenshotInFileManager: () => Promise.resolve(),
        ScanScreenshotDir: function() {
          return Promise.resolve(0).then(function(count) {
            setTimeout(function() {
              emitWailsEvent('gallery:scan-done', { count: count });
            }, 0);
            return count;
          });
        },
        IsGalleryScanning: () => Promise.resolve(false),
        ReindexScreenshotDir: () => Promise.resolve(0),
        Encounters: () => Promise.resolve(encounters),
        EncountersByVRCUserID: (vrcUserId) =>
          Promise.resolve(encountersByVrcUserId(encounters, vrcUserId)),
        EncountersByWorldID: (worldId) =>
          Promise.resolve(encountersByWorldId(encounters, worldId)),
        RotateEncounters: () => Promise.resolve(0),
        GetActivityStats: (_fromISO, _toISO) => Promise.resolve(activityStats),
        Friends: () => Promise.resolve(friends),
        ResolveUserProfileNavigation: (id) =>
          Promise.resolve(resolveUserProfileNavigation(id)),
        SetFavorite: () => Promise.resolve(),
        SetStatus: () => Promise.resolve(),
        SetStatusDescription: () => Promise.resolve(),
        SetStatusAndDescription: () => Promise.resolve(),
        Login: () => Promise.resolve({ ok: false, error: 'E2E mock' }),
        Logout: () => Promise.resolve(),
        IsLoggedIn: () => Promise.resolve(false),
        HasStoredCredential: () => Promise.resolve(false),
        GetCredentialBlob: () => Promise.resolve(''),
        UnlockVRChatSession: (_token) => Promise.resolve(),
        PersistWrappedCredential: (_blob) => Promise.resolve(),
        ClearStoredCredential: () => Promise.resolve(),
        GetVRChatCurrentUser: (_forceRefresh) =>
          Promise.reject(new Error('E2E mock: not logged in')),
        RefreshFriends: () => Promise.resolve(),
        ReconcileVRChatSocialCache: () => Promise.resolve(),
        VacuumDb: () => Promise.resolve(),
        ClearEncounters: () => Promise.resolve(0),
        ClearScreenshots: () => Promise.resolve(0),
        ClearFriendsCache: () => Promise.resolve(0),
        ListAutomationRules: () => Promise.resolve(automationRules),
        SaveAutomationRule: () => Promise.resolve(),
        DeleteAutomationRule: () => Promise.resolve(),
        ToggleAutomationRule: () => Promise.resolve(),
        VRChatConfigExists: () => Promise.resolve(true),
        GetVRChatConfig: () => Promise.resolve(vrchatConfig),
        SaveVRChatConfig: () => Promise.resolve(),
        DeleteVRChatConfig: () => Promise.resolve(),
        DefaultVRChatPictureFolder: () =>
          Promise.resolve('C:\\\\Temp\\\\VRChatTweakerE2E\\\\Pictures\\\\VRChat'),
      };
    })();
  `.trim();
}
