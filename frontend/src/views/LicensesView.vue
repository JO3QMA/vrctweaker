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
          <div class="licenses-table-scroll__track">
            <el-table
              :data="npmLicensesArray"
              class="licenses-table"
              size="small"
              stripe
              :fit="false"
              :aria-label="t('licenses.frontend')"
              :empty-text="t('licenses.tableEmptyFrontend')"
            >
              <el-table-column
                prop="name"
                :label="t('licenses.colPackage')"
                min-width="200"
                :show-overflow-tooltip="false"
                class-name="licenses-table__package-cell"
              >
                <template #default="{ row }">
                  <span class="package-name">{{ row.name }}</span>
                </template>
              </el-table-column>
              <el-table-column
                prop="version"
                :label="t('licenses.colVersion')"
                width="100"
              />
              <el-table-column
                :label="t('licenses.colLicense')"
                width="120"
                align="left"
                class-name="licenses-table__license-cell"
              >
                <template #default="{ row }">
                  <VtTag size="small" variant="neutral">{{
                    row.licenses
                  }}</VtTag>
                </template>
              </el-table-column>
              <el-table-column
                :label="t('licenses.colRepo')"
                min-width="200"
                :show-overflow-tooltip="false"
                class-name="licenses-table__repo-cell"
              >
                <template #default="{ row }">
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
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </section>

      <section class="licenses-section">
        <h2 class="text-h2 section-title">{{ t("licenses.backend") }}</h2>
        <div
          class="licenses-table-scroll"
          data-testid="licenses-table-scroll-backend"
        >
          <div class="licenses-table-scroll__track">
            <el-table
              :data="goLicenses"
              class="licenses-table"
              size="small"
              stripe
              :fit="false"
              :aria-label="t('licenses.backend')"
              :empty-text="t('licenses.tableEmptyBackend')"
            >
              <el-table-column
                prop="path"
                :label="t('licenses.colPackageGo')"
                min-width="220"
                :show-overflow-tooltip="false"
                class-name="licenses-table__package-cell"
              >
                <template #default="{ row }">
                  <span class="package-name">{{ row.path }}</span>
                </template>
              </el-table-column>
              <el-table-column
                :label="t('licenses.colLicense')"
                width="120"
                align="left"
                class-name="licenses-table__license-cell"
              >
                <template #default="{ row }">
                  <VtTag size="small" variant="neutral">{{
                    row.license
                  }}</VtTag>
                </template>
              </el-table-column>
              <el-table-column
                :label="t('licenses.colRepo')"
                min-width="200"
                :show-overflow-tooltip="false"
                class-name="licenses-table__repo-cell"
              >
                <template #default="{ row }">
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
                </template>
              </el-table-column>
            </el-table>
          </div>
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

/* Tighter than other views' intro: unofficial note + legal links follow immediately. */
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

.package-name {
  font-family: ui-monospace, monospace;
  font-size: var(--font-size-14);
}

.repo-link {
  color: var(--color-link);
  text-decoration: none;
}

.repo-link:hover {
  color: var(--color-link-hover);
  text-decoration: underline;
}

.repo-missing {
  color: var(--color-text-secondary);
}

.licenses-table {
  width: 100%;
}

.licenses-table :deep(.licenses-table__license-cell .cell) {
  text-align: left;
}

.licenses-table-scroll {
  width: 100%;
  max-width: 100%;
}

@media (max-width: 768px) {
  .licenses-table-scroll {
    overflow-x: auto;
  }

  .licenses-table-scroll__track {
    width: max-content;
    min-width: 100%;
  }

  .licenses-table-scroll .licenses-table {
    width: auto;
    min-width: 100%;
  }

  .licenses-table :deep(.el-table__inner-wrapper),
  .licenses-table :deep(.el-table__header-wrapper),
  .licenses-table :deep(.el-table__body-wrapper) {
    overflow: visible;
  }

  .licenses-table :deep(.el-table__body),
  .licenses-table :deep(.el-table__header),
  .licenses-table :deep(table) {
    width: max-content;
    min-width: 100%;
    table-layout: auto;
  }

  .licenses-table :deep(td.licenses-table__package-cell),
  .licenses-table :deep(td.licenses-table__repo-cell) {
    width: max-content;
    max-width: none;
    overflow: visible;
  }

  .licenses-table :deep(.licenses-table__package-cell .cell),
  .licenses-table :deep(.licenses-table__repo-cell .cell),
  .licenses-table :deep(.licenses-table__license-cell .cell) {
    overflow: visible;
    text-overflow: clip;
    white-space: nowrap;
    max-width: none;
    width: max-content;
    min-width: 100%;
  }

  .licenses-table-scroll .package-name,
  .licenses-table-scroll .repo-link {
    display: inline-block;
    max-width: none;
    overflow: visible;
    text-overflow: clip;
    white-space: nowrap;
  }
}
</style>
