import { flushPromises, mount } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import FriendAvatarList from "../FriendAvatarList.vue";
import { App } from "../../wails/app";

vi.mock("../../wails/app", () => ({
  App: {
    friendAvatarUsageByVRCUserID: vi.fn(),
  },
}));

const mockFetch = vi.mocked(App.friendAvatarUsageByVRCUserID);

describe("FriendAvatarList", () => {
  beforeEach(() => {
    mockFetch.mockReset();
    mockFetch.mockResolvedValue([]);
  });

  it("loads avatar usage for user id", async () => {
    mockFetch.mockResolvedValue([
      {
        avatarName: "Fox",
        useCount: 2,
        firstSeenAt: "2026-01-01T00:00:00.000Z",
        lastSeenAt: "2026-01-02T00:00:00.000Z",
      },
    ]);
    const wrapper = mount(FriendAvatarList, {
      props: { userId: "usr_test" },
    });
    await flushPromises();
    expect(mockFetch).toHaveBeenCalledWith("usr_test");
    expect(wrapper.text()).toContain("Fox");
    expect(wrapper.find(".el-table").exists()).toBe(true);
  });

  it("shows empty message when no rows", async () => {
    const wrapper = mount(FriendAvatarList, {
      props: { userId: "usr_test" },
    });
    await flushPromises();
    expect(wrapper.text()).toContain(
      "ログから記録されたアバターはまだありません",
    );
  });

  it("shows generic error on fetch failure", async () => {
    mockFetch.mockRejectedValue(new Error("friend avatar usage: db locked"));
    const wrapper = mount(FriendAvatarList, {
      props: { userId: "usr_test" },
    });
    await flushPromises();
    expect(wrapper.text()).toContain("アバター一覧の取得に失敗しました");
  });

  it("ignores stale fetch when userId changes quickly", async () => {
    let resolveFirst!: (
      rows: {
        avatarName: string;
        useCount: number;
        firstSeenAt: string;
        lastSeenAt: string;
      }[],
    ) => void;
    const firstPromise = new Promise<
      {
        avatarName: string;
        useCount: number;
        firstSeenAt: string;
        lastSeenAt: string;
      }[]
    >((resolve) => {
      resolveFirst = resolve;
    });
    mockFetch.mockImplementationOnce(() => firstPromise);
    mockFetch.mockResolvedValueOnce([
      {
        avatarName: "Current",
        useCount: 1,
        firstSeenAt: "2026-01-01T00:00:00.000Z",
        lastSeenAt: "2026-01-01T00:00:00.000Z",
      },
    ]);

    const wrapper = mount(FriendAvatarList, {
      props: { userId: "usr_old" },
    });
    await wrapper.setProps({ userId: "usr_new" });
    await flushPromises();
    expect(wrapper.text()).toContain("Current");

    resolveFirst([
      {
        avatarName: "Stale",
        useCount: 9,
        firstSeenAt: "2026-01-01T00:00:00.000Z",
        lastSeenAt: "2026-01-01T00:00:00.000Z",
      },
    ]);
    await flushPromises();
    expect(wrapper.text()).toContain("Current");
    expect(wrapper.text()).not.toContain("Stale");
  });
});
