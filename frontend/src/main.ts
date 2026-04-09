import { createApp } from "vue";
import { createRouter, createWebHashHistory } from "vue-router";
import ElementPlus from "element-plus";
import "element-plus/dist/index.css";
import "element-plus/theme-chalk/dark/css-vars.css";
import * as ElementPlusIconsVue from "@element-plus/icons-vue";
import AppRoot from "./App.vue";
import { App } from "./wails/app";
import { createAppI18n } from "./i18n";
import "./assets/style.css";
import type { RouteRecordRaw } from "vue-router";

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    name: "dashboard",
    component: () => import("./views/DashboardView.vue"),
    meta: { titleKey: "meta.dashboard" },
  },
  {
    path: "/launcher",
    name: "launcher",
    component: () => import("./views/LauncherView.vue"),
    meta: { titleKey: "meta.launcher" },
  },
  {
    path: "/gallery",
    name: "gallery",
    component: () => import("./views/GalleryView.vue"),
    meta: { titleKey: "meta.gallery" },
  },
  {
    path: "/activity",
    name: "activity",
    component: () => import("./views/ActivityView.vue"),
    meta: { titleKey: "meta.activity" },
  },
  {
    path: "/activity/encounter-history",
    name: "encounter-history",
    component: () => import("./views/EncounterHistoryDetailView.vue"),
    meta: { titleKey: "meta.encounterHistory", bare: true },
  },
  {
    path: "/friends",
    name: "friends",
    component: () => import("./views/FriendsView.vue"),
    meta: { titleKey: "meta.friends" },
  },
  {
    path: "/user-profile",
    name: "user-profile",
    component: () => import("./views/UserProfileDetailView.vue"),
    meta: { titleKey: "meta.userProfile" },
  },
  {
    path: "/automation",
    name: "automation",
    component: () => import("./views/AutomationView.vue"),
    meta: { titleKey: "meta.automation" },
  },
  {
    path: "/config",
    name: "config",
    component: () => import("./views/ConfigView.vue"),
    meta: { titleKey: "meta.config" },
  },
  {
    path: "/settings",
    name: "settings",
    component: () => import("./views/SettingsView.vue"),
    meta: { titleKey: "meta.settings" },
  },
  {
    path: "/licenses",
    name: "licenses",
    component: () => import("./views/LicensesView.vue"),
    meta: { titleKey: "meta.licenses" },
  },
];

async function bootstrap() {
  let code = "ja";
  try {
    code = await App.getUILanguage();
  } catch {
    // DB or IPC error – fall back to Japanese so the app still renders.
  }
  const i18n = createAppI18n(code);
  const router = createRouter({
    history: createWebHashHistory(),
    routes,
  });

  router.afterEach((to) => {
    const titleKey = to.meta.titleKey;
    if (typeof titleKey === "string" && titleKey.length > 0) {
      document.title = `${i18n.global.t(titleKey)} - ${i18n.global.t("appTitle")}`;
    }
  });

  const app = createApp(AppRoot);
  app.use(i18n);
  app.use(router);
  app.use(ElementPlus);
  for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component);
  }
  app.mount("#app");
}

void bootstrap();
