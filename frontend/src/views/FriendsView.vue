<template>
  <div class="friends-view">
    <div class="friends-section">
      <section class="friends-pane friends-pane--left">
        <h1 class="page-title">フレンド</h1>
        <FriendsViewToolbar
          v-model:show-offline-list="showOfflineList"
          v-model:display-name-query="displayNameQuery"
          :is-logged-in="isLoggedIn"
          :refresh-loading="refreshLoading"
          @refresh="doRefresh"
        />
        <el-alert
          v-if="!isLoggedIn"
          title="フレンド一覧の更新にはログインが必要です。設定画面でログインしてください。"
          type="info"
          :closable="false"
          show-icon
          class="login-hint"
        />
        <div class="friends-list-wrap">
          <FriendsListPanel
            :friends="filteredFriends"
            :selected="selected"
            :loading="loading"
            :empty-message="emptyListMessage"
            @select="selected = $event"
            @toggle-favorite="toggleFavorite"
          />
        </div>
      </section>
      <FriendsDetailPane
        :selected="selected"
        @favorite-change="onDetailFavoriteChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import FriendsDetailPane from "./friends/FriendsDetailPane.vue";
import FriendsListPanel from "./friends/FriendsListPanel.vue";
import FriendsViewToolbar from "./friends/FriendsViewToolbar.vue";
import { friendIsOffline } from "./friends/friendsViewUtils";
import { App } from "../wails/app";
import type { UserCacheDTO } from "../wails/app";

const route = useRoute();
const router = useRouter();

function firstQueryString(v: unknown): string {
  if (v == null) return "";
  if (typeof v === "string") return v;
  if (Array.isArray(v)) {
    for (const x of v) {
      if (typeof x === "string") return x;
    }
  }
  return "";
}

/** false: オンラインのみ / true: オフラインのみ */
const showOfflineList = ref(false);
const displayNameQuery = ref("");
const friends = ref<UserCacheDTO[]>([]);
const selected = ref<UserCacheDTO | null>(null);
const isLoggedIn = ref(false);
const loading = ref(true);
const refreshLoading = ref(false);

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

async function stripVrcUserIdFromQuery(): Promise<void> {
  const q = { ...route.query } as Record<string, string | string[] | undefined>;
  if (q.vrcUserId == null) return;
  delete q.vrcUserId;
  await router.replace({ path: route.path, query: q });
}

async function applyVrcUserIdFromQuery(): Promise<void> {
  const id = firstQueryString(route.query.vrcUserId).trim();
  if (!id) return;
  let f = friends.value.find((x) => x.vrcUserId === id);
  if (!f) {
    await loadFriends();
    f = friends.value.find((x) => x.vrcUserId === id);
  }
  if (!f) {
    ElMessage.warning(
      "指定されたユーザーはフレンド一覧に見つかりませんでした。",
    );
    await stripVrcUserIdFromQuery();
    return;
  }
  selected.value = f;
  showOfflineList.value = friendIsOffline(f.status);
  displayNameQuery.value = "";
  await stripVrcUserIdFromQuery();
}

onMounted(async () => {
  await loadFriends();
  isLoggedIn.value = await App.isLoggedIn();
  await applyVrcUserIdFromQuery();
});

watch(
  () => firstQueryString(route.query.vrcUserId),
  (id) => {
    if (id.trim() !== "") void applyVrcUserIdFromQuery();
  },
);

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

async function onDetailFavoriteChange(f: UserCacheDTO, isFavorite: boolean) {
  f.isFavorite = isFavorite;
  try {
    await App.setFavorite(f.vrcUserId, f.isFavorite);
  } catch {
    f.isFavorite = !f.isFavorite;
  }
}
</script>

<style scoped>
.friends-view {
  height: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.login-hint {
  margin-bottom: 1rem;
}

.friends-section {
  display: flex;
  align-items: stretch;
  flex: 1;
  min-height: 0;
  overflow: hidden;
  gap: 1.5rem;
}

.friends-pane--left {
  width: 320px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
}

.friends-list-wrap {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}
</style>
