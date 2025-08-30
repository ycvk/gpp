package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/sagernet/sing-box/option"
)

var Debug atomic.Bool

type Peer struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Port     uint16 `json:"port"`
	Addr     string `json:"addr"`
	UUID     string `json:"uuid"`
	Ping     uint   `json:"ping"`
}

func (p *Peer) Domain() string {
	host := strings.Split(p.Addr, ":")[0]
	_, err := netip.ParseAddr(host)
	if err != nil {
		return host
	}
	return "placeholder.com"
}

type Config struct {
	PeerList []*Peer       `json:"peer_list"`
	SubAddr  string        `json:"sub_addr"`
	Rules    []option.Rule `json:"rules"`
	GamePeer string        `json:"game_peer"`
	HTTPPeer string        `json:"http_peer"`
	ProxyDNS string        `json:"proxy_dns"`
	LocalDNS string        `json:"local_dns"`
	Debug    bool          `json:"debug"`
}

// InitConfig 初始化配置文件（使用新的配置加载器）
func InitConfig() {
	loader := NewConfigLoader()
	if err := loader.Init(); err != nil {
		fmt.Printf("Error initializing config: %v\n", err)
	}
}

// LoadConfig 加载配置（使用新的配置加载器）
func LoadConfig() (*Config, error) {
	loader := NewConfigLoader()
	return loader.Load()
}

// SaveConfig 保存配置（使用新的配置加载器）
func SaveConfig(config *Config) error {
	loader := NewConfigLoader()
	return loader.Save(config)
}
func ParsePeer(token string) (error, *Peer) {
	split := strings.Split(token, "#")
	name := ""
	if len(split) == 2 {
		token = split[0]
		name = split[1]
	}
	tokenBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return err, nil
	}
	token = string(tokenBytes)
	split = strings.Split(token, "@")
	protocol := strings.ReplaceAll(split[0], "gpp://", "")
	switch protocol {
	case "vless", "shadowsocks", "socks", "hysteria2":
	default:
		return fmt.Errorf("unknown protocol: %s", protocol), nil
	}
	if len(split) != 2 {
		return fmt.Errorf("invalid token: %s", token), nil
	}
	split = strings.Split(split[1], "/")
	addr := strings.Split(split[0], ":")
	if len(addr) != 2 {
		return errors.New("invalid addr: " + split[0]), nil
	}
	if len(split) != 2 {
		return fmt.Errorf("invalid token: %s", token), nil
	}
	uuid := split[1]
	if name == "" {
		name = fmt.Sprintf("%s:%s", addr[0], addr[1])
	}
	port, _ := strconv.ParseInt(addr[1], 10, 64)
	return nil, &Peer{
		Name:     name,
		Protocol: protocol,
		Port:     uint16(port),
		Addr:     addr[0],
		UUID:     uuid,
	}
}

// ParseSingBoxConfig 解析sing-box配置JSON，提取outbounds转换为Peer列表
func ParseSingBoxConfig(configJson string) ([]*Peer, error) {
	// 首先尝试解析为完整的sing-box配置
	var singBoxConfig struct {
		Outbounds []json.RawMessage `json:"outbounds"`
	}
	
	err := json.Unmarshal([]byte(configJson), &singBoxConfig)
	if err == nil && len(singBoxConfig.Outbounds) > 0 {
		// 解析多个outbounds
		var peers []*Peer
		for i, outboundRaw := range singBoxConfig.Outbounds {
			peer, err := parseSingBoxOutbound(outboundRaw, i)
			if err != nil {
				continue // 跳过无法解析的outbound
			}
			if peer != nil {
				peers = append(peers, peer)
			}
		}
		
		if len(peers) == 0 {
			return nil, fmt.Errorf("未找到有效的出站节点配置")
		}
		return peers, nil
	}
	
	// 尝试解析为单个节点配置
	var singleNode struct {
		Type       string `json:"type"`
		Tag        string `json:"tag"`
		Listen     string `json:"listen"`
		ListenPort uint16 `json:"listen_port"`
		Server     string `json:"server"`
		ServerPort uint16 `json:"server_port"`
	}
	
	err = json.Unmarshal([]byte(configJson), &singleNode)
	if err != nil {
		return nil, fmt.Errorf("无效的JSON格式: %v", err)
	}
	
	// 检查是否为inbound配置
	if singleNode.Listen != "" || singleNode.ListenPort != 0 {
		return nil, fmt.Errorf("这是一个入站(inbound)配置，不是出站(outbound)节点。请提供客户端连接配置")
	}
	
	// 尝试作为单个outbound解析
	peer, err := parseSingBoxOutbound(json.RawMessage(configJson), 0)
	if err != nil {
		return nil, fmt.Errorf("无法解析节点配置: %v", err)
	}
	
	if peer == nil {
		return nil, fmt.Errorf("不支持的节点类型或配置无效")
	}
	
	return []*Peer{peer}, nil
}

// parseSingBoxOutbound 解析单个outbound配置
func parseSingBoxOutbound(outboundRaw json.RawMessage, index int) (*Peer, error) {
	// 首先解析基础结构以获取类型
	var baseOutbound struct {
		Type string `json:"type"`
		Tag  string `json:"tag"`
	}
	
	err := json.Unmarshal(outboundRaw, &baseOutbound)
	if err != nil {
		return nil, err
	}
	
	// 跳过非代理类型的outbound
	switch baseOutbound.Type {
	case "shadowsocks":
		return parseShadowsocksOutbound(outboundRaw, baseOutbound.Tag, index)
	case "vless":
		return parseVLESSOutbound(outboundRaw, baseOutbound.Tag, index)
	case "vmess":
		return parseVMessOutbound(outboundRaw, baseOutbound.Tag, index)
	case "trojan":
		return parseTrojanOutbound(outboundRaw, baseOutbound.Tag, index)
	case "hysteria2":
		return parseHysteria2Outbound(outboundRaw, baseOutbound.Tag, index)
	case "direct", "block", "dns":
		return nil, nil // 跳过非代理outbound
	default:
		return nil, fmt.Errorf("unsupported outbound type: %s", baseOutbound.Type)
	}
}

