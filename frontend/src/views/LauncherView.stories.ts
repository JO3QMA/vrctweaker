import type { Meta, StoryObj } from "@storybook/vue3-vite";
import LauncherView from "./LauncherView.vue";

const meta = {
  title: "Views/LauncherView",
  component: LauncherView,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof LauncherView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
