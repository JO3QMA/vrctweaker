import type { Meta, StoryObj } from "@storybook/vue3-vite";
import LicensesView from "./LicensesView.vue";

const meta = {
  title: "Views/LicensesView",
  component: LicensesView,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof LicensesView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
