import type { FormItemRule } from "element-plus";

export type FormRuleTranslator = (
  key: string,
  params?: Record<string, string>,
) => string;

const LINE_BREAK_FORBIDDEN = /[\n\r]/;
const COOKIES_FILE_FORBIDDEN = /[\n\r"']/;

export function pathLineHasForbiddenChars(value: string): boolean {
  return LINE_BREAK_FORBIDDEN.test(value);
}

export function cookiesFilePathHasForbiddenChars(value: string): boolean {
  return COOKIES_FILE_FORBIDDEN.test(value);
}

export function launchArgLineHasForbiddenChars(value: string): boolean {
  return LINE_BREAK_FORBIDDEN.test(value);
}

function requiredMessage(t: FormRuleTranslator): string {
  return t("formValidation.required");
}

function parseIntegerFieldValue(value: unknown): number | null | typeof NaN {
  if (value === null || value === undefined || value === "") {
    return null;
  }
  if (typeof value === "string" && value.trim() === "") {
    return null;
  }
  if (typeof value !== "number") {
    return NaN;
  }
  if (!Number.isFinite(value) || !Number.isInteger(value)) {
    return NaN;
  }
  return value;
}

export function requiredTrimmedRule(t: FormRuleTranslator): FormItemRule {
  return {
    validator: (_rule, value, callback) => {
      const s = typeof value === "string" ? value.trim() : "";
      if (!s) {
        callback(new Error(requiredMessage(t)));
        return;
      }
      callback();
    },
    trigger: ["blur", "change"],
  };
}

/** Optional folder/file path: empty OK; non-empty must not contain line breaks. */
export function optionalPathLineRule(t: FormRuleTranslator): FormItemRule {
  return {
    validator: (_rule, value, callback) => {
      const s = typeof value === "string" ? value : String(value ?? "");
      if (!s.trim()) {
        callback();
        return;
      }
      if (pathLineHasForbiddenChars(s)) {
        callback(new Error(t("formValidation.pathInvalid")));
        return;
      }
      callback();
    },
    trigger: ["blur", "change"],
  };
}

/** Cookies file path (yt-dlp): required; matches backend forbidden characters. */
export function cookiesFilePathRule(t: FormRuleTranslator): FormItemRule {
  return {
    validator: (_rule, value, callback) => {
      const s = typeof value === "string" ? value.trim() : "";
      if (!s) {
        callback(new Error(requiredMessage(t)));
        return;
      }
      if (cookiesFilePathHasForbiddenChars(s)) {
        callback(new Error(t("formValidation.pathInvalid")));
        return;
      }
      callback();
    },
    trigger: ["blur", "change"],
  };
}

export function integerMinRule(
  min: number,
  t: FormRuleTranslator,
): FormItemRule {
  return {
    type: "number",
    validator: (_rule, value, callback) => {
      const parsed = parseIntegerFieldValue(value);
      if (parsed === null) {
        callback(new Error(requiredMessage(t)));
        return;
      }
      if (typeof parsed !== "number" || Number.isNaN(parsed)) {
        callback(new Error(t("formValidation.integerRequired")));
        return;
      }
      if (parsed < min) {
        callback(
          new Error(t("formValidation.integerMin", { min: String(min) })),
        );
        return;
      }
      callback();
    },
    trigger: ["blur", "change"],
  };
}

export function integerRangeRule(
  min: number,
  max: number,
  t: FormRuleTranslator,
): FormItemRule {
  return {
    type: "number",
    validator: (_rule, value, callback) => {
      const parsed = parseIntegerFieldValue(value);
      if (parsed === null) {
        callback(new Error(requiredMessage(t)));
        return;
      }
      if (typeof parsed !== "number" || Number.isNaN(parsed)) {
        callback(new Error(t("formValidation.integerRequired")));
        return;
      }
      if (parsed < min || parsed > max) {
        callback(
          new Error(
            t("formValidation.integerRange", {
              min: String(min),
              max: String(max),
            }),
          ),
        );
        return;
      }
      callback();
    },
    trigger: ["blur", "change"],
  };
}

/** Single-line launch argument fragment (custom / OSC / MIDI / trackers). */
export function optionalLaunchArgLineRule(t: FormRuleTranslator): FormItemRule {
  return {
    validator: (_rule, value, callback) => {
      const s = typeof value === "string" ? value : String(value ?? "");
      if (!s.trim()) {
        callback();
        return;
      }
      if (launchArgLineHasForbiddenChars(s)) {
        callback(new Error(t("formValidation.launchArgInvalid")));
        return;
      }
      callback();
    },
    trigger: ["blur", "change"],
  };
}

export function scheduleWeekdaysRule(t: FormRuleTranslator): FormItemRule {
  return {
    type: "array",
    validator: (_rule, value, callback) => {
      const days = Array.isArray(value) ? value : [];
      if (days.length === 0) {
        callback(new Error(t("formValidation.scheduleWeekdaysRequired")));
        return;
      }
      callback();
    },
    trigger: "change",
  };
}
