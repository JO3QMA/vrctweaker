import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { createI18n } from "vue-i18n";
import ja from "../../i18n/locales/ja.json";
import VrcInstanceKeyLabel from "../VrcInstanceKeyLabel.vue";

function mountLabel(props: { instanceKey?: string } = {}) {
  const i18n = createI18n({
    legacy: false,
    locale: "ja",
    messages: { ja },
  });
  return mount(VrcInstanceKeyLabel, {
    props,
    global: { plugins: [i18n] },
  });
}

describe("VrcInstanceKeyLabel", () => {
  it("shows dash when instanceKey is undefined", () => {
    const wrapper = mountLabel();
    expect(wrapper.text()).toBe(ja.common.dash);
  });

  it("shows dash when instanceKey is empty or whitespace only", () => {
    for (const instanceKey of ["", "   ", "\t"]) {
      const wrapper = mountLabel({ instanceKey });
      expect(wrapper.text()).toBe(ja.common.dash);
    }
  });

  it("shows trimmed legacy id as mono fallback when not a wrld key", () => {
    const wrapper = mountLabel({ instanceKey: "  inst_legacy_001  " });
    expect(wrapper.find(".mono").text()).toBe("inst_legacy_001");
  });

  it("renders labeled text for a parseable wrld instance key", () => {
    const instanceKey = "wrld_w:88577~region(jp)";
    const wrapper = mountLabel({ instanceKey });
    const publicLabel = ja.vrc.instanceType.public;
    expect(wrapper.text()).toContain(`${publicLabel} #88577`);
    expect(wrapper.text()).toContain("[JP]");
  });
});
