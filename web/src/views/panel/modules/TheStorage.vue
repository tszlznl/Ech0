<template>
  <div class="w-full px-2">
    <!-- 存储设置 -->
    <TheStorageSetting />
    <!-- 文件列表 -->
    <TheStorageFileList />
  </div>
</template>

<script setup lang="ts">
import TheStorageSetting from './TheSetting/TheStorageSetting.vue'
import TheStorageFileList from './TheSetting/TheStorageFileList.vue'
</script>

<style scoped></style>
