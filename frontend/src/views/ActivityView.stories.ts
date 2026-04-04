import type { Meta, StoryObj } from "@storybook/vue3-vite";
import ActivityView from "./ActivityView.vue";
import {
  activityViewWailsDecorator,
  sampleActivityEncounters,
  sampleActivityStats,
} from "./activityViewStoryDecorator";

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
    activityViewWailsDecorator(sampleActivityEncounters, sampleActivityStats()),
  ],
};
