import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { createI18n } from "vue-i18n";
import ja from "../../i18n/locales/ja.json";
import EncounterHistoryInstanceCell from "../EncounterHistoryInstanceCell.vue";

function mountCell(props: { instanceId?: string } = {}) {
  const i18n = createI18n({
    legacy: false,
    locale: "ja",
    messages: { ja },
  });
  return mount(EncounterHistoryInstanceCell, {
    props,
    global: { plugins: [i18n] },
  });
}

describe("EncounterHistoryInstanceCell", () => {
  it("shows dash when instanceId is undefined", () => {
    const wrapper = mountCell();
    expect(wrapper.text()).toBe(ja.common.dash);
    expect(
      wrapper.find('[data-testid="encounter-instance-link"]').exists(),
    ).toBe(false);
  });

  it("shows dash when instanceId is empty or whitespace only", () => {
    for (const instanceId of ["", "   ", "\t"]) {
      const wrapper = mountCell({ instanceId });
      expect(wrapper.text()).toBe(ja.common.dash);
    }
  });

  it("shows trimmed legacy id as mono fallback when not a wrld key", () => {
    const wrapper = mountCell({ instanceId: "  inst_legacy_001  " });
    expect(wrapper.find(".mono").text()).toBe("inst_legacy_001");
    expect(
      wrapper.find('[data-testid="encounter-instance-link"]').exists(),
    ).toBe(false);
  });

  it("renders labeled link for a parseable wrld instance key", () => {
    const instanceId = "wrld_w:88577~region(jp)";
    const wrapper = mountCell({ instanceId });
    const link = wrapper.get('[data-testid="encounter-instance-link"]');
    expect(link.text()).toContain("Public #88577");
    expect(link.text()).toContain("[JP]");
    expect(link.attributes("href")).toContain("vrchat.com/home/launch");
    expect(link.attributes("href")).toContain("worldId=wrld_w");
    expect(link.attributes("aria-label")).toBe(
      ja.encounterHistory.openInVrchat.replace("{text}", link.text()),
    );
  });
});
