/** Stable substring from Go vrchatapi.ErrTwoFactorRequired. */
export const LOGIN_TWO_FACTOR_REQUIRED_MARKER =
  "two-factor authentication required";

export function isLoginTwoFactorRequired(error: string | undefined): boolean {
  if (!error) return false;
  return error.trim().toLowerCase() === LOGIN_TWO_FACTOR_REQUIRED_MARKER;
}
