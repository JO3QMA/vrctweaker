import type { Meta, StoryObj } from "@storybook/vue3-vite";
import GalleryView from "./GalleryView.vue";

const meta = {
  title: "Views/GalleryView",
  component: GalleryView,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof GalleryView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
