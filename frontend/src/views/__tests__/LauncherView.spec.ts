import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import LauncherView from "../LauncherView.vue";
import type { LaunchProfileDTO, LaunchArgsParsedDTO } from "../../wails/app";

const {
  mockLaunchProfiles,
  mockParseLaunchArgsForGUI,
  mockMergeLaunchArgsForGUI,
  mockSaveLaunchProfile,
  mockDeleteLaunchProfile,
  mockLaunchVRChatWithArgs,
} = vi.hoisted(() => ({
  mockLaunchProfiles: vi.fn(),
  mockParseLaunchArgsForGUI: vi.fn(),
  mockMergeLaunchArgsForGUI: vi.fn(),
  mockSaveLaunchProfile: vi.fn(),
  mockDeleteLaunchProfile: vi.fn(),
  mockLaunchVRChatWithArgs: vi.fn(),
}));

vi.mock("../../wails/app", () => ({
  App: {
    launchProfiles: mockLaunchProfiles,
    parseLaunchArgsForGUI: mockParseLaunchArgsForGUI,
    mergeLaunchArgsForGUI: mockMergeLaunchArgsForGUI,
    saveLaunchProfile: mockSaveLaunchProfile,
    deleteLaunchProfile: mockDeleteLaunchProfile,
    launchVRChatWithArgs: mockLaunchVRChatWithArgs,
  },
}));

const sampleProfiles: LaunchProfileDTO[] = [
  {
    id: "1",
    name: "Default",
    arguments: "--no-vr",
    isDefault: true,
  },
  {
    id: "2",
    name: "With cache clear",
    arguments: "--no-vr --clear-cache -screen-fullscreen 1",
    isDefault: false,
  },
];

