import type { FormItemRule } from "element-plus";

export type FormRuleTranslator = (
  key: string,
  params?: Record<string, string>,
) => string;

const PATH_LINE_FORBIDDEN = /[\n\r]/;
const COOKIES_FILE_FORBIDDEN = /[\n\r"']/;
const LAUNCH_ARG_LINE_FORBIDDEN = /[\n\r]/;

export function pathLineHasForbiddenChars(value: string): boolean {
  return PATH_LINE_FORBIDDEN.test(value);
}

export function cookiesFilePathHasForbiddenChars(value: string): boolean {
  return COOKIES_FILE_FORBIDDEN.test(value);
}

export function launchArgLineHasForbiddenChars(value: string): boolean {
  return LAUNCH_ARG_LINE_FORBIDDEN.test(value);
}

function requiredMessage(t: FormRuleTranslator): string {
  return t("formValidation.required");
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
      if (value === null || value === undefined || value === "") {
        callback(new Error(requiredMessage(t)));
        return;
      }
      const n = Number(value);
      if (!Number.isFinite(n) || !Number.isInteger(n) || n < min) {
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
      if (value === null || value === undefined || value === "") {
        callback(new Error(requiredMessage(t)));
        return;
      }
      const n = Number(value);
      if (!Number.isFinite(n) || !Number.isInteger(n) || n < min || n > max) {
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
