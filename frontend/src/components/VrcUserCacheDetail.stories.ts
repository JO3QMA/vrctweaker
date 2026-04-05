import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { ref } from "vue";
import VrcUserCacheDetail from "./VrcUserCacheDetail.vue";
import type { UserEncounterDTO } from "../wails/app";
import { sampleFriendsList } from "../views/friends/friendsSampleData";
import { wailsEncountersByUserDecorator } from "./vrcUserCacheDetailStoryDecorator";

const storyEncounters: UserEncounterDTO[] = [
  {
    id: "enc_sb_1",
    vrcUserId: sampleFriendsList[0]!.vrcUserId,
    displayName: sampleFriendsList[0]!.displayName,
    instanceId: "instance~abc",
    worldId: "wrld_sample_001",
    worldDisplayName: "Storybook 用ワールド",
    joinedAt: "2026-01-15T12:00:00+09:00",
    leftAt: "2026-01-15T13:30:00+09:00",
  },
];

const meta = {
  title: "Components/VrcUserCacheDetail",
  component: VrcUserCacheDetail,
  tags: ["autodocs"],
  parameters: {
    layout: "padded",
  },
} satisfies Meta<typeof VrcUserCacheDetail>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithSampleUser: Story = {
  args: {
    selected: null,
  },
  decorators: [wailsEncountersByUserDecorator(storyEncounters)],
  render: () => ({
    components: { VrcUserCacheDetail },
    setup() {
      const selected = ref({ ...sampleFriendsList[0]! });
      return { selected };
    },
    template: `
      <div style="max-width: 42rem; height: 28rem; display: flex; flex-direction: column; min-height: 0">
        <VrcUserCacheDetail
          :selected="selected"
          @favorite-change="(f, v) => { f.isFavorite = v }"
        />
      </div>
    `,
  }),
};

export const NoSelection: Story = {
  args: {
    selected: null,
  },
  render: () => ({
    components: { VrcUserCacheDetail },
    setup() {
      const selected = ref(null);
      return { selected };
    },
    template: `<VrcUserCacheDetail :selected="selected" />`,
  }),
};
