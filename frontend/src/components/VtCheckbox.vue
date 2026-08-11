<script setup lang="ts">
defineOptions({
  inheritAttrs: false,
});
</script>

<template>
  <el-checkbox v-bind="$attrs">
    <slot />
  </el-checkbox>
</template>
