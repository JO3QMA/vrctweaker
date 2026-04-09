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
    expect(header.text()).toContain("詳細設定");
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

    const collapse = wrapper.find(".args-collapse");
    await collapse.find(".el-collapse-item__header").trigger("click");
    await flushPromises();

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

    const collapse = wrapper.find(".args-collapse");
    await collapse.find(".el-collapse-item__header").trigger("click");
    await flushPromises();

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

    const collapse = wrapper.find(".args-collapse");
    await collapse.find(".el-collapse-item__header").trigger("click");
    await flushPromises();

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
    // ElMessageBox.confirm をモック（resolve = 削除確定）
    const confirmSpy = vi
      .spyOn(ElMessageBox, "confirm")
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      .mockResolvedValue("confirm" as any);

    const deleteBtn = wrapper.find('[data-testid="delete-profile-btn"]');
    await deleteBtn.trigger("click");
    await flushPromises();

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

    const collapse = wrapper.find(".args-collapse");
    await collapse.find(".el-collapse-item__header").trigger("click");
    await flushPromises();
    await checkInput(wrapper, "resolution-enabled-checkbox").setValue(true);
    await flushPromises();

    const resRg = wrapper
      .get("[data-testid='resolution-preset-hd']")
      .element.closest("[role='radiogroup']");
    expect(resRg?.getAttribute("aria-label")).toBe("プリセット");

    const debugHeaders = wrapper.findAll(".el-collapse-item__header");
    const debugHeader = debugHeaders.find((h) =>
      h.text().includes("クリエイター・デバッグ向け"),
    );
    expect(debugHeader?.exists()).toBe(true);
    await debugHeader!.trigger("click");
    await flushPromises();

    const vdRg = wrapper
      .get('[data-testid="video-decoding-default"]')
      .element.closest("[role='radiogroup']");
    expect(vdRg?.getAttribute("aria-label")).toBe("動画デコーディング");
  });
});
