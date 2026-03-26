<template>
  <div class="friends-view">
    <h1 class="page-title">フレンド</h1>
    <div class="friends-header">
      <div class="tabs">
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'online' }"
          @click="activeTab = 'online'"
        >
          Online
        </button>
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'offline' }"
          @click="activeTab = 'offline'"
        >
          Offline
        </button>
      </div>
      <button
        type="button"
        class="btn-refresh"
        :disabled="!isLoggedIn || refreshLoading"
        :title="
          isLoggedIn ? 'フレンド一覧をAPIから再取得' : 'ログインが必要です'
        "
        @click="doRefresh"
      >
        {{ refreshLoading ? "更新中..." : "更新" }}
      </button>
    </div>
    <p v-if="!isLoggedIn" class="hint">
      フレンド一覧の更新にはログインが必要です。設定画面でログインしてください。
    </p>
    <div class="friends-section">
      <div class="friends-list">
        <div
          v-for="f in filteredFriends"
          :key="f.vrcUserId"
          class="friend-card"
          :class="{ active: selected?.vrcUserId === f.vrcUserId }"
          @click="selected = f"
        >
          <img
            v-if="friendThumbUrl(f)"
            class="friend-thumb"
            :src="friendThumbUrl(f)!"
            alt=""
            width="40"
            height="40"
          />
          <div v-else class="friend-thumb friend-thumb-placeholder" />
          <span class="friend-name">{{ f.displayName }}</span>
          <span class="friend-status">{{ f.status || "—" }}</span>
          <button
            type="button"
            class="btn-favorite"
            :class="{ on: f.isFavorite }"
            :title="f.isFavorite ? 'お気に入り解除' : 'お気に入り登録'"
            @click.stop="toggleFavorite(f)"
          >
            ★
          </button>
        </div>
        <p
          v-if="filteredFriends.length === 0 && !loading"
          class="empty-message"
        >
          {{
            activeTab === "online"
              ? "オンラインのフレンドはいません"
              : "オフラインのフレンドはいません"
          }}
        </p>
      </div>
      <div v-if="selected" class="friend-detail">
        <div class="detail-head">
          <img
            v-if="friendThumbUrl(selected)"
            class="detail-avatar"
            :src="friendThumbUrl(selected)!"
            alt=""
            width="96"
            height="96"
          />
          <h3>詳細</h3>
        </div>
        <dl class="detail-list">
          <dt>表示名</dt>
          <dd>{{ selected.displayName }}</dd>
          <dt>ユーザーID</dt>
          <dd class="mono">{{ selected.vrcUserId }}</dd>
          <template v-if="selected.username">
            <dt>ユーザー名</dt>
            <dd>{{ selected.username }}</dd>
          </template>
          <dt>ステータス</dt>
          <dd>{{ selected.status || "—" }}</dd>
          <template v-if="selected.statusDescription">
            <dt>ステータス説明</dt>
            <dd>{{ selected.statusDescription }}</dd>
          </template>
          <template v-if="selected.state">
            <dt>状態 (state)</dt>
            <dd>{{ selected.state }}</dd>
          </template>
          <template v-if="selected.bio">
            <dt>自己紹介</dt>
            <dd class="multiline">{{ selected.bio }}</dd>
          </template>
          <template v-if="jsonStringArray(selected.bioLinksJson).length">
            <dt>bio リンク</dt>
            <dd>
              <ul class="link-list">
                <li
                  v-for="(u, i) in jsonStringArray(selected.bioLinksJson)"
                  :key="i"
                >
                  <a :href="u" target="_blank" rel="noopener noreferrer">{{
                    u
                  }}</a>
                </li>
              </ul>
            </dd>
          </template>
          <template v-if="selected.location">
            <dt>ロケーション</dt>
            <dd class="mono wrap">{{ selected.location }}</dd>
          </template>
          <template v-if="selected.developerType">
            <dt>開発者種別</dt>
            <dd>{{ selected.developerType }}</dd>
          </template>
          <template v-if="selected.lastPlatform || selected.platform">
            <dt>プラットフォーム</dt>
            <dd>
              {{
                [selected.platform, selected.lastPlatform]
                  .filter(Boolean)
                  .join(" / ")
              }}
            </dd>
          </template>
          <template v-if="selected.lastLogin">
            <dt>最終ログイン</dt>
            <dd>{{ selected.lastLogin }}</dd>
          </template>
          <template v-if="selected.lastActivity">
            <dt>最終アクティビティ</dt>
            <dd>{{ selected.lastActivity }}</dd>
          </template>
          <template v-if="selected.lastMobile">
            <dt>最終モバイル</dt>
            <dd>{{ selected.lastMobile }}</dd>
          </template>
          <template v-if="jsonStringArray(selected.tagsJson).length">
            <dt>タグ</dt>
            <dd>
              <span
                v-for="tag in jsonStringArray(selected.tagsJson)"
                :key="tag"
                class="tag-chip"
                >{{ tag }}</span
              >
            </dd>
          </template>
          <template
            v-if="jsonStringArray(selected.currentAvatarTagsJson).length"
          >
            <dt>アバタータグ</dt>
            <dd>
              <span
                v-for="tag in jsonStringArray(selected.currentAvatarTagsJson)"
                :key="tag"
                class="tag-chip"
                >{{ tag }}</span
              >
            </dd>
          </template>
          <template v-if="selected.currentAvatarImageUrl">
            <dt>アバター画像 URL</dt>
            <dd>
              <a
                :href="selected.currentAvatarImageUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="wrap"
                >{{ selected.currentAvatarImageUrl }}</a
              >
            </dd>
          </template>
          <template v-if="selected.userIcon">
            <dt>ユーザーアイコン URL</dt>
            <dd>
              <a
                :href="selected.userIcon"
                target="_blank"
                rel="noopener noreferrer"
                class="wrap"
                >{{ selected.userIcon }}</a
              >
            </dd>
          </template>
          <template v-if="selected.imageUrl">
            <dt>imageUrl</dt>
            <dd>
              <a
                :href="selected.imageUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="wrap"
                >{{ selected.imageUrl }}</a
              >
            </dd>
          </template>
          <template v-if="selected.profilePicOverride">
            <dt>プロフィール画像 (上書き)</dt>
            <dd>
              <a
                :href="selected.profilePicOverride"
                target="_blank"
                rel="noopener noreferrer"
                class="wrap"
                >{{ selected.profilePicOverride }}</a
              >
            </dd>
          </template>
          <template v-if="selected.profilePicOverrideThumbnail">
            <dt>プロフィール画像サムネ</dt>
            <dd>
              <a
                :href="selected.profilePicOverrideThumbnail"
                target="_blank"
                rel="noopener noreferrer"
                class="wrap"
                >{{ selected.profilePicOverrideThumbnail }}</a
              >
            </dd>
          </template>
          <template v-if="selected.friendKey">
            <dt>friendKey</dt>
            <dd class="mono wrap">{{ selected.friendKey }}</dd>
          </template>
          <dt>キャッシュ更新</dt>
          <dd>{{ selected.lastUpdated }}</dd>
        </dl>
        <label class="favorite-toggle">
          <input
            v-model="selected.isFavorite"
            type="checkbox"
            @change="applyFavorite(selected)"
          />
          お気に入り
        </label>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { App } from "../wails/app";
