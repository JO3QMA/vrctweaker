import type { Meta, StoryObj } from "@storybook/vue3-vite";
import VtButton from "./VtButton.vue";

const meta = {
  title: "Components/VtButton",
  component: VtButton,
  tags: ["autodocs"],
  args: {
    variant: "primary",
  },
  argTypes: {
    variant: {
      control: "select",
      options: ["primary", "secondary", "tertiary", "danger"],
    },
    plain: { control: "boolean" },
  },
} satisfies Meta<typeof VtButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Primary: Story = {
  args: { variant: "primary" },
  render: (args) => ({
    components: { VtButton },
    setup: () => ({ args }),
    template: `<VtButton v-bind="args">Primary</VtButton>`,
  }),
};

export const Secondary: Story = {
  args: { variant: "secondary" },
  render: (args) => ({
    components: { VtButton },
    setup: () => ({ args }),
    template: `<VtButton v-bind="args">Secondary</VtButton>`,
  }),
};

export const Tertiary: Story = {
  args: { variant: "tertiary" },
  render: (args) => ({
    components: { VtButton },
    setup: () => ({ args }),
    template: `<VtButton v-bind="args">Tertiary</VtButton>`,
  }),
};

export const Danger: Story = {
  args: { variant: "danger" },
  render: (args) => ({
    components: { VtButton },
    setup: () => ({ args }),
    template: `<VtButton v-bind="args">Danger</VtButton>`,
  }),
};

export const DangerPlain: Story = {
  args: { variant: "danger", plain: true },
  render: (args) => ({
    components: { VtButton },
    setup: () => ({ args }),
    template: `<VtButton v-bind="args">Danger plain</VtButton>`,
  }),
};

export const DisabledStates: Story = {
  args: { variant: "primary" },
  render: () => ({
    components: { VtButton },
    template: `
      <div style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center;">
        <VtButton variant="primary" disabled>Primary</VtButton>
        <VtButton variant="secondary" disabled>Secondary</VtButton>
        <VtButton variant="tertiary" disabled>Tertiary</VtButton>
        <VtButton variant="danger" disabled>Danger</VtButton>
      </div>
    `,
  }),
};

export const LoadingStates: Story = {
  args: { variant: "primary" },
  render: () => ({
    components: { VtButton },
    template: `
      <div style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center;">
        <VtButton variant="primary" loading>Primary</VtButton>
        <VtButton variant="secondary" loading>Secondary</VtButton>
        <VtButton variant="tertiary" loading>Tertiary</VtButton>
        <VtButton variant="danger" loading>Danger</VtButton>
      </div>
    `,
  }),
};

export const ActionGroupPrimarySecondary: Story = {
  name: "Action group: Primary + Secondary",
  args: { variant: "primary" },
  render: () => ({
    components: { VtButton },
    template: `
      <div style="display: flex; gap: 8px; justify-content: flex-end;">
        <VtButton variant="secondary">Launch</VtButton>
        <VtButton variant="primary">Save</VtButton>
      </div>
    `,
  }),
};

export const ActionGroupDangerConfirm: Story = {
  name: "Action group: Danger confirm + Secondary cancel",
  args: { variant: "danger" },
  render: () => ({
    components: { VtButton },
    template: `
      <div style="display: flex; gap: 8px; justify-content: flex-end;">
        <VtButton variant="secondary">Cancel</VtButton>
        <VtButton variant="danger">Delete</VtButton>
      </div>
    `,
  }),
};

export const ActionGroupPrimaryDangerPlain: Story = {
  name: "Action group: Primary + Danger plain",
  args: { variant: "primary" },
  render: () => ({
    components: { VtButton },
    template: `
      <div style="display: flex; gap: 8px; flex-wrap: wrap;">
        <VtButton variant="primary">Refresh profile</VtButton>
        <VtButton variant="danger" plain>Logout</VtButton>
      </div>
    `,
  }),
};
