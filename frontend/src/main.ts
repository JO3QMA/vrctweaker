import { createApp } from "vue";
import { createRouter, createWebHashHistory } from "vue-router";
import ElementPlus from "element-plus";
import "element-plus/dist/index.css";
import "element-plus/theme-chalk/dark/css-vars.css";
import * as ElementPlusIconsVue from "@element-plus/icons-vue";
import App from "./App.vue";
import "./assets/style.css";
import type { RouteRecordRaw } from "vue-router";
import { i18n } from "./i18n";

declare module "vue-router" {
  interface RouteMeta {
    /** vue-i18n key under `routes.*` */
    titleKey?: string;
    bare?: boolean;
  }
}

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    name: "dashboard",
    component: () => import("./views/DashboardView.vue"),
    meta: { titleKey: "routes.dashboard" },
  },
  {
    path: "/launcher",
    name: "launcher",
    component: () => import("./views/LauncherView.vue"),
    meta: { titleKey: "routes.launcher" },
  },
  {
    path: "/gallery",
    name: "gallery",
    component: () => import("./views/GalleryView.vue"),
    meta: { titleKey: "routes.gallery" },
  },
  {
    path: "/activity",
    name: "activity",
    component: () => import("./views/ActivityView.vue"),
    meta: { titleKey: "routes.activity" },
  },
  {
    path: "/activity/encounter-history",
    name: "encounter-history",
    component: () => import("./views/EncounterHistoryDetailView.vue"),
    meta: { titleKey: "routes.encounterHistory", bare: true },
  },
  {
    path: "/friends",
    name: "friends",
    component: () => import("./views/FriendsView.vue"),
    meta: { titleKey: "routes.friends" },
  },
  {
    path: "/user-profile",
    name: "user-profile",
    component: () => import("./views/UserProfileDetailView.vue"),
    meta: { titleKey: "routes.user" },
  },
  {
    path: "/automation",
    name: "automation",
    component: () => import("./views/AutomationView.vue"),
    meta: { titleKey: "routes.automation" },
  },
  {
    path: "/config",
    name: "config",
    component: () => import("./views/ConfigView.vue"),
    meta: { titleKey: "routes.configOther" },
  },
  {
    path: "/settings",
    name: "settings",
    component: () => import("./views/SettingsView.vue"),
    meta: { titleKey: "routes.settings" },
  },
  {
    path: "/licenses",
    name: "licenses",
    component: () => import("./views/LicensesView.vue"),
    meta: { titleKey: "routes.licenses" },
  },
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

const app = createApp(App);
app.use(router);
app.use(i18n);
app.use(ElementPlus);
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component);
}
app.mount("#app");
