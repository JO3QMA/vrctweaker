import type { Meta, StoryObj } from "@storybook/vue3-vite";
import EncounterHistoryDetailView from "./EncounterHistoryDetailView.vue";
import {
  encounterByUserSample,
  encounterByWorldSample,
} from "../stories/fixtures/encounters";
import { withRouter, withWailsApp } from "../stories/wailsDecorator";

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
    withRouter({
      name: "encounter-history",
      query: { kind: "user", vrcUserId: "usr_enc_story" },
    }),
    withWailsApp({
      EncountersByVRCUserID: () => Promise.resolve(encounterByUserSample),
      EncountersByWorldID: () => Promise.resolve([]),
    }),
  ],
};

/** ワールド別遭遇履歴 + サンプル行 */
export const ByWorld: Story = {
  decorators: [
    withRouter({
      name: "encounter-history",
      query: { kind: "world", worldId: "wrld_enc_story" },
    }),
    withWailsApp({
      EncountersByVRCUserID: () => Promise.resolve([]),
      EncountersByWorldID: () => Promise.resolve(encounterByWorldSample),
    }),
  ],
};
