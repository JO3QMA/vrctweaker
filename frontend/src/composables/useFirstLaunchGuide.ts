import { readonly, ref } from "vue";

const manualOpen = ref(false);

export function resetFirstLaunchGuideForTest(): void {
  manualOpen.value = false;
}

export function useFirstLaunchGuide() {
  function requestFirstLaunchGuide(): void {
    manualOpen.value = true;
  }

  function clearManualFirstLaunchRequest(): void {
    manualOpen.value = false;
  }

  return {
    manualOpen: readonly(manualOpen),
    requestFirstLaunchGuide,
    clearManualFirstLaunchRequest,
  };
}
