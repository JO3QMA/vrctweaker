import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import VrcStatusTag from "../VrcStatusTag.vue";

describe("VrcStatusTag", () => {
  it("shows API text and el-tag type for join me", () => {
    const wrapper = mount(VrcStatusTag, { props: { status: "join me" } });
    expect(wrapper.text()).toContain("join me");
    expect(wrapper.find(".el-tag--success").exists()).toBe(true);
  });

  it("uses primary for active (no empty ElTag type)", () => {
    const wrapper = mount(VrcStatusTag, { props: { status: "active" } });
    expect(wrapper.text()).toContain("active");
    expect(wrapper.find(".el-tag--primary").exists()).toBe(true);
  });

  it("shows em dash when status empty", () => {
    const wrapper = mount(VrcStatusTag, { props: { status: "" } });
    expect(wrapper.text()).toContain("—");
    expect(wrapper.find(".el-tag--info").exists()).toBe(true);
  });
});
