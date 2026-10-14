import { onMounted, onUnmounted, ref } from "vue";
import { App, type AppUpdateStatusDTO } from "../wails/app";
import { getRuntime } from "../wails/runtime";

const dialogVisible = ref(false);
const status = ref<AppUpdateStatusDTO | null>(null);
const applying = ref(false);
let eventsBound = false;

function openDialog(st: AppUpdateStatusDTO) {
  status.value = st;
  dialogVisible.value = true;
}

export function useAppUpdate() {
  function bindEvents() {
    if (eventsBound) return;
    const rt = getRuntime();
    const off = rt?.EventsOn?.("app:update-available", (data?: unknown) => {
      const st = data as AppUpdateStatusDTO | undefined;
      if (st?.shouldPrompt) {
        openDialog(st);
      }
    });
    eventsBound = true;
    return off;
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
    if (!st?.downloadUrl || !st.zipSha256) return;
    applying.value = true;
    try {
      await App.applyAppUpdate(st.downloadUrl, st.zipSha256);
    } finally {
      applying.value = false;
    }
  }

  async function openReleasePage() {
    const url = status.value?.releasePageUrl;
    if (!url) return;
    await App.openAppReleasePage(url);
  }

  function closeDialog() {
    dialogVisible.value = false;
  }

  return {
    dialogVisible,
    status,
    applying,
    bindEvents,
    checkManual,
    dismissLater,
    applyUpdate,
    openReleasePage,
    closeDialog,
  };
}

export function useAppUpdateLifecycle() {
  const api = useAppUpdate();
  let off: (() => void) | undefined;
  onMounted(() => {
    off = api.bindEvents();
  });
  onUnmounted(() => {
    off?.();
  });
  return api;
}
