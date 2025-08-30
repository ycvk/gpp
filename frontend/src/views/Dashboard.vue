<template>
  <div class="dashboard">
    <!-- 连接状态显示 -->
    <ConnectionStatus
      :status="connectionStatus"
      :game-node="status.gamePeer"
      :http-node="status.httpPeer" 
      :game-ping="gamePing"
      :http-ping="httpPing"
      :traffic-data="trafficData"
      :show-traffic="!!isConnected"
      @node-select="handleNodeSelect"
    />
    
    <!-- 主操作按钮 -->
    <div class="action-section">
      <n-button
        :type="isConnected ? 'error' : 'primary'"
        :loading="isLoading"
        :disabled="!canPerformAction"
        size="large"
        block
        @click="handlePrimaryAction"
      >
        {{ primaryButtonText }}
      </n-button>
      
      <!-- 辅助操作 -->
      <div class="secondary-actions">
        <n-button
          type="tertiary"
          size="small"
          @click="handleNodeSettings"
        >
          <template #icon>
            <i class="i-ionicons-settings-outline" />
          </template>
          节点设置
        </n-button>
        
        <n-button
          type="tertiary"
          size="small"
          @click="handleRefreshNodes"
          :loading="isRefreshing"
        >
          <template #icon>
            <i class="i-ionicons-refresh-outline" />
          </template>
          刷新
        </n-button>
      </div>
      
      <!-- 版本信息 -->
      <div class="version-info">
        <n-text depth="3" style="font-size: var(--font-size-xs)">
          v1.4.6
        </n-text>
      </div>
    </div>
    
    <!-- 节点选择模态框 -->
    <NodeSelectorModal
      v-model:show="showNodeSelector"
      :type="selectedNodeType"
      :current-game-node="status.gamePeer"
      :current-http-node="status.httpPeer"
      @confirm="handleNodeConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useProxyState } from '@/composables/useProxyState'
import { useNodeManager } from '@/composables/useNodeManager'
import ConnectionStatus from '@/components/business/ConnectionStatus.vue'
import NodeSelectorModal from '@/components/business/NodeSelectorModal.vue'
import type { Peer } from '@/types/models'

// 状态管理
const {
  status,
  isLoading,
  connectionStatus,
  isConnected,
  startProxy,
  stopProxy,
  setPeerNodes
} = useProxyState()

// 临时节点管理模拟 - 实际应该使用 useNodeManager
const isRefreshing = ref(false)
const refreshNodes = async () => {
  isRefreshing.value = true
  // Simulate refresh
  await new Promise(resolve => setTimeout(resolve, 1000))
  isRefreshing.value = false
}

// UI状态
const showNodeSelector = ref(false)
const selectedNodeType = ref<'game' | 'http'>('game')
const gamePing = ref<number>()
const httpPing = ref<number>()

// 计算属性
const primaryButtonText = computed(() => {
  if (isLoading.value) {
    return isConnected.value ? '正在停止...' : '正在连接...'
  }
  return isConnected.value ? '结束加速' : '开始加速'
})

const canPerformAction = computed(() => {
  return !isLoading.value && (
    isConnected.value || (status.gamePeer && status.httpPeer)
  )
})

const trafficData = computed(() => {
  if (!isConnected.value) return null
  
  return {
    uploadSpeed: status.upSpeed,
    downloadSpeed: status.downSpeed,
    totalUpload: status.totalUp,
    totalDownload: status.totalDown
  }
})

// 事件处理
const handlePrimaryAction = async () => {
  if (isConnected.value) {
    await stopProxy()
  } else {
    await startProxy()
  }
}

const handleNodeSelect = (type: 'game' | 'http') => {
  selectedNodeType.value = type
  showNodeSelector.value = true
}

const handleNodeSettings = () => {
  selectedNodeType.value = 'game'
  showNodeSelector.value = true
}

const handleRefreshNodes = async () => {
  await refreshNodes()
  // 这里应该也调用 pingNodes 等方法
}

const handleNodeConfirm = async (gameNode: Peer | null, httpNode: Peer | null) => {
  // 使用正确的API签名：SetPeer需要两个节点名称
  if (gameNode && httpNode) {
    await setPeerNodes(gameNode, httpNode)
  }
  showNodeSelector.value = false
}

// 定期ping检测 - 这里应该集成到 useNodeManager 中
const updatePingValues = async () => {
  try {
    if (status.gamePeer) {
      // 实际应该调用PingAll或单独ping方法
      gamePing.value = Math.floor(Math.random() * 200) + 20
    }
    if (status.httpPeer) {
      httpPing.value = Math.floor(Math.random() * 200) + 20  
    }
  } catch (error) {
    console.error('Ping检测失败:', error)
  }
}

// 定期更新ping值
setInterval(updatePingValues, 30000)
</script>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: var(--space-8);
  height: 100%;
}

.action-section {
  margin-top: auto;
  padding-top: var(--space-6);
}

.secondary-actions {
  display: flex;
  justify-content: space-between;
  margin-top: var(--space-4);
}

.version-info {
  display: flex;
  justify-content: center;
  margin-top: var(--space-4);
}
</style>