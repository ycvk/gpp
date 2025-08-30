package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cloverstd/tcping/ping"
	"github.com/danbai225/gpp/backend/client"
	"github.com/danbai225/gpp/backend/config"
	"github.com/danbai225/gpp/backend/data"
	"github.com/danbai225/gpp/backend/errors"
	"github.com/danbai225/gpp/systray"
	box "github.com/sagernet/sing-box"
	netutils "github.com/shirou/gopsutil/v3/net"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx      context.Context
	conf     *config.Config
	gamePeer *config.Peer
	httpPeer *config.Peer
	box      *box.Box
	lock     sync.Mutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	conf := config.Config{}
	app := App{
		conf: &conf,
	}
	return &app
}
func (a *App) systemTray() {
	systray.SetIcon(logo) // read the icon from a file
	show := systray.AddMenuItem("显示窗口", "显示窗口")
	systray.AddSeparator()
	exit := systray.AddMenuItem("退出加速器", "退出加速器")
	show.Click(func() { runtime.WindowShow(a.ctx) })
	exit.Click(func() {
		a.Stop()
		runtime.Quit(a.ctx)
		systray.Quit()
		time.Sleep(time.Second)
		os.Exit(0)
	})
	systray.SetOnClick(func(menu systray.IMenu) { runtime.WindowShow(a.ctx) })
	go func() {
		listener, err := net.Listen("tcp", "127.0.0.1:54713")
		if err != nil {
			_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
				Type:    runtime.ErrorDialog,
				Title:   "监听错误",
				Message: fmt.Sprintln("Error listening0:", err),
			})
		}
		var conn net.Conn
		for {
			conn, err = listener.Accept()
			if err != nil {
				_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
					Type:    runtime.ErrorDialog,
					Title:   "监听错误",
					Message: fmt.Sprintln("Error listening1:", err),
				})
				continue
			}
			// 读取指令
			buffer := make([]byte, 1024)
			n, err := conn.Read(buffer)
			if err != nil {
				_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
					Type:    runtime.ErrorDialog,
					Title:   "监听错误",
					Message: fmt.Sprintln("Error read:", err),
				})
				continue
			}
			command := string(buffer[:n])
			// 如果收到显示窗口的命令，则显示窗口
			if command == "SHOW_WINDOW" {
				// 展示窗口的代码
				runtime.WindowShow(a.ctx)
			}
			_ = conn.Close()
		}
	}()
}

func (a *App) testPing() {
	initialDelay := time.Second * 5 // 初始延迟5秒
	normalDelay := time.Second * 30 // 正常延迟30秒
	maxDelay := time.Minute * 2     // 最大延迟2分钟
	currentDelay := initialDelay
	consecutiveStable := 0 // 连续稳定次数

	for {
		a.PingAll()

		// 根据网络稳定性调整测试频率
		a.lock.Lock()
		isRunning := a.box != nil
		a.lock.Unlock()

		if isRunning {
			// 运行时减少测试频率
			currentDelay = maxDelay
		} else {
			// 未运行时，根据稳定性调整
			consecutiveStable++
			if consecutiveStable > 10 {
				// 稳定10次后使用正常延迟
				currentDelay = normalDelay
			} else if consecutiveStable > 20 {
				// 非常稳定，使用最大延迟
				currentDelay = maxDelay
			}
		}

		time.Sleep(currentDelay)
	}
}
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go systray.Run(a.systemTray, func() {})
	loadConfig, err := config.LoadConfig()
	if err != nil {
		_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.WarningDialog,
			Title:   "配置加载错误",
			Message: err.Error(),
		})
	} else {
		a.conf = loadConfig
	}
	if len(a.conf.PeerList) > 0 {
		if a.conf.GamePeer == "" {
			a.conf.GamePeer = a.conf.PeerList[0].Name
		} else {
			for _, peer := range a.conf.PeerList {
				if peer.Name == a.conf.GamePeer {
					a.gamePeer = peer
				}
			}
		}
		if a.conf.HTTPPeer == "" {
			a.conf.HTTPPeer = a.conf.PeerList[0].Name
		} else {
			for _, peer := range a.conf.PeerList {
				if peer.Name == a.conf.HTTPPeer {
					a.httpPeer = peer
				}
			}
		}
	}
	go a.testPing()
}
func (a *App) PingAll() {
	a.lock.Lock()
	if a.box != nil {
		a.lock.Unlock()
		return
	}
	a.lock.Unlock()

	// 使用 worker pool 限制并发数
	const maxWorkers = 5 // 最多同时 ping 5个节点

	type pingJob struct {
		peer *config.Peer
	}

	jobs := make(chan pingJob, len(a.conf.PeerList))
	results := make(chan struct{}, len(a.conf.PeerList))

	// 启动 worker pool
	var wg sync.WaitGroup
	for w := 0; w < maxWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				job.peer.Ping = pingPort(job.peer.Addr, job.peer.Port)
				results <- struct{}{}
			}
		}()
	}

	// 发送任务到队列
	jobCount := 0
	for i := range a.conf.PeerList {
		if a.conf.PeerList[i].Protocol == "direct" {
			continue
		}
		jobs <- pingJob{peer: a.conf.PeerList[i]}
		jobCount++
	}
	close(jobs)

	// 等待所有任务完成
	wg.Wait()
	close(results)
}

