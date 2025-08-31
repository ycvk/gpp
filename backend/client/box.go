package client

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"os"
	"time"

	"github.com/danbai225/gpp/backend/config"
	"github.com/google/uuid"
	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/option"
	dns "github.com/sagernet/sing-dns"
	"github.com/sagernet/sing/common/json/badoption"
)

func getOUt(peer *config.Peer) option.Outbound {
	var out option.Outbound
	switch peer.Protocol {
	case "shadowsocks":
		out = option.Outbound{
			Type: "shadowsocks",
			Options: &option.ShadowsocksOutboundOptions{
				ServerOptions: option.ServerOptions{
					Server:     peer.Addr,
					ServerPort: peer.Port,
				},
				Method:   "aes-256-gcm",
				Password: peer.UUID,
				UDPOverTCP: &option.UDPOverTCPOptions{
					Enabled: true,
					Version: 2,
				},
				Multiplex: &option.OutboundMultiplexOptions{
					Enabled:        true,
					Protocol:       "h2mux",
					MaxConnections: 16,
					MinStreams:     32,
					Padding:        false,
				},
			},
		}
	case "socks":
		out = option.Outbound{
			Type: "socks",
			Options: &option.SOCKSOutboundOptions{
				ServerOptions: option.ServerOptions{
					Server:     peer.Addr,
					ServerPort: peer.Port,
				},
				Username: "gpp",
				Password: peer.UUID,
				UDPOverTCP: &option.UDPOverTCPOptions{
					Enabled: true,
					Version: 2,
				},
			},
		}
	case "hysteria2":
		out = option.Outbound{
			Type: "hysteria2",
			Options: &option.Hysteria2OutboundOptions{
				ServerOptions: option.ServerOptions{
					Server:     peer.Addr,
					ServerPort: peer.Port,
				},
				Password: peer.UUID,
				OutboundTLSOptionsContainer: option.OutboundTLSOptionsContainer{
					TLS: &option.OutboundTLSOptions{
						Enabled:    true,
						ServerName: "gpp",
						Insecure:   true,
						ALPN:       []string{"h3"},
					},
				},
				BrutalDebug: false,
			},
		}
	case "direct":
		out = option.Outbound{
			Type: "direct",
		}
	default:
		out = option.Outbound{
			Type: "vless",
			Options: &option.VLESSOutboundOptions{
				ServerOptions: option.ServerOptions{
					Server:     peer.Addr,
					ServerPort: peer.Port,
				},
				UUID: peer.UUID,
				Multiplex: &option.OutboundMultiplexOptions{
					Enabled:        true,
					Protocol:       "h2mux",
					MaxConnections: 16,
					MinStreams:     32,
					Padding:        false,
				},
			},
		}
	}
	out.Tag = uuid.New().String()
	return out
}
func Client(gamePeer, httpPeer *config.Peer, proxyDNS, localDNS string, rules []option.Rule) (*box.Box, error) {
	home, _ := os.UserHomeDir()
	proxyOut := getOUt(gamePeer)
	httpOut := proxyOut
	if httpPeer != nil {
		httpOut = getOUt(httpPeer)
	}
	httpOut.Tag = "http"
	proxyOut.Tag = "proxy"
	
	options := box.Options{
		Options: option.Options{
			Log: &option.LogOptions{
				Disabled: true,
			},
			DNS: &option.DNSOptions{
				RawDNSOptions: option.RawDNSOptions{
					Servers: []option.DNSServerOptions{
						{
							Tag: "proxyDns",
							Options: &option.LegacyDNSServerOptions{
								Address:  proxyDNS,
								Detour:   "proxy",
								Strategy: option.DomainStrategy(dns.DomainStrategyUseIPv4),
							},
						},
						{
							Tag: "localDns",
							Options: &option.LegacyDNSServerOptions{
								Address:  localDNS,
								Detour:   "direct",
								Strategy: option.DomainStrategy(dns.DomainStrategyUseIPv4),
							},
						},
						{
							Tag: "block",
							Options: &option.LegacyDNSServerOptions{
								Address:  "rcode://success",
								Strategy: option.DomainStrategy(dns.DomainStrategyUseIPv4),
							},
						},
					},
					Rules: []option.DNSRule{
						{
							Type: "default",
							DefaultOptions: option.DefaultDNSRule{
								RawDefaultDNSRule: option.RawDefaultDNSRule{
									Domain: badoption.Listable[string]{
										gamePeer.Domain(),
										httpPeer.Domain(),
									},
								},
								DNSRuleAction: option.DNSRuleAction{
									RouteOptions: option.DNSRouteActionOptions{
										Server: "localDns",
									},
								},
							},
						},
						{
							Type: "default",
							DefaultOptions: option.DefaultDNSRule{
								RawDefaultDNSRule: option.RawDefaultDNSRule{
									Geosite: badoption.Listable[string]{"cn"},
								},
								DNSRuleAction: option.DNSRuleAction{
									RouteOptions: option.DNSRouteActionOptions{
										Server: "localDns",
									},
								},
							},
						},
						{
							Type: "default",
							DefaultOptions: option.DefaultDNSRule{
								RawDefaultDNSRule: option.RawDefaultDNSRule{
									Geosite: badoption.Listable[string]{"geolocation-!cn"},
								},
								DNSRuleAction: option.DNSRuleAction{
									RouteOptions: option.DNSRouteActionOptions{
										Server: "proxyDns",
									},
								},
							},
						},
					},
					DNSClientOptions: option.DNSClientOptions{
						DisableCache: false,
					},
				},
			},
			Inbounds: []option.Inbound{
				{
					Type: "tun",
					Tag:  "tun-in",
					Options: &option.TunInboundOptions{
						InterfaceName: "utun225",
						MTU:           1420,
						Address: badoption.Listable[netip.Prefix]{
							netip.MustParsePrefix("172.25.0.1/30"),
						},
						AutoRoute:              true,
						StrictRoute:            true,
						EndpointIndependentNat: true,
						UDPTimeout:             option.UDPTimeoutCompat(time.Second * 300),
						Stack:                  "system",
						InboundOptions: option.InboundOptions{
							SniffEnabled: true,
						},
					},
				},
				{
					Type: "socks",
					Tag:  "socks-in",
					Options: &option.SocksInboundOptions{
						ListenOptions: option.ListenOptions{
							ListenPort: 5123,
							InboundOptions: option.InboundOptions{
								SniffEnabled: true,
							},
						},
					},
				},
			},
			Route: &option.RouteOptions{
				AutoDetectInterface: true,
				GeoIP: &option.GeoIPOptions{
					Path:           fmt.Sprintf("%s%c%s%c%s", home, os.PathSeparator, ".gpp", os.PathSeparator, "geoip.db"),
					DownloadURL:    "https://github.com/SagerNet/sing-geoip/releases/latest/download/geoip.db",
					DownloadDetour: "http",
				},
				Geosite: &option.GeositeOptions{
					Path:           fmt.Sprintf("%s%c%s%c%s", home, os.PathSeparator, ".gpp", os.PathSeparator, "geosite.db"),
					DownloadURL:    "https://github.com/SagerNet/sing-geosite/releases/latest/download/geosite.db",
					DownloadDetour: "http",
				},
				Rules: []option.Rule{
					{
						Type: "default",
						DefaultOptions: option.DefaultRule{
							RawDefaultRule: option.RawDefaultRule{
								Protocol: badoption.Listable[string]{"dns"},
							},
							RuleAction: option.RuleAction{
								RouteOptions: option.RouteActionOptions{
									Outbound: "dns_out",
								},
							},
						},
					},
					{
						Type: "default",
						DefaultOptions: option.DefaultRule{
							RawDefaultRule: option.RawDefaultRule{
								Inbound: badoption.Listable[string]{"dns_in"},
							},
							RuleAction: option.RuleAction{
								RouteOptions: option.RouteActionOptions{
									Outbound: "dns_out",
								},
							},
						},
					},
				},
			},
			Outbounds: []option.Outbound{
				proxyOut,
				httpOut,
				{
					Type: "block",
					Tag:  "block",
				},
				{
					Type: "direct",
					Tag:  "direct",
				}, {
					Type: "dns",
					Tag:  "dns_out",
				},
			},
		},
	}

	options.Options.Route.Rules = append(options.Options.Route.Rules, []option.Rule{
		{
			Type: "default",
			DefaultOptions: option.DefaultRule{
				RawDefaultRule: option.RawDefaultRule{
					Network: badoption.Listable[string]{"udp"},
					Port:    badoption.Listable[uint16]{443},
				},
				RuleAction: option.RuleAction{
					RouteOptions: option.RouteActionOptions{
						Outbound: "block",
					},
				},
			},
		},
		{
			Type: "default",
			DefaultOptions: option.DefaultRule{
				RawDefaultRule: option.RawDefaultRule{
					Geosite: badoption.Listable[string]{"cn"},
				},
				RuleAction: option.RuleAction{
					RouteOptions: option.RouteActionOptions{
						Outbound: "direct",
					},
				},
			},
		},
		{
			Type: "default",
			DefaultOptions: option.DefaultRule{
				RawDefaultRule: option.RawDefaultRule{
					GeoIP: badoption.Listable[string]{"cn", "private"},
				},
				RuleAction: option.RuleAction{
					RouteOptions: option.RouteActionOptions{
						Outbound: "direct",
					},
				},
			},
		},
		{
			Type: "default",
			DefaultOptions: option.DefaultRule{
				RawDefaultRule: option.RawDefaultRule{
					IPCIDR: badoption.Listable[string]{
						"85.236.96.0/21",
						"188.42.95.0/24",
						"188.42.147.0/24",
					},
				},
				RuleAction: option.RuleAction{
					RouteOptions: option.RouteActionOptions{
						Outbound: "direct",
					},
				},
			},
		},
		{
			Type: "default",
			DefaultOptions: option.DefaultRule{
				RawDefaultRule: option.RawDefaultRule{
					DomainSuffix: badoption.Listable[string]{
						"vivox.com",
						"cm.steampowered.com",
						"steamchina.com",
						"steamcontent.com",
						"steamserver.net",
						"steamusercontent.com",
						"csgo.wmsj.cn",
						"dl.steam.clngaa.com",
						"dl.steam.ksyna.com",
						"dota2.wmsj.cn",
						"st.dl.bscstorage.net",
						"st.dl.eccdnx.com",
						"st.dl.pinyuncloud.com",
						"steampipe.steamcontent.tnkjmec.com",
						"steampowered.com.8686c.com",
						"steamstatic.com.8686c.com",
						"wmsjsteam.com",
						"xz.pphimalayanrt.com",
					},
				},
				RuleAction: option.RuleAction{
					RouteOptions: option.RouteActionOptions{
						Outbound: "direct",
					},
				},
			},
		},
	}...)
	options.Options.Route.Rules = append(options.Options.Route.Rules, rules...)
	// http
	if httpPeer != nil && httpPeer.Name != gamePeer.Name {
		options.Options.Route.Rules = append(options.Options.Route.Rules, option.Rule{
			Type: "default",
			DefaultOptions: option.DefaultRule{
				RawDefaultRule: option.RawDefaultRule{
					Protocol: badoption.Listable[string]{"http"},
				},
				RuleAction: option.RuleAction{
					RouteOptions: option.RouteActionOptions{
						Outbound: httpOut.Tag,
					},
				},
			},
		})
		options.Options.Route.Rules = append(options.Options.Route.Rules, option.Rule{
			Type: "default",
			DefaultOptions: option.DefaultRule{
				RawDefaultRule: option.RawDefaultRule{
					Network: badoption.Listable[string]{"tcp"},
					Port:    badoption.Listable[uint16]{80, 443, 8080, 8443},
				},
				RuleAction: option.RuleAction{
					RouteOptions: option.RouteActionOptions{
						Outbound: httpOut.Tag,
					},
				},
			},
		})
	}
	if config.Debug.Load() {
		options.Log = &option.LogOptions{
			Disabled:     false,
			Level:        "trace",
			Output:       "debug.log",
			Timestamp:    true,
			DisableColor: true,
		}
		indent, _ := json.MarshalIndent(options, "", " ")
		_ = os.WriteFile("sing.json", indent, os.ModePerm)
	}
	var instance, err = box.New(options)
	if err != nil {
		return nil, err
	}
	return instance, nil
}