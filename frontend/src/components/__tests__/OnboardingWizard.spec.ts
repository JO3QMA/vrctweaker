import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import OnboardingWizard from "../OnboardingWizard.vue";
import { App } from "@/wails/app";
import { resetSessionUnlockForStorybook } from "@/composables/useSessionUnlock";
import { LOGIN_TWO_FACTOR_REQUIRED_MARKER } from "@/utils/vrchatLoginResult";

vi.mock("@/wails/app", () => ({
  App: {
    isLoggedIn: vi.fn(),
    getSelfProfile: vi.fn(),
    login: vi.fn(),
    getPathSettings: vi.fn(),
    getOnboardingFolderDefaults: vi.fn(),
    saveOnboardingFolderSettings: vi.fn(),
    setFirstLaunchGuideCompleted: vi.fn(),
    openDirectoryDialog: vi.fn(),
    openFileDialog: vi.fn(),
  },
}));

describe("OnboardingWizard", () => {
  beforeEach(() => {
    resetSessionUnlockForStorybook();
    vi.mocked(App.isLoggedIn).mockResolvedValue(false);
    vi.mocked(App.getSelfProfile).mockResolvedValue({
      displayName: "Tester",
      vrcUserId: "usr_test",
    } as never);
    vi.mocked(App.getPathSettings).mockResolvedValue({
      vrchatPathWindows: "",
      steamPathLinux: "",
      outputLogPath: "",
    });
    vi.mocked(App.getOnboardingFolderDefaults).mockResolvedValue({
      vrchatPathWindows: "C:\\VRChat\\launch.exe",
      steamPathLinux: "/usr/bin/steam",
      outputLogPath: "C:\\logs",
      pictureOutputFolder: "C:\\Pictures\\VRChat",
      savedPictureOutputFolder: "",
    });
    vi.mocked(App.setFirstLaunchGuideCompleted).mockResolvedValue(undefined);
    vi.mocked(App.saveOnboardingFolderSettings).mockResolvedValue(undefined);
  });

  it("shows login step with keyring hint", async () => {
    const wrapper = mount(OnboardingWizard);
    await flushPromises();
    expect(wrapper.find('[data-testid="onboarding-step-login"]').exists()).toBe(
      true,
    );
    expect(wrapper.text()).toContain("キーリング");
  });

  it("advances to folder step on skip login", async () => {
    const wrapper = mount(OnboardingWizard);
    await flushPromises();
    await wrapper.find('[data-testid="onboarding-skip-step"]').trigger("click");
    expect(
      wrapper.find('[data-testid="onboarding-step-folders"]').exists(),
    ).toBe(true);
  });

  it("lists folder rows with change buttons", async () => {
    const wrapper = mount(OnboardingWizard);
    await flushPromises();
    await wrapper.find('[data-testid="onboarding-skip-step"]').trigger("click");
    await flushPromises();
    expect(
      wrapper
        .find('[data-testid="onboarding-folder-change-outputLogPath"]')
        .exists(),
    ).toBe(true);
    expect(wrapper.text()).toContain("C:\\logs");
  });

  it("completes wizard and saves auto-detected folders", async () => {
    const wrapper = mount(OnboardingWizard);
    await flushPromises();
    await wrapper.find('[data-testid="onboarding-skip-step"]').trigger("click");
    await flushPromises();
    await wrapper.find('[data-testid="onboarding-next"]').trigger("click");
    await flushPromises();
    expect(App.saveOnboardingFolderSettings).toHaveBeenCalledWith({
      vrchatPathWindows: "C:\\VRChat\\launch.exe",
      steamPathLinux: "/usr/bin/steam",
      outputLogPath: "C:\\logs",
      pictureOutputFolder: "C:\\Pictures\\VRChat",
      applyVrchatPathWindows: true,
      applySteamPathLinux: true,
      applyOutputLogPath: true,
      applyPictureOutputFolder: true,
    });
    await wrapper.find('[data-testid="onboarding-next"]').trigger("click");
    await flushPromises();
    expect(App.setFirstLaunchGuideCompleted).toHaveBeenCalledWith(true);
    expect(wrapper.emitted("finished")).toBeTruthy();
  });

  it("blocks folder step when defaults fail to load", async () => {
    vi.useFakeTimers();
    vi.mocked(App.getOnboardingFolderDefaults).mockRejectedValue(
      new Error("backend down"),
    );
    const wrapper = mount(OnboardingWizard);
    await flushPromises();
    await vi.runAllTimersAsync();
    await flushPromises();
    await wrapper.find('[data-testid="onboarding-skip-step"]').trigger("click");
    await flushPromises();
    await wrapper.find('[data-testid="onboarding-next"]').trigger("click");
    await vi.runAllTimersAsync();
    await flushPromises();
    expect(App.saveOnboardingFolderSettings).not.toHaveBeenCalled();
    expect(
      wrapper.find('[data-testid="onboarding-step-folders"]').exists(),
    ).toBe(true);
    vi.useRealTimers();
  });

  it("does not emit finished when completion flag save fails", async () => {
    vi.mocked(App.setFirstLaunchGuideCompleted).mockRejectedValue(
      new Error("disk full"),
    );
    const wrapper = mount(OnboardingWizard);
    await flushPromises();
    await wrapper.find('[data-testid="onboarding-skip-step"]').trigger("click");
    await flushPromises();
    await wrapper.find('[data-testid="onboarding-next"]').trigger("click");
    await flushPromises();
    await wrapper.find('[data-testid="onboarding-next"]').trigger("click");
    await flushPromises();
    expect(wrapper.emitted("finished")).toBeFalsy();
    expect(
      wrapper.find('[data-testid="onboarding-complete-error"]').text(),
    ).toContain("disk full");
  });

  it("shows login error below fields and keeps username", async () => {
    vi.mocked(App.login).mockResolvedValue({
      ok: false,
      error: "invalid credentials",
    });
    const wrapper = mount(OnboardingWizard);
    await flushPromises();
    await wrapper
      .find('[data-testid="onboarding-login-username"]')
      .setValue("user1");
    await wrapper
      .find('[data-testid="onboarding-login-password"]')
      .setValue("secret");
    await wrapper
      .find('[data-testid="onboarding-login-submit"]')
      .trigger("click");
    await flushPromises();
    expect(
      wrapper.find('[data-testid="onboarding-login-error"]').text(),
    ).toContain("invalid credentials");
    expect(App.login).toHaveBeenCalledWith("user1", "secret", undefined);
  });

  it("uses short keyring hint copy", async () => {
    const wrapper = mount(OnboardingWizard);
    await flushPromises();
    expect(wrapper.text()).toContain(
      "認証情報はOSのキーリングに保存されます。",
    );
    expect(wrapper.text()).not.toContain("平文");
  });

  it("auto-submits when six-digit 2FA code is entered", async () => {
    vi.mocked(App.login)
      .mockResolvedValueOnce({
        ok: false,
        error: LOGIN_TWO_FACTOR_REQUIRED_MARKER,
      })
      .mockResolvedValueOnce({ ok: true });
    const wrapper = mount(OnboardingWizard);
    await flushPromises();
    await wrapper
      .find('[data-testid="onboarding-login-username"]')
      .setValue("user1");
    await wrapper
      .find('[data-testid="onboarding-login-password"]')
      .setValue("secret");
    await wrapper
      .find('[data-testid="onboarding-login-submit"]')
      .trigger("click");
    await flushPromises();
    await wrapper
      .find('[data-testid="onboarding-login-2fa"]')
      .setValue("123456");
    await flushPromises();
    expect(App.login).toHaveBeenCalledTimes(2);
    expect(App.login).toHaveBeenLastCalledWith("user1", "secret", "123456");
  });

  it("reveals 2FA field without error when required", async () => {
    vi.mocked(App.login).mockResolvedValue({
      ok: false,
      error: LOGIN_TWO_FACTOR_REQUIRED_MARKER,
    });
    const wrapper = mount(OnboardingWizard);
    await flushPromises();
    await wrapper
      .find('[data-testid="onboarding-login-username"]')
      .setValue("user1");
    await wrapper
      .find('[data-testid="onboarding-login-password"]')
      .setValue("secret");
    await wrapper
      .find('[data-testid="onboarding-login-submit"]')
      .trigger("click");
    await flushPromises();
    expect(
      wrapper.find('[data-testid="onboarding-login-2fa-field"]').exists(),
    ).toBe(true);
    expect(
      wrapper.find('[data-testid="onboarding-login-error"]').exists(),
    ).toBe(false);
  });

  it("skip all finishes without saving folders", async () => {
    const wrapper = mount(OnboardingWizard);
    await flushPromises();
    await wrapper.find('[data-testid="onboarding-skip-all"]').trigger("click");
    await flushPromises();
    expect(App.saveOnboardingFolderSettings).not.toHaveBeenCalled();
    expect(App.setFirstLaunchGuideCompleted).toHaveBeenCalledWith(true);
  });
});
