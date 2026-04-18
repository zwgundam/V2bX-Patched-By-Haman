package sing

import (
	"net/netip"
	"testing"

	"github.com/MoeclubM/V2bX/api/panel"
	"github.com/MoeclubM/V2bX/conf"
	"github.com/sagernet/sing-box/option"
)

func TestGetInboundOptionsUsesInlineCertAndECH(t *testing.T) {
	inbound, err := getInboundOptions("test", &panel.NodeInfo{
		Type:     "vmess",
		Security: panel.Tls,
		Common: &panel.CommonNode{
			ServerPort: 443,
		},
		VAllss: &panel.VAllssNode{
			CommonNode: panel.CommonNode{
				ServerPort: 443,
			},
			Tls:     panel.Tls,
			Network: "ws",
			TlsSettings: panel.TlsSettings{
				ECH: &panel.ECHSettings{
					Enabled: true,
					Key:     panel.StringList{"ECH-KEY"},
				},
			},
		},
	}, &conf.Options{
		ListenIP:    "0.0.0.0",
		SingOptions: conf.NewSingOptions(),
		CertConfig: &conf.CertConfig{
			Certificate: []string{"CERT-PEM"},
			Key:         []string{"KEY-PEM"},
		},
	})
	if err != nil {
		t.Fatalf("get inbound options error: %v", err)
	}
	vmess, ok := inbound.Options.(*option.VMessInboundOptions)
	if !ok {
		t.Fatalf("unexpected inbound options type: %T", inbound.Options)
	}
	if vmess.TLS == nil || !vmess.TLS.Enabled {
		t.Fatal("expected tls to be enabled")
	}
	if len(vmess.TLS.Certificate) != 1 || vmess.TLS.Certificate[0] != "CERT-PEM" {
		t.Fatalf("unexpected inline certificate: %+v", vmess.TLS.Certificate)
	}
	if len(vmess.TLS.Key) != 1 || vmess.TLS.Key[0] != "KEY-PEM" {
		t.Fatalf("unexpected inline key: %+v", vmess.TLS.Key)
	}
	if vmess.TLS.ECH == nil || len(vmess.TLS.ECH.Key) != 1 || vmess.TLS.ECH.Key[0] != "ECH-KEY" {
		t.Fatalf("unexpected ech config: %+v", vmess.TLS.ECH)
	}
}

func TestGetInboundOptionsUsesFixedDualStackListenIP(t *testing.T) {
	inbound, err := getInboundOptions("test", &panel.NodeInfo{
		Type:     "vmess",
		Security: panel.None,
		Common: &panel.CommonNode{
			ListenIP:   "0.0.0.0",
			ServerPort: 443,
		},
		VAllss: &panel.VAllssNode{
			CommonNode: panel.CommonNode{
				ListenIP:   "0.0.0.0",
				ServerPort: 443,
			},
			Network: "ws",
		},
	}, &conf.Options{
		ListenIP:    "0.0.0.0",
		SingOptions: conf.NewSingOptions(),
		CertConfig:  conf.NewCertConfig(),
	})
	if err != nil {
		t.Fatalf("get inbound options error: %v", err)
	}
	vmess, ok := inbound.Options.(*option.VMessInboundOptions)
	if !ok {
		t.Fatalf("unexpected inbound options type: %T", inbound.Options)
	}
	if vmess.ListenOptions.Listen == nil {
		t.Fatal("expected listen address")
	}
	if netip.Addr(*vmess.ListenOptions.Listen).String() != "::" {
		t.Fatalf("unexpected listen ip: %s", netip.Addr(*vmess.ListenOptions.Listen).String())
	}
}
