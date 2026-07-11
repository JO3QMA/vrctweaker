import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { ElMessage, ElMessageBox } from "element-plus";
import { createRouter, createWebHashHistory } from "vue-router";
import SettingsView from "../SettingsView.vue";
import { App } from "../../wails/app";
import * as I18nModule from "../../i18n";
import { resetSessionUnlockForStorybook } from "../../composables/useSessionUnlock";

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: "/settings", component: SettingsView },
    { path: "/me", name: "me", component: { template: "<div>Me</div>" } },
    { path: "/licenses", component: { template: "<div>Licenses</div>" } },
  ],
});

const defaultPathSettings = {
  vrchatPathWindows: "C:\\VRChat\\VRChat.exe",
  steamPathLinux: "/home/user/.steam/steam",
  outputLogPath:
    "C:\\Users\\x\\AppData\\LocalLow\\VRChat\\VRChat\\output_log.txt",
};

function mountSettings() {
  return mount(SettingsView, { global: { plugins: [router] } });
}

function pathRowInput(wrapper: ReturnType<typeof mount>, rowIndex: number) {
  return wrapper.findAll(".path-row .el-input input")[rowIndex]!;
}

function validateExistsBtn(
  wrapper: ReturnType<typeof mount>,
  rowIndex: number,
) {
  return wrapper.findAll(".path-row .el-button--primary")[rowIndex]!;
}

