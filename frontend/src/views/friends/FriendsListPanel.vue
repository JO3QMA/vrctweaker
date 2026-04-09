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
      <el-button
        link
        :type="f.isFavorite ? 'primary' : 'info'"
        :title="
          f.isFavorite
            ? t('friendsList.favoriteRemove')
            : t('friendsList.favoriteAdd')
        "
        class="btn-favorite"
        @click.stop="emit('toggleFavorite', f)"
      >
        ★
      </el-button>
    </div>
    <p v-if="friends.length === 0 && !loading" class="empty-message">
      {{ emptyMessage }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import VrcStatusTag from "../../components/VrcStatusTag.vue";
import type { UserCacheDTO } from "../../wails/app";
import { friendThumbUrl } from "./friendsViewUtils";

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
</style>
