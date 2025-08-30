import { createApp } from 'vue'
import { createRouter, createWebHashHistory } from 'vue-router'
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
  NIcon,
  NTag,
  NText,
  NTooltip
} from 'naive-ui'
import App from './App.vue'
import routes from './router'
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
    NIcon,
    NTag,
    NText,
    NTooltip
  ]
})

// 路由系统
const router = createRouter({
  history: createWebHashHistory(),
  routes
})

const app = createApp(App)
app.use(naive)
app.use(router)
app.mount('#app')
