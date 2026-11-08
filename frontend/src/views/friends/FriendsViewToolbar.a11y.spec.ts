import { describe, it, expect, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import ElementPlus from "element-plus";
import { createI18n } from "vue-i18n";
import FriendsViewToolbar from "./FriendsViewToolbar.vue";
import en from "../../i18n/locales/en.json";
import ja from "../../i18n/locales/ja.json";

function mountToolbar(locale: "en" | "ja" = "ja") {
  const i18n = createI18n({
    legacy: false,
    locale,
    messages: { en, ja },
  });
  return mount(FriendsViewToolbar, {
    props: {
      isLoggedIn: true,
      refreshLoading: false,
      showOfflineList: false,
      displayNameQuery: "",
    },
    global: {
      plugins: [ElementPlus, i18n],
    },
    attachTo: document.body,
  });
}

describe("FriendsViewToolbar accessibility", () => {
  beforeEach(() => {
    document.body.innerHTML = "";
  });

  it("keeps Online / switch / Offline visible layout", async () => {
    const wrapper = mountToolbar();
    await flushPromises();
    expect(wrapper.text()).toMatch(/Online/);
    expect(wrapper.text()).toMatch(/Offline/);
    expect(wrapper.find('[data-testid="friends-filter-mode"]').exists()).toBe(
      true,
    );
    wrapper.unmount();
  });

  it("uses group label and switch operation label separately (ja)", async () => {
    const wrapper = mountToolbar("ja");
    await flushPromises();
    const group = wrapper.get('[role="group"]');
    const groupLabel = group.attributes("aria-label");
    const switchInput = wrapper.get(
      '[data-testid="friends-filter-mode"] input[role="switch"]',
    );
    const switchLabel = switchInput.attributes("aria-label");
    expect(groupLabel).toBe(ja.friends.toolbarOnlineOffline);
    expect(switchLabel).toBe(ja.friends.toolbarOfflineSwitch);
    expect(groupLabel).not.toBe(switchLabel);
    wrapper.unmount();
  });
});
