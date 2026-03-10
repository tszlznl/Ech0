<template>
  <div class="w-full px-2">
    <div class="empty-state">
      <h2 class="empty-title">Dashboard 指标模块已下线</h2>
      <p class="empty-desc">当前版本暂不提供系统指标展示，你可以前往“系统日志”查看运行信息。</p>
    </div>
  </div>
</template>

<style scoped>
.empty-state {
  min-height: 18rem;
  border: 1px dashed var(--border-color-300);
  border-radius: 0.75rem;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  text-align: center;
  padding: 1.5rem;
  gap: 0.75rem;
}

.empty-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--text-color-next-500);
}

.empty-desc {
  max-width: 32rem;
  font-size: 0.95rem;
  color: var(--text-color-next-400);
}
</style>
