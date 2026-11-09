import type { FormInstance } from "element-plus";
import { nextTick } from "vue";

const FORM_A11Y_SCOPE_ATTR = "data-vt-form-a11y-scope";

const FOCUSABLE_CONTROL_SELECTOR =
  "input:not([type='hidden']):not([disabled]), textarea:not([disabled]), select:not([disabled])";

/** Stable id for Element Plus inline validation error copy. */
export function formFieldValidationErrorId(
  formScopeId: string,
  prop: string,
): string {
  const safeProp = prop.replace(/\./g, "-");
  return `${formScopeId}-${safeProp}-error`;
}

export function ensureFormA11yScopeId(formEl: HTMLElement): string {
  let scopeId = formEl.getAttribute(FORM_A11Y_SCOPE_ATTR);
  if (!scopeId) {
    scopeId =
      typeof crypto !== "undefined" && "randomUUID" in crypto
        ? `vt-form-${crypto.randomUUID()}`
        : `vt-form-${Math.random().toString(36).slice(2, 10)}`;
    formEl.setAttribute(FORM_A11Y_SCOPE_ATTR, scopeId);
  }
  return scopeId;
}

function rangeHintIdsInFormItem(formItemEl: HTMLElement): string[] {
  return Array.from(
    formItemEl.querySelectorAll<HTMLElement>(".form-field-range-hint[id]"),
  )
    .map((el) => el.id)
    .filter((id) => id.length > 0);
}

function findControlsInFormItem(formItemEl: HTMLElement): HTMLElement[] {
  const inputNumberInputs = formItemEl.querySelectorAll<HTMLElement>(
    ".el-input-number input:not([type='hidden'])",
  );
  if (inputNumberInputs.length > 0) {
    return Array.from(inputNumberInputs);
  }
  return Array.from(
    formItemEl.querySelectorAll<HTMLElement>(FOCUSABLE_CONTROL_SELECTOR),
  );
}

function applyControlDisabledA11y(control: HTMLElement): void {
  const disabled =
    control instanceof HTMLInputElement ||
    control instanceof HTMLTextAreaElement ||
    control instanceof HTMLSelectElement ||
    control instanceof HTMLButtonElement
      ? control.disabled
      : control.hasAttribute("disabled");
  if (disabled) {
    control.setAttribute("aria-disabled", "true");
  } else {
    control.removeAttribute("aria-disabled");
  }
}

/** Skips ids that are not present in the document (e.g. error node not painted yet). */
function describedByForField(
  errorId: string | undefined,
  hintIds: string[],
): string | undefined {
  const ids = [
    ...(errorId && document.getElementById(errorId) ? [errorId] : []),
    ...hintIds.filter((id) => document.getElementById(id)),
  ];
  if (ids.length === 0) return undefined;
  return ids.join(" ");
}

function applyControlDescribedBy(
  control: HTMLElement,
  describedBy: string | undefined,
): void {
  if (describedBy) {
    control.setAttribute("aria-describedby", describedBy);
  } else {
    control.removeAttribute("aria-describedby");
  }
}

/** Wait one frame so Element Plus can paint inline errors after nextTick. */
async function waitForFormItemErrorPaint(formRef: FormInstance): Promise<void> {
  await nextTick();
  if (!formRef.fields.some((field) => field.validateState === "error")) {
    return;
  }
  await new Promise<void>((resolve) => {
    requestAnimationFrame(() => resolve());
  });
  await nextTick();
}

/**
 * Sets aria-invalid and aria-describedby (error id + optional range hints) on
 * controls inside validated el-form items.
 */
export async function syncValidatedFormFieldA11y(
  formRef: FormInstance,
): Promise<void> {
  await nextTick();
  if (formRef.fields.some((field) => field.validateState === "error")) {
    await waitForFormItemErrorPaint(formRef);
  }
  const formEl = formRef.$el as HTMLElement | undefined;
  if (!formEl) return;
  const scopeId = ensureFormA11yScopeId(formEl);

  for (const field of formRef.fields) {
    const prop = field.propString;
    const itemEl = field.$el;
    if (!prop || !itemEl) continue;

    const controls = findControlsInFormItem(itemEl);
    if (controls.length === 0) continue;

    const hintIds = rangeHintIdsInFormItem(itemEl);
    const isError = field.validateState === "error";

    if (isError) {
      const errorEl = itemEl.querySelector<HTMLElement>(".el-form-item__error");
      const errorId = formFieldValidationErrorId(scopeId, prop);
      const paintedErrorId = errorEl ? errorId : undefined;
      if (errorEl) {
        errorEl.id = errorId;
      }
      const describedBy = describedByForField(paintedErrorId, hintIds);
      for (const control of controls) {
        if (errorEl) {
          control.setAttribute("aria-invalid", "true");
        } else {
          control.removeAttribute("aria-invalid");
        }
        applyControlDescribedBy(control, describedBy);
        applyControlDisabledA11y(control);
      }
    } else {
      const describedBy = describedByForField(undefined, hintIds);
      for (const control of controls) {
        control.removeAttribute("aria-invalid");
        applyControlDescribedBy(control, describedBy);
        applyControlDisabledA11y(control);
      }
      const staleErrorId = formFieldValidationErrorId(scopeId, prop);
      const stale = itemEl.querySelector(`#${CSS.escape(staleErrorId)}`);
      stale?.removeAttribute("id");
    }
  }
}

/** Bottom inset (px) for sticky bars obscuring invalid fields. */
export function stickyBottomInsetForForm(
  formEl: HTMLElement,
  stickyBarRootSelector = ".automation-rule-editor",
  stickyBarSelector = ".automation-action-bar",
  extraPaddingPx = 8,
): number {
  const bar = formEl
    .closest(stickyBarRootSelector)
    ?.querySelector<HTMLElement>(stickyBarSelector);
  if (!bar) return 0;
  return Math.ceil(bar.getBoundingClientRect().height) + extraPaddingPx;
}

/**
 * Scrolls a form item into view, keeping it above a sticky footer when present.
 */
export function scrollFormItemIntoView(
  itemEl: HTMLElement,
  bottomInset = 0,
): void {
  itemEl.scrollIntoView({ block: "nearest", inline: "nearest" });
  if (bottomInset <= 0 || typeof window === "undefined") return;
  const rect = itemEl.getBoundingClientRect();
  const maxBottom = window.innerHeight - bottomInset;
  if (rect.bottom > maxBottom) {
    window.scrollBy({ top: rect.bottom - maxBottom, behavior: "auto" });
  } else if (rect.top < 0) {
    window.scrollBy({ top: rect.top, behavior: "auto" });
  }
}

export function scrollFormFieldIntoView(
  formRef: FormInstance,
  prop: string,
): void {
  const field = formRef.fields.find((f) => f.propString === prop);
  const formEl = formRef.$el as HTMLElement | undefined;
  if (field?.$el && formEl) {
    scrollFormItemIntoView(field.$el, stickyBottomInsetForForm(formEl));
    return;
  }
  formRef.scrollToField(prop);
}
