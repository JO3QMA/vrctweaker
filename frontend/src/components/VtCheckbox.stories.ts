import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { ref } from "vue";
import VtCheckbox from "./VtCheckbox.vue";

const catalogParameters = {
  controls: { disable: true },
};

const meta = {
  title: "Components/VtCheckbox",
  component: VtCheckbox,
  tags: ["autodocs"],
} satisfies Meta<typeof VtCheckbox>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => ({
    components: { VtCheckbox },
    setup() {
      const checked = ref(true);
      return { checked };
    },
    template: `<VtCheckbox v-model="checked">Desktop mode</VtCheckbox>`,
  }),
};

export const Disabled: Story = {
  parameters: catalogParameters,
  render: () => ({
    components: { VtCheckbox },
    setup() {
      const checked = ref(false);
      return { checked };
    },
    template: `<VtCheckbox v-model="checked" disabled>Desktop mode</VtCheckbox>`,
  }),
};
