import type { Meta, StoryObj } from "@storybook/vue3-vite";
import VrcStatusTag from "./VrcStatusTag.vue";

const meta = {
  title: "Components/VrcStatusTag",
  component: VrcStatusTag,
  tags: ["autodocs"],
  argTypes: {
    status: { control: "text" },
  },
} satisfies Meta<typeof VrcStatusTag>;

export default meta;
type Story = StoryObj<typeof meta>;

export const JoinMe: Story = { args: { status: "join me" } };
export const AskMe: Story = { args: { status: "ask me" } };
export const Busy: Story = { args: { status: "busy" } };
export const Offline: Story = { args: { status: "offline" } };
export const Active: Story = { args: { status: "active" } };
export const Empty: Story = { args: { status: "" } };
