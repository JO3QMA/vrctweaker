import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import LauncherView from "../LauncherView.vue";
import type { LaunchProfileDTO, LaunchArgsParsedDTO } from "../../wails/app";

const {
  mockLaunchProfiles,
  mockParseLaunchArgsForGUI,
  mockMergeLaunchArgsForGUI,
  mockSaveLaunchProfile,
  mockLaunchVRChatWithArgs,
} = vi.hoisted(() => ({
  mockLaunchProfiles: vi.fn(),
  mockParseLaunchArgsForGUI: vi.fn(),
  mockMergeLaunchArgsForGUI: vi.fn(),
  mockSaveLaunchProfile: vi.fn(),
  mockLaunchVRChatWithArgs: vi.fn(),
}));

vi.mock("../../wails/app", () => ({
  App: {
    launchProfiles: mockLaunchProfiles,
    parseLaunchArgsForGUI: mockParseLaunchArgsForGUI,
    mergeLaunchArgsForGUI: mockMergeLaunchArgsForGUI,
    saveLaunchProfile: mockSaveLaunchProfile,
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
        if (args.includes("--no-vr")) {
          return {
            noVr: true,
            clearCache: args.includes("--clear-cache"),
            fullscreen: args.includes("-screen-fullscreen 1"),
            custom: args.includes("-batchmode") ? "-batchmode" : "",
          };
        }
        return {
          noVr: false,
          clearCache: false,
          fullscreen: false,
          custom: args.trim() || "",
        };
      },
    );
    mockMergeLaunchArgsForGUI.mockImplementation(
      async (dto: LaunchArgsParsedDTO): Promise<string> => {
        const parts: string[] = [];
        if (dto.noVr) parts.push("-no-vr");
        if (dto.clearCache) parts.push("--clear-cache");
        if (dto.fullscreen) parts.push("-screen-fullscreen 1");
        if (dto.custom) parts.push(dto.custom);
        return parts.join(" ");
      },
    );
    mockSaveLaunchProfile.mockResolvedValue(undefined);
    mockLaunchVRChatWithArgs.mockResolvedValue(undefined);
  });

  it("renders launcher title", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();
    expect(wrapper.find(".page-title").text()).toBe("ランチャー");
  });

  it("renders GUI items: desktop mode, clear cache, fullscreen, custom args", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();
    // Select first profile to show editor
    const card = wrapper.findAll(".profile-card")[0];
    await card?.trigger("click");
    await flushPromises();

    expect(wrapper.find('[data-testid="no-vr-checkbox"]').exists()).toBe(true);
    expect(wrapper.find('[data-testid="clear-cache-checkbox"]').exists()).toBe(
      true,
    );
    expect(wrapper.find('[data-testid="fullscreen-checkbox"]').exists()).toBe(
      true,
    );
    expect(wrapper.find('[data-testid="custom-args-input"]').exists()).toBe(
      true,
    );
  });

  it("parses arguments on profile select and reflects GUI state", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    mockParseLaunchArgsForGUI.mockResolvedValue({
      noVr: true,
      clearCache: true,
      fullscreen: true,
      custom: "-batchmode",
    });

    const cardWithCache = wrapper.findAll(".profile-card")[1];
    await cardWithCache?.trigger("click");
    await flushPromises();

    expect(mockParseLaunchArgsForGUI).toHaveBeenLastCalledWith(
      "--no-vr --clear-cache -screen-fullscreen 1",
    );

    const noVrCheckbox = wrapper.find('[data-testid="no-vr-checkbox"]');
    const clearCacheCheckbox = wrapper.find(
      '[data-testid="clear-cache-checkbox"]',
    );
    const fullscreenCheckbox = wrapper.find(
      '[data-testid="fullscreen-checkbox"]',
    );
    const customInput = wrapper.find('[data-testid="custom-args-input"]');

    expect((noVrCheckbox.element as HTMLInputElement).checked).toBe(true);
    expect((clearCacheCheckbox.element as HTMLInputElement).checked).toBe(true);
    expect((fullscreenCheckbox.element as HTMLInputElement).checked).toBe(true);
    expect((customInput.element as HTMLInputElement).value).toBe("-batchmode");
  });

  it("merges GUI state to arguments on save", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    const card = wrapper.findAll(".profile-card")[0];
    await card?.trigger("click");
    await flushPromises();

    mockParseLaunchArgsForGUI.mockResolvedValue({
      noVr: true,
      clearCache: false,
      fullscreen: false,
      custom: "",
    });
    await flushPromises();

    const noVrCheckbox = wrapper.find('[data-testid="no-vr-checkbox"]');
    await noVrCheckbox.setValue(true);
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
        noVr: true,
        clearCache: true,
        fullscreen: false,
        custom: "-batchmode",
      }),
    );
    expect(mockSaveLaunchProfile).toHaveBeenCalled();
  });

  it("launch uses current GUI state via merge and launchVRChatWithArgs", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    const card = wrapper.findAll(".profile-card")[0];
    await card?.trigger("click");
    await flushPromises();

    mockMergeLaunchArgsForGUI.mockResolvedValue("--no-vr -screen-fullscreen 1");

    const noVrCheckbox = wrapper.find('[data-testid="no-vr-checkbox"]');
    await noVrCheckbox.setValue(true);
    const fullscreenCheckbox = wrapper.find(
      '[data-testid="fullscreen-checkbox"]',
    );
    await fullscreenCheckbox.setValue(true);
    await flushPromises();

    const launchBtn = wrapper.find(".btn-launch");
    await launchBtn.trigger("click");
    await flushPromises();

    expect(mockMergeLaunchArgsForGUI).toHaveBeenCalledWith(
      expect.objectContaining({
        noVr: true,
        fullscreen: true,
      }),
    );
    expect(mockLaunchVRChatWithArgs).toHaveBeenCalledWith(
      "--no-vr -screen-fullscreen 1",
    );
  });
});
