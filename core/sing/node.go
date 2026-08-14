package sing

import (
	"fmt"
	"net/netip"
	"net/url"
	"strconv"
	"strings"

	"github.com/sagernet/sing-box/log"
	"time"

	"encoding/json"

	"github.com/MoeclubM/V2bX/api/panel"
	"github.com/MoeclubM/V2bX/conf"
	"github.com/sagernet/sing-box/option"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/common/json/badoption"
)

type HttpNetworkConfig struct {
	Header struct {
		Type     string           `json:"type"`
		Request  *json.RawMessage `json:"request"`
		Response *json.RawMessage `json:"response"`
	} `json:"header"`
}

type HttpRequest struct {
	Version string   `json:"version"`
	Method  string   `json:"method"`
	Path    []string `json:"path"`
	Headers struct {
		Host []string `json:"Host"`
	} `json:"headers"`
}

type WsNetworkConfig struct {
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
}

type GrpcNetworkConfig struct {
	ServiceName string `json:"serviceName"`
}

type HttpupgradeNetworkConfig struct {
	Path string `json:"path"`
	Host string `json:"host"`
}

type H2NetworkConfig struct {
	Host json.RawMessage `json:"host"`
	Path string          `json:"path"`
}

func getInboundOptions(tag string, info *panel.NodeInfo, c *conf.Options) (option.Inbound, error) {
	listenIP := "::"
	addr, err := netip.ParseAddr(listenIP)
	if err != nil {
		return option.Inbound{}, fmt.Errorf("the listen ip not vail")
	}
	listen := option.ListenOptions{
		Listen:      (*badoption.Addr)(&addr),
		ListenPort:  uint16(info.Common.ServerPort),
		TCPFastOpen: c.SingOptions.TCPFastOpen,
	}
	multiplexConfig := c.SingOptions.Multiplex
	if info.Type == "vless" && info.VLESS != nil && info.VLESS.Multiplex != nil {
		multiplexConfig = info.VLESS.Multiplex
	}
	var multiplex *option.InboundMultiplexOptions
	if multiplexConfig != nil {
		multiplexOption := option.InboundMultiplexOptions{
			Enabled: multiplexConfig.Enabled,
			Padding: multiplexConfig.Padding,
			Brutal: &option.BrutalOptions{
				Enabled:  multiplexConfig.Brutal.Enabled,
				UpMbps:   multiplexConfig.Brutal.UpMbps,
				DownMbps: multiplexConfig.Brutal.DownMbps,
			},
		}
		multiplex = &multiplexOption
	}
	var tls option.InboundTLSOptions
	switch info.Security {
	case panel.Tls:
		if c.CertConfig == nil {
			return option.Inbound{}, fmt.Errorf("the CertConfig is not vail")
		}
		if len(c.CertConfig.Certificate) > 0 || len(c.CertConfig.Key) > 0 {
			if len(c.CertConfig.Certificate) == 0 || len(c.CertConfig.Key) == 0 {
				return option.Inbound{}, fmt.Errorf("tls certificate or key not found")
			}
			tls.Enabled = true
			tls.Certificate = badoption.Listable[string](c.CertConfig.Certificate)
			tls.Key = badoption.Listable[string](c.CertConfig.Key)
		} else {
			switch c.CertConfig.CertMode {
			case "none", "":
				if c.CertConfig.CertFile == "" || c.CertConfig.KeyFile == "" {
					return option.Inbound{}, fmt.Errorf("tls certificate not found")
				}
			}
			tls.Enabled = true
			tls.CertificatePath = c.CertConfig.CertFile
			tls.KeyPath = c.CertConfig.KeyFile
		}
	case panel.Reality:
		tls.Enabled = true
		v := info.VLESS
		serverName := v.TlsSettings.ServerName
		port, _ := strconv.Atoi(v.TlsSettings.ServerPort)
		if port == 0 {
			port = 443
		}
		var dest string
		if v.TlsSettings.Dest != "" {
			dest = v.TlsSettings.Dest
		} else {
			dest = serverName
		}
		if serverName == "" && dest != "" {
			// 如果 ServerName 未填，自动提取 dest 的主机名/域名作为 SNI
			serverName = strings.Split(dest, ":")[0]
			log.Info(fmt.Sprintf("REALITY ServerName is empty, auto-fallback to dest host (%s) for tag %s", serverName, tag))
		}
		tls.ServerName = serverName

		mtd, _ := time.ParseDuration(v.RealityConfig.MaxTimeDiff)
		tls.Reality = &option.InboundRealityOptions{
			Enabled:    true,
			ShortID:    []string{v.TlsSettings.ShortId},
			PrivateKey: v.TlsSettings.PrivateKey,
			Handshake: option.InboundRealityHandshakeOptions{
				ServerOptions: option.ServerOptions{
					Server:     dest,
					ServerPort: uint16(port),
				},
			},
			MaxTimeDifference: badoption.Duration(mtd),
		}
	}
	if ech := info.ECH(); ech != nil && ech.Enabled {
		if len(ech.Key) == 0 && ech.KeyPath == "" {
			return option.Inbound{}, fmt.Errorf("ech key not found")
		}
		tls.ECH = &option.InboundECHOptions{
			Enabled: true,
			Key:     badoption.Listable[string](ech.Key),
			KeyPath: ech.KeyPath,
		}
	}
	in := option.Inbound{
		Tag: tag,
	}
	switch info.Type {
	case "vless":
		n := info.VLESS
		t := option.V2RayTransportOptions{
			Type: n.Network,
		}
		switch n.Network {
		case "tcp":
			if len(n.NetworkSettings) != 0 {
				network := HttpNetworkConfig{}
				err := json.Unmarshal(n.NetworkSettings, &network)
				if err != nil {
					return option.Inbound{}, fmt.Errorf("decode NetworkSettings error: %s", err)
				}
				if network.Header.Type == "http" {
					t.Type = network.Header.Type
					var request HttpRequest
					if network.Header.Request != nil {
						err = json.Unmarshal(*network.Header.Request, &request)
						if err != nil {
							return option.Inbound{}, fmt.Errorf("decode HttpRequest error: %s", err)
						}
						t.HTTPOptions.Host = badoption.Listable[string](request.Headers.Host)
						t.HTTPOptions.Path = request.Path[0]
						t.HTTPOptions.Method = request.Method
					}
				} else {
					t.Type = ""
				}
			} else {
				t.Type = ""
			}
		case "ws":
			var (
				path    string
				ed      int
				headers map[string]badoption.Listable[string]
			)
			if len(n.NetworkSettings) != 0 {
				network := WsNetworkConfig{}
				err := json.Unmarshal(n.NetworkSettings, &network)
				if err != nil {
					return option.Inbound{}, fmt.Errorf("decode NetworkSettings error: %s", err)
				}
				var u *url.URL
				u, err = url.Parse(network.Path)
				if err != nil {
					return option.Inbound{}, fmt.Errorf("parse path error: %s", err)
				}
				path = u.Path
				ed, _ = strconv.Atoi(u.Query().Get("ed"))
				headers = make(map[string]badoption.Listable[string], len(network.Headers))
				for k, v := range network.Headers {
					headers[k] = badoption.Listable[string]{
						v,
					}
				}
			}
			t.WebsocketOptions = option.V2RayWebsocketOptions{
				Path:                path,
				EarlyDataHeaderName: "Sec-WebSocket-Protocol",
				MaxEarlyData:        uint32(ed),
				Headers:             headers,
			}
		case "grpc":
			network := GrpcNetworkConfig{}
			if len(n.NetworkSettings) != 0 {
				err := json.Unmarshal(n.NetworkSettings, &network)
				if err != nil {
					return option.Inbound{}, fmt.Errorf("decode NetworkSettings error: %s", err)
				}
			}
			t.GRPCOptions = option.V2RayGRPCOptions{
				ServiceName: network.ServiceName,
			}
		case "httpupgrade":
			network := HttpupgradeNetworkConfig{}
			if len(n.NetworkSettings) != 0 {
				err := json.Unmarshal(n.NetworkSettings, &network)
				if err != nil {
					return option.Inbound{}, fmt.Errorf("decode NetworkSettings error: %s", err)
				}
			}
			t.HTTPUpgradeOptions = option.V2RayHTTPUpgradeOptions{
				Path: network.Path,
				Host: network.Host,
			}
		case "http", "h2":
			t.Type = "http"
			network := H2NetworkConfig{}
			if len(n.NetworkSettings) != 0 {
				err := json.Unmarshal(n.NetworkSettings, &network)
				if err != nil {
					return option.Inbound{}, fmt.Errorf("decode NetworkSettings error: %s", err)
				}
			}
			var host badoption.Listable[string]
			if len(network.Host) > 0 {
				var singleHost string
				if err := json.Unmarshal(network.Host, &singleHost); err == nil {
					host = badoption.Listable[string]{singleHost}
				} else {
					var hostList []string
					if err := json.Unmarshal(network.Host, &hostList); err == nil {
						host = badoption.Listable[string](hostList)
					} else {
						return option.Inbound{}, fmt.Errorf("decode NetworkSettings host error: %s", err)
					}
				}
			}
			t.HTTPOptions = option.V2RayHTTPOptions{
				Host: host,
				Path: network.Path,
			}
		case "quic":
			t.QUICOptions = option.V2RayQUICOptions{}
		}
		in.Type = "vless"
		in.Options = &option.VLESSInboundOptions{
			ListenOptions: listen,
			InboundTLSOptionsContainer: option.InboundTLSOptionsContainer{
				TLS: &tls,
			},
			Transport: &t,
			Multiplex: multiplex,
		}
	case "anytls":
		in.Type = "anytls"
		in.Options = &option.AnyTLSInboundOptions{
			ListenOptions: listen,
			PaddingScheme: []string(info.AnyTls.PaddingScheme),
			InboundTLSOptionsContainer: option.InboundTLSOptionsContainer{
				TLS: &tls,
			},
		}
	case "hysteria2":
		in.Type = "hysteria2"
		var obfs *option.Hysteria2Obfs
		if info.Hysteria2.ObfsType != "" && info.Hysteria2.ObfsPassword != "" {
			obfs = &option.Hysteria2Obfs{
				Type:     info.Hysteria2.ObfsType,
				Password: info.Hysteria2.ObfsPassword,
			}
		} else if info.Hysteria2.ObfsType != "" {
			obfs = &option.Hysteria2Obfs{
				Type:     "salamander",
				Password: info.Hysteria2.ObfsType,
			}
		}
		in.Options = &option.Hysteria2InboundOptions{
			ListenOptions:         listen,
			UpMbps:                info.Hysteria2.UpMbps,
			DownMbps:              info.Hysteria2.DownMbps,
			IgnoreClientBandwidth: info.Hysteria2.Ignore_Client_Bandwidth,
			Obfs:                  obfs,
			InboundTLSOptionsContainer: option.InboundTLSOptionsContainer{
				TLS: &tls,
			},
		}
	default:
		return option.Inbound{}, fmt.Errorf("unsupported node type: %s", info.Type)
	}
	return in, nil
}

