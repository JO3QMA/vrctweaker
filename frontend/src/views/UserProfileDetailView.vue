<template>
  <div class="user-profile-detail-view">
    <h1 class="page-title">ユーザー</h1>
    <div v-if="loading" class="msg">読み込み中…</div>
    <el-alert
      v-else-if="loadError && !selected"
      :title="loadError"
      type="warning"
      :closable="false"
      show-icon
    />
    <div v-else-if="selected" class="detail-wrap">
      <VrcUserCacheDetail
        :selected="selected"
        @favorite-change="onFavoriteChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import { useRoute } from "vue-router";
import VrcUserCacheDetail from "../components/VrcUserCacheDetail.vue";
import { App } from "../wails/app";
import type { UserCacheDTO } from "../wails/app";

const route = useRoute();

const loading = ref(true);
const loadError = ref<string | null>(null);
const selected = ref<UserCacheDTO | null>(null);

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

function minimalFromQuery(
  vrcUserId: string,
  displayName: string,
): UserCacheDTO {
  return {
    vrcUserId,
    displayName: displayName || vrcUserId,
    status: "",
    isFavorite: false,
    lastUpdated: "",
  } as UserCacheDTO;
}

async function load(): Promise<void> {
  const vrcUserId = firstQueryString(route.query.vrcUserId).trim();
  const hint = firstQueryString(route.query.displayName);
  loading.value = true;
  loadError.value = null;
  selected.value = null;
  if (!vrcUserId) {
    loadError.value = "ユーザー ID が指定されていません。";
    loading.value = false;
    return;
  }
  try {
    const nav = await App.resolveUserProfileNavigation(vrcUserId);
    selected.value = nav.user;
    if (!selected.value?.vrcUserId) {
      selected.value = minimalFromQuery(vrcUserId, hint);
    }
  } catch {
    loadError.value =
      "プロフィールを取得できませんでした。キャッシュまたはログイン状態を確認してください。";
    selected.value = minimalFromQuery(vrcUserId, hint);
  } finally {
    loading.value = false;
  }
}

async function onFavoriteChange(f: UserCacheDTO, isFavorite: boolean) {
  f.isFavorite = isFavorite;
  try {
    await App.setFavorite(f.vrcUserId, f.isFavorite);
  } catch {
    f.isFavorite = !f.isFavorite;
  }
}

watch(
  () => firstQueryString(route.query.vrcUserId),
  () => {
    void load();
  },
);

void load();
</script>

<style scoped>
.user-profile-detail-view {
  height: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.msg {
  padding: 1rem;
  color: var(--text-secondary);
}

.detail-wrap {
  flex: 1;
  min-height: 0;
  min-width: 0;
  width: 100%;
  align-self: stretch;
  display: flex;
  flex-direction: column;
}
</style>
