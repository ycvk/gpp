import { ref, reactive, computed, readonly } from 'vue'
// import { List, Add, PingAll } from '@/wailsjs/go/main/App'
import type { Peer } from '@/types/models'
import { useMessage } from 'naive-ui'

// Temporary mock functions for development
const List = async () => [] as Peer[]
const Add = async (url: string) => 'ok'
const PingAll = async () => {}

export const useNodeManager = () => {
  const message = useMessage()
  
  // 状态管理
  const allNodes = ref<Peer[]>([])
  const isLoading = ref(false)
  const isPinging = ref(false)
  const lastRefreshTime = ref<Date | null>(null)
  
  // 计算分类节点
  const gameNodes = computed(() => {
    return allNodes.value.filter(node => 
      node.name.startsWith('game') || 
      node.name.includes('游戏') ||
      node.type === 'game'
    )
  })
  
  const httpNodes = computed(() => {
    return allNodes.value.filter(node => 
      node.name.startsWith('http') || 
      node.name.includes('网页') ||
      node.type === 'http' ||
      (!node.name.startsWith('game') && !node.name.includes('游戏'))
    )
  })
  
  // 刷新节点列表 - 优化：避免重复调用
  const refreshNodes = async (force = false) => {
    // 避免频繁刷新（5分钟内不重复刷新）
    if (!force && lastRefreshTime.value) {
      const timeDiff = Date.now() - lastRefreshTime.value.getTime()
      if (timeDiff < 5 * 60 * 1000) {
        return
      }
    }
    
    try {
      isLoading.value = true
      
      const nodes = await List()
      allNodes.value = nodes || []
      lastRefreshTime.value = new Date()
      
      message.success(`已更新 ${nodes.length} 个节点`)
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : '获取节点列表失败'
      message.error(errorMsg)
      console.error('刷新节点失败:', error)
    } finally {
      isLoading.value = false
    }
  }
  
  // Ping所有节点
  const pingAll = async () => {
    if (allNodes.value.length === 0) {
      message.warning('暂无节点可测试')
      return
    }
    
    try {
      isPinging.value = true
      await PingAll()
      message.success('延迟测试完成')
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Ping测试失败'
      message.error(errorMsg)
    } finally {
      isPinging.value = false
    }
  }
  
  // 导入订阅
  const importSubscription = async (url: string) => {
    if (!url.trim()) {
      message.error('请输入有效的订阅URL')
      return false
    }
    
    try {
      isLoading.value = true
      
      // 调用后端导入订阅API
      const result = await Add(url.trim())
      
      if (result === 'ok') {
        message.success('订阅导入成功')
        // 导入后自动刷新节点列表
        await refreshNodes(true)
        return true
      } else {
        throw new Error(result)
      }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : '导入订阅失败'
      message.error(errorMsg)
      return false
    } finally {
      isLoading.value = false
    }
  }
  
  // 搜索节点
  const searchNodes = (keyword: string) => {
    if (!keyword.trim()) {
      return allNodes.value
    }
    
    const searchTerm = keyword.toLowerCase()
    return allNodes.value.filter(node =>
      node.name.toLowerCase().includes(searchTerm) ||
      (node.server && node.server.toLowerCase().includes(searchTerm))
    )
  }
  
  return {
    // 状态
    allNodes: readonly(allNodes),
    gameNodes,
    httpNodes,
    isLoading: readonly(isLoading),
    isPinging: readonly(isPinging),
    lastRefreshTime: readonly(lastRefreshTime),
    
    // 方法
    refreshNodes,
    pingAll,
    importSubscription,
    searchNodes
  }
}