import type { Meta, StoryObj } from "@storybook/vue3-vite";
import LegalDocLinks from "./LegalDocLinks.vue";

const meta: Meta<typeof LegalDocLinks> = {
  title: "Components/LegalDocLinks",
  component: LegalDocLinks,
  tags: ["autodocs"],
};

export default meta;

type Story = StoryObj<typeof LegalDocLinks>;

export const StackedInApp: Story = {
  args: {
    intro: "データの取り扱いと利用条件は次のドキュメントをご確認ください。",
    stacked: true,
    showOssLicensesLink: true,
  },
};

export const DialogPresentation: Story = {
  args: {
    intro: "ログインすると、次のドキュメントに同意したものとみなされます。",
    stacked: true,
    presentation: "dialog",
    showGitHubSupplement: false,
  },
};
