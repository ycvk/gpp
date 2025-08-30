# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

GPP 是一个基于 sing-box + Wails 的跨平台 VPN/代理加速器桌面应用，支持 Windows、Linux 和 macOS。应用采用 Go 后端 + Vue.js 前端的混合架构，通过 Wails 框架桥接。

## 开发命令

### 前端开发
```bash
npm install          # 安装前端依赖
npm run dev          # 启动前端开发服务器（带热重载）
npm run build        # 构建前端生产版本
npm run type-check   # TypeScript 类型检查
```

#### 前端注意事项
- 使用 TypeScript strict 模式，注意类型定义
- ref 初始值为 null 时需要显式声明类型：`ref<string | null>(null)`
- 实现了智能轮询机制：状态变化时加快轮询，稳定后逐渐减慢

### 完整应用开发
```bash
wails dev           # 启动完整应用开发模式（前后端热重载）
wails build         # 构建生产版本
./build.sh          # 生产构建脚本（包含优化标志）
```

### 测试
```bash
go test ./...       # 运行所有 Go 单元测试
```

### 构建标志
生产构建使用：`-m -trimpath -tags webkit2_41,with_quic`

## 依赖版本

### 核心依赖
- **Go**: 1.23+ (sing-box v1.12+ 要求)
- **sing-box**: v1.12.4
- **Wails**: v2.10.2
- **Vue**: 3.x + TypeScript
- **Naive UI**: 最新版

### 重要 Go 包
```go
github.com/sagernet/sing-box v1.12.4
github.com/sagernet/sing v0.7.6
github.com/sagernet/sing-dns v0.4.6
github.com/sagernet/sing/common/json/badoption  // API 兼容性关键包
```

## 架构说明

### 核心组件
- **cmd/gpp/**: 主入口，包含 Wails 应用初始化和系统托盘集成
- **backend/**: Go 后端逻辑
  - `client/`: 客户端核心逻辑，包括代理管理
  - `config/`: 配置管理系统
  - `data/`: 数据模型和结构
- **frontend/**: Vue 3 + TypeScript 前端
  - 使用 Naive UI 组件库
  - Vite 构建工具
- **systray/**: 跨平台系统托盘实现

### 关键设计模式
1. **单实例管理**: 通过 TCP 端口 54713 防止重复启动
2. **双代理模式**: 
   - gamePeer: 游戏流量代理
   - httpPeer: HTTP 流量代理
3. **配置系统**: 支持本地和用户目录配置文件（`config.json`）
4. **协议支持**: 通过 sing-box 支持 VLESS、VMess、Shadowsocks、WireGuard 等多种协议

### 网络架构
- 使用 TUN 接口（utun225）进行流量劫持
- 内置 DNS 配置（代理 DNS: Cloudflare，本地 DNS: AliDNS）
- 支持订阅更新和节点 ping 测试
- 基于规则的流量路由（通过 sing-box options）

### 前后端通信
Wails 框架自动生成 Go 方法的 JavaScript 绑定，前端通过这些绑定调用后端功能。主要接口在 `app.go` 中定义。

## sing-box API 兼容性说明

### v1.12.4 API 关键变化

#### 1. Outbound 配置模式
```go
// 旧版本（< v1.12）
out = option.Outbound{
    Type: "shadowsocks",
    ShadowsocksOptions: option.ShadowsocksOutboundOptions{...},
}

// 新版本（v1.12.4）
out = option.Outbound{
    Type: "shadowsocks",
    Options: &option.ShadowsocksOutboundOptions{...},  // 使用通用 Options 字段
}
```

#### 2. DNS 配置结构
```go
// 必须使用 RawDNSOptions 包装
DNS: &option.DNSOptions{
    RawDNSOptions: option.RawDNSOptions{
        Servers: []option.DNSServerOptions{
            {
                Tag: "proxyDns",
                Options: &option.LegacyDNSServerOptions{...},
            },
        },
        Rules: []option.DNSRule{...},
    },
}
```

#### 3. DNS 规则 Action
```go
// DNS 规则使用专门的 DNSRouteActionOptions
DNSRuleAction: option.DNSRuleAction{
    RouteOptions: option.DNSRouteActionOptions{
        Server: "localDns",  // 不是 Outbound
    },
}
```

#### 4. 路由规则结构
```go
// 使用 RawDefaultRule + RuleAction 组合
DefaultOptions: option.DefaultRule{
    RawDefaultRule: option.RawDefaultRule{
        Protocol: badoption.Listable[string]{"dns"},  // 注意 badoption.Listable
    },
    RuleAction: option.RuleAction{
        RouteOptions: option.RouteActionOptions{
            Outbound: "dns_out",
        },
    },
}
```

#### 5. Listable 类型迁移
- 所有 `option.Listable[T]` 改为 `badoption.Listable[T]`
- 需要导入：`github.com/sagernet/sing/common/json/badoption`

#### 6. Inbound 配置
```go
// TUN/SOCKS 等都使用 Options 字段
{
    Type: "tun",
    Options: &option.TunInboundOptions{...},
}
```

## 重要配置文件
- `wails.json`: Wails 框架配置
- `go.mod`: Go 依赖管理
- `package.json`: 前端依赖管理
- `config.json`: 应用运行时配置（自动生成）

## 常见问题与解决方案

### 1. sing-box API 编译错误
**问题**: `unknown field XXXOptions in struct literal`
**解决**: 
- 检查是否使用了正确的 Options 字段模式
- 确认导入了 badoption 包
- 参考上面的 API 兼容性说明

### 2. TypeScript 类型错误
**问题**: `Type 'string' is not assignable to type 'null'`
**解决**:
```typescript
// 错误
let lastState = ref(null)
// 正确
let lastState = ref<string | null>(null)
```

### 3. Wails Dev 启动失败
**常见原因**:
- Go 版本过低（需要 1.23+）
- 前端依赖未安装：运行 `npm install`
- 端口被占用：检查 5173、34115 端口

### 4. DNS 配置错误
**问题**: `unknown field Route in struct`
**解决**: DNS 规则使用 `DNSRouteActionOptions` 而非 `RouteActionOptions`

## 调试技巧

### 后端调试
```bash
# 单独编译测试特定包
go build ./backend/client

# 查看详细错误
go build -v ./...

# 生成调试配置
config.Debug.Store(true)  # 启用调试模式，生成 sing.json
```

### 前端调试
- 开发服务器：http://localhost:5173
- Wails 调试界面：http://localhost:34115
- Vue DevTools 可用于组件调试

## 注意事项
- 应用需要管理员权限运行（TUN 接口需要）
- 窗口固定大小：360x520
- 支持后台运行（系统托盘）
- 配置自动保存机制
- macOS 构建包含私有 API，仅用于测试，不能上架 App Store