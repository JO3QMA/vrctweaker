import type { Meta, StoryObj } from "@storybook/vue3-vite";
import EncounterHistoryDetailView from "./EncounterHistoryDetailView.vue";

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

export const Default: Story = {};
