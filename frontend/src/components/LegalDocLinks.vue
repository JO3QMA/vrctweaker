<template>
  <div class="legal-doc-links" data-testid="legal-doc-links">
    <p v-if="intro" class="text-body-sm intro">{{ intro }}</p>
    <ul class="link-list" :class="{ 'link-list--stacked': stacked }">
      <li>
        <a
          :href="privacyUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="legal-link"
          data-testid="legal-link-privacy"
        >
          {{ t("legal.privacyPolicy") }}
        </a>
      </li>
      <li>
        <a
          :href="termsUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="legal-link"
          data-testid="legal-link-terms"
        >
          {{ t("legal.termsOfUse") }}
        </a>
      </li>
      <li v-if="showOssLicensesLink">
        <router-link
          to="/licenses"
          class="legal-link"
          data-testid="settings-oss-licenses-link"
        >
          {{ t("settings.ossButton") }}
        </router-link>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { privacyPolicyUrl, termsOfUseUrl } from "../constants/legalUrls";

defineProps<{
  intro?: string;
  stacked?: boolean;
  showOssLicensesLink?: boolean;
}>();

const { locale, t } = useI18n();

// URL language: see legalDocLocale (ja only; other locales → EN legal docs).
const privacyUrl = computed(() => privacyPolicyUrl(locale.value));
const termsUrl = computed(() => termsOfUseUrl(locale.value));
</script>

<style scoped>
.intro {
  margin: 0 0 var(--space-block);
  color: var(--color-text-secondary);
  line-height: var(--line-height-relaxed);
}

.link-list {
  margin: 0;
  padding-left: 1.25rem;
  color: var(--color-text-secondary);
}

.legal-link {
  color: var(--color-brand);
  text-decoration: none;
}

.legal-link:hover {
  color: var(--color-brand-hover);
  text-decoration: underline;
}

.link-list--stacked {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: var(--space-inline-tight);
}
</style>
