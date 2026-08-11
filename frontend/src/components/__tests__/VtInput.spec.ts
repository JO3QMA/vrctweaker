import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import VtInput from "../VtInput.vue";

describe("VtInput", () => {
  it("renders el-input", () => {
    const wrapper = mount(VtInput, {
      attrs: { placeholder: "Name" },
    });
    expect(wrapper.find(".el-input").exists()).toBe(true);
    expect(wrapper.find("input").attributes("placeholder")).toBe("Name");
  });

  it("forwards disabled and data-testid", () => {
    const wrapper = mount(VtInput, {
      attrs: {
        disabled: true,
        "data-testid": "profile-name",
      },
    });
    const input = wrapper.find("input");
    expect(input.attributes("data-testid")).toBe("profile-name");
    expect(input.attributes("disabled")).toBeDefined();
  });

  it("forwards prefix slot", () => {
    const wrapper = mount(VtInput, {
      slots: {
        prefix: '<span class="prefix-icon">*</span>',
      },
    });
    expect(wrapper.find(".prefix-icon").exists()).toBe(true);
  });
});
