<template>
  <div class="activity-view">
    <h1 class="page-title">アクティビティ</h1>

    <!-- 統計セクション（直近14日） -->
    <el-card class="section-card section-card--playtime" shadow="never">
      <template #header>
        <span>プレイ時間（直近14日）</span>
      </template>
      <div v-if="statsLoading" class="loading">読み込み中…</div>
      <div v-else-if="!statsRangeFrom" class="empty-stats">
        データがありません
      </div>
      <PlayTimeChart v-else :series="dailyPlayChartSeries" />
    </el-card>

    <!-- タイムラインセクション -->
    <el-card class="section-card section-card--encounters" shadow="never">
      <template #header>
        <span>遭遇ログ（滞在区間）</span>
      </template>
      <!-- フィルタ -->
      <div class="filters">
        <el-input
          v-model="displayNameFilter"
          placeholder="表示名で検索"
          clearable
          style="max-width: 220px"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button @click="loadEncounters">更新</el-button>
      </div>

      <div class="encounter-log-scroll">
        <div v-if="encountersLoading" class="loading">読み込み中…</div>
        <div v-else-if="filteredEncounters.length === 0" class="empty">
          遭遇ログがありません。
        </div>
        <el-table
          v-else
          :data="filteredEncounters"
          style="width: 100%"
          size="small"
          :border="false"
          stripe
        >
          <el-table-column label="入室" width="150">
            <template #default="{ row }">
              <span class="timeline-time">{{
                formatEncounteredAt(row.joinedAt)
              }}</span>
            </template>
          </el-table-column>
          <el-table-column label="退室" width="150">
            <template #default="{ row }">
              <span class="timeline-time">{{
                row.leftAt ? formatEncounteredAt(row.leftAt) : "—"
              }}</span>
            </template>
          </el-table-column>
          <el-table-column label="表示名" min-width="120">
            <template #default="{ row }">
              <el-button
                v-if="row.vrcUserId"
                link
                type="primary"
                class="timeline-link"
                @click="openUserHistory(row.vrcUserId)"
              >
                {{ row.displayName }}
              </el-button>
              <span v-else class="timeline-name-muted">{{
                row.displayName
              }}</span>
            </template>
          </el-table-column>
          <el-table-column label="ワールド名" min-width="120">
            <template #default="{ row }">
              <el-button
                v-if="row.worldId"
                link
                type="primary"
                class="timeline-link"
                :title="row.worldId"
                @click="openWorldHistory(row.worldId)"
              >
                {{ row.worldDisplayName || row.worldId }}
              </el-button>
              <span v-else>—</span>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import {
  App,
  type UserEncounterDTO,
  type ActivityStatsDTO,
} from "../wails/app";
import { getRuntime } from "../wails/runtime";
import PlayTimeChart, {
  type PlayTimeDayPoint,
} from "../components/PlayTimeChart.vue";
import { openEncounterHistoryWindow } from "../utils/openEncounterHistoryWindow";

const PLAYTIME_CHART_MAX_DAYS = 14;
const ACTIVITY_ENCOUNTERS_CHANGED_DEBOUNCE_MS = 400;

const router = useRouter();

const encounters = ref<UserEncounterDTO[]>([]);
const encountersLoading = ref(false);
const displayNameFilter = ref("");

const stats = ref<ActivityStatsDTO>({ dailyPlaySeconds: [], topWorlds: [] });
const statsLoading = ref(false);
const statsRangeFrom = ref("");
const statsRangeTo = ref("");

const dailyPlayChartSeries = computed((): PlayTimeDayPoint[] => {
  const from = statsRangeFrom.value;
  const to = statsRangeTo.value;
  if (!from || !to) return [];
  const byDate = new Map(
    stats.value.dailyPlaySeconds.map((d) => [d.date, d.seconds]),
  );
  const out: PlayTimeDayPoint[] = [];
  const start = new Date(from + "T00:00:00.000Z");
  const end = new Date(to + "T00:00:00.000Z");
  for (
    let cur = new Date(start);
    cur <= end;
    cur.setUTCDate(cur.getUTCDate() + 1)
  ) {
    const iso = cur.toISOString().slice(0, 10);
    out.push({
      date: iso,
      label: formatDateShort(iso),
      seconds: byDate.get(iso) ?? 0,
    });
  }
  return out;
});

const filteredEncounters = computed(() => {
  const list = encounters.value;
  const q = displayNameFilter.value.trim().toLowerCase();
  if (!q) return list;
  return list.filter((e) => e.displayName.toLowerCase().includes(q));
});

