<script setup lang="ts">
defineOptions({
  inheritAttrs: false,
});
</script>

<template>
  <el-checkbox v-bind="$attrs">
    <template v-for="(_, slotName) in $slots" #[slotName]="slotProps">
      <slot :name="slotName" v-bind="slotProps ?? {}" />
    </template>
  </el-checkbox>
</template>
