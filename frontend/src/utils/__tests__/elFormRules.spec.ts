import { describe, it, expect, vi } from "vitest";
import {
  cookiesFilePathHasForbiddenChars,
  cookiesFilePathRule,
  integerMinRule,
  integerRangeRule,
  launchArgLineHasForbiddenChars,
  optionalLaunchArgLineRule,
  optionalPathLineRule,
  pathLineHasForbiddenChars,
  requiredTrimmedRule,
  scheduleWeekdaysRule,
} from "../elFormRules";

const t = vi.fn((key: string, params?: Record<string, string>) => {
  if (params) return `${key}:${JSON.stringify(params)}`;
  return key;
});

function runValidator(
  rule: ReturnType<typeof requiredTrimmedRule>,
  value: unknown,
): Promise<void> {
  return new Promise((resolve, reject) => {
    const validator = rule.validator;
    if (typeof validator !== "function") {
      reject(new Error("missing validator"));
      return;
    }
    validator(
      rule,
      value,
      (err?: Error | string) => {
        if (err) reject(err instanceof Error ? err : new Error(String(err)));
        else resolve();
      },
      {},
      {},
    );
  });
}

describe("elFormRules helpers", () => {
  it("detects forbidden path line breaks", () => {
    expect(pathLineHasForbiddenChars("ok")).toBe(false);
    expect(pathLineHasForbiddenChars("a\nb")).toBe(true);
    expect(cookiesFilePathHasForbiddenChars(`a"b`)).toBe(true);
    expect(launchArgLineHasForbiddenChars("-x\r")).toBe(true);
  });
});

describe("requiredTrimmedRule", () => {
  it("rejects whitespace-only values", async () => {
    const rule = requiredTrimmedRule(t);
    await expect(runValidator(rule, "   ")).rejects.toThrow(
      "formValidation.required",
    );
    await expect(runValidator(rule, "name")).resolves.toBeUndefined();
  });
});

describe("optionalPathLineRule", () => {
  it("allows empty and rejects line breaks", async () => {
    const rule = optionalPathLineRule(t);
    await expect(runValidator(rule, "")).resolves.toBeUndefined();
    await expect(runValidator(rule, "C:\\path")).resolves.toBeUndefined();
    await expect(runValidator(rule, "a\nb")).rejects.toThrow(
      "formValidation.pathInvalid",
    );
  });
});

describe("cookiesFilePathRule", () => {
  it("requires non-empty path without forbidden chars", async () => {
    const rule = cookiesFilePathRule(t);
    await expect(runValidator(rule, "")).rejects.toThrow(
      "formValidation.required",
    );
    await expect(runValidator(rule, "cookies.txt")).resolves.toBeUndefined();
    await expect(runValidator(rule, "bad'file")).rejects.toThrow(
      "formValidation.pathInvalid",
    );
  });
});

describe("integerMinRule", () => {
  it("enforces minimum integer", async () => {
    const rule = integerMinRule(30, t);
    await expect(runValidator(rule, 29)).rejects.toThrow(/integerMin/);
    await expect(runValidator(rule, 30)).resolves.toBeUndefined();
    await expect(runValidator(rule, null)).rejects.toThrow(
      "formValidation.required",
    );
    await expect(runValidator(rule, " ")).rejects.toThrow(
      "formValidation.required",
    );
    await expect(runValidator(rule, 3.5)).rejects.toThrow(
      "formValidation.integerRequired",
    );
  });
});

describe("integerRangeRule", () => {
  it("enforces inclusive range", async () => {
    const rule = integerRangeRule(0, 23, t);
    await expect(runValidator(rule, 24)).rejects.toThrow(/integerRange/);
    await expect(runValidator(rule, 12)).resolves.toBeUndefined();
    await expect(runValidator(rule, 3.5)).rejects.toThrow(
      "formValidation.integerRequired",
    );
  });
});

describe("optionalLaunchArgLineRule", () => {
  it("allows empty and rejects multiline", async () => {
    const rule = optionalLaunchArgLineRule(t);
    await expect(runValidator(rule, "")).resolves.toBeUndefined();
    await expect(runValidator(rule, "-batchmode")).resolves.toBeUndefined();
    await expect(runValidator(rule, "-a\n-b")).rejects.toThrow(
      "formValidation.launchArgInvalid",
    );
  });
});

describe("scheduleWeekdaysRule", () => {
  it("requires at least one weekday", async () => {
    const rule = scheduleWeekdaysRule(t);
    await expect(runValidator(rule, [])).rejects.toThrow(
      "formValidation.scheduleWeekdaysRequired",
    );
    await expect(runValidator(rule, [1])).resolves.toBeUndefined();
  });
});
