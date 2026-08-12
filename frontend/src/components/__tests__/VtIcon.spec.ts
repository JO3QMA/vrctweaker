import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import { Search } from "@element-plus/icons-vue";
import VtIcon from "../VtIcon.vue";
import { VT_ICON_SIZES } from "../../design/iconTokens";

describe("VtIcon", () => {
  it.each(VT_ICON_SIZES)("applies size class for %s", (size) => {
    const wrapper = mount(VtIcon, {
      props: { size },
      slots: { default: Search },
    });
    expect(wrapper.find(`.vt-icon--size-${size}`).exists()).toBe(true);
  });

  it("renders el-icon wrapper with slotted icon", () => {
    const wrapper = mount(VtIcon, {
      props: { size: "default" },
      slots: { default: Search },
    });
    expect(wrapper.find(".el-icon").exists()).toBe(true);
  });

  it("defaults decorative icons to aria-hidden", () => {
    const wrapper = mount(VtIcon, {
      props: { size: "default" },
      slots: { default: Search },
    });
    expect(wrapper.find(".el-icon").attributes("aria-hidden")).toBe("true");
  });

  it("passes through aria-hidden when decorative is false", () => {
    const wrapper = mount(VtIcon, {
      props: { size: "default", decorative: false },
      attrs: { "aria-hidden": "true" },
      slots: { default: Search },
    });
    expect(wrapper.find(".el-icon").attributes("aria-hidden")).toBe("true");
  });

  it("forwards attrs such as data-testid", () => {
    const wrapper = mount(VtIcon, {
      props: { size: "compact" },
      attrs: { "data-testid": "search-icon" },
      slots: { default: Search },
    });
    expect(wrapper.find(".el-icon").attributes("data-testid")).toBe(
      "search-icon",
    );
  });
});
