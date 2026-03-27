import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import FriendsView from "../FriendsView.vue";
import type { UserCacheDTO } from "../../wails/app";

const { mockFriends, mockIsLoggedIn } = vi.hoisted(() => ({
  mockFriends: vi.fn(),
  mockIsLoggedIn: vi.fn(),
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

describe("FriendsView", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockIsLoggedIn.mockResolvedValue(false);
  });

  it("renders single Online / Offline mode toggle", async () => {
    mockFriends.mockResolvedValue([]);
    const wrapper = mount(FriendsView);
    await flushPromises();

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
    const wrapper = mount(FriendsView);
    await flushPromises();

    const mode = wrapper.find('[data-testid="friends-filter-mode"]')
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
    const wrapper = mount(FriendsView);
    await flushPromises();

    await wrapper.find('[data-testid="friends-filter-mode"]').setValue(true);
    await wrapper.vm.$nextTick();

    expect(wrapper.text()).not.toContain("OnUser");
    expect(wrapper.text()).toContain("OffUser");
  });

  it("renders display name search input below toolbar", async () => {
    mockFriends.mockResolvedValue([]);
    const wrapper = mount(FriendsView);
    await flushPromises();

    const input = wrapper.find('[data-testid="friends-search-display-name"]');
    expect(input.exists()).toBe(true);
    expect((input.element as HTMLInputElement).placeholder).toBe(
      "表示名で検索",
    );
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
    const wrapper = mount(FriendsView);
    await flushPromises();

    await wrapper
      .find('[data-testid="friends-search-display-name"]')
      .setValue("beta");
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
    const wrapper = mount(FriendsView);
    await flushPromises();

    await wrapper
      .find('[data-testid="friends-search-display-name"]')
      .setValue("nope");
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
    const wrapper = mount(FriendsView);
    await flushPromises();

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
    const wrapper = mount(FriendsView);
    await flushPromises();

    await wrapper.find(".friend-card").trigger("click");
    await wrapper.vm.$nextTick();

    await wrapper
      .find('[data-testid="friend-copy-display-name"]')
      .trigger("click");
    expect(writeText).toHaveBeenCalledWith("CopyMe Friend");
  });
});
