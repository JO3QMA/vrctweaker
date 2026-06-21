import {
  afterEach,
  beforeEach,
  describe,
  expect,
  it,
  vi,
  type Mock,
} from "vitest";
import {
  App,
  callApp,
  isWailsRuntime,
  PRIORITY_OMIT,
  type ActivityStatsDTO,
  type AppBindings,
  type AutomationRuleDTO,
  type LaunchArgsParsedDTO,
  type LaunchProfileDTO,
  type LoginResultDTO,
  type PathSettingsDTO,
  type ScreenshotDTO,
  type ScreenshotSearchDTO,
  type UserCacheDTO,
  type UserEncounterDTO,
  type UserProfileNavigationDTO,
  type VRChatConfigDTO,
  type VRChatCurrentUserDTO,
} from "../app";

type MockAppBindings = {
  [K in keyof AppBindings]: Mock<AppBindings[K]>;
};

function setWindowGoApp(mockBindings: MockAppBindings): void {
  window.go = { main: { App: mockBindings as AppBindings } };
}

function createMockBindings(): MockAppBindings {
  return {
    Greet: vi.fn(),
    LaunchProfiles: vi.fn(),
    LaunchVRChat: vi.fn(),
    LaunchVRChatWithArgs: vi.fn(),
    ParseLaunchArgsForGUI: vi.fn(),
    MergeLaunchArgsForGUI: vi.fn(),
    JoinWorld: vi.fn(),
    JoinWorldFromScreenshot: vi.fn(),
    SaveLaunchProfile: vi.fn(),
    DeleteLaunchProfile: vi.fn(),
    GetLogRetentionDays: vi.fn(),
    SetLogRetentionDays: vi.fn(),
    GetLanguage: vi.fn(),
    SetLanguage: vi.fn(),
    GetSystemLocale: vi.fn(),
    GetPathSettings: vi.fn(),
    SetPathSettings: vi.fn(),
    GetSuppressSleepWhileVRChat: vi.fn(),
    SetSuppressSleepWhileVRChat: vi.fn(),
    ValidatePath: vi.fn(),
    ValidateOutputLogPath: vi.fn(),
    OpenVRChatLogFolder: vi.fn(),
    OpenFileDialog: vi.fn(),
    OpenDirectoryDialog: vi.fn(),
    Screenshots: vi.fn(),
    SearchScreenshots: vi.fn(),
    GetScreenshot: vi.fn(),
    ScreenshotThumbnailDataURL: vi.fn(),
    OpenScreenshotExternally: vi.fn(),
    RevealScreenshotInFileManager: vi.fn(),
    ScanScreenshotDir: vi.fn(),
    IsGalleryScanning: vi.fn(),
    ReindexScreenshotDir: vi.fn(),
    Encounters: vi.fn(),
    EncountersByVRCUserID: vi.fn(),
    EncountersByWorldID: vi.fn(),
    RotateEncounters: vi.fn(),
    GetActivityStats: vi.fn(),
    Friends: vi.fn(),
    ResolveUserProfileNavigation: vi.fn(),
    SetFavorite: vi.fn(),
    SetStatus: vi.fn(),
    SetStatusDescription: vi.fn(),
    SetStatusAndDescription: vi.fn(),
    Login: vi.fn(),
    Logout: vi.fn(),
    IsLoggedIn: vi.fn(),
    HasStoredCredential: vi.fn(),
    GetCredentialBlob: vi.fn(),
    UnlockVRChatSession: vi.fn(),
    PersistWrappedCredential: vi.fn(),
    ClearStoredCredential: vi.fn(),
    GetVRChatCurrentUser: vi.fn(),
    RefreshFriends: vi.fn(),
    ReconcileVRChatSocialCache: vi.fn(),
    VacuumDb: vi.fn(),
    ClearEncounters: vi.fn(),
    ClearScreenshots: vi.fn(),
    ClearFriendsCache: vi.fn(),
    ListAutomationRules: vi.fn(),
    SaveAutomationRule: vi.fn(),
    DeleteAutomationRule: vi.fn(),
    ToggleAutomationRule: vi.fn(),
    VRChatConfigExists: vi.fn(),
    GetVRChatConfig: vi.fn(),
    SaveVRChatConfig: vi.fn(),
    DeleteVRChatConfig: vi.fn(),
    DefaultVRChatPictureFolder: vi.fn(),
  };
}

const sampleLaunchProfile: LaunchProfileDTO = {
  id: "lp-1",
  name: "Default",
  arguments: "-screen-fullscreen 1",
  isDefault: true,
};

const sampleLaunchArgs: LaunchArgsParsedDTO = {
  noVr: true,
  screenMode: "windowed",
  screenWidth: 1920,
  screenHeight: 1080,
  fps: 90,
  skipRegistry: false,
  processPriority: PRIORITY_OMIT,
  mainThreadPriority: 1,
  monitor: 1,
  profile: 0,
  enableDebugGui: false,
  enableSDKLogLevels: false,
  enableUdonDebugLogging: false,
  midi: "",
  watchWorlds: false,
  watchAvatars: false,
  ignoreTrackers: "",
  videoDecoding: "hardware",
  disableAMDStutterWorkaround: false,
  osc: "",
  affinity: "",
  enforceWorldServerChecks: false,
  custom: "",
};

