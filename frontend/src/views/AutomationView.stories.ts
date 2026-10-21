import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { expect, userEvent, within } from "storybook/test";
import AutomationView from "./AutomationView.vue";
import { withAutomationWails } from "../stories/fixtures/automationWails";

const meta = {
  title: "Views/AutomationView",
  component: AutomationView,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
  decorators: [withAutomationWails()],
} satisfies Meta<typeof AutomationView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Rules: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(
      await canvas.findByText("フレンド参加で取り込み中"),
    ).toBeVisible();
  },
};

export const EditingRule: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(await canvas.findByText("フレンド参加で取り込み中"));
    await expect(canvas.getByTestId("item-name-input")).toHaveValue(
      "フレンド参加で取り込み中",
    );
    await expect(canvas.getByText("いつ（トリガー）")).toBeVisible();
  },
};

export const RunLog: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(await canvas.findByRole("tab", { name: /実行ログ/ }));
    await expect(canvas.getByText("スケジュール（毎分評価）")).toBeVisible();
  },
};

export const NarrowList: Story = {
  decorators: [
    (story) => ({
      components: { story },
      template: '<div style="width: 600px; height: 720px"><story /></div>',
    }),
  ],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(
      await canvas.findByText("フレンド参加で取り込み中"),
    ).toBeVisible();
  },
};
