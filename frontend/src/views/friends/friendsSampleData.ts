import type { UserCacheDTO } from "../../wails/app";

/**
 * Storybook / プレビュー用フレンド一覧。
 *
 * フィールド構成とステータス種別の出方は、実DB
 * `docs/ai_dlc/vrchat-tweaker.db` の `users_cache`（user_kind = friend）を
 * 集計・サンプル抽出して基準にしている（例: active / offline が多く、busy は稀、など）。
 * vrcUserId・表示名・bio・各種 URL・インスタンス表記はプライバシーと再配布のため架空値。
 */
export function sampleFriend(
  partial: Partial<UserCacheDTO> &
    Pick<UserCacheDTO, "vrcUserId" | "displayName" | "status">,
): UserCacheDTO {
  return {
    isFavorite: false,
    lastUpdated: "2026-03-27T13:10:00+09:00",
    ...partial,
  } as UserCacheDTO;
}

/** 実DBに近い tags_json の一例（system_* / language_* が混ざる形） */
const tagsSampleA =
  '["language_jpn","system_world_access","system_trust_basic","system_avatar_access","system_feedback_access"]';

const tagsSampleB =
  '["show_social_rank","language_jpn","system_world_access","system_trust_basic","language_kor"]';

const thumb = (n: number) =>
  `https://cdn.example.invalid/vrctweaker-story/avatar/${n}/256.jpg`;
const fileImage = (n: number) =>
  `https://cdn.example.invalid/vrctweaker-story/file/${n}/image`;

export const sampleFriendsList: UserCacheDTO[] = [
  sampleFriend({
    vrcUserId: "usr_a1111111-1111-4111-8111-111111111101",
    displayName: "サンプル・ジョインミー",
    status: "join me",
    isFavorite: true,
    username: "sample_join_user",
    statusDescription: "イベントワールド（ダミー）",
    bio: "VRChat 中心に活動しています。\n気軽に Join してください（ダミー文面）。",
    bioLinksJson: JSON.stringify([
      "https://example.invalid/social/profile-join",
      "https://example.invalid/link/booth-dummy",
    ]),
    location:
      "wrld_aaaaaaaa-bbbb-4ccc-dddd-eeeeeeeeeeee:12345~group(grp_00000000-0000-4000-8000-000000000001)~groupAccessType(public)",
    developerType: "none",
    platform: "standalonewindows",
    lastPlatform: "standalonewindows",
    lastLogin: "2026-03-27T08:10:57.917Z",
    lastActivity: "2026-03-27T13:09:55.308Z",
    firstSeenAt: "2026-03-26T10:00:00+09:00",
    lastContactAt: "2026-03-27T23:14:31+09:00",
    currentAvatarThumbnailImageUrl: thumb(1),
    currentAvatarImageUrl: fileImage(1),
    imageUrl: thumb(1),
    tagsJson: tagsSampleA,
  }),
  sampleFriend({
    vrcUserId: "usr_a1111111-1111-4111-8111-111111111102",
    displayName: "サンプル・アスクミー",
    status: "ask me",
    statusDescription: "まったり",
    bio: "はじめまして。ワールド巡りが好きです（ダミー）。\n仲良くしてくれると嬉しいです。",
    bioLinksJson: JSON.stringify([
      "https://example.invalid/social/profile-ask",
      "https://example.invalid/shop/dummy-booth",
    ]),
    location: "private",
    developerType: "none",
    platform: "standalonewindows",
    lastPlatform: "standalonewindows",
    lastLogin: "2026-03-27T12:34:32.708Z",
    lastActivity: "2026-03-27T13:09:13.698Z",
    firstSeenAt: "2026-03-20T18:00:00+09:00",
    lastContactAt: "2026-03-27T21:00:00+09:00",
    currentAvatarThumbnailImageUrl: thumb(2),
    currentAvatarImageUrl: fileImage(2),
    imageUrl: thumb(2),
    tagsJson: tagsSampleA,
  }),
  sampleFriend({
    vrcUserId: "usr_a1111111-1111-4111-8111-111111111103",
    displayName: "サンプル・ビジー",
    status: "busy",
    statusDescription: "作業中（ダミー）",
    bio: "平日は返信遅めかもしれません（ダミー）。\nイベントは週末によくいます。",
    bioLinksJson: JSON.stringify([
      "https://example.invalid/social/profile-busy",
    ]),
    location: "offline",
    developerType: "none",
    platform: "web",
    lastPlatform: "standalonewindows",
    lastLogin: "2026-03-26T13:02:36.489Z",
    lastActivity: "2026-03-26T16:07:09.576Z",
    firstSeenAt: "2026-03-01T12:00:00+09:00",
    lastContactAt: "2026-03-26T20:00:00+09:00",
    currentAvatarThumbnailImageUrl: thumb(3),
    currentAvatarImageUrl: fileImage(3),
    imageUrl: thumb(3),
    userIcon: fileImage(30),
    tagsJson: tagsSampleA,
  }),
  sampleFriend({
    vrcUserId: "usr_a1111111-1111-4111-8111-111111111104",
    displayName: "サンプル・アクティブ",
    status: "active",
    statusDescription: "․․․",
    bio: "言語学習ワールドによくいます（ダミー）。\nフレンド申請は一言あいさつ付きでお願いします。",
    location:
      "wrld_beddab1e-fee1-cafe-f00d-ca7c0dd1eca7:53938~hidden(usr_b2222222-2222-4222-8222-222222222202)~region(jp)",
    developerType: "none",
    platform: "standalonewindows",
    lastPlatform: "standalonewindows",
    lastLogin: "2026-03-26T13:17:45.695Z",
    lastActivity: "2026-03-27T13:09:04.013Z",
    firstSeenAt: "2026-02-15T09:00:00+09:00",
    lastContactAt: "2026-03-27T12:00:00+09:00",
    currentAvatarThumbnailImageUrl: thumb(4),
    currentAvatarImageUrl: fileImage(4),
    imageUrl: fileImage(40),
    profilePicOverride: fileImage(40),
    profilePicOverrideThumbnail: thumb(40),
    userIcon: fileImage(41),
    tagsJson: tagsSampleB,
    currentAvatarTagsJson: JSON.stringify(["author_permitted"]),
  }),
  sampleFriend({
    vrcUserId: "usr_a1111111-1111-4111-8111-111111111105",
    displayName: "サンプル・オフライン",
    status: "offline",
    bio: "たまにインします（ダミー）。",
    bioLinksJson: JSON.stringify([
      "https://example.invalid/social/profile-off",
    ]),
    location: "offline",
    developerType: "none",
    platform: "standalonewindows",
    lastPlatform: "standalonewindows",
    lastLogin: "2026-03-26T13:32:00.433Z",
    lastActivity: "2026-03-26T15:48:41.694Z",
    firstSeenAt: "2026-01-10T20:00:00+09:00",
    lastContactAt: "2026-03-26T18:00:00+09:00",
    currentAvatarThumbnailImageUrl: thumb(5),
    currentAvatarImageUrl: fileImage(5),
    imageUrl: fileImage(50),
    profilePicOverride: fileImage(50),
    profilePicOverrideThumbnail: thumb(50),
    userIcon: fileImage(51),
    tagsJson: tagsSampleA,
  }),
];
