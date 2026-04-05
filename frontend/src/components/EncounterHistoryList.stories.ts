import type { Meta, StoryObj } from "@storybook/vue3-vite";
import EncounterHistoryList from "./EncounterHistoryList.vue";
import type { UserEncounterDTO } from "../wails/app";
import { wailsEncountersByUserDecorator } from "./vrcUserCacheDetailStoryDecorator";

const rows: UserEncounterDTO[] = [
  {
    id: "e1",
    vrcUserId: "usr_story",
    displayName: "Story User",
    instanceId: "inst_1",
    worldId: "wrld_1",
    worldDisplayName: "Test World",
    joinedAt: "2026-02-01T10:00:00+09:00",
    leftAt: "2026-02-01T11:00:00+09:00",
  },
];

const meta = {
  title: "Components/EncounterHistoryList",
  component: EncounterHistoryList,
  tags: ["autodocs"],
  decorators: [wailsEncountersByUserDecorator(rows)],
  parameters: {
    layout: "padded",
  },
} satisfies Meta<typeof EncounterHistoryList>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ByUser: Story = {
  args: {
    mode: "user",
    userId: "usr_story",
    hideDisplayNameColumn: true,
  },
  render: (args) => ({
    components: { EncounterHistoryList },
    setup() {
      return { args };
    },
    template: `
      <div style="max-width: 56rem">
        <EncounterHistoryList v-bind="args" />
      </div>
    `,
  }),
};
