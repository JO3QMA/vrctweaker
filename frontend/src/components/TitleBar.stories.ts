import type { Meta, StoryObj } from "@storybook/vue3-vite";
import TitleBar from "./TitleBar.vue";

const meta = {
  title: "Components/TitleBar",
  component: TitleBar,
  tags: ["autodocs"],
  parameters: {
    layout: "padded",
  },
} satisfies Meta<typeof TitleBar>;

export default meta;
type Story = StoryObj<typeof meta>;

/** Storybook では Wails runtime が無く、ウィンドウ操作ボタンは no-op */
export const Default: Story = {
  render: () => ({
    components: { TitleBar },
    template: `
      <div style="max-width: 720px; border: 1px solid var(--border); border-radius: var(--radius); overflow: hidden">
        <TitleBar />
      </div>
    `,
  }),
};
