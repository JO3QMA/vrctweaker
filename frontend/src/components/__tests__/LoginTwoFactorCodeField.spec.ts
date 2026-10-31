import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import LoginTwoFactorCodeField from "../LoginTwoFactorCodeField.vue";

describe("LoginTwoFactorCodeField", () => {
  it("emits sanitized digits and completes when pasting a formatted code", async () => {
    const wrapper = mount(LoginTwoFactorCodeField, {
      props: { modelValue: "" },
    });

    await wrapper.find("input").trigger("paste", {
      clipboardData: { getData: () => "12 34-56" },
    });

    expect(wrapper.emitted("update:modelValue")).toEqual([["123456"]]);
    expect(wrapper.emitted("complete")).toHaveLength(1);
  });

  it("does not emit complete until six digits are present", async () => {
    const wrapper = mount(LoginTwoFactorCodeField, {
      props: { modelValue: "" },
    });

    await wrapper.find("input").setValue("12ab34");

    const values = wrapper.emitted("update:modelValue") ?? [];
    expect(values[values.length - 1]).toEqual(["1234"]);
    expect(wrapper.emitted("complete")).toBeUndefined();
  });
});
