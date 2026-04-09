import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import EncounterHistoryList from "../EncounterHistoryList.vue";
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
};

describe("EncounterHistoryList", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls encountersByVRCUserID in user mode", async () => {
    vi.mocked(wailsApp.App.encountersByVRCUserID).mockResolvedValue([
      sampleEncounter,
    ]);

    mount(EncounterHistoryList, {
      props: { mode: "user", userId: "usr_x" },
    });
    await flushPromises();

    expect(wailsApp.App.encountersByVRCUserID).toHaveBeenCalledWith("usr_x");
    expect(wailsApp.App.encountersByWorldID).not.toHaveBeenCalled();
  });

  it("calls encountersByWorldID in world mode", async () => {
    vi.mocked(wailsApp.App.encountersByWorldID).mockResolvedValue([
      sampleEncounter,
    ]);

    mount(EncounterHistoryList, {
      props: { mode: "world", worldId: "wrld_abc" },
    });
    await flushPromises();

    expect(wailsApp.App.encountersByWorldID).toHaveBeenCalledWith("wrld_abc");
    expect(wailsApp.App.encountersByVRCUserID).not.toHaveBeenCalled();
  });

  it("does not call backend when userId is empty in user mode", async () => {
    mount(EncounterHistoryList, {
      props: { mode: "user", userId: "" },
    });
    await flushPromises();

    expect(wailsApp.App.encountersByVRCUserID).not.toHaveBeenCalled();
    expect(wailsApp.App.encountersByWorldID).not.toHaveBeenCalled();
  });

  it("hides display name column when hideDisplayNameColumn is true", async () => {
    vi.mocked(wailsApp.App.encountersByVRCUserID).mockResolvedValue([
      sampleEncounter,
    ]);

    const wrapper = mount(EncounterHistoryList, {
      props: {
        mode: "user",
        userId: "u1",
        hideDisplayNameColumn: true,
      },
    });
    await flushPromises();

    const headerTexts = wrapper.findAll("th").map((th) => th.text());
    expect(headerTexts.some((t) => t.includes("表示名"))).toBe(false);
    expect(headerTexts.some((t) => t.includes("ワールド名"))).toBe(true);
  });

  it("shows error alert when fetch fails", async () => {
    vi.mocked(wailsApp.App.encountersByVRCUserID).mockRejectedValue(
      new Error("network down"),
    );

    const wrapper = mount(EncounterHistoryList, {
      props: { mode: "user", userId: "u1" },
    });
    await flushPromises();

    expect(wrapper.find(".el-alert--error").exists()).toBe(true);
    expect(wrapper.text()).toContain("network down");
  });

  it("shows translated fallback when fetch rejects with non-Error", async () => {
    vi.mocked(wailsApp.App.encountersByVRCUserID).mockRejectedValue("boom");

    const wrapper = mount(EncounterHistoryList, {
      props: { mode: "user", userId: "u1" },
    });
    await flushPromises();

    expect(wrapper.find(".el-alert--error").exists()).toBe(true);
    expect(wrapper.text()).toContain("データの取得に失敗しました。");
    expect(wrapper.text()).not.toContain("encounterHistory.fetchFailedGeneric");
  });
});
