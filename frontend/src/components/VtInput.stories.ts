import type { Meta, StoryObj } from "@storybook/vue3-vite";
import VtInput from "./VtInput.vue";
import "./FormControl.stories.css";

const catalogParameters = {
  controls: { disable: true },
};

const meta = {
  title: "Components/VtInput",
  component: VtInput,
  tags: ["autodocs"],
  parameters: catalogParameters,
} satisfies Meta<typeof VtInput>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => ({
    components: { VtInput },
    template: `<VtInput placeholder="Profile name" />`,
  }),
};

export const Disabled: Story = {
  render: () => ({
    components: { VtInput },
    template: `<VtInput placeholder="Profile name" disabled />`,
  }),
};

export const WithPrefix: Story = {
  render: () => ({
    components: { VtInput },
    template: `
      <VtInput placeholder="Search display name" clearable>
        <template #prefix><span class="vt-form-story-prefix">⌕</span></template>
      </VtInput>
    `,
  }),
};
