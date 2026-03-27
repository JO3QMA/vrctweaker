import type { Meta, StoryObj } from "@storybook/vue3-vite";
import AutomationView from "./AutomationView.vue";

const meta = {
  title: "Views/AutomationView",
  component: AutomationView,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof AutomationView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
