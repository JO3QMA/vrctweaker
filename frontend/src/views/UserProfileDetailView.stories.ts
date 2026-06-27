import type { Meta, StoryObj } from "@storybook/vue3-vite";
import UserProfileDetailView from "./UserProfileDetailView.vue";
import { userProfileEncounters } from "../stories/fixtures/encounters";
import { withRouter, withWailsApp } from "../stories/wailsDecorator";

const meta = {
  title: "Views/UserProfileDetailView",
  component: UserProfileDetailView,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof UserProfileDetailView>;

export default meta;
type Story = StoryObj<typeof meta>;

/** vrcUserId なし → 警告のみ */
export const MissingUserId: Story = {
  decorators: [withRouter({ name: "user-profile", query: {} })],
};

/** 解決成功。遭遇タブ用に Encounters スタブあり */
export const ResolvedProfile: Story = {
  decorators: [
    withRouter({
      name: "user-profile",
      query: { vrcUserId: "usr_story", displayName: "Hint Name" },
    }),
    withWailsApp({
      ResolveUserProfileNavigation: (id: string) =>
        Promise.resolve({
          user: {
            vrcUserId: id,
            displayName: "Story User",
            status: "active",
            isFavorite: false,
            lastUpdated: "",
          },
          openInFriendsView: false,
          openInSelfProfile: false,
        }),
      SetFavorite: () => Promise.resolve(),
      EncountersByVRCUserID: () => Promise.resolve(userProfileEncounters),
      EncountersByWorldID: () => Promise.resolve([]),
    }),
  ],
};

/** resolve 失敗時はクエリの displayName でフォールバック表示 */
export const ResolveFailedUsesQueryDisplayName: Story = {
  decorators: [
    withRouter({
      name: "user-profile",
      query: { vrcUserId: "u2", displayName: "FallbackName" },
    }),
    withWailsApp({
      ResolveUserProfileNavigation: () => Promise.reject(new Error("network")),
      SetFavorite: () => Promise.resolve(),
      EncountersByVRCUserID: () => Promise.resolve([]),
      EncountersByWorldID: () => Promise.resolve([]),
    }),
  ],
};
