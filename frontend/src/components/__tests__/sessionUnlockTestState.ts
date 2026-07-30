import { ref } from "vue";
import type { UnlockState } from "../../composables/useSessionUnlock";

export const unlockState = ref<UnlockState>("idle");
