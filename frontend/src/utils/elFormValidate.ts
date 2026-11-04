import type { FormInstance } from "element-plus";

/** Runs Element Plus form validation; returns false when invalid or ref is missing. */
export async function validateElForm(
  formRef: FormInstance | undefined,
): Promise<boolean> {
  if (!formRef) return false;
  try {
    await formRef.validate();
    return true;
  } catch {
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
