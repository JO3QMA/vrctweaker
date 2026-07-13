import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { createRouter, createWebHashHistory } from "vue-router";
import Sidebar from "../Sidebar.vue";
import { App } from "../../wails/app";

vi.mock("../../wails/app", () => ({
  App: {
    runtimeIsWindows: vi.fn(),
  },
}));

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: "/", component: { template: "<div />" } },
    { path: "/video", component: { template: "<div />" } },
    { path: "/settings", component: { template: "<div />" } },
  ],
});

describe("Sidebar", () => {
  beforeEach(() => {
    vi.mocked(App.runtimeIsWindows).mockResolvedValue(false);
  });

  it("renders menu items", async () => {
    await router.push("/");
    await router.isReady();
    const wrapper = mount(Sidebar, {
      global: {
        plugins: [router],
      },
    });
    await flushPromises();
    const links = wrapper.findAll(".el-menu-item");
    expect(links.length).toBeGreaterThanOrEqual(7); // 7 main + settings
  });

  it("has dashboard link", async () => {
    const wrapper = mount(Sidebar, {
      global: {
        plugins: [router],
      },
    });
    await flushPromises();
    expect(wrapper.text()).toContain("ダッシュボード");
  });

  it("shows video nav on Windows", async () => {
    vi.mocked(App.runtimeIsWindows).mockResolvedValue(true);
    const wrapper = mount(Sidebar, {
      global: {
        plugins: [router],
      },
    });
    await flushPromises();
    expect(wrapper.text()).toContain("動画");
  });
});
