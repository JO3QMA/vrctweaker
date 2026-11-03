import { onMounted, onUnmounted, ref } from "vue";
import { getRuntime } from "../wails/runtime";
import { App as WailsApp } from "../wails/app";

export type StartupInitStatus = {
  ready: boolean;
  loading: boolean;
  phase: string;
  summaryKey: string;
  detail: string;
  dataDir: string;
  canOpenDataFolder: boolean;
};

const emptyStatus = (): StartupInitStatus => ({
  ready: import.meta.env.MODE === "e2e",
  loading: import.meta.env.MODE !== "e2e",
  phase: import.meta.env.MODE === "e2e" ? "" : "opening",
  summaryKey: "",
  detail: "",
  dataDir: "",
  canOpenDataFolder: false,
});

export function useStartupInit() {
  const status = ref<StartupInitStatus>(emptyStatus());
  const retrying = ref(false);

  async function refreshStatus(): Promise<void> {
    status.value = await WailsApp.getStartupInitStatus();
  }

  async function retryInit(): Promise<void> {
    retrying.value = true;
    try {
      status.value = await WailsApp.retryStartupInit();
    } finally {
      retrying.value = false;
    }
  }

  async function openDataFolder(): Promise<void> {
    await WailsApp.openAppDataFolder();
  }

  let cancelListen: (() => void) | undefined;

  onMounted(async () => {
    await refreshStatus();
    const eventsOn = getRuntime()?.EventsOn;
    if (eventsOn) {
      cancelListen = eventsOn("app:startup-status-changed", (data?: unknown) => {
        status.value = data as StartupInitStatus;
      });
    }
  });

  onUnmounted(() => {
    cancelListen?.();
  });

  return {
    status,
    retrying,
    refreshStatus,
    retryInit,
    openDataFolder,
  };
}
