import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import FriendsListPanel from "../FriendsListPanel.vue";
import type { UserCacheDTO } from "../../../wails/app";

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

describe("FriendsListPanel", () => {
  it("renders thumbnail when avatar url is available", () => {
    const user = minimalUser({
      vrcUserId: "1",
      displayName: "ThumbUser",
      status: "active",
      currentAvatarThumbnailImageUrl: "https://example.com/thumb.png",
    });
    const wrapper = mount(FriendsListPanel, {
      props: {
        friends: [user],
        selected: null,
        loading: false,
        emptyMessage: "empty",
      },
    });

    const img = wrapper.find(".friend-thumb");
    expect(img.element.tagName).toBe("IMG");
    expect(img.attributes("src")).toBe("https://example.com/thumb.png");
  });

  it("renders placeholder when no thumbnail url", () => {
    const wrapper = mount(FriendsListPanel, {
      props: {
        friends: [
          minimalUser({
            vrcUserId: "1",
            displayName: "NoThumb",
            status: "active",
          }),
        ],
        selected: null,
        loading: false,
        emptyMessage: "empty",
      },
    });

    expect(wrapper.find(".friend-thumb-placeholder").exists()).toBe(true);
  });

  it("emits select when a friend card is clicked", async () => {
    const user = minimalUser({
      vrcUserId: "1",
      displayName: "PickMe",
      status: "active",
    });
    const wrapper = mount(FriendsListPanel, {
      props: {
        friends: [user],
        selected: null,
        loading: false,
        emptyMessage: "empty",
      },
    });

    await wrapper.find(".friend-card").trigger("click");
    expect(wrapper.emitted("select")?.[0]).toEqual([user]);
  });

  it("emits toggleFavorite when star button is clicked", async () => {
    const user = minimalUser({
      vrcUserId: "1",
      displayName: "StarMe",
      status: "active",
      isFavorite: true,
    });
    const wrapper = mount(FriendsListPanel, {
      props: {
        friends: [user],
        selected: user,
        loading: false,
        emptyMessage: "empty",
      },
    });

    await wrapper.find(".btn-favorite").trigger("click");
    expect(wrapper.emitted("toggleFavorite")?.[0]).toEqual([user]);
  });

  it("shows empty message only when not loading", () => {
    const wrapper = mount(FriendsListPanel, {
      props: {
        friends: [],
        selected: null,
        loading: false,
        emptyMessage: "オンラインのフレンドはいません",
      },
    });

    expect(wrapper.find(".empty-message").text()).toBe(
      "オンラインのフレンドはいません",
    );
  });

  it("hides empty message while loading", () => {
    const wrapper = mount(FriendsListPanel, {
      props: {
        friends: [],
        selected: null,
        loading: true,
        emptyMessage: "オンラインのフレンドはいません",
      },
    });

    expect(wrapper.find(".empty-message").exists()).toBe(false);
  });

  it("marks selected friend card as active", () => {
    const user = minimalUser({
      vrcUserId: "1",
      displayName: "ActiveUser",
      status: "active",
    });
    const wrapper = mount(FriendsListPanel, {
      props: {
        friends: [user],
        selected: user,
        loading: false,
        emptyMessage: "empty",
      },
    });

    expect(wrapper.find(".friend-card.active").exists()).toBe(true);
  });

  it("uses different favorite button title for favorite state", () => {
    const favorite = minimalUser({
      vrcUserId: "1",
      displayName: "Fav",
      status: "active",
      isFavorite: true,
    });
    const notFavorite = minimalUser({
      vrcUserId: "2",
      displayName: "Plain",
      status: "active",
      isFavorite: false,
    });
    const wrapper = mount(FriendsListPanel, {
      props: {
        friends: [favorite, notFavorite],
        selected: null,
        loading: false,
        emptyMessage: "empty",
      },
    });

    const buttons = wrapper.findAll(".btn-favorite");
    expect(buttons[0]!.attributes("title")).toBe("お気に入り解除");
    expect(buttons[1]!.attributes("title")).toBe("お気に入り登録");
  });
});
