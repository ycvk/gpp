import { createApp } from 'vue'
import {
  create,
  NMessageProvider,
  NConfigProvider,
  NLoadingBarProvider,
  NButton,
  NCard,
  NModal,
  NProgress,
  NSpace,
  NSelect,
  NInput,
  NInputNumber,
  NSwitch,
  NIcon,
  NTag,
  NText,
  NTooltip,
  NTabs,
  NTabPane,
  NResult,
  NSkeleton
} from 'naive-ui'
import App from './App.vue'
import router from './router'
import '@/styles/globals.css'

// Naive UI 组件注册
const naive = create({
  components: [
    NMessageProvider,
    NConfigProvider, 
    NLoadingBarProvider,
    NButton,
    NCard,
    NModal,
    NProgress,
    NSpace,
    NSelect,
    NInput,
    NInputNumber,
    NSwitch,
    NIcon,
    NTag,
    NText,
    NTooltip,
    NTabs,
    NTabPane,
    NResult,
    NSkeleton
  ]
})

const app = createApp(App)
app.use(naive)
app.use(router)
app.mount('#app')
