import type { FormInstance } from "element-plus";
import { onBeforeUnmount, watchEffect, type Ref } from "vue";
import { syncValidatedFormFieldA11y } from "../utils/formFieldA11y";

/** Keeps aria-invalid / aria-describedby in sync when EP re-validates fields. */
export function useValidatedFormA11y(
  formRef: Ref<FormInstance | undefined>,
): void {
  const stop = watchEffect(
    () => {
      const form = formRef.value;
      if (!form) return;
      void form.fields.length;
      for (const field of form.fields) {
        void field.validateState;
        void field.validateMessage;
        void field.propString;
      }
      void syncValidatedFormFieldA11y(form).catch(() => {});
    },
    { flush: "post" },
  );

  onBeforeUnmount(() => {
    stop();
  });
}

/** Watches several validated forms (e.g. Config save group). */
export function useValidatedFormsA11y(
  formRefs: Ref<FormInstance | undefined>[],
): void {
  for (const ref of formRefs) {
    useValidatedFormA11y(ref);
  }
}
