import type { FormRuleTranslator } from "./elFormRules";

/** Small hint under numeric inputs, e.g. `1280〜7680` (from i18n + limits). */
export function formatNumericRangeHint(
  t: FormRuleTranslator,
  min: number,
  max: number,
): string {
  return t("formValidation.numericRangeHint", {
    min: String(min),
    max: String(max),
  });
}
