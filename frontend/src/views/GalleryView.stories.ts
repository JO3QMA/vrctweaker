import type { Meta, StoryObj } from "@storybook/vue3-vite";
import GalleryView from "./GalleryView.vue";

const meta = {
  title: "Views/GalleryView",
  component: GalleryView,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
    docs: {
      description: {
        component:
          "Filter bar: world search (wrld_ prefix → ID, else world name) and date range picker.",
      },
    },
  },
} satisfies Meta<typeof GalleryView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
