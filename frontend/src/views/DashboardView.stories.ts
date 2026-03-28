import type { Meta, StoryObj } from "@storybook/vue3-vite";
import DashboardView from "./DashboardView.vue";

const meta = {
  title: "Views/DashboardView",
  component: DashboardView,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof DashboardView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
