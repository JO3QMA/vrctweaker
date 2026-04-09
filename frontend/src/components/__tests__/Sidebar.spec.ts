import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { createRouter, createWebHashHistory } from "vue-router";
import Sidebar from "../Sidebar.vue";

const stub = { template: "<div />" };
const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: "/", component: stub },
    { path: "/launcher", component: stub },
    { path: "/gallery", component: stub },
    { path: "/activity", component: stub },
    { path: "/friends", component: stub },
    { path: "/automation", component: stub },
    { path: "/config", component: stub },
    { path: "/settings", component: stub },
  ],
});

describe("Sidebar", () => {
  it("renders menu items", async () => {
    await router.push("/");
    await router.isReady();
    const wrapper = mount(Sidebar, {
      global: {
        plugins: [router],
      },
    });
    const links = wrapper.findAll(".sidebar-item");
    expect(links.length).toBe(8); // 7 main + settings
  });

  it("has dashboard link", () => {
    const wrapper = mount(Sidebar, {
      global: {
        plugins: [router],
      },
    });
    expect(wrapper.text()).toContain("ダッシュボード");
  });
});
