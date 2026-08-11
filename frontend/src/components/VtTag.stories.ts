import type { Meta, StoryObj } from "@storybook/vue3-vite";
import VtTag from "./VtTag.vue";
import { VT_TAG_VARIANTS, type VtTagProps } from "./vtTagVariants";
import "./VtTag.stories.css";

const catalogParameters = {
  controls: { disable: true },
};

const meta = {
  title: "Components/VtTag",
  component: VtTag,
  tags: ["autodocs"],
  args: {
    variant: "neutral",
  },
  argTypes: {
    variant: {
      control: "select",
      options: [...VT_TAG_VARIANTS],
    },
    size: {
      control: "select",
      options: ["large", "default", "small"],
    },
  },
} satisfies Meta<typeof VtTag>;

export default meta;
type Story = StoryObj<typeof meta>;

function variantStory(label: string, args: VtTagProps): Story {
  return {
    args,
    render: (storyArgs) => ({
      components: { VtTag },
      setup: () => ({ args: storyArgs, label }),
      template: `<VtTag v-bind="args">{{ label }}</VtTag>`,
    }),
  };
}

export const Success = variantStory("Success", { variant: "success" });
export const Warning = variantStory("Warning", { variant: "warning" });
export const Danger = variantStory("Danger", { variant: "danger" });
export const Info = variantStory("Info", { variant: "info" });
export const Neutral = variantStory("3 licenses", { variant: "neutral" });
export const Primary = variantStory("Default", { variant: "primary" });

export const SemanticRow: Story = {
  name: "Semantic status badges",
  parameters: catalogParameters,
  render: () => ({
    components: { VtTag },
    template: `
      <div class="vt-tag-story-row">
        <VtTag variant="success" size="small">Logged in</VtTag>
        <VtTag variant="danger" size="small">Failed</VtTag>
        <VtTag variant="warning" size="small">Needs review</VtTag>
        <VtTag variant="info" size="small">Rule</VtTag>
      </div>
    `,
  }),
};

export const LaunchProfileDefaultLabel: Story = {
  name: "Example: launch profile default label",
  parameters: catalogParameters,
  render: () => ({
    components: { VtTag },
    template: `
      <VtTag variant="primary" size="small">Default</VtTag>
    `,
  }),
};
