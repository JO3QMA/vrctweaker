import type { Meta, StoryObj } from "@storybook/vue3-vite";
import ActivityView from "./ActivityView.vue";
import { sampleActivityStats } from "../stories/fixtures/activity";
import { sampleActivityEncounters } from "../stories/fixtures/encounters";
import { withWailsApp } from "../stories/wailsDecorator";

const meta = {
  title: "Views/ActivityView",
  component: ActivityView,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof ActivityView>;

export default meta;
type Story = StoryObj<typeof meta>;

/** Wails なし（空データ）。折りたたみヘッダーとレイアウトの確認用 */
export const Default: Story = {};

/** サンプル遭遇ログ・プレイ時間で折りたたみカードの中身まで確認 */
export const WithSampleData: Story = {
  decorators: [
    withWailsApp({
      Encounters: () => Promise.resolve(sampleActivityEncounters),
      GetActivityStats: (_from: string, _to: string) =>
        Promise.resolve(sampleActivityStats()),
      GetLogRetentionDays: () => Promise.resolve(30),
      ResolveUserProfileNavigation: (id: string) =>
        Promise.resolve({
          user: {
            vrcUserId: id,
            displayName: "mock",
            status: "",
            isFavorite: false,
            lastUpdated: "",
          },
          openInFriendsView: false,
          openInSelfProfile: false,
        }),
    }),
  ],
};
