import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { ElMessageBox } from "element-plus";
import AutomationView from "../AutomationView.vue";
import type { AutomationItemDTO } from "../../wails/app";
import { showToast } from "../../utils/showToast";

vi.mock("../../utils/showToast", () => ({
  showToast: {
    success: vi.fn(),
    warning: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
  },
}));

const {
  mockListAutomationItems,
  mockSaveAutomationItem,
  mockToggleAutomationItem,
  mockGetAutomationRunLog,
  mockGetAutomationRuntimeStatus,
  mockListDetectedPowerPlans,
  mockFriends,
} = vi.hoisted(() => ({
  mockListAutomationItems: vi.fn(),
  mockSaveAutomationItem: vi.fn(),
  mockToggleAutomationItem: vi.fn(),
  mockGetAutomationRunLog: vi.fn(),
  mockGetAutomationRuntimeStatus: vi.fn(),
  mockListDetectedPowerPlans: vi.fn(),
  mockFriends: vi.fn(),
}));

vi.mock("../../wails/app", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../../wails/app")>();
  return {
    ...actual,
    App: {
      listAutomationItems: mockListAutomationItems,
      saveAutomationItem: mockSaveAutomationItem,
      toggleAutomationItem: mockToggleAutomationItem,
      deleteAutomationItem: vi.fn(),
      getAutomationRunLog: mockGetAutomationRunLog,
      getAutomationRuntimeStatus: mockGetAutomationRuntimeStatus,
      listDetectedPowerPlans: mockListDetectedPowerPlans,
      friends: mockFriends,
    },
  };
});

vi.mock("../../wails/runtime", () => ({
  getRuntime: () => undefined,
}));

const seeded: AutomationItemDTO = {
  id: "rule_1",
  name: "Seeded",
  kind: "rule",
  isEnabled: true,
  triggerType: "friend_joined",
  conditionsJson: "[]",
  actionsJson: JSON.stringify([
    { type: "change_status", payload: { status: "busy" } },
  ]),
};

describe("AutomationView unsaved guard", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockListAutomationItems.mockResolvedValue([seeded]);
    mockGetAutomationRunLog.mockResolvedValue([]);
    mockGetAutomationRuntimeStatus.mockResolvedValue({ available: true });
    mockListDetectedPowerPlans.mockResolvedValue([]);
    mockFriends.mockResolvedValue([]);
    mockSaveAutomationItem.mockResolvedValue(undefined);
  });

  it("blocks switching items when save-and-continue fails", async () => {
    const wrapper = mount(AutomationView);
    await flushPromises();

    await wrapper.find('[data-testid="add-rule"]').trigger("click");
    await flushPromises();

    const nameInput = wrapper.find(".rule-editor input");
    await nameInput.setValue("Draft rule");
    await flushPromises();

    expect(wrapper.find('[data-testid="unsaved-banner"]').exists()).toBe(true);

    mockSaveAutomationItem.mockRejectedValueOnce(new Error("save failed"));
    const confirmSpy = vi
      .spyOn(ElMessageBox, "confirm")
      .mockResolvedValue("confirm" as never);

    await wrapper.find(".rule-card").trigger("click");
    await flushPromises();

    expect(confirmSpy).toHaveBeenCalled();
    expect(mockSaveAutomationItem).toHaveBeenCalled();
    expect(showToast.error).toHaveBeenCalled();
    // Still editing the draft (did not switch to seeded card editor).
    expect(wrapper.find(".rule-editor").exists()).toBe(true);
    expect(
      (wrapper.find(".rule-editor input").element as HTMLInputElement).value,
    ).toBe("Draft rule");

    confirmSpy.mockRestore();
  });

  it("does not treat list refresh failure as save failure", async () => {
    const wrapper = mount(AutomationView);
    await flushPromises();

    await wrapper.find('[data-testid="add-rule"]').trigger("click");
    await flushPromises();
    await wrapper.find(".rule-editor input").setValue("Saved ok");
    await flushPromises();

    mockListAutomationItems.mockRejectedValueOnce(new Error("list failed"));

    await wrapper.find('[data-testid="save-item"]').trigger("click");
    await flushPromises();

    expect(mockSaveAutomationItem).toHaveBeenCalled();
    expect(showToast.error).not.toHaveBeenCalled();
    expect(wrapper.find('[data-testid="unsaved-banner"]').exists()).toBe(false);
  });

  it("clears the editor when selecting an item with invalid JSON", async () => {
    const broken: AutomationItemDTO = {
      id: "broken_1",
      name: "Broken",
      kind: "rule",
      isEnabled: true,
      triggerType: "friend_joined",
      conditionsJson: "[]",
      actionsJson: "null",
    };
    mockListAutomationItems.mockResolvedValue([seeded, broken]);

    const wrapper = mount(AutomationView);
    await flushPromises();

    await wrapper.findAll(".rule-card")[0]?.trigger("click");
    await flushPromises();
    expect(wrapper.find(".rule-editor").exists()).toBe(true);

    await wrapper.findAll(".rule-card")[1]?.trigger("click");
    await flushPromises();

    expect(showToast.error).toHaveBeenCalled();
    expect(wrapper.find(".rule-editor").exists()).toBe(false);
  });

  it("does not keep a generated id on the editor when save fails", async () => {
    const wrapper = mount(AutomationView);
    await flushPromises();

    await wrapper.find('[data-testid="add-rule"]').trigger("click");
    await flushPromises();
    await wrapper.find(".rule-editor input").setValue("Will fail");
    await flushPromises();

    mockSaveAutomationItem.mockRejectedValueOnce(new Error("save failed"));

    await wrapper.find('[data-testid="save-item"]').trigger("click");
    await flushPromises();

    const firstId = (
      mockSaveAutomationItem.mock.calls[0]?.[0] as AutomationItemDTO
    ).id;
    expect(firstId).toBeTruthy();
    expect(showToast.error).toHaveBeenCalled();
    expect(wrapper.find('[data-testid="unsaved-banner"]').exists()).toBe(true);

    mockSaveAutomationItem.mockResolvedValueOnce(undefined);
    mockListAutomationItems.mockResolvedValueOnce([
      {
        ...seeded,
        id: "retry-id",
        name: "Will fail",
      },
    ]);
    await wrapper.find('[data-testid="save-item"]').trigger("click");
    await flushPromises();

    const secondId = (
      mockSaveAutomationItem.mock.calls[1]?.[0] as AutomationItemDTO
    ).id;
    // Retry must mint a new id (editor did not keep the failed attempt's UUID).
    expect(secondId).toBeTruthy();
    expect(secondId).not.toBe(firstId);
  });
});