describe("LauncherView", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockLaunchProfiles.mockResolvedValue([...sampleProfiles]);
    mockParseLaunchArgsForGUI.mockImplementation(
      async (args: string): Promise<LaunchArgsParsedDTO> => {
        const base = {
          fpfc: false,
          screenWidth: 0,
          screenHeight: 0,
          fps: 0,
          safe: false,
          noSplash: false,
          noAudio: false,
          skipRegistry: false,
          forceD3d11: false,
          forceVulkan: false,
          log: false,
          processPriority: 0,
        };
        let vrMode: "" | "desktop" | "vr" = "";
        if (args.includes("--no-vr") || args.includes("-no-vr"))
          vrMode = "desktop";
        else if (args.includes("-vr")) vrMode = "vr";

        return {
          ...base,
          vrMode,
          clearCache: args.includes("--clear-cache"),
          screenMode: args.includes("-screen-fullscreen 1")
            ? "fullscreen"
            : args.includes("-popupwindow")
              ? "popupwindow"
              : args.includes("-windowed")
                ? "windowed"
                : "",
          custom: args.includes("-batchmode") ? "-batchmode" : "",
        };
      },
    );
    mockMergeLaunchArgsForGUI.mockImplementation(
      async (dto: LaunchArgsParsedDTO): Promise<string> => {
        const parts: string[] = [];
        if (dto.vrMode === "desktop") parts.push("-no-vr");
        if (dto.vrMode === "vr") parts.push("-vr");
        if (dto.clearCache) parts.push("--clear-cache");
        if (dto.screenMode === "fullscreen") parts.push("-screen-fullscreen 1");
        if (dto.screenMode === "windowed") parts.push("-windowed");
        if (dto.screenMode === "popupwindow") parts.push("-popupwindow");
        if (dto.fpfc) parts.push("-fpfc");
        if (dto.screenWidth)
          parts.push("-screen-width", String(dto.screenWidth));
        if (dto.screenHeight)
          parts.push("-screen-height", String(dto.screenHeight));
        if (dto.fps) parts.push(`--fps=${dto.fps}`);
        if (dto.safe) parts.push("-safe");
        if (dto.noSplash) parts.push("-nosplash");
        if (dto.noAudio) parts.push("-noaudio");
        if (dto.skipRegistry) parts.push("--skip-registry-install");
        if (dto.forceD3d11) parts.push("-force-d3d11");
        if (dto.forceVulkan) parts.push("-force-vulkan");
        if (dto.log) parts.push("-log");
        if (dto.processPriority)
          parts.push(`--process-priority=${dto.processPriority}`);
        if (dto.custom) parts.push(dto.custom);
        return parts.join(" ");
      },
    );
    mockSaveLaunchProfile.mockResolvedValue(undefined);
    mockDeleteLaunchProfile.mockResolvedValue(undefined);
    mockLaunchVRChatWithArgs.mockResolvedValue(undefined);
  });

  it("renders launcher title", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();
    expect(wrapper.find(".page-title").text()).toBe("ランチャー");
  });

  it("renders GUI items: VR mode toggle, clear cache, screen mode toggle, custom args", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();
    // Select first profile to show editor
    const card = wrapper.findAll(".profile-card")[0];
    await card?.trigger("click");
    await flushPromises();

    expect(wrapper.find('[data-testid="vr-mode-desktop"]').exists()).toBe(true);
    expect(wrapper.find('[data-testid="vr-mode-none"]').exists()).toBe(true);
    expect(wrapper.find('[data-testid="vr-mode-vr"]').exists()).toBe(true);
    expect(wrapper.find('[data-testid="clear-cache-checkbox"]').exists()).toBe(
      true,
    );
    expect(
      wrapper.find('[data-testid="screen-mode-fullscreen"]').exists(),
    ).toBe(true);
    expect(wrapper.find('[data-testid="custom-args-input"]').exists()).toBe(
      true,
    );
  });

  it("parses arguments on profile select and reflects GUI state", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    mockParseLaunchArgsForGUI.mockResolvedValue({
      vrMode: "desktop",
      clearCache: true,
      screenMode: "fullscreen",
      custom: "-batchmode",
    });

    const cardWithCache = wrapper.findAll(".profile-card")[1];
    await cardWithCache?.trigger("click");
    await flushPromises();

    expect(mockParseLaunchArgsForGUI).toHaveBeenLastCalledWith(
      "--no-vr --clear-cache -screen-fullscreen 1",
    );

    const desktopRadio = wrapper.find('[data-testid="vr-mode-desktop"]');
    const clearCacheCheckbox = wrapper.find(
      '[data-testid="clear-cache-checkbox"]',
    );
    const fullscreenRadio = wrapper.find(
      '[data-testid="screen-mode-fullscreen"]',
    );
    const customInput = wrapper.find('[data-testid="custom-args-input"]');

    expect((desktopRadio.element as HTMLInputElement).checked).toBe(true);
    expect((clearCacheCheckbox.element as HTMLInputElement).checked).toBe(true);
    expect((fullscreenRadio.element as HTMLInputElement).checked).toBe(true);
    expect((customInput.element as HTMLInputElement).value).toBe("-batchmode");
  });

  it("merges GUI state to arguments on save", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    const card = wrapper.findAll(".profile-card")[0];
    await card?.trigger("click");
    await flushPromises();

    mockParseLaunchArgsForGUI.mockResolvedValue({
      vrMode: "desktop",
      clearCache: false,
      screenMode: "",
      custom: "",
    });
    await flushPromises();

    const desktopRadio = wrapper.find('[data-testid="vr-mode-desktop"]');
    await desktopRadio.setValue("desktop");
    const clearCacheCheckbox = wrapper.find(
      '[data-testid="clear-cache-checkbox"]',
    );
    await clearCacheCheckbox.setValue(true);
    const customInput = wrapper.find('[data-testid="custom-args-input"]');
    await customInput.setValue("-batchmode");
    await flushPromises();

    const saveBtn = wrapper.find(".btn-save");
    await saveBtn.trigger("click");
    await flushPromises();

    expect(mockMergeLaunchArgsForGUI).toHaveBeenCalledWith(
      expect.objectContaining({
        vrMode: "desktop",
        clearCache: true,
        screenMode: "",
        custom: "-batchmode",
      }),
    );
    expect(mockSaveLaunchProfile).toHaveBeenCalled();
  });

  it("renders detailed settings when expanded", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();
    const card = wrapper.findAll(".profile-card")[0];
    await card?.trigger("click");
    await flushPromises();

    const details = wrapper.find(".details-advanced");
    expect(details.exists()).toBe(true);
    const summary = details.find("summary");
    expect(summary.text()).toContain("詳細設定");

    await summary.trigger("click");
    await flushPromises();

    expect(wrapper.find('[data-testid="vr-mode-vr"]').exists()).toBe(true);
    expect(wrapper.find('[data-testid="fpfc-checkbox"]').exists()).toBe(true);
    expect(wrapper.find('[data-testid="screen-mode-windowed"]').exists()).toBe(
      true,
    );
    expect(wrapper.find('[data-testid="screen-width-input"]').exists()).toBe(
      true,
    );
    expect(wrapper.find('[data-testid="fps-input"]').exists()).toBe(true);
    expect(wrapper.find('[data-testid="safe-checkbox"]').exists()).toBe(true);
  });

  it("launch uses current GUI state via merge and launchVRChatWithArgs", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    const card = wrapper.findAll(".profile-card")[0];
    await card?.trigger("click");
    await flushPromises();

    mockMergeLaunchArgsForGUI.mockResolvedValue("--no-vr -screen-fullscreen 1");

    const desktopRadio = wrapper.find('[data-testid="vr-mode-desktop"]');
    await desktopRadio.setValue("desktop");
    const fullscreenRadio = wrapper.find(
      '[data-testid="screen-mode-fullscreen"]',
    );
    await fullscreenRadio.setValue("fullscreen");
    await flushPromises();

    const launchBtn = wrapper.find(".btn-launch");
    await launchBtn.trigger("click");
    await flushPromises();

    expect(mockMergeLaunchArgsForGUI).toHaveBeenCalledWith(
      expect.objectContaining({
        vrMode: "desktop",
        screenMode: "fullscreen",
      }),
    );
    expect(mockLaunchVRChatWithArgs).toHaveBeenCalledWith(
      "--no-vr -screen-fullscreen 1",
    );
  });

  it("shows delete button only for saved profiles and deletes on confirm", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    const card = wrapper.findAll(".profile-card")[0];
    await card?.trigger("click");
    await flushPromises();

    expect(wrapper.find('[data-testid="delete-profile-btn"]').exists()).toBe(
      true,
    );

    mockDeleteLaunchProfile.mockResolvedValue(undefined);
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(true);

    const deleteBtn = wrapper.find('[data-testid="delete-profile-btn"]');
    await deleteBtn.trigger("click");
    await flushPromises();

    expect(confirmSpy).toHaveBeenCalledWith("「Default」を削除しますか？");
    expect(mockDeleteLaunchProfile).toHaveBeenCalledWith("1");
    expect(mockLaunchProfiles).toHaveBeenCalled();

    confirmSpy.mockRestore();
  });

  it("does not delete when user cancels confirmation", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    const card = wrapper.findAll(".profile-card")[1];
    await card?.trigger("click");
    await flushPromises();

    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(false);

    const deleteBtn = wrapper.find('[data-testid="delete-profile-btn"]');
    await deleteBtn.trigger("click");
    await flushPromises();

    expect(mockDeleteLaunchProfile).not.toHaveBeenCalled();
    confirmSpy.mockRestore();
  });

  it("does not show delete button for new unsaved profile", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    const addBtn = wrapper.find(".btn-add");
    await addBtn.trigger("click");
    await flushPromises();

    expect(wrapper.find('[data-testid="delete-profile-btn"]').exists()).toBe(
      false,
    );
  });
});
