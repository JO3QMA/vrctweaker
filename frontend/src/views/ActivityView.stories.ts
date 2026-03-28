import type { Meta, StoryObj } from "@storybook/vue3-vite";
import ActivityView from "./ActivityView.vue";

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

export const Default: Story = {};
