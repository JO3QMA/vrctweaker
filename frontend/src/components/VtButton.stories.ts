import type { Meta, StoryObj } from "@storybook/vue3-vite";
import VtButton from "./VtButton.vue";
import { VT_BUTTON_VARIANTS, type VtButtonProps } from "./vtButtonVariants";
import "./VtButton.stories.css";

const catalogParameters = {
  controls: { disable: true },
};

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
      options: [...VT_BUTTON_VARIANTS],
    },
    plain: { control: "boolean" },
  },
} satisfies Meta<typeof VtButton>;

export default meta;
type Story = StoryObj<typeof meta>;

function variantStory(label: string, args: VtButtonProps): Story {
  return {
    args,
    render: (storyArgs) => ({
      components: { VtButton },
      setup: () => ({ args: storyArgs }),
      template: `<VtButton v-bind="args">${label}</VtButton>`,
    }),
  };
}

export const Primary = variantStory("Primary", { variant: "primary" });
export const Secondary = variantStory("Secondary", { variant: "secondary" });
export const Tertiary = variantStory("Tertiary", { variant: "tertiary" });
export const Danger = variantStory("Danger", { variant: "danger" });
export const DangerPlain = variantStory("Danger plain", {
  variant: "danger",
  plain: true,
});

export const DisabledStates: Story = {
  parameters: catalogParameters,
  render: () => ({
    components: { VtButton },
    template: `
      <div class="vt-button-story-row">
        <VtButton variant="primary" disabled>Primary</VtButton>
        <VtButton variant="secondary" disabled>Secondary</VtButton>
        <VtButton variant="tertiary" disabled>Tertiary</VtButton>
        <VtButton variant="danger" disabled>Danger</VtButton>
      </div>
    `,
  }),
};

export const LoadingStates: Story = {
  parameters: catalogParameters,
  render: () => ({
    components: { VtButton },
    template: `
      <div class="vt-button-story-row">
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
  parameters: catalogParameters,
  render: () => ({
    components: { VtButton },
    template: `
      <div class="vt-button-story-actions">
        <VtButton variant="secondary">Launch</VtButton>
        <VtButton variant="primary">Save</VtButton>
      </div>
    `,
  }),
};

export const ActionGroupDangerConfirm: Story = {
  name: "Action group: Danger confirm + Secondary cancel",
  parameters: catalogParameters,
  render: () => ({
    components: { VtButton },
    template: `
      <div class="vt-button-story-actions">
        <VtButton variant="secondary">Cancel</VtButton>
        <VtButton variant="danger">Delete</VtButton>
      </div>
    `,
  }),
};

export const ActionGroupPrimaryDangerPlain: Story = {
  name: "Action group: Primary + Danger plain",
  parameters: catalogParameters,
  render: () => ({
    components: { VtButton },
    template: `
      <div class="vt-button-story-actions-wrap">
        <VtButton variant="primary">Refresh profile</VtButton>
        <VtButton variant="danger" plain>Logout</VtButton>
      </div>
    `,
  }),
};
