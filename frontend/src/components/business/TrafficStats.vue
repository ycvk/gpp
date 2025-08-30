<template>
  <div class="traffic-stats">
    <div class="stats-header">
      <h4 class="stats-title">流量统计</h4>
      <div class="stats-controls">
        <n-button 
          size="tiny"
          type="tertiary"
          @click="handleReset"
        >
          重置
        </n-button>
      </div>
    </div>
    
    <div class="stats-grid">
      <!-- 实时速度 -->
      <div class="stat-item">
        <div class="stat-icon">
          <i class="i-ionicons-arrow-up text-success" />
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ formatSpeed(uploadSpeed) }}</div>
          <div class="stat-label">上传</div>
        </div>
      </div>
      
      <div class="stat-item">
        <div class="stat-icon">
          <i class="i-ionicons-arrow-down text-primary" />
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ formatSpeed(downloadSpeed) }}</div>
          <div class="stat-label">下载</div>
        </div>
      </div>
      
      <!-- 总流量 -->
      <div class="stat-item stat-item--total">
        <div class="stat-icon">
          <i class="i-ionicons-stats-chart" />
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ formatBytes(totalUpload + totalDownload) }}</div>
          <div class="stat-label">总流量</div>
        </div>
      </div>
    </div>
    
    <!-- 可选的迷你图表 -->
    <div v-if="showChart" class="mini-chart">
      <!-- 这里可以集成简单的图表组件 -->
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  uploadSpeed: number
  downloadSpeed: number
  totalUpload: number
  totalDownload: number
  showChart?: boolean
}

withDefaults(defineProps<Props>(), {
  showChart: false
})

const emit = defineEmits<{
  reset: []
}>()

// 格式化速度显示
const formatSpeed = (bytesPerSecond: number): string => {
  if (bytesPerSecond === 0) return '0 B/s'
  
  const units = ['B/s', 'KB/s', 'MB/s', 'GB/s']
  let value = bytesPerSecond
  let unitIndex = 0
  
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex++
  }
  
  return `${value.toFixed(1)} ${units[unitIndex]}`
}

// 格式化字节显示
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let value = bytes
  let unitIndex = 0
  
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex++
  }
  
  return `${value.toFixed(2)} ${units[unitIndex]}`
}

const handleReset = () => {
  emit('reset')
}
</script>

<style scoped>
.traffic-stats {
  background: var(--bg-secondary);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-md);
  padding: var(--space-4);
}

.stats-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-4);
}

.stats-title {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-neutral-800);
  margin: 0;
}

.stats-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-4);
}

.stat-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2);
  background: var(--bg-elevated);
  border-radius: var(--radius-sm);
  transition: var(--transition-base);
}

.stat-item--total {
  grid-column: 1 / -1;
  background: linear-gradient(135deg, var(--color-primary), var(--color-primary-hover));
  color: white;
}

.stat-item:hover {
  transform: translateY(-1px);
  box-shadow: var(--shadow-sm);
}

.stat-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: var(--color-neutral-100);
  border-radius: var(--radius-sm);
  flex-shrink: 0;
}

.stat-item--total .stat-icon {
  background: rgba(255, 255, 255, 0.2);
  color: white;
}

.stat-content {
  flex: 1;
  min-width: 0;
}

.stat-value {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-bold);
  color: var(--color-neutral-900);
  line-height: var(--line-height-tight);
}

.stat-item--total .stat-value {
  color: white;
}

.stat-label {
  font-size: var(--font-size-xs);
  color: var(--color-neutral-500);
  line-height: var(--line-height-normal);
}

.stat-item--total .stat-label {
  color: rgba(255, 255, 255, 0.8);
}
</style>