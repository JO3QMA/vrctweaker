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
    template: `<VtSwitch v-model="enabled" />`,
  }),
};

export const Off: Story = {
  parameters: catalogParameters,
  render: () => ({
    components: { VtSwitch },
    setup() {
      const enabled = ref(false);
      return { enabled };
    },
    template: `<VtSwitch v-model="enabled" />`,
  }),
};

export const Disabled: Story = {
  parameters: catalogParameters,
  render: () => ({
    components: { VtSwitch },
    setup() {
      const enabled = ref(true);
      return { enabled };
    },
    template: `<VtSwitch v-model="enabled" disabled />`,
  }),
};

export const Small: Story = {
  name: "Small (list enable toggle)",
  parameters: catalogParameters,
  render: () => ({
    components: { VtSwitch },
    setup() {
      const enabled = ref(true);
      return { enabled };
    },
    template: `<VtSwitch v-model="enabled" size="small" />`,
  }),
};
