import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import VtAlert from "../VtAlert.vue";

describe("VtAlert", () => {
  it("maps danger variant to el-alert--error", () => {
    const wrapper = mount(VtAlert, {
      props: { variant: "danger", title: "Failed" },
    });
    expect(wrapper.find(".el-alert--error").exists()).toBe(true);
  });

  it("maps success variant to el-alert--success", () => {
    const wrapper = mount(VtAlert, {
      props: { variant: "success", title: "OK" },
    });
    expect(wrapper.find(".el-alert--success").exists()).toBe(true);
  });

  it("defaults to show-icon and not closable", () => {
    const wrapper = mount(VtAlert, {
      props: { variant: "info", title: "Hint" },
    });
    expect(wrapper.find(".el-alert__icon").exists()).toBe(true);
    expect(wrapper.find(".el-alert__close-btn").exists()).toBe(false);
  });

  it("forwards description and data-testid", () => {
    const wrapper = mount(VtAlert, {
      props: {
        variant: "warning",
        title: "Warn",
        description: "Details",
      },
      attrs: { "data-testid": "block-alert" },
    });
    expect(wrapper.text()).toContain("Details");
    expect(wrapper.attributes("data-testid")).toBe("block-alert");
  });

  it("forwards default slot content", () => {
    const wrapper = mount(VtAlert, {
      props: { variant: "info", title: "Hint" },
      slots: { default: "<strong>Extra</strong>" },
    });
    expect(wrapper.find("strong").text()).toBe("Extra");
  });
});
