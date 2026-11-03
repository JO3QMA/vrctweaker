import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { nextTick } from "vue";
import { createI18n } from "vue-i18n";
import AppStartupScreen from "../AppStartupScreen.vue";
import en from "../../i18n/locales/en.json";

import type { StartupInitStatus } from "../../composables/useStartupInit";

function mountScreen(
  status: StartupInitStatus,
  retrying = false,
  attachTo?: HTMLElement,
) {
  const i18n = createI18n({
    legacy: false,
    locale: "en",
    messages: { en },
  });
  return mount(AppStartupScreen, {
    props: { status, retrying },
    attachTo: attachTo ?? document.body,
    global: { plugins: [i18n] },
  });
}

describe("AppStartupScreen", () => {
  it("shows migrating message when dedupe phase is active", () => {
    const wrapper = mountScreen({
      ready: false,
      loading: true,
      phase: "dedupe",
      summaryKey: "",
      detail: "",
      dataDir: "",
      canOpenDataFolder: false,
    });
    expect(wrapper.get('[data-testid="startup-loading"]').text()).toContain(
      "Organizing data",
    );
  });

  it("shows failure summary and hides open-folder when unavailable", () => {
    const wrapper = mountScreen({
      ready: false,
      loading: false,
      phase: "",
      summaryKey: "app.startup.failed.dataDirResolve",
      detail: "permission denied",
      dataDir: "",
      canOpenDataFolder: false,
    });
    expect(wrapper.get('[data-testid="startup-failure"]').text()).toContain(
      "Could not resolve the data folder",
    );
    expect(
      wrapper.find('[data-testid="startup-open-data-folder"]').exists(),
    ).toBe(false);
  });

  it("hides details collapse when detail is empty", () => {
    const wrapper = mountScreen({
      ready: false,
      loading: false,
      phase: "",
      summaryKey: "app.startup.failed.dataDirResolve",
      detail: "",
      dataDir: "",
      canOpenDataFolder: false,
    });
    expect(wrapper.find(".startup-detail").exists()).toBe(false);
  });

  it("shows status fetch detail in collapse", () => {
    const wrapper = mountScreen({
      ready: false,
      loading: false,
      phase: "",
      summaryKey: "app.startup.failed.statusFetch",
      detail: "offline",
      dataDir: "",
      canOpenDataFolder: false,
    });
    expect(wrapper.get('[data-testid="startup-failure-detail"]').text()).toBe(
      "offline",
    );
  });

  it("restores focus to retry after retrying when retry had focus", async () => {
    const wrapper = mountScreen(
      {
        ready: false,
        loading: false,
        phase: "",
        summaryKey: "app.startup.failed.databaseOpen",
        detail: "db error",
        dataDir: "",
        canOpenDataFolder: false,
      },
      false,
    );
    const btn = wrapper.get('[data-testid="startup-retry"]')
      .element as HTMLElement;
    btn.focus();
    await wrapper.get('[data-testid="startup-retry"]').trigger("click");
    await wrapper.setProps({ retrying: true });
    await wrapper.setProps({ retrying: false });
    await nextTick();
    expect(document.activeElement).toBe(btn);
    wrapper.unmount();
  });
});
