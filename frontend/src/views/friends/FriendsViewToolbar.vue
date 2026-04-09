<template>
  <div class="friends-toolbar">
    <div class="friends-header">
      <div
        class="filter-mode"
        role="group"
        :aria-label="t('friends.toolbarOnlineOffline')"
      >
        <span :class="['mode-label', { active: !showOfflineList }]"
          >Online</span
        >
        <el-switch
          v-model="showOfflineList"
          data-testid="friends-filter-mode"
          :aria-label="t('friends.toolbarOfflineSwitch')"
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
          isLoggedIn
            ? t('friends.refreshTitleOk')
            : t('friends.refreshTitleNeedLogin')
        "
        @click="emit('refresh')"
      >
        {{ refreshLoading ? t("common.updating") : t("common.refresh") }}
      </el-button>
    </div>
    <el-input
      v-model.trim="displayNameQuery"
      type="search"
      :placeholder="t('friends.searchPlaceholder')"
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
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";

const { t } = useI18n();

defineProps<{
  isLoggedIn: boolean;
  refreshLoading: boolean;
}>();

const emit = defineEmits<{
  refresh: [];
}>();

const showOfflineList = defineModel<boolean>("showOfflineList", {
  required: true,
});
const displayNameQuery = defineModel<string>("displayNameQuery", {
  required: true,
});
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
</style>
