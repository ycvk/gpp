package config

import (
	"encoding/base64"
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