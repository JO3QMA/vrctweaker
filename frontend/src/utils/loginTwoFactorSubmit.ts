import { isSanitizedTwoFactorCodeComplete } from "./twoFactorCodeInput";

/**
 * True when login must stay disabled because 2FA is incomplete.
 * `code` comes from LoginTwoFactorCodeField, so it is already sanitized.
 */
export function isLoginBlockedByTwoFactorCode(
  code: string,
  twoFactorRequired: boolean,
): boolean {
  if (twoFactorRequired) {
    return !isSanitizedTwoFactorCodeComplete(code);
  }
  return code.length > 0 && !isSanitizedTwoFactorCodeComplete(code);
}
