import type { Preview } from "@storybook/vue3-vite";
import { setup } from "@storybook/vue3-vite";
import { createMemoryHistory, createRouter } from "vue-router";
import ElementPlus from "element-plus";
import "element-plus/dist/index.css";
import "element-plus/theme-chalk/dark/css-vars.css";
import * as ElementPlusIconsVue from "@element-plus/icons-vue";
import "../src/assets/style.css";

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: "/", name: "dashboard", component: { template: "<div />" } },
    { path: "/launcher", name: "launcher", component: { template: "<div />" } },
    { path: "/gallery", name: "gallery", component: { template: "<div />" } },
    { path: "/activity", name: "activity", component: { template: "<div />" } },
    {
      path: "/activity/encounter-history",
      name: "encounter-history",
      component: { template: "<div />" },
    },
    { path: "/friends", name: "friends", component: { template: "<div />" } },
    {
      path: "/automation",
      name: "automation",
      component: { template: "<div />" },
    },
    { path: "/config", name: "config", component: { template: "<div />" } },
    { path: "/settings", name: "settings", component: { template: "<div />" } },
    { path: "/licenses", name: "licenses", component: { template: "<div />" } },
  ],
});

setup((app) => {
  app.use(router);
  app.use(ElementPlus);
  for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component);
  }
});

const preview: Preview = {
  parameters: {
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
  },
};

export default preview;
