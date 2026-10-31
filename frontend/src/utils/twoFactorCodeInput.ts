export const TWO_FACTOR_CODE_LENGTH = 6;

/** Keeps up to six digits for TOTP / email OTP entry. */
export function sanitizeTwoFactorCodeInput(raw: string): string {
  return raw.replace(/\D/g, "").slice(0, TWO_FACTOR_CODE_LENGTH);
}

/** True when a value already sanitized by the 2FA field is exactly six digits. */
export function isSanitizedTwoFactorCodeComplete(code: string): boolean {
  return code.length === TWO_FACTOR_CODE_LENGTH && /^\d+$/.test(code);
}
