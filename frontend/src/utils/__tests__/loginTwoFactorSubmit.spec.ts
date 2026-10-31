import { describe, expect, it } from "vitest";
import { isLoginBlockedByTwoFactorCode } from "../loginTwoFactorSubmit";

describe("isLoginBlockedByTwoFactorCode", () => {
  it("requires six digits when 2FA is required", () => {
    expect(isLoginBlockedByTwoFactorCode("", true)).toBe(true);
    expect(isLoginBlockedByTwoFactorCode("12345", true)).toBe(true);
    expect(isLoginBlockedByTwoFactorCode("123456", true)).toBe(false);
  });

  it("allows empty optional 2FA but blocks partial entry", () => {
    expect(isLoginBlockedByTwoFactorCode("", false)).toBe(false);
    expect(isLoginBlockedByTwoFactorCode("12", false)).toBe(true);
    expect(isLoginBlockedByTwoFactorCode("123456", false)).toBe(false);
  });

  it("requires an already-sanitized six-digit value", () => {
    expect(isLoginBlockedByTwoFactorCode("12ab34", true)).toBe(true);
    expect(isLoginBlockedByTwoFactorCode(" 123456", true)).toBe(true);
  });
});
