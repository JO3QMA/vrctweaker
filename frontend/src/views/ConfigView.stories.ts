import type { Meta, StoryObj } from "@storybook/vue3-vite";
import ConfigView from "./ConfigView.vue";

const meta = {
  title: "Views/ConfigView",
  component: ConfigView,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof ConfigView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
