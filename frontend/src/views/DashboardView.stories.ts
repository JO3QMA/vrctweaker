import type { Meta, StoryObj } from "@storybook/vue3-vite";
import DashboardView from "./DashboardView.vue";
import { dashboardViewWailsDecorator } from "./dashboardViewStoryDecorator";

const meta = {
  title: "Views/DashboardView",
  component: DashboardView,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
  decorators: [dashboardViewWailsDecorator()],
} satisfies Meta<typeof DashboardView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
