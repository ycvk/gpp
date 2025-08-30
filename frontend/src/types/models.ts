// 导入Wails生成的类型
import { config, data } from '../../wailsjs/go/models'

// 直接使用Wails生成的类型
export type Peer = config.Peer
export type WailsStatus = data.Status

// 扩展状态接口以适应前端需求  
export interface ProxyStatus {
  isRunning: boolean
  gamePeer: Peer | null
  httpPeer: Peer | null
  upSpeed: number
  downSpeed: number
  totalUp: number
  totalDown: number
}

export interface ConnectionState {
  status: 'idle' | 'connecting' | 'connected' | 'error'
  message?: string
  timestamp: Date
}

export interface TrafficStats {
  uploadSpeed: number
  downloadSpeed: number
  totalUpload: number
  totalDownload: number
  sessionStart: Date
}

export interface AppSettings {
  theme: 'light' | 'dark' | 'auto'
  autoStart: boolean
  minimizeToTray: boolean
  language: 'zh-CN' | 'en-US'
  updateChannel: 'stable' | 'beta'
}

// 事件类型
export interface AppEvents {
  'connection-state-change': ConnectionState
  'traffic-update': TrafficStats
  'node-change': { type: 'game' | 'http'; node: Peer }
  'error': Error
}

// API响应类型
export type ApiResponse<T = any> = {
  success: boolean
  data?: T
  error?: string
  timestamp: Date
}

// 组件Props类型辅助
export type ComponentProps<T> = T extends (...args: any) => any 
  ? Parameters<T>[0] 
  : T