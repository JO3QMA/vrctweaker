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
    scopeId = `vt-form-${Math.random().toString(36).slice(2, 10)}`;
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
  return Array.from(
    formItemEl.querySelectorAll<HTMLElement>(FOCUSABLE_CONTROL_SELECTOR),
  );
}

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

/** Wait for Element Plus debounced error message to appear in the DOM. */
async function waitForFormItemErrorPaint(formRef: FormInstance): Promise<void> {
  await nextTick();
  const deadline = Date.now() + 500;
  while (Date.now() < deadline) {
    const painted = formRef.fields.some((field) => {
      if (field.validateState !== "error") return false;
      return !!field.$el?.querySelector(".el-form-item__error");
    });
    if (painted) return;
    await new Promise((resolve) => setTimeout(resolve, 25));
  }
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
      if (errorEl) {
        errorEl.id = errorId;
      }
      const describedBy = describedByForField(errorId, hintIds);
      for (const control of controls) {
        control.setAttribute("aria-invalid", "true");
        applyControlDescribedBy(control, describedBy);
      }
    } else {
      const describedBy = describedByForField(undefined, hintIds);
      for (const control of controls) {
        control.removeAttribute("aria-invalid");
        applyControlDescribedBy(control, describedBy);
      }
    }
  }
}

/** Bottom inset (px) for sticky bars obscuring invalid fields. */
export function stickyBottomInsetForForm(formEl: HTMLElement): number {
  const editorRoot = formEl.closest(".automation-rule-editor");
  const bar = editorRoot?.querySelector<HTMLElement>(".automation-action-bar");
  if (!bar) return 0;
  return Math.ceil(bar.getBoundingClientRect().height) + 8;
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
