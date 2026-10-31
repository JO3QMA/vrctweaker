import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { ElMessageBox } from "element-plus";
import { createRouter, createWebHashHistory } from "vue-router";
import SettingsView from "../SettingsView.vue";
import { settingsChildRoutes } from "../settings/settingsSections";
import { App } from "../../wails/app";
import * as I18nModule from "../../i18n";
import { resetSessionUnlockForStorybook } from "../../composables/useSessionUnlock";
import { showToast } from "../../utils/showToast";

vi.mock("../../utils/showToast", () => ({
  showToast: {
    success: vi.fn(),
    warning: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
  },
}));

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: "/settings",
      component: SettingsView,
      children: settingsChildRoutes(),
    },
    { path: "/me", name: "me", component: { template: "<div>Me</div>" } },
    {
      path: "/video",
      name: "video",
      component: { template: "<div>Video</div>" },
    },
    { path: "/licenses", component: { template: "<div>Licenses</div>" } },
    {
      path: "/legal/privacy",
      name: "legal-privacy",
      component: { template: "<div>Privacy</div>" },
    },
    {
      path: "/legal/terms",
      name: "legal-terms",
      component: { template: "<div>Terms</div>" },
    },
  ],
});

const defaultPathSettings = {
  vrchatPathWindows: "C:\\VRChat\\VRChat.exe",
  steamPathLinux: "/home/user/.steam/steam",
  outputLogPath:
    "C:\\Users\\x\\AppData\\LocalLow\\VRChat\\VRChat\\output_log.txt",
};

let activeWrapper: ReturnType<typeof mount> | undefined;

function mountSettings() {
  activeWrapper?.unmount();
  activeWrapper = mount(SettingsView, { global: { plugins: [router] } });
  return activeWrapper;
}

async function mountAt(path: string) {
  activeWrapper?.unmount();
  activeWrapper = undefined;
  await router.push(path);
  await router.isReady();
  const wrapper = mountSettings();
  await flushPromises();
  return wrapper;
}

function pathRowInput(wrapper: ReturnType<typeof mount>, rowIndex: number) {
  return wrapper.findAll(".path-row .el-input input")[rowIndex]!;
}

const pathFieldKeys = [
  "vrchatPathWindows",
  "steamPathLinux",
  "outputLogPath",
] as const;

function validateExistsBtn(
  wrapper: ReturnType<typeof mount>,
  rowIndex: number,
) {
  return wrapper.find(
    `[data-testid="path-validate-${pathFieldKeys[rowIndex]}"]`,
  );
}

function setupAppMocks() {
  vi.spyOn(App, "hasStoredCredential").mockResolvedValue(false);
  vi.spyOn(App, "isLoggedIn").mockResolvedValue(false);
  vi.spyOn(App, "getLogRetentionDays").mockResolvedValue(45);
  vi.spyOn(App, "getGalleryAutoEnrichMetadata").mockResolvedValue(true);
  vi.spyOn(App, "setGalleryAutoEnrichMetadata").mockResolvedValue(undefined);
  vi.spyOn(App, "getSuppressSleepWhileVRChat").mockResolvedValue(true);
  vi.spyOn(App, "getCloseToTray").mockResolvedValue(true);
  vi.spyOn(App, "getAppUpdateAutoCheck").mockResolvedValue(true);
  vi.spyOn(App, "setAppUpdateAutoCheck").mockResolvedValue(undefined);
  vi.spyOn(App, "checkForAppUpdate").mockResolvedValue({
    currentVersion: "0.1.0",
    updateAvailable: false,
    shouldPrompt: false,
    latestVersion: "0.1.0",
    latestTag: "v0.1.0",
    releaseNotes: "",
    downloadUrl: "",
    zipSha256: "",
    releasePageUrl: "",
    autoApplySupported: true,
  });
  vi.spyOn(App, "getAppVersion").mockResolvedValue("0.1.0");
  vi.spyOn(App, "getPathSettings").mockResolvedValue({
    ...defaultPathSettings,
  });
  vi.spyOn(App, "setPathSettings").mockResolvedValue(undefined);
  vi.spyOn(App, "setLogRetentionDays").mockResolvedValue(undefined);
  vi.spyOn(App, "setSuppressSleepWhileVRChat").mockResolvedValue(undefined);
  vi.spyOn(App, "setCloseToTray").mockResolvedValue(undefined);
  vi.spyOn(App, "openFileDialog").mockResolvedValue(null);
  vi.spyOn(App, "openDirectoryDialog").mockResolvedValue(null);
  vi.spyOn(App, "validatePath").mockResolvedValue(true);
  vi.spyOn(App, "validateOutputLogPath").mockResolvedValue(true);
  vi.spyOn(App, "openVRChatLogFolder").mockResolvedValue(undefined);
  vi.spyOn(App, "vacuumDb").mockResolvedValue(undefined);
  vi.spyOn(App, "clearEncounters").mockResolvedValue(5);
  vi.spyOn(App, "clearScreenshots").mockResolvedValue(3);
  vi.spyOn(App, "clearFriendsCache").mockResolvedValue(2);
  vi.spyOn(App, "setLanguage").mockResolvedValue(undefined);
}

