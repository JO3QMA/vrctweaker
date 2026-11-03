import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { createI18n } from "vue-i18n";
import AppStartupScreen from "../AppStartupScreen.vue";
import en from "../../i18n/locales/en.json";

import type { StartupInitStatus } from "../../composables/useStartupInit";

function mountScreen(status: StartupInitStatus) {
  const i18n = createI18n({
    legacy: false,
    locale: "en",
    messages: { en },
  });
  return mount(AppStartupScreen, {
    props: { status, retrying: false },
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
});
