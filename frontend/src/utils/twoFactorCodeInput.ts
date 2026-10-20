export const TWO_FACTOR_CODE_LENGTH = 6;

/** Keeps up to six digits for TOTP / email OTP entry. */
export function sanitizeTwoFactorCodeInput(raw: string): string {
  return raw.replace(/\D/g, "").slice(0, TWO_FACTOR_CODE_LENGTH);
}

/** Strips separators then digits-only (for paste). */
export function sanitizeTwoFactorCodePaste(raw: string): string {
  return sanitizeTwoFactorCodeInput(raw.replace(/[\s-]/g, ""));
}

export function isTwoFactorCodeComplete(code: string): boolean {
  return sanitizeTwoFactorCodeInput(code).length === TWO_FACTOR_CODE_LENGTH;
}
