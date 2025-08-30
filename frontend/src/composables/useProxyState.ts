import { ref, reactive, computed, onMounted, onBeforeUnmount, readonly } from 'vue'
// Note: These imports should be available once the Wails app is built
// import { Status, Start, Stop, SetPeer, PingAll } from '@/wailsjs/go/main/App'
import type { Peer, ProxyStatus } from '@/types/models'
import { useMessage } from 'naive-ui'

// Temporary mock functions for development - replace with actual Wails imports
const Status = async () => ({ running: false, game_peer: null, http_peer: null, up: 0, down: 0 })
const Start = async () => 'ok'
const Stop = async () => 'ok'  
const SetPeer = async (game: string, http: string) => 'ok'

export const useProxyState = () => {
  const message = useMessage()
  
  // 状态管理
  const status = reactive<ProxyStatus>({
    isRunning: false,
    gamePeer: null,
    httpPeer: null,
    upSpeed: 0,
    downSpeed: 0,
    totalUp: 0,
    totalDown: 0
  })
  
  // UI状态
  const isLoading = ref(false)
  const connectionStatus = ref<'idle' | 'connecting' | 'connected' | 'error'>('idle')
  
  // 智能轮询相关
  const pollInterval = ref(5000)
  const lastStateSnapshot = ref('')
  let pollTimer: number | null = null
  
  // 计算属性
  const isConnected = computed(() => status.isRunning && status.gamePeer && status.httpPeer)
  
  // 状态快照，用于检测变化
  const createStateSnapshot = () => {
    return JSON.stringify({
      isRunning: status.isRunning,
      gameNode: status.gamePeer?.name || null,
      httpNode: status.httpPeer?.name || null
    })
  }
  
  // 获取状态的核心方法
  const updateStatus = async () => {
    try {
      const result = await Status()
      
      // 更新状态 - 根据实际API返回字段调整
      status.isRunning = result.running || false  // 注意：字段名是running不是is_running
      status.gamePeer = result.game_peer || null
      status.httpPeer = result.http_peer || null
      status.upSpeed = result.up || 0
      status.downSpeed = result.down || 0
      // 注意：当前API没有total_up/total_down字段，需要前端累计
      status.totalUp = (status.totalUp || 0) + (result.up || 0)
      status.totalDown = (status.totalDown || 0) + (result.down || 0)
      
      // 更新连接状态
      if (status.isRunning && status.gamePeer && status.httpPeer) {
        connectionStatus.value = 'connected'
      } else if (status.gamePeer || status.httpPeer) {
        connectionStatus.value = 'idle'
      }
      
      // 智能轮询频率调整
      const currentSnapshot = createStateSnapshot()
      if (lastStateSnapshot.value !== currentSnapshot) {
        pollInterval.value = 1000 // 状态变化时加快轮询
        lastStateSnapshot.value = currentSnapshot
      } else if (pollInterval.value < 5000) {
        pollInterval.value = Math.min(pollInterval.value + 500, 5000)
      }
      
    } catch (error) {
      console.error('获取状态失败:', error)
      connectionStatus.value = 'error'
    }
  }
  
  // 启动智能轮询
  const startPolling = () => {
    if (pollTimer) clearTimeout(pollTimer)
    
    const poll = async () => {
      await updateStatus()
      pollTimer = window.setTimeout(poll, pollInterval.value)
    }
    
    poll()
  }
  
  // 停止轮询
  const stopPolling = () => {
    if (pollTimer) {
      clearTimeout(pollTimer)
      pollTimer = null
    }
  }
  
  // 启动代理
  const startProxy = async () => {
    if (!status.gamePeer || !status.httpPeer) {
      message.error('请先选择游戏节点和网页节点')
      return false
    }
    
    try {
      isLoading.value = true
      connectionStatus.value = 'connecting'
      
      const result = await Start()
      
      if (result === 'ok' || result === 'running') {
        message.success('代理启动成功')
        connectionStatus.value = 'connected'
        return true
      } else {
        throw new Error(result)
      }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error)
      message.error(`启动失败: ${errorMsg}`)
      connectionStatus.value = 'error'
      return false
    } finally {
      isLoading.value = false
    }
  }
  
  // 停止代理
  const stopProxy = async () => {
    try {
      isLoading.value = true
      connectionStatus.value = 'connecting'
      
      const result = await Stop()
      
      if (result === 'ok') {
        message.success('代理已停止')
        connectionStatus.value = 'idle'
        return true
      } else {
        throw new Error(result)
      }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error)
      message.error(`停止失败: ${errorMsg}`)
      return false
    } finally {
      isLoading.value = false
    }
  }
  
  // 设置节点 - 根据实际API签名 SetPeer(game: string, http: string)
  const setPeerNodes = async (gameNode: Peer | null, httpNode: Peer | null) => {
    if (!gameNode || !httpNode) {
      message.error('请选择完整的游戏和网页节点')
      return false
    }
    
    try {
      const result = await SetPeer(gameNode.name, httpNode.name)
      
      if (result === 'ok') {
        status.gamePeer = gameNode
        status.httpPeer = httpNode
        message.success('节点设置成功')
        return true
      } else {
        throw new Error(result)
      }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error)
      message.error(`设置节点失败: ${errorMsg}`)
      return false
    }
  }
  
  // 组件生命周期
  onMounted(() => {
    startPolling()
  })
  
  onBeforeUnmount(() => {
    stopPolling()
  })
  
  return {
    // 状态
    status: readonly(status),
    isLoading: readonly(isLoading),
    connectionStatus: readonly(connectionStatus),
    
    // 计算属性
    isConnected,
    
    // 方法
    updateStatus,
    startProxy,
    stopProxy,
    setPeerNodes,
    
    // 轮询控制
    startPolling,
    stopPolling
  }
}