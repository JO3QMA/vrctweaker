<template>
  <div class="legal-doc-links" data-testid="legal-doc-links">
    <p v-if="intro" class="text-body-sm intro">{{ intro }}</p>
    <ul class="link-list" :class="{ 'link-list--stacked': stacked }">
      <li class="legal-doc-item">
        <component
          :is="privacyLinkComponent"
          v-bind="privacyLinkBind"
          class="legal-link"
          data-testid="legal-link-privacy"
          @click="onPrivacyClick"
        >
          {{ t("legal.privacyPolicy") }}
        </component>
        <SafeExternalUrlLink
          v-if="showGitHubSupplement"
          :url="privacyUrl"
          test-id="legal-github-privacy"
          class="github-supplement"
        >
          {{ t("legal.viewOnGitHub") }}
        </SafeExternalUrlLink>
      </li>
      <li class="legal-doc-item">
        <component
          :is="termsLinkComponent"
          v-bind="termsLinkBind"
          class="legal-link"
          data-testid="legal-link-terms"
          @click="onTermsClick"
        >
          {{ t("legal.termsOfUse") }}
        </component>
        <SafeExternalUrlLink
          v-if="showGitHubSupplement"
          :url="termsUrl"
          test-id="legal-github-terms"
          class="github-supplement"
        >
          {{ t("legal.viewOnGitHub") }}
        </SafeExternalUrlLink>
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
    <LegalDocumentDialog v-model="dialogOpen" :kind="dialogKind" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";
import LegalDocumentDialog from "./LegalDocumentDialog.vue";
import SafeExternalUrlLink from "./SafeExternalUrlLink.vue";
import { privacyPolicyUrl, termsOfUseUrl } from "../constants/legalUrls";
import type { LegalDocKind } from "@/data/legalDocuments";

const props = withDefaults(
  defineProps<{
    intro?: string;
    stacked?: boolean;
    showOssLicensesLink?: boolean;
    /** `route`: in-app page; `dialog`: modal (e.g. onboarding wizard). */
    presentation?: "route" | "dialog";
    /** Hide per-item GitHub links when the panel already shows one (dialog mode). */
    showGitHubSupplement?: boolean;
  }>(),
  {
    intro: undefined,
    stacked: false,
    showOssLicensesLink: false,
    presentation: "route",
    showGitHubSupplement: true,
  },
);

const { locale, t } = useI18n();

const privacyUrl = computed(() => privacyPolicyUrl(locale.value));
const termsUrl = computed(() => termsOfUseUrl(locale.value));

const dialogOpen = ref(false);
const dialogKind = ref<LegalDocKind | null>(null);

const useDialog = computed(() => props.presentation === "dialog");

const privacyLinkComponent = computed(() =>
  useDialog.value ? "button" : RouterLink,
);
const termsLinkComponent = computed(() =>
  useDialog.value ? "button" : RouterLink,
);

const privacyLinkBind = computed(() =>
  useDialog.value
    ? { type: "button" as const }
    : { to: "/legal/privacy" as const },
);
const termsLinkBind = computed(() =>
  useDialog.value
    ? { type: "button" as const }
    : { to: "/legal/terms" as const },
);

function openDialog(kind: LegalDocKind): void {
  dialogKind.value = kind;
  dialogOpen.value = true;
}

function onPrivacyClick(event: MouseEvent): void {
  if (!useDialog.value) {
    return;
  }
  event.preventDefault();
  openDialog("privacy");
}

function onTermsClick(event: MouseEvent): void {
  if (!useDialog.value) {
    return;
  }
  event.preventDefault();
  openDialog("terms");
}
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
  background: none;
  border: none;
  padding: 0;
  font: inherit;
  cursor: pointer;
  text-align: left;
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

.legal-doc-item {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
}

.github-supplement {
  font-size: var(--font-size-12);
  color: var(--color-text-secondary);
}
</style>
