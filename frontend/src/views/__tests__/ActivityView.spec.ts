import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { createRouter, createWebHashHistory } from "vue-router";
import ActivityView from "../ActivityView.vue";

vi.mock("../../components/PlayTimeChart.vue", () => ({
  default: {
    name: "PlayTimeChartStub",
    template: '<div class="playtime-chart-stub" />',
  },
}));

vi.mock("../../wails/app", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../../wails/app")>();
  return {
    ...actual,
    App: {
      ...actual.App,
      encounters: vi.fn().mockResolvedValue([]),
      getActivityStats: vi.fn().mockResolvedValue({
        dailyPlaySeconds: [],
        topWorlds: [],
      }),
    },
  };
});

describe("ActivityView", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("uses layout classes for playtime card and encounter log scroll region", async () => {
    const router = createRouter({
      history: createWebHashHistory(),
      routes: [{ path: "/activity", component: ActivityView }],
    });
    await router.push("/activity");
    await router.isReady();

    const wrapper = mount(ActivityView, {
      global: { plugins: [router] },
    });
    await flushPromises();
    await wrapper.vm.$nextTick();

    const playtimeCard = wrapper.find(".section-card--playtime");
    expect(playtimeCard.exists()).toBe(true);

    const scrollRegion = wrapper.find(".encounter-log-scroll");
    expect(scrollRegion.exists()).toBe(true);

    const encountersCard = wrapper.find(".section-card--encounters");
    expect(encountersCard.exists()).toBe(true);
  });
});
