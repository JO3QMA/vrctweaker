import { createI18n } from "vue-i18n";
import en from "../locales/en.json";
import ja from "../locales/ja.json";
import ko from "../locales/ko.json";
import zhCN from "../locales/zh-CN.json";
import zhTW from "../locales/zh-TW.json";

/** UI locale codes shown in settings (Japanese uses BCP47 ja-JP like zh-CN / zh-TW). */
export type AppLocale = "ja-JP" | "en" | "zh-CN" | "zh-TW" | "ko";

export const SUPPORTED_UI_LOCALES: AppLocale[] = [
  "ja-JP",
  "en",
  "zh-CN",
  "zh-TW",
  "ko",
];

export function isAppLocale(s: string): s is AppLocale {
  return (SUPPORTED_UI_LOCALES as string[]).includes(s);
}

/** Maps persisted / backend codes (Go stores `ja`) to vue-i18n locale keys. */
export function backendToAppLocale(code: string): AppLocale {
  const s = code.trim();
  if (s === "ja" || s === "ja-JP") {
    return "ja-JP";
  }
  if (isAppLocale(s)) {
    return s;
  }
  return "en";
}

export function createAppI18n(initialLocale: string) {
  const locale = backendToAppLocale(initialLocale);
  return createI18n({
    legacy: false,
    locale,
    fallbackLocale: "en",
    messages: {
      "ja-JP": ja,
      en,
      "zh-CN": zhCN,
      "zh-TW": zhTW,
      ko,
    },
    missing: (_loc, key) => key,
  });
}

export { elementPlusLocaleFor } from "./elementPlusLocale";
