import type { Meta, StoryObj } from "@storybook/vue3-vite";
import SettingsView from "./SettingsView.vue";

const meta = {
  title: "Views/SettingsView",
  component: SettingsView,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof SettingsView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
