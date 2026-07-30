<template>
  <div class="dashboard">
    <h1 class="page-title">{{ t("routes.dashboard") }}</h1>
    <div class="quick-actions">
      <ServerStatusSection />
      <DashboardLaunchBlock />
      <PresenceChangeSection />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import ServerStatusSection from "../components/ServerStatusSection.vue";
import DashboardLaunchBlock from "../components/DashboardLaunchBlock.vue";
import PresenceChangeSection from "../components/PresenceChangeSection.vue";

const { t } = useI18n();
</script>

<style scoped>
.dashboard {
  max-width: 600px;
  margin: 0 auto;
}

.quick-actions {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
</style>
