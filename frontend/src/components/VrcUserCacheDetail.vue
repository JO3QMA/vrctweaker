<template>
  <div v-if="selected" ref="detailRoot" class="friend-detail">
    <el-card
      class="friend-detail-card"
      shadow="never"
      :body-style="{ padding: 0 }"
    >
      <div
        class="detail-sticky-header"
        :class="{ visible: stickyHeaderVisible }"
      >
        <img
          v-if="avatarSrc"
          class="detail-sticky-avatar"
          :src="avatarSrc"
          alt=""
          width="32"
          height="32"
        />
        <div
          v-else
          class="detail-sticky-avatar detail-sticky-avatar--placeholder"
        />
        <div class="detail-sticky-main">
          <p class="detail-sticky-name">{{ selected.displayName }}</p>
          <VrcStatusTag :status="selected.status" />
        </div>
      </div>
      <div class="profile-hero">
        <div class="profile-banner" aria-hidden="true">
          <img
            v-if="bannerSrc"
            class="profile-banner-img"
            :src="bannerSrc"
            alt=""
          />
          <div v-else class="profile-banner-fallback" />
        </div>
        <div class="profile-toolbar">
          <div class="profile-avatar-wrap">
            <img
              v-if="avatarSrc"
              class="profile-avatar"
              :src="avatarSrc"
              alt=""
              width="88"
              height="88"
            />
            <div v-else class="profile-avatar profile-avatar--placeholder" />
          </div>
          <div class="profile-toolbar-actions">
            <el-checkbox
              :model-value="selected.isFavorite"
              @update:model-value="onFavoriteUpdate"
            >
              お気に入り
            </el-checkbox>
          </div>
        </div>
      </div>

      <div class="profile-body">
        <div ref="nameAnchor" class="profile-name-row">
          <h2 class="profile-display-name">{{ selected.displayName }}</h2>
          <el-button
            link
            type="primary"
            title="表示名をコピー"
            aria-label="表示名をコピー"
            data-testid="friend-copy-display-name"
            class="profile-copy-name"
            @click="copyDisplayName(selected.displayName)"
          >
            <el-icon><CopyDocument /></el-icon>
          </el-button>
        </div>
        <p v-if="selected.username" class="profile-handle">
          @{{ selected.username }}
        </p>
        <div class="profile-status-row">
          <VrcStatusTag :status="selected.status" />
          <span v-if="selected.statusDescription" class="profile-status-desc">{{
            selected.statusDescription
          }}</span>
        </div>
        <p v-if="selected.bio" class="profile-bio multiline">
          {{ selected.bio }}
        </p>
        <ul
          v-if="jsonStringArray(selected.bioLinksJson).length"
          class="profile-bio-links"
        >
          <li v-for="(u, i) in jsonStringArray(selected.bioLinksJson)" :key="i">
            <a :href="u" target="_blank" rel="noopener noreferrer">{{ u }}</a>
          </li>
        </ul>
      </div>

      <div class="profile-details-wrap">
        <el-tabs v-model="detailTab" class="profile-detail-tabs">
          <el-tab-pane label="詳細" name="detail">
            <el-descriptions :column="1" border size="small">
              <el-descriptions-item v-if="selected.state" label="状態 (state)">
                {{ selected.state }}
              </el-descriptions-item>
              <el-descriptions-item v-if="locationLabel" label="ロケーション">
                <span class="mono wrap">{{ locationLabel }}</span>
              </el-descriptions-item>
              <el-descriptions-item
                v-if="selected.developerType"
                label="開発者種別"
              >
                {{ selected.developerType }}
              </el-descriptions-item>
              <el-descriptions-item
                v-if="selected.lastPlatform || selected.platform"
                label="プラットフォーム"
              >
                {{
                  [selected.platform, selected.lastPlatform]
                    .filter(Boolean)
                    .join(" / ")
                }}
              </el-descriptions-item>
              <el-descriptions-item
                v-if="selected.lastLogin"
                label="最終ログイン"
              >
                {{ selected.lastLogin }}
              </el-descriptions-item>
              <el-descriptions-item
                v-if="selected.lastActivity"
                label="最終アクティビティ"
              >
                {{ selected.lastActivity }}
              </el-descriptions-item>
              <el-descriptions-item
                v-if="selected.lastMobile"
                label="最終モバイル"
              >
                {{ selected.lastMobile }}
              </el-descriptions-item>
              <el-descriptions-item
                v-if="jsonStringArray(selected.tagsJson).length"
                label="タグ"
              >
                <el-tag
                  v-for="tag in jsonStringArray(selected.tagsJson)"
                  :key="tag"
                  size="small"
                  class="tag-chip"
                >
                  {{ tag }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item
                v-if="jsonStringArray(selected.currentAvatarTagsJson).length"
                label="アバタータグ"
              >
                <el-tag
                  v-for="tag in jsonStringArray(selected.currentAvatarTagsJson)"
                  :key="tag"
                  size="small"
                  class="tag-chip"
                >
                  {{ tag }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item
                v-if="selected.currentAvatarImageUrl"
                label="アバター画像 URL"
              >
                <a
                  :href="selected.currentAvatarImageUrl"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="wrap"
                  >{{ selected.currentAvatarImageUrl }}</a
                >
              </el-descriptions-item>
              <el-descriptions-item
                v-if="selected.userIcon"
                label="ユーザーアイコン URL"
              >
                <a
                  :href="selected.userIcon"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="wrap"
                  >{{ selected.userIcon }}</a
                >
              </el-descriptions-item>
              <el-descriptions-item v-if="selected.imageUrl" label="imageUrl">
                <a
                  :href="selected.imageUrl"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="wrap"
                  >{{ selected.imageUrl }}</a
                >
              </el-descriptions-item>
              <el-descriptions-item
                v-if="selected.profilePicOverride"
                label="プロフィール画像 (上書き)"
              >
                <a
                  :href="selected.profilePicOverride"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="wrap"
                  >{{ selected.profilePicOverride }}</a
                >
              </el-descriptions-item>
              <el-descriptions-item
                v-if="selected.profilePicOverrideThumbnail"
                label="プロフィール画像サムネ"
              >
                <a
                  :href="selected.profilePicOverrideThumbnail"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="wrap"
                  >{{ selected.profilePicOverrideThumbnail }}</a
                >
              </el-descriptions-item>
              <el-descriptions-item v-if="selected.friendKey" label="friendKey">
                <span class="mono wrap">{{ selected.friendKey }}</span>
              </el-descriptions-item>
              <el-descriptions-item label="キャッシュ更新">
                {{ selected.lastUpdated }}
              </el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>
          <el-tab-pane label="遭遇履歴" name="encounters" lazy>
            <EncounterHistoryList
              mode="user"
              :user-id="selected.vrcUserId"
              hide-display-name-column
            />
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import EncounterHistoryList from "./EncounterHistoryList.vue";
import VrcStatusTag from "./VrcStatusTag.vue";
import type { UserCacheDTO } from "../wails/app";
import {
  copyDisplayName,
  friendDetailStickyHeaderVisible,
  friendLocationLabel,
  friendProfileBannerUrl,
  friendThumbUrl,
  jsonStringArray,
} from "../utils/vrcUserCacheDisplay";

const props = defineProps<{
  selected: UserCacheDTO | null;
}>();

const detailTab = ref<"detail" | "encounters">("detail");

const detailRoot = ref<HTMLElement | null>(null);
const nameAnchor = ref<HTMLElement | null>(null);
const stickyHeaderVisible = ref(false);

let cardBodyScrollEl: HTMLElement | null = null;

function getCardBodyEl(): HTMLElement | null {
  return detailRoot.value?.querySelector(".el-card__body") ?? null;
}

function detachCardBodyScroll() {
  if (cardBodyScrollEl) {
    cardBodyScrollEl.removeEventListener("scroll", onScroll);
    cardBodyScrollEl = null;
  }
}

function attachCardBodyScroll() {
  detachCardBodyScroll();
  const body = getCardBodyEl();
  if (!body) return;
  cardBodyScrollEl = body;
  body.addEventListener("scroll", onScroll, { passive: true });
}

const emit = defineEmits<{
  favoriteChange: [user: UserCacheDTO, isFavorite: boolean];
}>();

const bannerSrc = computed(() =>
  props.selected ? friendProfileBannerUrl(props.selected) : undefined,
);

const avatarSrc = computed(() =>
  props.selected ? friendThumbUrl(props.selected) : undefined,
);

const locationLabel = computed(() =>
  props.selected ? friendLocationLabel(props.selected.location) : "",
);

function onFavoriteUpdate(val: boolean | string | number | undefined) {
  const f = props.selected;
  if (!f) return;
  emit("favoriteChange", f, Boolean(val));
}

function updateStickyHeaderVisibility() {
  const body = getCardBodyEl();
  const anchor = nameAnchor.value;
  if (!body || !anchor) return;
  const bodyRect = body.getBoundingClientRect();
  const anchorRect = anchor.getBoundingClientRect();
  stickyHeaderVisible.value = friendDetailStickyHeaderVisible({
    scrollTop: body.scrollTop,
    anchorTopViewport: anchorRect.top,
    bodyTopViewport: bodyRect.top,
  });
}

function onScroll() {
  updateStickyHeaderVisibility();
}

watch(detailTab, () => {
  void nextTick(() => updateStickyHeaderVisibility());
});

watch(
  () => props.selected?.vrcUserId,
  () => {
    detailTab.value = "detail";
  },
);

watch(
  () => props.selected,
  async (sel, prev) => {
    detachCardBodyScroll();
    stickyHeaderVisible.value = false;
    if (!sel) return;
    await nextTick();
    attachCardBodyScroll();
    const body = getCardBodyEl();
    if (body && (!prev || prev.vrcUserId !== sel.vrcUserId)) {
      body.scrollTop = 0;
    }
    updateStickyHeaderVisibility();
  },
  { immediate: true },
);

onMounted(() => {
  void nextTick(() => updateStickyHeaderVisibility());
});

onUnmounted(() => {
  detachCardBodyScroll();
});
</script>

<style scoped>
.friend-detail {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.friend-detail-card {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  background: var(--bg-secondary) !important;
  border-color: var(--border) !important;
}

.friend-detail-card :deep(.el-card__body) {
  flex: 1;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
}

.detail-sticky-header {
  position: sticky;
  top: 0;
  z-index: 20;
  display: flex;
  align-items: center;
  gap: 0.55rem;
  box-sizing: border-box;
  max-height: 0;
  min-height: 0;
  padding: 0 0.75rem;
  overflow: hidden;
  border-bottom: 0 solid var(--border);
  background: color-mix(in srgb, var(--bg-secondary) 92%, #000);
  backdrop-filter: blur(6px);
  opacity: 0;
  transform: translateY(-0.35rem);
  pointer-events: none;
  transition:
    max-height 0.2s ease,
    opacity 0.15s ease,
    transform 0.15s ease,
    padding 0.2s ease,
    border-bottom-width 0.2s ease;
}

.detail-sticky-header.visible {
  max-height: 3.5rem;
  padding: 0.55rem 0.75rem;
  border-bottom-width: 1px;
  opacity: 1;
  transform: translateY(0);
  pointer-events: auto;
}

.detail-sticky-avatar {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  object-fit: cover;
  flex-shrink: 0;
}

.detail-sticky-avatar--placeholder {
  background: var(--el-fill-color-light);
  border: 1px solid var(--border);
}

.detail-sticky-main {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 0.45rem;
}

.detail-sticky-name {
  margin: 0;
  max-width: 22rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.92rem;
  font-weight: 700;
  color: var(--el-text-color-primary);
}

.profile-hero {
  position: relative;
}

.profile-banner {
  position: relative;
  height: 120px;
  overflow: hidden;
  background: linear-gradient(
    145deg,
    var(--el-fill-color-dark) 0%,
    var(--el-fill-color-darker, #1a1a1a) 100%
  );
}

.profile-banner-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.profile-banner-fallback {
  width: 100%;
  height: 100%;
  background: linear-gradient(
    160deg,
    color-mix(in srgb, var(--el-color-primary) 35%, transparent),
    var(--el-fill-color-dark)
  );
}

.profile-toolbar {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0 1rem;
  margin-top: -44px;
  position: relative;
  z-index: 1;
}

.profile-avatar-wrap {
  flex-shrink: 0;
}

.profile-avatar {
  display: block;
  width: 88px;
  height: 88px;
  border-radius: 12px;
  object-fit: cover;
  border: 3px solid var(--bg-secondary);
  box-sizing: border-box;
  background: var(--bg-secondary);
}

.profile-avatar--placeholder {
  background: var(--el-fill-color-light);
  border-style: solid;
}

.profile-toolbar-actions {
  padding-bottom: 0.35rem;
}

.profile-body {
  padding: 0.65rem 1rem 1rem;
}

.profile-name-row {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  flex-wrap: wrap;
}

.profile-display-name {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 700;
  line-height: 1.25;
  color: var(--el-text-color-primary);
}

.profile-copy-name {
  flex-shrink: 0;
}

.profile-handle {
  margin: 0.15rem 0 0;
  font-size: 0.9rem;
  color: var(--el-text-color-secondary);
}

.profile-status-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem 0.5rem;
  margin-top: 0.5rem;
}

.profile-status-desc {
  font-size: 0.85rem;
  color: var(--el-text-color-regular);
}

.profile-bio {
  margin: 0.75rem 0 0;
  font-size: 0.9rem;
  line-height: 1.45;
  color: var(--el-text-color-primary);
}

.profile-bio-links {
  margin: 0.5rem 0 0;
  padding-left: 1.1rem;
  font-size: 0.9rem;
}

.profile-bio-links a {
  color: var(--el-color-primary);
  word-break: break-all;
}

.profile-details-wrap {
  padding: 0 1rem 1rem;
  min-height: 0;
}

.profile-detail-tabs {
  min-height: 0;
}

.profile-detail-tabs :deep(.el-tabs__header) {
  margin-bottom: 0.65rem;
}

.profile-detail-tabs :deep(.el-tabs__content),
.profile-detail-tabs :deep(.el-tab-pane) {
  min-height: 0;
}

.tag-chip {
  margin: 0.15rem 0.2rem 0 0;
}

.mono {
  font-family: ui-monospace, monospace;
  font-size: 0.85rem;
}

.wrap {
  word-break: break-all;
}

.multiline {
  white-space: pre-wrap;
}
</style>
