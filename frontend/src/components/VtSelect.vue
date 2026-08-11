<script setup lang="ts">
defineOptions({
  inheritAttrs: false,
});
</script>

<template>
  <el-select v-bind="$attrs">
    <slot />
  </el-select>
</template>
