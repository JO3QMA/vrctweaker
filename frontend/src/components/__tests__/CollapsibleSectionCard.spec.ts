import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import CollapsibleSectionCard from "../CollapsibleSectionCard.vue";

describe("CollapsibleSectionCard", () => {
  it("toggles expanded state and section-card--collapsed class on the card root", async () => {
    const wrapper = mount(CollapsibleSectionCard, {
      props: { title: "テストセクション" },
      slots: { default: '<p class="body-marker">本文</p>' },
    });

    const card = wrapper.find(".section-card");
    const btn = wrapper.find(".section-card__toggle");

    expect(btn.attributes("aria-expanded")).toBe("true");
    expect(card.classes()).not.toContain("section-card--collapsed");

    await btn.trigger("click");
    expect(btn.attributes("aria-expanded")).toBe("false");
    expect(card.classes()).toContain("section-card--collapsed");

    await btn.trigger("click");
    expect(btn.attributes("aria-expanded")).toBe("true");
    expect(card.classes()).not.toContain("section-card--collapsed");
  });

  it("uses v-model to control expanded state", async () => {
    const wrapper = mount(CollapsibleSectionCard, {
      props: { modelValue: false, title: "M" },
    });

    const btn = wrapper.find(".section-card__toggle");
    expect(btn.attributes("aria-expanded")).toBe("false");

    await wrapper.setProps({ modelValue: true });
    await wrapper.vm.$nextTick();
    expect(btn.attributes("aria-expanded")).toBe("true");
  });
});
