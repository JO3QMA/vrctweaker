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

const STATUS_FETCH_FAILURE_KEY = "app.startup.failed.statusFetch";

const emptyStatus = (): StartupInitStatus => ({
  ready: import.meta.env.MODE === "e2e",
  loading: import.meta.env.MODE !== "e2e",
  phase: import.meta.env.MODE === "e2e" ? "" : "opening",
  summaryKey: "",
  detail: "",
  dataDir: "",
  canOpenDataFolder: false,
});

function statusFetchFailure(): StartupInitStatus {
  return {
    ready: false,
    loading: false,
    phase: "",
    summaryKey: STATUS_FETCH_FAILURE_KEY,
    detail: "",
    dataDir: "",
    canOpenDataFolder: false,
  };
}

export function useStartupInit() {
  const status = ref<StartupInitStatus>(emptyStatus());
  const retrying = ref(false);
  let receivedEvent = false;

  async function refreshStatus(): Promise<void> {
    try {
      const snapshot = await WailsApp.getStartupInitStatus();
      if (!receivedEvent) {
        status.value = snapshot;
      }
    } catch {
      if (!receivedEvent) {
        status.value = statusFetchFailure();
      }
    }
  }

  async function retryInit(): Promise<void> {
    retrying.value = true;
    try {
      status.value = await WailsApp.retryStartupInit();
    } catch {
      try {
        await refreshStatus();
      } catch {
        status.value = statusFetchFailure();
      }
    } finally {
      retrying.value = false;
    }
  }

  async function openDataFolder(): Promise<void> {
    await WailsApp.openAppDataFolder();
  }

  let cancelListen: (() => void) | undefined;

  onMounted(async () => {
    const eventsOn = getRuntime()?.EventsOn;
    if (eventsOn) {
      cancelListen = eventsOn(
        "app:startup-status-changed",
        (data?: unknown) => {
          receivedEvent = true;
          status.value = data as StartupInitStatus;
        },
      );
    }
    await refreshStatus();
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
