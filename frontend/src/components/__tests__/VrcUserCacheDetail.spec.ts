import { describe, expect, it } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { nextTick } from "vue";
import VrcUserCacheDetail from "../VrcUserCacheDetail.vue";
import type { UserCacheDTO } from "../../wails/app";

function minimalUser(): UserCacheDTO {
  return {
    vrcUserId: "u_detail_1",
    displayName: "Detail Test",
    status: "active",
    isFavorite: false,
    lastUpdated: "",
  } as UserCacheDTO;
}

describe("VrcUserCacheDetail", () => {
  it("renders profile content inside Element Plus card body", async () => {
    const wrapper = mount(VrcUserCacheDetail, {
      props: { selected: minimalUser() },
    });
    await flushPromises();
    await nextTick();

    expect(wrapper.find(".friend-detail").exists()).toBe(true);
    expect(wrapper.find(".el-card__body").exists()).toBe(true);
    expect(wrapper.find(".profile-display-name").text()).toBe("Detail Test");
  });
});
