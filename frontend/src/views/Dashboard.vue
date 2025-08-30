<template>
  <div class="dashboard">
    <!-- 错误状态显示 -->
    <ErrorState
      v-if="showErrorState"
      :type="errorType"
      :message="errorMessage"
      @retry="handleErrorRetry"
      @help="handleErrorHelp"
    />
    
    <!-- 连接状态显示 -->
    <ConnectionStatus
      v-else
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
            <n-icon :size="18">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                <path d="M12 15.5A3.5 3.5 0 0 1 8.5 12A3.5 3.5 0 0 1 12 8.5a3.5 3.5 0 0 1 3.5 3.5a3.5 3.5 0 0 1-3.5 3.5m7.43-2.53c.04-.32.07-.64.07-.97c0-.33-.03-.66-.07-1l2.11-1.63c.19-.15.24-.42.12-.64l-2-3.46c-.12-.22-.39-.3-.61-.22l-2.49 1c-.52-.39-1.06-.73-1.69-.98l-.37-2.65A.506.506 0 0 0 14 2h-4c-.25 0-.46.18-.5.42l-.37 2.65c-.63.25-1.17.59-1.69.98l-2.49-1c-.22-.09-.49 0-.61.22l-2 3.46c-.13.22-.07.49.12.64L4.57 11c-.04.34-.07.67-.07 1c0 .33.03.65.07.97l-2.11 1.66c-.19.15-.25.42-.12.64l2 3.46c.12.22.39.3.61.22l2.49-1.01c.52.4 1.06.74 1.69.99l.37 2.65c.04.24.25.42.5.42h4c.25 0 .46-.18.5-.42l.37-2.65c.63-.26 1.17-.59 1.69-.99l2.49 1.01c.22.08.49 0 .61-.22l2-3.46c.12-.22.07-.49-.12-.64z"/>
              </svg>
            </n-icon>
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
            <n-icon :size="18">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                <path d="M17.65 6.35C16.2 4.9 14.21 4 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08c-.82 2.33-3.04 4-5.65 4c-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"/>
              </svg>
            </n-icon>
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
import { computed, ref, onMounted, watchEffect } from 'vue'
import { useProxyState } from '@/composables/useProxyState'
import { useNodeManager } from '@/composables/useNodeManager'
import { NIcon } from 'naive-ui'
import ConnectionStatus from '@/components/business/ConnectionStatus.vue'
import NodeSelectorModal from '@/components/business/NodeSelectorModal.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
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

// 节点管理
const {
  refreshNodes,
  pingAll,
  isLoading: isRefreshing
} = useNodeManager()

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

// 错误状态计算
const showErrorState = computed(() => {
  return connectionStatus.value === 'error' && !isConnected.value
})

const errorType = computed(() => {
  if (connectionStatus.value === 'error') {
    if (!status.gamePeer || !status.httpPeer) {
      return 'config' as const
    }
    return 'connection' as const
  }
  return 'unknown' as const
})

const errorMessage = computed(() => {
  if (!status.gamePeer || !status.httpPeer) {
    return '请先选择游戏节点和网页节点'
  }
  return '无法连接到代理服务器，请检查网络设置'
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
  await pingAll()
}

const handleNodeConfirm = async (gameNode: Peer | null, httpNode: Peer | null) => {
  // 使用正确的API签名：SetPeer需要两个节点名称
  if (gameNode && httpNode) {
    await setPeerNodes(gameNode, httpNode)
  }
  showNodeSelector.value = false
}

// 错误处理方法
const handleErrorRetry = async () => {
  if (!status.gamePeer || !status.httpPeer) {
    // 如果没有节点，打开节点选择器
    showNodeSelector.value = true
  } else {
    // 否则尝试重新连接
    await startProxy()
  }
}

const handleErrorHelp = () => {
  // 可以打开帮助文档或显示故障排除提示
  console.log('显示帮助信息')
  // 这里可以集成帮助文档或FAQ
}

// 定期更新ping值（使用composable管理）
onMounted(() => {
  // 初始化时更新节点ping值
  if (status.gamePeer || status.httpPeer) {
    pingAll()
  }
})

// 使用 watchEffect 监听节点变化
watchEffect(() => {
  if (status.gamePeer || status.httpPeer) {
    // 节点变化时自动更新ping
    pingAll()
  }
})
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