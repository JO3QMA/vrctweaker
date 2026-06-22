import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { createRouter, createWebHashHistory } from "vue-router";
import SelfProfileView from "../SelfProfileView.vue";
import { App } from "../../wails/app";
import { resetSessionUnlockForStorybook } from "../../composables/useSessionUnlock";

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: "/me", component: SelfProfileView },
    { path: "/settings", name: "settings", component: { template: "<div/>" } },
  ],
});

const sampleSelf = {
  vrcUserId: "usr_me",
  displayName: "Self User",
  status: "active",
  isFavorite: false,
  lastUpdated: "2025-01-01T00:00:00Z",
};

describe("SelfProfileView", () => {
  beforeEach(async () => {
    resetSessionUnlockForStorybook();
    await router.push("/me");
    await router.isReady();
    vi.spyOn(App, "isLoggedIn").mockResolvedValue(false);
    vi.spyOn(App, "getSelfProfile").mockResolvedValue(sampleSelf);
  });

  it("shows login hint when not logged in", async () => {
    const wrapper = mount(SelfProfileView, {
      global: { plugins: [router] },
    });
    await flushPromises();

    expect(wrapper.text()).toContain("ログインが必要");
    expect(wrapper.find(".settings-link").attributes("href")).toBe(
      "#/settings",
    );
  });

  it("loads self profile when logged in", async () => {
    vi.mocked(App.isLoggedIn).mockResolvedValue(true);
    const wrapper = mount(SelfProfileView, {
      global: { plugins: [router] },
    });
    await flushPromises();

    expect(App.getSelfProfile).toHaveBeenCalledWith(false);
    expect(wrapper.text()).toContain("Self User");
  });
});
