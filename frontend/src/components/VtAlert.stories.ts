import type { Meta, StoryObj } from "@storybook/vue3-vite";
import VtAlert from "./VtAlert.vue";
import { VT_ALERT_VARIANTS, type VtAlertProps } from "./vtAlertVariants";
import "./VtAlert.stories.css";

const catalogParameters = {
  controls: { disable: true },
};

const meta = {
  title: "Components/VtAlert",
  component: VtAlert,
  tags: ["autodocs"],
  args: {
    variant: "info",
  },
  argTypes: {
    variant: {
      control: "select",
      options: [...VT_ALERT_VARIANTS],
    },
  },
} satisfies Meta<typeof VtAlert>;

export default meta;
type Story = StoryObj<typeof meta>;

function variantStory(alertTitle: string, args: VtAlertProps): Story {
  return {
    args: { ...args, title: alertTitle },
    render: (storyArgs) => ({
      components: { VtAlert },
      setup: () => ({ args: storyArgs }),
      template: `<VtAlert v-bind="args" />`,
    }),
  };
}

export const Success = variantStory("Operation completed", {
  variant: "success",
});
export const Warning = variantStory("Review before continuing", {
  variant: "warning",
});
export const Danger = variantStory("Could not load data", {
  variant: "danger",
});
export const Info = variantStory("Official build required for cookies", {
  variant: "info",
});

export const WithDescription: Story = {
  args: {
    variant: "warning",
    title: "Account risk",
    description:
      "Using cookies with a main account may risk a ban. Prefer a throwaway account.",
  },
  render: (storyArgs) => ({
    components: { VtAlert },
    setup: () => ({ args: storyArgs }),
    template: `<VtAlert v-bind="args" />`,
  }),
};

export const BlockErrorExample: Story = {
  name: "Example: block fetch failure",
  parameters: catalogParameters,
  render: () => ({
    components: { VtAlert },
    template: `
      <div class="vt-alert-story-stack">
        <VtAlert variant="danger" title="Could not load gallery" />
      </div>
    `,
  }),
};

export const PersistentWarningExample: Story = {
  name: "Example: persistent section warning",
  parameters: catalogParameters,
  render: () => ({
    components: { VtAlert },
    template: `
      <div class="vt-alert-story-stack">
        <VtAlert variant="warning" title="Replacing the bundled yt-dlp is not officially supported." />
      </div>
    `,
  }),
};
