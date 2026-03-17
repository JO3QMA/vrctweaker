import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { createRouter, createWebHashHistory } from "vue-router";
import SettingsView from "../SettingsView.vue";

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: "/settings", component: SettingsView },
    { path: "/licenses", component: { template: "<div>Licenses</div>" } },
  ],
});

describe("SettingsView", () => {
  it("renders settings title", async () => {
    await router.push("/settings");
    await router.isReady();
    const wrapper = mount(SettingsView, {
      global: {
        plugins: [router],
      },
    });
    expect(wrapper.find(".page-title").text()).toBe("設定");
  });

  it("has link to OSS licenses page", async () => {
    await router.push("/settings");
    await router.isReady();
    const wrapper = mount(SettingsView, {
      global: {
        plugins: [router],
      },
    });
    const licensesLink = wrapper.find(".btn-licenses");
    expect(licensesLink.exists()).toBe(true);
    expect(licensesLink.attributes("href")).toContain("/licenses");
    expect(licensesLink.text()).toContain("OSS ライセンス一覧");
  });
});
