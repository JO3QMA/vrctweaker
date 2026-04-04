<template>
  <section class="friends-detail-pane">
    <VrcUserCacheDetail
      :selected="selected"
      @favorite-change="onFavoriteChange"
    />
  </section>
</template>

<script setup lang="ts">
import VrcUserCacheDetail from "../../components/VrcUserCacheDetail.vue";
import type { UserCacheDTO } from "../../wails/app";

defineProps<{
  selected: UserCacheDTO | null;
}>();

const emit = defineEmits<{
  favoriteChange: [user: UserCacheDTO, isFavorite: boolean];
}>();

function onFavoriteChange(user: UserCacheDTO, isFavorite: boolean) {
  emit("favoriteChange", user, isFavorite);
}
</script>

<style scoped>
.friends-detail-pane {
  flex: 1;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
