import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { ElMessage } from "element-plus";
import DashboardView from "../DashboardView.vue";
import type { LaunchProfileDTO } from "../../wails/app";

const {
  mockLaunchProfiles,
  mockSetStatus,
  mockSetStatusDescription,
  mockSetStatusAndDescription,
  mockLaunchVRChat,
  mockGetInstanceRejoinSection,
  mockInstanceRejoin,
  mockEventsOn,
  triggerEncountersChanged,
} = vi.hoisted(() => {
  let encountersChangedCb: (() => void) | undefined;
  return {
    mockLaunchProfiles: vi.fn(),
    mockSetStatus: vi.fn(),
    mockSetStatusDescription: vi.fn(),
    mockSetStatusAndDescription: vi.fn(),
    mockLaunchVRChat: vi.fn(),
    mockGetInstanceRejoinSection: vi.fn(),
    mockInstanceRejoin: vi.fn(),
    mockEventsOn: vi.fn((_event: string, cb: () => void) => {
      encountersChangedCb = cb;
      return vi.fn();
    }),
    triggerEncountersChanged: () => encountersChangedCb?.(),
  };
});

vi.mock("../../wails/app", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../../wails/app")>();
  return {
    ...actual,
    App: {
      ...actual.App,
      launchProfiles: mockLaunchProfiles,
      launchVRChat: mockLaunchVRChat,
      getInstanceRejoinSection: mockGetInstanceRejoinSection,
      instanceRejoin: mockInstanceRejoin,
      setStatus: mockSetStatus,
      setStatusDescription: mockSetStatusDescription,
      setStatusAndDescription: mockSetStatusAndDescription,
    },
  };
});

vi.mock("../../wails/runtime", () => ({
  getRuntime: () => ({
    EventsOn: mockEventsOn,
  }),
}));

const emptyProfiles: LaunchProfileDTO[] = [];

const defaultLaunchProfile: LaunchProfileDTO = {
  id: "p-default",
  name: "Default",
  arguments: "",
  isDefault: true,
};

