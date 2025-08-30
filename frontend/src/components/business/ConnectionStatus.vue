<template>
  <div class="connection-status">
    <!-- 状态指示器 -->
    <div class="status-indicator">
      <div 
        :class="[
          'status-dot',
          `status-dot--${status}`
        ]"
      >
        <div v-if="status === 'connecting'" class="pulse-ring" />
      </div>
      <div class="status-info">
        <h3 class="status-title">{{ statusTitle }}</h3>
        <p class="status-description">{{ statusDescription }}</p>
      </div>
    </div>
    
    <!-- 节点信息卡片组 -->
    <div v-if="showNodes" class="nodes-grid">
      <NodeCard 
        v-if="gameNode"
        :node="gameNode"
        type="game"
        :ping="gamePing"
        @change="handleNodeSelect('game')"
      />
      <NodeCard 
        v-if="httpNode" 
        :node="httpNode"
        type="http"
        :ping="httpPing"
        @change="handleNodeSelect('http')"
      />
    </div>
    
    <!-- 流量统计 -->
    <TrafficStats 
      v-if="showTraffic && trafficData"
      :upload-speed="trafficData.uploadSpeed"
      :download-speed="trafficData.downloadSpeed"
      :total-upload="trafficData.totalUpload"
      :total-download="trafficData.totalDownload"
    />
  </div>
</template>

<script setup lang="ts">
import type { Peer } from '@/types/models'
import NodeCard from './NodeCard.vue'
import TrafficStats from './TrafficStats.vue'

interface TrafficData {
  uploadSpeed: number
  downloadSpeed: number  
  totalUpload: number
  totalDownload: number
}

interface Props {
  status: 'idle' | 'connecting' | 'connected' | 'error'
  gameNode?: Peer | null
  httpNode?: Peer | null
  gamePing?: number
  httpPing?: number
  trafficData?: TrafficData | null
  showNodes?: boolean
  showTraffic?: boolean | null
}

const props = withDefaults(defineProps<Props>(), {
  showNodes: true,
  showTraffic: false
})

const emit = defineEmits<{
  nodeSelect: [type: 'game' | 'http']
}>()

// 计算状态显示文本
const statusTitle = computed(() => {
  switch (props.status) {
    case 'idle':
      return '未连接'
    case 'connecting':
      return '正在连接...'
    case 'connected':
      return '已连接'
    case 'error':
      return '连接失败'
    default:
      return '未知状态'
  }
})

const statusDescription = computed(() => {
  switch (props.status) {
    case 'idle':
      return '选择节点开始加速'
    case 'connecting':
      return '正在建立安全连接'
    case 'connected':
      return '加速服务运行中'
    case 'error':
      return '请检查网络设置'
    default:
      return ''
  }
})

const handleNodeSelect = (type: 'game' | 'http') => {
  emit('nodeSelect', type)
}
</script>

<style scoped>
.connection-status {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-4);
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-neutral-200);
}

.status-dot {
  position: relative;
  width: 12px;
  height: 12px;
  border-radius: var(--radius-full);
  flex-shrink: 0;
}

.status-dot--idle {
  background: var(--color-neutral-400);
}

.status-dot--connecting {
  background: var(--color-warning);
}

.status-dot--connected {
  background: var(--color-success);
}

.status-dot--error {
  background: var(--color-danger);
}

.pulse-ring {
  position: absolute;
  top: -4px;
  left: -4px;
  right: -4px;
  bottom: -4px;
  border: 2px solid var(--color-warning);
  border-radius: var(--radius-full);
  animation: pulse 1.5s infinite;
  opacity: 0.6;
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

.status-info {
  flex: 1;
  min-width: 0;
}

.status-title {
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-semibold);
  color: var(--color-neutral-900);
  margin: 0;
  line-height: var(--line-height-tight);
}

.status-description {
  font-size: var(--font-size-sm);
  color: var(--color-neutral-500);
  margin: 2px 0 0;
  line-height: var(--line-height-normal);
}

.nodes-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-4);
}

@media (max-width: 320px) {
  .nodes-grid {
    grid-template-columns: 1fr;
  }
}
</style>