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
});
