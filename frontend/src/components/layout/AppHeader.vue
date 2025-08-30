<template>
  <header class="app-header">
    <div class="header-content">
      <!-- 左侧 - 应用标题 -->
      <div class="header-left">
        <div class="app-title">
          <i class="i-ionicons-rocket text-primary" />
          <span>GPP 加速器</span>
        </div>
      </div>
      
      <!-- 中间 - 状态指示 -->
      <div class="header-center">
        <div 
          :class="[
            'connection-indicator',
            `connection-indicator--${connectionStatus}`
          ]"
        >
          <div class="indicator-dot"></div>
          <span class="indicator-text">{{ statusText }}</span>
        </div>
      </div>
      
      <!-- 右侧 - 操作按钮 -->
      <div class="header-right">
        <n-button
          type="tertiary"
          size="small"
          circle
          @click="$router.push('/settings')"
        >
          <template #icon>
            <i class="i-ionicons-settings-outline" />
          </template>
        </n-button>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
interface Props {
  connectionStatus?: 'idle' | 'connecting' | 'connected' | 'error'
  isConnected?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  connectionStatus: 'idle',
  isConnected: false
})

const statusText = computed(() => {
  switch (props.connectionStatus) {
    case 'idle':
      return '待机'
    case 'connecting':
      return '连接中'
    case 'connected':
      return '已连接'
    case 'error':
      return '错误'
    default:
      return '未知'
  }
})
</script>

<style scoped>
.app-header {
  background: var(--bg-elevated);
  border-bottom: 1px solid var(--color-neutral-200);
  padding: var(--space-3) var(--space-4);
  flex-shrink: 0;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  max-width: 100%;
}

.header-left,
.header-center,
.header-right {
  flex: 1;
  display: flex;
  align-items: center;
}

.header-center {
  justify-content: center;
}

.header-right {
  justify-content: flex-end;
}

.app-title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-semibold);
  color: var(--color-neutral-900);
}

.app-title i {
  font-size: 18px;
}

.connection-indicator {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-1) var(--space-3);
  background: var(--color-neutral-50);
  border-radius: var(--radius-full);
  transition: var(--transition-base);
}

.indicator-dot {
  width: 8px;
  height: 8px;
  border-radius: var(--radius-full);
  transition: var(--transition-base);
}

.connection-indicator--idle .indicator-dot {
  background: var(--color-neutral-400);
}

.connection-indicator--connecting .indicator-dot {
  background: var(--color-warning);
  animation: pulse 1.5s infinite;
}

.connection-indicator--connected .indicator-dot {
  background: var(--color-success);
}

.connection-indicator--error .indicator-dot {
  background: var(--color-danger);
}

.indicator-text {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  color: var(--color-neutral-600);
}

@keyframes pulse {
  0% {
    transform: scale(0.8);
    opacity: 0.6;
  }
  50% {
    transform: scale(1.2);
    opacity: 0.3;
  }
  100% {
    transform: scale(1.4);
    opacity: 0;
  }
}
</style>