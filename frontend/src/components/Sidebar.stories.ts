import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { onMounted } from "vue";
import { useRouter } from "vue-router";
import Sidebar from "./Sidebar.vue";

function routePathDecorator(path: string) {
  return (story: () => unknown) => ({
    components: { story },
    setup() {
      const router = useRouter();
      onMounted(() => {
        void router.replace(path);
      });
      return {};
    },
    template: "<story />",
  });
}

const meta = {
  title: "Components/Sidebar",
  component: Sidebar,
  tags: ["autodocs"],
  parameters: {
    layout: "padded",
  },
} satisfies Meta<typeof Sidebar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  decorators: [routePathDecorator("/")],
  render: () => ({
    components: { Sidebar },
    template: `
      <div style="height: 28rem; border: 1px solid var(--border); border-radius: var(--radius); overflow: hidden">
        <Sidebar />
      </div>
    `,
  }),
};

export const ActiveActivity: Story = {
  decorators: [routePathDecorator("/activity")],
  render: () => ({
    components: { Sidebar },
    template: `
      <div style="height: 28rem; border: 1px solid var(--border); border-radius: var(--radius); overflow: hidden">
        <Sidebar />
      </div>
    `,
  }),
};
