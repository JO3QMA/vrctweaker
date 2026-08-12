import type { Component } from "vue";
import {
  Cpu,
  DataLine,
  EditPen,
  House,
  Picture,
  Promotion,
  Setting,
  User,
  UserFilled,
  VideoCamera,
} from "@element-plus/icons-vue";

export type NavMenuItem = {
  path: string;
  labelKey: string;
  icon: Component;
  windowsOnly?: boolean;
};

/** Main sidebar navigation entries (footer settings excluded). */
export const NAV_MAIN_MENU_ITEMS: readonly NavMenuItem[] = [
  { path: "/", labelKey: "nav.dashboard", icon: House },
  { path: "/launcher", labelKey: "nav.launcher", icon: Promotion },
  { path: "/gallery", labelKey: "nav.gallery", icon: Picture },
  { path: "/activity", labelKey: "nav.activity", icon: DataLine },
  { path: "/me", labelKey: "nav.me", icon: User },
  { path: "/friends", labelKey: "nav.friends", icon: UserFilled },
  { path: "/automation", labelKey: "nav.automation", icon: Cpu },
  {
    path: "/video",
    labelKey: "nav.video",
    icon: VideoCamera,
    windowsOnly: true,
  },
  { path: "/config", labelKey: "nav.configOther", icon: EditPen },
];

export const NAV_SETTINGS_MENU_ITEM: NavMenuItem = {
  path: "/settings",
  labelKey: "nav.settings",
  icon: Setting,
};
