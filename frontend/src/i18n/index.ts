import { createI18n } from "vue-i18n";
import { shallowRef, type ShallowRef } from "vue";
import type { Language } from "element-plus/es/locale";
import enEl from "element-plus/es/locale/lang/en";
import jaEl from "element-plus/es/locale/lang/ja";
import koEl from "element-plus/es/locale/lang/ko";
import zhCnEl from "element-plus/es/locale/lang/zh-cn";
import zhTwEl from "element-plus/es/locale/lang/zh-tw";

import en from "./locales/en.json";
import ja from "./locales/ja.json";
import ko from "./locales/ko.json";
import zhTW from "./locales/zh-TW.json";
import zhCN from "./locales/zh-CN.json";

export type AppLocale = "en" | "ja" | "ko" | "zh-TW" | "zh-CN";

export const APP_LOCALES: AppLocale[] = ["ja", "en", "ko", "zh-TW", "zh-CN"];

const messages = {
  en,
  ja,
  ko,
  "zh-TW": zhTW,
  "zh-CN": zhCN,
} as const;

export const i18n = createI18n({
  legacy: false,
  locale: "ja",
  fallbackLocale: "en",
  messages: messages as Record<AppLocale, typeof en>,
  globalInjection: true,
});

const elLocaleByApp: Record<AppLocale, Language> = {
  en: enEl,
  ja: jaEl,
  ko: koEl,
  "zh-CN": zhCnEl,
  "zh-TW": zhTwEl,
};

/** Element Plus `el-config-provider` locale (synced by setLanguage). */
export const elLocale: ShallowRef<Language> = shallowRef(jaEl);

/** Maps persisted app locale to BCP 47 for `Intl` / `toLocaleString`. */
export function appLocaleToBcp47(locale: string): string {
  switch (locale) {
    case "ja":
      return "ja-JP";
    case "ko":
      return "ko-KR";
    case "zh-TW":
      return "zh-TW";
    case "zh-CN":
      return "zh-CN";
    default:
      return "en-US";
  }
}

export function isAppLocale(s: string): s is AppLocale {
  return (
    s === "en" || s === "ja" || s === "ko" || s === "zh-TW" || s === "zh-CN"
  );
}

export function setLanguage(lang: AppLocale): void {
  i18n.global.locale.value = lang;
  elLocale.value = elLocaleByApp[lang] ?? enEl;
}
