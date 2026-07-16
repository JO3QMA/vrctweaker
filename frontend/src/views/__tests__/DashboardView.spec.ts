import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { ElMessage } from "element-plus";
import DashboardView from "../DashboardView.vue";

const { mockSetStatus, mockSetStatusDescription, mockSetStatusAndDescription } =
  vi.hoisted(() => ({
    mockSetStatus: vi.fn(),
    mockSetStatusDescription: vi.fn(),
    mockSetStatusAndDescription: vi.fn(),
  }));

vi.mock("../../components/DashboardLaunchBlock.vue", () => ({
  default: {
    name: "DashboardLaunchBlock",
    template: '<div data-testid="dashboard-launch-block-stub" />',
  },
}));

vi.mock("../../wails/app", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../../wails/app")>();
  return {
    ...actual,
    App: {
      ...actual.App,
      setStatus: mockSetStatus,
      setStatusDescription: mockSetStatusDescription,
      setStatusAndDescription: mockSetStatusAndDescription,
    },
  };
});

describe("DashboardView", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSetStatus.mockResolvedValue(undefined);
    mockSetStatusDescription.mockResolvedValue(undefined);
    mockSetStatusAndDescription.mockResolvedValue(undefined);
  });

  it("renders launch block stub and page title", async () => {
    const wrapper = mount(DashboardView);
    await flushPromises();
    expect(wrapper.find(".page-title").text()).toBe("ダッシュボード");
    expect(
      wrapper.find('[data-testid="dashboard-launch-block-stub"]').exists(),
    ).toBe(true);
  });

  it("renders quick status buttons", async () => {
    const wrapper = mount(DashboardView);
    await flushPromises();
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
});
