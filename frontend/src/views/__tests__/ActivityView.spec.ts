import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { createRouter, createWebHashHistory } from "vue-router";
import ActivityView from "../ActivityView.vue";

const {
  mockEncounters,
  mockGetActivityStats,
  mockResolveUserProfileNavigation,
} = vi.hoisted(() => ({
  mockEncounters: vi.fn().mockResolvedValue([]),
  mockGetActivityStats: vi.fn().mockResolvedValue({
    dailyPlaySeconds: [],
    topWorlds: [],
  }),
  mockResolveUserProfileNavigation: vi.fn().mockResolvedValue({
    openInFriendsView: false,
    user: {
      vrcUserId: "stub",
      displayName: "Stub",
      status: "",
      isFavorite: false,
      lastUpdated: "",
    },
  }),
}));

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
      encounters: mockEncounters,
      getActivityStats: mockGetActivityStats,
      resolveUserProfileNavigation: mockResolveUserProfileNavigation,
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

  it("collapses playtime section when header toggle is clicked", async () => {
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
    const toggle = playtimeCard.find(".section-card__toggle");
    expect(toggle.attributes("aria-expanded")).toBe("true");

    await toggle.trigger("click");
    expect(playtimeCard.classes()).toContain("section-card--collapsed");
    expect(toggle.attributes("aria-expanded")).toBe("false");
  });

  it("display name click pushes friends route when resolve says friend", async () => {
    mockEncounters.mockResolvedValue([
      {
        id: "1",
        vrcUserId: "u1",
        displayName: "EncUser",
        instanceId: "inst",
        joinedAt: "2024-01-01T12:00:00.000Z",
      },
    ]);
    mockResolveUserProfileNavigation.mockResolvedValue({
      openInFriendsView: true,
      user: {
        vrcUserId: "u1",
        displayName: "EncUser",
        status: "active",
        isFavorite: false,
        lastUpdated: "",
      },
    });

    const router = createRouter({
      history: createWebHashHistory(),
      routes: [
        { path: "/activity", component: ActivityView },
        {
          path: "/friends",
          name: "friends",
          component: { template: "<div/>" },
        },
        {
          path: "/user-profile",
          name: "user-profile",
          component: { template: "<div/>" },
        },
      ],
    });
    await router.push("/activity");
    await router.isReady();
    const pushSpy = vi.spyOn(router, "push");

    const wrapper = mount(ActivityView, {
      global: { plugins: [router] },
    });
    await flushPromises();
    await wrapper.vm.$nextTick();

    await wrapper.find(".timeline-link").trigger("click");
    await flushPromises();

    expect(mockResolveUserProfileNavigation).toHaveBeenCalledWith("u1");
    expect(pushSpy).toHaveBeenCalledWith({
      name: "friends",
      query: { vrcUserId: "u1" },
    });
  });

  it("display name click pushes user-profile when resolve says non-friend", async () => {
    mockEncounters.mockResolvedValue([
      {
        id: "1",
        vrcUserId: "u2",
        displayName: "Stranger",
        instanceId: "inst",
        joinedAt: "2024-01-01T12:00:00.000Z",
      },
    ]);
    mockResolveUserProfileNavigation.mockResolvedValue({
      openInFriendsView: false,
      user: {
        vrcUserId: "u2",
        displayName: "Stranger",
        status: "",
        isFavorite: false,
        lastUpdated: "",
      },
    });

    const router = createRouter({
      history: createWebHashHistory(),
      routes: [
        { path: "/activity", component: ActivityView },
        {
          path: "/friends",
          name: "friends",
          component: { template: "<div/>" },
        },
        {
          path: "/user-profile",
          name: "user-profile",
          component: { template: "<div/>" },
        },
      ],
    });
    await router.push("/activity");
    await router.isReady();
    const pushSpy = vi.spyOn(router, "push");

    const wrapper = mount(ActivityView, {
      global: { plugins: [router] },
    });
    await flushPromises();
    await wrapper.vm.$nextTick();

    await wrapper.find(".timeline-link").trigger("click");
    await flushPromises();

    expect(pushSpy).toHaveBeenCalledWith({
      name: "user-profile",
      query: { vrcUserId: "u2", displayName: "Stranger" },
    });
  });

  it("display name click falls back to user-profile when resolve rejects", async () => {
    mockEncounters.mockResolvedValue([
      {
        id: "1",
        vrcUserId: "u3",
        displayName: "FailUser",
        instanceId: "inst",
        joinedAt: "2024-01-01T12:00:00.000Z",
      },
    ]);
    mockResolveUserProfileNavigation.mockRejectedValue(new Error("boom"));

    const router = createRouter({
      history: createWebHashHistory(),
      routes: [
        { path: "/activity", component: ActivityView },
        {
          path: "/friends",
          name: "friends",
          component: { template: "<div/>" },
        },
        {
          path: "/user-profile",
          name: "user-profile",
          component: { template: "<div/>" },
        },
      ],
    });
    await router.push("/activity");
    await router.isReady();
    const pushSpy = vi.spyOn(router, "push");

    const wrapper = mount(ActivityView, {
      global: { plugins: [router] },
    });
    await flushPromises();
    await wrapper.vm.$nextTick();

    await wrapper.find(".timeline-link").trigger("click");
    await flushPromises();

    expect(pushSpy).toHaveBeenCalledWith({
      name: "user-profile",
      query: { vrcUserId: "u3", displayName: "FailUser" },
    });
  });
});
