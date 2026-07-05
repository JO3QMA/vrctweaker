import type { VrcStatusElementTagType } from "./vrcStatus";

export type UserTagTranslateFn = (
  key: string,
  params?: Record<string, string>,
) => string;

/** vrchat.community/tags/user に掲載のユーザータグ（language_* 除く） */
export const KNOWN_USER_TAG_IDS = new Set([
  "admin_avatar_access",
  "admin_can_grant_licenses",
  "admin_canny_access",
  "admin_lock_tags",
  "admin_lock_level",
  "admin_moderator",
  "admin_official_thumbnail",
  "admin_scripting_access",
  "admin_world_access",
  "show_social_rank",
  "show_mod_tag",
  "system_avatar_access",
  "system_early_adopter",
  "system_feedback_access",
  "system_generated",
  "system_probable_troll",
  "system_supporter",
  "system_legend",
  "system_scripting_access",
  "system_troll",
  "system_trust_basic",
  "system_trust_known",
  "system_trust_trusted",
  "system_trust_veteran",
  "system_trust_intermediate",
  "system_trust_advanced",
  "system_trust_legend",
  "system_world_access",
  "system_vital_test_user_do_not_delete",
]);

/** vrchat.community/tags の Language Tags */
export const KNOWN_LANGUAGE_TAG_IDS = new Set([
  "language_afr",
  "language_ara",
  "language_ase",
  "language_asf",
  "language_ben",
  "language_bfi",
  "language_bul",
  "language_ces",
  "language_cmn",
  "language_cym",
  "language_dan",
  "language_deu",
  "language_dse",
  "language_ell",
  "language_eng",
  "language_epo",
  "language_est",
  "language_fil",
  "language_fin",
  "language_fra",
  "language_fsl",
  "language_gla",
  "language_gle",
  "language_gsg",
  "language_heb",
  "language_hin",
  "language_hmn",
  "language_hrv",
  "language_hun",
  "language_hye",
  "language_ind",
  "language_isl",
  "language_ita",
  "language_jpn",
  "language_jsl",
  "language_kor",
  "language_kvk",
  "language_lav",
  "language_lit",
  "language_ltz",
  "language_mar",
  "language_mkd",
  "language_mlt",
  "language_mri",
  "language_msa",
  "language_nld",
  "language_nor",
  "language_nzs",
  "language_pol",
  "language_por",
  "language_ron",
  "language_rus",
  "language_sco",
  "language_slk",
  "language_slv",
  "language_spa",
  "language_swe",
  "language_tel",
  "language_tha",
  "language_tok",
  "language_tur",
  "language_tws",
  "language_ukr",
  "language_vie",
  "language_wuu",
  "language_yue",
  "language_zho",
  "language_zxx",
]);

export const DEPRECATED_USER_TAG_IDS = new Set([
  "show_social_rank",
  "admin_scripting_access",
  "system_scripting_access",
  "system_legend",
  "system_trust_intermediate",
  "system_trust_advanced",
  "system_trust_legend",
  "system_generated",
  "system_vital_test_user_do_not_delete",
]);

const LANGUAGE_USER_TAG_PREFIX = "language_";

export function isLanguageUserTag(tag: string): boolean {
  return tag.startsWith(LANGUAGE_USER_TAG_PREFIX);
}

export function isKnownUserTag(tag: string): boolean {
  return KNOWN_USER_TAG_IDS.has(tag) || KNOWN_LANGUAGE_TAG_IDS.has(tag);
}

export function isDeprecatedUserTag(tag: string): boolean {
  return DEPRECATED_USER_TAG_IDS.has(tag);
}

function userTagI18nKey(tag: string, field: "label" | "description"): string {
  return `userDetail.userTags.${tag}.${field}`;
}

function userTagMetaI18nKey(
  field: "deprecated" | "unknown" | "tag_id",
): string {
  return `userDetail.userTags.${field}`;
}

export function userTagElementType(tag: string): VrcStatusElementTagType {
  switch (tag.trim()) {
    case "system_trust_basic":
      return "info";
    case "system_trust_known":
      return "success";
    case "system_trust_trusted":
      return "warning";
    case "system_trust_veteran":
    case "system_trust_legend":
      return "primary";
    case "admin_moderator":
    case "show_mod_tag":
      return "danger";
    case "system_troll":
    case "system_probable_troll":
      return "danger";
    case "system_supporter":
    case "system_early_adopter":
      return "success";
    default:
      return "info";
  }
}

export interface UserTagDisplay {
  label: string;
  tooltip: string;
  isKnown: boolean;
  deprecated: boolean;
}

export function resolveUserTagDisplay(
  tag: string,
  t: UserTagTranslateFn,
): UserTagDisplay {
  const trimmed = tag.trim();
  const deprecated = isDeprecatedUserTag(trimmed);
  const tagIdLine = `${t(userTagMetaI18nKey("tag_id"))}: ${trimmed}`;

  if (isKnownUserTag(trimmed)) {
    const label = t(userTagI18nKey(trimmed, "label"));
    const description = t(userTagI18nKey(trimmed, "description"));
    const deprecatedLine = deprecated
      ? `(${t(userTagMetaI18nKey("deprecated"))})`
      : "";
    const tooltip = [description, deprecatedLine, tagIdLine]
      .filter(Boolean)
      .join("\n");

    return { label, tooltip, isKnown: true, deprecated };
  }

  const unknown = t(userTagMetaI18nKey("unknown"));
  return {
    label: trimmed,
    tooltip: `${unknown}\n${tagIdLine}`,
    isKnown: false,
    deprecated: false,
  };
}
