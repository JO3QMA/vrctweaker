<template>
  <div class="settings-frame">
    <div class="settings-view">
      <h1 class="page-title">{{ t("settings.title") }}</h1>
      <div class="settings-shell">
        <div class="settings-layout">
          <nav
            class="settings-nav"
            :aria-label="t('settings.navLabel')"
            data-testid="settings-nav"
          >
            <router-link
              v-for="item in navItems"
              :key="item.id"
              class="settings-nav__link"
              :to="`/settings/${item.id}`"
              :data-testid="`settings-nav-${item.id}`"
            >
              {{ t(item.labelKey) }}
            </router-link>
            <div
              class="settings-nav__divider"
              role="separator"
              data-testid="settings-nav-divider"
            />
            <router-link
              class="settings-nav__link"
              to="/settings/reset"
              data-testid="settings-nav-reset"
            >
              {{ t("settings.resetSetupSection") }}
            </router-link>
          </nav>
          <div class="settings-panel">
            <SettingsAccountPanel v-if="section === 'account'" />
            <SettingsGeneralPanel v-else-if="section === 'general'" />
            <SettingsDataPanel v-else-if="section === 'data'" />
            <SettingsAboutPanel v-else-if="section === 'about'" />
            <SettingsResetPanel v-else-if="section === 'reset'" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";
import SettingsAboutPanel from "../components/settings/SettingsAboutPanel.vue";
import SettingsAccountPanel from "../components/settings/SettingsAccountPanel.vue";
import SettingsDataPanel from "../components/settings/SettingsDataPanel.vue";
import SettingsGeneralPanel from "../components/settings/SettingsGeneralPanel.vue";
import SettingsResetPanel from "../components/settings/SettingsResetPanel.vue";
import { settingsSectionFromPath } from "./settings/settingsSections";

const { t } = useI18n();
const route = useRoute();

const section = computed(() => settingsSectionFromPath(route.path));

const navItems = [
  { id: "account", labelKey: "settings.navAccount" },
  { id: "general", labelKey: "settings.navGeneral" },
  { id: "data", labelKey: "settings.navData" },
  { id: "about", labelKey: "settings.navAbout" },
] as const;
</script>
