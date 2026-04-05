import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { ref } from "vue";
import type { UserCacheDTO } from "../../wails/app";
import FriendsListPanel from "./FriendsListPanel.vue";
import { sampleFriendsList } from "./friendsSampleData";

const meta = {
  title: "Views/FriendsView/List",
  component: FriendsListPanel,
  tags: ["autodocs"],
  parameters: {
    layout: "padded",
  },
} satisfies Meta<typeof FriendsListPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithSamples: Story = {
  args: {
    friends: [],
    selected: null,
    loading: false,
    emptyMessage: "",
  },
  render: () => ({
    components: { FriendsListPanel },
    setup() {
      const friends = ref([...sampleFriendsList]);
      const selected = ref(sampleFriendsList[0] ?? null);
      return { friends, selected };
    },
    template: `
      <FriendsListPanel
        :friends="friends"
        :selected="selected"
        :loading="false"
        empty-message="該当するフレンドはいません"
        @select="selected = $event"
        @toggle-favorite="(f) => { f.isFavorite = !f.isFavorite }"
      />
    `,
  }),
};

export const Empty: Story = {
  args: {
    friends: [],
    selected: null,
    loading: false,
    emptyMessage: "オンラインのフレンドはいません",
  },
  render: () => ({
    components: { FriendsListPanel },
    setup() {
      const friends = ref<UserCacheDTO[]>([]);
      const selected = ref(null);
      return { friends, selected };
    },
    template: `
      <FriendsListPanel
        :friends="friends"
        :selected="selected"
        :loading="false"
        empty-message="オンラインのフレンドはいません"
        @select="selected = $event"
        @toggle-favorite="() => {}"
      />
    `,
  }),
};
