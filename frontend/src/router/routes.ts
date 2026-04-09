import type { RouteRecordRaw } from "vue-router";
import ActivityView from "../views/ActivityView.vue";
import AutomationView from "../views/AutomationView.vue";
import ConfigView from "../views/ConfigView.vue";
import DashboardView from "../views/DashboardView.vue";
import EncounterHistoryDetailView from "../views/EncounterHistoryDetailView.vue";
import FriendsView from "../views/FriendsView.vue";
import GalleryView from "../views/GalleryView.vue";
import LauncherView from "../views/LauncherView.vue";
import LicensesView from "../views/LicensesView.vue";
import SettingsView from "../views/SettingsView.vue";
import UserProfileDetailView from "../views/UserProfileDetailView.vue";

/**
 * ルートはすべて同期コンポーネント。
 * Wails 本番（埋め込み dist・相対 base）では `() => import()` のチャンク URL が
 * WebView で解決失敗し router-view が空になることがあるため、コード分割しない。
 */
export const appRoutes: RouteRecordRaw[] = [
  {
    path: "/",
    name: "dashboard",
    component: DashboardView,
    meta: { titleKey: "meta.dashboard" },
  },
  {
    path: "/launcher",
    name: "launcher",
    component: LauncherView,
    meta: { titleKey: "meta.launcher" },
  },
  {
    path: "/gallery",
    name: "gallery",
    component: GalleryView,
    meta: { titleKey: "meta.gallery" },
  },
  {
    path: "/activity",
    name: "activity",
    component: ActivityView,
    meta: { titleKey: "meta.activity" },
  },
  {
    path: "/activity/encounter-history",
    name: "encounter-history",
    component: EncounterHistoryDetailView,
    meta: { titleKey: "meta.encounterHistory", bare: true },
  },
  {
    path: "/friends",
    name: "friends",
    component: FriendsView,
    meta: { titleKey: "meta.friends" },
  },
  {
    path: "/user-profile",
    name: "user-profile",
    component: UserProfileDetailView,
    meta: { titleKey: "meta.userProfile" },
  },
  {
    path: "/automation",
    name: "automation",
    component: AutomationView,
    meta: { titleKey: "meta.automation" },
  },
  {
    path: "/config",
    name: "config",
    component: ConfigView,
    meta: { titleKey: "meta.config" },
  },
  {
    path: "/settings",
    name: "settings",
    component: SettingsView,
    meta: { titleKey: "meta.settings" },
  },
  {
    path: "/licenses",
    name: "licenses",
    component: LicensesView,
    meta: { titleKey: "meta.licenses" },
  },
];
