import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import VtTag from "../VtTag.vue";

describe("VtTag", () => {
  it("maps danger variant to el-tag--danger", () => {
    const wrapper = mount(VtTag, {
      props: { variant: "danger" },
      slots: { default: "Failed" },
    });
    expect(wrapper.find(".el-tag--danger").exists()).toBe(true);
  });

  it("maps neutral variant to neutral styling class", () => {
    const wrapper = mount(VtTag, {
      props: { variant: "neutral" },
      slots: { default: "3" },
    });
    expect(wrapper.find(".vt-tag--neutral").exists()).toBe(true);
    expect(wrapper.find(".el-tag--primary").exists()).toBe(false);
    expect(wrapper.find(".el-tag--success").exists()).toBe(false);
  });

  it("strips conflicting type attr when variant is neutral", () => {
    const wrapper = mount(VtTag, {
      props: { variant: "neutral" },
      attrs: { type: "danger" },
      slots: { default: "3" },
    });
    expect(wrapper.find(".el-tag--danger").exists()).toBe(false);
    expect(wrapper.find(".vt-tag--neutral").exists()).toBe(true);
  });

  it("maps primary variant to el-tag--primary", () => {
    const wrapper = mount(VtTag, {
      props: { variant: "primary" },
      slots: { default: "Default" },
    });
    expect(wrapper.find(".el-tag--primary").exists()).toBe(true);
  });

  it("forwards size and data-testid", () => {
    const wrapper = mount(VtTag, {
      props: { variant: "success", size: "small" },
      attrs: { "data-testid": "status-tag" },
      slots: { default: "OK" },
    });
    expect(wrapper.attributes("data-testid")).toBe("status-tag");
    expect(wrapper.find(".el-tag--small").exists()).toBe(true);
  });
});
