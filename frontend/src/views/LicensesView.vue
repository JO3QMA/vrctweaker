<template>
  <div class="licenses-view">
    <h1 class="page-title">{{ t("licenses.title") }}</h1>
    <p class="text-body-sm intro">
      {{ t("licenses.intro") }}
    </p>

    <section class="licenses-section">
      <h2 class="text-h2 section-title">{{ t("licenses.frontend") }}</h2>
      <el-table
        :data="npmLicensesArray"
        class="licenses-table"
        style="width: 100%"
        size="small"
        stripe
      >
        <el-table-column
          prop="name"
          :label="t('licenses.colPackage')"
          min-width="200"
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
        <el-table-column :label="t('licenses.colLicense')" width="120">
          <template #default="{ row }">
            <VtTag size="small" variant="neutral">{{ row.licenses }}</VtTag>
          </template>
        </el-table-column>
        <el-table-column :label="t('licenses.colRepo')" min-width="200">
          <template #default="{ row }">
            <a
              v-if="row.repository"
              :href="row.repository"
              target="_blank"
              rel="noopener noreferrer"
              class="repo-link"
            >
              {{ truncateUrl(row.repository) }}
            </a>
            <span v-else class="text-caption repo-missing">{{
              t("common.dash")
            }}</span>
          </template>
        </el-table-column>
      </el-table>
    </section>

    <section class="licenses-section">
      <h2 class="text-h2 section-title">{{ t("licenses.backend") }}</h2>
      <el-table :data="goLicenses" style="width: 100%" size="small" stripe>
        <el-table-column
          prop="path"
          :label="t('licenses.colPackageGo')"
          min-width="220"
        >
          <template #default="{ row }">
            <span class="package-name">{{ row.path }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('licenses.colLicense')" width="120">
          <template #default="{ row }">
            <VtTag size="small" variant="neutral">{{ row.license }}</VtTag>
          </template>
        </el-table-column>
        <el-table-column :label="t('licenses.colRepo')" min-width="200">
          <template #default="{ row }">
            <a
              v-if="row.repository"
              :href="row.repository"
              target="_blank"
              rel="noopener noreferrer"
              class="repo-link"
            >
              {{ truncateUrl(row.repository) }}
            </a>
            <span v-else class="text-caption repo-missing">{{
              t("common.dash")
            }}</span>
          </template>
        </el-table-column>
      </el-table>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import VtTag from "../components/VtTag.vue";

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

function truncateUrl(url: string): string {
  try {
    const u = new URL(url);
    const path = u.pathname.replace(/\/$/, "");
    return path ? `${u.hostname}${path}` : u.hostname;
  } catch {
    return url;
  }
}
</script>

<style scoped>
.licenses-view {
  max-width: 900px;
}

.intro {
  display: block;
  margin: 0 0 var(--space-page);
  line-height: var(--line-height-relaxed);
  color: var(--color-text-secondary);
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
  color: var(--color-brand);
  text-decoration: none;
}

.repo-link:hover {
  color: var(--color-brand-hover);
  text-decoration: underline;
}

.repo-missing {
  color: var(--color-text-secondary);
}
</style>
