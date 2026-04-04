<template>
  <div class="encounter-history-list">
    <template v-if="canLoad">
      <div v-if="loading" class="message">読み込み中…</div>
      <el-alert
        v-else-if="error"
        :title="error"
        type="error"
        :closable="false"
        show-icon
      />
      <div v-else-if="rows.length === 0" class="message">
        該当する遭遇ログがありません。
      </div>
      <el-table v-else :data="rows" style="width: 100%" size="small" stripe>
        <el-table-column label="入室" width="155">
          <template #default="{ row }">
            {{ formatEncounteredAt(row.joinedAt) }}
          </template>
        </el-table-column>
        <el-table-column label="退室" width="155">
          <template #default="{ row }">
            {{ row.leftAt ? formatEncounteredAt(row.leftAt) : "—" }}
          </template>
        </el-table-column>
        <el-table-column
          v-if="!hideDisplayNameColumn"
          label="表示名"
          min-width="120"
          prop="displayName"
        />
        <el-table-column label="ワールド名" min-width="120">
          <template #default="{ row }">
            <span :title="row.worldId || ''">
              {{ row.worldDisplayName || row.worldId || "—" }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="インスタンス" min-width="120">
          <template #default="{ row }">
            <span class="mono">{{ row.instanceId || "—" }}</span>
          </template>
        </el-table-column>
      </el-table>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { App, type UserEncounterDTO } from "../wails/app";
import { formatEncounteredAt } from "../utils/formatEncounteredAt";

const props = withDefaults(
  defineProps<{
    mode: "user" | "world";
    userId?: string;
    worldId?: string;
    /** 同一ユーザーのプロフィール内など、表示名が自明なときは true */
    hideDisplayNameColumn?: boolean;
  }>(),
  {
    hideDisplayNameColumn: false,
    userId: undefined,
    worldId: undefined,
  },
);

const loading = ref(false);
const error = ref<string | null>(null);
const rows = ref<UserEncounterDTO[]>([]);

const canLoad = computed(() => {
  if (props.mode === "user") return Boolean(props.userId?.trim());
  return Boolean(props.worldId?.trim());
});

async function load(): Promise<void> {
  if (!canLoad.value) {
    rows.value = [];
    error.value = null;
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = null;
  try {
    if (props.mode === "user") {
      rows.value = await App.encountersByVRCUserID(props.userId!.trim());
    } else {
      rows.value = await App.encountersByWorldID(props.worldId!.trim());
    }
  } catch (e) {
    rows.value = [];
    error.value =
      e instanceof Error ? e.message : "データの取得に失敗しました。";
  } finally {
    loading.value = false;
  }
}

watch(
  () => [props.mode, props.userId, props.worldId] as const,
  () => {
    void load();
  },
  { immediate: true },
);
</script>

<style scoped>
.encounter-history-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-height: 0;
}

.message {
  padding: 1rem;
  text-align: center;
  color: var(--text-secondary);
}

.mono {
  font-family: monospace;
  font-size: 0.78rem;
  word-break: break-all;
}
</style>
