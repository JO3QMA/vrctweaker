import { ref } from "vue";

const manualOpen = ref(false);

export function useFirstLaunchGuide() {
  function requestFirstLaunchGuide(): void {
    manualOpen.value = true;
  }

  function clearManualFirstLaunchRequest(): void {
    manualOpen.value = false;
  }

  return {
    manualOpen,
    requestFirstLaunchGuide,
    clearManualFirstLaunchRequest,
  };
}
