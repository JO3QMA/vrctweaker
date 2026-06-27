import type { Meta, StoryObj } from "@storybook/vue3-vite";
import DashboardView from "./DashboardView.vue";
import { sampleLaunchProfiles } from "../stories/fixtures/launcher";
import { withWailsApp } from "../stories/wailsDecorator";

const meta = {
  title: "Views/DashboardView",
  component: DashboardView,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
  decorators: [
    withWailsApp({
      LaunchProfiles: () => Promise.resolve([...sampleLaunchProfiles]),
      LaunchVRChat: () => Promise.resolve(),
      SetStatus: () => Promise.resolve(),
      SetStatusDescription: () => Promise.resolve(),
      SetStatusAndDescription: () => Promise.resolve(),
    }),
  ],
} satisfies Meta<typeof DashboardView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
