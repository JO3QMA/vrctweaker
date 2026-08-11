import { ElMessage } from "element-plus";
import type { MessageHandler } from "element-plus";

export type ToastVariant = "success" | "warning" | "danger" | "info";

function show(variant: ToastVariant, message: string): MessageHandler {
  switch (variant) {
    case "success":
      return ElMessage.success(message);
    case "warning":
      return ElMessage.warning(message);
    case "danger":
      return ElMessage.error(message);
    case "info":
      return ElMessage.info(message);
    default: {
      const _exhaustive: never = variant;
      throw new Error(`Unknown toast variant: ${String(_exhaustive)}`);
    }
  }
}

/** Toast feedback (i18n-ready string only). See ADR 0022 / showToast adoption. */
export const showToast = {
  success: (message: string) => show("success", message),
  warning: (message: string) => show("warning", message),
  error: (message: string) => show("danger", message),
  info: (message: string) => show("info", message),
};
