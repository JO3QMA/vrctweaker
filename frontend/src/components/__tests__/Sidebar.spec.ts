import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { createRouter, createWebHashHistory } from "vue-router";
import Sidebar from "../Sidebar.vue";

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: "/", component: { template: "<div />" } },
    { path: "/settings", component: { template: "<div />" } },
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
    const links = wrapper.findAll(".el-menu-item");
    expect(links.length).toBeGreaterThanOrEqual(7); // 7 main + settings
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
