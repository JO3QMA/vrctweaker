import { describe, expect, it } from "vitest";
import {
  isLoginTwoFactorRequired,
  LOGIN_TWO_FACTOR_REQUIRED_MARKER,
} from "../vrchatLoginResult";

describe("isLoginTwoFactorRequired", () => {
  it("matches the Go sentinel", () => {
    expect(isLoginTwoFactorRequired(LOGIN_TWO_FACTOR_REQUIRED_MARKER)).toBe(
      true,
    );
  });

  it("rejects other errors", () => {
    expect(isLoginTwoFactorRequired("invalid credentials")).toBe(false);
  });
});
