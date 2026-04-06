import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { createRouter, createWebHashHistory } from "vue-router";
import { ElMessage } from "element-plus";
import FriendsView from "../FriendsView.vue";
import { App, type UserCacheDTO } from "../../wails/app";

async function mountFriendsView(query: Record<string, string> = {}) {
  const router = createRouter({
    history: createWebHashHistory(),
    routes: [
      { path: "/friends", name: "friends", component: { template: "<div/>" } },
    ],
  });
  await router.push({ path: "/friends", query });
  await router.isReady();
  const wrapper = mount(FriendsView, {
    global: { plugins: [router] },
  });
  await flushPromises();
  return wrapper;
}

const { mockFriends, mockIsLoggedIn } = vi.hoisted(() => ({
  mockFriends: vi.fn(),
  mockIsLoggedIn: vi.fn(),
}));

vi.mock("../../composables/useSessionUnlock", () => ({
  useSessionUnlock: () => ({
    beginStartupUnlock: vi.fn().mockResolvedValue(undefined),
  }),
}));

vi.mock("../../wails/app", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../../wails/app")>();
  return {
    ...actual,
    App: {
      ...actual.App,
      friends: mockFriends,
      isLoggedIn: mockIsLoggedIn,
      refreshFriends: vi.fn(),
      reconcileVRChatSocialCache: vi.fn().mockResolvedValue(undefined),
      setFavorite: vi.fn(),
    },
  };
});

function minimalUser(
  partial: Partial<UserCacheDTO> &
    Pick<UserCacheDTO, "vrcUserId" | "displayName" | "status">,
): UserCacheDTO {
  return {
    isFavorite: false,
    lastUpdated: "",
    ...partial,
  } as UserCacheDTO;
}

/**
 * ElSwitch の inner checkbox input を返す
 * （ElSwitch は data-testid を外側 div に残すため `input` サフィックスが必要）
 */
function switchInput(wrapper: ReturnType<typeof mount>, testId: string) {
  return wrapper.find(`[data-testid="${testId}"] input`);
}

/**
 * ElInput の inner input を返す
 * （ElInput は data-testid を inner input に転送するため直接セレクタで取得できる）
 */
function elInputEl(wrapper: ReturnType<typeof mount>, testId: string) {
  return wrapper.find(`[data-testid="${testId}"]`);
}

