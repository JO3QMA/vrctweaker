import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { ref } from "vue";
import FriendsViewToolbar from "./FriendsViewToolbar.vue";

const meta = {
  title: "Views/FriendsView/Toolbar",
  component: FriendsViewToolbar,
  tags: ["autodocs"],
  parameters: {
    layout: "padded",
  },
} satisfies Meta<typeof FriendsViewToolbar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const LoggedIn: Story = {
  args: {
    isLoggedIn: true,
    refreshLoading: false,
    showOfflineList: false,
    displayNameQuery: "",
  },
  render: (args) => ({
    components: { FriendsViewToolbar },
    setup() {
      const showOfflineList = ref(args.showOfflineList);
      const displayNameQuery = ref(args.displayNameQuery);
      return { args, showOfflineList, displayNameQuery };
    },
    template: `
      <div style="max-width: 28rem">
        <FriendsViewToolbar
          v-model:show-offline-list="showOfflineList"
          v-model:display-name-query="displayNameQuery"
          :is-logged-in="args.isLoggedIn"
          :refresh-loading="args.refreshLoading"
          @refresh="() => {}"
        />
      </div>
    `,
  }),
};

export const NotLoggedIn: Story = {
  ...LoggedIn,
  args: {
    ...LoggedIn.args,
    isLoggedIn: false,
    refreshLoading: false,
  },
};

export const Refreshing: Story = {
  ...LoggedIn,
  args: {
    ...LoggedIn.args,
    isLoggedIn: true,
    refreshLoading: true,
  },
};
