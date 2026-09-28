import type { Decorator, Meta, StoryObj } from "@storybook/vue3-vite";
import FriendAvatarList from "./FriendAvatarList.vue";
import { withWailsApp } from "../stories/wailsDecorator";

function withFriendAvatarWails(
  overrides: Record<string, unknown> = {},
): Decorator {
  return withWailsApp({
    FriendAvatarUsageByVRCUserID: () =>
      Promise.resolve([
        {
          avatarName: "Story Avatar",
          useCount: 3,
          firstSeenAt: "2026-01-01T12:00:00.000Z",
          lastSeenAt: "2026-02-01T12:00:00.000Z",
        },
      ]),
    ...overrides,
  });
}

const meta = {
  title: "Components/FriendAvatarList",
  component: FriendAvatarList,
  tags: ["autodocs"],
  parameters: {
    layout: "padded",
  },
} satisfies Meta<typeof FriendAvatarList>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  decorators: [withFriendAvatarWails()],
  args: { userId: "usr_story" },
  render: (args) => ({
    components: { FriendAvatarList },
    setup() {
      return { args };
    },
    template: `
      <div style="max-width: 56rem">
        <FriendAvatarList v-bind="args" />
      </div>
    `,
  }),
};

export const Empty: Story = {
  decorators: [
    withFriendAvatarWails({
      FriendAvatarUsageByVRCUserID: () => Promise.resolve([]),
    }),
  ],
  args: { userId: "usr_empty" },
  render: (args) => ({
    components: { FriendAvatarList },
    setup() {
      return { args };
    },
    template: `
      <div style="max-width: 56rem">
        <FriendAvatarList v-bind="args" />
      </div>
    `,
  }),
};
