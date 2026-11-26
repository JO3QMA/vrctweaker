import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { createRouter, createWebHashHistory } from "vue-router";
import LicensesView from "../LicensesView.vue";

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
        plugins: [router],
      },
    });
    expect(wrapper.find(".page-title").text()).toBe("OSSライセンス");
    expect(wrapper.find(".intro-with-siblings").text()).toContain(
      "オープンソースソフトウェア",
    );
  });

  it("shows unofficial disclaimer and legal doc links", async () => {
    await router.push("/licenses");
    await router.isReady();
    const wrapper = mount(LicensesView, {
      global: {
        plugins: [router],
      },
    });
    expect(
      wrapper.find('[data-testid="licenses-unofficial-disclaimer"]').text(),
    ).toContain("非公式");
    expect(
      wrapper.find('[data-testid="legal-link-privacy"]').attributes("href"),
    ).toContain("/legal/privacy");
  });

  it("renders npm licenses section with table", async () => {
    await router.push("/licenses");
    await router.isReady();
    const wrapper = mount(LicensesView, {
      global: {
        plugins: [router],
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
        plugins: [router],
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

  it("disables overflow tooltips on package and repository columns", async () => {
    await router.push("/licenses");
    await router.isReady();
    const wrapper = mount(LicensesView, {
      global: {
        plugins: [router],
      },
    });
    const columns = wrapper.findAllComponents({ name: "ElTableColumn" });
    const packageCol = columns.find((c) =>
      (c.props("className" as never) as string | undefined)?.includes(
        "licenses-table__package-cell",
      ),
    );
    const repoCol = columns.find((c) =>
      (c.props("className" as never) as string | undefined)?.includes(
        "licenses-table__repo-cell",
      ),
    );
    expect(packageCol?.props("showOverflowTooltip")).toBe(false);
    expect(repoCol?.props("showOverflowTooltip")).toBe(false);
  });

  it("wraps license tables for horizontal scroll on narrow viewports", async () => {
    await router.push("/licenses");
    await router.isReady();
    const wrapper = mount(LicensesView, {
      global: {
        plugins: [router],
      },
    });
    expect(
      wrapper.find('[data-testid="licenses-table-scroll-frontend"]').exists(),
    ).toBe(true);
    expect(
      wrapper.find('[data-testid="licenses-table-scroll-backend"]').exists(),
    ).toBe(true);
  });

  it("filters out vrchat-tweaker-frontend from npm licenses", async () => {
    await router.push("/licenses");
    await router.isReady();
    const wrapper = mount(LicensesView, {
      global: {
        plugins: [router],
      },
    });
    const packageNames = wrapper
      .findAll(".package-name")
      .map((el) => el.text());
    expect(packageNames).not.toContain("vrchat-tweaker-frontend");
  });
});
