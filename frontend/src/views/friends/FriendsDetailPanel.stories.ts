import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { ref } from "vue";
import FriendsDetailPanel from "./FriendsDetailPanel.vue";
import { sampleFriendsList } from "./friendsSampleData";

const meta = {
  title: "Views/FriendsView/Detail",
  component: FriendsDetailPanel,
  tags: ["autodocs"],
  parameters: {
    layout: "padded",
  },
} satisfies Meta<typeof FriendsDetailPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithSampleUser: Story = {
  args: {
    selected: null,
  },
  render: () => ({
    components: { FriendsDetailPanel },
    setup() {
      const selected = ref({ ...sampleFriendsList[0]! });
      return { selected };
    },
    template: `
      <div style="max-width: 42rem; height: 28rem; display: flex; flex-direction: column; min-height: 0">
        <FriendsDetailPanel
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
    components: { FriendsDetailPanel },
    setup() {
      const selected = ref(null);
      return { selected };
    },
    template: `<FriendsDetailPanel :selected="selected" />`,
  }),
};
