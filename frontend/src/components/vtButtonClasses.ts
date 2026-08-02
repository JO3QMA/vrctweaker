/**
 * ElMessageBox button classes shared with VtButton styling (ADR 0017).
 * MessageBox cannot render VtButton; use these constants for confirm/cancel.
 */

/** Destructive confirm (Danger button in dialogs). */
export const VT_BUTTON_DANGER_CONFIRM_CLASS = "el-button--danger";

/**
 * Cancel in destructive confirm dialogs (Secondary button).
 * Empty — Element Plus default button matches our secondary filled style.
 */
export const VT_BUTTON_SECONDARY_CANCEL_CLASS = "";
