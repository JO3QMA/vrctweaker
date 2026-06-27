import type { Meta, StoryObj } from "@storybook/vue3-vite";
import FriendsView from "./FriendsView.vue";
import { sampleFriendsList } from "../stories/fixtures/friends";
import { withWailsApp } from "../stories/wailsDecorator";

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
  decorators: [
    withWailsApp({
      Friends: () => Promise.resolve(sampleFriendsList),
      IsLoggedIn: () => Promise.resolve(true),
      RefreshFriends: () => Promise.resolve(),
      ReconcileVRChatSocialCache: () => Promise.resolve(),
      SetFavorite: () => Promise.resolve(),
      HasStoredCredential: () => Promise.resolve(false),
      GetCredentialBlob: () => Promise.resolve(""),
      UnlockVRChatSession: (_token: string) => Promise.resolve(),
      PersistWrappedCredential: (_blob: string) => Promise.resolve(),
      ClearStoredCredential: () => Promise.resolve(),
    }),
  ],
};
