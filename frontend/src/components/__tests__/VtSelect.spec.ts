import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import VtSelect from "../VtSelect.vue";

describe("VtSelect", () => {
  it("renders el-select", () => {
    const wrapper = mount(VtSelect, {
      slots: {
        default: '<el-option label="A" value="a" />',
      },
    });
    expect(wrapper.find(".el-select").exists()).toBe(true);
  });

  it("forwards disabled and data-testid", () => {
    const wrapper = mount(VtSelect, {
      attrs: {
        disabled: true,
        "data-testid": "language-select",
      },
      slots: {
        default: '<el-option label="JA" value="ja" />',
      },
    });
    const select = wrapper.find(".el-select");
    expect(select.attributes("data-testid")).toBe("language-select");
    expect(wrapper.find(".is-disabled").exists()).toBe(true);
  });
});