function formatDateShort(dateStr: string): string {
  try {
    const d = new Date(dateStr + "T12:00:00Z");
    const m = d.getMonth() + 1;
    const day = d.getDate();
    return `${m}/${day}`;
  } catch {
    return dateStr;
  }
}

function formatEncounteredAt(iso: string): string {
  try {
    const d = new Date(iso);
    return d.toLocaleString("ja-JP");
  } catch {
    return iso;
  }
}

function openUserHistory(vrcUserId: string): void {
  openEncounterHistoryWindow(router, "user", vrcUserId);
}

function openWorldHistory(worldId: string): void {
  openEncounterHistoryWindow(router, "world", worldId);
}

let encountersChangedDebounceTimer: ReturnType<typeof setTimeout> | null = null;
let unsubscribeEncountersChanged: (() => void) | undefined;

function scheduleLoadEncounters(): void {
  if (encountersChangedDebounceTimer !== null) {
    clearTimeout(encountersChangedDebounceTimer);
  }
  encountersChangedDebounceTimer = setTimeout(() => {
    encountersChangedDebounceTimer = null;
    void loadEncounters();
  }, ACTIVITY_ENCOUNTERS_CHANGED_DEBOUNCE_MS);
}

async function loadEncounters(): Promise<void> {
  encountersLoading.value = true;
  try {
    encounters.value = await App.encounters();
  } finally {
    encountersLoading.value = false;
  }
}

async function loadStats(): Promise<void> {
  statsLoading.value = true;
  try {
    const to = new Date();
    const from = new Date();
    from.setDate(from.getDate() - (PLAYTIME_CHART_MAX_DAYS - 1));
    const fromStr = from.toISOString().slice(0, 10);
    const toStr = to.toISOString().slice(0, 10);
    statsRangeFrom.value = fromStr;
    statsRangeTo.value = toStr;
    stats.value = await App.getActivityStats(fromStr, toStr);
  } finally {
    statsLoading.value = false;
  }
}

onMounted(() => {
  const rt = getRuntime();
  const off = rt?.EventsOn?.("activity:encounters-changed", () => {
    scheduleLoadEncounters();
  });
  if (typeof off === "function") {
    unsubscribeEncountersChanged = off;
  }
  void loadEncounters();
  void loadStats();
});

onUnmounted(() => {
  if (encountersChangedDebounceTimer !== null) {
    clearTimeout(encountersChangedDebounceTimer);
    encountersChangedDebounceTimer = null;
  }
  unsubscribeEncountersChanged?.();
});
</script>

<style scoped>
.activity-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  min-width: 0;
  width: 100%;
  overflow-y: auto;
}

.section-card {
  background: var(--bg-secondary) !important;
  border-color: var(--border) !important;
  width: 100%;
  min-width: 0;
}

.section-card :deep(.el-card__header) {
  font-weight: 600;
  border-bottom-color: var(--border);
  color: var(--text-secondary);
}

/* プレイ時間カード全体の縦幅を固定（ヘッダー + グラフ 280px 相当の body） */
.section-card--playtime {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  height: 336px;
  min-height: 336px;
  max-height: 336px;
}

.section-card--playtime :deep(.el-card__header) {
  flex-shrink: 0;
}

/* グラフ枠: 残り領域いっぱい・内部スクロールなし（PlayTimeChart は 280px 固定で中央寄せ可） */
.section-card--playtime :deep(.el-card__body) {
  flex: 1 1 0;
  min-height: 0;
  overflow: hidden;
  box-sizing: border-box;
  padding-top: 0;
  padding-bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
}

.section-card--playtime :deep(.el-card__body) > * {
  width: 100%;
}

.section-card--playtime :deep(.el-card__body) .loading,
.section-card--playtime :deep(.el-card__body) .empty-stats {
  padding: 1rem;
}

.section-card--encounters {
  min-height: 320px;
}

.section-card--encounters :deep(.el-card__body) {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
  width: 100%;
}

.filters {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  margin-bottom: 1rem;
  flex-wrap: wrap;
  flex-shrink: 0;
}

/* 一覧・空表示のみスクロール（el-card__body 全体はスクロールさせない） */
.encounter-log-scroll {
  overflow-y: auto;
  min-height: 0;
  flex: 1 1 auto;
  max-height: min(60vh, 32rem);
  width: 100%;
}

.loading,
.empty,
.empty-stats {
  padding: 2rem;
  text-align: center;
  color: var(--text-secondary);
}

.timeline-time {
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.timeline-link {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
}

.timeline-name-muted {
  color: var(--text-secondary);
}
</style>
