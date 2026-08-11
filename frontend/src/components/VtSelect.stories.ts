import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { ref } from "vue";
import VtSelect from "./VtSelect.vue";

const catalogParameters = {
  controls: { disable: true },
};

const meta = {
  title: "Components/VtSelect",
  component: VtSelect,
  tags: ["autodocs"],
  parameters: catalogParameters,
} satisfies Meta<typeof VtSelect>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => ({
    components: { VtSelect },
    setup() {
      const value = ref("ja");
      return { value };
    },
    template: `
      <VtSelect v-model="value" placeholder="Language">
        <el-option value="ja" label="日本語" />
        <el-option value="en" label="English" />
      </VtSelect>
    `,
  }),
};

export const Disabled: Story = {
  render: () => ({
    components: { VtSelect },
    setup() {
      const value = ref("ja");
      return { value };
    },
    template: `
      <VtSelect v-model="value" disabled placeholder="Language">
        <el-option value="ja" label="日本語" />
        <el-option value="en" label="English" />
      </VtSelect>
    `,
  }),
};
