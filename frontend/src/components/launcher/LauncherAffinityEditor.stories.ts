import type { Meta, StoryObj } from "@storybook/vue3-vite";
import LauncherAffinityEditor from "./LauncherAffinityEditor.vue";
import { withWailsApp } from "../../stories/wailsDecorator";

const meta = {
  title: "Launcher/LauncherAffinityEditor",
  component: LauncherAffinityEditor,
  tags: ["autodocs"],
  decorators: [
    withWailsApp({
      GetLogicalProcessorCount: () => Promise.resolve(32),
    }),
  ],
  args: {
    modelValue: "FFFF",
  },
} satisfies Meta<typeof LauncherAffinityEditor>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default32Cores: Story = {
  args: {
    modelValue: "FFFFFFFF",
  },
};

export const LowerCcdOnly: Story = {
  args: {
    modelValue: "FFFF",
  },
};
