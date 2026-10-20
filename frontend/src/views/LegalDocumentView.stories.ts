import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { createRouter, createWebHashHistory } from "vue-router";
import LegalDocumentView from "./LegalDocumentView.vue";

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: "/legal/privacy",
      component: LegalDocumentView,
      props: { kind: "privacy" },
    },
    {
      path: "/legal/terms",
      component: LegalDocumentView,
      props: { kind: "terms" },
    },
  ],
});

const meta: Meta<typeof LegalDocumentView> = {
  title: "Views/LegalDocumentView",
  component: LegalDocumentView,
  tags: ["autodocs"],
  decorators: [
    () => ({
      template: "<router-view />",
      setup() {
        return {};
      },
    }),
  ],
  parameters: {
    vueRouter: { router },
  },
};

export default meta;

type Story = StoryObj<typeof LegalDocumentView>;

export const Privacy: Story = {
  render: () => ({
    components: { LegalDocumentView },
    template: '<LegalDocumentView kind="privacy" />',
  }),
};

export const Terms: Story = {
  render: () => ({
    components: { LegalDocumentView },
    template: '<LegalDocumentView kind="terms" />',
  }),
};
