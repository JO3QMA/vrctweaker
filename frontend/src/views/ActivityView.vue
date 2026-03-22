<template>
  <div class="activity-view">
    <h1 class="page-title">アクティビティ</h1>

    <!-- 統計セクション（直近14日） -->
    <section class="stats-section">
      <h2 class="section-title">プレイ時間（直近14日）</h2>
      <div v-if="statsLoading" class="loading">読み込み中…</div>
      <div v-else-if="!statsRangeFrom" class="empty-stats">
        データがありません
      </div>
      <PlayTimeChart v-else :series="dailyPlayChartSeries" />
    </section>

    <!-- タイムラインセクション -->
    <section class="timeline-section">
      <h2 class="section-title">遭遇ログ（Join/Leave）</h2>

      <!-- フィルタ -->
      <div class="filters">
        <input
          v-model="displayNameFilter"
          type="text"
          placeholder="表示名で検索"
          class="filter-input"
        />
        <button class="btn-refresh" @click="loadEncounters">更新</button>
      </div>

      <div v-if="encountersLoading" class="loading">読み込み中…</div>
      <div v-else-if="filteredEncounters.length === 0" class="empty">
        遭遇ログがありません。
      </div>
      <ul v-else class="timeline">
        <li
          v-for="enc in filteredEncounters"
          :key="enc.id"
          class="timeline-item"
        >
          <span class="timeline-time">{{
            formatEncounteredAt(enc.encounteredAt)
          }}</span>
          <span class="timeline-name">{{ enc.displayName }}</span>
          <span class="timeline-action" :class="enc.action">{{
            actionLabel(enc.action)
          }}</span>
          <span class="timeline-world" :title="enc.worldId || ''">{{
            enc.worldDisplayName || enc.worldId || "—"
          }}</span>
          <span class="timeline-user-meta">{{ encounterUserMeta(enc) }}</span>
          <span class="timeline-instance">{{ enc.instanceId || "—" }}</span>
        </li>
      </ul>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import {
  App,
  type UserEncounterDTO,
  type ActivityStatsDTO,
} from "../wails/app";
import { getRuntime } from "../wails/runtime";
import PlayTimeChart, {
  type PlayTimeDayPoint,
} from "../components/PlayTimeChart.vue";

/** プレイ時間グラフに表示する暦日数（最大14日、今日を含む） */
const PLAYTIME_CHART_MAX_DAYS = 14;

const ACTIVITY_ENCOUNTERS_CHANGED_DEBOUNCE_MS = 400;

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

function actionLabel(action: string): string {
  if (action === "join") return "参加";
  if (action === "leave") return "退出";
  return action;
}

function encounterUserMeta(enc: UserEncounterDTO): string {
  if (enc.isFirstEncounter) {
    return "初めての遭遇";
  }
  if (enc.userFirstSeenAt) {
    return `初見: ${formatEncounteredAt(enc.userFirstSeenAt)} / 最終: ${
      enc.userLastContactAt ? formatEncounteredAt(enc.userLastContactAt) : "—"
    }`;
  }
  if (enc.userLastContactAt) {
    return `最終接触: ${formatEncounteredAt(enc.userLastContactAt)}`;
  }
  return "—";
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
  gap: 2rem;
}

.page-title {
  margin: 0;
  font-size: 1.5rem;
}

.section-title {
  margin: 0 0 1rem;
  font-size: 1.1rem;
  color: var(--text-secondary);
}

.stats-section,
.timeline-section {
  padding: 1rem;
  background: var(--bg-secondary);
  border-radius: var(--radius);
  border: 1px solid var(--border);
}

.filters {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  margin-bottom: 1rem;
}

.filter-input {
  width: 220px;
  padding: 0.5rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
}

.btn-refresh {
  padding: 0.5rem 1rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
  cursor: pointer;
}

.btn-refresh:hover {
  background: var(--accent);
  color: white;
  border-color: var(--accent);
}

.loading,
.empty,
.empty-stats {
  padding: 2rem;
  text-align: center;
  color: var(--text-secondary);
}

.timeline {
  list-style: none;
  margin: 0;
  padding: 0;
}

.timeline-item {
  display: grid;
  grid-template-columns:
    10rem 8rem 4rem minmax(0, 11rem) minmax(0, 16rem)
    minmax(0, 1fr);
  gap: 0.75rem;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--border);
  font-size: 0.9rem;
  align-items: center;
}

.timeline-item:last-child {
  border-bottom: none;
}

.timeline-time {
  color: var(--text-secondary);
  font-size: 0.85rem;
}

.timeline-name {
  font-weight: 500;
}

.timeline-action.join {
  color: var(--success);
}

.timeline-action.leave {
  color: var(--text-secondary);
}

.timeline-world {
  font-size: 0.8rem;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.timeline-user-meta {
  font-size: 0.75rem;
  color: var(--text-secondary);
  line-height: 1.3;
  overflow: hidden;
}

.timeline-instance {
  font-family: monospace;
  font-size: 0.8rem;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