describe("DashboardView", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockLaunchProfiles.mockResolvedValue([...emptyProfiles]);
    mockGetInstanceRejoinSection.mockResolvedValue(null);
    mockInstanceRejoin.mockResolvedValue(undefined);
    mockLaunchVRChat.mockResolvedValue(undefined);
    mockSetStatus.mockResolvedValue(undefined);
    mockSetStatusDescription.mockResolvedValue(undefined);
    mockSetStatusAndDescription.mockResolvedValue(undefined);
  });

  it("calls App.launchProfiles on mount", async () => {
    mount(DashboardView);
    await flushPromises();

    expect(mockLaunchProfiles).toHaveBeenCalledTimes(1);
  });

  it("disables the launch button when launchProfiles returns no profiles", async () => {
    const wrapper = mount(DashboardView);
    await flushPromises();

    const launchBtn = wrapper.find(".launch-btn");
    expect((launchBtn.element as HTMLButtonElement).disabled).toBe(true);
  });

  it("enables the launch button when launchProfiles returns a default profile", async () => {
    mockLaunchProfiles.mockResolvedValue([defaultLaunchProfile]);
    const wrapper = mount(DashboardView);
    await flushPromises();

    expect(mockLaunchProfiles).toHaveBeenCalledTimes(1);
    const launchBtn = wrapper.find(".launch-btn");
    expect((launchBtn.element as HTMLButtonElement).disabled).toBe(false);
    expect(launchBtn.text()).toContain("Default");
  });

  it("renders page title and quick status buttons", async () => {
    const wrapper = mount(DashboardView);
    await flushPromises();

    expect(wrapper.find(".page-title").text()).toBe("ダッシュボード");
    expect(
      wrapper.find('[data-testid="dashboard-quick-status-join-me"]').exists(),
    ).toBe(true);
    expect(
      wrapper.find('[data-testid="dashboard-quick-status-active"]').exists(),
    ).toBe(true);
    expect(
      wrapper.find('[data-testid="dashboard-quick-status-ask-me"]').exists(),
    ).toBe(true);
    expect(
      wrapper.find('[data-testid="dashboard-quick-status-busy"]').exists(),
    ).toBe(true);
  });

  it("setStatusOnly calls App.setStatus and shows success message", async () => {
    const successSpy = vi
      .spyOn(ElMessage, "success")
      .mockImplementation(() => ({
        close: () => {},
      }));
    const wrapper = mount(DashboardView);
    await flushPromises();

    await wrapper
      .find('[data-testid="dashboard-quick-status-active"]')
      .trigger("click");
    await flushPromises();

    expect(mockSetStatus).toHaveBeenCalledWith("active");
    expect(successSpy).toHaveBeenCalledWith("ステータスを更新しました");
    successSpy.mockRestore();
  });

  it("applyCustomDescription trims input and calls App.setStatusDescription", async () => {
    const successSpy = vi
      .spyOn(ElMessage, "success")
      .mockImplementation(() => ({
        close: () => {},
      }));
    const wrapper = mount(DashboardView);
    await flushPromises();

    const input = wrapper.find(".custom-status-input input");
    await input.setValue("  hello world  ");
    await wrapper.find(".apply-btn").trigger("click");
    await flushPromises();

    expect(mockSetStatusDescription).toHaveBeenCalledWith("hello world");
    expect(successSpy).toHaveBeenCalledWith("ステータスを更新しました");
    successSpy.mockRestore();
  });

  it("applyTemplate calls App.setStatusAndDescription with template status and label", async () => {
    const successSpy = vi
      .spyOn(ElMessage, "success")
      .mockImplementation(() => ({
        close: () => {},
      }));
    const wrapper = mount(DashboardView);
    await flushPromises();

    const busyTemplateBtn = wrapper
      .findAll(".templates-panel .status-btn")
      .find((b) => b.text() === "作業中");
    expect(busyTemplateBtn).toBeDefined();
    await busyTemplateBtn!.trigger("click");
    await flushPromises();

    expect(mockSetStatusAndDescription).toHaveBeenCalledWith("busy", "作業中");
    expect(successSpy).toHaveBeenCalledWith("ステータスを更新しました");
    successSpy.mockRestore();
  });

  it("shows ElMessage.error with error text when App methods reject", async () => {
    mockSetStatus.mockRejectedValueOnce(new Error("set status failed"));
    const errorSpy = vi.spyOn(ElMessage, "error").mockImplementation(() => ({
      close: () => {},
    }));

    const wrapper = mount(DashboardView);
    await flushPromises();

    await wrapper
      .find('[data-testid="dashboard-quick-status-busy"]')
      .trigger("click");
    await flushPromises();

    expect(errorSpy).toHaveBeenCalledWith("set status failed");
    errorSpy.mockRestore();
  });

  it("shows backend string rejections in ElMessage.error (not only Error instances)", async () => {
    mockSetStatus.mockRejectedValueOnce("  backend string err  ");
    const errorSpy = vi.spyOn(ElMessage, "error").mockImplementation(() => ({
      close: () => {},
    }));

    const wrapper = mount(DashboardView);
    await flushPromises();

    await wrapper
      .find('[data-testid="dashboard-quick-status-join-me"]')
      .trigger("click");
    await flushPromises();

    expect(errorSpy).toHaveBeenCalledWith("backend string err");
    errorSpy.mockRestore();
  });

  it("calls App.launchVRChat with default profile id when launch is clicked", async () => {
    mockLaunchProfiles.mockResolvedValue([defaultLaunchProfile]);
    const wrapper = mount(DashboardView);
    await flushPromises();

    await wrapper.find(".launch-btn").trigger("click");
    await flushPromises();

    expect(mockLaunchVRChat).toHaveBeenCalledWith("p-default");
  });

  it("uses first profile when none is marked default", async () => {
    mockLaunchProfiles.mockResolvedValue([
      {
        id: "p-first",
        name: "FirstOnly",
        arguments: "",
        isDefault: false,
      },
    ]);
    const wrapper = mount(DashboardView);
    await flushPromises();

    const launchBtn = wrapper.find(".launch-btn");
    expect((launchBtn.element as HTMLButtonElement).disabled).toBe(false);
    expect(launchBtn.text()).toContain("FirstOnly");
  });

  it("setStatusOnly calls join me and ask me statuses", async () => {
    const wrapper = mount(DashboardView);
    await flushPromises();

    await wrapper
      .find('[data-testid="dashboard-quick-status-join-me"]')
      .trigger("click");
    await flushPromises();
    expect(mockSetStatus).toHaveBeenCalledWith("join me");

    await wrapper
      .find('[data-testid="dashboard-quick-status-ask-me"]')
      .trigger("click");
    await flushPromises();
    expect(mockSetStatus).toHaveBeenCalledWith("ask me");
  });

  it("shows error when applyCustomDescription fails", async () => {
    mockSetStatusDescription.mockRejectedValueOnce(new Error("desc failed"));
    const errorSpy = vi.spyOn(ElMessage, "error").mockImplementation(() => ({
      close: () => {},
    }));

    const wrapper = mount(DashboardView);
    await flushPromises();
    await wrapper.find(".custom-status-input input").setValue("bad");
    await wrapper.find(".apply-btn").trigger("click");
    await flushPromises();

    expect(errorSpy).toHaveBeenCalledWith("desc failed");
    errorSpy.mockRestore();
  });

  it("shows error when applyTemplate fails", async () => {
    mockSetStatusAndDescription.mockRejectedValueOnce({ message: "tpl fail" });
    const errorSpy = vi.spyOn(ElMessage, "error").mockImplementation(() => ({
      close: () => {},
    }));

    const wrapper = mount(DashboardView);
    await flushPromises();

    const joinTemplateBtn = wrapper
      .findAll(".templates-panel .status-btn")
      .find((b) => b.text() === "だれでもどうぞ");
    await joinTemplateBtn!.trigger("click");
    await flushPromises();

    expect(errorSpy).toHaveBeenCalledWith("tpl fail");
    errorSpy.mockRestore();
  });

  it("hides instance rejoin section when unavailable", async () => {
    const wrapper = mount(DashboardView);
    await flushPromises();
    expect(
      wrapper.find('[data-testid="instance-rejoin-section"]').exists(),
    ).toBe(false);
  });

  it("shows world name on rejoin button", async () => {
    mockGetInstanceRejoinSection.mockResolvedValue({
      playSessionId: "ps-1",
      worldDisplayName: "Test World",
      profiles: [defaultLaunchProfile],
      selectedProfileId: "p-default",
    });
    const wrapper = mount(DashboardView);
    await flushPromises();
    expect(wrapper.find('[data-testid="instance-rejoin-btn"]').text()).toBe(
      "Test World に参加",
    );
  });

  it("shows generic rejoin label without world name", async () => {
    mockGetInstanceRejoinSection.mockResolvedValue({
      playSessionId: "ps-1",
      worldDisplayName: "",
      profiles: [defaultLaunchProfile],
      selectedProfileId: "p-default",
    });
    const wrapper = mount(DashboardView);
    await flushPromises();
    expect(wrapper.find('[data-testid="instance-rejoin-btn"]').text()).toBe(
      "最後のインスタンスに参加",
    );
  });

  it("shows error on instance rejoin failure", async () => {
    mockGetInstanceRejoinSection.mockResolvedValue({
      playSessionId: "ps-1",
      worldDisplayName: "",
      profiles: [defaultLaunchProfile],
      selectedProfileId: "p-default",
    });
    mockInstanceRejoin.mockRejectedValueOnce(new Error("rejoin failed"));
    const errorSpy = vi.spyOn(ElMessage, "error").mockImplementation(() => ({
      close: () => {},
    }));
    const wrapper = mount(DashboardView);
    await flushPromises();
    await wrapper.find('[data-testid="instance-rejoin-btn"]').trigger("click");
    await flushPromises();
    expect(mockInstanceRejoin).toHaveBeenCalledWith("p-default", "ps-1");
    expect(errorSpy).toHaveBeenCalledWith("rejoin failed");
    expect(mockGetInstanceRejoinSection).toHaveBeenCalledTimes(2);
    errorSpy.mockRestore();
  });

  it("shows error when instance rejoin section load fails", async () => {
    mockGetInstanceRejoinSection.mockRejectedValueOnce(new Error("db down"));
    const errorSpy = vi.spyOn(ElMessage, "error").mockImplementation(() => ({
      close: () => {},
    }));
    const wrapper = mount(DashboardView);
    await flushPromises();
    expect(
      wrapper.find('[data-testid="instance-rejoin-section"]').exists(),
    ).toBe(false);
    expect(errorSpy).toHaveBeenCalledWith("db down");
    errorSpy.mockRestore();
  });

  it("refreshes instance rejoin on activity event", async () => {
    vi.useFakeTimers();
    mockGetInstanceRejoinSection.mockClear();
    mount(DashboardView);
    await flushPromises();
    expect(mockGetInstanceRejoinSection).toHaveBeenCalledTimes(1);
    triggerEncountersChanged();
    await vi.advanceTimersByTimeAsync(400);
    expect(mockGetInstanceRejoinSection).toHaveBeenCalledTimes(2);
    vi.useRealTimers();
  });
});
