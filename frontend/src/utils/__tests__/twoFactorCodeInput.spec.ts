import { describe, expect, it } from "vitest";
import {
  isTwoFactorCodeComplete,
  sanitizeTwoFactorCodeInput,
  sanitizeTwoFactorCodePaste,
} from "../twoFactorCodeInput";

describe("twoFactorCodeInput", () => {
  it("sanitizes input to digits only with max length", () => {
    expect(sanitizeTwoFactorCodeInput("12a34b56c78")).toBe("123456");
    expect(sanitizeTwoFactorCodeInput("1234567890")).toBe("123456");
  });

  it("sanitizes paste with spaces and hyphens", () => {
    expect(sanitizeTwoFactorCodePaste("12 34-56")).toBe("123456");
  });

  it("detects complete codes", () => {
    expect(isTwoFactorCodeComplete("123456")).toBe(true);
    expect(isTwoFactorCodeComplete("12345")).toBe(false);
    expect(isTwoFactorCodeComplete("12ab3456")).toBe(true);
  });
});
