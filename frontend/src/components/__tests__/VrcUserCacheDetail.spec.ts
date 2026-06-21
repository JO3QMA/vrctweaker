import { describe, expect, it, vi } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { nextTick } from "vue";
import VrcUserCacheDetail from "../VrcUserCacheDetail.vue";
import type { UserCacheDTO } from "../../wails/app";
import * as vrcUserCacheDisplay from "../../utils/vrcUserCacheDisplay";

function minimalUser(overrides: Partial<UserCacheDTO> = {}): UserCacheDTO {
  return {
    vrcUserId: "u_detail_1",
    displayName: "Detail Test",
    status: "active",
    isFavorite: false,
    lastUpdated: "2025-01-01",
    ...overrides,
  } as UserCacheDTO;
}

function mountDetail(
  selected: UserCacheDTO | null,
  stubs: Record<string, boolean | { template: string }> = {
    EncounterHistoryList: true,
  },
) {
  return mount(VrcUserCacheDetail, {
    props: { selected },
    global: { stubs },
  });
}

describe("VrcUserCacheDetail", () => {
  it("renders profile content inside Element Plus card body", async () => {
    const wrapper = mountDetail(minimalUser());
    await flushPromises();
    await nextTick();

    expect(wrapper.find(".friend-detail").exists()).toBe(true);
    expect(wrapper.find(".el-card__body").exists()).toBe(true);
    expect(wrapper.find(".profile-display-name").text()).toBe("Detail Test");
    expect(wrapper.find(".el-tabs").exists()).toBe(true);
    expect(wrapper.text()).toContain("遭遇履歴");
    expect(wrapper.find(".el-descriptions").exists()).toBe(true);
  });

  it("renders nothing when selected is null", () => {
    const wrapper = mountDetail(null);
    expect(wrapper.find(".friend-detail").exists()).toBe(false);
  });

  it("shows username, bio, status description, and bio links", async () => {
    const wrapper = mountDetail(
      minimalUser({
        username: "detail_user",
        bio: "Hello\nWorld",
        statusDescription: "Building worlds",
        bioLinksJson: '["https://example.com/a","https://example.com/b"]',
      }),
    );
    await flushPromises();

    expect(wrapper.find(".profile-handle").text()).toBe("@detail_user");
    expect(wrapper.find(".profile-bio").text()).toBe("Hello\nWorld");
    expect(wrapper.find(".profile-status-desc").text()).toBe("Building worlds");
    const links = wrapper.findAll(".profile-bio-links a");
    expect(links).toHaveLength(2);
    expect(links[0]?.attributes("href")).toBe("https://example.com/a");
    expect(links[1]?.text()).toBe("https://example.com/b");
  });

  it("emits favoriteChange when favorite checkbox toggles", async () => {
    const user = minimalUser({ isFavorite: false });
    const wrapper = mountDetail(user);
    await flushPromises();

    const checkbox = wrapper.find(".profile-toolbar-actions input");
    await checkbox.setValue(true);
    await flushPromises();

    expect(wrapper.emitted("favoriteChange")).toEqual([[user, true]]);
  });

  it("copies display name when copy button is clicked", async () => {
    const copySpy = vi
      .spyOn(vrcUserCacheDisplay, "copyDisplayName")
      .mockResolvedValue(undefined);
    const wrapper = mountDetail(minimalUser({ displayName: "Copy Me" }));
    await flushPromises();

    await wrapper
      .find('[data-testid="friend-copy-display-name"]')
      .trigger("click");
    await flushPromises();

    expect(copySpy).toHaveBeenCalledWith("Copy Me");
    copySpy.mockRestore();
  });

  it("shows avatar and banner images when URLs are available", async () => {
    const wrapper = mountDetail(
      minimalUser({
        currentAvatarThumbnailImageUrl: "https://cdn/avatar.png",
        profilePicOverride: "https://cdn/banner.png",
      }),
    );
    await flushPromises();

    expect(wrapper.find(".profile-avatar").attributes("src")).toBe(
      "https://cdn/avatar.png",
    );
    expect(wrapper.find(".profile-banner-img").attributes("src")).toBe(
      "https://cdn/banner.png",
    );
  });

  it("shows detail tab fields including tags and location", async () => {
    const wrapper = mountDetail(
      minimalUser({
        state: "online",
        location: "wrld_x:123~grp",
        developerType: "trusted",
        platform: "standalonewindows",
        lastPlatform: "android",
        tagsJson: '["system_trust_basic"]',
        currentAvatarTagsJson: '["author_tag"]',
      }),
    );
    await flushPromises();

    const text = wrapper.find(".el-descriptions").text();
    expect(text).toContain("online");
    expect(text).toContain("wrld_x:123~grp");
    expect(text).toContain("trusted");
    expect(text).toContain("standalonewindows");
    expect(text).toContain("system_trust_basic");
    expect(text).toContain("author_tag");
  });

  it("switches to encounters tab", async () => {
    const wrapper = mountDetail(minimalUser(), {
      EncounterHistoryList: {
        template: '<div data-testid="encounter-list-stub" />',
      },
    });
    await flushPromises();

    const tabs = wrapper.findAll(".el-tabs__item");
    const encountersTab = tabs.find((t) => t.text().includes("遭遇履歴"));
    await encountersTab?.trigger("click");
    await flushPromises();
    await nextTick();

    expect(wrapper.find('[data-testid="encounter-list-stub"]').exists()).toBe(
      true,
    );
  });

  it("resets tab to detail when selected user changes", async () => {
    const wrapper = mountDetail(minimalUser());
    await flushPromises();

    const tabs = wrapper.findAll(".el-tabs__item");
    const encountersTab = tabs.find((t) => t.text().includes("遭遇履歴"));
    await encountersTab?.trigger("click");
    await flushPromises();

    await wrapper.setProps({
      selected: minimalUser({
        vrcUserId: "u_detail_2",
        displayName: "Other User",
      }),
    });
    await flushPromises();
    await nextTick();

    expect(wrapper.find(".profile-display-name").text()).toBe("Other User");
    const activeTab = wrapper.find(".el-tabs__item.is-active");
    expect(activeTab.text()).toContain("詳細");
  });

  it("shows sticky header after scrolling card body", async () => {
    const wrapper = mountDetail(minimalUser());
    await flushPromises();
    await nextTick();

    const body = wrapper.find(".el-card__body").element as HTMLElement;
    const anchor = wrapper.find(".profile-name-row").element as HTMLElement;
    vi.spyOn(body, "getBoundingClientRect").mockReturnValue({
      top: 0,
      left: 0,
      right: 0,
      bottom: 400,
      width: 400,
      height: 400,
      x: 0,
      y: 0,
      toJSON: () => ({}),
    });
    vi.spyOn(anchor, "getBoundingClientRect").mockReturnValue({
      top: 0,
      left: 0,
      right: 0,
      bottom: 24,
      width: 200,
      height: 24,
      x: 0,
      y: 0,
      toJSON: () => ({}),
    });
    Object.defineProperty(body, "scrollTop", {
      configurable: true,
      value: 120,
      writable: true,
    });
    body.dispatchEvent(new Event("scroll"));
    await nextTick();

    expect(wrapper.find(".detail-sticky-header").classes()).toContain(
      "visible",
    );
  });

  it("shows extended detail rows and image links", async () => {
    const wrapper = mountDetail(
      minimalUser({
        lastLogin: "2025-01-01T00:00:00Z",
        lastActivity: "2025-01-02T00:00:00Z",
        lastMobile: "2025-01-03T00:00:00Z",
        friendKey: "fk_abc",
        currentAvatarImageUrl: "https://cdn/avatar-big.png",
        userIcon: "https://cdn/icon.png",
        imageUrl: "https://cdn/image.png",
        profilePicOverride: "https://cdn/override.png",
        profilePicOverrideThumbnail: "https://cdn/override-thumb.png",
      }),
    );
    await flushPromises();

    const text = wrapper.find(".el-descriptions").text();
    expect(text).toContain("2025-01-01T00:00:00Z");
    expect(text).toContain("2025-01-02T00:00:00Z");
    expect(text).toContain("2025-01-03T00:00:00Z");
    expect(text).toContain("fk_abc");
    expect(wrapper.find('a[href="https://cdn/avatar-big.png"]').exists()).toBe(
      true,
    );
    expect(wrapper.find('a[href="https://cdn/icon.png"]').exists()).toBe(true);
    expect(wrapper.find('a[href="https://cdn/image.png"]').exists()).toBe(true);
    expect(wrapper.find('a[href="https://cdn/override.png"]').exists()).toBe(
      true,
    );
    expect(
      wrapper.find('a[href="https://cdn/override-thumb.png"]').exists(),
    ).toBe(true);
  });

  it("shows placeholder avatar and banner when URLs are missing", async () => {
    const wrapper = mountDetail(minimalUser());
    await flushPromises();

    expect(wrapper.find(".profile-avatar--placeholder").exists()).toBe(true);
    expect(wrapper.find(".profile-banner-fallback").exists()).toBe(true);
  });
});
