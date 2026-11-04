import type { FormInstance } from "element-plus";
import { nextTick } from "vue";

type InvalidFieldsMap = Record<string, unknown> | undefined;

function invalidFieldsFromFailure(failure: unknown): InvalidFieldsMap {
  if (failure && typeof failure === "object" && !Array.isArray(failure)) {
    return failure as Record<string, unknown>;
  }
  return undefined;
}

async function focusFirstInvalidField(
  formRef: FormInstance,
  invalidFields: InvalidFieldsMap,
): Promise<void> {
  const firstProp = invalidFields ? Object.keys(invalidFields)[0] : undefined;
  if (firstProp) {
    formRef.scrollToField(firstProp);
  }
  await nextTick();
  const formEl = formRef.$el as HTMLElement | undefined;
  if (!formEl) return;
  const focusTarget = formEl.querySelector<HTMLElement>(
    ".el-form-item.is-error input:not([type='hidden']):not([disabled]), " +
      ".el-form-item.is-error textarea:not([disabled])",
  );
  focusTarget?.focus({ preventScroll: true });
}

/** Runs Element Plus form validation; returns false when invalid or ref is missing. */
export async function validateElForm(
  formRef: FormInstance | undefined,
): Promise<boolean> {
  if (!formRef) return false;
  try {
    await formRef.validate();
    return true;
  } catch (failure) {
    await focusFirstInvalidField(formRef, invalidFieldsFromFailure(failure));
    return false;
  }
}

/** Validates every form; stops at the first failure. */
export async function validateElForms(
  formRefs: (FormInstance | undefined)[],
): Promise<boolean> {
  for (const ref of formRefs) {
    if (!(await validateElForm(ref))) {
      return false;
    }
  }
  return true;
}
