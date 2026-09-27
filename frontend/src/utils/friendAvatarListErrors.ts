import type { ComposerTranslation } from "vue-i18n";

/** Maps backend errors for friend avatar list fetch to safe i18n messages. */
export function friendAvatarListFetchErrorMessage(
  _err: unknown,
  t: ComposerTranslation,
): string {
  return t("friendAvatars.fetchFailedGeneric");
}
