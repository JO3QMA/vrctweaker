import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { createRouter, createWebHashHistory } from "vue-router";
import ActivityView from "../ActivityView.vue";

const {
  mockEncounters,
  mockGetActivityStats,
  mockGetLogRetentionDays,
  mockResolveUserProfileNavigation,
} = vi.hoisted(() => ({
  mockEncounters: vi.fn().mockResolvedValue([]),
  mockGetActivityStats: vi.fn().mockResolvedValue({
    dailyPlaySeconds: [],
    topWorlds: [],
  }),
  mockGetLogRetentionDays: vi.fn().mockResolvedValue(30),
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

const runtimeHooks = vi.hoisted(() => ({
  encountersChangedHandler: null as (() => void) | null,
}));

vi.mock("../../utils/openEncounterHistoryWindow", () => ({
  openEncounterHistoryWindow: vi.fn(),
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
      getLogRetentionDays: mockGetLogRetentionDays,
      resolveUserProfileNavigation: mockResolveUserProfileNavigation,
    },
  };
});

vi.mock("../../wails/runtime", () => ({
  getRuntime: () => ({
    EventsOn: (event: string, handler: () => void) => {
      if (event === "activity:encounters-changed") {
        runtimeHooks.encountersChangedHandler = handler;
      }
      return () => {};
    },
  }),
}));

import { openEncounterHistoryWindow } from "../../utils/openEncounterHistoryWindow";

describe("ActivityView", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    runtimeHooks.encountersChangedHandler = null;
    mockEncounters.mockResolvedValue([]);
    mockGetActivityStats.mockResolvedValue({
      dailyPlaySeconds: [],
      topWorlds: [],
    });
    mockGetLogRetentionDays.mockResolvedValue(30);
  });

  async function mountActivity() {
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
    const wrapper = mount(ActivityView, {
      global: { plugins: [router] },
    });
    await flushPromises();
    await wrapper.vm.$nextTick();
    return { wrapper, router };
  }

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

    const cards = wrapper.findAll(".collapsible-section-card");
    expect(cards[0]!.classes()).toContain("section-card--encounters");
    expect(cards[1]!.classes()).toContain("section-card--playtime");
  });

  it("starts with playtime section collapsed and encounters expanded", async () => {
    const { wrapper } = await mountActivity();
    const playtimeCard = wrapper.find(".section-card--playtime");
    const encountersCard = wrapper.find(".section-card--encounters");

    expect(
      playtimeCard.find(".section-card__toggle").attributes("aria-expanded"),
    ).toBe("false");
    expect(
      encountersCard.find(".section-card__toggle").attributes("aria-expanded"),
    ).toBe("true");
  });

  it("expands playtime section when header toggle is clicked", async () => {
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
    expect(toggle.attributes("aria-expanded")).toBe("false");

    await toggle.trigger("click");
    expect(playtimeCard.classes()).not.toContain("section-card--collapsed");
    expect(toggle.attributes("aria-expanded")).toBe("true");
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

  it("loads encounters, stats, and retention days on mount", async () => {
    await mountActivity();
    expect(mockEncounters).toHaveBeenCalled();
    expect(mockGetActivityStats).toHaveBeenCalled();
    expect(mockGetLogRetentionDays).toHaveBeenCalled();
  });

  it("shows retention hint with configured days", async () => {
    mockGetLogRetentionDays.mockResolvedValue(45);
    const { wrapper } = await mountActivity();
    expect(wrapper.find(".retention-hint").text()).toContain("45");
  });

  it("shows playtime chart when stats are available and section is expanded", async () => {
    mockGetActivityStats.mockResolvedValue({
      dailyPlaySeconds: [{ date: "2026-06-01", seconds: 3600 }],
      topWorlds: [],
    });
    const { wrapper } = await mountActivity();
    await wrapper
      .find(".section-card--playtime .section-card__toggle")
      .trigger("click");
    await wrapper.vm.$nextTick();
    expect(wrapper.find(".playtime-chart-stub").exists()).toBe(true);
  });

  it("filters encounters by display name", async () => {
    mockEncounters.mockResolvedValue([
      {
        id: "1",
        vrcUserId: "u1",
        displayName: "Alpha",
        instanceId: "inst",
        joinedAt: "2024-01-01T12:00:00.000Z",
      },
      {
        id: "2",
        vrcUserId: "u2",
        displayName: "Beta",
        instanceId: "inst",
        joinedAt: "2024-01-02T12:00:00.000Z",
      },
    ]);
    const { wrapper } = await mountActivity();

    const input = wrapper.find(".filters .el-input input");
    await input.setValue("beta");
    await wrapper.vm.$nextTick();

    expect(wrapper.text()).toContain("Beta");
    expect(wrapper.text()).not.toContain("Alpha");
  });

  it("reloads encounters when refresh button is clicked", async () => {
    const { wrapper } = await mountActivity();
    mockEncounters.mockClear();

    const buttons = wrapper.findAll(".filters .el-button");
    await buttons[0]!.trigger("click");
    await flushPromises();

    expect(mockEncounters).toHaveBeenCalledTimes(1);
  });

  it("renders muted display name when encounter has no vrcUserId", async () => {
    mockEncounters.mockResolvedValue([
      {
        id: "1",
        vrcUserId: "",
        displayName: "Anonymous",
        instanceId: "inst",
        joinedAt: "2024-01-01T12:00:00.000Z",
      },
    ]);
    const { wrapper } = await mountActivity();

    expect(wrapper.find(".timeline-name-muted").text()).toBe("Anonymous");
    expect(wrapper.find(".timeline-link").exists()).toBe(false);
  });

  it("shows still-present label for open encounter", async () => {
    mockEncounters.mockResolvedValue([
      {
        id: "1",
        vrcUserId: "u1",
        displayName: "StillHere",
        instanceId: "inst",
        joinedAt: "2024-01-01T12:00:00.000Z",
        leftAt: "",
      },
    ]);
    const { wrapper } = await mountActivity();
    expect(wrapper.text()).toContain("滞在中");
  });

  it("opens world encounter history when world link is clicked", async () => {
    mockEncounters.mockResolvedValue([
      {
        id: "1",
        vrcUserId: "u1",
        displayName: "User",
        worldId: "wrld_abc",
        worldDisplayName: "Test World",
        instanceId: "inst",
        joinedAt: "2024-01-01T12:00:00.000Z",
      },
    ]);
    const { wrapper, router } = await mountActivity();

    const worldLinks = wrapper.findAll(".timeline-link");
    await worldLinks[worldLinks.length - 1]!.trigger("click");

    expect(openEncounterHistoryWindow).toHaveBeenCalledWith(
      router,
      "world",
      "wrld_abc",
    );
  });

  it("debounces encounter reload on activity:encounters-changed event", async () => {
    vi.useFakeTimers();
    await mountActivity();
    mockEncounters.mockClear();

    runtimeHooks.encountersChangedHandler?.();
    runtimeHooks.encountersChangedHandler?.();
    expect(mockEncounters).not.toHaveBeenCalled();

    await vi.advanceTimersByTimeAsync(400);
    expect(mockEncounters).toHaveBeenCalledTimes(1);
    vi.useRealTimers();
  });

  it("shows empty encounters message when list is empty", async () => {
    mockEncounters.mockResolvedValue([]);
    const { wrapper } = await mountActivity();
    expect(wrapper.find(".empty").text()).toContain("遭遇");
  });

  it("collapses encounter section when header toggle is clicked", async () => {
    const { wrapper } = await mountActivity();
    const encountersCard = wrapper.find(".section-card--encounters");
    const toggle = encountersCard.find(".section-card__toggle");

    await toggle.trigger("click");
    expect(encountersCard.classes()).toContain("section-card--collapsed");
  });
});
