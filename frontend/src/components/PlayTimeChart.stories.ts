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

/** 最大10時間・14日分。Y軸は 0時間〜10時間、右端の日付ラベルが切れないこと。 */
const fourteenDayPeakTenHours = [
  { date: "2025-10-17", label: "10/17", seconds: 0 },
  { date: "2025-10-18", label: "10/18", seconds: 5400 },
  { date: "2025-10-19", label: "10/19", seconds: 7200 },
  { date: "2025-10-20", label: "10/20", seconds: 3600 },
  { date: "2025-10-21", label: "10/21", seconds: 10800 },
  { date: "2025-10-22", label: "10/22", seconds: 5400 },
  { date: "2025-10-23", label: "10/23", seconds: 1800 },
  { date: "2025-10-24", label: "10/24", seconds: 0 },
  { date: "2025-10-25", label: "10/25", seconds: 9000 },
  { date: "2025-10-26", label: "10/26", seconds: 36000 },
  { date: "2025-10-27", label: "10/27", seconds: 5400 },
  { date: "2025-10-28", label: "10/28", seconds: 0 },
  { date: "2025-10-29", label: "10/29", seconds: 7200 },
  { date: "2025-10-30", label: "10/30", seconds: 5400 },
];

export const FourteenDaysPeakTenHours: Story = {
  args: {
    series: fourteenDayPeakTenHours,
  },
  decorators: [
    () => ({
      template:
        '<div style="width: 780px; padding: 16px; background: var(--color-bg-elevated);"><story /></div>',
    }),
  ],
};

/** PR / visual check: card ~720px. */
export const FourteenDaysPeakTenHoursWidth720: Story = {
  args: {
    series: fourteenDayPeakTenHours,
  },
  decorators: [
    () => ({
      template:
        '<div style="width: 720px; padding: 16px; background: var(--color-bg-elevated);"><story /></div>',
    }),
  ],
};

/** PR / visual check: card ~960px. */
export const FourteenDaysPeakTenHoursWidth960: Story = {
  args: {
    series: fourteenDayPeakTenHours,
  },
  decorators: [
    () => ({
      template:
        '<div style="width: 960px; padding: 16px; background: var(--color-bg-elevated);"><story /></div>',
    }),
  ],
};

/** PR / visual check: narrow card (~600px). */
export const FourteenDaysPeakTenHoursWidth600: Story = {
  args: {
    series: fourteenDayPeakTenHours,
  },
  decorators: [
    () => ({
      template:
        '<div style="width: 600px; padding: 16px; background: var(--color-bg-elevated);"><story /></div>',
    }),
  ],
};

/** PR / visual check: wide card (~1280px). */
export const FourteenDaysPeakTenHoursWidth1280: Story = {
  args: {
    series: fourteenDayPeakTenHours,
  },
  decorators: [
    () => ({
      template:
        '<div style="width: 1280px; padding: 16px; background: var(--color-bg-elevated);"><story /></div>',
    }),
  ],
};

/** 最大55分: Y軸は 0分〜60分（小数時間ラベルなし）。 */
export const FiftyFiveMinuteDailyMax: Story = {
  args: {
    series: [
      { date: "2025-03-01", label: "3/1", seconds: 1800 },
      { date: "2025-03-02", label: "3/2", seconds: 3300 },
      { date: "2025-03-03", label: "3/3", seconds: 900 },
      { date: "2025-03-04", label: "3/4", seconds: 2700 },
      { date: "2025-03-05", label: "3/5", seconds: 0 },
    ],
  },
  decorators: [
    () => ({
      template:
        '<div style="width: 720px; padding: 16px; background: var(--color-bg-elevated);"><story /></div>',
    }),
  ],
};

/** 最大が1時間未満のとき Y軸はすべて「分」。 */
export const SubHourScale: Story = {
  args: {
    series: [
      { date: "2025-03-01", label: "3/1", seconds: 0 },
      { date: "2025-03-02", label: "3/2", seconds: 900 },
      { date: "2025-03-03", label: "3/3", seconds: 1800 },
      { date: "2025-03-04", label: "3/4", seconds: 600 },
    ],
  },
};
