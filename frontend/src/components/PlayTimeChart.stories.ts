import type { Meta, StoryObj } from "@storybook/vue3-vite";
import PlayTimeChart from "./PlayTimeChart.vue";

const meta = {
  title: "Components/PlayTimeChart",
  component: PlayTimeChart,
  tags: ["autodocs"],
} satisfies Meta<typeof PlayTimeChart>;

export default meta;
type Story = StoryObj<typeof meta>;

const sampleSeries = [
  { date: "2025-03-01", label: "3/1", seconds: 3600 },
  { date: "2025-03-02", label: "3/2", seconds: 7200 },
  { date: "2025-03-03", label: "3/3", seconds: 1800 },
];

export const Default: Story = {
  args: {
    series: sampleSeries,
  },
};

export const Empty: Story = {
  args: {
    series: [],
  },
};
