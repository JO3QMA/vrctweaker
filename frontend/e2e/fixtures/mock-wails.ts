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
  E2E_SELF_USER_ID,
  E2E_TEST_USER_ID,
  SEED_ACTIVITY_STATS,
  SEED_AUTOMATION_RULES,
  SEED_ENCOUNTERS,
  SEED_FRIENDS,
  SEED_LAUNCH_PROFILES,
  SEED_PATH_SETTINGS,
  SEED_SCREENSHOTS,
  SEED_SELF_PROFILE,
  SEED_VRCHAT_CONFIG,
  seedUserProfileNavigation,
} from "./seed-data";

export interface MockWailsOptions {
  /** true のとき IsLoggedIn / GetSelfProfile がログイン済みを返す */
  loggedIn?: boolean;
}

/** ページ読み込み前に注入する window.go スタブの初期化スクリプトを返す（中身はブラウザでそのまま実行されるため TypeScript 構文不可） */
export function getMockWailsInitScript(options: MockWailsOptions = {}): string {
  const loggedIn = options.loggedIn ?? false;
  const profilesJson = JSON.stringify(SEED_LAUNCH_PROFILES);
  const pathSettingsJson = JSON.stringify(SEED_PATH_SETTINGS);
  const screenshotsJson = JSON.stringify(SEED_SCREENSHOTS);
  const encountersJson = JSON.stringify(SEED_ENCOUNTERS);
  const friendsJson = JSON.stringify(SEED_FRIENDS);
  const activityStatsJson = JSON.stringify(SEED_ACTIVITY_STATS);
  const automationRulesJson = JSON.stringify(SEED_AUTOMATION_RULES);
  const vrchatConfigJson = JSON.stringify(SEED_VRCHAT_CONFIG);
  const e2eTestUserIdJson = JSON.stringify(E2E_TEST_USER_ID);
  const e2eSelfUserIdJson = JSON.stringify(E2E_SELF_USER_ID);
  const selfProfileJson = JSON.stringify(SEED_SELF_PROFILE);
  const resolveUserProfileSeedJson = JSON.stringify(
    seedUserProfileNavigation(E2E_TEST_USER_ID),
  );
  const resolveSelfProfileSeedJson = JSON.stringify(
    seedUserProfileNavigation(E2E_SELF_USER_ID),
  );
  const loggedInJson = JSON.stringify(loggedIn);

  return `
    (function() {
      if (typeof window === 'undefined') return;
      const profiles = ${profilesJson};
      let launchProfiles = profiles.slice();
      const pathSettings = ${pathSettingsJson};
      const screenshots = ${screenshotsJson};
      const encounters = ${encountersJson};
      const friends = ${friendsJson};
      const activityStats = ${activityStatsJson};
      const automationRules = ${automationRulesJson};
      const vrchatConfig = ${vrchatConfigJson};
      const e2eTestUserId = ${e2eTestUserIdJson};
      const e2eSelfUserId = ${e2eSelfUserIdJson};
      const selfProfileSeed = ${selfProfileJson};
      const resolveUserProfileSeed = ${resolveUserProfileSeedJson};
      const resolveSelfProfileSeed = ${resolveSelfProfileSeedJson};
      const loggedIn = ${loggedInJson};
      var selfProfile = Object.assign({}, selfProfileSeed);
      var selfProfileRefreshCount = 0;

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
        if (vrcUserId === e2eSelfUserId) {
          return {
            user: resolveSelfProfileSeed.user,
            openInFriendsView: resolveSelfProfileSeed.openInFriendsView,
            openInSelfProfile: resolveSelfProfileSeed.openInSelfProfile,
          };
        }
        if (vrcUserId === e2eTestUserId) {
          return {
            user: resolveUserProfileSeed.user,
            openInFriendsView: resolveUserProfileSeed.openInFriendsView,
            openInSelfProfile: resolveUserProfileSeed.openInSelfProfile,
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
          openInSelfProfile: false,
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
        LaunchProfiles: () => Promise.resolve(launchProfiles),
        SaveLaunchProfile: (p) => {
          const idx = launchProfiles.findIndex(function(x) { return x.id === p.id; });
          if (idx >= 0) launchProfiles[idx] = p;
          else launchProfiles.push(p);
          return Promise.resolve();
        },
        DeleteLaunchProfile: (id) => {
          launchProfiles = launchProfiles.filter(function(x) { return x.id !== id; });
          return Promise.resolve();
        },
        LaunchVRChat: () => Promise.resolve(),
        LaunchVRChatWithArgs: (_args, _profileId) => Promise.resolve(),
        GetInstanceRejoinSection: () => Promise.resolve(null),
        InstanceRejoin: (_profileId, _playSessionId) => Promise.resolve(),
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
        RuntimeIsWindows: () => Promise.resolve(true),
        GetYTDLPMaintainStatus: () =>
          Promise.resolve({
            supported: true,
            unsupportedReason: '',
            maintainDesired: false,
            riskAcknowledged: false,
            effectiveOfficial: false,
            cachePresent: false,
            cacheVersion: '',
            toolsPath: '',
            cachePath: '',
            pendingError: '',
            latestVersion: '',
            latestTag: '',
            latestDownloadUrl: '',
            latestError: '',
          }),
        AcknowledgeYTDLPToolsReplaceRisk: () => Promise.resolve(),
        SetYTDLPToolsReplaceMaintain: () => Promise.resolve(),
        CheckYTDLPLatestRelease: () =>
          Promise.resolve({
            supported: true,
            unsupportedReason: '',
            maintainDesired: false,
            riskAcknowledged: false,
            effectiveOfficial: false,
            cachePresent: false,
            cacheVersion: '',
            toolsPath: '',
            cachePath: '',
            pendingError: '',
            latestVersion: '2025.01.01',
            latestTag: '2025.01.01',
            latestDownloadUrl: 'https://example.com/yt-dlp.exe',
            latestError: '',
          }),
        UpdateOfficialYTDLPCache: () =>
          Promise.resolve({
            supported: true,
            unsupportedReason: '',
            maintainDesired: false,
            riskAcknowledged: false,
            effectiveOfficial: false,
            cachePresent: true,
            cacheVersion: '2025.01.01',
            toolsPath: '',
            cachePath: '',
            pendingError: '',
            latestVersion: '',
            latestTag: '',
            latestDownloadUrl: '',
            latestError: '',
          }),
        OpenYTDLPCacheFolder: () => Promise.resolve(),
        OpenYTDLPToolsFolder: () => Promise.resolve(),
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
        IsLoggedIn: () => Promise.resolve(loggedIn),
        HasStoredCredential: () => Promise.resolve(loggedIn),
        GetCredentialBlob: () => Promise.resolve(''),
        UnlockVRChatSession: (_token) => Promise.resolve(),
        PersistWrappedCredential: (_blob) => Promise.resolve(),
        ClearStoredCredential: () => Promise.resolve(),
        GetVRChatCurrentUser: (_forceRefresh) =>
          loggedIn
            ? Promise.resolve({
                id: selfProfile.vrcUserId,
                displayName: selfProfile.displayName,
                username: selfProfile.username || '',
                status: selfProfile.status || '',
                statusDescription: selfProfile.statusDescription || '',
                state: selfProfile.state || '',
                currentAvatarThumbnailImageUrl: '',
                userIcon: '',
                profilePicOverrideThumbnail: '',
              })
            : Promise.reject(new Error('E2E mock: not logged in')),
        GetSelfProfile: (forceRefresh) => {
          if (!loggedIn) {
            return Promise.reject(new Error('E2E mock: not logged in'));
          }
          if (forceRefresh) {
            selfProfileRefreshCount += 1;
            selfProfile = Object.assign({}, selfProfileSeed, {
              statusDescription:
                'E2E refreshed ' + String(selfProfileRefreshCount),
            });
          }
          return Promise.resolve(selfProfile);
        },
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