func (b *Sing) AddNode(tag string, info *panel.NodeInfo, config *conf.Options) error {
	b.nodeReportMinTrafficBytes[tag] = config.ReportMinTraffic * 1024
	c, err := getInboundOptions(tag, info, config)
	if err != nil {
		return err
	}
	in := b.box.Inbound()
	_ = in.Remove(tag)
	// Initial settle after remove
	time.Sleep(200 * time.Millisecond)

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		err = in.Create(
			b.ctx,
			b.box.Router(),
			b.logFactory.NewLogger(F.ToString("inbound/", c.Type, "[", tag, "]")),
			tag,
			c.Type,
			c.Options,
		)
		if err == nil {
			return nil
		}
		lastErr = err
		errStr := err.Error()
		// Retry only on transient bind/listen failures (port in TIME_WAIT, FD reclaim)
		if !strings.Contains(errStr, "address already in use") && !strings.Contains(errStr, "bind:") {
			return fmt.Errorf("add inbound error: %s", err)
		}
		log.Warn(fmt.Sprintf("add inbound retry %d/3 for tag %s: %s", attempt, tag, err))
		// Re-remove to release any partial resources before retry
		_ = in.Remove(tag)
		time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
	}
	return fmt.Errorf("add inbound error after 3 retries: %s", lastErr)
}

func (b *Sing) DelNode(tag string) error {
	in := b.box.Inbound()
	err := in.Remove(tag)
	time.Sleep(50 * time.Millisecond)
	if err != nil {
		return fmt.Errorf("delete inbound error: %s", err)
	}
	return nil
}
