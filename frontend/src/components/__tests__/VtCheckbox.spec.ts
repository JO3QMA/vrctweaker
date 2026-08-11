import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import VtCheckbox from "../VtCheckbox.vue";

describe("VtCheckbox", () => {
  it("renders el-checkbox", () => {
    const wrapper = mount(VtCheckbox, {
      slots: { default: "Desktop mode" },
    });
    expect(wrapper.find(".el-checkbox").exists()).toBe(true);
    expect(wrapper.text()).toContain("Desktop mode");
  });

  it("forwards disabled and data-testid", () => {
    const wrapper = mount(VtCheckbox, {
      attrs: {
        disabled: true,
        "data-testid": "opt-desktop",
      },
      slots: { default: "Option" },
    });
    const checkbox = wrapper.find(".el-checkbox");
    expect(checkbox.attributes("data-testid")).toBe("opt-desktop");
    expect(checkbox.classes()).toContain("is-disabled");
  });
});
