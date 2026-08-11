<script setup lang="ts">
defineOptions({
  inheritAttrs: false,
});
</script>

<template>
  <el-switch v-bind="$attrs">
    <slot />
  </el-switch>
</template>
