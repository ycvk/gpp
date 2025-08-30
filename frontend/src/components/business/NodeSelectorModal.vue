<template>
  <n-modal 
    v-model:show="showModal"
    :mask-closable="false"
    preset="card"
    title="节点设置"
    style="width: 320px;"
    :segmented="true"
  >
    <n-tabs v-model:value="activeTab" type="segment">
      <n-tab-pane name="select" tab="选择节点">
        <div class="node-selection">
          <!-- 游戏节点选择 -->
          <div class="node-group">
            <label class="node-label">游戏节点</label>
            <n-select
              v-model:value="selectedGameNode"
              :options="gameNodeOptions"
              placeholder="选择游戏节点"
              :loading="isLoadingNodes"
              clearable
              filterable
            />
          </div>
          
          <!-- 网页节点选择 -->
          <div class="node-group">
            <label class="node-label">网页节点</label>
            <n-select
              v-model:value="selectedHttpNode"
              :options="httpNodeOptions"
              placeholder="选择网页节点"
              :loading="isLoadingNodes"
              clearable
              filterable
            />
          </div>
          
          <!-- 节点ping信息 -->
          <div v-if="showPingInfo" class="ping-info">
            <n-button
              type="tertiary"
              size="small"
              :loading="isPinging"
              @click="handlePingTest"
            >
              <template #icon>
                <i class="i-ionicons-pulse-outline" />
              </template>
              测试延迟
            </n-button>
          </div>
        </div>
      </n-tab-pane>
      
      <n-tab-pane name="import" tab="导入订阅">
        <div class="import-subscription">
          <div class="input-group">
            <label class="input-label">订阅链接</label>
            <n-input
              v-model:value="subscriptionUrl"
              placeholder="输入订阅URL"
              type="text"
              clearable
            />
          </div>
          
          <n-button
            type="primary"
            :loading="isImporting"
            :disabled="!subscriptionUrl"
            block
            @click="handleImportSubscription"
          >
            导入订阅
          </n-button>
        </div>
      </n-tab-pane>
    </n-tabs>
    
    <template #action>
      <div class="modal-actions">
        <n-button @click="handleCancel">取消</n-button>
        <n-button 
          type="primary"
          :disabled="!canConfirm"
          @click="handleConfirm"
        >
          确认
        </n-button>
      </div>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { Peer } from '@/types/models'

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

// Mock node management - 实际应该使用 useNodeManager
const gameNodes = ref<Peer[]>([])
const httpNodes = ref<Peer[]>([])
const isLoadingNodes = ref(false)
const isPinging = ref(false)

const refreshNodes = async () => {
  // Mock implementation
  isLoadingNodes.value = true
  await new Promise(resolve => setTimeout(resolve, 1000))
  gameNodes.value = [
    { name: '游戏节点1', server: 'hk.example.com' },
    { name: '游戏节点2', server: 'sg.example.com' }
  ]
  httpNodes.value = [
    { name: 'HTTP节点1', server: 'us.example.com' },
    { name: 'HTTP节点2', server: 'jp.example.com' }
  ]
  isLoadingNodes.value = false
}

const pingAll = async () => {
  isPinging.value = true
  await new Promise(resolve => setTimeout(resolve, 2000))
  isPinging.value = false
}

const importSubscription = async (url: string) => {
  // Mock implementation
  await new Promise(resolve => setTimeout(resolve, 1500))
  return true
}

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
  return gameNodes.value.map(node => ({
    label: `${node.name} (${node.server || '未知'})`,
    value: node.name,
    node: node
  }))
})

const httpNodeOptions = computed(() => {
  return httpNodes.value.map(node => ({
    label: `${node.name} (${node.server || '未知'})`,
    value: node.name, 
    node: node
  }))
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
  if (!subscriptionUrl.value) return
  
  try {
    isImporting.value = true
    await importSubscription(subscriptionUrl.value)
    
    // 导入成功后切换到选择标签页
    activeTab.value = 'select'
    await refreshNodes()
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

.input-label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-neutral-700);
}

.modal-actions {
  display: flex;
  gap: var(--space-3);
}
</style>