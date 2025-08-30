<template>
  <div class="app-layout">
    <!-- 顶部状态栏 -->
    <AppHeader 
      :connection-status="connectionStatus"
      :is-connected="isConnected"
    />
    
    <!-- 主内容区域 -->
    <main class="main-content">
      <slot />
    </main>
    
    <!-- 底部操作区 -->
    <AppFooter 
      v-if="showFooter"
      :is-loading="isLoading"
      @primary-action="handlePrimaryAction"
    />
  </div>
</template>

<script setup lang="ts">
import AppHeader from './AppHeader.vue'

interface Props {
  connectionStatus?: 'idle' | 'connecting' | 'connected' | 'error'
  isConnected?: boolean
  showFooter?: boolean
  isLoading?: boolean
}

withDefaults(defineProps<Props>(), {
  connectionStatus: 'idle',
  isConnected: false,
  showFooter: false,
  isLoading: false
})

const emit = defineEmits<{
  primaryAction: []
}>()

const handlePrimaryAction = () => {
  emit('primaryAction')
}
</script>

<style scoped>
.app-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: var(--bg-primary);
}

.main-content {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-6) var(--space-5);
  min-height: 0; /* 防止flex子元素撑开 */
}
</style>