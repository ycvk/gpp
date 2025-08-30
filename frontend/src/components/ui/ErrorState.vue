<template>
  <div class="error-state">
    <div class="error-icon">
      <i :class="iconClass" />
    </div>
    <div class="error-content">
      <h3 class="error-title">{{ title }}</h3>
      <p class="error-message">{{ message }}</p>
      <div v-if="showActions" class="error-actions">
        <n-button 
          v-if="showRetry"
          type="primary"
          size="small"
          :loading="retrying"
          @click="handleRetry"
        >
          重试
        </n-button>
        <n-button 
          v-if="showHelp"
          type="tertiary"
          size="small"
          @click="handleHelp"
        >
          获取帮助
        </n-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  type?: 'connection' | 'network' | 'config' | 'unknown'
  title?: string
  message?: string
  showRetry?: boolean
  showHelp?: boolean
  showActions?: boolean
  retrying?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  type: 'unknown',
  showRetry: true,
  showHelp: true,
  showActions: true,
  retrying: false
})

const emit = defineEmits<{
  retry: []
  help: []
}>()

const iconClass = computed(() => {
  switch (props.type) {
    case 'connection':
      return 'i-ionicons-wifi-outline text-danger'
    case 'network':
      return 'i-ionicons-cloud-offline-outline text-warning'
    case 'config':
      return 'i-ionicons-settings-outline text-primary'
    default:
      return 'i-ionicons-warning-outline text-danger'
  }
})

const title = computed(() => {
  if (props.title) return props.title
  
  switch (props.type) {
    case 'connection':
      return '连接失败'
    case 'network':
      return '网络异常'
    case 'config':
      return '配置错误'
    default:
      return '出现错误'
  }
})

const message = computed(() => {
  if (props.message) return props.message
  
  switch (props.type) {
    case 'connection':
      return '无法连接到代理服务器，请检查节点设置'
    case 'network':
      return '网络连接不稳定，请检查网络设置'
    case 'config':
      return '配置信息有误，请重新设置'
    default:
      return '发生了未知错误，请重试'
  }
})

const handleRetry = () => {
  emit('retry')
}

const handleHelp = () => {
  emit('help')
}
</script>

<style scoped>
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: var(--space-8);
  color: var(--color-neutral-600);
}

.error-icon {
  margin-bottom: var(--space-4);
  font-size: 48px;
}

.error-content {
  max-width: 240px;
}

.error-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-neutral-900);
  margin: 0 0 var(--space-2);
  line-height: var(--line-height-tight);
}

.error-message {
  font-size: var(--font-size-sm);
  color: var(--color-neutral-600);
  margin: 0 0 var(--space-6);
  line-height: var(--line-height-normal);
}

.error-actions {
  display: flex;
  gap: var(--space-3);
  justify-content: center;
}
</style>