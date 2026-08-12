<template>
  <div class="friends-list">
    <div
      v-for="f in friends"
      :key="f.vrcUserId"
      class="friend-card"
      :class="{ active: selected?.vrcUserId === f.vrcUserId }"
      @click="emit('select', f)"
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
      <VrcStatusTag :status="f.status" />
      <VtButton
        link
        :variant="f.isFavorite ? 'primary' : 'tertiary'"
        :title="
          f.isFavorite
            ? t('friendsList.favoriteRemove')
            : t('friendsList.favoriteAdd')
        "
        class="btn-favorite"
        @click.stop="emit('toggleFavorite', f)"
      >
        ★
      </VtButton>
    </div>
    <p v-if="friends.length === 0 && !loading" class="empty-message">
      {{ emptyMessage }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import VrcStatusTag from "../../components/VrcStatusTag.vue";
import VtButton from "../../components/VtButton.vue";
import type { UserCacheDTO } from "../../wails/app";
import { friendThumbUrl } from "@/utils/vrcUserCacheDisplay";

const { t } = useI18n();

defineProps<{
  friends: UserCacheDTO[];
  selected: UserCacheDTO | null;
  loading: boolean;
  emptyMessage: string;
}>();

const emit = defineEmits<{
  select: [user: UserCacheDTO];
  toggleFavorite: [user: UserCacheDTO];
}>();
</script>

<style scoped>
.friends-list {
  align-self: flex-start;
  box-sizing: border-box;
  width: 320px;
  flex-shrink: 0;
  min-height: 0;
  max-height: 100%;
  overflow-y: auto;
}

.friend-card {
  display: flex;
  align-items: center;
  gap: var(--space-action-group);
  padding: var(--space-form-field);
  margin-bottom: var(--space-action-group);
  background: var(--color-bg-elevated);
  border-radius: var(--radius);
  cursor: pointer;
  transition: background 0.15s;
}

.friend-card:hover,
.friend-card.active {
  background: var(--color-bg-muted);
}

.friend-thumb {
  width: 40px;
  height: 40px;
  border-radius: var(--radius);
  object-fit: cover;
  flex-shrink: 0;
}

.friend-thumb-placeholder {
  background: var(--color-bg-muted);
  border: 1px solid var(--color-border);
}

.friend-name {
  flex: 1;
  font-weight: var(--font-weight-500);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.btn-favorite {
  flex-shrink: 0;
  font-size: var(--font-size-16) !important;
  padding: 0 var(--space-inline-tight) !important;
}

.empty-message {
  font-size: var(--font-size-14);
  color: var(--color-text-secondary);
  margin: var(--space-block) 0;
}
</style>
