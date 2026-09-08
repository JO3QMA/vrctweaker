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
          "Flat thumbnail grid with sticky day section headers. Year dividers appear when the calendar year changes; day labels omit the year (e.g. 9月4日). Filter toolbar: criteria group (world search + date range) and actions group (refresh + scan folder), height-aligned at default control size.",
      },
    },
  },
} satisfies Meta<typeof GalleryView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
