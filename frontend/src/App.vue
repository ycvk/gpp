<template>
  <n-config-provider :theme="theme">
    <n-message-provider>
      <n-loading-bar-provider>
        <AppLayout>
          <router-view />
        </AppLayout>
      </n-loading-bar-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'

// 主题系统初始化
const theme = computed(() => {
  // 未来可扩展为用户主题切换
  return null // 默认亮色主题
})
</script>

<style>
#app {
  height: 100vh;
  width: 360px;
  overflow: hidden;
  position: relative;
}
</style>
