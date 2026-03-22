<template>
  <div class="activity-view">
    <h1 class="page-title">アクティビティ</h1>

    <!-- 統計セクション（直近7日） -->
    <section class="stats-section">
      <h2 class="section-title">プレイ時間（直近7日）</h2>
      <div v-if="statsLoading" class="loading">読み込み中…</div>
      <div v-else-if="stats.dailyPlaySeconds.length === 0" class="empty-stats">
        データがありません
      </div>
      <div v-else class="bar-chart">
        <div
          v-for="day in stats.dailyPlaySeconds"
          :key="day.date"
          class="bar-row"
        >
          <span class="bar-label">{{ formatDateShort(day.date) }}</span>
          <div class="bar-track">
            <div
              class="bar-fill"
              :style="{ width: barWidthPercent(day) + '%' }"
            />
          </div>
          <span class="bar-value">{{ formatSeconds(day.seconds) }}</span>
        </div>
      </div>
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

const ACTIVITY_ENCOUNTERS_CHANGED_DEBOUNCE_MS = 400;

const encounters = ref<UserEncounterDTO[]>([]);
const encountersLoading = ref(false);
const displayNameFilter = ref("");

const stats = ref<ActivityStatsDTO>({ dailyPlaySeconds: [], topWorlds: [] });
const statsLoading = ref(false);

const maxBarSeconds = computed(() => {
  const daily = stats.value.dailyPlaySeconds;
  if (daily.length === 0) return 1;
  return Math.max(...daily.map((d) => d.seconds), 1);
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

function formatSeconds(seconds: number): string {
  if (seconds < 60) return `${seconds}秒`;
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  if (s === 0) return `${m}分`;
  return `${m}分${s}秒`;
}

function barWidthPercent(day: { date: string; seconds: number }): number {
  const max = maxBarSeconds.value;
  return Math.min(100, (day.seconds / max) * 100);
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
    from.setDate(from.getDate() - 6);
    const fromStr = from.toISOString().slice(0, 10);
    const toStr = to.toISOString().slice(0, 10);
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

.bar-chart {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.bar-row {
  display: grid;
  grid-template-columns: 4rem 1fr 5rem;
  align-items: center;
  gap: 0.75rem;
}

.bar-label {
  font-size: 0.9rem;
  color: var(--text-secondary);
}

.bar-track {
  height: 1.25rem;
  background: var(--bg-tertiary);
  border-radius: 4px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  background: var(--accent);
  border-radius: 4px;
  min-width: 2px;
  transition: width 0.2s ease;
}

.bar-value {
  font-size: 0.85rem;
  color: var(--text-secondary);
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