const samplePathSettings: PathSettingsDTO = {
  vrchatPathWindows: "C:\\VRChat\\VRChat.exe",
  steamPathLinux: "/home/user/.steam",
  outputLogPath: "C:\\VRChat\\output_log.txt",
};

const sampleScreenshot: ScreenshotDTO = {
  id: "ss-1",
  filePath: "/pics/shot.png",
  worldId: "wrld_abc",
  worldName: "Test World",
};

const sampleEncounter: UserEncounterDTO = {
  id: "enc-1",
  vrcUserId: "usr_abc",
  displayName: "Alice",
  instanceId: "inst_1",
  joinedAt: "2025-01-01T00:00:00Z",
};

const sampleUser: UserCacheDTO = {
  vrcUserId: "usr_abc",
  displayName: "Alice",
  status: "active",
  isFavorite: false,
  lastUpdated: "2025-01-01T00:00:00Z",
};

const sampleNavigation: UserProfileNavigationDTO = {
  user: sampleUser,
  openInFriendsView: true,
};

const sampleLoginResult: LoginResultDTO = {
  ok: true,
  plaintextToken: "token",
};

const sampleCurrentUser: VRChatCurrentUserDTO = {
  id: "usr_abc",
  displayName: "Alice",
  username: "alice",
  status: "active",
  statusDescription: "hi",
  state: "online",
  currentAvatarThumbnailImageUrl: "https://example.com/a.png",
  userIcon: "https://example.com/icon.png",
  profilePicOverrideThumbnail: "",
};

const sampleActivityStats: ActivityStatsDTO = {
  dailyPlaySeconds: [{ date: "2025-01-01", seconds: 3600 }],
  topWorlds: [
    { worldId: "wrld_abc", worldName: "Test", seconds: 3600, sessions: 1 },
  ],
};

const sampleAutomationRule: AutomationRuleDTO = {
  id: "rule-1",
  name: "Rule",
  triggerType: "on_launch",
  conditionJson: "{}",
  actionType: "noop",
  actionPayload: "",
  isEnabled: true,
};

const sampleVRChatConfig: VRChatConfigDTO = {
  cameraResWidth: 1920,
  cameraResHeight: 1080,
  screenshotResWidth: 3840,
  screenshotResHeight: 2160,
  pictureOutputFolder: "/pics",
  pictureOutputSplitByDate: true,
  fpvSteadycamFov: 60,
  cacheDirectory: "/cache",
  cacheSize: 30,
  cacheExpiryDelay: 7,
  disableRichPresence: false,
};

describe("app exports", () => {
  it("exports PRIORITY_OMIT as -999", () => {
    expect(PRIORITY_OMIT).toBe(-999);
  });
});

describe("isWailsRuntime", () => {
  let mockBindings: MockAppBindings;
  let prevGo: typeof window.go;

  beforeEach(() => {
    prevGo = window.go;
    mockBindings = createMockBindings();
    setWindowGoApp(mockBindings);
  });

  afterEach(() => {
    window.go = prevGo;
  });

  it("returns true when window.go.main.App exists", () => {
    expect(isWailsRuntime()).toBe(true);
  });

  it("returns false when bindings are missing", () => {
    window.go = undefined;
    expect(isWailsRuntime()).toBe(false);
  });
});

describe("callApp", () => {
  let mockBindings: MockAppBindings;
  let prevGo: typeof window.go;

  beforeEach(() => {
    prevGo = window.go;
    mockBindings = createMockBindings();
    setWindowGoApp(mockBindings);
  });

  afterEach(() => {
    window.go = prevGo;
  });

  it("returns fallback when bindings are missing (Vitest skips dev wait)", async () => {
    window.go = undefined;
    const out = await callApp(async () => "invoked", "fallback");
    expect(out).toBe("fallback");
  });

  it("invokes fn with bindings and returns its result", async () => {
    mockBindings.Greet.mockResolvedValue("from-go");
    const out = await callApp((a) => a.Greet("Alice"), "fallback");
    expect(mockBindings.Greet).toHaveBeenCalledWith("Alice");
    expect(out).toBe("from-go");
  });

  it("propagates rejection from the binding", async () => {
    mockBindings.Greet.mockRejectedValue(new Error("backend failed"));
    await expect(callApp((a) => a.Greet("Bob"), "fallback")).rejects.toThrow(
      "backend failed",
    );
  });
});

