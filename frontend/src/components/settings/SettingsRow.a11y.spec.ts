import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import ElementPlus from "element-plus";
import { h } from "vue";
import { ElSwitch } from "element-plus";
import SettingsRow from "./SettingsRow.vue";
import SettingsGeneralPanel from "./SettingsGeneralPanel.vue";
import { createI18n } from "vue-i18n";
import ja from "../../i18n/locales/ja.json";

vi.mock("../../wails/app", () => ({
  App: {
    getCloseToTray: vi.fn().mockResolvedValue(false),
    setCloseToTray: vi.fn().mockResolvedValue(undefined),
    getSuppressSleepWhileVRChat: vi.fn().mockResolvedValue(false),
    setSuppressSleepWhileVRChat: vi.fn().mockResolvedValue(undefined),
  },
}));

describe("SettingsRow switch accessibility", () => {
  beforeEach(() => {
    document.body.innerHTML = "";
  });

  it("passes ariaLabel slot prop to el-switch input", async () => {
    const i18n = createI18n({
      legacy: false,
      locale: "ja",
      messages: { ja },
    });
    const wrapper = mount(SettingsRow, {
      props: {
        label: "トレイに閉じる",
        control: "switch",
      },
      slots: {
        default: ({ ariaLabel }: { ariaLabel: string }) =>
          h(ElSwitch, {
            "aria-label": ariaLabel,
            "data-testid": "sw",
          }),
      },
      global: {
        plugins: [ElementPlus, i18n],
      },
      attachTo: document.body,
    });
    await flushPromises();
    const input = wrapper.get('[data-testid="sw"] input[role="switch"]');
    expect(input.attributes("aria-label")).toBe("トレイに閉じる");
    wrapper.unmount();
  });

  it("wires settings panel switches to visible row labels", async () => {
    const i18n = createI18n({
      legacy: false,
      locale: "ja",
      messages: { ja },
    });
    const wrapper = mount(SettingsGeneralPanel, {
      global: { plugins: [ElementPlus, i18n] },
      attachTo: document.body,
    });
    await flushPromises();
    const closeSwitch = wrapper.get(
      '[data-testid="settings-close-to-tray"] input[role="switch"]',
    );
    expect(closeSwitch.attributes("aria-label")).toBe(ja.settings.closeToTray);
    wrapper.unmount();
  });
});
