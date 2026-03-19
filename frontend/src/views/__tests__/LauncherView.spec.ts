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

vi.mock("../../wails/app", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../../wails/app")>();
  return {
    ...actual,
    App: {
      launchProfiles: mockLaunchProfiles,
      parseLaunchArgsForGUI: mockParseLaunchArgsForGUI,
      mergeLaunchArgsForGUI: mockMergeLaunchArgsForGUI,
      saveLaunchProfile: mockSaveLaunchProfile,
      deleteLaunchProfile: mockDeleteLaunchProfile,
      launchVRChatWithArgs: mockLaunchVRChatWithArgs,
    },
  };
});

const sampleProfiles: LaunchProfileDTO[] = [
  {
    id: "1",
    name: "Default",
    arguments: "-screen-fullscreen 1",
    isDefault: true,
  },
  {
    id: "2",
    name: "With fullscreen",
    arguments: "-screen-fullscreen 1 -windowed",
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
          noVr: args.includes("-no-vr") || args.includes("--no-vr"),
          screenWidth: 0,
          screenHeight: 0,
          fps: 0,
          skipRegistry: false,
          processPriority: -999,
          mainThreadPriority: -999,
          monitor: 0,
          profile: -1,
          enableDebugGui: false,
          enableSDKLogLevels: false,
          enableUdonDebugLogging: false,
          midi: "",
          watchWorlds: false,
          watchAvatars: false,
          ignoreTrackers: "",
          videoDecoding: "" as "" | "software" | "hardware",
          disableAMDStutterWorkaround: false,
          osc: "",
          affinity: "",
          enforceWorldServerChecks: false,
        };
        return {
          ...base,
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
        if (dto.noVr) parts.push("-no-vr");
        if (dto.screenMode === "fullscreen") parts.push("-screen-fullscreen 1");
        if (dto.screenMode === "windowed") parts.push("-windowed");
        if (dto.screenMode === "popupwindow") parts.push("-popupwindow");
        if (dto.screenWidth)
          parts.push("-screen-width", String(dto.screenWidth));
        if (dto.screenHeight)
          parts.push("-screen-height", String(dto.screenHeight));
        if (dto.fps) parts.push(`--fps=${dto.fps}`);
        if (dto.skipRegistry) parts.push("--skip-registry-install");
        if (
          typeof dto.processPriority === "number" &&
          dto.processPriority >= -2 &&
          dto.processPriority <= 2
        )
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

  it("renders GUI items: desktop mode checkbox, screen mode toggle, custom args", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();
    // Select first profile to show editor
    const card = wrapper.findAll(".profile-card")[0];
    await card?.trigger("click");
    await flushPromises();

    expect(wrapper.find('[data-testid="no-vr-checkbox"]').exists()).toBe(true);
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
      noVr: false,
      screenMode: "fullscreen",
      screenWidth: 0,
      screenHeight: 0,
      fps: 0,
      processPriority: -999,
      mainThreadPriority: -999,
      monitor: 0,
      profile: -1,
      midi: "",
      ignoreTrackers: "",
      osc: "",
      affinity: "",
      custom: "-batchmode",
    } as LaunchArgsParsedDTO);

    const card2 = wrapper.findAll(".profile-card")[1];
    await card2?.trigger("click");
    await flushPromises();

    expect(mockParseLaunchArgsForGUI).toHaveBeenLastCalledWith(
      "-screen-fullscreen 1 -windowed",
    );

    const fullscreenRadio = wrapper.find(
      '[data-testid="screen-mode-fullscreen"]',
    );
    const customInput = wrapper.find('[data-testid="custom-args-input"]');

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
      noVr: false,
      screenMode: "",
      screenWidth: 0,
      screenHeight: 0,
      fps: 0,
      processPriority: -999,
      mainThreadPriority: -999,
      monitor: 0,
      profile: -1,
      midi: "",
      ignoreTrackers: "",
      osc: "",
      affinity: "",
      custom: "",
    } as LaunchArgsParsedDTO);
    await flushPromises();

    const fullscreenRadio = wrapper.find(
      '[data-testid="screen-mode-fullscreen"]',
    );
    await fullscreenRadio.setValue("fullscreen");
    const customInput = wrapper.find('[data-testid="custom-args-input"]');
    await customInput.setValue("-batchmode");
    await flushPromises();

    const saveBtn = wrapper.find(".btn-save");
    await saveBtn.trigger("click");
    await flushPromises();

    expect(mockMergeLaunchArgsForGUI).toHaveBeenCalledWith(
      expect.objectContaining({
        screenMode: "fullscreen",
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

    expect(wrapper.find('[data-testid="screen-mode-windowed"]').exists()).toBe(
      true,
    );
    expect(
      wrapper.find('[data-testid="resolution-enabled-checkbox"]').exists(),
    ).toBe(true);
    expect(wrapper.find('[data-testid="fps-enabled-checkbox"]').exists()).toBe(
      true,
    );
  });

  it("has resolution preset toggles when resolution is enabled", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();
    const card = wrapper.findAll(".profile-card")[0];
    await card?.trigger("click");
    await flushPromises();

    const details = wrapper.find(".details-advanced");
    await details.find("summary").trigger("click");
    await flushPromises();

    const resolutionCheckbox = wrapper.find(
      '[data-testid="resolution-enabled-checkbox"]',
    );
    await resolutionCheckbox.setValue(true);
    await flushPromises();

    expect(wrapper.find("[data-testid='resolution-preset-hd']").exists()).toBe(
      true,
    );
    expect(wrapper.find("[data-testid='resolution-preset-fhd']").exists()).toBe(
      true,
    );
    expect(wrapper.find("[data-testid='resolution-preset-4k']").exists()).toBe(
      true,
    );
    expect(
      wrapper.find("[data-testid='resolution-preset-custom']").exists(),
    ).toBe(true);
  });

  it("disables resolution inputs when preset is not custom", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();
    const card = wrapper.findAll(".profile-card")[0];
    await card?.trigger("click");
    await flushPromises();

    const details = wrapper.find(".details-advanced");
    await details.find("summary").trigger("click");
    await flushPromises();

    const resolutionCheckbox = wrapper.find(
      '[data-testid="resolution-enabled-checkbox"]',
    );
    await resolutionCheckbox.setValue(true);
    await flushPromises();

    const widthInput = wrapper.find("[data-testid='screen-width-input']");
    const heightInput = wrapper.find("[data-testid='screen-height-input']");
    expect((widthInput.element as HTMLInputElement).disabled).toBe(true);
    expect((heightInput.element as HTMLInputElement).disabled).toBe(true);
  });

  it("enables resolution inputs when preset is custom", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();
    const card = wrapper.findAll(".profile-card")[0];
    await card?.trigger("click");
    await flushPromises();

    const details = wrapper.find(".details-advanced");
    await details.find("summary").trigger("click");
    await flushPromises();

    const resolutionCheckbox = wrapper.find(
      '[data-testid="resolution-enabled-checkbox"]',
    );
    await resolutionCheckbox.setValue(true);
    await flushPromises();

    const customRadio = wrapper.find(
      "[data-testid='resolution-preset-custom']",
    );
    await customRadio.setValue(true);
    await flushPromises();

    const widthInput = wrapper.find("[data-testid='screen-width-input']");
    const heightInput = wrapper.find("[data-testid='screen-height-input']");
    expect((widthInput.element as HTMLInputElement).disabled).toBe(false);
    expect((heightInput.element as HTMLInputElement).disabled).toBe(false);
  });

  it("launch uses current GUI state via merge and launchVRChatWithArgs", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    const card = wrapper.findAll(".profile-card")[0];
    await card?.trigger("click");
    await flushPromises();

    mockMergeLaunchArgsForGUI.mockResolvedValue("-screen-fullscreen 1");

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
        screenMode: "fullscreen",
      }),
    );
    expect(mockLaunchVRChatWithArgs).toHaveBeenCalledWith(
      "-screen-fullscreen 1",
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