describe("App bindings", () => {
  let mockBindings: MockAppBindings;
  let prevGo: typeof window.go;

  beforeEach(() => {
    prevGo = window.go;
    mockBindings = createMockBindings();
    setWindowGoApp(mockBindings);
  });

  afterEach(() => {
    window.go = prevGo;
  });

  describe("launcher", () => {
    it("greet delegates to Greet", async () => {
      mockBindings.Greet.mockResolvedValue("Hello Alice");
      await expect(App.greet("Alice")).resolves.toBe("Hello Alice");
      expect(mockBindings.Greet).toHaveBeenCalledWith("Alice");
    });

    it("launchProfiles delegates to LaunchProfiles", async () => {
      mockBindings.LaunchProfiles.mockResolvedValue([sampleLaunchProfile]);
      await expect(App.launchProfiles()).resolves.toEqual([
        sampleLaunchProfile,
      ]);
      expect(mockBindings.LaunchProfiles).toHaveBeenCalled();
    });

    it("launchVRChat delegates to LaunchVRChat", async () => {
      mockBindings.LaunchVRChat.mockResolvedValue(undefined);
      await App.launchVRChat("lp-1");
      expect(mockBindings.LaunchVRChat).toHaveBeenCalledWith("lp-1");
    });

    it("launchVRChatWithArgs delegates to LaunchVRChatWithArgs", async () => {
      mockBindings.LaunchVRChatWithArgs.mockResolvedValue(undefined);
      await App.launchVRChatWithArgs("-no-vr");
      expect(mockBindings.LaunchVRChatWithArgs).toHaveBeenCalledWith("-no-vr");
    });

    it("parseLaunchArgsForGUI delegates to ParseLaunchArgsForGUI", async () => {
      mockBindings.ParseLaunchArgsForGUI.mockResolvedValue(sampleLaunchArgs);
      await expect(App.parseLaunchArgsForGUI("-no-vr")).resolves.toEqual(
        sampleLaunchArgs,
      );
      expect(mockBindings.ParseLaunchArgsForGUI).toHaveBeenCalledWith("-no-vr");
    });

    it("mergeLaunchArgsForGUI delegates to MergeLaunchArgsForGUI", async () => {
      mockBindings.MergeLaunchArgsForGUI.mockResolvedValue("-no-vr");
      await expect(App.mergeLaunchArgsForGUI(sampleLaunchArgs)).resolves.toBe(
        "-no-vr",
      );
      expect(mockBindings.MergeLaunchArgsForGUI).toHaveBeenCalledWith(
        sampleLaunchArgs,
      );
    });

    it("joinWorld delegates to JoinWorld", async () => {
      mockBindings.JoinWorld.mockResolvedValue(undefined);
      await App.joinWorld("wrld_abc");
      expect(mockBindings.JoinWorld).toHaveBeenCalledWith("wrld_abc");
    });

    it("joinWorldFromScreenshot delegates to JoinWorldFromScreenshot", async () => {
      mockBindings.JoinWorldFromScreenshot.mockResolvedValue(undefined);
      await App.joinWorldFromScreenshot("ss-1");
      expect(mockBindings.JoinWorldFromScreenshot).toHaveBeenCalledWith("ss-1");
    });

    it("saveLaunchProfile delegates to SaveLaunchProfile", async () => {
      mockBindings.SaveLaunchProfile.mockResolvedValue(undefined);
      await App.saveLaunchProfile(sampleLaunchProfile);
      expect(mockBindings.SaveLaunchProfile).toHaveBeenCalledWith(
        sampleLaunchProfile,
      );
    });

    it("deleteLaunchProfile delegates to DeleteLaunchProfile", async () => {
      mockBindings.DeleteLaunchProfile.mockResolvedValue(undefined);
      await App.deleteLaunchProfile("lp-1");
      expect(mockBindings.DeleteLaunchProfile).toHaveBeenCalledWith("lp-1");
    });
  });

  describe("settings", () => {
    it("getLogRetentionDays delegates to GetLogRetentionDays", async () => {
      mockBindings.GetLogRetentionDays.mockResolvedValue(14);
      await expect(App.getLogRetentionDays()).resolves.toBe(14);
    });

    it("setLogRetentionDays delegates to SetLogRetentionDays", async () => {
      mockBindings.SetLogRetentionDays.mockResolvedValue(undefined);
      await App.setLogRetentionDays(14);
      expect(mockBindings.SetLogRetentionDays).toHaveBeenCalledWith(14);
    });

    it("getLanguage delegates to GetLanguage", async () => {
      mockBindings.GetLanguage.mockResolvedValue("ja");
      await expect(App.getLanguage()).resolves.toBe("ja");
    });

    it("setLanguage delegates to SetLanguage", async () => {
      mockBindings.SetLanguage.mockResolvedValue(undefined);
      await App.setLanguage("en");
      expect(mockBindings.SetLanguage).toHaveBeenCalledWith("en");
    });

    it("getSystemLocale delegates to GetSystemLocale", async () => {
      mockBindings.GetSystemLocale.mockResolvedValue("ja-JP");
      await expect(App.getSystemLocale()).resolves.toBe("ja-JP");
    });

    it("getPathSettings delegates to GetPathSettings", async () => {
      mockBindings.GetPathSettings.mockResolvedValue(samplePathSettings);
      await expect(App.getPathSettings()).resolves.toEqual(samplePathSettings);
    });

    it("setPathSettings delegates to SetPathSettings", async () => {
      mockBindings.SetPathSettings.mockResolvedValue(undefined);
      await App.setPathSettings(samplePathSettings);
      expect(mockBindings.SetPathSettings).toHaveBeenCalledWith(
        samplePathSettings,
      );
    });

    it("getSuppressSleepWhileVRChat delegates to GetSuppressSleepWhileVRChat", async () => {
      mockBindings.GetSuppressSleepWhileVRChat.mockResolvedValue(true);
      await expect(App.getSuppressSleepWhileVRChat()).resolves.toBe(true);
    });

    it("setSuppressSleepWhileVRChat delegates to SetSuppressSleepWhileVRChat", async () => {
      mockBindings.SetSuppressSleepWhileVRChat.mockResolvedValue(undefined);
      await App.setSuppressSleepWhileVRChat(true);
      expect(mockBindings.SetSuppressSleepWhileVRChat).toHaveBeenCalledWith(
        true,
      );
    });

    it("validatePath delegates to ValidatePath", async () => {
      mockBindings.ValidatePath.mockResolvedValue(true);
      await expect(App.validatePath("/path")).resolves.toBe(true);
      expect(mockBindings.ValidatePath).toHaveBeenCalledWith("/path");
    });

    it("validateOutputLogPath delegates to ValidateOutputLogPath", async () => {
      mockBindings.ValidateOutputLogPath.mockResolvedValue(false);
      await expect(App.validateOutputLogPath("/log.txt")).resolves.toBe(false);
      expect(mockBindings.ValidateOutputLogPath).toHaveBeenCalledWith(
        "/log.txt",
      );
    });

    it("openVRChatLogFolder delegates to OpenVRChatLogFolder", async () => {
      mockBindings.OpenVRChatLogFolder.mockResolvedValue(undefined);
      await App.openVRChatLogFolder();
      expect(mockBindings.OpenVRChatLogFolder).toHaveBeenCalled();
    });
  });

  describe("dialogs", () => {
    it("openFileDialog returns path when non-empty", async () => {
      mockBindings.OpenFileDialog.mockResolvedValue("/file.txt");
      await expect(
        App.openFileDialog("Pick file", "/dir", "*.txt"),
      ).resolves.toBe("/file.txt");
      expect(mockBindings.OpenFileDialog).toHaveBeenCalledWith(
        "Pick file",
        "/dir",
        "*.txt",
      );
    });

    it("openFileDialog returns null when binding returns empty string", async () => {
      mockBindings.OpenFileDialog.mockResolvedValue("");
      await expect(
        App.openFileDialog("Pick file", "/dir", "*.txt"),
      ).resolves.toBeNull();
    });

    it("openDirectoryDialog returns path when non-empty", async () => {
      mockBindings.OpenDirectoryDialog.mockResolvedValue("/dir");
      await expect(App.openDirectoryDialog("Pick dir", "/start")).resolves.toBe(
        "/dir",
      );
      expect(mockBindings.OpenDirectoryDialog).toHaveBeenCalledWith(
        "Pick dir",
        "/start",
      );
    });

    it("openDirectoryDialog returns null when binding returns empty string", async () => {
      mockBindings.OpenDirectoryDialog.mockResolvedValue("");
      await expect(
        App.openDirectoryDialog("Pick dir", "/start"),
      ).resolves.toBeNull();
    });
  });

  describe("gallery", () => {
    it("screenshots passes empty string when worldId omitted", async () => {
      mockBindings.Screenshots.mockResolvedValue([sampleScreenshot]);
      await expect(App.screenshots()).resolves.toEqual([sampleScreenshot]);
      expect(mockBindings.Screenshots).toHaveBeenCalledWith("");
    });

    it("screenshots passes worldId when provided", async () => {
      mockBindings.Screenshots.mockResolvedValue([]);
      await App.screenshots("wrld_abc");
      expect(mockBindings.Screenshots).toHaveBeenCalledWith("wrld_abc");
    });

    it("searchScreenshots delegates to SearchScreenshots", async () => {
      const filter: ScreenshotSearchDTO = { worldName: "Test" };
      mockBindings.SearchScreenshots.mockResolvedValue([sampleScreenshot]);
      await expect(App.searchScreenshots(filter)).resolves.toEqual([
        sampleScreenshot,
      ]);
      expect(mockBindings.SearchScreenshots).toHaveBeenCalledWith(filter);
    });

    it("getScreenshot delegates to GetScreenshot", async () => {
      mockBindings.GetScreenshot.mockResolvedValue(sampleScreenshot);
      await expect(App.getScreenshot("ss-1")).resolves.toEqual(
        sampleScreenshot,
      );
      expect(mockBindings.GetScreenshot).toHaveBeenCalledWith("ss-1");
    });

    it("screenshotThumbnailDataURL delegates to ScreenshotThumbnailDataURL", async () => {
      mockBindings.ScreenshotThumbnailDataURL.mockResolvedValue(
        "data:image/png",
      );
      await expect(App.screenshotThumbnailDataURL("ss-1")).resolves.toBe(
        "data:image/png",
      );
      expect(mockBindings.ScreenshotThumbnailDataURL).toHaveBeenCalledWith(
        "ss-1",
      );
    });

    it("openScreenshotExternally delegates to OpenScreenshotExternally", async () => {
      mockBindings.OpenScreenshotExternally.mockResolvedValue(undefined);
      await App.openScreenshotExternally("ss-1");
      expect(mockBindings.OpenScreenshotExternally).toHaveBeenCalledWith(
        "ss-1",
      );
    });

    it("revealScreenshotInFileManager delegates to RevealScreenshotInFileManager", async () => {
      mockBindings.RevealScreenshotInFileManager.mockResolvedValue(undefined);
      await App.revealScreenshotInFileManager("ss-1");
      expect(mockBindings.RevealScreenshotInFileManager).toHaveBeenCalledWith(
        "ss-1",
      );
    });

    it("scanScreenshotDir delegates to ScanScreenshotDir", async () => {
      mockBindings.ScanScreenshotDir.mockResolvedValue(5);
      await expect(App.scanScreenshotDir("/pics")).resolves.toBe(5);
      expect(mockBindings.ScanScreenshotDir).toHaveBeenCalledWith("/pics");
    });

    it("isGalleryScanning delegates to IsGalleryScanning", async () => {
      mockBindings.IsGalleryScanning.mockResolvedValue(true);
      await expect(App.isGalleryScanning()).resolves.toBe(true);
    });

    it("reindexScreenshotDir delegates to ReindexScreenshotDir", async () => {
      mockBindings.ReindexScreenshotDir.mockResolvedValue(3);
      await expect(App.reindexScreenshotDir("/pics")).resolves.toBe(3);
      expect(mockBindings.ReindexScreenshotDir).toHaveBeenCalledWith("/pics");
    });

    it("clearScreenshots delegates to ClearScreenshots", async () => {
      mockBindings.ClearScreenshots.mockResolvedValue(10);
      await expect(App.clearScreenshots()).resolves.toBe(10);
    });
  });

  describe("encounters & activity", () => {
    it("encounters delegates to Encounters", async () => {
      mockBindings.Encounters.mockResolvedValue([sampleEncounter]);
      await expect(App.encounters()).resolves.toEqual([sampleEncounter]);
    });

    it("encountersByVRCUserID delegates to EncountersByVRCUserID", async () => {
      mockBindings.EncountersByVRCUserID.mockResolvedValue([sampleEncounter]);
      await expect(App.encountersByVRCUserID("usr_abc")).resolves.toEqual([
        sampleEncounter,
      ]);
      expect(mockBindings.EncountersByVRCUserID).toHaveBeenCalledWith(
        "usr_abc",
      );
    });

    it("encountersByWorldID delegates to EncountersByWorldID", async () => {
      mockBindings.EncountersByWorldID.mockResolvedValue([]);
      await App.encountersByWorldID("wrld_abc");
      expect(mockBindings.EncountersByWorldID).toHaveBeenCalledWith("wrld_abc");
    });

    it("clearEncounters delegates to ClearEncounters", async () => {
      mockBindings.ClearEncounters.mockResolvedValue(2);
      await expect(App.clearEncounters()).resolves.toBe(2);
    });

    it("getActivityStats delegates to GetActivityStats", async () => {
      mockBindings.GetActivityStats.mockResolvedValue(sampleActivityStats);
      await expect(
        App.getActivityStats("2025-01-01", "2025-01-31"),
      ).resolves.toEqual(sampleActivityStats);
      expect(mockBindings.GetActivityStats).toHaveBeenCalledWith(
        "2025-01-01",
        "2025-01-31",
      );
    });
  });

  describe("friends & social", () => {
    it("friends delegates to Friends", async () => {
      mockBindings.Friends.mockResolvedValue([sampleUser]);
      await expect(App.friends()).resolves.toEqual([sampleUser]);
    });

    it("resolveUserProfileNavigation delegates to ResolveUserProfileNavigation", async () => {
      mockBindings.ResolveUserProfileNavigation.mockResolvedValue(
        sampleNavigation,
      );
      await expect(
        App.resolveUserProfileNavigation("usr_abc"),
      ).resolves.toEqual(sampleNavigation);
      expect(mockBindings.ResolveUserProfileNavigation).toHaveBeenCalledWith(
        "usr_abc",
      );
    });

    it("setFavorite delegates to SetFavorite", async () => {
      mockBindings.SetFavorite.mockResolvedValue(undefined);
      await App.setFavorite("usr_abc", true);
      expect(mockBindings.SetFavorite).toHaveBeenCalledWith("usr_abc", true);
    });

    it("setStatus delegates to SetStatus", async () => {
      mockBindings.SetStatus.mockResolvedValue(undefined);
      await App.setStatus("join me");
      expect(mockBindings.SetStatus).toHaveBeenCalledWith("join me");
    });

    it("setStatusDescription delegates to SetStatusDescription", async () => {
      mockBindings.SetStatusDescription.mockResolvedValue(undefined);
      await App.setStatusDescription("hello");
      expect(mockBindings.SetStatusDescription).toHaveBeenCalledWith("hello");
    });

    it("setStatusAndDescription delegates to SetStatusAndDescription", async () => {
      mockBindings.SetStatusAndDescription.mockResolvedValue(undefined);
      await App.setStatusAndDescription("active", "busy");
      expect(mockBindings.SetStatusAndDescription).toHaveBeenCalledWith(
        "active",
        "busy",
      );
    });

    it("refreshFriends delegates to RefreshFriends", async () => {
      mockBindings.RefreshFriends.mockResolvedValue(undefined);
      await App.refreshFriends();
      expect(mockBindings.RefreshFriends).toHaveBeenCalled();
    });

    it("reconcileVRChatSocialCache delegates to ReconcileVRChatSocialCache", async () => {
      mockBindings.ReconcileVRChatSocialCache.mockResolvedValue(undefined);
      await App.reconcileVRChatSocialCache();
      expect(mockBindings.ReconcileVRChatSocialCache).toHaveBeenCalled();
    });

    it("clearFriendsCache delegates to ClearFriendsCache", async () => {
      mockBindings.ClearFriendsCache.mockResolvedValue(4);
      await expect(App.clearFriendsCache()).resolves.toBe(4);
    });
  });

  describe("auth", () => {
    it("login passes empty twoFactorCode when omitted", async () => {
      mockBindings.Login.mockResolvedValue(sampleLoginResult);
      await expect(App.login("user", "pass")).resolves.toEqual(
        sampleLoginResult,
      );
      expect(mockBindings.Login).toHaveBeenCalledWith("user", "pass", "");
    });

    it("login passes twoFactorCode when provided", async () => {
      mockBindings.Login.mockResolvedValue(sampleLoginResult);
      await App.login("user", "pass", "123456");
      expect(mockBindings.Login).toHaveBeenCalledWith("user", "pass", "123456");
    });

    it("logout delegates to Logout", async () => {
      mockBindings.Logout.mockResolvedValue(undefined);
      await App.logout();
      expect(mockBindings.Logout).toHaveBeenCalled();
    });

    it("isLoggedIn delegates to IsLoggedIn", async () => {
      mockBindings.IsLoggedIn.mockResolvedValue(true);
      await expect(App.isLoggedIn()).resolves.toBe(true);
    });

    it("hasStoredCredential delegates to HasStoredCredential", async () => {
      mockBindings.HasStoredCredential.mockResolvedValue(true);
      await expect(App.hasStoredCredential()).resolves.toBe(true);
    });

    it("getCredentialBlob delegates to GetCredentialBlob", async () => {
      mockBindings.GetCredentialBlob.mockResolvedValue("blob");
      await expect(App.getCredentialBlob()).resolves.toBe("blob");
    });

    it("unlockVRChatSession delegates to UnlockVRChatSession", async () => {
      mockBindings.UnlockVRChatSession.mockResolvedValue(undefined);
      await App.unlockVRChatSession("token");
      expect(mockBindings.UnlockVRChatSession).toHaveBeenCalledWith("token");
    });

    it("persistWrappedCredential delegates to PersistWrappedCredential", async () => {
      mockBindings.PersistWrappedCredential.mockResolvedValue(undefined);
      await App.persistWrappedCredential("wrapped");
      expect(mockBindings.PersistWrappedCredential).toHaveBeenCalledWith(
        "wrapped",
      );
    });

    it("clearStoredCredential delegates to ClearStoredCredential", async () => {
      mockBindings.ClearStoredCredential.mockResolvedValue(undefined);
      await App.clearStoredCredential();
      expect(mockBindings.ClearStoredCredential).toHaveBeenCalled();
    });

    it("getVRChatCurrentUser defaults forceRefresh to false", async () => {
      mockBindings.GetVRChatCurrentUser.mockResolvedValue(sampleCurrentUser);
      await expect(App.getVRChatCurrentUser()).resolves.toEqual(
        sampleCurrentUser,
      );
      expect(mockBindings.GetVRChatCurrentUser).toHaveBeenCalledWith(false);
    });

    it("getVRChatCurrentUser passes forceRefresh when provided", async () => {
      mockBindings.GetVRChatCurrentUser.mockResolvedValue(sampleCurrentUser);
      await App.getVRChatCurrentUser(true);
      expect(mockBindings.GetVRChatCurrentUser).toHaveBeenCalledWith(true);
    });
  });

  describe("maintenance", () => {
    it("vacuumDb delegates to VacuumDb", async () => {
      mockBindings.VacuumDb.mockResolvedValue(undefined);
      await App.vacuumDb();
      expect(mockBindings.VacuumDb).toHaveBeenCalled();
    });
  });

  describe("automation", () => {
    it("listAutomationRules delegates to ListAutomationRules", async () => {
      mockBindings.ListAutomationRules.mockResolvedValue([
        sampleAutomationRule,
      ]);
      await expect(App.listAutomationRules()).resolves.toEqual([
        sampleAutomationRule,
      ]);
    });

    it("saveAutomationRule delegates to SaveAutomationRule", async () => {
      mockBindings.SaveAutomationRule.mockResolvedValue(undefined);
      await App.saveAutomationRule(sampleAutomationRule);
      expect(mockBindings.SaveAutomationRule).toHaveBeenCalledWith(
        sampleAutomationRule,
      );
    });

    it("deleteAutomationRule delegates to DeleteAutomationRule", async () => {
      mockBindings.DeleteAutomationRule.mockResolvedValue(undefined);
      await App.deleteAutomationRule("rule-1");
      expect(mockBindings.DeleteAutomationRule).toHaveBeenCalledWith("rule-1");
    });

    it("toggleAutomationRule delegates to ToggleAutomationRule", async () => {
      mockBindings.ToggleAutomationRule.mockResolvedValue(undefined);
      await App.toggleAutomationRule("rule-1", false);
      expect(mockBindings.ToggleAutomationRule).toHaveBeenCalledWith(
        "rule-1",
        false,
      );
    });
  });

  describe("vrchat config", () => {
    it("vrchatConfigExists delegates to VRChatConfigExists", async () => {
      mockBindings.VRChatConfigExists.mockResolvedValue(true);
      await expect(App.vrchatConfigExists()).resolves.toBe(true);
    });

    it("getVRChatConfig delegates to GetVRChatConfig", async () => {
      mockBindings.GetVRChatConfig.mockResolvedValue(sampleVRChatConfig);
      await expect(App.getVRChatConfig()).resolves.toEqual(sampleVRChatConfig);
    });

    it("saveVRChatConfig delegates to SaveVRChatConfig", async () => {
      mockBindings.SaveVRChatConfig.mockResolvedValue(undefined);
      await App.saveVRChatConfig(sampleVRChatConfig);
      expect(mockBindings.SaveVRChatConfig).toHaveBeenCalledWith(
        sampleVRChatConfig,
      );
    });

    it("deleteVRChatConfig delegates to DeleteVRChatConfig", async () => {
      mockBindings.DeleteVRChatConfig.mockResolvedValue(undefined);
      await App.deleteVRChatConfig();
      expect(mockBindings.DeleteVRChatConfig).toHaveBeenCalled();
    });

    it("defaultVRChatPictureFolder delegates to DefaultVRChatPictureFolder", async () => {
      mockBindings.DefaultVRChatPictureFolder.mockResolvedValue("/Pictures");
      await expect(App.defaultVRChatPictureFolder()).resolves.toBe("/Pictures");
    });
  });
});

