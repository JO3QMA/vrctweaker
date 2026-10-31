import type { Meta, StoryObj } from "@storybook/vue3-vite";
import ServerStatusSection from "./ServerStatusSection.vue";
import { withWailsApp } from "../stories/wailsDecorator";

type MaintenanceFixture = {
  id: string;
  name: string;
  status: string;
  scheduledFor?: string;
  scheduledUntil?: string;
  updates?: string[];
};

function atLocal(dayOffset: number, hour: number, minute: number): string {
  const date = new Date();
  date.setDate(date.getDate() + dayOffset);
  date.setHours(hour, minute, 0, 0);
  return date.toISOString();
}

function statusDto(
  maintenances: MaintenanceFixture[],
  summary: { indicator: string; description: string } = {
    indicator: "none",
    description: "All Systems Operational",
  },
) {
  return {
    fetchState: "ok",
    summary,
    components: [],
    incidents: [],
    maintenances,
  };
}

function withStatus(dto: ReturnType<typeof statusDto>) {
  return withWailsApp({
    GetServerStatus: () => Promise.resolve(dto),
  });
}

const scheduled: MaintenanceFixture = {
  id: "scheduled",
  name: "API",
  status: "scheduled",
  scheduledFor: atLocal(2, 12, 34),
  scheduledUntil: atLocal(2, 23, 45),
  updates: ["API will be unavailable during this window."],
};

const inProgress: MaintenanceFixture = {
  id: "live",
  name: "API",
  status: "in_progress",
  scheduledFor: atLocal(0, 12, 34),
  scheduledUntil: atLocal(0, 23, 45),
  updates: ["Maintenance is in progress. Verification follows."],
};

const overnight: MaintenanceFixture = {
  id: "overnight",
  name: "API",
  status: "scheduled",
  scheduledFor: atLocal(1, 23, 0),
  scheduledUntil: atLocal(2, 2, 0),
  updates: ["The window crosses midnight in local time."],
};

const meta = {
  title: "Components/ServerStatusSection",
  component: ServerStatusSection,
  tags: ["autodocs"],
  parameters: {
    layout: "padded",
  },
} satisfies Meta<typeof ServerStatusSection>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Operational: Story = {
  decorators: [withStatus(statusDto([]))],
};

export const ScheduledNotice: Story = {
  decorators: [withStatus(statusDto([scheduled]))],
};

export const InProgress: Story = {
  decorators: [
    withStatus(
      statusDto([inProgress], {
        indicator: "maintenance",
        description: "Service Under Maintenance",
      }),
    ),
  ],
};

const openEndedScheduled: MaintenanceFixture = {
  id: "open-scheduled",
  name: "API",
  status: "scheduled",
  scheduledFor: atLocal(2, 12, 34),
  updates: ["No end time has been announced."],
};

const openEndedLive: MaintenanceFixture = {
  id: "open-live",
  name: "API",
  status: "in_progress",
  scheduledFor: atLocal(0, 9, 5),
  updates: ["Started without a published end time."],
};

export const Overnight: Story = {
  decorators: [withStatus(statusDto([overnight]))],
};

export const OpenEnded: Story = {
  decorators: [withStatus(statusDto([openEndedLive, openEndedScheduled]))],
};

export const Mixed: Story = {
  decorators: [
    withStatus(
      statusDto([
        {
          id: "ended",
          name: "Ended",
          status: "scheduled",
          scheduledFor: atLocal(-2, 1, 0),
          scheduledUntil: atLocal(-2, 2, 0),
          updates: ["This scheduled window has already ended."],
        },
        overnight,
        inProgress,
        scheduled,
      ]),
    ),
  ],
};
