import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { ref } from "vue";
import FriendsDetailPane from "./FriendsDetailPane.vue";
import { sampleFriendsList } from "./friendsSampleData";

const meta = {
  title: "Views/FriendsView/DetailPane",
  component: FriendsDetailPane,
  tags: ["autodocs"],
  parameters: {
    layout: "padded",
  },
} satisfies Meta<typeof FriendsDetailPane>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithSampleUser: Story = {
  args: {
    selected: null,
  },
  render: () => ({
    components: { FriendsDetailPane },
    setup() {
      const selected = ref({ ...sampleFriendsList[0]! });
      return { selected };
    },
    template: `
      <div style="max-width: 42rem; height: 28rem; display: flex; flex-direction: column; min-height: 0">
        <FriendsDetailPane
          :selected="selected"
          @favorite-change="(u, v) => { u.isFavorite = v }"
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
    components: { FriendsDetailPane },
    setup() {
      const selected = ref(null);
      return { selected };
    },
    template: `<FriendsDetailPane :selected="selected" />`,
  }),
};