func (a *App) Status() *data.Status {
	a.lock.Lock()
	defer a.lock.Unlock()
	status := data.Status{
		Running:  a.box != nil,
		GamePeer: a.gamePeer,
		HttpPeer: a.httpPeer,
	}

	counters, _ := netutils.IOCounters(true)
	for _, counter := range counters {
		if counter.Name == "utun225" {
			status.Up = counter.BytesSent
			status.Down = counter.BytesRecv
		}
	}
	return &status
}

func (a *App) List() []*config.Peer {
	list := a.conf.PeerList
	sort.Slice(list, func(i, j int) bool { return list[i].Ping < list[j].Ping })
	return list
}
func (a *App) Add(token string) string {
	if a.conf.PeerList == nil {
		a.conf.PeerList = make([]*config.Peer, 0)
	}
	
	// sing-box配置检测
	if isValidSingBoxConfig(token) {
		return a.importSingBoxConfig(token)
	}
	
	if strings.HasPrefix(token, "http") {
		_, err := http.Get(token)
		if err != nil {
			_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
				Type:    runtime.ErrorDialog,
				Title:   "订阅错误",
				Message: err.Error(),
			})
			return err.Error()
		}
		a.conf.SubAddr = token
	} else {
		err, peer := config.ParsePeer(token)
		if err != nil {
			_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
				Type:    runtime.ErrorDialog,
				Title:   "导入错误",
				Message: err.Error(),
			})
			return err.Error()
		}
		for _, p := range a.conf.PeerList {
			if p.Name == peer.Name {
				_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
					Type:    runtime.ErrorDialog,
					Title:   "导入错误",
					Message: fmt.Sprintf("节点 %s 已存在", peer.Name),
				})
				return fmt.Sprintf("peer %s already exists", peer.Name)
			}
		}
		a.conf.PeerList = append(a.conf.PeerList, peer)
	}
	err := config.SaveConfig(a.conf)
	if err != nil {
		_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "导入错误",
			Message: err.Error(),
		})
		return err.Error()
	}
	return "ok"
}

// isValidSingBoxConfig 检测是否为sing-box配置
func isValidSingBoxConfig(content string) bool {
	content = strings.TrimSpace(content)
	// 检查是否为JSON格式
	if !strings.HasPrefix(content, "{") {
		return false
	}
	// 支持完整配置（包含outbounds）或单节点配置（包含type字段）
	return strings.Contains(content, "\"outbounds\"") || 
		   (strings.Contains(content, "\"type\"") && strings.Contains(content, "\"server\""))
}

// importSingBoxConfig 导入sing-box配置
func (a *App) importSingBoxConfig(configJson string) string {
	peers, err := config.ParseSingBoxConfig(configJson)
	if err != nil {
		_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "配置解析错误",
			Message: fmt.Sprintf("解析sing-box配置失败: %v", err),
		})
		return err.Error()
	}
	
	successCount, skipCount := a.checkAndMergePeers(peers)
	
	// 保存配置
	err = config.SaveConfig(a.conf)
	if err != nil {
		_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "保存配置失败",
			Message: err.Error(),
		})
		return err.Error()
	}
	
	// 返回结果信息
	if skipCount > 0 {
		message := fmt.Sprintf("成功导入 %d 个节点，跳过 %d 个重复节点", successCount, skipCount)
		_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.InfoDialog,
			Title:   "导入完成",
			Message: message,
		})
		return fmt.Sprintf("imported %d nodes, skipped %d duplicates", successCount, skipCount)
	}
	
	if successCount > 0 {
		_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.InfoDialog,
			Title:   "导入成功",
			Message: fmt.Sprintf("成功导入 %d 个节点", successCount),
		})
	}
	
	return "ok"
}

