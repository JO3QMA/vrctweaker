import { isTwoFactorCodeComplete } from "./twoFactorCodeInput";

/** True when login must stay disabled because 2FA is incomplete. */
export function isLoginBlockedByTwoFactorCode(
  code: string,
  twoFactorRequired: boolean,
): boolean {
  const trimmed = code.trim();
  if (twoFactorRequired) {
    return !isTwoFactorCodeComplete(trimmed);
  }
  if (trimmed.length === 0) {
    return false;
  }
  return !isTwoFactorCodeComplete(trimmed);
}
