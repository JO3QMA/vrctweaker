import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { ElMessageBox } from "element-plus";
import AutomationView from "../AutomationView.vue";
import AutomationItemList from "../automation/AutomationItemList.vue";
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
  mockDeleteAutomationItem,
  mockGetAutomationRunLog,
  mockGetAutomationRuntimeStatus,
  mockListDetectedPowerPlans,
  mockFriends,
} = vi.hoisted(() => ({
  mockListAutomationItems: vi.fn(),
  mockSaveAutomationItem: vi.fn(),
  mockToggleAutomationItem: vi.fn(),
  mockDeleteAutomationItem: vi.fn(),
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
      deleteAutomationItem: mockDeleteAutomationItem,
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

async function chooseCreate(
  wrapper: ReturnType<typeof mount>,
  command: "rule" | "script",
) {
  const dropdown = wrapper
    .getComponent(AutomationItemList)
    .findComponent({ name: "ElDropdown" });
  await dropdown.vm.$emit("command", command);
  await flushPromises();
}

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

    await chooseCreate(wrapper, "rule");
    await flushPromises();

    const nameInput = wrapper.get('[data-testid="item-name-input"]');
    await nameInput.setValue("Draft rule");
    await flushPromises();

    expect(wrapper.find('[data-testid="unsaved-banner"]').exists()).toBe(true);

    mockSaveAutomationItem.mockRejectedValueOnce(new Error("save failed"));
    const confirmSpy = vi
      .spyOn(ElMessageBox, "confirm")
      .mockResolvedValue("confirm" as never);

    await wrapper.find(".rule-card__select").trigger("click");
    await flushPromises();

    expect(confirmSpy).toHaveBeenCalled();
    expect(mockSaveAutomationItem).toHaveBeenCalled();
    expect(showToast.error).toHaveBeenCalled();
    // Still editing the draft (did not switch to seeded card editor).
    expect(wrapper.find(".rule-editor").exists()).toBe(true);
    expect(
      (
        wrapper.get('[data-testid="item-name-input"]')
          .element as HTMLInputElement
      ).value,
    ).toBe("Draft rule");

    confirmSpy.mockRestore();
  });

  it("does not treat list refresh failure as save failure", async () => {
    const wrapper = mount(AutomationView);
    await flushPromises();

    await chooseCreate(wrapper, "rule");
    await flushPromises();
    await wrapper.get('[data-testid="item-name-input"]').setValue("Saved ok");
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

    await wrapper.findAll(".rule-card__select")[0]?.trigger("click");
    await flushPromises();
    expect(wrapper.find(".rule-editor").exists()).toBe(true);

    await wrapper.findAll(".rule-card__select")[1]?.trigger("click");
    await flushPromises();

    expect(showToast.error).toHaveBeenCalled();
    expect(wrapper.find(".rule-editor").exists()).toBe(false);
  });

  it("does not keep a generated id on the editor when save fails", async () => {
    const wrapper = mount(AutomationView);
    await flushPromises();

    await chooseCreate(wrapper, "rule");
    await flushPromises();
    await wrapper.get('[data-testid="item-name-input"]').setValue("Will fail");
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

describe("AutomationView layout", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockListAutomationItems.mockResolvedValue([seeded]);
    mockGetAutomationRunLog.mockResolvedValue([
      {
        at: "2026-10-21 21:00",
        itemId: "rule_1",
        itemName: "Seeded",
        eventType: "friend_joined",
        success: true,
        actionsCompleted: 1,
        actionsTotal: 1,
        contextLabel: "参加",
      },
    ]);
    mockGetAutomationRuntimeStatus.mockResolvedValue({ available: true });
    mockListDetectedPowerPlans.mockResolvedValue([]);
    mockFriends.mockResolvedValue([]);
    mockSaveAutomationItem.mockResolvedValue(undefined);
  });

  it("shows rule and run-log tabs with counts", async () => {
    const wrapper = mount(AutomationView);
    await flushPromises();

    const rules = wrapper.get('[data-testid="automation-tab-rules"]');
    const log = wrapper.get('[data-testid="automation-tab-run-log"]');
    expect(rules.text()).toContain("ルール");
    expect(wrapper.get('[data-testid="automation-rules-count"]').text()).toBe(
      "1",
    );
    expect(log.text()).toContain("実行ログ");
    expect(wrapper.get('[data-testid="automation-log-count"]').text()).toBe(
      "1",
    );
    expect(rules.attributes("aria-selected")).toBe("true");
    expect(log.attributes("aria-selected")).toBe("false");
  });

  it("moves the run log into its own tab", async () => {
    const wrapper = mount(AutomationView);
    await flushPromises();

    expect(wrapper.find("#automation-panel-rules").exists()).toBe(true);
    expect(wrapper.find("#automation-panel-log").exists()).toBe(true);
    expect(
      wrapper.get("#automation-panel-log").attributes("style") ?? "",
    ).toContain("display: none");

    await wrapper
      .get('[data-testid="automation-tab-run-log"]')
      .trigger("click");
    expect(
      wrapper.get("#automation-panel-rules").attributes("style") ?? "",
    ).toContain("display: none");
    expect(
      wrapper.get("#automation-panel-log").attributes("style") ?? "",
    ).not.toContain("display: none");
    expect(wrapper.text()).toContain("フレンド参加時");
    expect(wrapper.text()).not.toContain("friend_joined");

    await wrapper.get('[data-testid="refresh-run-log"]').trigger("click");
    await flushPromises();
    expect(mockGetAutomationRunLog).toHaveBeenCalledTimes(2);
  });

  it("groups the selected rule into sections and a bottom action bar", async () => {
    const wrapper = mount(AutomationView);
    await flushPromises();

    await wrapper.find(".rule-card__select").trigger("click");
    await flushPromises();

    expect(
      wrapper.get('[data-testid="automation-section-when"]').text(),
    ).toContain("いつ");
    expect(
      wrapper.get('[data-testid="automation-section-if"]').text(),
    ).toContain("もし");
    expect(
      wrapper.get('[data-testid="automation-section-then"]').text(),
    ).toContain("したら");
    const bar = wrapper.get('[data-testid="automation-action-bar"]');
    expect(bar.find('[data-testid="save-item"]').exists()).toBe(true);
    expect(bar.find('[data-testid="cancel-item"]').exists()).toBe(true);
    expect(bar.find('[data-testid="delete-item"]').exists()).toBe(true);
    expect(bar.classes()).toContain("automation-action-bar");
  });

  it("shows an empty-name error under the field and does not save", async () => {
    const wrapper = mount(AutomationView);
    await flushPromises();

    await chooseCreate(wrapper, "rule");
    await flushPromises();
    await wrapper.find('[data-testid="save-item"]').trigger("click");
    await flushPromises();

    const error = wrapper.get('[data-testid="name-error"]');
    expect(error.text()).toContain("名前を入力してください");
    const nameField = wrapper.get('[data-testid="item-name-input"]');
    expect(
      nameField.element.compareDocumentPosition(error.element) &
        Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
    expect(mockSaveAutomationItem).not.toHaveBeenCalled();
    expect(showToast.error).not.toHaveBeenCalled();
  });

  it("disables list and editor controls when automation is unavailable", async () => {
    mockGetAutomationRuntimeStatus.mockResolvedValue({
      available: false,
      reasonKey: "subsystemUnavailable",
    });
    const wrapper = mount(AutomationView);
    await flushPromises();

    expect(
      wrapper.get('[data-testid="add-item"]').attributes("disabled"),
    ).toBeDefined();
    expect(wrapper.get(".rule-card .el-switch").classes()).toContain(
      "is-disabled",
    );

    await wrapper.find(".rule-card__select").trigger("click");
    await flushPromises();

    expect(
      wrapper.get('[data-testid="save-item"]').attributes("disabled"),
    ).toBe("");
    expect(
      (
        wrapper.get('[data-testid="item-name-input"]')
          .element as HTMLInputElement
      ).disabled,
    ).toBe(true);
    expect(wrapper.get('[data-testid="editor-enabled"]').classes()).toContain(
      "is-disabled",
    );
  });

  it("toasts when refreshing the run log fails", async () => {
    const wrapper = mount(AutomationView);
    await flushPromises();
    await wrapper
      .get('[data-testid="automation-tab-run-log"]')
      .trigger("click");
    mockGetAutomationRunLog.mockRejectedValueOnce(new Error("log failed"));

    await wrapper.get('[data-testid="refresh-run-log"]').trigger("click");
    await flushPromises();

    expect(showToast.error).toHaveBeenCalledWith(
      "実行ログを取得できませんでした",
    );
  });

  it("moves between tabs with the arrow keys", async () => {
    const wrapper = mount(AutomationView);
    await flushPromises();
    const rules = wrapper.get('[data-testid="automation-tab-rules"]');
    expect(rules.attributes("tabindex")).toBe("0");
    expect(
      wrapper
        .get('[data-testid="automation-tab-run-log"]')
        .attributes("tabindex"),
    ).toBe("-1");

    await rules.trigger("keydown", { key: "ArrowRight" });
    await flushPromises();

    expect(
      wrapper
        .get('[data-testid="automation-tab-run-log"]')
        .attributes("aria-selected"),
    ).toBe("true");
    expect(
      wrapper
        .get('[data-testid="automation-tab-run-log"]')
        .attributes("tabindex"),
    ).toBe("0");
    expect(
      wrapper.get("#automation-panel-log").attributes("style") ?? "",
    ).not.toContain("display: none");
  });

  it("asks before leaving the editor for the list when there are unsaved changes", async () => {
    const wrapper = mount(AutomationView);
    await flushPromises();
    await chooseCreate(wrapper, "rule");
    await flushPromises();
    await wrapper.get('[data-testid="item-name-input"]').setValue("Draft rule");
    await flushPromises();

    const confirmSpy = vi
      .spyOn(ElMessageBox, "confirm")
      .mockRejectedValueOnce("close");
    await wrapper.get('[data-testid="automation-back"]').trigger("click");
    await flushPromises();

    expect(confirmSpy).toHaveBeenCalled();
    expect(wrapper.find(".rule-editor").isVisible()).toBe(true);
    expect(
      (
        wrapper.get('[data-testid="item-name-input"]')
          .element as HTMLInputElement
      ).value,
    ).toBe("Draft rule");
    confirmSpy.mockRestore();
  });

  it("moves to the first and last tab with Home and End", async () => {
    const wrapper = mount(AutomationView);
    await flushPromises();
    const log = wrapper.get('[data-testid="automation-tab-run-log"]');
    await log.trigger("keydown", { key: "End" });
    await flushPromises();
    expect(log.attributes("aria-selected")).toBe("true");

    await log.trigger("keydown", { key: "Home" });
    await flushPromises();
    expect(
      wrapper
        .get('[data-testid="automation-tab-rules"]')
        .attributes("aria-selected"),
    ).toBe("true");
  });

  it("labels an afk_detected rule with the AFK trigger name", async () => {
    mockListAutomationItems.mockResolvedValue([
      { ...seeded, triggerType: "afk_detected" },
    ]);
    const wrapper = mount(AutomationView);
    await flushPromises();

    expect(wrapper.text()).toContain("AFK検出時");
    expect(wrapper.text()).not.toContain("afk_detected");

    await wrapper.find(".rule-card__select").trigger("click");
    await flushPromises();
    expect(wrapper.get('[data-testid="rule-editor"]').text()).toContain(
      "AFK検出時",
    );
  });

  it("toasts and keeps the editor open when delete fails", async () => {
    mockDeleteAutomationItem.mockRejectedValueOnce(new Error("delete failed"));
    const wrapper = mount(AutomationView);
    await flushPromises();
    await wrapper.find(".rule-card__select").trigger("click");
    await flushPromises();

    const confirmSpy = vi
      .spyOn(ElMessageBox, "confirm")
      .mockResolvedValue("confirm" as never);
    await wrapper.get('[data-testid="delete-item"]').trigger("click");
    await flushPromises();

    expect(showToast.error).toHaveBeenCalledWith(
      "ルールを削除できませんでした",
    );
    expect(wrapper.find('[data-testid="rule-editor"]').exists()).toBe(true);
    expect(mockDeleteAutomationItem).toHaveBeenCalledWith("rule_1");
    confirmSpy.mockRestore();
  });
});
