import { describe, expect, it } from "vitest";
import {
  isSanitizedTwoFactorCodeComplete,
  isTwoFactorCodeComplete,
  sanitizeTwoFactorCodeInput,
} from "../twoFactorCodeInput";

describe("twoFactorCodeInput", () => {
  it("sanitizes input to digits only with max length", () => {
    expect(sanitizeTwoFactorCodeInput("12a34b56c78")).toBe("123456");
    expect(sanitizeTwoFactorCodeInput("1234567890")).toBe("123456");
  });

  it("strips paste separators while sanitizing", () => {
    expect(sanitizeTwoFactorCodeInput("12 34-56")).toBe("123456");
  });

  it("detects complete raw codes after sanitizing", () => {
    expect(isTwoFactorCodeComplete("123456")).toBe(true);
    expect(isTwoFactorCodeComplete("12345")).toBe(false);
    expect(isTwoFactorCodeComplete("12ab3456")).toBe(true);
  });

  it("detects complete already-sanitized codes by length only", () => {
    expect(isSanitizedTwoFactorCodeComplete("123456")).toBe(true);
    expect(isSanitizedTwoFactorCodeComplete("12345")).toBe(false);
    expect(isSanitizedTwoFactorCodeComplete("12ab3456")).toBe(false);
  });
});
