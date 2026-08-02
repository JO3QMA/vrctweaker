import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import VtButton from "../VtButton.vue";
import { type VtButtonProps } from "../vtButtonVariants";

describe("VtButton", () => {
  it("maps primary variant to el-button--primary", () => {
    const wrapper = mount(VtButton, {
      props: { variant: "primary" },
      slots: { default: "Save" },
    });
    expect(wrapper.find(".el-button--primary").exists()).toBe(true);
    expect(wrapper.find(".is-text").exists()).toBe(false);
  });

  it("maps secondary variant to default filled button", () => {
    const wrapper = mount(VtButton, {
      props: { variant: "secondary" },
      slots: { default: "Launch" },
    });
    const btn = wrapper.find(".el-button");
    expect(btn.classes()).not.toContain("el-button--primary");
    expect(btn.classes()).not.toContain("is-text");
    expect(btn.classes()).not.toContain("el-button--danger");
  });

  it("maps tertiary variant to text button", () => {
    const wrapper = mount(VtButton, {
      props: { variant: "tertiary" },
      slots: { default: "Reload" },
    });
    expect(wrapper.find(".is-text").exists()).toBe(true);
  });

  it("maps danger variant to el-button--danger", () => {
    const wrapper = mount(VtButton, {
      props: { variant: "danger" },
      slots: { default: "Delete" },
    });
    expect(wrapper.find(".el-button--danger").exists()).toBe(true);
    expect(wrapper.find(".is-plain").exists()).toBe(false);
  });

  it("maps danger plain when plain prop is set", () => {
    const wrapper = mount(VtButton, {
      props: { variant: "danger", plain: true } satisfies VtButtonProps,
      slots: { default: "Logout" },
    });
    expect(wrapper.find(".el-button--danger").exists()).toBe(true);
    expect(wrapper.find(".is-plain").exists()).toBe(true);
  });

  it("forwards attrs such as disabled and data-testid", () => {
    const wrapper = mount(VtButton, {
      props: { variant: "primary" },
      attrs: {
        disabled: true,
        "data-testid": "save-btn",
      },
      slots: { default: "Save" },
    });
    const btn = wrapper.find(".el-button");
    expect(btn.attributes("data-testid")).toBe("save-btn");
    expect(btn.attributes("disabled")).toBeDefined();
  });

  it("forwards loading to el-button", () => {
    const wrapper = mount(VtButton, {
      props: { variant: "primary" },
      attrs: { loading: true },
      slots: { default: "Apply" },
    });
    expect(wrapper.find(".is-loading").exists()).toBe(true);
  });
});
