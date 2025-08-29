package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestPathManager 测试路径管理器
func TestPathManager(t *testing.T) {
	t.Run("GetConfigPath", func(t *testing.T) {
		pm := NewPathManager()
		path := pm.GetPath()
		if path == "" {
			t.Error("config path should not be empty")
		}
	})
	
	t.Run("EnsureDir", func(t *testing.T) {
		tmpDir := t.TempDir()
		pm := &PathManager{
			configPath: filepath.Join(tmpDir, "test", "config.json"),
		}
		
		err := pm.EnsureDir()
		if err != nil {
			t.Errorf("EnsureDir failed: %v", err)
		}
		
		// 检查目录是否创建
		dir := filepath.Dir(pm.configPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Error("directory was not created")
		}
	})
}

// TestConfigValidator 测试配置验证器
func TestConfigValidator(t *testing.T) {
	validator := NewConfigValidator()
	
	t.Run("ValidateEmptyConfig", func(t *testing.T) {
		conf := &Config{}
		err := validator.Validate(conf)
		if err != nil {
			t.Errorf("Validate failed: %v", err)
		}
		
		// 检查是否添加了直连节点
		hasDirect := false
		for _, peer := range conf.PeerList {
			if peer.Name == "直连" && peer.Protocol == "direct" {
				hasDirect = true
				break
			}
		}
		if !hasDirect {
			t.Error("direct peer was not added")
		}
		
		// 检查默认 DNS
		if conf.ProxyDNS != "https://1.1.1.1/dns-query" {
			t.Errorf("ProxyDNS not set correctly: %s", conf.ProxyDNS)
		}
		if conf.LocalDNS != "https://223.5.5.5/dns-query" {
			t.Errorf("LocalDNS not set correctly: %s", conf.LocalDNS)
		}
	})
	
	t.Run("ValidateWithExistingDirect", func(t *testing.T) {
		conf := &Config{
			PeerList: []*Peer{
				{Name: "直连", Protocol: "direct"},
				{Name: "节点1", Protocol: "vless"},
			},
		}
		
		err := validator.Validate(conf)
		if err != nil {
			t.Errorf("Validate failed: %v", err)
		}
		
		// 确保没有重复添加直连节点
		directCount := 0
		for _, peer := range conf.PeerList {
			if peer.Name == "直连" {
				directCount++
			}
		}
		if directCount != 1 {
			t.Errorf("Expected 1 direct peer, got %d", directCount)
		}
	})
}

// TestSubscriptionManager 测试订阅管理器
func TestSubscriptionManager(t *testing.T) {
	t.Run("UpdateFromSubscription", func(t *testing.T) {
		// 创建测试服务器
		testPeers := []*Peer{
			{Name: "test1", Protocol: "vless", Port: 443, Addr: "test1.com"},
			{Name: "test2", Protocol: "shadowsocks", Port: 8080, Addr: "test2.com"},
		}
		
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(testPeers)
		}))
		defer server.Close()
		
		sm := NewSubscriptionManager()
		existingPeers := []*Peer{
			{Name: "existing", Protocol: "direct"},
		}
		
		result, err := sm.UpdateFromSubscription(server.URL, existingPeers)
		if err != nil {
			t.Errorf("UpdateFromSubscription failed: %v", err)
		}
		
		// 检查结果
		if len(result) != 3 {
			t.Errorf("Expected 3 peers, got %d", len(result))
		}
		
		// 检查是否包含所有节点
		names := make(map[string]bool)
		for _, peer := range result {
			names[peer.Name] = true
		}
		
		if !names["existing"] || !names["test1"] || !names["test2"] {
			t.Error("Not all peers were merged correctly")
		}
	})
	
	t.Run("UpdateFromEmptySubscription", func(t *testing.T) {
		sm := NewSubscriptionManager()
		existingPeers := []*Peer{
			{Name: "existing", Protocol: "direct"},
		}
		
		result, err := sm.UpdateFromSubscription("", existingPeers)
		if err != nil {
			t.Errorf("UpdateFromSubscription failed: %v", err)
		}
		
		if len(result) != len(existingPeers) {
			t.Error("Peers should remain unchanged with empty subscription")
		}
	})
}

// TestConfigLoader 测试配置加载器
func TestConfigLoader(t *testing.T) {
	t.Run("InitAndSaveLoad", func(t *testing.T) {
		// 使用临时目录
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")
		
		// 创建临时配置加载器
		cl := &ConfigLoader{
			pathManager: &PathManager{configPath: configPath},
			subManager:  NewSubscriptionManager(),
			validator:   NewConfigValidator(),
		}
		
		// 初始化
		err := cl.Init()
		if err != nil {
			t.Errorf("Init failed: %v", err)
		}
		
		// 检查文件是否创建
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("config file was not created")
		}
		
		// 保存配置
		testConfig := &Config{
			PeerList: []*Peer{
				{Name: "test", Protocol: "vless", Port: 443},
			},
			GamePeer: "test",
			ProxyDNS: "https://8.8.8.8/dns-query",
		}
		
		err = cl.Save(testConfig)
		if err != nil {
			t.Errorf("Save failed: %v", err)
		}
		
		// 加载配置
		loadedConfig, err := cl.Load()
		if err != nil {
			t.Errorf("Load failed: %v", err)
		}
		
		// 验证加载的配置
		if loadedConfig.GamePeer != testConfig.GamePeer {
			t.Errorf("GamePeer mismatch: got %s, want %s", 
				loadedConfig.GamePeer, testConfig.GamePeer)
		}
		
		// 应该有两个节点（测试节点 + 自动添加的直连节点）
		if len(loadedConfig.PeerList) != 2 {
			t.Errorf("Expected 2 peers, got %d", len(loadedConfig.PeerList))
		}
	})
}

// TestPeerDomain 测试 Peer 的 Domain 方法
func TestPeerDomain(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		expected string
	}{
		{
			name:     "Domain",
			addr:     "example.com",
			expected: "example.com",
		},
		{
			name:     "DomainWithPort",
			addr:     "example.com:443",
			expected: "example.com",
		},
		{
			name:     "IPv4",
			addr:     "192.168.1.1",
			expected: "placeholder.com",
		},
		{
			name:     "IPv4WithPort",
			addr:     "192.168.1.1:8080",
			expected: "placeholder.com",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			peer := &Peer{Addr: tt.addr}
			domain := peer.Domain()
			if domain != tt.expected {
				t.Errorf("Domain() = %v, want %v", domain, tt.expected)
			}
		})
	}
}

// BenchmarkConfigLoad 性能测试
func BenchmarkConfigLoad(b *testing.B) {
	// 创建临时配置文件
	tmpDir := b.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	
	testConfig := &Config{
		PeerList: make([]*Peer, 100), // 100个节点
	}
	for i := 0; i < 100; i++ {
		testConfig.PeerList[i] = &Peer{
			Name:     fmt.Sprintf("node%d", i),
			Protocol: "vless",
			Port:     uint16(443 + i),
			Addr:     fmt.Sprintf("node%d.example.com", i),
		}
	}
	
	data, _ := json.Marshal(testConfig)
	os.WriteFile(configPath, data, 0o600)
	
	cl := &ConfigLoader{
		pathManager: &PathManager{configPath: configPath},
		subManager:  NewSubscriptionManager(),
		validator:   NewConfigValidator(),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cl.Load()
	}
}