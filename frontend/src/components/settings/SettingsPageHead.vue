<script setup lang="ts">
defineProps<{
  title: string;
  description: string;
}>();
</script>

<template>
  <header class="settings-page-head">
    <h2 class="settings-page-head__title text-h2">{{ title }}</h2>
    <p class="settings-page-head__lead text-body-sm">{{ description }}</p>
  </header>
</template>
