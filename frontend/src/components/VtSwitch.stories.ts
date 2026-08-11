import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { ref } from "vue";
import VtSwitch from "./VtSwitch.vue";

const catalogParameters = {
  controls: { disable: true },
};

const meta = {
  title: "Components/VtSwitch",
  component: VtSwitch,
  tags: ["autodocs"],
  parameters: catalogParameters,
} satisfies Meta<typeof VtSwitch>;

export default meta;
type Story = StoryObj<typeof meta>;

export const On: Story = {
  render: () => ({
    components: { VtSwitch },
    setup() {
      const enabled = ref(true);
      return { enabled };
    },
    template: `<VtSwitch v-model="enabled" aria-label="Enable feature" />`,
  }),
};

export const Off: Story = {
  render: () => ({
    components: { VtSwitch },
    setup() {
      const enabled = ref(false);
      return { enabled };
    },
    template: `<VtSwitch v-model="enabled" aria-label="Enable feature" />`,
  }),
};

export const Disabled: Story = {
  render: () => ({
    components: { VtSwitch },
    setup() {
      const enabled = ref(true);
      return { enabled };
    },
    template: `<VtSwitch v-model="enabled" disabled aria-label="Enable feature" />`,
  }),
};

export const Small: Story = {
  name: "Small (list enable toggle)",
  render: () => ({
    components: { VtSwitch },
    setup() {
      const enabled = ref(true);
      return { enabled };
    },
    template: `<VtSwitch v-model="enabled" size="small" aria-label="Enable list item" />`,
  }),
};
