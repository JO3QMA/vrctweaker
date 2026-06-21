import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { ElMessage } from "element-plus";
import { createRouter, createWebHashHistory } from "vue-router";
import SettingsView from "../SettingsView.vue";
import { App } from "../../wails/app";
import * as I18nModule from "../../i18n";

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: "/settings", component: SettingsView },
    { path: "/licenses", component: { template: "<div>Licenses</div>" } },
  ],
});

describe("SettingsView", () => {
  beforeEach(() => {
    vi.spyOn(App, "getMicMuteSyncSettings").mockResolvedValue({
      enabled: false,
      oscEndpoint: "",
    });
    vi.spyOn(App, "getMicMuteSyncStatus").mockResolvedValue({
      available: true,
      enabled: false,
      oscEndpoint: "9000:127.0.0.1:9001",
      vrchatOscListening: true,
      vrchatOscConnected: false,
      vrchatMuteKnown: false,
      vrchatMuted: false,
      syncEngineState: "off",
      discordRpcConnected: false,
      discordMuteKnown: false,
      discordMuted: false,
      toggleVoiceKnown: false,
      toggleVoiceOk: false,
    });
    vi.spyOn(App, "saveMicMuteSyncSettings").mockResolvedValue(undefined);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("renders settings title", async () => {
    await router.push("/settings");
    await router.isReady();
    const wrapper = mount(SettingsView, {
      global: {
        plugins: [router],
      },
    });
    expect(wrapper.find(".page-title").text()).toBe("設定");
  });

  it("has link to OSS licenses page", async () => {
    await router.push("/settings");
    await router.isReady();
    const wrapper = mount(SettingsView, {
      global: {
        plugins: [router],
      },
    });
    const licensesLink = wrapper.find(".btn-licenses");
    expect(licensesLink.exists()).toBe(true);
    expect(licensesLink.attributes("href")).toContain("/licenses");
    expect(licensesLink.text()).toContain("OSS ライセンス一覧");
  });

  it("shows Mic Mute Sync section when available", async () => {
    await router.push("/settings");
    await router.isReady();
    const wrapper = mount(SettingsView, {
      global: {
        plugins: [router],
      },
    });
    await flushPromises();
    expect(wrapper.find('[data-testid="mic-mute-sync-card"]').exists()).toBe(
      true,
    );
  });
});

describe("SettingsView onLanguageChange", () => {
  beforeEach(async () => {
    await router.push("/settings");
    await router.isReady();
    vi.spyOn(App, "setLanguage").mockResolvedValue(undefined);
    vi.spyOn(I18nModule, "setLanguage");
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("calls i18n setLanguage after App.setLanguage succeeds", async () => {
    const wrapper = mount(SettingsView, {
      global: {
        plugins: [router],
      },
    });
    await flushPromises();

    const select = wrapper.findComponent({ name: "ElSelect" });
    await select.vm.$emit("update:modelValue", "en");
    await flushPromises();

    expect(App.setLanguage).toHaveBeenCalledWith("en");
    expect(I18nModule.setLanguage).toHaveBeenCalledWith("en");
  });

  it("does not call i18n setLanguage when App.setLanguage fails", async () => {
    vi.mocked(App.setLanguage).mockRejectedValueOnce(new Error("save failed"));
    const elErrorSpy = vi.spyOn(ElMessage, "error").mockImplementation(() => ({
      close: () => {},
    }));

    const wrapper = mount(SettingsView, {
      global: {
        plugins: [router],
      },
    });
    await flushPromises();

    vi.mocked(I18nModule.setLanguage).mockClear();

    const select = wrapper.findComponent({ name: "ElSelect" });
    await select.vm.$emit("update:modelValue", "en");
    await flushPromises();

    expect(I18nModule.setLanguage).not.toHaveBeenCalled();
    expect(elErrorSpy).toHaveBeenCalledWith("save failed");
  });
});