// checkAndMergePeers 批量节点去重和合并
func (a *App) checkAndMergePeers(newPeers []*config.Peer) (int, int) {
	successCount := 0
	skipCount := 0
	
	for _, newPeer := range newPeers {
		exists := false
		for _, existingPeer := range a.conf.PeerList {
			if existingPeer.Name == newPeer.Name {
				exists = true
				skipCount++
				break
			}
		}
		
		if !exists {
			a.conf.PeerList = append(a.conf.PeerList, newPeer)
			successCount++
		}
	}
	
	return successCount, skipCount
}
func (a *App) Del(Name string) string {
	for i, peer := range a.conf.PeerList {
		if peer.Name == Name {
			a.conf.PeerList = append(a.conf.PeerList[:i], a.conf.PeerList[i+1:]...)
			break
		}
	}
	err := config.SaveConfig(a.conf)
	if err != nil {
		return err.Error()
	}
	return "ok"
}
func (a *App) SetPeer(game, http string) string {
	for _, peer := range a.conf.PeerList {
		if peer.Name == game {
			a.gamePeer = peer
			a.conf.GamePeer = peer.Name
			break
		}
	}
	for _, peer := range a.conf.PeerList {
		if peer.Name == http {
			a.httpPeer = peer
			a.conf.HTTPPeer = peer.Name
			break
		}
	}
	err := config.SaveConfig(a.conf)
	if err != nil {
		_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "保存错误",
			Message: err.Error(),
		})
		return err.Error()
	}
	return "ok"
}

// Start 启动加速
func (a *App) Start() string {
	a.lock.Lock()
	defer a.lock.Unlock()
	if a.box != nil {
		return "running"
	}
	var err error
	a.box, err = client.Client(a.gamePeer, a.httpPeer, a.conf.ProxyDNS, a.conf.LocalDNS, a.conf.Rules)
	if err != nil {
		// 根据错误类型提供更友好的提示
		appErr := errors.NewNetworkError("创建代理客户端失败", err)

		// 特殊处理常见错误
		if strings.Contains(err.Error(), "permission") {
			appErr = errors.NewPermissionError("权限不足", err).
				WithUserMessage("需要管理员权限来创建网络接口").
				WithSuggestion("请以管理员身份运行程序")
		} else if strings.Contains(err.Error(), "address already in use") {
			appErr = errors.NewNetworkError("端口已被占用", err).
				WithUserMessage("代理端口已被其他程序占用").
				WithSuggestion("请检查是否有其他VPN或代理程序正在运行")
		}

		_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "加速失败",
			Message: fmt.Sprintf("%s\n\n建议：%s", appErr.UserMessage, appErr.Suggestion),
		})
		a.box = nil
		return appErr.Error()
	}
	err = a.box.Start()
	if err != nil {
		appErr := errors.NewSystemError("启动代理服务失败", err)

		// 检查特定错误类型
		if strings.Contains(err.Error(), "TUN") {
			appErr = errors.NewPermissionError("创建TUN接口失败", err).
				WithUserMessage("无法创建虚拟网络接口").
				WithSuggestion("请确保以管理员权限运行，并检查系统是否支持TUN设备")
		}

		_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "加速失败",
			Message: fmt.Sprintf("%s\n\n建议：%s", appErr.UserMessage, appErr.Suggestion),
		})
		a.box = nil
		return appErr.Error()
	}
	return "ok"
}

// Stop 停止加速
func (a *App) Stop() string {
	a.lock.Lock()
	defer a.lock.Unlock()
	if a.box == nil {
		return "not running"
	}
	err := a.box.Close()
	if err != nil {
		_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "停止失败",
			Message: err.Error(),
		})
		return err.Error()
	}
	a.box = nil
	return "ok"
}
func pingPort(host string, port uint16) uint {
	tcPing := ping.NewTCPing()
	tcPing.SetTarget(&ping.Target{
		Host:     host,
		Port:     int(port),
		Counter:  1,
		Interval: time.Millisecond * 200,
		Timeout:  time.Second * 3,
	})
	start := tcPing.Start()
	<-start
	result := tcPing.Result()
	return uint(result.Avg().Milliseconds())
}
func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	return io.ReadAll(resp.Body)
}