import type { Meta, StoryObj } from "@storybook/vue3-vite";
import FriendsView from "./FriendsView.vue";

const meta = {
  title: "Views/FriendsView",
  component: FriendsView,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof FriendsView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
