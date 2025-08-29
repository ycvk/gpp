package core

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/netip"
	"time"

	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/auth"
	"github.com/sagernet/sing/common/json/badoption"
)

func Server(conf Peer) error {
	var in option.Inbound
	switch conf.Protocol {
	case "shadowsocks":
		listenAddr := badoption.Addr(netip.MustParseAddr(conf.Addr))
		in = option.Inbound{
			Type: "shadowsocks",
			Tag:  "ss-in",
			Options: &option.ShadowsocksInboundOptions{
				ListenOptions: option.ListenOptions{
					Listen:     &listenAddr,
					ListenPort: conf.Port,
				},
				Method:   "aes-256-gcm",
				Password: conf.UUID,
				Multiplex: &option.InboundMultiplexOptions{
					Enabled: true,
				},
			},
		}
	case "socks":
		listenAddr := badoption.Addr(netip.MustParseAddr(conf.Addr))
		in = option.Inbound{
			Type: "socks",
			Tag:  "socks-in",
			Options: &option.SocksInboundOptions{
				ListenOptions: option.ListenOptions{
					Listen:     &listenAddr,
					ListenPort: conf.Port,
				},
				Users: []auth.User{
					{
						Username: "gpp",
						Password: conf.UUID,
					},
				},
			},
		}
	case "hysteria2":
		c, k := generateKey()
		listenAddr := badoption.Addr(netip.MustParseAddr(conf.Addr))
		in = option.Inbound{
			Type: "hysteria2",
			Tag:  "hy2-in",
			Options: &option.Hysteria2InboundOptions{
				ListenOptions: option.ListenOptions{
					Listen:     &listenAddr,
					ListenPort: conf.Port,
				},
				Users: []option.Hysteria2User{
					{
						Name:     "gpp",
						Password: conf.UUID,
					},
				},
				InboundTLSOptionsContainer: option.InboundTLSOptionsContainer{
					TLS: &option.InboundTLSOptions{
						Enabled:     true,
						ServerName:  "gpp",
						ALPN:        badoption.Listable[string]{"h3"},
						Certificate: badoption.Listable[string]{c},
						Key:         badoption.Listable[string]{k},
					},
				},
			},
		}
	default:
		listenAddr := badoption.Addr(netip.MustParseAddr(conf.Addr))
		in = option.Inbound{
			Type: "vless",
			Tag:  "vless-in",
			Options: &option.VLESSInboundOptions{
				ListenOptions: option.ListenOptions{
					Listen:     &listenAddr,
					ListenPort: conf.Port,
				},
				Users: []option.VLESSUser{
					{
						Name: "gpp",
						UUID: conf.UUID,
					},
				},
				Multiplex: &option.InboundMultiplexOptions{
					Enabled: true,
				},
			},
		}
	}
	var instance, err = box.New(box.Options{
		Context: context.Background(),
		Options: option.Options{
			Log: &option.LogOptions{
				Disabled:     false,
				Level:        "info",
				Output:       "run.log",
				Timestamp:    true,
				DisableColor: true,
			},
			Inbounds: []option.Inbound{in},
			Outbounds: []option.Outbound{
				{
					Type: "direct",
					Tag:  "direct-out",
				},
			},
		},
	})
	if err != nil {
		return err
	}
	err = instance.Start()
	if err != nil {
		return err
	}
	return nil
}
func generateKey() (string, string) {
	// 生成RSA密钥对
	pvk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", ""
	}

	// 设置证书信息
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"GPP"},
			CommonName:   "gpp",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
	}

	// 生成证书
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &pvk.PublicKey, pvk)
	if err != nil {
		return "", ""
	}
	buffer := bytes.NewBuffer([]byte{})
	_ = pem.Encode(buffer, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	buffer2 := bytes.NewBuffer([]byte{})
	pvkBytes, _ := x509.MarshalPKCS8PrivateKey(pvk)
	_ = pem.Encode(buffer2, &pem.Block{Type: "PRIVATE KEY", Bytes: pvkBytes})
	return buffer.String(), buffer2.String()
}