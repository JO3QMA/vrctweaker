import { describe, expect, it } from "vitest";
import {
  isSanitizedTwoFactorCodeComplete,
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

  it("detects complete already-sanitized codes as exactly six digits", () => {
    expect(isSanitizedTwoFactorCodeComplete("123456")).toBe(true);
    expect(isSanitizedTwoFactorCodeComplete("12345")).toBe(false);
    expect(isSanitizedTwoFactorCodeComplete("12ab34")).toBe(false);
    expect(isSanitizedTwoFactorCodeComplete(" 123456")).toBe(false);
  });
});
