import type { Meta, StoryObj } from "@storybook/vue3-vite";
import EncounterHistoryDetailView from "./EncounterHistoryDetailView.vue";
import {
  encounterHistoryDetailRouterDecorator,
  encounterHistoryDetailWailsDecorator,
} from "./encounterHistoryDetailViewStoryDecorator";

const meta = {
  title: "Views/EncounterHistoryDetailView",
  component: EncounterHistoryDetailView,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof EncounterHistoryDetailView>;

export default meta;
type Story = StoryObj<typeof meta>;

/** クエリ不正 → 警告のみ（一覧なし） */
export const InvalidQuery: Story = {};

/** ユーザー別遭遇履歴 + サンプル行 */
export const ByUser: Story = {
  decorators: [
    encounterHistoryDetailRouterDecorator({
      kind: "user",
      vrcUserId: "usr_enc_story",
    }),
    encounterHistoryDetailWailsDecorator("user"),
  ],
};

/** ワールド別遭遇履歴 + サンプル行 */
export const ByWorld: Story = {
  decorators: [
    encounterHistoryDetailRouterDecorator({
      kind: "world",
      worldId: "wrld_enc_story",
    }),
    encounterHistoryDetailWailsDecorator("world"),
  ],
};