import type { UserCacheDTO } from "../wails/app";

const activeTab = ref<"online" | "offline">("online");
const friends = ref<UserCacheDTO[]>([]);
const selected = ref<UserCacheDTO | null>(null);
const isLoggedIn = ref(false);
const loading = ref(true);
const refreshLoading = ref(false);

const filteredFriends = computed(() => {
  const list = friends.value;
  const isOffline = (s: string) => !s || s.toLowerCase() === "offline";
  if (activeTab.value === "online") {
    return list.filter((f) => !isOffline(f.status));
  }
  return list.filter((f) => isOffline(f.status));
});

function friendThumbUrl(f: UserCacheDTO): string | undefined {
  return (
    f.currentAvatarThumbnailImageUrl ||
    f.profilePicOverrideThumbnail ||
    f.userIcon ||
    f.imageUrl
  );
}

function jsonStringArray(raw: string | undefined): string[] {
  if (!raw?.trim()) return [];
  try {
    const v = JSON.parse(raw) as unknown;
    if (!Array.isArray(v)) return [];
    return v.filter((x): x is string => typeof x === "string");
  } catch {
    return [];
  }
}

onMounted(async () => {
  await loadFriends();
  isLoggedIn.value = await App.isLoggedIn();
});

async function loadFriends() {
  loading.value = true;
  try {
    friends.value = await App.friends();
  } finally {
    loading.value = false;
  }
}

