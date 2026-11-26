<template>
  <div class="licenses-view route-scroll-root">
    <div class="licenses-view__content">
      <h1 class="page-title">{{ t("licenses.title") }}</h1>
      <p class="text-body-sm intro-with-siblings">
        {{ t("licenses.intro") }}
      </p>
      <el-text
        type="info"
        size="small"
        class="hint block-hint unofficial-note"
        data-testid="licenses-unofficial-disclaimer"
      >
        {{ t("common.unofficialDisclaimer") }}
      </el-text>
      <LegalDocLinks :intro="t('legal.licensesIntro')" />

      <section class="licenses-section">
        <h2 class="text-h2 section-title">{{ t("licenses.frontend") }}</h2>
        <div
          class="licenses-table-scroll"
          data-testid="licenses-table-scroll-frontend"
        >
          <table
            class="licenses-table licenses-data-table"
            :aria-label="t('licenses.frontend')"
          >
            <thead>
              <tr>
                <th scope="col" class="licenses-table__package-col">
                  {{ t("licenses.colPackage") }}
                </th>
                <th scope="col" class="licenses-table__version-col">
                  {{ t("licenses.colVersion") }}
                </th>
                <th scope="col" class="licenses-table__license-col">
                  {{ t("licenses.colLicense") }}
                </th>
                <th scope="col" class="licenses-table__repo-col">
                  {{ t("licenses.colRepo") }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="npmLicensesArray.length === 0">
                <td colspan="4" class="licenses-table__empty">
                  {{ t("licenses.tableEmptyFrontend") }}
                </td>
              </tr>
              <tr
                v-for="row in npmLicensesArray"
                :key="`${row.name}@${row.version}`"
              >
                <td class="licenses-table__package-cell">
                  <span class="package-name">{{ row.name }}</span>
                </td>
                <td class="licenses-table__version-cell">{{ row.version }}</td>
                <td class="licenses-table__license-cell">
                  <VtTag size="small" variant="neutral">{{
                    row.licenses
                  }}</VtTag>
                </td>
                <td class="licenses-table__repo-cell">
                  <a
                    v-if="row.repository"
                    :href="row.repository"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="repo-link"
                  >
                    {{ row.repository }}
                  </a>
                  <span v-else class="text-caption repo-missing">{{
                    t("common.dash")
                  }}</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section class="licenses-section">
        <h2 class="text-h2 section-title">{{ t("licenses.backend") }}</h2>
        <div
          class="licenses-table-scroll"
          data-testid="licenses-table-scroll-backend"
        >
          <table
            class="licenses-table licenses-data-table"
            :aria-label="t('licenses.backend')"
          >
            <thead>
              <tr>
                <th scope="col" class="licenses-table__package-col">
                  {{ t("licenses.colPackageGo") }}
                </th>
                <th scope="col" class="licenses-table__license-col">
                  {{ t("licenses.colLicense") }}
                </th>
                <th scope="col" class="licenses-table__repo-col">
                  {{ t("licenses.colRepo") }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="goLicenses.length === 0">
                <td colspan="3" class="licenses-table__empty">
                  {{ t("licenses.tableEmptyBackend") }}
                </td>
              </tr>
              <tr v-for="row in goLicenses" :key="row.path">
                <td class="licenses-table__package-cell">
                  <span class="package-name">{{ row.path }}</span>
                </td>
                <td class="licenses-table__license-cell">
                  <VtTag size="small" variant="neutral">{{
                    row.license
                  }}</VtTag>
                </td>
                <td class="licenses-table__repo-cell">
                  <a
                    v-if="row.repository"
                    :href="row.repository"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="repo-link"
                  >
                    {{ row.repository }}
                  </a>
                  <span v-else class="text-caption repo-missing">{{
                    t("common.dash")
                  }}</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import VtTag from "../components/VtTag.vue";
import LegalDocLinks from "../components/LegalDocLinks.vue";

const { t } = useI18n();
import npmLicensesData from "../data/licenses.json";
import { goLicenses } from "../data/go-licenses";

interface NpmLicenseEntry {
  licenses: string;
  repository?: string;
}

type NpmLicensesRecord = Record<string, NpmLicenseEntry>;

const npmLicensesArray = computed(() => {
  return Object.entries(npmLicensesData as NpmLicensesRecord)
    .filter(([key]) => !key.startsWith("vrchat-tweaker-frontend@"))
    .map(([key, info]) => {
      const atIdx = key.lastIndexOf("@");
      const name = atIdx >= 0 ? key.slice(0, atIdx) : key;
      const version = atIdx >= 0 ? key.slice(atIdx + 1) : "";
      return {
        name,
        version,
        licenses: info.licenses || "-",
        repository: info.repository,
      };
    })
    .sort((a, b) => a.name.localeCompare(b.name));
});
</script>

<style scoped>
.licenses-view__content {
  max-width: 900px;
}

.intro-with-siblings {
  display: block;
  margin: 0 0 var(--space-block);
  line-height: var(--line-height-relaxed);
  color: var(--color-text-secondary);
}

.unofficial-note {
  display: block;
  margin-bottom: var(--space-block);
  max-width: 42rem;
}

.licenses-view :deep(.legal-doc-links) {
  margin-bottom: var(--space-section);
}

.licenses-section {
  margin-bottom: var(--space-48);
}

.section-title {
  margin: 0 0 var(--space-block);
  color: var(--color-text-primary);
}

.licenses-table-scroll {
  width: 100%;
  max-width: 100%;
  overflow-x: auto;
}

.licenses-data-table {
  width: 100%;
  min-width: max-content;
  border-collapse: collapse;
  table-layout: auto;
  font-size: var(--font-size-14);
  line-height: var(--line-height-normal);
  color: var(--color-text-primary);
}

.licenses-data-table th,
.licenses-data-table td {
  padding: var(--space-8) var(--space-12);
  text-align: left;
  vertical-align: middle;
  border-bottom: 1px solid var(--color-border);
}

.licenses-data-table thead th {
  background: var(--color-bg-elevated);
  color: var(--color-text-secondary);
  font-weight: var(--font-weight-600);
  white-space: nowrap;
}

.licenses-data-table tbody tr:nth-child(odd) {
  background: var(--color-bg-base);
}

.licenses-data-table tbody tr:nth-child(even) {
  background: var(--color-bg-elevated);
}

.licenses-table__package-cell,
.licenses-table__repo-cell,
.licenses-table__version-cell {
  white-space: nowrap;
}

.licenses-table__license-cell {
  white-space: nowrap;
}

.licenses-table__license-cell :deep(.el-tag) {
  --el-tag-bg-color: var(--color-bg-muted);
  --el-tag-border-color: var(--color-border);
  --el-tag-text-color: var(--color-text-secondary);
  display: inline-flex;
  vertical-align: middle;
}

.package-name {
  font-family: ui-monospace, monospace;
  font-size: var(--font-size-14);
  white-space: nowrap;
}

.repo-link {
  color: var(--color-link);
  text-decoration: none;
  white-space: nowrap;
}

.repo-link:hover {
  color: var(--color-link-hover);
  text-decoration: underline;
}

.repo-missing {
  color: var(--color-text-secondary);
}

.licenses-table__empty {
  text-align: center;
  color: var(--color-text-secondary);
  padding: var(--space-24);
}
</style>
