import { describe, expect, it } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { nextTick } from "vue";
import VrcUserTagChip from "../VrcUserTagChip.vue";

describe("VrcUserTagChip", () => {
  it("shows localized label for known user tag", async () => {
    const wrapper = mount(VrcUserTagChip, {
      props: { tag: "system_trust_basic" },
    });
    await flushPromises();
    await nextTick();

    const chip = wrapper.find("[data-testid='user-tag-chip']");
    expect(chip.text()).toContain("New User");
    expect(chip.text()).not.toContain("新規ユーザー");
    expect(chip.attributes("data-tag-id")).toBe("system_trust_basic");
  });

  it("shows endonym for language tag", async () => {
    const wrapper = mount(VrcUserTagChip, {
      props: { tag: "language_jpn" },
    });
    await flushPromises();
    await nextTick();

    expect(wrapper.text()).toBe("日本語");
  });

  it("falls back to raw tag id for unknown tags", async () => {
    const wrapper = mount(VrcUserTagChip, {
      props: { tag: "system_slug" },
    });
    await flushPromises();
    await nextTick();

    expect(wrapper.text()).toBe("system_slug");
  });
});
