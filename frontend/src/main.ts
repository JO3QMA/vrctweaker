import { createApp } from "vue";
import { createRouter, createWebHashHistory } from "vue-router";
import App from "./App.vue";
import "./assets/style.css";
import type { RouteRecordRaw } from "vue-router";

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    name: "dashboard",
    component: () => import("./views/DashboardView.vue"),
    meta: { title: "ダッシュボード" },
  },
  {
    path: "/launcher",
    name: "launcher",
    component: () => import("./views/LauncherView.vue"),
    meta: { title: "ランチャー" },
  },
  {
    path: "/gallery",
    name: "gallery",
    component: () => import("./views/GalleryView.vue"),
    meta: { title: "ギャラリー" },
  },
  {
    path: "/activity",
    name: "activity",
    component: () => import("./views/ActivityView.vue"),
    meta: { title: "アクティビティ" },
  },
  {
    path: "/activity/encounter-history",
    name: "encounter-history",
    component: () => import("./views/EncounterHistoryDetailView.vue"),
    meta: { title: "遭遇履歴", bare: true },
  },
  {
    path: "/friends",
    name: "friends",
    component: () => import("./views/FriendsView.vue"),
    meta: { title: "フレンド" },
  },
  {
    path: "/automation",
    name: "automation",
    component: () => import("./views/AutomationView.vue"),
    meta: { title: "オートメーション" },
  },
  {
    path: "/config",
    name: "config",
    component: () => import("./views/ConfigView.vue"),
    meta: { title: "その他の設定" },
  },
  {
    path: "/settings",
    name: "settings",
    component: () => import("./views/SettingsView.vue"),
    meta: { title: "設定" },
  },
  {
    path: "/licenses",
    name: "licenses",
    component: () => import("./views/LicensesView.vue"),
    meta: { title: "OSSライセンス" },
  },
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

const app = createApp(App);
app.use(router);
app.mount("#app");
