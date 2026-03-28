import type { Meta, StoryObj } from "@storybook/vue3-vite";
import FriendsView from "./FriendsView.vue";
import { sampleFriendsList } from "./friends/friendsSampleData";
import { wailsFriendsLoggedInDecorator } from "./friends/wailsFriendsStoryDecorator";

const meta = {
  title: "Views/FriendsView",
  component: FriendsView,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof FriendsView>;

export default meta;
type Story = StoryObj<typeof meta>;

/** Wails 無し: isLoggedIn=false・friends=[]（本番の未ログインに近い） */
export const Default: Story = {};

/**
 * ログイン済み＋サンプルフレンド（更新・お気に入りは no-op）。
 * ツールバー／一覧／詳細は各サブストーリー（Views/FriendsView/Toolbar 等）でも個別に確認できます。
 */
export const LoggedInWithSampleFriends: Story = {
  decorators: [wailsFriendsLoggedInDecorator(sampleFriendsList)],
};