describe("SettingsView", () => {
  beforeEach(async () => {
    resetSessionUnlockForStorybook();
    await router.push("/settings");
    await router.isReady();
    setupAppMocks();
  });

  afterEach(() => {
    activeWrapper?.unmount();
    activeWrapper = undefined;
    vi.restoreAllMocks();
  });

  it("renders settings title", async () => {
    const wrapper = mountSettings();
    await flushPromises();
    expect(wrapper.find(".page-title").text()).toBe("設定");
  });

  it("redirects /settings to the account page", async () => {
    await router.push("/settings");
    await router.isReady();
    expect(router.currentRoute.value.path).toBe("/settings/account");
    const wrapper = mountSettings();
    await flushPromises();
    expect(
      wrapper.find('[data-testid="settings-account-panel"]').exists(),
    ).toBe(true);
    expect(
      wrapper.get('[data-testid="settings-credential-keyring"]').text(),
    ).toContain("キーリング");
  });

  it("lists five nav entries with a divider before reset", async () => {
    const wrapper = mountSettings();
    await flushPromises();
    const ids = [
      ...wrapper.get('[data-testid="settings-nav"]').element.children,
    ].map((el) => el.getAttribute("data-testid"));
    expect(ids).toEqual([
      "settings-nav-account",
      "settings-nav-general",
      "settings-nav-data",
      "settings-nav-about",
      "settings-nav-divider",
      "settings-nav-reset",
    ]);
  });

  it("has link to OSS licenses page in about section", async () => {
    const wrapper = await mountAt("/settings/about");
    const licensesLink = wrapper.find(
      '[data-testid="settings-oss-licenses-link"]',
    );
    expect(licensesLink.exists()).toBe(true);
    expect(licensesLink.attributes("href")).toContain("/licenses");
    expect(licensesLink.text()).toContain("OSSライセンス一覧");
  });

  it("shows legal document links in about section", async () => {
    const wrapper = await mountAt("/settings/about");
    const privacy = wrapper.find('[data-testid="legal-link-privacy"]');
    expect(privacy.exists()).toBe(true);
    expect(privacy.attributes("href")).toContain("/legal/privacy");
    const terms = wrapper.find('[data-testid="legal-link-terms"]');
    expect(terms.attributes("href")).toContain("/legal/terms");
  });

  it("loads path settings on mount", async () => {
    const wrapper = await mountAt("/settings/data");

    expect(App.getPathSettings).toHaveBeenCalled();
    expect((pathRowInput(wrapper, 0).element as HTMLInputElement).value).toBe(
      defaultPathSettings.vrchatPathWindows,
    );
    expect((pathRowInput(wrapper, 1).element as HTMLInputElement).value).toBe(
      defaultPathSettings.steamPathLinux,
    );
    expect((pathRowInput(wrapper, 2).element as HTMLInputElement).value).toBe(
      defaultPathSettings.outputLogPath,
    );
    expect(
      wrapper.findAll(".path-row")[1]?.find(".path-row__desc").text(),
    ).toBe("空欄なら既定のコマンドを使います。");
    const browseLabels = [
      "vrchat-path-browse",
      "steam-path-browse",
      "output-log-dir-browse",
    ].map((id) => wrapper.get(`[data-testid="${id}"]`).text());
    expect(new Set(browseLabels).size).toBe(1);
    expect(browseLabels[0]).toBe("参照");
  });

  it("falls back to dev when getAppVersion fails", async () => {
    vi.mocked(App.getAppVersion).mockRejectedValueOnce(
      new Error("version unavailable"),
    );
    const wrapper = await mountAt("/settings/about");

    expect(
      wrapper.get('[data-testid="settings-app-version"]').text(),
    ).toContain("dev");
  });

  it("loads general controls without the OSS link", async () => {
    const wrapper = await mountAt("/settings/general");

    expect(App.getSuppressSleepWhileVRChat).toHaveBeenCalled();
    expect(App.getCloseToTray).toHaveBeenCalled();
    expect(
      wrapper
        .find(".power-switch")
        .findComponent({ name: "ElSwitch" })
        .props("modelValue"),
    ).toBe(true);
    expect(
      wrapper.find('[data-testid="settings-close-to-tray"]').exists(),
    ).toBe(true);
    expect(wrapper.find(".setting-row--fixed").exists()).toBe(true);
    expect(wrapper.find('[data-testid="settings-ui-language"]').exists()).toBe(
      true,
    );
    expect(
      wrapper.find('[data-testid="settings-oss-licenses-link"]').exists(),
    ).toBe(false);
  });

  it("loads data retention, gallery enrich, and storage note", async () => {
    const wrapper = await mountAt("/settings/data");

    expect(App.getLogRetentionDays).toHaveBeenCalled();
    expect(App.getGalleryAutoEnrichMetadata).toHaveBeenCalled();
    expect(
      wrapper.get('[data-testid="settings-app-data-storage"]').text(),
    ).toContain("vrchat-tweaker");
    expect(
      (
        wrapper.find(".setting-row .el-input-number input")
          .element as HTMLInputElement
      ).value,
    ).toBe("45");
    const retentionRow = wrapper
      .find('[data-testid="settings-log-retention"]')
      .element.closest(".setting-row");
    expect(retentionRow?.classList.contains("setting-row--number")).toBe(true);
    expect(retentionRow?.classList.contains("setting-row--fixed")).toBe(true);
    expect(retentionRow?.textContent).toContain("日");
    expect(retentionRow?.textContent).not.toContain("（日）");
    expect(
      wrapper.find('[data-testid="settings-gallery-auto-enrich"]').exists(),
    ).toBe(true);
  });

  it("loads about and reset sections on their routes", async () => {
    const about = await mountAt("/settings/about");
    expect(App.getAppVersion).toHaveBeenCalled();
    expect(App.getAppUpdateAutoCheck).toHaveBeenCalled();
    expect(about.get('[data-testid="settings-app-version"]').text()).toContain(
      "0.1.0",
    );
    expect(
      about
        .get('[data-testid="settings-app-version"]')
        .element.closest(".setting-row")
        ?.classList.contains("setting-row--fixed"),
    ).toBe(true);
    expect(
      about.get('[data-testid="settings-app-disclaimer"]').text(),
    ).toContain("非公式");
    expect(
      about.find('[data-testid="settings-check-app-update"]').exists(),
    ).toBe(true);
    expect(
      about.find('[data-testid="settings-oss-licenses-link"]').exists(),
    ).toBe(true);

    const reset = await mountAt("/settings/reset");
    expect(
      reset.find('[data-testid="settings-reset-setup-section"]').exists(),
    ).toBe(true);
    expect(
      reset.find('[data-testid="settings-open-onboarding"]').exists(),
    ).toBe(true);
    expect(reset.find('[data-testid="settings-vacuum"]').exists()).toBe(true);
    const danger = reset.get('[data-testid="settings-destructive-section"]');
    expect(danger.classes()).toContain("settings-danger-card");
    expect(danger.find('[data-testid="settings-vacuum"]').exists()).toBe(false);
    expect(
      danger.find('[data-testid="settings-clear-encounters"]').exists(),
    ).toBe(true);
    expect(
      danger.find('[data-testid="settings-clear-screenshots"]').exists(),
    ).toBe(true);
    expect(danger.find('[data-testid="settings-clear-friends"]').exists()).toBe(
      true,
    );
    expect(danger.text()).toContain("遭遇ログ");
    expect(danger.text()).toContain("user_encounters");
    expect(danger.text()).not.toContain("よろしいですか");
    expect(danger.text()).toContain("スクリーンショットインデックス");
    expect(danger.text()).not.toContain("スクショインデックス");
    expect(
      danger.get('[data-testid="settings-clear-screenshots"]').text(),
    ).toContain("スクリーンショットインデックスをクリア");
    const screenshotLabel = danger
      .get('[data-testid="settings-clear-screenshots"]')
      .element.closest(".setting-row")
      ?.querySelector(".setting-row__label");
    expect(screenshotLabel?.textContent).toBe("スクリーンショットインデックス");
    expect(
      danger.get('[data-testid="settings-clear-encounters"]').text(),
    ).toContain("クリア");
    const clearRow = danger
      .get('[data-testid="settings-clear-encounters"]')
      .element.closest(".setting-row");
    expect(clearRow?.classList.contains("setting-row--center")).toBe(false);
  });

  it("shows a failed update check under the button without asking the card to scroll sideways", async () => {
    const detail = `https://updates.example.invalid/${"x".repeat(180)}`;
    vi.mocked(App.checkForAppUpdate).mockRejectedValueOnce(new Error(detail));
    const wrapper = await mountAt("/settings/about");

    await wrapper
      .get('[data-testid="settings-check-app-update"]')
      .trigger("click");
    await flushPromises();

    const alert = wrapper.get('[data-testid="settings-app-update-error"]');
    expect(alert.text()).toContain("更新を確認できませんでした");
    expect(alert.text()).toContain("ネットワークを確認して");
    expect(alert.text()).toContain(detail);
    const row = wrapper
      .get('[data-testid="settings-check-app-update"]')
      .element.closest(".setting-row");
    expect(row?.contains(alert.element)).toBe(true);
    expect(row?.querySelector(".settings-update-actions")).not.toBeNull();
  });

  it("shows up to date after manual update check", async () => {
    const wrapper = await mountAt("/settings/about");
    await wrapper
      .get('[data-testid="settings-check-app-update"]')
      .trigger("click");
    await flushPromises();
    expect(App.checkForAppUpdate).toHaveBeenCalled();
    expect(
      wrapper.get('[data-testid="settings-app-update-status"]').text(),
    ).toContain("最新");
  });

  it("lists all UI language options", async () => {
    const wrapper = await mountAt("/settings/general");

    const options = wrapper.findAllComponents({ name: "ElOption" });
    expect(options).toHaveLength(5);
    expect(options.map((o) => o.props("value"))).toEqual([
      "ja",
      "en",
      "ko",
      "zh-TW",
      "zh-CN",
    ]);
  });

  it("saves path settings when a path input changes", async () => {
    const wrapper = await mountAt("/settings/data");
    vi.mocked(App.setPathSettings).mockClear();

    const input = pathRowInput(wrapper, 0);
    await input.setValue("D:\\New\\VRChat.exe");
    await input.trigger("change");
    await flushPromises();

    expect(App.setPathSettings).toHaveBeenCalledWith(
      expect.objectContaining({ vrchatPathWindows: "D:\\New\\VRChat.exe" }),
    );
  });

  it("browse VRChat path saves selected file", async () => {
    vi.mocked(App.openFileDialog).mockResolvedValueOnce(
      "D:\\Picked\\VRChat.exe",
    );
    const wrapper = await mountAt("/settings/data");
    vi.mocked(App.setPathSettings).mockClear();

    await wrapper.find('[data-testid="vrchat-path-browse"]').trigger("click");
    await flushPromises();

    expect(App.openFileDialog).toHaveBeenCalled();
    expect(App.setPathSettings).toHaveBeenCalledWith(
      expect.objectContaining({ vrchatPathWindows: "D:\\Picked\\VRChat.exe" }),
    );
  });

  it("browse Steam path saves selected file", async () => {
    vi.mocked(App.openFileDialog).mockResolvedValueOnce("/usr/bin/steam");
    const wrapper = await mountAt("/settings/data");
    vi.mocked(App.setPathSettings).mockClear();

    await wrapper.find('[data-testid="steam-path-browse"]').trigger("click");
    await flushPromises();

    expect(App.setPathSettings).toHaveBeenCalledWith(
      expect.objectContaining({ steamPathLinux: "/usr/bin/steam" }),
    );
  });

  it("browse output log directory saves selected folder", async () => {
    vi.mocked(App.openDirectoryDialog).mockResolvedValueOnce("C:\\logs");
    const wrapper = await mountAt("/settings/data");
    vi.mocked(App.setPathSettings).mockClear();

    await wrapper
      .find('[data-testid="output-log-dir-browse"]')
      .trigger("click");
    await flushPromises();

    expect(App.setPathSettings).toHaveBeenCalledWith(
      expect.objectContaining({ outputLogPath: "C:\\logs" }),
    );
  });

  it("does not show output log file browse button", async () => {
    const wrapper = await mountAt("/settings/data");
    expect(
      wrapper.find('[data-testid="output-log-path-browse"]').exists(),
    ).toBe(false);
  });

  it("opens VRChat log folder via App", async () => {
    const wrapper = await mountAt("/settings/data");

    const openLink = wrapper.get('[data-testid="settings-open-log-folder"]');
    expect(openLink.element.tagName).toBe("A");
    await openLink.trigger("click");
    await flushPromises();

    expect(App.openVRChatLogFolder).toHaveBeenCalled();
  });

  it("validates executable paths with App.validatePath", async () => {
    vi.mocked(App.validatePath).mockResolvedValueOnce(false);
    const wrapper = await mountAt("/settings/data");

    await validateExistsBtn(wrapper, 0).trigger("click");
    await flushPromises();

    expect(App.validatePath).toHaveBeenCalledWith(
      defaultPathSettings.vrchatPathWindows,
    );
    expect(wrapper.text()).toContain("存在しません");
  });

  it("validates output log path with App.validateOutputLogPath", async () => {
    vi.mocked(App.validateOutputLogPath).mockResolvedValueOnce(true);
    const wrapper = await mountAt("/settings/data");

    await validateExistsBtn(wrapper, 2).trigger("click");
    await flushPromises();

    expect(App.validateOutputLogPath).toHaveBeenCalledWith(
      defaultPathSettings.outputLogPath,
    );
    expect(wrapper.text()).toContain("存在します");
  });

  it("saves log retention days on change", async () => {
    const wrapper = await mountAt("/settings/data");
    vi.mocked(App.setLogRetentionDays).mockClear();

    const retentionInput = wrapper.find(".setting-row .el-input-number input");
    await retentionInput.setValue("60");
    await retentionInput.trigger("change");
    await flushPromises();

    expect(App.setLogRetentionDays).toHaveBeenCalledWith(60);
  });

  it("saves suppress sleep toggle on change", async () => {
    const wrapper = await mountAt("/settings/general");
    vi.mocked(App.setSuppressSleepWhileVRChat).mockClear();

    await wrapper.find(".power-switch").trigger("click");
    await flushPromises();

    expect(App.setSuppressSleepWhileVRChat).toHaveBeenCalledWith(false);
  });

  it("saves close to tray toggle on change", async () => {
    const wrapper = await mountAt("/settings/general");
    vi.mocked(App.setCloseToTray).mockClear();

    await wrapper
      .find('[data-testid="settings-close-to-tray"]')
      .trigger("click");
    await flushPromises();

    expect(App.setCloseToTray).toHaveBeenCalledWith(false);
  });

  it("runs vacuum DB after confirmation", async () => {
    vi.spyOn(ElMessageBox, "confirm").mockResolvedValue(undefined as never);
    const wrapper = await mountAt("/settings/reset");

    await wrapper.get('[data-testid="settings-vacuum"]').trigger("click");
    await flushPromises();

    expect(App.vacuumDb).toHaveBeenCalled();
    expect(showToast.success).toHaveBeenCalled();
  });

  it("clears encounters after confirmation", async () => {
    vi.spyOn(ElMessageBox, "confirm").mockResolvedValue(undefined as never);
    const wrapper = await mountAt("/settings/reset");

    await wrapper
      .get('[data-testid="settings-clear-encounters"]')
      .trigger("click");
    await flushPromises();

    expect(App.clearEncounters).toHaveBeenCalled();
  });

  it("clears screenshots after confirmation", async () => {
    vi.spyOn(ElMessageBox, "confirm").mockResolvedValue(undefined as never);
    const wrapper = await mountAt("/settings/reset");

    await wrapper
      .get('[data-testid="settings-clear-screenshots"]')
      .trigger("click");
    await flushPromises();

    expect(ElMessageBox.confirm).toHaveBeenCalledWith(
      "スクリーンショットインデックス（screenshots）をすべて削除します。よろしいですか？",
      "確認",
      expect.any(Object),
    );
    expect(App.clearScreenshots).toHaveBeenCalled();
    expect(showToast.success).toHaveBeenCalledWith(
      "3件のスクリーンショットインデックスを削除しました",
    );
  });

  it("clears friends cache after confirmation", async () => {
    vi.spyOn(ElMessageBox, "confirm").mockResolvedValue(undefined as never);
    const wrapper = await mountAt("/settings/reset");

    await wrapper
      .get('[data-testid="settings-clear-friends"]')
      .trigger("click");
    await flushPromises();

    expect(App.clearFriendsCache).toHaveBeenCalled();
  });

  it("skips maintenance when confirmation is cancelled", async () => {
    vi.spyOn(ElMessageBox, "confirm").mockRejectedValue(new Error("cancel"));
    const wrapper = await mountAt("/settings/reset");

    await wrapper.get('[data-testid="settings-vacuum"]').trigger("click");
    await flushPromises();

    expect(App.vacuumDb).not.toHaveBeenCalled();
  });

  it("shows maintenance error when operation fails", async () => {
    vi.spyOn(ElMessageBox, "confirm").mockResolvedValue(undefined as never);
    vi.mocked(App.vacuumDb).mockRejectedValueOnce(new Error("vacuum failed"));
    const wrapper = await mountAt("/settings/reset");

    await wrapper.get('[data-testid="settings-vacuum"]').trigger("click");
    await flushPromises();

    expect(wrapper.find(".el-alert--error").text()).toContain("vacuum failed");
  });

  it("handles isLoggedIn failure on mount", async () => {
    vi.mocked(App.isLoggedIn).mockRejectedValueOnce(new Error("backend down"));
    const wrapper = mountSettings();
    await flushPromises();

    expect(wrapper.find("#login-username").exists()).toBe(true);
  });

  it("logs error when openVRChatLogFolder fails", async () => {
    vi.mocked(App.openVRChatLogFolder).mockRejectedValueOnce(
      new Error("open failed"),
    );
    const consoleSpy = vi
      .spyOn(console, "error")
      .mockImplementation(() => undefined);
    const wrapper = await mountAt("/settings/data");

    await wrapper
      .get('[data-testid="settings-open-log-folder"]')
      .trigger("click");
    await flushPromises();

    expect(consoleSpy).toHaveBeenCalled();
  });
});