describe("App fallbacks without bindings", () => {
  let prevGo: typeof window.go;

  beforeEach(() => {
    prevGo = window.go;
    window.go = undefined;
  });

  afterEach(() => {
    window.go = prevGo;
  });

  it("greet returns Hello fallback", async () => {
    await expect(App.greet("Bob")).resolves.toBe("Hello Bob, Welcome!");
  });

  it("parseLaunchArgsForGUI returns default DTO with PRIORITY_OMIT", async () => {
    const dto = await App.parseLaunchArgsForGUI("");
    expect(dto.processPriority).toBe(PRIORITY_OMIT);
    expect(dto.mainThreadPriority).toBe(PRIORITY_OMIT);
    expect(dto.fps).toBe(90);
  });

  it("login returns unavailable error", async () => {
    await expect(App.login("u", "p")).resolves.toEqual({
      ok: false,
      error: "App not available",
    });
  });

  it("openFileDialog returns null without bindings", async () => {
    await expect(App.openFileDialog("t", "d", "*")).resolves.toBeNull();
  });

  it("getLogRetentionDays returns 30", async () => {
    await expect(App.getLogRetentionDays()).resolves.toBe(30);
  });
});

describe("callApp DEV binding race recovery", () => {
  let prevGo: typeof window.go;
  const injectedScripts: HTMLScriptElement[] = [];
  let origCreateElement: typeof document.createElement;

  function removeAllWailsScripts() {
    for (const el of document.querySelectorAll('head script[src*="wails/"]')) {
      el.remove();
    }
  }

  function injectWailsScript(src = "/wails/runtime.js") {
    const script = document.createElement("script");
    script.setAttribute("src", src);
    document.head.appendChild(script);
    injectedScripts.push(script);
    return script;
  }

  beforeEach(() => {
    prevGo = window.go;
    removeAllWailsScripts();
    injectedScripts.length = 0;
    vi.resetModules();
    vi.stubEnv("DEV", true);
    vi.stubEnv("MODE", "development");
    origCreateElement = document.createElement.bind(document);
    vi.spyOn(document, "createElement").mockImplementation((tag, options) => {
      const el = origCreateElement(tag, options);
      if (String(tag).toLowerCase() === "script") {
        queueMicrotask(() => {
          el.dispatchEvent(new Event("load"));
        });
      }
      return el;
    });
  });

  afterEach(() => {
    window.go = prevGo;
    for (const script of injectedScripts.splice(0)) {
      script.remove();
    }
    removeAllWailsScripts();
    vi.restoreAllMocks();
    vi.unstubAllEnvs();
    vi.resetModules();
  });

  it("waits for bindings via requestAnimationFrame when scripts expect Wails", async () => {
    injectWailsScript();
    window.go = undefined;

    const mockBindings = createMockBindings();
    mockBindings.Greet.mockResolvedValue("from-wait");

    let frames = 0;
    vi.spyOn(window, "requestAnimationFrame").mockImplementation((cb) => {
      frames += 1;
      if (frames === 1) {
        setWindowGoApp(mockBindings);
      }
      cb(0);
      return frames;
    });

    const { callApp: freshCallApp } = await import("../app");
    await expect(
      freshCallApp((a) => a.Greet("Wait"), "fallback"),
    ).resolves.toBe("from-wait");
  });

  it("returns fallback after exhausting rAF wait without bindings", async () => {
    injectWailsScript();
    window.go = undefined;

    vi.spyOn(window, "requestAnimationFrame").mockImplementation((cb) => {
      cb(0);
      return 1;
    });

    const { callApp: freshCallApp } = await import("../app");
    await expect(freshCallApp(async () => "never", "fallback")).resolves.toBe(
      "fallback",
    );
  });

  it("reloads wails scripts once then falls back when bindings never appear", async () => {
    injectWailsScript("/wails/ipc.js");
    window.go = undefined;

    vi.spyOn(window, "requestAnimationFrame").mockImplementation((cb) => {
      cb(0);
      return 1;
    });

    const { callApp: freshCallApp } = await import("../app");
    await expect(freshCallApp(async () => "never", "fallback")).resolves.toBe(
      "fallback",
    );
    expect(
      document.querySelector('script[src*="wails/ipc.js?wailsRetry="]'),
    ).not.toBeNull();
  });

  it("falls back when dev script reload fails", async () => {
    injectWailsScript("/wails/ipc.js");
    window.go = undefined;

    vi.spyOn(window, "requestAnimationFrame").mockImplementation((cb) => {
      cb(0);
      return 1;
    });

    vi.mocked(document.createElement).mockImplementation((tag, options) => {
      const el = origCreateElement(tag, options);
      if (String(tag).toLowerCase() === "script") {
        queueMicrotask(() => {
          el.dispatchEvent(new Event("error"));
        });
      }
      return el;
    });

    const { callApp: freshCallApp } = await import("../app");
    await expect(freshCallApp(async () => "never", "fallback")).resolves.toBe(
      "fallback",
    );
  });

  it("skips dev wait when head has no wails script tags", async () => {
    window.go = undefined;
    const raf = vi.spyOn(window, "requestAnimationFrame");

    const { callApp: freshCallApp } = await import("../app");
    await expect(freshCallApp(async () => "never", "fallback")).resolves.toBe(
      "fallback",
    );
    expect(raf).not.toHaveBeenCalled();
  });

  it("treats page as not expecting Wails when document is unavailable", async () => {
    window.go = undefined;
    const doc = globalThis.document;
    // @ts-expect-error test-only: exercise pageExpectsWailsBindings guard
    delete globalThis.document;

    try {
      const { callApp: freshCallApp } = await import("../app");
      await expect(freshCallApp(async () => "never", "fallback")).resolves.toBe(
        "fallback",
      );
    } finally {
      globalThis.document = doc;
    }
  });

  it("rejects reload when runtime script fails to load", async () => {
    injectWailsScript("/wails/ipc.js");
    window.go = undefined;

    vi.spyOn(window, "requestAnimationFrame").mockImplementation((cb) => {
      cb(0);
      return 1;
    });

    vi.mocked(document.createElement).mockImplementation((tag, options) => {
      const el = origCreateElement(tag, options);
      if (String(tag).toLowerCase() === "script") {
        queueMicrotask(() => {
          const src = el.getAttribute("src") ?? "";
          el.dispatchEvent(
            new Event(src.includes("runtime.js") ? "error" : "load"),
          );
        });
      }
      return el;
    });

    const { callApp: freshCallApp } = await import("../app");
    await expect(freshCallApp(async () => "never", "fallback")).resolves.toBe(
      "fallback",
    );
  });
});
