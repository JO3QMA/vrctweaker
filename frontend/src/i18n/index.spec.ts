import { describe, expect, it, beforeEach } from "vitest";
import enEl from "element-plus/es/locale/lang/en";
import jaEl from "element-plus/es/locale/lang/ja";
import koEl from "element-plus/es/locale/lang/ko";
import zhCnEl from "element-plus/es/locale/lang/zh-cn";
import zhTwEl from "element-plus/es/locale/lang/zh-tw";
import {
  APP_LOCALES,
  appLocaleToBcp47,
  elLocale,
  i18n,
  isAppLocale,
  setLanguage,
  type AppLocale,
} from "./index";

describe("appLocaleToBcp47", () => {
  it.each([
    ["ja", "ja-JP"],
    ["ko", "ko-KR"],
    ["zh-TW", "zh-TW"],
    ["zh-CN", "zh-CN"],
    ["en", "en-US"],
    ["fr", "en-US"],
    ["", "en-US"],
  ] as const)("maps %s to %s", (locale, expected) => {
    expect(appLocaleToBcp47(locale)).toBe(expected);
  });
});

describe("isAppLocale", () => {
  it("accepts supported app locales", () => {
    for (const locale of APP_LOCALES) {
      expect(isAppLocale(locale)).toBe(true);
    }
  });

  it("rejects unknown locale strings", () => {
    expect(isAppLocale("de")).toBe(false);
    expect(isAppLocale("zh")).toBe(false);
    expect(isAppLocale("")).toBe(false);
  });
});

describe("setLanguage", () => {
  beforeEach(() => {
    setLanguage("ja");
  });

  it.each([
    ["ja", jaEl],
    ["en", enEl],
    ["ko", koEl],
    ["zh-CN", zhCnEl],
    ["zh-TW", zhTwEl],
  ] as const)("syncs vue-i18n and Element Plus locale for %s", (lang, el) => {
    setLanguage(lang);
    expect(i18n.global.locale.value).toBe(lang);
    expect(elLocale.value).toBe(el);
  });

  it("falls back Element Plus locale to English for unknown keys", () => {
    setLanguage("en" as AppLocale);
    setLanguage("ja");

    setLanguage("xx" as AppLocale);
    expect(i18n.global.locale.value).toBe("xx");
    expect(elLocale.value).toBe(enEl);
  });
});
