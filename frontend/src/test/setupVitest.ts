import { config } from "@vue/test-utils";
import ElementPlus from "element-plus";
import * as ElementPlusIconsVue from "@element-plus/icons-vue";
import { createI18n } from "vue-i18n";
import en from "../i18n/locales/en.json";
import ja from "../i18n/locales/ja.json";
import ko from "../i18n/locales/ko.json";
import zhTW from "../i18n/locales/zh-TW.json";
import zhCN from "../i18n/locales/zh-CN.json";

const testI18n = createI18n({
  legacy: false,
  locale: "ja",
  fallbackLocale: "en",
  messages: { en, ja, ko, "zh-TW": zhTW, "zh-CN": zhCN },
  globalInjection: true,
});

// テスト環境で Element Plus・vue-i18n・アイコンをグローバル登録
config.global.plugins = [ElementPlus, testI18n];
config.global.components = Object.fromEntries(
  Object.entries(ElementPlusIconsVue),
);

/**
 * jsdom does not implement ResizeObserver; TanStack Virtual and the gallery grid rely on it.
 */
class ResizeObserverMock implements ResizeObserver {
  constructor(private cb: ResizeObserverCallback) {}

  observe(target: Element): void {
    const w = target.clientWidth > 0 ? target.clientWidth : 480;
    const h = target.clientHeight > 0 ? target.clientHeight : 400;
    this.cb(
      [
        {
          target,
          contentRect: {
            x: 0,
            y: 0,
            width: w,
            height: h,
            top: 0,
            left: 0,
            bottom: h,
            right: w,
            toJSON() {
              return {};
            },
          },
          borderBoxSize: [],
          contentBoxSize: [],
          devicePixelContentBoxSize: [],
        } as ResizeObserverEntry,
      ],
      this,
    );
  }

  unobserve(): void {}

  disconnect(): void {}
}

globalThis.ResizeObserver = ResizeObserverMock;
