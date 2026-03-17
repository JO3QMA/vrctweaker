<template>
  <div class="friends-view">
    <h1 class="page-title">
      フレンド
    </h1>
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
    <p
      v-if="!isLoggedIn"
      class="hint"
    >
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
      <div
        v-if="selected"
        class="friend-detail"
      >
        <h3>詳細</h3>
        <dl class="detail-list">
          <dt>表示名</dt>
          <dd>{{ selected.displayName }}</dd>
          <dt>ステータス</dt>
          <dd>{{ selected.status || "—" }}</dd>
        </dl>
        <label class="favorite-toggle">
          <input
            v-model="selected.isFavorite"
            type="checkbox"
            @change="applyFavorite(selected)"
          >
          お気に入り
        </label>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { App } from "../wails/app";
import type { FriendCacheDTO } from "../wails/app";

const activeTab = ref<"online" | "offline">("online");
const friends = ref<FriendCacheDTO[]>([]);
const selected = ref<FriendCacheDTO | null>(null);
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

async function toggleFavorite(f: FriendCacheDTO) {
  const next = !f.isFavorite;
  try {
    await App.setFavorite(f.vrcUserId, next);
    f.isFavorite = next;
  } catch {
    // 失敗時は変化なし（一覧の星ボタンではまだ反映していない）
  }
}

async function applyFavorite(f: FriendCacheDTO) {
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

.friend-name {
  flex: 1;
  font-weight: 500;
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
}

.friend-detail h3 {
  margin: 0 0 1rem;
  font-size: 1.1rem;
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

.favorite-toggle {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.9rem;
  cursor: pointer;
}
</style>
