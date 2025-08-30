<template>
  <n-modal
    v-model:show="showModal"
    preset="card"
    :title="type === 'game' ? '游戏节点设置' : 'HTTP节点设置'"
    style="width: 600px"
    :bordered="false"
  >
    <n-tabs v-model:value="activeTab" type="line">
      <!-- 节点选择标签页 -->
      <n-tab-pane name="select" tab="选择节点">
        <div class="node-selection">
          <!-- 游戏节点选择 -->
          <div class="node-group">
            <label class="node-label">游戏节点</label>
            <n-select
              v-model:value="selectedGameNode"
              :options="gameNodeOptions"
              placeholder="请选择游戏节点"
              :loading="isLoading"
              clearable
            />
          </div>
          
          <!-- HTTP节点选择 -->
          <div class="node-group">
            <label class="node-label">HTTP节点</label>
            <n-select
              v-model:value="selectedHttpNode"
              :options="httpNodeOptions"
              placeholder="请选择HTTP节点"
              :loading="isLoading"
              clearable
            />
          </div>
          
          <!-- Ping测试按钮 -->
          <div v-if="showPingInfo" class="ping-info">
            <n-button
              type="primary"
              ghost
              @click="handlePingTest"
              :loading="isPinging"
            >
              {{ isPinging ? '测试中...' : '测试延迟' }}
            </n-button>
          </div>
        </div>
      </n-tab-pane>
      
      <!-- 导入订阅标签页 -->
      <n-tab-pane name="import" tab="导入订阅">
        <div class="import-subscription">
          <div class="input-group">
            <label class="node-label">订阅地址或配置</label>
            <n-input
              v-model:value="subscriptionUrl"
              type="textarea"
              :rows="4"
              placeholder="请输入订阅URL、GPP token或sing-box配置JSON"
              :disabled="isImporting"
            />
          </div>
          
          <n-button
            type="primary"
            block
            @click="handleImportSubscription"
            :loading="isImporting"
            :disabled="!subscriptionUrl.trim()"
          >
            {{ isImporting ? '导入中...' : '导入' }}
          </n-button>
          
          <n-alert type="info" :show-icon="false">
            <div style="font-size: 12px;">
              <p>支持的格式：</p>
              <ul style="margin: 4px 0 0 20px; padding: 0;">
                <li>HTTP订阅链接（http://...）</li>
                <li>GPP分享token</li>
                <li>sing-box配置JSON（包含outbounds）</li>
              </ul>
            </div>
          </n-alert>
        </div>
      </n-tab-pane>
    </n-tabs>
    
    <!-- 底部按钮 -->
    <template #footer>
      <div style="display: flex; justify-content: flex-end; gap: 8px;">
        <n-button @click="handleCancel">
          取消
        </n-button>
        <n-button
          v-if="activeTab === 'select'"
          type="primary"
          @click="handleConfirm"
          :disabled="!canConfirm"
        >
          确定
        </n-button>
      </div>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { Peer } from '@/types/models'
import { useNodeManager } from '@/composables/useNodeManager'
import { useMessage } from 'naive-ui'

interface Props {
  show: boolean
  type?: 'game' | 'http'
  currentGameNode?: Peer | null
  currentHttpNode?: Peer | null
}

const props = withDefaults(defineProps<Props>(), {
  type: 'game'
})

const emit = defineEmits<{
  'update:show': [value: boolean]
  confirm: [gameNode: Peer | null, httpNode: Peer | null]
}>()

const message = useMessage()

// 使用真实的节点管理器
const {
  allNodes,
  gameNodes,
  httpNodes,
  isLoading,
  isPinging,
  refreshNodes,
  pingAll,
  importSubscription,
  importSingBoxConfig
} = useNodeManager()

// 本地状态
const activeTab = ref('select')
const selectedGameNode = ref<string | null>(null)
const selectedHttpNode = ref<string | null>(null)
const subscriptionUrl = ref('')
const isImporting = ref(false)

// 计算属性
const showModal = computed({
  get: () => props.show,
  set: (value) => emit('update:show', value)
})

const gameNodeOptions = computed(() => {
  return gameNodes.value.map(node => {
    let label = node.name
    if (node.addr) {
      label += ` (${node.addr})`
    }
    if (node.ping && node.ping > 0) {
      label += ` - ${node.ping}ms`
    }
    return {
      label,
      value: node.name,
      node: node
    }
  })
})

const httpNodeOptions = computed(() => {
  return httpNodes.value.map(node => {
    let label = node.name
    if (node.addr) {
      label += ` (${node.addr})`
    }
    if (node.ping && node.ping > 0) {
      label += ` - ${node.ping}ms`
    }
    return {
      label,
      value: node.name, 
      node: node
    }
  })
})

const showPingInfo = computed(() => {
  return selectedGameNode.value || selectedHttpNode.value
})

const canConfirm = computed(() => {
  return selectedGameNode.value && selectedHttpNode.value
})

// 监听器
watch(() => props.show, (show) => {
  if (show) {
    // 重置选择状态
    selectedGameNode.value = props.currentGameNode?.name || null
    selectedHttpNode.value = props.currentHttpNode?.name || null
    subscriptionUrl.value = ''
    activeTab.value = 'select'
    
    // 刷新节点列表
    refreshNodes()
  }
})

// 事件处理
const handlePingTest = async () => {
  await pingAll()
}

const handleImportSubscription = async () => {
  if (!subscriptionUrl.value.trim()) {
    message.warning('请输入订阅地址或配置内容')
    return
  }
  
  try {
    isImporting.value = true
    const content = subscriptionUrl.value.trim()
    
    // 判断是否为 sing-box 配置
    let success = false
    if (content.startsWith('{') && (content.includes('"type"') || content.includes('"outbounds"'))) {
      // 可能是 sing-box 配置
      try {
        success = await importSingBoxConfig(content)
        if (success) {
          message.success('sing-box 配置导入成功')
        }
      } catch (error: any) {
        // 如果 sing-box 导入失败，显示具体错误
        if (error.message && (error.message.includes('入站') || error.message.includes('inbound'))) {
          message.error(error.message)
          return
        }
        // 其他错误，尝试作为订阅URL
        success = await importSubscription(content)
        if (success) {
          message.success('订阅导入成功')
        }
      }
    } else {
      // 作为订阅URL或GPP token处理
      success = await importSubscription(content)
      if (success) {
        message.success('导入成功')
      }
    }
    
    if (success) {
      // 导入成功后切换到选择标签页
      activeTab.value = 'select'
      subscriptionUrl.value = ''
      await refreshNodes()
    }
  } catch (error: any) {
    console.error('Import error:', error)
    message.error(`导入失败: ${error.message || error}`)
  } finally {
    isImporting.value = false
  }
}

const handleCancel = () => {
  showModal.value = false
}

const handleConfirm = () => {
  const gameNode = gameNodes.value.find(n => n.name === selectedGameNode.value) || null
  const httpNode = httpNodes.value.find(n => n.name === selectedHttpNode.value) || null
  
  emit('confirm', gameNode, httpNode)
}
</script>

<style scoped>
.node-selection {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.node-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.node-label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-neutral-700);
}

.ping-info {
  display: flex;
  justify-content: center;
  padding-top: var(--space-4);
  border-top: 1px solid var(--color-neutral-200);
}

.import-subscription {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.input-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}
</style>