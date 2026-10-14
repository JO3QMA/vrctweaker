import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import OnboardingWizard from "../OnboardingWizard.vue";
import { App } from "@/wails/app";
import { resetSessionUnlockForStorybook } from "@/composables/useSessionUnlock";

vi.mock("@/wails/app", () => ({
  App: {
    isLoggedIn: vi.fn(),
    getSelfProfile: vi.fn(),
    login: vi.fn(),
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
    vi.mocked(App.getOnboardingFolderDefaults).mockResolvedValue({
      vrchatPathWindows: "C:\\VRChat\\launch.exe",
      steamPathLinux: "/usr/bin/steam",
      outputLogPath: "C:\\logs",
      pictureOutputFolder: "C:\\Pictures\\VRChat",
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

  it("completes wizard and emits finished", async () => {
    const wrapper = mount(OnboardingWizard);
    await flushPromises();
    await wrapper.find('[data-testid="onboarding-skip-step"]').trigger("click");
    await flushPromises();
    await wrapper.find('[data-testid="onboarding-next"]').trigger("click");
    await flushPromises();
    expect(App.saveOnboardingFolderSettings).toHaveBeenCalled();
    await wrapper.find('[data-testid="onboarding-next"]').trigger("click");
    await flushPromises();
    expect(App.setFirstLaunchGuideCompleted).toHaveBeenCalledWith(true);
    expect(wrapper.emitted("finished")).toBeTruthy();
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
