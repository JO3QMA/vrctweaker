<template>
  <div class="activity-view">
    <h1 class="page-title">{{ t("activity.title") }}</h1>

    <!-- 統計セクション（直近14日） -->
    <CollapsibleSectionCard
      class="section-card--playtime"
      :title="t('activity.playtime14')"
    >
      <div v-if="statsLoading" class="loading">{{ t("common.loading") }}</div>
      <div v-else-if="!statsRangeFrom" class="empty-stats">
        {{ t("activity.noData") }}
      </div>
      <PlayTimeChart v-else :series="dailyPlayChartSeries" />
    </CollapsibleSectionCard>

    <!-- タイムラインセクション -->
    <CollapsibleSectionCard
      class="section-card--encounters"
      :title="t('activity.encounterLog')"
    >
      <!-- フィルタ -->
      <div class="filters">
        <el-input
          v-model="displayNameFilter"
          :placeholder="t('activity.searchDisplayName')"
          clearable
          style="max-width: 220px"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button @click="loadEncounters">{{ t("common.refresh") }}</el-button>
      </div>

      <div class="encounter-log-scroll">
        <div v-if="encountersLoading" class="loading">
          {{ t("common.loading") }}
        </div>
        <div v-else-if="filteredEncounters.length === 0" class="empty">
          {{ t("activity.noEncounters") }}
        </div>
        <el-table
          v-else
          :data="filteredEncounters"
          style="width: 100%"
          size="small"
          :border="false"
          stripe
        >
          <el-table-column :label="t('common.joined')" width="150">
            <template #default="{ row }">
              <span class="timeline-time">{{
                formatEncounteredAt(row.joinedAt)
              }}</span>
            </template>
          </el-table-column>
          <el-table-column :label="t('common.left')" width="150">
            <template #default="{ row }">
              <span class="timeline-time">{{
                row.leftAt ? formatEncounteredAt(row.leftAt) : "—"
              }}</span>
            </template>
          </el-table-column>
          <el-table-column :label="t('common.displayName')" min-width="120">
            <template #default="{ row }">
              <el-button
                v-if="row.vrcUserId"
                link
                type="primary"
                class="timeline-link"
                @click="openUserFromEncounter(row)"
              >
                {{ row.displayName }}
              </el-button>
              <span v-else class="timeline-name-muted">{{
                row.displayName
              }}</span>
            </template>
          </el-table-column>
          <el-table-column :label="t('common.worldName')" min-width="120">
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
    </CollapsibleSectionCard>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import CollapsibleSectionCard from "../components/CollapsibleSectionCard.vue";
import {
  App,
  type ActivityStatsDTO,
  type UserEncounterDTO,
} from "../wails/app";
import { getRuntime } from "../wails/runtime";
import PlayTimeChart, {
  type PlayTimeDayPoint,
} from "../components/PlayTimeChart.vue";
import { openEncounterHistoryWindow } from "../utils/openEncounterHistoryWindow";

const PLAYTIME_CHART_MAX_DAYS = 14;
const ACTIVITY_ENCOUNTERS_CHANGED_DEBOUNCE_MS = 400;

const router = useRouter();
const { locale, t } = useI18n();

const encounters = ref<UserEncounterDTO[]>([]);
const encountersLoading = ref(false);
const displayNameFilter = ref("");

const stats = ref<ActivityStatsDTO>({ dailyPlaySeconds: [], topWorlds: [] });
const statsLoading = ref(false);
const statsRangeFrom = ref("");
const statsRangeTo = ref("");

const dailyPlayChartSeries = computed((): PlayTimeDayPoint[] => {
  void locale.value;
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
    return d.toLocaleDateString(locale.value, {
      month: "numeric",
      day: "numeric",
    });
  } catch {
    return dateStr;
  }
}

function formatEncounteredAt(iso: string): string {
  try {
    const d = new Date(iso);
    return d.toLocaleString(locale.value);
  } catch {
    return iso;
  }
}

async function openUserFromEncounter(row: UserEncounterDTO): Promise<void> {
  const vrcUserId = row.vrcUserId;
  if (!vrcUserId) return;
  const displayName = row.displayName ?? "";
  try {
    const nav = await App.resolveUserProfileNavigation(vrcUserId);
    if (nav.openInFriendsView) {
      await router.push({ name: "friends", query: { vrcUserId } });
    } else {
      await router.push({
        name: "user-profile",
        query: { vrcUserId, displayName },
      });
    }
  } catch {
    await router.push({
      name: "user-profile",
      query: { vrcUserId, displayName },
    });
  }
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
/* メイン領域（router-outlet-host）の残り高さを使い、遭遇ログはカード内でスクロール */
.activity-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  min-width: 0;
  width: 100%;
  min-height: 0;
  flex: 1 1 0;
  overflow: hidden;
}

.activity-view > .page-title {
  flex-shrink: 0;
}

.section-card {
  background: var(--bg-secondary) !important;
  border-color: var(--border) !important;
  width: 100%;
  min-width: 0;
  padding: 0;
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

.section-card--playtime.section-card--collapsed {
  height: auto;
  min-height: 0;
  max-height: none;
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

.section-card--encounters.section-card--collapsed {
  flex: 0 0 auto;
}

.section-card--encounters:not(.section-card--collapsed) {
  flex: 1 1 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.section-card--encounters:not(.section-card--collapsed)
  :deep(.el-card__header) {
  flex-shrink: 0;
}

.section-card--encounters:not(.section-card--collapsed) :deep(.el-card__body) {
  display: flex;
  flex-direction: column;
  flex: 1 1 0;
  overflow: hidden;
  min-height: 0;
  width: 100%;
}

.section-card--encounters:not(.section-card--collapsed)
  :deep(.section-card__panel) {
  display: flex;
  flex-direction: column;
  flex: 1 1 0;
  min-height: 0;
  overflow: hidden;
  width: 100%;
  min-width: 0;
}

.filters {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  margin-bottom: 1rem;
  flex-wrap: wrap;
  flex-shrink: 0;
}

/* セクションカードの残り高さに合わせてスクロール（親チェーンに flex + min-height:0 あり） */
.encounter-log-scroll {
  flex: 1 1 0;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
  width: 100%;
  box-sizing: border-box;
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
