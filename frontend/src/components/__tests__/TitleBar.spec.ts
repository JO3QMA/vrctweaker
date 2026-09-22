import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import { createI18n } from "vue-i18n";
import en from "../../i18n/locales/en.json";
import TitleBar from "../TitleBar.vue";

describe("TitleBar", () => {
  it("shows the app icon beside the title", () => {
    const i18n = createI18n({
      legacy: false,
      locale: "en",
      messages: { en },
    });
    const wrapper = mount(TitleBar, {
      global: { plugins: [i18n] },
    });
    const icon = wrapper.get('[data-testid="title-bar-app-icon"]');
    expect(icon.attributes("alt")).toBe("VRChat Tweaker");
    expect(icon.attributes("src")).toMatch(/appicon\.png$/);
  });
});
