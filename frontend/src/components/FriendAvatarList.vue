<template>
  <div class="friend-avatar-list" data-testid="friend-avatar-list">
    <template v-if="canLoad">
      <div v-if="loading" class="message">{{ t("common.loading") }}</div>
      <VtAlert v-else-if="error" variant="danger" :title="error" />
      <div v-else-if="rows.length === 0" class="message">
        {{ t("friendAvatars.empty") }}
      </div>
      <el-table v-else :data="rows" style="width: 100%" size="small" stripe>
        <el-table-column
          :label="t('friendAvatars.colAvatarName')"
          min-width="160"
          prop="avatarName"
        />
        <el-table-column
          :label="t('friendAvatars.colLastSeen')"
          :width="ENCOUNTER_LOG_TIME_COL_WIDTH"
        >
          <template #default="{ row }">
            <span class="friend-avatar-log-time">{{
              formatSeenLocal(row.lastSeenAt)
            }}</span>
          </template>
        </el-table-column>
        <el-table-column
          :label="t('friendAvatars.colFirstSeen')"
          :width="ENCOUNTER_LOG_TIME_COL_WIDTH"
        >
          <template #default="{ row }">
            <span class="friend-avatar-log-time">{{
              formatSeenLocal(row.firstSeenAt)
            }}</span>
          </template>
        </el-table-column>
        <el-table-column
          :label="t('friendAvatars.colCount')"
          width="88"
          align="right"
          prop="useCount"
        />
      </el-table>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import VtAlert from "./VtAlert.vue";
import { App, type FriendAvatarUsageDTO } from "../wails/app";
import {
  ENCOUNTER_LOG_TIME_COL_WIDTH,
  formatEncounterLogTimestamp,
} from "../utils/formatEncounteredAt";
import { friendAvatarListFetchErrorMessage } from "../utils/friendAvatarListErrors";

const { t } = useI18n();

function formatSeenLocal(iso: string): string {
  return formatEncounterLogTimestamp(iso) ?? t("common.dash");
}

const props = defineProps<{
  userId?: string;
}>();

const loading = ref(false);
const error = ref<string | null>(null);
const rows = ref<FriendAvatarUsageDTO[]>([]);

let loadToken = 0;

const canLoad = computed(() => Boolean(props.userId?.trim()));

async function load(): Promise<void> {
  if (!canLoad.value) {
    loadToken += 1;
    rows.value = [];
    error.value = null;
    loading.value = false;
    return;
  }
  const userId = props.userId!.trim();
  const token = ++loadToken;
  loading.value = true;
  error.value = null;
  try {
    const result = await App.friendAvatarUsageByVRCUserID(userId);
    if (token !== loadToken) return;
    rows.value = result;
  } catch (e) {
    if (token !== loadToken) return;
    rows.value = [];
    error.value = friendAvatarListFetchErrorMessage(e, t);
  } finally {
    if (token === loadToken) {
      loading.value = false;
    }
  }
}

watch(
  () => props.userId,
  () => {
    void load();
  },
  { immediate: true },
);
</script>

<style scoped>
.friend-avatar-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-action-group);
  min-height: 0;
}

.message {
  padding: var(--space-block);
  text-align: center;
  color: var(--color-text-secondary);
}

.friend-avatar-log-time {
  white-space: nowrap;
}
</style>