describe("SettingsView login and profile", () => {
  const sampleUser = {
    vrcUserId: "usr_abc",
    displayName: "Test User",
    username: "testuser",
    status: "active",
    statusDescription: "Hello",
    state: "online",
    isFavorite: false,
    lastUpdated: "2025-01-01T00:00:00Z",
    currentAvatarThumbnailImageUrl: "https://example.com/avatar.png",
    userIcon: "",
    profilePicOverrideThumbnail: "",
  };

  beforeEach(async () => {
    resetSessionUnlockForStorybook();
    await router.push("/settings");
    await router.isReady();
    setupAppMocks();
    vi.spyOn(App, "login").mockResolvedValue({ ok: false, error: "unused" });
    vi.spyOn(App, "logout").mockResolvedValue(undefined);
    vi.spyOn(App, "refreshFriends").mockResolvedValue(undefined);
    vi.spyOn(App, "clearStoredCredential").mockResolvedValue(undefined);
  });

  afterEach(() => {
    activeWrapper?.unmount();
    activeWrapper = undefined;
    vi.restoreAllMocks();
  });

  function mountLoggedIn(user = sampleUser) {
    vi.mocked(App.isLoggedIn).mockResolvedValue(true);
    vi.spyOn(App, "getSelfProfile").mockResolvedValue(user);
    return mountSettings();
  }

  it("shows current user profile when logged in", async () => {
    const wrapper = mountLoggedIn();
    await flushPromises();

    expect(wrapper.text()).toContain("Test User");
    expect(wrapper.text()).toContain("@testuser");
    expect(wrapper.text()).toContain("usr_abc");
    expect(wrapper.find(".current-user-avatar").attributes("src")).toBe(
      "https://example.com/avatar.png",
    );
    expect(
      wrapper.find('[data-testid="settings-view-self-profile"]').text(),
    ).toBe("詳細を見る");
  });

  it("shows profile error when getSelfProfile fails", async () => {
    vi.mocked(App.isLoggedIn).mockResolvedValue(true);
    vi.spyOn(App, "getSelfProfile").mockRejectedValue(
      new Error("profile unavailable"),
    );
    const wrapper = mountSettings();
    await flushPromises();

    expect(wrapper.find(".login-status .el-alert--error").text()).toContain(
      "profile unavailable",
    );
  });

  it("refreshes profile when refresh button is clicked", async () => {
    const getUser = vi
      .spyOn(App, "getSelfProfile")
      .mockResolvedValue(sampleUser);
    const wrapper = mountLoggedIn();
    await flushPromises();
    getUser.mockClear();

    const refreshBtn = wrapper
      .findAll(".login-actions button")
      .find((b) => b.text().includes("プロフィール再取得"));
    await refreshBtn!.trigger("click");
    await flushPromises();

    expect(getUser).toHaveBeenCalledWith(true);
  });

  it("logs in successfully and loads profile", async () => {
    vi.mocked(App.login).mockResolvedValue({ ok: true });
    vi.spyOn(App, "getSelfProfile").mockResolvedValue(sampleUser);
    const wrapper = mountSettings();
    await flushPromises();

    await wrapper.find("#login-username").setValue("user");
    await wrapper.find("#login-password").setValue("pass");
    await wrapper.find("#login-2fa").setValue("123456");
    await flushPromises();

    expect(App.login).toHaveBeenCalledWith("user", "pass", "123456");
    expect(wrapper.text()).toContain("Test User");
  });

  it("shows login error when credentials are rejected", async () => {
    vi.mocked(App.login).mockResolvedValue({
      ok: false,
      error: "invalid credentials",
    });
    const wrapper = mountSettings();
    await flushPromises();

    await wrapper.find("#login-username").setValue("user");
    await wrapper.find("#login-password").setValue("wrong");
    await wrapper
      .findAll("button")
      .find((b) => b.text().includes("ログイン") && !b.text().includes("中"))!
      .trigger("click");
    await flushPromises();

    expect(wrapper.find(".login-form .el-alert--error").text()).toContain(
      "invalid credentials",
    );
  });

  it("logs out and returns to login form", async () => {
    const wrapper = mountLoggedIn();
    await flushPromises();

    await wrapper
      .findAll("button")
      .find((b) => b.text().includes("ログアウト"))!
      .trigger("click");
    await flushPromises();

    expect(App.logout).toHaveBeenCalled();
    expect(App.clearStoredCredential).not.toHaveBeenCalled();
    expect(wrapper.find("#login-username").exists()).toBe(true);
  });

  it("shows logout error after failed backend logout", async () => {
    vi.mocked(App.logout).mockRejectedValueOnce(new Error("logout failed"));
    const wrapper = mountLoggedIn();
    await flushPromises();

    await wrapper
      .findAll("button")
      .find((b) => b.text().includes("ログアウト"))!
      .trigger("click");
    await flushPromises();

    expect(wrapper.find("#login-username").exists()).toBe(true);
    expect(wrapper.find(".login-form .el-alert--error").text()).toContain(
      "logout failed",
    );
    expect(App.clearStoredCredential).toHaveBeenCalledTimes(1);
  });

  it("calls refreshFriends from logged-in actions", async () => {
    const wrapper = mountLoggedIn();
    await flushPromises();

    await wrapper
      .findAll("button")
      .find((b) => b.text().includes("フレンド一覧を更新"))!
      .trigger("click");
    await flushPromises();

    expect(App.refreshFriends).toHaveBeenCalled();
  });

  it("handles refreshFriends failure", async () => {
    vi.mocked(App.refreshFriends).mockRejectedValueOnce(
      new Error("friends refresh failed"),
    );
    const wrapper = mountLoggedIn();
    await flushPromises();

    await wrapper
      .findAll("button")
      .find((b) => b.text().includes("フレンド一覧を更新"))!
      .trigger("click");
    await flushPromises();

    expect(App.refreshFriends).toHaveBeenCalled();
  });
});

