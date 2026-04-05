import type { Meta, StoryObj } from "@storybook/vue3-vite";
import UserProfileDetailView from "./UserProfileDetailView.vue";
import {
  userProfileDetailRouterDecorator,
  userProfileDetailWailsDecorator,
} from "./userProfileDetailViewStoryDecorator";

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
  decorators: [userProfileDetailRouterDecorator({})],
};

/** 解決成功。遭遇タブ用に Encounters スタブあり */
export const ResolvedProfile: Story = {
  decorators: [
    userProfileDetailRouterDecorator({
      vrcUserId: "usr_story",
      displayName: "Hint Name",
    }),
    userProfileDetailWailsDecorator({}),
  ],
};

/** resolve 失敗時はクエリの displayName でフォールバック表示 */
export const ResolveFailedUsesQueryDisplayName: Story = {
  decorators: [
    userProfileDetailRouterDecorator({
      vrcUserId: "u2",
      displayName: "FallbackName",
    }),
    userProfileDetailWailsDecorator({
      resolveUserProfileNavigation: () => Promise.reject(new Error("network")),
      encountersByUser: [],
    }),
  ],
};
