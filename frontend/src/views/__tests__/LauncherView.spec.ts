import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { ElMessageBox } from "element-plus";
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

/** ElRadioButton 内の native radio input を返す（checked 確認用） */
function radioInput(wrapper: ReturnType<typeof mount>, testId: string) {
  return wrapper.find(`[data-testid="${testId}"] input`);
}

/** ElCheckbox 内の native checkbox input を返す（setValue / checked 用） */
function checkInput(wrapper: ReturnType<typeof mount>, testId: string) {
  return wrapper.find(`[data-testid="${testId}"] input`);
}

/** ElInput / ElInputNumber の内側 input を返す */
function inputInner(wrapper: ReturnType<typeof mount>, testId: string) {
  return wrapper.find(`[data-testid="${testId}"] input`);
}

/** ElInputNumber 内 input の disabled を返す */
function isDisabled(wrapper: ReturnType<typeof mount>, testId: string) {
  return (inputInner(wrapper, testId).element as HTMLInputElement).disabled;
}

async function openAdvancedCollapse(wrapper: ReturnType<typeof mount>) {
  const collapse = wrapper.find(".args-collapse");
  if (!collapse.find(".el-collapse-item.is-active").exists()) {
    await collapse.find(".el-collapse-item__header").trigger("click");
    await flushPromises();
  }
}

