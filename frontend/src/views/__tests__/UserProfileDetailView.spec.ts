import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { createRouter, createWebHashHistory } from "vue-router";
import UserProfileDetailView from "../UserProfileDetailView.vue";
import VrcUserCacheDetail from "../../components/VrcUserCacheDetail.vue";

const { mockResolve } = vi.hoisted(() => ({
  mockResolve: vi.fn(),
}));

vi.mock("../../wails/app", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../../wails/app")>();
  return {
    ...actual,
    App: {
      ...actual.App,
      resolveUserProfileNavigation: mockResolve,
      setFavorite: vi.fn(),
    },
  };
});

describe("UserProfileDetailView", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockResolve.mockResolvedValue({
      openInFriendsView: false,
      user: {
        vrcUserId: "u1",
        displayName: "Resolved",
        status: "active",
        isFavorite: false,
        lastUpdated: "",
      },
    });
  });

  async function mountWithQuery(query: Record<string, string>) {
    const router = createRouter({
      history: createWebHashHistory(),
      routes: [
        {
          path: "/user-profile",
          name: "user-profile",
          component: UserProfileDetailView,
        },
      ],
    });
    await router.push({ path: "/user-profile", query });
    await router.isReady();
    const wrapper = mount(UserProfileDetailView, {
      global: { plugins: [router] },
    });
    await flushPromises();
    return wrapper;
  }

  it("mounts VrcUserCacheDetail after resolve returns user", async () => {
    const wrapper = await mountWithQuery({
      vrcUserId: "u1",
      displayName: "Hint",
    });
    expect(mockResolve).toHaveBeenCalledWith("u1");
    const detail = wrapper.findComponent(VrcUserCacheDetail);
    expect(detail.exists()).toBe(true);
    expect(detail.props("selected")?.displayName).toBe("Resolved");
  });

  it("uses query displayName as fallback detail when resolve fails", async () => {
    mockResolve.mockRejectedValue(new Error("network"));
    const wrapper = await mountWithQuery({
      vrcUserId: "u2",
      displayName: "FallbackName",
    });
    const detail = wrapper.findComponent(VrcUserCacheDetail);
    expect(detail.exists()).toBe(true);
    expect(detail.props("selected")?.displayName).toBe("FallbackName");
    expect(detail.props("selected")?.vrcUserId).toBe("u2");
  });
});
