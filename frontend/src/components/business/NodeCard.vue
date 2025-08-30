<template>
  <div 
    :class="[
      'node-card',
      { 'node-card--clickable': clickable }
    ]"
    @click="handleClick"
  >
    <div class="node-header">
      <div class="node-type">
        <i :class="typeIcon" />
        <span class="node-type-text">{{ typeLabel }}</span>
      </div>
      <div v-if="ping !== undefined" class="node-ping">
        <span :class="pingClass">{{ ping }}ms</span>
      </div>
    </div>
    
    <div class="node-content">
      <h4 class="node-name" :title="node?.name">
        {{ displayName }}
      </h4>
      <p v-if="node?.addr" class="node-server">
        {{ node.addr }}
      </p>
    </div>
    
    <div v-if="showActions" class="node-actions">
      <n-button 
        size="small" 
        type="primary"
        ghost
        @click.stop="$emit('change')"
      >
        更换
      </n-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Peer } from '@/types/models'

interface Props {
  node?: Peer | null
  type: 'game' | 'http'
  ping?: number
  clickable?: boolean
  showActions?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  clickable: true,
  showActions: true
})

const emit = defineEmits<{
  click: []
  change: []
}>()

// 计算显示属性
const typeLabel = computed(() => {
  return props.type === 'game' ? '游戏节点' : '网页节点'
})

const typeIcon = computed(() => {
  return props.type === 'game' 
    ? 'i-ionicons-game-controller-outline'
    : 'i-ionicons-globe-outline'
})

const displayName = computed(() => {
  if (!props.node?.name) return '未选择节点'
  // 截断过长的节点名
  return props.node.name.length > 15 
    ? props.node.name.substring(0, 15) + '...'
    : props.node.name
})

const pingClass = computed(() => {
  if (props.ping === undefined) return ''
  if (props.ping < 50) return 'ping-excellent'
  if (props.ping < 100) return 'ping-good'
  if (props.ping < 200) return 'ping-fair'
  return 'ping-poor'
})

const handleClick = () => {
  if (props.clickable) {
    emit('click')
  }
}
</script>

<style scoped>
.node-card {
  background: var(--bg-elevated);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-md);
  padding: var(--space-4);
  transition: var(--transition-base);
}

.node-card--clickable {
  cursor: pointer;
}

.node-card--clickable:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-sm);
  transform: translateY(-1px);
}

.node-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-3);
}

.node-type {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--font-size-xs);
  color: var(--color-neutral-500);
  font-weight: var(--font-weight-medium);
}

.node-type i {
  font-size: 14px;
}

.node-ping {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
}

.ping-excellent { color: var(--color-success); }
.ping-good { color: var(--color-success-light); }
.ping-fair { color: var(--color-warning); }
.ping-poor { color: var(--color-danger); }

.node-content {
  margin-bottom: var(--space-3);
}

.node-name {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-neutral-900);
  margin: 0 0 var(--space-1);
  line-height: var(--line-height-tight);
}

.node-server {
  font-size: var(--font-size-xs);
  color: var(--color-neutral-500);
  margin: 0;
  line-height: var(--line-height-normal);
}

.node-actions {
  display: flex;
  justify-content: flex-end;
}
</style>