describe("FriendsView", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockIsLoggedIn.mockResolvedValue(false);
  });

  it("renders single Online / Offline mode toggle", async () => {
    mockFriends.mockResolvedValue([]);
    const wrapper = await mountFriendsView();

    expect(wrapper.find('[data-testid="friends-filter-mode"]').exists()).toBe(
      true,
    );
    expect(wrapper.text()).toMatch(/Online/);
    expect(wrapper.text()).toMatch(/Offline/);
  });

  it("defaults to Online: shows online friends, hides offline", async () => {
    mockFriends.mockResolvedValue([
      minimalUser({
        vrcUserId: "1",
        displayName: "OnUser",
        status: "join me",
      }),
      minimalUser({
        vrcUserId: "2",
        displayName: "OffUser",
        status: "offline",
      }),
    ]);
    const wrapper = await mountFriendsView();

    // ElSwitch の inner checkbox input で確認
    const mode = switchInput(wrapper, "friends-filter-mode")
      .element as HTMLInputElement;
    expect(mode.checked).toBe(false);

    expect(wrapper.text()).toContain("OnUser");
    expect(wrapper.text()).not.toContain("OffUser");
  });

  it("when toggle is on, shows offline friends only", async () => {
    mockFriends.mockResolvedValue([
      minimalUser({
        vrcUserId: "1",
        displayName: "OnUser",
        status: "active",
      }),
      minimalUser({
        vrcUserId: "2",
        displayName: "OffUser",
        status: "offline",
      }),
    ]);
    const wrapper = await mountFriendsView();

    await switchInput(wrapper, "friends-filter-mode").setValue(true);
    await wrapper.vm.$nextTick();

    expect(wrapper.text()).not.toContain("OnUser");
    expect(wrapper.text()).toContain("OffUser");
  });

  it("renders display name search input below toolbar", async () => {
    mockFriends.mockResolvedValue([]);
    const wrapper = await mountFriendsView();

    expect(
      wrapper.find('[data-testid="friends-search-display-name"]').exists(),
    ).toBe(true);
    // ElInput は data-testid を inner input に転送するため直接 .element で参照可
    const inner = elInputEl(wrapper, "friends-search-display-name")
      .element as HTMLInputElement;
    expect(inner.placeholder).toBe("表示名で検索");
  });

  it("filters friends by display name (case-insensitive)", async () => {
    mockFriends.mockResolvedValue([
      minimalUser({
        vrcUserId: "1",
        displayName: "AlphaTest",
        status: "join me",
      }),
      minimalUser({
        vrcUserId: "2",
        displayName: "BetaUser",
        status: "join me",
      }),
    ]);
    const wrapper = await mountFriendsView();

    // ElInput の data-testid は inner input に行くため直接 setValue
    await elInputEl(wrapper, "friends-search-display-name").setValue("beta");
    await wrapper.vm.$nextTick();

    expect(wrapper.text()).toContain("BetaUser");
    expect(wrapper.text()).not.toContain("AlphaTest");
  });

  it("shows message when search matches no one", async () => {
    mockFriends.mockResolvedValue([
      minimalUser({
        vrcUserId: "1",
        displayName: "OnlyOne",
        status: "join me",
      }),
    ]);
    const wrapper = await mountFriendsView();

    await elInputEl(wrapper, "friends-search-display-name").setValue("nope");
    await wrapper.vm.$nextTick();

    expect(wrapper.text()).toContain("検索に一致するフレンドはいません");
  });

  it("detail omits user id row and offers copy on display name", async () => {
    mockFriends.mockResolvedValue([
      minimalUser({
        vrcUserId: "usr_internal_id_value",
        displayName: "VisibleName",
        status: "active",
      }),
    ]);
    const wrapper = await mountFriendsView();

    await wrapper.find(".friend-card").trigger("click");
    await wrapper.vm.$nextTick();

    expect(wrapper.text()).not.toContain("ユーザーID");
    expect(wrapper.text()).not.toContain("usr_internal_id_value");
    expect(
      wrapper.find('[data-testid="friend-copy-display-name"]').exists(),
    ).toBe(true);
  });

  it("copy button writes display name to clipboard", async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.assign(navigator, { clipboard: { writeText } });

    mockFriends.mockResolvedValue([
      minimalUser({
        vrcUserId: "1",
        displayName: "CopyMe Friend",
        status: "active",
      }),
    ]);
    const wrapper = await mountFriendsView();

    await wrapper.find(".friend-card").trigger("click");
    await wrapper.vm.$nextTick();

    await wrapper
      .find('[data-testid="friend-copy-display-name"]')
      .trigger("click");
    expect(writeText).toHaveBeenCalledWith("CopyMe Friend");
  });

  it("selects friend from vrcUserId query and switches to offline list when needed", async () => {
    mockFriends.mockResolvedValue([
      minimalUser({
        vrcUserId: "2",
        displayName: "OffUser",
        status: "offline",
      }),
    ]);
    const wrapper = await mountFriendsView({ vrcUserId: "2" });

    const mode = switchInput(wrapper, "friends-filter-mode")
      .element as HTMLInputElement;
    expect(mode.checked).toBe(true);
    expect(wrapper.find(".friend-card.active").exists()).toBe(true);
    expect(wrapper.text()).toContain("OffUser");
  });

  it("throttles reconcileVRChatSocialCache on visibility within 60s", async () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-01-01T12:00:00.000Z"));
    mockFriends.mockResolvedValue([]);
    mockIsLoggedIn.mockResolvedValue(true);
    const wrapper = await mountFriendsView();
    await flushPromises();

    const reconcile = vi.mocked(App.reconcileVRChatSocialCache);
    reconcile.mockClear();

    vi.setSystemTime(new Date("2026-01-01T12:00:30.000Z"));
    document.dispatchEvent(new Event("visibilitychange"));
    await flushPromises();
    expect(reconcile).not.toHaveBeenCalled();

    vi.setSystemTime(new Date("2026-01-01T12:01:01.000Z"));
    document.dispatchEvent(new Event("visibilitychange"));
    await flushPromises();
    expect(reconcile).toHaveBeenCalledTimes(1);

    wrapper.unmount();
    vi.useRealTimers();
  });

  it("warns and removes vrcUserId query when id is not in friends list", async () => {
    const noopHandler = { close: () => {} };
    const warnSpy = vi
      .spyOn(ElMessage, "warning")
      .mockImplementation(() => noopHandler);
    mockFriends.mockResolvedValue([
      minimalUser({
        vrcUserId: "1",
        displayName: "OnlyOne",
        status: "join me",
      }),
    ]);
    const wrapper = await mountFriendsView({ vrcUserId: "missing-id" });
    await flushPromises();

    expect(warnSpy).toHaveBeenCalled();
    const router = wrapper.vm.$router;
    expect(router.currentRoute.value.query.vrcUserId).toBeUndefined();
    warnSpy.mockRestore();
  });
});
