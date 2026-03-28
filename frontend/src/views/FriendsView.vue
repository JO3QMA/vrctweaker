<template>
  <div class="friends-view">
    <h1 class="page-title">フレンド</h1>
    <div class="friends-toolbar">
      <div class="friends-header">
        <div
          class="filter-mode"
          role="group"
          aria-label="フレンド一覧: Online または Offline"
        >
          <span :class="['mode-label', { active: !showOfflineList }]"
            >Online</span
          >
          <el-switch
            v-model="showOfflineList"
            data-testid="friends-filter-mode"
            aria-label="Offline 一覧を表示する（オフのときは Online）"
          />
          <span :class="['mode-label', { active: showOfflineList }]"
            >Offline</span
          >
        </div>
        <el-button
          type="primary"
          :disabled="!isLoggedIn || refreshLoading"
          :loading="refreshLoading"
          :title="
            isLoggedIn ? 'フレンド一覧をAPIから再取得' : 'ログインが必要です'
          "
          @click="doRefresh"
        >
          {{ refreshLoading ? "更新中..." : "更新" }}
        </el-button>
      </div>
      <el-input
        v-model.trim="displayNameQuery"
        type="search"
        placeholder="表示名で検索"
        data-testid="friends-search-display-name"
        clearable
        class="friends-search-input"
        autocomplete="off"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
    </div>
    <el-alert
      v-if="!isLoggedIn"
      title="フレンド一覧の更新にはログインが必要です。設定画面でログインしてください。"
      type="info"
      :closable="false"
      show-icon
      class="login-hint"
    />
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
          <el-tag
            size="small"
            :type="statusTagType(f.status)"
            class="friend-status-tag"
          >
            {{ f.status || "—" }}
          </el-tag>
          <el-button
            link
            :type="f.isFavorite ? 'primary' : 'info'"
            :title="f.isFavorite ? 'お気に入り解除' : 'お気に入り登録'"
            class="btn-favorite"
            @click.stop="toggleFavorite(f)"
          >
            ★
          </el-button>
        </div>
        <p
          v-if="filteredFriends.length === 0 && !loading"
          class="empty-message"
        >
          {{ emptyListMessage }}
        </p>
      </div>
      <el-card v-if="selected" class="friend-detail" shadow="never">
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
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="表示名">
            <div class="detail-display-name">
              <span>{{ selected.displayName }}</span>
              <el-button
                link
                type="primary"
                title="表示名をコピー"
                aria-label="表示名をコピー"
                data-testid="friend-copy-display-name"
                @click="copyDisplayName(selected.displayName)"
              >
                <el-icon><CopyDocument /></el-icon>
              </el-button>
            </div>
          </el-descriptions-item>
          <el-descriptions-item v-if="selected.username" label="ユーザー名">
            {{ selected.username }}
          </el-descriptions-item>
          <el-descriptions-item label="ステータス">
            {{ selected.status || "—" }}
          </el-descriptions-item>
          <el-descriptions-item
            v-if="selected.statusDescription"
            label="ステータス説明"
          >
            {{ selected.statusDescription }}
          </el-descriptions-item>
          <el-descriptions-item v-if="selected.state" label="状態 (state)">
            {{ selected.state }}
          </el-descriptions-item>
          <el-descriptions-item v-if="selected.bio" label="自己紹介">
            <span class="multiline">{{ selected.bio }}</span>
          </el-descriptions-item>
          <el-descriptions-item
            v-if="jsonStringArray(selected.bioLinksJson).length"
            label="bio リンク"
          >
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
          </el-descriptions-item>
          <el-descriptions-item v-if="selected.location" label="ロケーション">
            <span class="mono wrap">{{ selected.location }}</span>
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
          <el-descriptions-item v-if="selected.lastLogin" label="最終ログイン">
            {{ selected.lastLogin }}
          </el-descriptions-item>
          <el-descriptions-item
            v-if="selected.lastActivity"
            label="最終アクティビティ"
          >
            {{ selected.lastActivity }}
          </el-descriptions-item>
          <el-descriptions-item v-if="selected.lastMobile" label="最終モバイル">
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
        <div class="favorite-toggle">
          <el-checkbox
            v-model="selected.isFavorite"
            @change="applyFavorite(selected!)"
          >
            お気に入り
          </el-checkbox>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { App } from "../wails/app";
