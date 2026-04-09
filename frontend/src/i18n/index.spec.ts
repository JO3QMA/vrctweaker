import { describe, it, expect } from "vitest";
import { unref } from "vue";
import { createAppI18n } from "./index";

describe("createAppI18n", () => {
  it("maps backend ja to ja-JP for vue-i18n", () => {
    const i18n = createAppI18n("ja");
    expect(unref(i18n.global.locale)).toBe("ja-JP");
  });

  it("accepts ja-JP from backend", () => {
    const i18n = createAppI18n("ja-JP");
    expect(unref(i18n.global.locale)).toBe("ja-JP");
  });

  it("falls back to en for unsupported codes", () => {
    const i18n = createAppI18n("xx");
    expect(unref(i18n.global.locale)).toBe("en");
  });

  it("resolves meta.dashboard in Japanese", () => {
    const i18n = createAppI18n("ja");
    expect(i18n.global.t("meta.dashboard")).toBe("ダッシュボード");
  });
});