// parseShadowsocksOutbound 解析shadowsocks outbound
func parseShadowsocksOutbound(outboundRaw json.RawMessage, tag string, index int) (*Peer, error) {
	var opts struct {
		Server     string `json:"server"`
		ServerPort uint16 `json:"server_port"`
		Password   string `json:"password"`
		Method     string `json:"method"`
	}
	
	err := json.Unmarshal(outboundRaw, &opts)
	if err != nil {
		return nil, fmt.Errorf("invalid shadowsocks outbound options: %v", err)
	}
	
	// 参数验证
	if opts.Server == "" || opts.ServerPort == 0 || opts.Password == "" {
		return nil, fmt.Errorf("shadowsocks outbound missing required parameters")
	}
	
	name := tag
	if name == "" {
		name = fmt.Sprintf("shadowsocks-%d", index+1)
	}
	
	return &Peer{
		Name:     name,
		Protocol: "shadowsocks",
		Addr:     opts.Server,
		Port:     opts.ServerPort,
		UUID:     opts.Password, // shadowsocks使用password字段
	}, nil
}

// parseVLESSOutbound 解析VLESS outbound
func parseVLESSOutbound(outboundRaw json.RawMessage, tag string, index int) (*Peer, error) {
	var opts struct {
		Server     string `json:"server"`
		ServerPort uint16 `json:"server_port"`
		UUID       string `json:"uuid"`
		Flow       string `json:"flow"`
	}
	
	err := json.Unmarshal(outboundRaw, &opts)
	if err != nil {
		return nil, fmt.Errorf("invalid vless outbound options: %v", err)
	}
	
	// 参数验证
	if opts.Server == "" || opts.ServerPort == 0 || opts.UUID == "" {
		return nil, fmt.Errorf("vless outbound missing required parameters")
	}
	
	name := tag
	if name == "" {
		name = fmt.Sprintf("vless-%d", index+1)
	}
	
	return &Peer{
		Name:     name,
		Protocol: "vless",
		Addr:     opts.Server,
		Port:     opts.ServerPort,
		UUID:     opts.UUID,
	}, nil
}

// parseVMessOutbound 解析VMess outbound
func parseVMessOutbound(outboundRaw json.RawMessage, tag string, index int) (*Peer, error) {
	var opts struct {
		Server     string `json:"server"`
		ServerPort uint16 `json:"server_port"`
		UUID       string `json:"uuid"`
		Security   string `json:"security"`
		AlterId    int    `json:"alter_id"`
	}
	
	err := json.Unmarshal(outboundRaw, &opts)
	if err != nil {
		return nil, fmt.Errorf("invalid vmess outbound options: %v", err)
	}
	
	// 参数验证
	if opts.Server == "" || opts.ServerPort == 0 || opts.UUID == "" {
		return nil, fmt.Errorf("vmess outbound missing required parameters")
	}
	
	name := tag
	if name == "" {
		name = fmt.Sprintf("vmess-%d", index+1)
	}
	
	// VMess在GPP中作为vless处理，因为getOUt函数中默认使用vless
	return &Peer{
		Name:     name,
		Protocol: "vless", // 映射为vless协议
		Addr:     opts.Server,
		Port:     opts.ServerPort,
		UUID:     opts.UUID,
	}, nil
}

// parseTrojanOutbound 解析Trojan outbound
func parseTrojanOutbound(outboundRaw json.RawMessage, tag string, index int) (*Peer, error) {
	var opts struct {
		Server     string `json:"server"`
		ServerPort uint16 `json:"server_port"`
		Password   string `json:"password"`
	}
	
	err := json.Unmarshal(outboundRaw, &opts)
	if err != nil {
		return nil, fmt.Errorf("invalid trojan outbound options: %v", err)
	}
	
	// 参数验证
	if opts.Server == "" || opts.ServerPort == 0 || opts.Password == "" {
		return nil, fmt.Errorf("trojan outbound missing required parameters")
	}
	
	name := tag
	if name == "" {
		name = fmt.Sprintf("trojan-%d", index+1)
	}
	
	// Trojan在GPP中作为vless处理
	return &Peer{
		Name:     name,
		Protocol: "vless", // 映射为vless协议
		Addr:     opts.Server,
		Port:     opts.ServerPort,
		UUID:     opts.Password, // trojan使用password字段
	}, nil
}

// parseHysteria2Outbound 解析Hysteria2 outbound
func parseHysteria2Outbound(outboundRaw json.RawMessage, tag string, index int) (*Peer, error) {
	var opts struct {
		Server     string `json:"server"`
		ServerPort uint16 `json:"server_port"`
		Password   string `json:"password"`
	}
	
	err := json.Unmarshal(outboundRaw, &opts)
	if err != nil {
		return nil, fmt.Errorf("invalid hysteria2 outbound options: %v", err)
	}
	
	// 参数验证
	if opts.Server == "" || opts.ServerPort == 0 || opts.Password == "" {
		return nil, fmt.Errorf("hysteria2 outbound missing required parameters")
	}
	
	name := tag
	if name == "" {
		name = fmt.Sprintf("hysteria2-%d", index+1)
	}
	
	return &Peer{
		Name:     name,
		Protocol: "hysteria2",
		Addr:     opts.Server,
		Port:     opts.ServerPort,
		UUID:     opts.Password,
	}, nil
}