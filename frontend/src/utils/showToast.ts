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

function patchMessageZIndex(node: HTMLElement): void {
  if (messageNeedsZIndexPatch(node)) {
    node.style.zIndex = zIndexTarget;
  }
}

function applyToastStackZIndex(): void {
  if (typeof document === "undefined") return;
  document.querySelectorAll(".el-message").forEach((node) => {
    if (node instanceof HTMLElement) {
      patchMessageZIndex(node);
    }
  });
}

const messageObserverByEl = new Map<HTMLElement, MutationObserver>();
let messageListObserver: MutationObserver | undefined;

function forEachMessageInNode(
  node: Node,
  visit: (el: HTMLElement) => void,
): void {
  if (!(node instanceof HTMLElement)) return;
  if (node.classList.contains("el-message")) {
    visit(node);
  }
  node.querySelectorAll(".el-message").forEach((el) => {
    if (el instanceof HTMLElement) visit(el);
  });
}

function observeMessageElement(el: HTMLElement): void {
  patchMessageZIndex(el);
  if (messageObserverByEl.has(el)) return;
  const observer = new MutationObserver(() => {
    patchMessageZIndex(el);
  });
  observer.observe(el, { attributes: true, attributeFilter: ["style"] });
  messageObserverByEl.set(el, observer);
}

function unobserveMessageElement(el: HTMLElement): void {
  messageObserverByEl.get(el)?.disconnect();
  messageObserverByEl.delete(el);
}

function teardownToastWatchersIfIdle(): void {
  if (document.querySelector(".el-message")) return;
  messageListObserver?.disconnect();
  messageListObserver = undefined;
  for (const [, observer] of messageObserverByEl) {
    observer.disconnect();
  }
  messageObserverByEl.clear();
}

/** Watch only toast mount/unmount; per-toast observers handle EP style rewrites. */
function ensureToastZIndexWatchers(): void {
  if (typeof document === "undefined" || messageListObserver) return;
  messageListObserver = new MutationObserver((records) => {
    for (const record of records) {
      record.addedNodes.forEach((node) => {
        forEachMessageInNode(node, observeMessageElement);
      });
      record.removedNodes.forEach((node) => {
        forEachMessageInNode(node, unobserveMessageElement);
      });
    }
    teardownToastWatchersIfIdle();
  });
  messageListObserver.observe(document.body, {
    childList: true,
    subtree: true,
  });
  document.querySelectorAll(".el-message").forEach((node) => {
    if (node instanceof HTMLElement) observeMessageElement(node);
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
  ensureToastZIndexWatchers();
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

/** @internal Vitest hooks for observer lifecycle */
export const showToastStackInternals = {
  activePerMessageObservers: (): number => messageObserverByEl.size,
  isMessageListObserverActive: (): boolean => messageListObserver !== undefined,
};
