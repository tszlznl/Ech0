<template>
  <div class="w-full px-2 space-y-3">
    <!-- 数据迁移 -->
    <TheMigrationSetting />

    <!-- 数据管理 -->
    <TheBackupSetting />

    <!-- 快照设置 -->
    <TheBackupScheduleSetting />
  </div>
</template>

<script setup lang="ts">
import TheBackupSetting from './TheSetting/TheBackupSetting.vue'
import TheBackupScheduleSetting from './TheSetting/TheBackupScheduleSetting.vue'
import TheMigrationSetting from './TheSetting/TheMigrationSetting.vue'
</script>

<style scoped></style>
