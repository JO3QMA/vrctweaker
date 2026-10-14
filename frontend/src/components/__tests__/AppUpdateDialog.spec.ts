import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import AppUpdateDialog from "../AppUpdateDialog.vue";
import { App } from "../../wails/app";

vi.mock("../../composables/useAppUpdate", async () => {
  const { ref } = await import("vue");
  const dialogVisible = ref(true);
  const status = ref({
    currentVersion: "1.0.0",
    updateAvailable: true,
    shouldPrompt: true,
    latestVersion: "2.0.0",
    latestTag: "v2.0.0",
    releaseNotes: "Notes",
    downloadUrl: "https://example.com/z.zip",
    zipSha256: "a".repeat(64),
    releasePageUrl: "https://example.com/r",
    autoApplySupported: true,
  });
  const applying = ref(false);
  return {
    useAppUpdateLifecycle: () => ({
      dialogVisible,
      status,
      applying,
      dismissLater: vi.fn(),
      applyUpdate: vi.fn(),
      openReleasePage: vi.fn(),
      closeDialog: vi.fn(),
    }),
  };
});

describe("AppUpdateDialog", () => {
  beforeEach(() => {
    vi.spyOn(App, "applyAppUpdate").mockResolvedValue(undefined);
  });

  it("renders release notes when visible", async () => {
    const wrapper = mount(AppUpdateDialog, { attachTo: document.body });
    await flushPromises();
    expect(wrapper.get('[data-testid="app-update-notes"]').text()).toContain(
      "Notes",
    );
    wrapper.unmount();
  });
});
