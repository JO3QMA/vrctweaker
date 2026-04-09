import { config } from "@vue/test-utils";
import ElementPlus from "element-plus";
import * as ElementPlusIconsVue from "@element-plus/icons-vue";
import { createAppI18n } from "../i18n";

// テスト環境で vue-i18n・Element Plus とアイコンをグローバル登録
config.global.plugins = [createAppI18n("ja"), ElementPlus];
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
