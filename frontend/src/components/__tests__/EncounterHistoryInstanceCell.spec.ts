import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import EncounterHistoryInstanceCell from "../EncounterHistoryInstanceCell.vue";

describe("EncounterHistoryInstanceCell", () => {
  it("shows dash when instanceId is undefined", () => {
    const wrapper = mount(EncounterHistoryInstanceCell, {});
    expect(wrapper.text()).toBe("—");
    expect(
      wrapper.find('[data-testid="encounter-instance-link"]').exists(),
    ).toBe(false);
  });

  it("shows dash when instanceId is empty or whitespace only", () => {
    for (const instanceId of ["", "   ", "\t"]) {
      const wrapper = mount(EncounterHistoryInstanceCell, {
        props: { instanceId },
      });
      expect(wrapper.text()).toBe("—");
    }
  });

  it("shows trimmed legacy id as mono fallback when not a wrld key", () => {
    const wrapper = mount(EncounterHistoryInstanceCell, {
      props: { instanceId: "  inst_legacy_001  " },
    });
    expect(wrapper.find(".mono").text()).toBe("inst_legacy_001");
    expect(
      wrapper.find('[data-testid="encounter-instance-link"]').exists(),
    ).toBe(false);
  });

  it("renders labeled link for a parseable wrld instance key", () => {
    const instanceId = "wrld_w:88577~region(jp)";
    const wrapper = mount(EncounterHistoryInstanceCell, {
      props: { instanceId },
    });
    const link = wrapper.get('[data-testid="encounter-instance-link"]');
    expect(link.text()).toContain("パブリック #88577");
    expect(link.text()).toContain("[JP]");
    expect(link.attributes("href")).toContain("vrchat.com/home/launch");
    expect(link.attributes("href")).toContain("worldId=wrld_w");
  });
});
