import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import SettingsPageHead from "./SettingsPageHead.vue";

describe("SettingsPageHead", () => {
  it("shows the page title and one-line description", () => {
    const wrapper = mount(SettingsPageHead, {
      props: { title: "データ", description: "参照先と保存期間です。" },
    });
    expect(wrapper.get("h2").text()).toBe("データ");
    expect(wrapper.get("p").text()).toBe("参照先と保存期間です。");
  });
});
