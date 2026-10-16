import { onMounted, onUnmounted, ref } from "vue";
import { App, type AppUpdateStatusDTO } from "../wails/app";
import { getRuntime } from "../wails/runtime";

/**
 * Single shared store: SettingsView manual checks and AppUpdateDialog must share
 * dialogVisible/status so a settings-triggered check opens the global dialog.
 */
function createAppUpdateStore() {
  const dialogVisible = ref(false);
  const status = ref<AppUpdateStatusDTO | null>(null);
  const applying = ref(false);
  let eventsBound = false;

  function openDialog(st: AppUpdateStatusDTO) {
    status.value = st;
    dialogVisible.value = true;
  }

  function bindEvents() {
    if (eventsBound) return undefined;
    const rt = getRuntime();
    const off = rt?.EventsOn?.("app:update-available", (data?: unknown) => {
      const st = data as AppUpdateStatusDTO | undefined;
      if (st?.shouldPrompt) {
        void App.clearStartupAppUpdatePrompt();
        openDialog(st);
      }
    });
    if (!off) return undefined;
    eventsBound = true;
    return () => {
      off();
      eventsBound = false;
    };
  }

  async function claimStartupPrompt() {
    const st = await App.claimStartupAppUpdatePrompt();
    if (st.updateAvailable && st.shouldPrompt) {
      openDialog(st);
    }
  }

  async function checkManual(): Promise<AppUpdateStatusDTO> {
    const st = await App.checkForAppUpdate();
    if (st.updateAvailable && st.shouldPrompt) {
      openDialog(st);
    }
    return st;
  }

  async function dismissLater() {
    const v = status.value?.latestVersion;
    if (v) {
      await App.dismissAppUpdate(v);
    }
    dialogVisible.value = false;
  }

  async function applyUpdate() {
    const st = status.value;
    if (!st?.latestVersion) {
      throw new Error("app update: missing version");
    }
    applying.value = true;
    try {
      await App.applyAppUpdate(st.latestVersion);
    } finally {
      applying.value = false;
    }
  }

  async function openReleasePage() {
    const url = status.value?.releasePageUrl;
    if (!url) {
      throw new Error("app update: missing release page url");
    }
    await App.openAppReleasePage(url);
  }

  return {
    dialogVisible,
    status,
    applying,
    bindEvents,
    claimStartupPrompt,
    checkManual,
    dismissLater,
    applyUpdate,
    openReleasePage,
  };
}

const appUpdateStore = createAppUpdateStore();

export function useAppUpdate() {
  return appUpdateStore;
}

export function useAppUpdateLifecycle() {
  const api = useAppUpdate();
  let off: (() => void) | undefined;
  onMounted(() => {
    off = api.bindEvents();
    void api.claimStartupPrompt();
  });
  onUnmounted(() => {
    off?.();
  });
  return api;
}