import type { UserCacheDTO } from "../wails/app";

/** false: オンラインのみ / true: オフラインのみ */
const showOfflineList = ref(false);
const displayNameQuery = ref("");
const friends = ref<UserCacheDTO[]>([]);
const selected = ref<UserCacheDTO | null>(null);
const isLoggedIn = ref(false);
const loading = ref(true);
const refreshLoading = ref(false);

function friendIsOffline(status: string): boolean {
  return !status || status.toLowerCase() === "offline";
}

function statusTagType(
  status: string,
): "success" | "warning" | "info" | "danger" | "" {
  if (!status || status.toLowerCase() === "offline") return "info";
  if (status.toLowerCase() === "join me") return "success";
  if (status.toLowerCase() === "busy") return "danger";
  if (status.toLowerCase() === "ask me") return "warning";
  return "";
}

const friendsByStatus = computed(() => {
  const list = friends.value;
  if (showOfflineList.value) {
    return list.filter((f) => friendIsOffline(f.status));
  }
  return list.filter((f) => !friendIsOffline(f.status));
});

const filteredFriends = computed(() => {
  const q = displayNameQuery.value.trim().toLowerCase();
  const base = friendsByStatus.value;
  if (!q) return base;
  return base.filter((f) => f.displayName.toLowerCase().includes(q));
});

const emptyListMessage = computed(() => {
  if (friendsByStatus.value.length === 0) {
    return showOfflineList.value
      ? "オフラインのフレンドはいません"
      : "オンラインのフレンドはいません";
  }
  if (
    displayNameQuery.value.trim() !== "" &&
    filteredFriends.value.length === 0
  ) {
    return "検索に一致するフレンドはいません";
  }
  return "該当するフレンドはいません";
});

async function copyDisplayName(name: string): Promise<void> {
  const text = name || "";
  if (!text) return;
  try {
    await navigator.clipboard.writeText(text);
  } catch {
    const ta = document.createElement("textarea");
    ta.value = text;
    ta.setAttribute("readonly", "");
    ta.style.position = "fixed";
    ta.style.left = "-9999px";
    document.body.appendChild(ta);
    ta.select();
    try {
      document.execCommand("copy");
    } finally {
      document.body.removeChild(ta);
    }
  }
}

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
    // 失敗時は変化なし
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
.friends-toolbar {
  margin-bottom: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}

.friends-header {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.friends-search-input {
  max-width: 20rem;
}

.filter-mode {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  flex-wrap: wrap;
}

.mode-label {
  font-size: 0.9rem;
  color: var(--text-secondary);
  min-width: 3.25rem;
  transition: color 0.15s ease;
}

.mode-label.active {
  color: var(--text-primary);
  font-weight: 600;
}

.login-hint {
  margin-bottom: 1rem;
}

.friends-section {
  display: flex;
  gap: 1.5rem;
}

.friends-list {
  width: 320px;
  flex-shrink: 0;
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
  transition: background 0.15s;
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
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.friend-status-tag {
  flex-shrink: 0;
}

.btn-favorite {
  flex-shrink: 0;
  font-size: 1rem !important;
  padding: 0 4px !important;
}

.empty-message {
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin: 1rem 0;
}

.friend-detail {
  flex: 1;
  max-height: 560px;
  overflow-y: auto;
  background: var(--bg-secondary) !important;
  border-color: var(--border) !important;
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

.detail-display-name {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.favorite-toggle {
  margin-top: 1rem;
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

.link-list {
  margin: 0;
  padding-left: 1.25rem;
}
</style>
