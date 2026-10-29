import { describe, expect, it } from "vitest";
import { defineComponent, ref } from "vue";
import { flushPromises, mount } from "@vue/test-utils";
import AutomationRuleEditor from "./AutomationRuleEditor.vue";
import { dtoToEditor } from "../automationEditorMapping";
import type { AutomationItemDTO } from "../../wails/app";

const item: AutomationItemDTO = {
  id: "rule-1",
  name: "Saved rule",
  kind: "rule",
  isEnabled: true,
  triggerType: "friend_joined",
  conditionsJson: "[]",
  actionsJson: JSON.stringify([
    { type: "change_status", payload: { status: "busy" } },
  ]),
};

const Harness = defineComponent({
  components: { AutomationRuleEditor },
  setup() {
    const editor = ref(dtoToEditor(item));
    return { editor };
  },
  template: `
    <AutomationRuleEditor
      v-model:editor="editor"
      :saving="false"
      :is-dirty="false"
      :disabled="false"
      name-error=""
      :trigger-options="[{ value: 'friend_joined', label: 'フレンド参加時' }]"
      :status-options="[]"
      :power-plan-presets="[]"
      :weekday-options="[]"
      :action-options="[]"
      :power-plans="[]"
      :friends="[]"
    />
  `,
});

describe("AutomationRuleEditor", () => {
  it("updates the parent editor through v-model", async () => {
    const wrapper = mount(Harness);
    const input = wrapper.get('[data-testid="item-name-input"]');
    await input.setValue("Renamed");
    expect(
      wrapper.getComponent(AutomationRuleEditor).props("editor"),
    ).toMatchObject({ name: "Renamed" });
  });

  it("emits delete only for the delete command", async () => {
    const wrapper = mount(Harness);
    const dropdowns = wrapper.findAllComponents({ name: "ElDropdown" });
    const menu = dropdowns[dropdowns.length - 1];
    expect(menu).toBeTruthy();
    await menu!.vm.$emit("command", "rename");
    expect(
      wrapper.findComponent(AutomationRuleEditor).emitted("delete"),
    ).toBeUndefined();
    await menu!.vm.$emit("command", "delete");
    expect(
      wrapper.findComponent(AutomationRuleEditor).emitted("delete"),
    ).toHaveLength(1);
  });

  it("labels an unknown trigger and keeps that option disabled", async () => {
    const unknown = dtoToEditor({
      ...item,
      triggerType: "future.unknown_trigger",
    });
    const wrapper = mount(AutomationRuleEditor, {
      props: {
        editor: unknown,
        saving: false,
        isDirty: false,
        disabled: false,
        nameError: "",
        triggerOptions: [{ value: "friend_joined", label: "フレンド参加時" }],
        statusOptions: [],
        powerPlanPresets: [],
        weekdayOptions: [],
        actionOptions: [],
        powerPlans: [],
        friends: [],
      },
    });

    await flushPromises();
    const option = document.body.querySelector(
      ".el-select-dropdown__item.is-disabled",
    );
    expect(option?.textContent).toContain("不明なトリガー");
    expect(document.body.textContent).not.toContain("future.unknown_trigger");
    const selected = wrapper.get(".el-select__placeholder");
    expect(selected.text()).toContain("不明なトリガー");
  });
});
