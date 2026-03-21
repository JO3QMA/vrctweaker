<template>
  <nav class="sidebar">
    <ul class="sidebar-menu">
      <li v-for="item in menuItems" :key="item.path">
        <router-link :to="item.path" class="sidebar-link" active-class="active">
          <span class="sidebar-icon">{{ item.icon }}</span>
          <span class="sidebar-label">{{ item.label }}</span>
        </router-link>
      </li>
    </ul>
    <div class="sidebar-footer">
      <router-link to="/settings" class="sidebar-link settings-link">
        <span class="sidebar-icon">⚙️</span>
        <span class="sidebar-label">設定</span>
      </router-link>
    </div>
  </nav>
</template>

<script setup lang="ts">
const menuItems = [
  { path: "/", icon: "🏠", label: "ダッシュボード" },
  { path: "/launcher", icon: "🚀", label: "ランチャー" },
  { path: "/gallery", icon: "🖼️", label: "ギャラリー" },
  { path: "/activity", icon: "📊", label: "アクティビティ" },
  { path: "/friends", icon: "👥", label: "フレンド" },
  { path: "/automation", icon: "🤖", label: "オートメーション" },
  { path: "/config", icon: "📝", label: "その他の設定" },
];
</script>

<style scoped>
.sidebar {
  width: var(--sidebar-width);
  background: var(--bg-secondary);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
}

.sidebar-menu {
  list-style: none;
  margin: 0;
  padding: 0.5rem 0;
}

.sidebar-link {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.6rem 1rem;
  color: var(--text-secondary);
  transition:
    background 0.15s,
    color 0.15s;
}

.sidebar-link:hover,
.sidebar-link.active {
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.sidebar-link.active {
  border-left: 3px solid var(--accent);
}

.sidebar-icon {
  font-size: 1.1rem;
}

.sidebar-label {
  font-size: 0.9rem;
}

.sidebar-footer {
  margin-top: auto;
  padding: 0.5rem 0;
  border-top: 1px solid var(--border);
}

.settings-link {
  color: var(--text-secondary);
}
</style>
