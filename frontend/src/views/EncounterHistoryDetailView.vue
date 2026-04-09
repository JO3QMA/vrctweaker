<template>
  <div class="encounter-history-view">
    <h1 class="page-title">{{ pageTitle }}</h1>
    <p v-if="idLine" class="id-line">{{ idLine }}</p>

    <el-alert
      v-if="invalidQuery"
      :title="t('encounterHistory.invalidQuery')"
      type="warning"
      :closable="false"
      show-icon
    />
    <EncounterHistoryList
      v-else
      :mode="listMode"
      :user-id="vrcUserId"
      :world-id="worldId"
      :hide-display-name-column="listMode === 'user'"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import EncounterHistoryList from "../components/EncounterHistoryList.vue";

const route = useRoute();
const { t } = useI18n();

function firstQueryString(v: unknown): string {
  if (v == null) return "";
  if (typeof v === "string") return v;
  if (Array.isArray(v)) {
    for (const x of v) {
      if (typeof x === "string") return x;
    }
  }
  return "";
}

const kind = computed(() => firstQueryString(route.query.kind));
const vrcUserId = computed(() => firstQueryString(route.query.vrcUserId));
const worldId = computed(() => firstQueryString(route.query.worldId));

const invalidQuery = computed(() => {
  if (kind.value !== "user" && kind.value !== "world") return true;
  if (kind.value === "user" && !vrcUserId.value.trim()) return true;
  if (kind.value === "world" && !worldId.value.trim()) return true;
  return false;
});

const listMode = computed<"user" | "world">(() =>
  kind.value === "world" ? "world" : "user",
);

const pageTitle = computed(() => {
  if (kind.value === "user") return t("encounterHistory.titleUser");
  if (kind.value === "world") return t("encounterHistory.titleWorld");
  return t("encounterHistory.titleDefault");
});

const idLine = computed(() => {
  if (invalidQuery.value) return "";
  if (kind.value === "user") return vrcUserId.value;
  if (kind.value === "world") return worldId.value;
  return "";
});
</script>

<style scoped>
.encounter-history-view {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  min-height: 0;
}

.id-line {
  margin: 0;
  font-size: 0.8rem;
  color: var(--text-secondary);
  word-break: break-all;
}
</style>
