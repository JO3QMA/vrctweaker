<template>
  <div class="encounter-history-view">
    <h1 class="page-title">{{ pageTitle }}</h1>
    <p v-if="idLine" class="id-line">{{ idLine }}</p>

    <el-alert
      v-if="invalidQuery"
      title="表示できません。URL の kind / vrcUserId / worldId を確認してください。"
      type="warning"
      :closable="false"
      show-icon
    />
    <EncounterHistoryList
      v-else
      :mode="listMode"
      :user-id="vrcUserId"
      :world-id="worldId"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import EncounterHistoryList from "../components/EncounterHistoryList.vue";

const route = useRoute();

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
  if (kind.value === "user") return "ユーザー別 遭遇履歴";
  if (kind.value === "world") return "ワールド別 遭遇履歴";
  return "遭遇履歴";
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
