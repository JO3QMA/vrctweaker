import type { RouteRecordRaw } from "vue-router";

export const SETTINGS_SECTION_IDS = [
  "account",
  "general",
  "data",
  "about",
  "reset",
] as const;

export type SettingsSectionId = (typeof SETTINGS_SECTION_IDS)[number];

export function isSettingsSectionId(value: string): value is SettingsSectionId {
  return (SETTINGS_SECTION_IDS as readonly string[]).includes(value);
}

/** `/settings` と不明なパスはアカウント（サブナビ先頭）に倒す。 */
export function settingsSectionFromPath(path: string): SettingsSectionId {
  const parts = path.split("/").filter(Boolean);
  const candidate = parts[0] === "settings" ? (parts[1] ?? "") : "";
  if (isSettingsSectionId(candidate)) return candidate;
  return "account";
}

const settingsChildPlaceholder = { render: () => null };

export function settingsChildRoutes(): RouteRecordRaw[] {
  return [
    { path: "", redirect: { name: "settings-account" } },
    ...SETTINGS_SECTION_IDS.map((id) => ({
      path: id,
      name: `settings-${id}`,
      component: settingsChildPlaceholder,
    })),
  ];
}
