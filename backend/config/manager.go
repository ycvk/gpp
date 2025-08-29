package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// PathManager 管理配置文件路径
type PathManager struct {
	configPath string
}

// NewPathManager 创建路径管理器
func NewPathManager() *PathManager {
	pm := &PathManager{}
	pm.configPath = pm.getConfigPath()
	return pm
}

// getConfigPath 获取配置文件路径
func (pm *PathManager) getConfigPath() string {
	// 优先使用当前目录
	localPath := "config.json"
	if _, err := os.Stat(localPath); err == nil {
		return localPath
	}
	
	// 使用用户目录
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gpp", "config.json")
}

// GetPath 获取配置路径
func (pm *PathManager) GetPath() string {
	return pm.configPath
}

// EnsureDir 确保配置目录存在
func (pm *PathManager) EnsureDir() error {
	dir := filepath.Dir(pm.configPath)
	return os.MkdirAll(dir, 0o755)
}

// SubscriptionManager 管理订阅更新
type SubscriptionManager struct {
	httpClient *http.Client
}

// NewSubscriptionManager 创建订阅管理器
func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// UpdateFromSubscription 从订阅地址更新节点列表
func (sm *SubscriptionManager) UpdateFromSubscription(subAddr string, existingPeers []*Peer) ([]*Peer, error) {
	if subAddr == "" {
		return existingPeers, nil
	}
	
	resp, err := sm.httpClient.Get(subAddr)
	if err != nil {
		return nil, fmt.Errorf("fetch subscription failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subscription returned status: %d", resp.StatusCode)
	}
	
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read subscription failed: %w", err)
	}
	
	var newPeers []*Peer
	if err := json.Unmarshal(data, &newPeers); err != nil {
		return nil, fmt.Errorf("parse subscription failed: %w", err)
	}
	
	// 合并并去重
	peerMap := make(map[string]*Peer)
	for _, peer := range existingPeers {
		peerMap[peer.Name] = peer
	}
	for _, peer := range newPeers {
		peerMap[peer.Name] = peer
	}
	
	result := make([]*Peer, 0, len(peerMap))
	for _, peer := range peerMap {
		result = append(result, peer)
	}
	
	return result, nil
}

// ConfigValidator 配置验证器
type ConfigValidator struct{}

// NewConfigValidator 创建配置验证器
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

// Validate 验证并修正配置
func (cv *ConfigValidator) Validate(conf *Config) error {
	// 确保有节点列表
	if conf.PeerList == nil {
		conf.PeerList = make([]*Peer, 0)
	}
	
	// 确保有直连节点
	hasDirect := false
	for _, peer := range conf.PeerList {
		if peer.Name == "直连" && peer.Protocol == "direct" {
			hasDirect = true
			break
		}
	}
	if !hasDirect {
		conf.PeerList = append(conf.PeerList, &Peer{
			Name:     "直连",
			Protocol: "direct",
			Port:     0,
			Addr:     "127.0.0.1",
			UUID:     "",
			Ping:     0,
		})
	}
	
	// 设置默认 DNS
	if conf.ProxyDNS == "" {
		conf.ProxyDNS = "https://1.1.1.1/dns-query"
	}
	if conf.LocalDNS == "" {
		conf.LocalDNS = "https://223.5.5.5/dns-query"
	}
	
	// 激活调试模式
	if conf.Debug {
		Debug.Store(true)
	}
	
	return nil
}

// ConfigLoader 配置加载器（组合所有功能）
type ConfigLoader struct {
	pathManager    *PathManager
	subManager     *SubscriptionManager
	validator      *ConfigValidator
}

// NewConfigLoader 创建配置加载器
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{
		pathManager:    NewPathManager(),
		subManager:     NewSubscriptionManager(),
		validator:      NewConfigValidator(),
	}
}

// Load 加载配置（重构后的 LoadConfig）
func (cl *ConfigLoader) Load() (*Config, error) {
	// 读取配置文件
	data, err := os.ReadFile(cl.pathManager.GetPath())
	if err != nil {
		return nil, fmt.Errorf("read config file failed: %w", err)
	}
	
	// 解析配置
	conf := &Config{}
	if err := json.Unmarshal(data, conf); err != nil {
		return nil, fmt.Errorf("parse config failed: %w", err)
	}
	
	// 更新订阅
	if conf.SubAddr != "" {
		updatedPeers, err := cl.subManager.UpdateFromSubscription(conf.SubAddr, conf.PeerList)
		if err != nil {
			// 订阅更新失败不应该导致加载失败，只记录错误
			fmt.Printf("Warning: subscription update failed: %v\n", err)
		} else {
			conf.PeerList = updatedPeers
		}
	}
	
	// 验证和修正配置
	if err := cl.validator.Validate(conf); err != nil {
		return nil, fmt.Errorf("validate config failed: %w", err)
	}
	
	return conf, nil
}

// Save 保存配置
func (cl *ConfigLoader) Save(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config failed: %w", err)
	}
	
	if err := os.WriteFile(cl.pathManager.GetPath(), data, 0o600); err != nil {
		return fmt.Errorf("write config file failed: %w", err)
	}
	
	return nil
}

// Init 初始化配置文件
func (cl *ConfigLoader) Init() error {
	if err := cl.pathManager.EnsureDir(); err != nil {
		return err
	}
	
	// 如果文件不存在，创建默认配置
	if _, err := os.Stat(cl.pathManager.GetPath()); os.IsNotExist(err) {
		defaultConfig := &Config{
			PeerList: make([]*Peer, 0),
			ProxyDNS: "https://1.1.1.1/dns-query",
			LocalDNS: "https://223.5.5.5/dns-query",
		}
		return cl.Save(defaultConfig)
	}
	
	return nil
}