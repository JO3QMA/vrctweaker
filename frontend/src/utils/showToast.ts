import { ElMessage } from "element-plus";
import type { MessageHandler, MessageParams } from "element-plus";
import { Z_INDEX_TOAST } from "../constants/elementPlusStack";

export type ToastVariant = "success" | "warning" | "danger" | "info";

const zIndexTarget = String(Z_INDEX_TOAST);

function messageNeedsZIndexPatch(node: HTMLElement): boolean {
  const inline = node.style.zIndex;
  if (inline === zIndexTarget) return false;
  if (inline !== "") return true;
  return getComputedStyle(node).zIndex !== zIndexTarget;
}

function applyToastStackZIndex(): void {
  if (typeof document === "undefined") return;
  document.querySelectorAll(".el-message").forEach((node) => {
    if (node instanceof HTMLElement && messageNeedsZIndexPatch(node)) {
      node.style.zIndex = zIndexTarget;
    }
  });
}

let toastZIndexGuard: MutationObserver | undefined;

/** Re-apply when EP rewrites inline z-index (e.g. after another overlay opens). */
function ensureToastZIndexGuard(): void {
  if (typeof document === "undefined" || toastZIndexGuard) return;
  toastZIndexGuard = new MutationObserver(() => {
    if (!document.querySelector(".el-message")) return;
    if (
      [...document.querySelectorAll(".el-message")].some(
        (n) => n instanceof HTMLElement && messageNeedsZIndexPatch(n),
      )
    ) {
      applyToastStackZIndex();
    }
  });
  toastZIndexGuard.observe(document.body, {
    childList: true,
    subtree: true,
    attributes: true,
    attributeFilter: ["style"],
  });
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
  ensureToastZIndexGuard();
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