function setupAppMocks() {
  vi.spyOn(App, "hasStoredCredential").mockResolvedValue(false);
  vi.spyOn(App, "isLoggedIn").mockResolvedValue(false);
  vi.spyOn(App, "getLogRetentionDays").mockResolvedValue(45);
  vi.spyOn(App, "getSuppressSleepWhileVRChat").mockResolvedValue(true);
  vi.spyOn(App, "getPathSettings").mockResolvedValue({
    ...defaultPathSettings,
  });
  vi.spyOn(App, "setPathSettings").mockResolvedValue(undefined);
  vi.spyOn(App, "setLogRetentionDays").mockResolvedValue(undefined);
  vi.spyOn(App, "setSuppressSleepWhileVRChat").mockResolvedValue(undefined);
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
    vi.restoreAllMocks();
  });

  it("renders settings title", async () => {
    const wrapper = mountSettings();
    await flushPromises();
    expect(wrapper.find(".page-title").text()).toBe("設定");
  });

  it("has link to OSS licenses page", async () => {
    const wrapper = mountSettings();
    await flushPromises();
    const licensesLink = wrapper.find(".btn-licenses");
    expect(licensesLink.exists()).toBe(true);
    expect(licensesLink.attributes("href")).toContain("/licenses");
    expect(licensesLink.text()).toContain("OSS ライセンス一覧");
  });

  it("loads path settings on mount", async () => {
    const wrapper = mountSettings();
    await flushPromises();

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
  });

  it("loads log retention and suppress sleep on mount", async () => {
    const wrapper = mountSettings();
    await flushPromises();

    expect(App.getLogRetentionDays).toHaveBeenCalled();
    expect(App.getSuppressSleepWhileVRChat).toHaveBeenCalled();
    expect(
      (
        wrapper.find(".setting-row .el-input-number input")
          .element as HTMLInputElement
      ).value,
    ).toBe("45");
    expect(
      wrapper.findComponent({ name: "ElSwitch" }).props("modelValue"),
    ).toBe(true);
  });

  it("lists all UI language options", async () => {
    const wrapper = mountSettings();
    await flushPromises();

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
    const wrapper = mountSettings();
    await flushPromises();
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
    const wrapper = mountSettings();
    await flushPromises();
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
    const wrapper = mountSettings();
    await flushPromises();
    vi.mocked(App.setPathSettings).mockClear();

    await wrapper.find('[data-testid="steam-path-browse"]').trigger("click");
    await flushPromises();

    expect(App.setPathSettings).toHaveBeenCalledWith(
      expect.objectContaining({ steamPathLinux: "/usr/bin/steam" }),
    );
  });

  it("browse output log directory saves selected folder", async () => {
    vi.mocked(App.openDirectoryDialog).mockResolvedValueOnce("C:\\logs");
    const wrapper = mountSettings();
    await flushPromises();
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
    const wrapper = mountSettings();
    await flushPromises();
    expect(
      wrapper.find('[data-testid="output-log-path-browse"]').exists(),
    ).toBe(false);
  });

  it("opens VRChat log folder via App", async () => {
    const wrapper = mountSettings();
    await flushPromises();

    const openBtn = wrapper
      .findAll(".path-row")[2]!
      .findAll("button")
      .find((b: { text: () => string }) =>
        b.text().includes("ログフォルダを開く"),
      );
    expect(openBtn).toBeDefined();
    await openBtn!.trigger("click");
    await flushPromises();

    expect(App.openVRChatLogFolder).toHaveBeenCalled();
  });

  it("validates executable paths with App.validatePath", async () => {
    vi.mocked(App.validatePath).mockResolvedValueOnce(false);
    const wrapper = mountSettings();
    await flushPromises();

    await validateExistsBtn(wrapper, 0).trigger("click");
    await flushPromises();

    expect(App.validatePath).toHaveBeenCalledWith(
      defaultPathSettings.vrchatPathWindows,
    );
    expect(wrapper.text()).toContain("存在しません");
  });

  it("validates output log path with App.validateOutputLogPath", async () => {
    vi.mocked(App.validateOutputLogPath).mockResolvedValueOnce(true);
    const wrapper = mountSettings();
    await flushPromises();

    await validateExistsBtn(wrapper, 2).trigger("click");
    await flushPromises();

    expect(App.validateOutputLogPath).toHaveBeenCalledWith(
      defaultPathSettings.outputLogPath,
    );
    expect(wrapper.text()).toContain("存在します");
  });

  it("saves log retention days on change", async () => {
    const wrapper = mountSettings();
    await flushPromises();
    vi.mocked(App.setLogRetentionDays).mockClear();

    const retentionInput = wrapper.find(".setting-row .el-input-number input");
    await retentionInput.setValue("60");
    await retentionInput.trigger("change");
    await flushPromises();

    expect(App.setLogRetentionDays).toHaveBeenCalledWith(60);
  });

  it("saves suppress sleep toggle on change", async () => {
    const wrapper = mountSettings();
    await flushPromises();
    vi.mocked(App.setSuppressSleepWhileVRChat).mockClear();

    await wrapper.find(".power-switch").trigger("click");
    await flushPromises();

    expect(App.setSuppressSleepWhileVRChat).toHaveBeenCalledWith(false);
  });

  it("runs vacuum DB after confirmation", async () => {
    vi.spyOn(ElMessageBox, "confirm").mockResolvedValue(undefined as never);
    const successSpy = vi
      .spyOn(ElMessage, "success")
      .mockImplementation(() => ({
        close: () => {},
      }));
    const wrapper = mountSettings();
    await flushPromises();

    const vacuumBtn = wrapper
      .findAll(".maintenance-actions button")
      .find((b) => b.text().includes("DB最適化"));
    await vacuumBtn!.trigger("click");
    await flushPromises();

    expect(App.vacuumDb).toHaveBeenCalled();
    expect(successSpy).toHaveBeenCalled();
  });

  it("clears encounters after confirmation", async () => {
    vi.spyOn(ElMessageBox, "confirm").mockResolvedValue(undefined as never);
    vi.spyOn(ElMessage, "success").mockImplementation(() => ({
      close: () => {},
    }));
    const wrapper = mountSettings();
    await flushPromises();

    const btn = wrapper
      .findAll(".maintenance-actions button")
      .find((b) => b.text().includes("遭遇ログ"));
    await btn!.trigger("click");
    await flushPromises();

    expect(App.clearEncounters).toHaveBeenCalled();
  });

  it("clears screenshots after confirmation", async () => {
    vi.spyOn(ElMessageBox, "confirm").mockResolvedValue(undefined as never);
    vi.spyOn(ElMessage, "success").mockImplementation(() => ({
      close: () => {},
    }));
    const wrapper = mountSettings();
    await flushPromises();

    const btn = wrapper
      .findAll(".maintenance-actions button")
      .find((b) => b.text().includes("スクショ"));
    await btn!.trigger("click");
    await flushPromises();

    expect(App.clearScreenshots).toHaveBeenCalled();
  });

  it("clears friends cache after confirmation", async () => {
    vi.spyOn(ElMessageBox, "confirm").mockResolvedValue(undefined as never);
    vi.spyOn(ElMessage, "success").mockImplementation(() => ({
      close: () => {},
    }));
    const wrapper = mountSettings();
    await flushPromises();

    const btn = wrapper
      .findAll(".maintenance-actions button")
      .find((b) => b.text().includes("フレンド"));
    await btn!.trigger("click");
    await flushPromises();

    expect(App.clearFriendsCache).toHaveBeenCalled();
  });

  it("skips maintenance when confirmation is cancelled", async () => {
    vi.spyOn(ElMessageBox, "confirm").mockRejectedValue(new Error("cancel"));
    const wrapper = mountSettings();
    await flushPromises();

    const vacuumBtn = wrapper
      .findAll(".maintenance-actions button")
      .find((b) => b.text().includes("DB最適化"));
    await vacuumBtn!.trigger("click");
    await flushPromises();

    expect(App.vacuumDb).not.toHaveBeenCalled();
  });

  it("shows maintenance error when operation fails", async () => {
    vi.spyOn(ElMessageBox, "confirm").mockResolvedValue(undefined as never);
    vi.mocked(App.vacuumDb).mockRejectedValueOnce(new Error("vacuum failed"));
    const wrapper = mountSettings();
    await flushPromises();

    const vacuumBtn = wrapper
      .findAll(".maintenance-actions button")
      .find((b) => b.text().includes("DB最適化"));
    await vacuumBtn!.trigger("click");
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
    const wrapper = mountSettings();
    await flushPromises();

    const openBtn = wrapper
      .findAll(".path-row")[2]!
      .findAll("button")
      .find((b: { text: () => string }) =>
        b.text().includes("ログフォルダを開く"),
      );
    await openBtn!.trigger("click");
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
    await wrapper
      .findAll("button")
      .find((b) => b.text().includes("ログイン") && !b.text().includes("中"))!
      .trigger("click");
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
    vi.restoreAllMocks();
  });

  it("calls i18n setLanguage after App.setLanguage succeeds", async () => {
    const wrapper = mountSettings();
    await flushPromises();

    const select = wrapper.findComponent({ name: "ElSelect" });
    await select.vm.$emit("update:modelValue", "en");
    await flushPromises();

    expect(App.setLanguage).toHaveBeenCalledWith("en");
    expect(I18nModule.setLanguage).toHaveBeenCalledWith("en");
  });

  it("does not call i18n setLanguage when App.setLanguage fails", async () => {
    vi.mocked(App.setLanguage).mockRejectedValueOnce(new Error("save failed"));
    const elErrorSpy = vi.spyOn(ElMessage, "error").mockImplementation(() => ({
      close: () => {},
    }));

    const wrapper = mountSettings();
    await flushPromises();

    vi.mocked(I18nModule.setLanguage).mockClear();

    const select = wrapper.findComponent({ name: "ElSelect" });
    await select.vm.$emit("update:modelValue", "en");
    await flushPromises();

    expect(I18nModule.setLanguage).not.toHaveBeenCalled();
    expect(elErrorSpy).toHaveBeenCalledWith("save failed");
  });

  it("ignores invalid locale values", async () => {
    const wrapper = mountSettings();
    await flushPromises();

    const select = wrapper.findComponent({ name: "ElSelect" });
    await select.vm.$emit("update:modelValue", "invalid-locale");
    await flushPromises();

    expect(App.setLanguage).not.toHaveBeenCalled();
    expect(I18nModule.setLanguage).not.toHaveBeenCalled();
  });
});
