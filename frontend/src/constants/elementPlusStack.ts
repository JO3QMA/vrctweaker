/**
 * Overlay stacking (bottom → top): onboarding → EP dialog → MessageBox → toast.
 * Document values in PR / docs when changing.
 */
export const Z_INDEX_ONBOARDING = 1500;

/** Passed to `el-config-provider` `:z-index`. Dialogs / poppers use `nextZIndex()` from this base. */
export const Z_INDEX_ELEMENT_PLUS_BASE = 2000;

/** Typical first dialog overlay after `nextZIndex()` (base + 1). */
export const Z_INDEX_DIALOG_FIRST = Z_INDEX_ELEMENT_PLUS_BASE + 1;

/** MessageBox uses the same EP z-index counter; opens above an open dialog when shown later. */
export const Z_INDEX_MESSAGEBOX_TYPICAL = Z_INDEX_ELEMENT_PLUS_BASE + 2;

/** Toast (`ElMessage`) — enforced in `style.css` so it stays above MessageBox. */
export const Z_INDEX_TOAST = 4000;
