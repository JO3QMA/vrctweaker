import type { Meta, StoryObj } from "@storybook/vue3-vite";
import VrcUserTagChip from "./VrcUserTagChip.vue";

const meta = {
  title: "Components/VrcUserTagChip",
  component: VrcUserTagChip,
  tags: ["autodocs"],
  argTypes: {
    tag: { control: "text" },
  },
} satisfies Meta<typeof VrcUserTagChip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const TrustBasic: Story = { args: { tag: "system_trust_basic" } };
export const LanguageJapanese: Story = { args: { tag: "language_jpn" } };
export const Deprecated: Story = { args: { tag: "show_social_rank" } };
export const Unknown: Story = { args: { tag: "system_slug" } };