async function clickDeleteFromOverflow(wrapper: ReturnType<typeof mount>) {
  await wrapper.find('[data-testid="profile-overflow-btn"]').trigger("click");
  await flushPromises();
  const deleteItem = document.querySelector(
    '[data-testid="delete-profile-btn"]',
  );
  expect(deleteItem).toBeTruthy();
  await (deleteItem as HTMLElement).click();
  await flushPromises();
}

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

    // ElRadioButton 内の native radio input で checked を確認
    const fullscreenRadioInput = radioInput(wrapper, "screen-mode-fullscreen");
    expect((fullscreenRadioInput.element as HTMLInputElement).checked).toBe(
      true,
    );

    // ElInput は data-testid を inner input に転送するため直接セレクタで取得
    const customInner = wrapper.find('[data-testid="custom-args-input"]');
    expect((customInner.element as HTMLInputElement).value).toBe("-batchmode");
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

    // fullscreen ラジオボタンの inner input に setValue で選択
    await wrapper
      .find('[data-testid="screen-mode-fullscreen"] input')
      .setValue(true);
    // ElInput は data-testid を inner input に転送するため直接セレクタで値をセット
    await wrapper
      .find('[data-testid="custom-args-input"]')
      .setValue("-batchmode");
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

    // el-collapse の存在確認
    const collapse = wrapper.find(".args-collapse");
    expect(collapse.exists()).toBe(true);

    // el-collapse-item のヘッダーをクリックして展開
    const header = collapse.find(".el-collapse-item__header");
    expect(header.text()).toContain("すべてのオプション");
    await header.trigger("click");
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

    await openAdvancedCollapse(wrapper);

    // ElCheckbox 内の input で setValue
    await checkInput(wrapper, "resolution-enabled-checkbox").setValue(true);
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

    await openAdvancedCollapse(wrapper);

    await checkInput(wrapper, "resolution-enabled-checkbox").setValue(true);
    await flushPromises();

    // デフォルトの HD プリセットでは disabled になる
    expect(isDisabled(wrapper, "screen-width-input")).toBe(true);
    expect(isDisabled(wrapper, "screen-height-input")).toBe(true);
  });

  it("enables resolution inputs when preset is custom", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();
    const card = wrapper.findAll(".profile-card")[0];
    await card?.trigger("click");
    await flushPromises();

    await openAdvancedCollapse(wrapper);

    await checkInput(wrapper, "resolution-enabled-checkbox").setValue(true);
    await flushPromises();

    // custom プリセットの inner radio input に setValue で選択
    await wrapper
      .find("[data-testid='resolution-preset-custom'] input")
      .setValue(true);
    await flushPromises();

    expect(isDisabled(wrapper, "screen-width-input")).toBe(false);
    expect(isDisabled(wrapper, "screen-height-input")).toBe(false);
  });

  it("launch uses current GUI state via merge and launchVRChatWithArgs", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    const card = wrapper.findAll(".profile-card")[0];
    await card?.trigger("click");
    await flushPromises();

    mockMergeLaunchArgsForGUI.mockResolvedValue("-screen-fullscreen 1");

    await wrapper
      .find('[data-testid="screen-mode-fullscreen"]')
      .trigger("click");
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

  it("shows delete in overflow menu for saved profiles and deletes on confirm", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    const card = wrapper.findAll(".profile-card")[0];
    await card?.trigger("click");
    await flushPromises();

    expect(wrapper.find('[data-testid="profile-overflow-btn"]').exists()).toBe(
      true,
    );

    mockDeleteLaunchProfile.mockResolvedValue(undefined);
    const confirmSpy = vi
      .spyOn(ElMessageBox, "confirm")
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      .mockResolvedValue("confirm" as any);

    await clickDeleteFromOverflow(wrapper);

    expect(confirmSpy).toHaveBeenCalledWith(
      "「Default」を削除しますか？",
      "確認",
      expect.any(Object),
    );
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

    // ElMessageBox.confirm をモック（reject = キャンセル）
    const confirmSpy = vi
      .spyOn(ElMessageBox, "confirm")
      .mockRejectedValue("cancel");

    await clickDeleteFromOverflow(wrapper);

    expect(mockDeleteLaunchProfile).not.toHaveBeenCalled();
    confirmSpy.mockRestore();
  });

  it("does not show overflow menu for new unsaved profile", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    const addBtn = wrapper.find(".btn-add");
    await addBtn.trigger("click");
    await flushPromises();

    expect(wrapper.find('[data-testid="profile-overflow-btn"]').exists()).toBe(
      false,
    );
  });

  it("addNew creates first profile as default when list is empty", async () => {
    mockLaunchProfiles.mockResolvedValue([]);
    const wrapper = mount(LauncherView);
    await flushPromises();

    await wrapper.find(".btn-add").trigger("click");
    await flushPromises();

    const defaultLabel = wrapper
      .findAll(".el-checkbox")
      .find((c) => c.text().includes("デフォルトに設定"));
    const defaultInput = defaultLabel?.find("input");
    expect((defaultInput?.element as HTMLInputElement).checked).toBe(true);
  });

  it("saves a new profile via merge and saveLaunchProfile", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    await wrapper.find(".btn-add").trigger("click");
    await flushPromises();

    await wrapper
      .find(".profile-editor .el-input input")
      .setValue("My Custom Profile");
    await checkInput(wrapper, "no-vr-checkbox").setValue(true);
    mockMergeLaunchArgsForGUI.mockResolvedValue("-no-vr");
    await wrapper.find(".btn-save").trigger("click");
    await flushPromises();

    expect(mockMergeLaunchArgsForGUI).toHaveBeenCalledWith(
      expect.objectContaining({ noVr: true }),
    );
    expect(mockSaveLaunchProfile).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "",
        name: "My Custom Profile",
        arguments: "-no-vr",
      }),
    );
    expect(mockLaunchProfiles).toHaveBeenCalled();
  });

  it("applies resolution preset dimensions on save", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    await wrapper.findAll(".profile-card")[0]?.trigger("click");
    await flushPromises();

    await openAdvancedCollapse(wrapper);

    await checkInput(wrapper, "resolution-enabled-checkbox").setValue(true);
    await flushPromises();
    await wrapper
      .find("[data-testid='resolution-preset-4k'] input")
      .setValue(true);
    await flushPromises();

    mockMergeLaunchArgsForGUI.mockResolvedValue(
      "-screen-width 3840 -screen-height 2160",
    );
    await wrapper.find(".btn-save").trigger("click");
    await flushPromises();

    expect(mockMergeLaunchArgsForGUI).toHaveBeenCalledWith(
      expect.objectContaining({
        screenWidth: 3840,
        screenHeight: 2160,
      }),
    );
  });

  it("clears resolution when disabled before save", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    mockParseLaunchArgsForGUI.mockResolvedValue({
      noVr: false,
      screenMode: "",
      screenWidth: 1920,
      screenHeight: 1080,
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

    await wrapper.findAll(".profile-card")[0]?.trigger("click");
    await flushPromises();

    await openAdvancedCollapse(wrapper);

    await checkInput(wrapper, "resolution-enabled-checkbox").setValue(false);
    await flushPromises();

    mockMergeLaunchArgsForGUI.mockResolvedValue("");
    await wrapper.find(".btn-save").trigger("click");
    await flushPromises();

    expect(mockMergeLaunchArgsForGUI).toHaveBeenCalledWith(
      expect.objectContaining({
        screenWidth: 0,
        screenHeight: 0,
      }),
    );
  });

  it("syncs advanced option toggles from parsed args", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    mockParseLaunchArgsForGUI.mockResolvedValue({
      noVr: false,
      screenMode: "",
      screenWidth: 0,
      screenHeight: 0,
      fps: 90,
      skipRegistry: false,
      processPriority: 1,
      mainThreadPriority: -999,
      monitor: 2,
      profile: 0,
      enableDebugGui: false,
      enableSDKLogLevels: false,
      enableUdonDebugLogging: false,
      midi: "device1",
      watchWorlds: false,
      watchAvatars: false,
      ignoreTrackers: "serial1",
      videoDecoding: "",
      disableAMDStutterWorkaround: false,
      osc: "9000",
      affinity: "0,1",
      enforceWorldServerChecks: false,
      custom: "",
    } as LaunchArgsParsedDTO);

    await wrapper.findAll(".profile-card")[1]?.trigger("click");
    await flushPromises();

    await openAdvancedCollapse(wrapper);

    expect(
      (
        checkInput(wrapper, "monitor-enabled-checkbox")
          .element as HTMLInputElement
      ).checked,
    ).toBe(true);
    expect(
      (checkInput(wrapper, "fps-enabled-checkbox").element as HTMLInputElement)
        .checked,
    ).toBe(true);
    expect(
      (
        checkInput(wrapper, "process-priority-enabled-checkbox")
          .element as HTMLInputElement
      ).checked,
    ).toBe(true);
    expect(
      (
        checkInput(wrapper, "profile-enabled-checkbox")
          .element as HTMLInputElement
      ).checked,
    ).toBe(true);

    expect(
      (checkInput(wrapper, "midi-enabled-checkbox").element as HTMLInputElement)
        .checked,
    ).toBe(true);
    expect(wrapper.find('[data-testid="midi-input"]').exists()).toBe(true);
    expect(
      (
        checkInput(wrapper, "ignore-trackers-enabled-checkbox")
          .element as HTMLInputElement
      ).checked,
    ).toBe(true);
    expect(
      (checkInput(wrapper, "osc-enabled-checkbox").element as HTMLInputElement)
        .checked,
    ).toBe(true);
    expect(
      (
        checkInput(wrapper, "affinity-enabled-checkbox")
          .element as HTMLInputElement
      ).checked,
    ).toBe(true);
  });

  it("clears optional fields when their enable toggles are turned off", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    mockParseLaunchArgsForGUI.mockResolvedValue({
      noVr: false,
      screenMode: "",
      screenWidth: 0,
      screenHeight: 0,
      fps: 90,
      processPriority: -999,
      mainThreadPriority: -999,
      monitor: 2,
      profile: -1,
      midi: "device1",
      ignoreTrackers: "serial1",
      osc: "9000",
      affinity: "0,1",
      custom: "",
    } as LaunchArgsParsedDTO);

    await wrapper.findAll(".profile-card")[0]?.trigger("click");
    await flushPromises();

    await openAdvancedCollapse(wrapper);

    await checkInput(wrapper, "monitor-enabled-checkbox").setValue(false);
    await checkInput(wrapper, "fps-enabled-checkbox").setValue(false);
    await flushPromises();

    await checkInput(wrapper, "midi-enabled-checkbox").setValue(false);
    await checkInput(wrapper, "osc-enabled-checkbox").setValue(false);
    await checkInput(wrapper, "ignore-trackers-enabled-checkbox").setValue(
      false,
    );
    await checkInput(wrapper, "affinity-enabled-checkbox").setValue(false);
    await flushPromises();

    mockMergeLaunchArgsForGUI.mockResolvedValue("");
    await wrapper.find(".btn-save").trigger("click");
    await flushPromises();

    expect(mockMergeLaunchArgsForGUI).toHaveBeenCalledWith(
      expect.objectContaining({
        monitor: 0,
        fps: 0,
        midi: "",
        osc: "",
        ignoreTrackers: "",
        affinity: "",
      }),
    );
  });

  it("enables main thread priority and profile index when toggled on", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    await wrapper.findAll(".profile-card")[0]?.trigger("click");
    await flushPromises();

    await openAdvancedCollapse(wrapper);

    await checkInput(wrapper, "main-thread-priority-enabled-checkbox").setValue(
      true,
    );
    await checkInput(wrapper, "profile-enabled-checkbox").setValue(true);
    await flushPromises();

    mockMergeLaunchArgsForGUI.mockResolvedValue("");
    await wrapper.find(".btn-save").trigger("click");
    await flushPromises();

    expect(mockMergeLaunchArgsForGUI).toHaveBeenCalledWith(
      expect.objectContaining({
        mainThreadPriority: 0,
        profile: 0,
      }),
    );
  });

  it("detects custom resolution preset from parsed args", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    mockParseLaunchArgsForGUI.mockResolvedValue({
      noVr: false,
      screenMode: "",
      screenWidth: 1600,
      screenHeight: 900,
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

    await wrapper.findAll(".profile-card")[1]?.trigger("click");
    await flushPromises();

    await openAdvancedCollapse(wrapper);

    expect(
      (
        checkInput(wrapper, "resolution-enabled-checkbox")
          .element as HTMLInputElement
      ).checked,
    ).toBe(true);
    expect(isDisabled(wrapper, "screen-width-input")).toBe(false);
    expect(isDisabled(wrapper, "screen-height-input")).toBe(false);
  });

  it("launches with popupwindow screen mode", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    await wrapper.findAll(".profile-card")[0]?.trigger("click");
    await flushPromises();

    await wrapper
      .find('[data-testid="screen-mode-popupwindow"] input')
      .setValue(true);
    mockMergeLaunchArgsForGUI.mockResolvedValue("-popupwindow");
    await wrapper.find(".btn-launch").trigger("click");
    await flushPromises();

    expect(mockMergeLaunchArgsForGUI).toHaveBeenCalledWith(
      expect.objectContaining({ screenMode: "popupwindow" }),
    );
    expect(mockLaunchVRChatWithArgs).toHaveBeenCalledWith("-popupwindow");
  });

  it("mounts without profiles and opens editor only after add", async () => {
    mockLaunchProfiles.mockResolvedValue([]);
    const wrapper = mount(LauncherView);
    await flushPromises();

    expect(wrapper.find(".profile-editor").exists()).toBe(false);
    await wrapper.find(".btn-add").trigger("click");
    await flushPromises();
    expect(wrapper.find(".profile-editor").exists()).toBe(true);
  });

  it("defaults resolution to FHD when enabling empty resolution fields", async () => {
    const wrapper = mount(LauncherView);
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

    await wrapper.findAll(".profile-card")[0]?.trigger("click");
    await flushPromises();

    await openAdvancedCollapse(wrapper);

    await checkInput(wrapper, "resolution-enabled-checkbox").setValue(true);
    await flushPromises();

    mockMergeLaunchArgsForGUI.mockResolvedValue(
      "-screen-width 1920 -screen-height 1080",
    );
    await wrapper.find(".btn-save").trigger("click");
    await flushPromises();

    expect(mockMergeLaunchArgsForGUI).toHaveBeenCalledWith(
      expect.objectContaining({
        screenWidth: 1920,
        screenHeight: 1080,
      }),
    );
  });

  it("merges monitor, fps, process priority, and debug flags on launch", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    await wrapper.findAll(".profile-card")[0]?.trigger("click");
    await flushPromises();

    await openAdvancedCollapse(wrapper);

    await checkInput(wrapper, "monitor-enabled-checkbox").setValue(true);
    await checkInput(wrapper, "fps-enabled-checkbox").setValue(true);
    await checkInput(wrapper, "process-priority-enabled-checkbox").setValue(
      true,
    );
    await checkInput(wrapper, "skip-registry-checkbox").setValue(true);
    await flushPromises();

    await inputInner(wrapper, "monitor-input").setValue(2);
    await inputInner(wrapper, "fps-input").setValue(144);
    await inputInner(wrapper, "process-priority-input").setValue(1);

    await checkInput(wrapper, "enable-debug-gui-checkbox").setValue(true);
    await checkInput(wrapper, "watch-worlds-checkbox").setValue(true);
    await checkInput(wrapper, "watch-avatars-checkbox").setValue(true);
    await checkInput(wrapper, "enforce-world-server-checks-checkbox").setValue(
      true,
    );
    await checkInput(
      wrapper,
      "disable-amd-stutter-workaround-checkbox",
    ).setValue(true);
    await wrapper
      .find('[data-testid="video-decoding-software"] input')
      .setValue(true);
    await flushPromises();

    mockMergeLaunchArgsForGUI.mockResolvedValue("merged-args");
    await wrapper.find(".btn-launch").trigger("click");
    await flushPromises();

    expect(mockMergeLaunchArgsForGUI).toHaveBeenCalledWith(
      expect.objectContaining({
        monitor: 2,
        fps: 144,
        processPriority: 1,
        skipRegistry: true,
        enableDebugGui: true,
        watchWorlds: true,
        watchAvatars: true,
        enforceWorldServerChecks: true,
        disableAMDStutterWorkaround: true,
        videoDecoding: "software",
      }),
    );
    expect(mockLaunchVRChatWithArgs).toHaveBeenCalledWith("merged-args");
  });

  it("selects remaining profile after delete succeeds", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();

    await wrapper.findAll(".profile-card")[0]?.trigger("click");
    await flushPromises();

    mockDeleteLaunchProfile.mockResolvedValue(undefined);
    mockLaunchProfiles.mockResolvedValue([sampleProfiles[1]!]);
    mockParseLaunchArgsForGUI.mockResolvedValue({
      noVr: false,
      screenMode: "windowed",
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

    const confirmSpy = vi
      .spyOn(ElMessageBox, "confirm")
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      .mockResolvedValue("confirm" as any);

    await clickDeleteFromOverflow(wrapper);
    await flushPromises();

    expect(mockDeleteLaunchProfile).toHaveBeenCalledWith("1");
    expect(wrapper.text()).toContain("With fullscreen");

    confirmSpy.mockRestore();
  });

  it("sets aria-label on screen mode, resolution preset, and video decoding radio groups", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();
    const card = wrapper.findAll(".profile-card")[0];
    await card?.trigger("click");
    await flushPromises();

    const screenRg = wrapper
      .get('[data-testid="screen-mode-fullscreen"]')
      .element.closest("[role='radiogroup']");
    expect(screenRg?.getAttribute("aria-label")).toBe("表示モード");

    await openAdvancedCollapse(wrapper);
    await checkInput(wrapper, "resolution-enabled-checkbox").setValue(true);
    await flushPromises();

    const resRg = wrapper
      .get("[data-testid='resolution-preset-hd']")
      .element.closest("[role='radiogroup']");
    expect(resRg?.getAttribute("aria-label")).toBe("プリセット");

    expect(
      wrapper.find('[data-testid="advanced-debug-section"]').exists(),
    ).toBe(true);

    const vdRg = wrapper
      .get('[data-testid="video-decoding-default"]')
      .element.closest("[role='radiogroup']");
    expect(vdRg?.getAttribute("aria-label")).toBe("動画デコーディング");
  });

  it("shows unsaved banner after editing launch args", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();
    await wrapper.findAll(".profile-card")[0]?.trigger("click");
    await flushPromises();

    expect(wrapper.find('[data-testid="unsaved-banner"]').exists()).toBe(false);

    await checkInput(wrapper, "no-vr-checkbox").setValue(true);
    await flushPromises();

    expect(wrapper.find('[data-testid="unsaved-banner"]').exists()).toBe(true);
    expect(wrapper.find('[data-testid="unsaved-dot"]').exists()).toBe(true);
  });

  it("prompts before switching profiles when edits are unsaved", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();
    await wrapper.findAll(".profile-card")[0]?.trigger("click");
    await flushPromises();

    await checkInput(wrapper, "no-vr-checkbox").setValue(true);
    await flushPromises();

    const confirmSpy = vi
      .spyOn(ElMessageBox, "confirm")
      .mockRejectedValue("close");

    await wrapper.findAll(".profile-card")[1]?.trigger("click");
    await flushPromises();

    expect(confirmSpy).toHaveBeenCalled();
    expect(mockParseLaunchArgsForGUI).toHaveBeenCalledTimes(1);

    confirmSpy.mockRestore();
  });

  it("discards unsaved edits when user chooses discard on switch", async () => {
    const wrapper = mount(LauncherView);
    await flushPromises();
    await wrapper.findAll(".profile-card")[0]?.trigger("click");
    await flushPromises();

    await checkInput(wrapper, "no-vr-checkbox").setValue(true);
    await flushPromises();

    const confirmSpy = vi
      .spyOn(ElMessageBox, "confirm")
      .mockRejectedValue("cancel");

    await wrapper.findAll(".profile-card")[1]?.trigger("click");
    await flushPromises();

    expect(mockParseLaunchArgsForGUI).toHaveBeenCalledTimes(2);
    expect(wrapper.find('[data-testid="unsaved-banner"]').exists()).toBe(false);

    confirmSpy.mockRestore();
  });
});