async function doRefresh() {
  if (!isLoggedIn.value) return;
  refreshLoading.value = true;
  try {
    await App.refreshFriends();
    await loadFriends();
    selected.value =
      friends.value.find((f) => f.vrcUserId === selected.value?.vrcUserId) ??
      null;
  } finally {
    refreshLoading.value = false;
  }
}

async function toggleFavorite(f: UserCacheDTO) {
  const next = !f.isFavorite;
  try {
    await App.setFavorite(f.vrcUserId, next);
    f.isFavorite = next;
  } catch {
    // 失敗時は変化なし（一覧の星ボタンではまだ反映していない）
  }
}

async function applyFavorite(f: UserCacheDTO) {
  try {
    await App.setFavorite(f.vrcUserId, f.isFavorite);
  } catch {
    f.isFavorite = !f.isFavorite;
  }
}
</script>

<style scoped>
.page-title {
  margin: 0 0 1rem;
  font-size: 1.5rem;
}

.friends-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
}

.tabs {
  display: flex;
  gap: 0.25rem;
}

.tab-btn {
  padding: 0.4rem 1rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-secondary);
  cursor: pointer;
}

.tab-btn:hover {
  background: var(--bg-secondary);
  color: var(--text-primary);
}

.tab-btn.active {
  background: var(--accent);
  color: white;
  border-color: var(--accent);
}

.btn-refresh {
  padding: 0.4rem 1rem;
  background: var(--accent);
  color: white;
  border: none;
  border-radius: var(--radius);
  cursor: pointer;
}

.btn-refresh:hover:not(:disabled) {
  opacity: 0.9;
}

.btn-refresh:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.hint {
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin: 0 0 1rem;
}

.friends-section {
  display: flex;
  gap: 1.5rem;
}

.friends-list {
  width: 320px;
  max-height: 480px;
  overflow-y: auto;
}

.friend-card {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem;
  margin-bottom: 0.5rem;
  background: var(--bg-secondary);
  border-radius: var(--radius);
  cursor: pointer;
}

.friend-card:hover,
.friend-card.active {
  background: var(--bg-tertiary);
}

.friend-thumb {
  width: 40px;
  height: 40px;
  border-radius: var(--radius);
  object-fit: cover;
  flex-shrink: 0;
}

.friend-thumb-placeholder {
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
}

.friend-name {
  flex: 1;
  font-weight: 500;
  min-width: 0;
}

.friend-status {
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.btn-favorite {
  padding: 0.2rem 0.4rem;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 1rem;
}

.btn-favorite:hover {
  color: var(--accent);
}

.btn-favorite.on {
  color: var(--accent);
}

.empty-message {
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin: 1rem 0;
}

.friend-detail {
  flex: 1;
  padding: 1rem;
  background: var(--bg-secondary);
  border-radius: var(--radius);
  max-height: 560px;
  overflow-y: auto;
}

.detail-head {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 0.75rem;
}

.detail-head h3 {
  margin: 0;
  font-size: 1.1rem;
}

.detail-avatar {
  border-radius: var(--radius);
  object-fit: cover;
  flex-shrink: 0;
}

.detail-list {
  margin: 0 0 1rem;
}

.detail-list dt {
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin-top: 0.5rem;
}

.detail-list dt:first-child {
  margin-top: 0;
}

.detail-list dd {
  margin: 0.2rem 0 0;
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

.link-list {
  margin: 0;
  padding-left: 1.25rem;
}

.tag-chip {
  display: inline-block;
  margin: 0.15rem 0.35rem 0 0;
  padding: 0.1rem 0.45rem;
  font-size: 0.75rem;
  background: var(--bg-tertiary);
  border-radius: 999px;
  border: 1px solid var(--border);
}

.favorite-toggle {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.9rem;
  cursor: pointer;
}
</style>
