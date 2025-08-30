<template>
  <footer v-if="showFooter" class="app-footer">
    <div class="footer-content">
      <!-- 主要操作按钮 -->
      <n-button
        v-if="showPrimaryButton" 
        :type="primaryButtonType"
        :loading="isLoading"
        :disabled="disabled"
        size="large"
        block
        @click="handlePrimaryAction"
      >
        <template #icon v-if="primaryIcon">
          <i :class="primaryIcon" />
        </template>
        {{ primaryButtonText }}
      </n-button>
      
      <!-- 次要操作区域 -->
      <div v-if="showSecondaryActions" class="secondary-actions">
        <slot name="secondary-actions" />
      </div>
      
      <!-- 额外信息区域 -->
      <div v-if="showExtraInfo" class="extra-info">
        <slot name="extra-info" />
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
interface Props {
  showFooter?: boolean      // v-if控制显示
  isLoading?: boolean       // 加载状态
  showPrimaryButton?: boolean
  primaryButtonText?: string
  primaryButtonType?: 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error'
  primaryIcon?: string
  disabled?: boolean
  showSecondaryActions?: boolean
  showExtraInfo?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  showFooter: true,
  isLoading: false,
  showPrimaryButton: false,
  primaryButtonText: '操作',
  primaryButtonType: 'primary',
  disabled: false,
  showSecondaryActions: false,
  showExtraInfo: false
})

const emit = defineEmits<{
  'primary-action': []      // 必须匹配AppLayout.vue的期望
}>()

const handlePrimaryAction = () => {
  if (!props.disabled && !props.isLoading) {
    emit('primary-action')
  }
}
</script>

<style scoped>
.app-footer {
  background: var(--bg-elevated);
  border-top: 1px solid var(--color-neutral-200);
  padding: var(--space-4);
  flex-shrink: 0;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.04);
}

.footer-content {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  max-width: 100%;
}

.secondary-actions {
  display: flex;
  justify-content: space-between;
  gap: var(--space-3);
}

.extra-info {
  display: flex;
  justify-content: center;
  font-size: var(--font-size-xs);
  color: var(--color-neutral-500);
  padding-top: var(--space-2);
  border-top: 1px solid var(--color-neutral-100);
}

/* 响应式适配 */
@media (max-width: 360px) {
  .app-footer {
    padding: var(--space-3);
  }
  
  .footer-content {
    gap: var(--space-2);
  }
}
</style>