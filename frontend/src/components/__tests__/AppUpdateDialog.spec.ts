import { describe, it, expect, vi, beforeEach } from "vitest";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { mount, flushPromises } from "@vue/test-utils";
import AppUpdateDialog from "../AppUpdateDialog.vue";
import { App } from "../../wails/app";

const appUpdateDialogSource = readFileSync(
  path.join(
    path.dirname(fileURLToPath(import.meta.url)),
    "../AppUpdateDialog.vue",
  ),
  "utf8",
);

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

  it("keeps keyboard focus inside the dialog", async () => {
    const outside = document.createElement("button");
    outside.textContent = "outside";
    document.body.appendChild(outside);
    outside.focus();

    mount(AppUpdateDialog, { attachTo: document.body });
    await flushPromises();

    const dialog = document.querySelector('[data-testid="app-update-dialog"]');
    expect(dialog).toBeTruthy();
    expect(dialog?.contains(document.activeElement)).toBe(true);

    document.body.innerHTML = "";
  });

  it("applies brand focus ring styles to teleported dialog buttons", async () => {
    expect(appUpdateDialogSource).toContain(".app-update-dialog");
    expect(appUpdateDialogSource).toContain(".dialog-actions");
    expect(appUpdateDialogSource).toContain(
      ".el-button.is-text:not(.is-disabled):focus-visible",
    );
    expect(appUpdateDialogSource).toContain("outline: 2px solid #5c9ce6");
    expect(appUpdateDialogSource).toContain("outline-offset: 2px");

    mount(AppUpdateDialog, { attachTo: document.body });
    await flushPromises();
    const later = document.querySelector(
      '[data-testid="app-update-later"]',
    ) as HTMLButtonElement | null;
    expect(later).toBeTruthy();
    expect(later!.closest(".app-update-dialog")).toBeTruthy();

    document.body.innerHTML = "";
  });
});
