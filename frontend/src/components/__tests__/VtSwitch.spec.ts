import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import VtSwitch from "../VtSwitch.vue";

describe("VtSwitch", () => {
  it("renders el-switch", () => {
    const wrapper = mount(VtSwitch, {
      attrs: { modelValue: true },
    });
    expect(wrapper.find(".el-switch").exists()).toBe(true);
    expect(wrapper.find(".is-checked").exists()).toBe(true);
  });

  it("forwards disabled and data-testid", () => {
    const wrapper = mount(VtSwitch, {
      attrs: {
        disabled: true,
        "data-testid": "item-enabled",
      },
    });
    const sw = wrapper.find(".el-switch");
    expect(sw.attributes("data-testid")).toBe("item-enabled");
    expect(sw.classes()).toContain("is-disabled");
  });

  it("forwards size small", () => {
    const wrapper = mount(VtSwitch, {
      attrs: { size: "small" },
    });
    expect(wrapper.find(".el-switch--small").exists()).toBe(true);
  });
});
