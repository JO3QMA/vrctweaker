import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import AutomationItemList from "./AutomationItemList.vue";
import type { AutomationItemDTO } from "../../wails/app";

const rule: AutomationItemDTO = {
  id: "rule-1",
  name: "AFK rule",
  kind: "rule",
  isEnabled: true,
  triggerType: "afk_detected",
  conditionsJson: "[]",
  actionsJson: "[]",
};

function mountList(items: AutomationItemDTO[] = [rule]) {
  return mount(AutomationItemList, {
    props: {
      items,
      activeId: "",
      triggerLabels: { afk_detected: "AFK検出時" },
      disabled: false,
    },
  });
}

describe("AutomationItemList", () => {
  it("creates a rule only for the rule command", async () => {
    const wrapper = mountList();
    const dropdown = wrapper.findComponent({ name: "ElDropdown" });

    await dropdown.vm.$emit("command", "unknown");
    expect(wrapper.emitted("addRule")).toBeUndefined();
    expect(wrapper.emitted("addScript")).toBeUndefined();

    await dropdown.vm.$emit("command", "rule");
    expect(wrapper.emitted("addRule")).toHaveLength(1);

    await dropdown.vm.$emit("command", "script");
    expect(wrapper.emitted("addScript")).toHaveLength(1);
  });

  it("shows the trigger label and never the raw trigger id", () => {
    const wrapper = mountList([
      rule,
      { ...rule, id: "unknown", name: "Unknown", triggerType: "not_a_trigger" },
    ]);
    expect(wrapper.text()).toContain("AFK検出時");
    expect(wrapper.text()).toContain("不明なトリガー");
    expect(wrapper.text()).not.toContain("afk_detected");
    expect(wrapper.text()).not.toContain("not_a_trigger");
  });

  it("keeps the enable switch outside the card's button", async () => {
    const wrapper = mountList();
    const select = wrapper.get(".rule-card__select");
    expect(select.attributes("role")).toBe("button");
    expect(select.findAll(".el-switch")).toHaveLength(0);
    expect(wrapper.find(".rule-card .el-switch").exists()).toBe(true);

    await wrapper.get(".toggle-wrap").trigger("click");
    expect(wrapper.emitted("select")).toBeUndefined();

    await select.trigger("click");
    expect(wrapper.emitted("select")).toHaveLength(1);
  });
});
