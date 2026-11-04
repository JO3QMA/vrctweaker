import { ElMessage } from "element-plus";
import type { MessageHandler, MessageParams } from "element-plus";
import { Z_INDEX_TOAST } from "../constants/elementPlusStack";

export type ToastVariant = "success" | "warning" | "danger" | "info";

function applyToastStackZIndex(): void {
  if (typeof document === "undefined") return;
  const nodes = document.querySelectorAll(".el-message");
  const last = nodes.item(nodes.length - 1);
  if (last instanceof HTMLElement) {
    last.style.zIndex = String(Z_INDEX_TOAST);
  }
}

/** ElMessage sets z-index from ConfigProvider counter on mount; patch after paint. */
function scheduleApplyToastStackZIndex(): void {
  const run = (): void => applyToastStackZIndex();
  queueMicrotask(run);
  requestAnimationFrame(() => {
    run();
    requestAnimationFrame(run);
  });
}

function toastOptions(message: string): MessageParams {
  return {
    message,
    zIndex: Z_INDEX_TOAST,
  };
}

function show(variant: ToastVariant, message: string): MessageHandler {
  const options = toastOptions(message);
  let handler: MessageHandler;
  switch (variant) {
    case "success":
      handler = ElMessage.success(options);
      break;
    case "warning":
      handler = ElMessage.warning(options);
      break;
    case "danger":
      handler = ElMessage.error(options);
      break;
    case "info":
      handler = ElMessage.info(options);
      break;
    default: {
      const _exhaustive: never = variant;
      throw new Error(`Unknown toast variant: ${String(_exhaustive)}`);
    }
  }
  scheduleApplyToastStackZIndex();
  return handler;
}

/** Toast feedback (i18n-ready string only). See ADR 0022 / showToast adoption. */
export const showToast = {
  success: (message: string) => show("success", message),
  warning: (message: string) => show("warning", message),
  error: (message: string) => show("danger", message),
  info: (message: string) => show("info", message),
};
