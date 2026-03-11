<template>
  <div class="w-full px-2">
    <!-- 数据迁移 -->
    <TheMigrationSetting class="mb-3" />

    <!-- 数据管理 -->
    <TheBackupSetting class="mb-3" />

    <!-- 备份计划设置 -->
    <TheBackupScheduleSetting />
  </div>
</template>

<script setup lang="ts">
import TheBackupSetting from './TheSetting/TheBackupSetting.vue'
import TheBackupScheduleSetting from './TheSetting/TheBackupScheduleSetting.vue'
import TheMigrationSetting from './TheSetting/TheMigrationSetting.vue'
</script>

<style scoped></style>
