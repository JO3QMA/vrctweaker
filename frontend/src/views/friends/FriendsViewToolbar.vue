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
      <VtButton
        variant="primary"
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
      </VtButton>
    </div>
    <VtInput
      v-model.trim="displayNameQuery"
      type="search"
      :placeholder="t('friends.searchPlaceholder')"
      data-testid="friends-search-display-name"
      clearable
      class="friends-search-input"
      autocomplete="off"
    >
      <template #prefix>
        <VtIcon size="default"><Search /></VtIcon>
      </template>
    </VtInput>
  </div>
</template>

<script setup lang="ts">
import { Search } from "@element-plus/icons-vue";
import { useI18n } from "vue-i18n";
import VtButton from "../../components/VtButton.vue";
import VtIcon from "../../components/VtIcon.vue";
import VtInput from "../../components/VtInput.vue";

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
  margin-bottom: var(--space-block);
  display: flex;
  flex-direction: column;
  gap: var(--space-form-field);
}

.friends-header {
  display: flex;
  align-items: center;
  gap: var(--space-block);
}

.friends-search-input {
  max-width: 20rem;
}

.filter-mode {
  display: flex;
  align-items: center;
  gap: var(--space-form-field);
  flex-wrap: wrap;
}

.mode-label {
  font-size: var(--font-size-14);
  color: var(--color-text-secondary);
  min-width: 3.25rem;
  transition: color 0.15s ease;
}

.mode-label.active {
  color: var(--color-text-primary);
  font-weight: var(--font-weight-600);
}
</style>
