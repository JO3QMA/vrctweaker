import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import SettingsRow from "./SettingsRow.vue";

describe("SettingsRow", () => {
  it("places the label on the left and the control on the right", () => {
    const wrapper = mount(SettingsRow, {
      props: { label: "表示言語", description: "すぐ反映されます" },
      slots: { default: '<select data-testid="control" />' },
    });

    expect(wrapper.get(".setting-row__label").text()).toBe("表示言語");
    expect(wrapper.get(".setting-row__desc").text()).toBe("すぐ反映されます");
    expect(wrapper.find("[data-testid='control']").exists()).toBe(true);
    expect(wrapper.classes()).not.toContain("setting-row--fixed");
    expect(wrapper.classes()).not.toContain("setting-row--center");
  });

  it("keeps switches and number inputs on a non-wrapping row", () => {
    const sw = mount(SettingsRow, {
      props: { label: "トレイ", control: "switch" },
    });
    const num = mount(SettingsRow, {
      props: { label: "日数", control: "number" },
    });

    expect(sw.classes()).toContain("setting-row--fixed");
    expect(sw.classes()).not.toContain("setting-row--number");
    expect(num.classes()).toContain("setting-row--fixed");
    expect(num.classes()).toContain("setting-row--number");
  });

  it("keeps a short value on the right even when the panel is narrow", () => {
    const wrapper = mount(SettingsRow, {
      props: { label: "バージョン", control: "value" },
      slots: { default: "<span>0.1.0</span>" },
    });
    expect(wrapper.classes()).toContain("setting-row--fixed");
    expect(wrapper.classes()).toContain("setting-row--center");
  });

  it("marks button rows so wrapped buttons can align left", () => {
    const wrapper = mount(SettingsRow, {
      props: { label: "更新", control: "button" },
    });
    expect(wrapper.classes()).toContain("setting-row--button");
    expect(wrapper.classes()).not.toContain("setting-row--fixed");
    expect(wrapper.classes()).toContain("setting-row--center");
  });
});
