<template>
  <div class="licenses-view">
    <h1 class="page-title">OSS ライセンス</h1>
    <p class="intro">
      本アプリケーションで使用しているオープンソースソフトウェア（OSS）のライセンス一覧です。
    </p>
    <section class="licenses-section">
      <h2>フロントエンド（npm）</h2>
      <div class="licenses-table-wrapper">
        <table class="licenses-table">
          <thead>
            <tr>
              <th>パッケージ名</th>
              <th>バージョン</th>
              <th>ライセンス</th>
              <th>リポジトリ</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="[key, info] in npmLicenses" :key="key">
              <td class="package-name">
                {{ info.name }}
              </td>
              <td>{{ info.version }}</td>
              <td>
                <span class="license-badge">{{ info.licenses }}</span>
              </td>
              <td>
                <a
                  v-if="info.repository"
                  :href="info.repository"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="repo-link"
                >
                  {{ truncateUrl(info.repository) }}
                </a>
                <span v-else class="no-repo">-</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
    <section class="licenses-section">
      <h2>バックエンド（Go）</h2>
      <div class="licenses-table-wrapper">
        <table class="licenses-table">
          <thead>
            <tr>
              <th>パッケージ</th>
              <th>ライセンス</th>
              <th>リポジトリ</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="pkg in goLicenses" :key="pkg.path">
              <td class="package-name">
                {{ pkg.path }}
              </td>
              <td>
                <span class="license-badge">{{ pkg.license }}</span>
              </td>
              <td>
                <a
                  v-if="pkg.repository"
                  :href="pkg.repository"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="repo-link"
                >
                  {{ truncateUrl(pkg.repository) }}
                </a>
                <span v-else class="no-repo">-</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import npmLicensesData from "../data/licenses.json";
import { goLicenses } from "../data/go-licenses";

interface NpmLicenseEntry {
  licenses: string;
  repository?: string;
}

type NpmLicensesRecord = Record<string, NpmLicenseEntry>;

const npmLicenses = computed(() => {
  const entries = Object.entries(npmLicensesData as NpmLicensesRecord)
    .filter(([key]) => !key.startsWith("vrchat-tweaker-frontend@"))
    .map(([key, info]) => {
      const atIdx = key.lastIndexOf("@");
      const name = atIdx >= 0 ? key.slice(0, atIdx) : key;
      const version = atIdx >= 0 ? key.slice(atIdx + 1) : "";
      return [
        key,
        {
          name,
          version,
          licenses: info.licenses || "-",
          repository: info.repository,
        },
      ] as const;
    })
    .sort((a, b) => a[1].name.localeCompare(b[1].name));
  return entries;
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
.page-title {
  font-size: 1.5rem;
  margin: 0 0 0.5rem;
}
.intro {
  font-size: 0.95rem;
  color: var(--text-secondary);
  margin: 0 0 2rem;
  line-height: 1.6;
}
.licenses-section {
  margin-bottom: 2.5rem;
}
.licenses-section h2 {
  font-size: 1.1rem;
  margin: 0 0 1rem;
  color: var(--text-primary);
}
.licenses-table-wrapper {
  overflow-x: auto;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: var(--bg-tertiary);
}
.licenses-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}
.licenses-table th,
.licenses-table td {
  padding: 0.6rem 1rem;
  text-align: left;
  border-bottom: 1px solid var(--border);
}
.licenses-table th {
  font-weight: 600;
  color: var(--text-secondary);
  background: var(--bg-secondary);
}
.licenses-table tbody tr:last-child td {
  border-bottom: none;
}
.licenses-table tbody tr:hover {
  background: rgba(255, 255, 255, 0.03);
}
.package-name {
  font-family: ui-monospace, monospace;
  font-size: 0.85rem;
}
.license-badge {
  display: inline-block;
  padding: 0.2rem 0.5rem;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: 4px;
  font-size: 0.8rem;
}
.repo-link {
  color: var(--accent);
  text-decoration: none;
}
.repo-link:hover {
  color: var(--accent-hover);
  text-decoration: underline;
}
.no-repo {
  color: var(--text-secondary);
}
</style>
