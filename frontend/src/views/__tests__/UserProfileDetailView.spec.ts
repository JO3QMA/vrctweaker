import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { createRouter, createWebHashHistory } from "vue-router";
import UserProfileDetailView from "../UserProfileDetailView.vue";
import VrcUserCacheDetail from "../../components/VrcUserCacheDetail.vue";

const { mockResolve, mockSetFavorite } = vi.hoisted(() => ({
  mockResolve: vi.fn(),
  mockSetFavorite: vi.fn(),
}));

vi.mock("../../wails/app", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../../wails/app")>();
  return {
    ...actual,
    App: {
      ...actual.App,
      resolveUserProfileNavigation: mockResolve,
      setFavorite: mockSetFavorite,
    },
  };
});

describe("UserProfileDetailView", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockResolve.mockResolvedValue({
      openInFriendsView: false,
      openInSelfProfile: false,
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

  it("redirects to me when resolve says self profile", async () => {
    mockResolve.mockResolvedValue({
      openInFriendsView: false,
      openInSelfProfile: true,
      user: {
        vrcUserId: "usr_me",
        displayName: "Me",
        status: "",
        isFavorite: false,
        lastUpdated: "",
      },
    });
    const router = createRouter({
      history: createWebHashHistory(),
      routes: [
        {
          path: "/user-profile",
          name: "user-profile",
          component: UserProfileDetailView,
        },
        { path: "/me", name: "me", component: { template: "<div/>" } },
      ],
    });
    const replaceSpy = vi.spyOn(router, "replace");
    await router.push({
      path: "/user-profile",
      query: { vrcUserId: "usr_me" },
    });
    await router.isReady();
    mount(UserProfileDetailView, { global: { plugins: [router] } });
    await flushPromises();

    expect(replaceSpy).toHaveBeenCalledWith({ name: "me" });
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

  it("shows warning when vrcUserId query is missing", async () => {
    const wrapper = await mountWithQuery({});
    expect(wrapper.find(".el-alert").exists()).toBe(true);
    expect(wrapper.text()).toContain("ユーザー ID");
    expect(wrapper.findComponent(VrcUserCacheDetail).exists()).toBe(false);
  });

  it("uses vrcUserId as display name when resolve omits user id", async () => {
    mockResolve.mockResolvedValue({
      openInFriendsView: false,
      openInSelfProfile: false,
      user: {
        vrcUserId: "",
        displayName: "",
        status: "",
        isFavorite: false,
        lastUpdated: "",
      },
    });
    const wrapper = await mountWithQuery({ vrcUserId: "usr_only" });
    const detail = wrapper.findComponent(VrcUserCacheDetail);
    expect(detail.props("selected")?.vrcUserId).toBe("usr_only");
    expect(detail.props("selected")?.displayName).toBe("usr_only");
  });

  it("reads displayName from array query values", async () => {
    mockResolve.mockRejectedValue(new Error("network"));
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
    await router.push({
      path: "/user-profile",
      query: { vrcUserId: "u-array", displayName: ["First", "Second"] },
    });
    await router.isReady();
    const wrapper = mount(UserProfileDetailView, {
      global: { plugins: [router] },
    });
    await flushPromises();

    expect(
      wrapper.findComponent(VrcUserCacheDetail).props("selected"),
    ).toMatchObject({
      vrcUserId: "u-array",
      displayName: "First",
    });
  });

  it("reloads when vrcUserId query changes", async () => {
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
    await router.push({
      path: "/user-profile",
      query: { vrcUserId: "first-id" },
    });
    await router.isReady();
    mount(UserProfileDetailView, { global: { plugins: [router] } });
    await flushPromises();
    expect(mockResolve).toHaveBeenCalledWith("first-id");

    mockResolve.mockClear();
    mockResolve.mockResolvedValue({
      openInFriendsView: false,
      openInSelfProfile: false,
      user: {
        vrcUserId: "second-id",
        displayName: "Second",
        status: "",
        isFavorite: false,
        lastUpdated: "",
      },
    });
    await router.push({
      path: "/user-profile",
      query: { vrcUserId: "second-id" },
    });
    await flushPromises();
    expect(mockResolve).toHaveBeenCalledWith("second-id");
  });

  it("persists favorite changes via App.setFavorite", async () => {
    mockSetFavorite.mockResolvedValue(undefined);
    const wrapper = await mountWithQuery({ vrcUserId: "u1" });
    const detail = wrapper.findComponent(VrcUserCacheDetail);

    await detail.vm.$emit("favorite-change", detail.props("selected"), true);
    await flushPromises();

    expect(mockSetFavorite).toHaveBeenCalledWith("u1", true);
    expect(detail.props("selected")?.isFavorite).toBe(true);
  });

  it("reverts favorite when App.setFavorite rejects", async () => {
    mockSetFavorite.mockRejectedValue(new Error("favorite failed"));
    const wrapper = await mountWithQuery({ vrcUserId: "u1" });
    const detail = wrapper.findComponent(VrcUserCacheDetail);
    expect(detail.props("selected")?.isFavorite).toBe(false);

    await detail.vm.$emit("favorite-change", detail.props("selected"), true);
    await flushPromises();

    expect(detail.props("selected")?.isFavorite).toBe(false);
  });
});
