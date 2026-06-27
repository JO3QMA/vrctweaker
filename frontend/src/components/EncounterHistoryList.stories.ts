import type { Meta, StoryObj } from "@storybook/vue3-vite";
import EncounterHistoryList from "./EncounterHistoryList.vue";
import { userProfileEncounters } from "../stories/fixtures/encounters";
import { withWailsApp } from "../stories/wailsDecorator";

const meta = {
  title: "Components/EncounterHistoryList",
  component: EncounterHistoryList,
  tags: ["autodocs"],
  decorators: [
    withWailsApp({
      EncountersByVRCUserID: () => Promise.resolve(userProfileEncounters),
      EncountersByWorldID: () => Promise.resolve([]),
    }),
  ],
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
