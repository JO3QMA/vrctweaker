import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import { defineComponent } from "vue";
import FriendsDetailPane from "../friends/FriendsDetailPane.vue";
import type { UserCacheDTO } from "../../wails/app";

function minimalUser(): UserCacheDTO {
  return {
    vrcUserId: "u_1",
    displayName: "Test User",
    status: "active",
    isFavorite: false,
    lastUpdated: "",
  } as UserCacheDTO;
}

describe("FriendsDetailPane", () => {
  it("renders right pane container", () => {
    const wrapper = mount(FriendsDetailPane, {
      props: { selected: null },
    });
    expect(wrapper.find(".friends-detail-pane").exists()).toBe(true);
  });

  it("re-emits favorite change from detail panel", async () => {
    const panelStub = defineComponent({
      name: "FriendsDetailPanel",
      emits: ["favorite-change"],
      data() {
        return { user: minimalUser() };
      },
      template: "<button @click=\"$emit('favorite-change', user, true)\" />",
    });

    const wrapper = mount(FriendsDetailPane, {
      props: { selected: minimalUser() },
      global: {
        stubs: {
          FriendsDetailPanel: panelStub,
        },
      },
    });
    await wrapper.find("button").trigger("click");

    expect(wrapper.emitted("favoriteChange")).toEqual([[minimalUser(), true]]);
  });
});
