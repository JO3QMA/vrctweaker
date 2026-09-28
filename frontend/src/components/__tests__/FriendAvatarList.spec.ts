import { flushPromises, mount } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createI18n } from "vue-i18n";
import ja from "../../i18n/locales/ja.json";
import FriendAvatarList from "../FriendAvatarList.vue";
import { App } from "../../wails/app";
import * as vrcUserCacheDisplay from "../../utils/vrcUserCacheDisplay";

vi.mock("../../wails/app", () => ({
  App: {
    friendAvatarUsageByVRCUserID: vi.fn(),
  },
}));

const mockFetch = vi.mocked(App.friendAvatarUsageByVRCUserID);

function mountList(props: { userId?: string } = { userId: "usr_test" }) {
  const i18n = createI18n({
    legacy: false,
    locale: "ja",
    messages: { ja },
  });
  return mount(FriendAvatarList, {
    props,
    global: { plugins: [i18n] },
  });
}

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
    const wrapper = mountList();
    await flushPromises();
    expect(mockFetch).toHaveBeenCalledWith("usr_test");
    expect(wrapper.text()).toContain("Fox");
    expect(wrapper.find(".el-table").exists()).toBe(true);
  });

  it("copies avatar id when copy control is used", async () => {
    const copySpy = vi
      .spyOn(vrcUserCacheDisplay, "copyTextToClipboard")
      .mockResolvedValue();
    mockFetch.mockResolvedValue([
      {
        avatarName: "Fox",
        avatarId: "avtr_11111111-2222-3333-4444-555555555555",
        useCount: 1,
        firstSeenAt: "2026-01-01T00:00:00.000Z",
        lastSeenAt: "2026-01-01T00:00:00.000Z",
      },
    ]);
    const wrapper = mountList();
    await flushPromises();
    const btn = wrapper.get('[data-testid="friend-avatar-copy-id-Fox"]');
    expect(btn.attributes("aria-label")).toBe(ja.friendAvatars.copyAvatarId);
    await btn.trigger("click");
    expect(copySpy).toHaveBeenCalledWith(
      "avtr_11111111-2222-3333-4444-555555555555",
    );
    copySpy.mockRestore();
  });

  it("shows empty message when no rows", async () => {
    const wrapper = mountList();
    await flushPromises();
    expect(wrapper.text()).toContain(
      "ログから記録されたアバターはまだありません",
    );
  });

  it("shows generic error on fetch failure", async () => {
    mockFetch.mockRejectedValue(new Error("friend avatar usage: db locked"));
    const wrapper = mountList();
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

    const wrapper = mountList({ userId: "usr_old" });
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
