<template>
  <n-config-provider :theme="theme">
    <n-message-provider>
      <n-loading-bar-provider>
        <AppLayout>
          <router-view v-slot="{ Component, route }">
            <transition name="page-transition" mode="out-in">
              <component :is="Component" :key="route.path" />
            </transition>
          </router-view>
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
  /* width removed - controlled by globals.css using var(--app-width) */
  overflow: hidden;
  position: relative;
}

/* 页面切换动画 - 适合桌面应用的轻量动画 */
.page-transition-enter-active,
.page-transition-leave-active {
  transition: all 0.25s ease-out;
}

.page-transition-enter-from {
  opacity: 0;
  transform: translateX(15px);
}

.page-transition-leave-to {
  opacity: 0;
  transform: translateX(-15px);
}

/* 确保切换过程中的定位 */
.page-transition-leave-active {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
}
</style>
