<template>
  <div class="settings-container">
    <div class="settings-header">
      <n-button 
        text 
        @click="$router.push('/')"
        class="back-button"
      >
        <template #icon>
          <n-icon :size="20">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
              <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"/>
            </svg>
          </n-icon>
        </template>
        返回
      </n-button>
      <h1>系统设置</h1>
      <div style="width: 60px"></div> <!-- 占位符，保持标题居中 -->
    </div>

    <n-card class="settings-card">
      <n-tabs type="line" animated>
        <!-- 基本设置 -->
        <n-tab-pane name="general" tab="基本设置">
          <div class="settings-section">
            <h3>应用设置</h3>
            
            <div class="setting-item">
              <label>主题</label>
              <n-select 
                v-model:value="settings.theme" 
                :options="themeOptions"
                style="width: 200px"
              />
            </div>
            
            <div class="setting-item">
              <label>语言</label>
              <n-select 
                v-model:value="settings.language" 
                :options="languageOptions"
                style="width: 200px"
              />
            </div>
            
            <div class="setting-item">
              <label>开机自启动</label>
              <n-switch v-model:value="settings.autoStart" />
            </div>
            
            <div class="setting-item">
              <label>最小化到托盘</label>
              <n-switch v-model:value="settings.minimizeToTray" />
            </div>
          </div>
        </n-tab-pane>
        
        <!-- 网络设置 -->
        <n-tab-pane name="network" tab="网络设置">
          <div class="settings-section">
            <h3>DNS配置</h3>
            
            <div class="setting-item">
              <label>代理DNS</label>
              <n-input 
                v-model:value="settings.proxyDns" 
                placeholder="例如: 1.1.1.1"
                style="width: 200px"
              />
            </div>
            
            <div class="setting-item">
              <label>本地DNS</label>
              <n-input 
                v-model:value="settings.localDns" 
                placeholder="例如: 223.5.5.5"
                style="width: 200px"
              />
            </div>
            
            <h3>代理设置</h3>
            
            <div class="setting-item">
              <label>自动选择最佳节点</label>
              <n-switch v-model:value="settings.autoSelectNode" />
            </div>
            
            <div class="setting-item">
              <label>连接超时(秒)</label>
              <n-input-number 
                v-model:value="settings.connectionTimeout" 
                :min="5"
                :max="60"
                style="width: 200px"
              />
            </div>
          </div>
        </n-tab-pane>
        
        <!-- 节点管理 -->
        <n-tab-pane name="nodes" tab="节点管理">
          <div class="settings-section">
            <h3>订阅管理</h3>
            
            <div class="subscription-list">
              <div v-for="(sub, index) in subscriptions" :key="index" class="subscription-item">
                <n-input 
                  v-model:value="sub.url" 
                  placeholder="订阅链接"
                  :disabled="!sub.editing"
                />
                <n-button 
                  v-if="!sub.editing"
                  @click="editSubscription(index)"
                  type="primary"
                  ghost
                >
                  编辑
                </n-button>
                <n-button 
                  v-else
                  @click="saveSubscription(index)"
                  type="success"
                >
                  保存
                </n-button>
                <n-button 
                  @click="removeSubscription(index)"
                  type="error"
                  ghost
                >
                  删除
                </n-button>
              </div>
            </div>
            
            <n-button 
              @click="addSubscription"
              type="primary"
              dashed
              block
              style="margin-top: var(--space-4)"
            >
              <template #icon>
                <i class="i-ionicons-add" />
              </template>
              添加订阅
            </n-button>
            
            <div class="setting-item" style="margin-top: var(--space-6)">
              <n-button @click="updateSubscriptions" type="primary">
                更新所有订阅
              </n-button>
              <n-button @click="testAllNodes" type="default">
                测试所有节点
              </n-button>
            </div>
          </div>
        </n-tab-pane>
        
        <!-- 高级设置 -->
        <n-tab-pane name="advanced" tab="高级设置">
          <div class="settings-section">
            <h3>调试选项</h3>
            
            <div class="setting-item">
              <label>启用调试模式</label>
              <n-switch v-model:value="settings.debugMode" />
            </div>
            
            <div class="setting-item">
              <label>日志级别</label>
              <n-select 
                v-model:value="settings.logLevel" 
                :options="logLevelOptions"
                style="width: 200px"
              />
            </div>
            
            <h3>数据管理</h3>
            
            <div class="setting-item">
              <n-button @click="clearCache">清理缓存</n-button>
              <n-button @click="exportConfig">导出配置</n-button>
              <n-button @click="importConfig">导入配置</n-button>
            </div>
            
            <h3>关于</h3>
            
            <div class="about-info">
              <p>GPP 加速器 v1.4.6</p>
              <p>基于 sing-box + Wails 构建</p>
              <n-button text @click="checkUpdate">检查更新</n-button>
            </div>
          </div>
        </n-tab-pane>
      </n-tabs>
    </n-card>
    
    <!-- 保存按钮 -->
    <div class="settings-footer">
      <n-button @click="resetSettings" :disabled="!hasChanges">
        重置
      </n-button>
      <n-button type="primary" @click="saveSettings" :disabled="!hasChanges">
        保存设置
      </n-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter, onBeforeRouteLeave } from 'vue-router'
