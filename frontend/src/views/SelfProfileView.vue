<template>
  <div class="self-profile-view">
    <h1 class="page-title">{{ t("selfProfile.title") }}</h1>
    <el-alert
      v-if="!isLoggedIn"
      :title="t('selfProfile.loginRequired')"
      type="info"
      :closable="false"
      show-icon
      class="login-hint"
    >
      <template #default>
        <router-link class="settings-link" :to="{ name: 'settings' }">
          {{ t("selfProfile.openSettings") }}
        </router-link>
      </template>
    </el-alert>
    <div v-else-if="loading" class="msg">{{ t("selfProfile.loading") }}</div>
    <el-alert
      v-else-if="loadError && !selected"
      :title="loadError"
      type="warning"
      :closable="false"
      show-icon
    />
    <div v-else-if="selected" class="detail-wrap">
      <VrcUserCacheDetail
        variant="self"
        :selected="selected"
        :refresh-loading="refreshLoading"
        @refresh="onRefresh"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import VrcUserCacheDetail from "../components/VrcUserCacheDetail.vue";
import { useSessionUnlock } from "../composables/useSessionUnlock";
import { App } from "../wails/app";
import type { UserCacheDTO } from "../wails/app";

const { t } = useI18n();
const { beginStartupUnlock } = useSessionUnlock();

const isLoggedIn = ref(false);
const loading = ref(true);
const refreshLoading = ref(false);
const loadError = ref<string | null>(null);
const selected = ref<UserCacheDTO | null>(null);

function formatBackendError(e: unknown, fallback: string): string {
  if (e instanceof Error && e.message) return e.message;
  if (typeof e === "string" && e) return e;
  return fallback;
}

async function load(forceRefresh = false): Promise<void> {
  if (!isLoggedIn.value) {
    loading.value = false;
    selected.value = null;
    loadError.value = null;
    return;
  }
  loadError.value = null;
  if (forceRefresh) {
    refreshLoading.value = true;
  } else {
    loading.value = true;
  }
  try {
    selected.value = await App.getSelfProfile(forceRefresh);
  } catch (e) {
    selected.value = null;
    loadError.value = formatBackendError(e, t("selfProfile.loadFailed"));
  } finally {
    loading.value = false;
    refreshLoading.value = false;
  }
}

async function onRefresh(): Promise<void> {
  await load(true);
}

onMounted(async () => {
  await beginStartupUnlock().catch(() => undefined);
  isLoggedIn.value = await App.isLoggedIn();
  await load();
});
</script>

<style scoped>
.self-profile-view {
  height: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.login-hint {
  margin-bottom: 1rem;
}

.settings-link {
  color: var(--el-color-primary);
  text-decoration: none;
}

.settings-link:hover {
  text-decoration: underline;
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
