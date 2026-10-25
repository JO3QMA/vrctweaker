import { describe, expect, it } from "vitest";
import { defineComponent, ref } from "vue";
import { mount } from "@vue/test-utils";
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
      :trigger-options="[]"
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
});
