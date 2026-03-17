<template>
  <div class="launcher-view">
    <h1 class="page-title">
      ランチャー
    </h1>
    <div class="profiles-section">
      <div class="profiles-list">
        <button
          class="btn-add"
          @click="addNew"
        >
          + 新規プロファイル
        </button>
        <div
          v-for="p in profiles"
          :key="p.id"
          class="profile-card"
          :class="{ active: selected?.id === p.id }"
          @click="select(p)"
        >
          <span class="profile-name">{{ p.name }}</span>
          <span
            v-if="p.isDefault"
            class="badge"
          >既定</span>
        </div>
      </div>
      <div
        v-if="selected"
        class="profile-editor"
      >
        <label>プロファイル名</label>
        <input
          v-model="selected.name"
          type="text"
        >
        <label>起動引数</label>
        <input
          v-model="selected.arguments"
          type="text"
          placeholder="--no-vr --clear-cache"
        >
        <label>
          <input
            v-model="selected.isDefault"
            type="checkbox"
          >
          デフォルトに設定
        </label>
        <div class="editor-actions">
          <button
            class="btn-save"
            @click="save"
          >
            保存
          </button>
          <button
            class="btn-launch"
            @click="launch"
          >
            この設定で起動
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { App, type LaunchProfileDTO } from '../wails/app'

const profiles = ref<LaunchProfileDTO[]>([])
const selected = ref<LaunchProfileDTO | null>(null)

onMounted(async () => {
  profiles.value = await App.launchProfiles()
  if (profiles.value.length > 0 && !selected.value) {
    selected.value = profiles.value.find(p => p.isDefault) ?? profiles.value[0]
  }
})

function select(p: LaunchProfileDTO) {
  selected.value = { ...p }
}

function addNew() {
  selected.value = {
    id: '',
    name: '新しいプロファイル',
    arguments: '',
    isDefault: profiles.value.length === 0,
  }
}

async function save() {
  if (!selected.value) return
  await App.saveLaunchProfile(selected.value)
  profiles.value = await App.launchProfiles()
}

async function launch() {
  if (!selected.value) return
  await App.launchVRChat(selected.value.id)
}
</script>

<style scoped>
.profiles-section { display: flex; gap: 1.5rem; }
.profiles-list { width: 240px; }
.btn-add { width: 100%; padding: 0.5rem; margin-bottom: 0.5rem; background: var(--bg-tertiary); border: 1px dashed var(--border); border-radius: var(--radius); color: var(--text-secondary); cursor: pointer; }
.btn-add:hover { background: var(--bg-secondary); color: var(--accent); }
.profile-card {
  padding: 0.75rem;
  margin-bottom: 0.5rem;
  background: var(--bg-secondary);
  border-radius: var(--radius);
  cursor: pointer;
  display: flex; align-items: center; gap: 0.5rem;
}
.profile-card:hover, .profile-card.active { background: var(--bg-tertiary); }
.badge { font-size: 0.7rem; color: var(--accent); }
.profile-editor { flex: 1; }
.profile-editor label { display: block; margin: 0.5rem 0 0.2rem; font-size: 0.85rem; }
.profile-editor input[type="text"] {
  width: 100%; padding: 0.5rem; background: var(--bg-tertiary); border: 1px solid var(--border);
  border-radius: var(--radius); color: var(--text-primary);
}
.editor-actions { margin-top: 1rem; display: flex; gap: 0.5rem; }
.btn-save, .btn-launch { padding: 0.5rem 1rem; border-radius: var(--radius); border: none; }
.btn-save { background: var(--bg-tertiary); color: var(--text-primary); }
.btn-launch { background: var(--accent); color: white; }
</style>
