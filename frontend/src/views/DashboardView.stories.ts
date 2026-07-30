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
      GetDashboardLaunchBlock: () =>
        Promise.resolve({
          profiles: [...sampleLaunchProfiles],
          selectedProfileId: sampleLaunchProfiles[0]?.id ?? "",
          rejoin: null,
        }),
      GetPresenceChangeSection: () =>
        Promise.resolve({
          loggedIn: true,
          status: "active",
          statusDescription: "",
          history: [],
        }),
      ApplyPresenceChange: (status: string, description: string) =>
        Promise.resolve({ status, statusDescription: description }),
      LaunchVRChat: () => Promise.resolve(),
    }),
  ],
} satisfies Meta<typeof DashboardView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
