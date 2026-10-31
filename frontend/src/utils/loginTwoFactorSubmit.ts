import { isSanitizedTwoFactorCodeComplete } from "./twoFactorCodeInput";

/**
 * True when login must stay disabled because 2FA is incomplete.
 * `sanitizedCode` must be the sanitized `LoginTwoFactorCodeField` model value.
 */
export function isLoginBlockedByTwoFactorCode(
  sanitizedCode: string,
  twoFactorRequired: boolean,
): boolean {
  if (sanitizedCode.length === 0) {
    return twoFactorRequired;
  }
  return !isSanitizedTwoFactorCodeComplete(sanitizedCode);
}
