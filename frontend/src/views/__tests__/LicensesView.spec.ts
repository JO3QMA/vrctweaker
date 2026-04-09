import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { createRouter, createWebHashHistory } from "vue-router";
import { createAppI18n } from "../../i18n";
import LicensesView from "../LicensesView.vue";

const i18n = createAppI18n("ja");

const router = createRouter({
  history: createWebHashHistory(),
  routes: [{ path: "/licenses", component: LicensesView }],
});

describe("LicensesView", () => {
  it("renders page title and intro", async () => {
    await router.push("/licenses");
    await router.isReady();
    const wrapper = mount(LicensesView, {
      global: {
        plugins: [i18n, router],
      },
    });
    expect(wrapper.find(".page-title").text()).toBe("OSS ライセンス");
    expect(wrapper.find(".intro").text()).toContain(
      "オープンソースソフトウェア",
    );
  });

  it("renders npm licenses section with table", async () => {
    await router.push("/licenses");
    await router.isReady();
    const wrapper = mount(LicensesView, {
      global: {
        plugins: [i18n, router],
      },
    });
    const section = wrapper.find(".licenses-section");
    expect(section.find("h2").text()).toBe("フロントエンド（npm）");
    const rows = wrapper.findAll(".licenses-table tbody tr");
    expect(rows.length).toBeGreaterThan(0);
  });

  it("renders go licenses section with table", async () => {
    await router.push("/licenses");
    await router.isReady();
    const wrapper = mount(LicensesView, {
      global: {
        plugins: [i18n, router],
      },
    });
    const sections = wrapper.findAll(".licenses-section");
    const goSection = sections.find(
      (s) => s.find("h2").text() === "バックエンド（Go）",
    );
    expect(goSection).toBeDefined();
    const rows = goSection!.findAll("tbody tr");
    expect(rows.length).toBeGreaterThan(0);
  });

  it("filters out vrchat-tweaker-frontend from npm licenses", async () => {
    await router.push("/licenses");
    await router.isReady();
    const wrapper = mount(LicensesView, {
      global: {
        plugins: [i18n, router],
      },
    });
    const packageNames = wrapper
      .findAll(".package-name")
      .map((el) => el.text());
    expect(packageNames).not.toContain("vrchat-tweaker-frontend");
  });
});