describe("SettingsView onLanguageChange", () => {
  beforeEach(async () => {
    resetSessionUnlockForStorybook();
    await router.push("/settings");
    await router.isReady();
    setupAppMocks();
    vi.spyOn(I18nModule, "setLanguage");
  });

  afterEach(() => {
    activeWrapper?.unmount();
    activeWrapper = undefined;
    vi.restoreAllMocks();
  });

  it("calls i18n setLanguage after App.setLanguage succeeds", async () => {
    const wrapper = await mountAt("/settings/general");

    const select = wrapper.findComponent({ name: "ElSelect" });
    await select.vm.$emit("update:modelValue", "en");
    await flushPromises();

    expect(App.setLanguage).toHaveBeenCalledWith("en");
    expect(I18nModule.setLanguage).toHaveBeenCalledWith("en");
  });

  it("does not call i18n setLanguage when App.setLanguage fails", async () => {
    vi.mocked(App.setLanguage).mockRejectedValueOnce(new Error("save failed"));

    const wrapper = await mountAt("/settings/general");

    vi.mocked(I18nModule.setLanguage).mockClear();

    const select = wrapper.findComponent({ name: "ElSelect" });
    await select.vm.$emit("update:modelValue", "en");
    await flushPromises();

    expect(I18nModule.setLanguage).not.toHaveBeenCalled();
    expect(showToast.error).toHaveBeenCalledWith("save failed");
  });

  it("ignores invalid locale values", async () => {
    const wrapper = await mountAt("/settings/general");

    const select = wrapper.findComponent({ name: "ElSelect" });
    await select.vm.$emit("update:modelValue", "invalid-locale");
    await flushPromises();

    expect(App.setLanguage).not.toHaveBeenCalled();
    expect(I18nModule.setLanguage).not.toHaveBeenCalled();
  });
});
