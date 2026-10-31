import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { createRouter, createWebHashHistory } from "vue-router";
import EncounterHistoryDetailView from "../EncounterHistoryDetailView.vue";
import EncounterHistoryList from "../../components/EncounterHistoryList.vue";
import * as wailsApp from "../../wails/app";
import type { UserEncounterDTO } from "../../wails/app";

vi.mock("../../wails/app", () => ({
  App: {
    encountersByVRCUserID: vi.fn(),
    encountersByWorldID: vi.fn(),
  },
}));

const sampleEncounter: UserEncounterDTO = {
  id: "e1",
  vrcUserId: "usr_a",
  displayName: "UserA",
  instanceId: "inst_1",
  worldId: "wrld_w",
  worldDisplayName: "World W",
  joinedAt: "2025-01-01T12:00:00Z",
  leftAt: "2025-01-01T13:00:00Z",
  isFirstEncounter: false,
  isListableFriend: false,
};

describe("EncounterHistoryDetailView", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls encountersByVRCUserID and renders rows for kind=user", async () => {
    vi.mocked(wailsApp.App.encountersByVRCUserID).mockResolvedValue([
      sampleEncounter,
    ]);

    const router = createRouter({
      history: createWebHashHistory(),
      routes: [
        {
          path: "/activity/encounter-history",
          component: EncounterHistoryDetailView,
        },
      ],
    });
    await router.push({
      path: "/activity/encounter-history",
      query: { kind: "user", vrcUserId: "usr_a" },
    });
    await router.isReady();

    mount(EncounterHistoryDetailView, {
      global: { plugins: [router] },
    });

    await flushPromises();

    expect(wailsApp.App.encountersByVRCUserID).toHaveBeenCalledWith("usr_a");
    expect(wailsApp.App.encountersByWorldID).not.toHaveBeenCalled();
  });

  it("passes trimmed ids to EncounterHistoryList", async () => {
    vi.mocked(wailsApp.App.encountersByVRCUserID).mockResolvedValue([]);

    const router = createRouter({
      history: createWebHashHistory(),
      routes: [
        {
          path: "/activity/encounter-history",
          component: EncounterHistoryDetailView,
        },
      ],
    });
    await router.push({
      path: "/activity/encounter-history",
      query: {
        kind: "user",
        vrcUserId: "  usr_a  ",
        worldId: "  wrld_w  ",
      },
    });
    await router.isReady();

    const wrapper = mount(EncounterHistoryDetailView, {
      global: { plugins: [router] },
    });
    await flushPromises();

    const list = wrapper.findComponent(EncounterHistoryList);
    expect(list.props("userId")).toBe("usr_a");
    expect(list.props("worldId")).toBe("wrld_w");
    expect(wailsApp.App.encountersByVRCUserID).toHaveBeenCalledWith("usr_a");
  });

  it("calls encountersByWorldID for kind=world", async () => {
    vi.mocked(wailsApp.App.encountersByWorldID).mockResolvedValue([
      sampleEncounter,
    ]);

    const router = createRouter({
      history: createWebHashHistory(),
      routes: [
        {
          path: "/activity/encounter-history",
          component: EncounterHistoryDetailView,
        },
      ],
    });
    await router.push({
      path: "/activity/encounter-history",
      query: { kind: "world", worldId: "wrld_w" },
    });
    await router.isReady();

    mount(EncounterHistoryDetailView, {
      global: { plugins: [router] },
    });

    await flushPromises();

    expect(wailsApp.App.encountersByWorldID).toHaveBeenCalledWith("wrld_w");
    expect(wailsApp.App.encountersByVRCUserID).not.toHaveBeenCalled();
  });

  it("does not call backend when query is invalid", async () => {
    const router = createRouter({
      history: createWebHashHistory(),
      routes: [
        {
          path: "/activity/encounter-history",
          component: EncounterHistoryDetailView,
        },
      ],
    });
    await router.push({
      path: "/activity/encounter-history",
      query: { kind: "user" },
    });
    await router.isReady();

    mount(EncounterHistoryDetailView, {
      global: { plugins: [router] },
    });

    await flushPromises();

    expect(wailsApp.App.encountersByVRCUserID).not.toHaveBeenCalled();
    expect(wailsApp.App.encountersByWorldID).not.toHaveBeenCalled();
  });
});