import { useMessage, NIcon } from 'naive-ui'

const router = useRouter()
const message = useMessage()

// 设置数据
const settings = reactive({
  // 基本设置
  theme: 'light',
  language: 'zh-CN',
  autoStart: false,
  minimizeToTray: true,
  
  // 网络设置
  proxyDns: '1.1.1.1',
  localDns: '223.5.5.5',
  autoSelectNode: false,
  connectionTimeout: 30,
  
  // 高级设置
  debugMode: false,
  logLevel: 'info'
})

// 原始设置（用于检测更改）
const originalSettings = reactive({ ...settings })

// 订阅列表
const subscriptions = ref([
  { url: '', editing: false }
])

// 选项列表
const themeOptions = [
  { label: '浅色', value: 'light' },
  { label: '深色', value: 'dark' },
  { label: '跟随系统', value: 'auto' }
]

const languageOptions = [
  { label: '简体中文', value: 'zh-CN' },
  { label: 'English', value: 'en-US' }
]

const logLevelOptions = [
  { label: 'Debug', value: 'debug' },
  { label: 'Info', value: 'info' },
  { label: 'Warning', value: 'warning' },
  { label: 'Error', value: 'error' }
]

// 计算属性
const hasChanges = computed(() => {
  return JSON.stringify(settings) !== JSON.stringify(originalSettings)
})

// 订阅管理
const addSubscription = () => {
  subscriptions.value.push({ url: '', editing: true })
}

const editSubscription = (index: number) => {
  subscriptions.value[index].editing = true
}

const saveSubscription = (index: number) => {
  subscriptions.value[index].editing = false
}

const removeSubscription = (index: number) => {
  subscriptions.value.splice(index, 1)
}

const updateSubscriptions = async () => {
  message.loading('正在更新订阅...')
  // TODO: 调用后端API更新订阅
  setTimeout(() => {
    message.success('订阅更新成功')
  }, 1000)
}

const testAllNodes = async () => {
  message.loading('正在测试节点...')
  // TODO: 调用后端API测试节点
  setTimeout(() => {
    message.success('节点测试完成')
  }, 2000)
}

// 高级设置操作
const clearCache = () => {
  message.warning('缓存已清理')
}

const exportConfig = () => {
  // TODO: 实现配置导出
  message.info('配置导出功能开发中...')
}

const importConfig = () => {
  // TODO: 实现配置导入
  message.info('配置导入功能开发中...')
}

const checkUpdate = () => {
  message.info('当前已是最新版本')
}

// 设置保存和重置
const saveSettings = async () => {
  try {
    // TODO: 调用后端API保存设置
    // await SaveConfig(settings)
    Object.assign(originalSettings, settings)
    message.success('设置已保存')
  } catch (error) {
    message.error('保存失败')
  }
}

const resetSettings = () => {
  Object.assign(settings, originalSettings)
  message.info('设置已重置')
}

// 路由守卫：检查未保存的更改
onBeforeRouteLeave((to, from) => {
  if (hasChanges.value) {
    const confirmLeave = confirm('有未保存的更改，确定要离开吗？')
    return confirmLeave
  }
  return true
})
</script>

<style scoped>
.settings-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  gap: var(--space-4);
}

.settings-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-4);
}

.back-button {
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  transition: all 0.3s;
}

.back-button:hover {
  color: var(--primary-color);
}

.settings-header h1 {
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-semibold);
  margin: 0;
}

.settings-card {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.settings-card :deep(.n-card__content) {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-4);
}

.settings-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.settings-section h3 {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-neutral-800);
  margin: var(--space-4) 0 var(--space-2);
}

.settings-section h3:first-child {
  margin-top: 0;
}

.setting-item {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-3) 0;
  border-bottom: 1px solid var(--color-neutral-100);
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-item label {
  flex: 1;
  font-size: var(--font-size-sm);
  color: var(--color-neutral-700);
}

.subscription-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.subscription-item {
  display: flex;
  gap: var(--space-2);
  align-items: center;
}

.subscription-item :deep(.n-input) {
  flex: 1;
}

.about-info {
  padding: var(--space-4);
  background: var(--color-neutral-50);
  border-radius: var(--radius-base);
  text-align: center;
}

.about-info p {
  margin: var(--space-2) 0;
  font-size: var(--font-size-sm);
  color: var(--color-neutral-600);
}

.settings-footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  padding: var(--space-4) 0;
  border-top: 1px solid var(--color-neutral-200);
}
</style>