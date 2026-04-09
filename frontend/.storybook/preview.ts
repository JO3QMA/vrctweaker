import type { Preview } from "@storybook/vue3-vite";
import { setup } from "@storybook/vue3-vite";
import { createMemoryHistory, createRouter } from "vue-router";
import { createI18n } from "vue-i18n";
import ElementPlus from "element-plus";
import "element-plus/dist/index.css";
import "element-plus/theme-chalk/dark/css-vars.css";
import * as ElementPlusIconsVue from "@element-plus/icons-vue";
import "../src/assets/style.css";
import en from "../src/i18n/locales/en.json";
import ja from "../src/i18n/locales/ja.json";
import ko from "../src/i18n/locales/ko.json";
import zhTW from "../src/i18n/locales/zh-TW.json";
import zhCN from "../src/i18n/locales/zh-CN.json";

const storybookI18n = createI18n({
  legacy: false,
  locale: "ja",
  fallbackLocale: "en",
  messages: { en, ja, ko, "zh-TW": zhTW, "zh-CN": zhCN },
  globalInjection: true,
});

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
      path: "/user-profile",
      name: "user-profile",
      component: { template: "<div />" },
    },
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
  app.use(storybookI18n);
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
