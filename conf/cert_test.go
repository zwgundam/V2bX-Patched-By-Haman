package conf

import "testing"

func TestCertConfigSupportsPanelInlineFields(t *testing.T) {
	var cfg CertConfig
	err := cfg.UnmarshalJSON([]byte(`{
		"cert_mode":"none",
		"certificate":"CERT",
		"key":["KEY-1","KEY-2"],
		"dns_env":{"A":"B"}
	}`))
	if err != nil {
		t.Fatalf("unmarshal cert config error: %v", err)
	}
	if cfg.CertMode != "none" {
		t.Fatalf("unexpected cert mode: %s", cfg.CertMode)
	}
	if len(cfg.Certificate) != 1 || cfg.Certificate[0] != "CERT" {
		t.Fatalf("unexpected certificate: %+v", cfg.Certificate)
	}
	if len(cfg.Key) != 2 || cfg.Key[1] != "KEY-2" {
		t.Fatalf("unexpected key: %+v", cfg.Key)
	}
	if cfg.DNSEnv["A"] != "B" {
		t.Fatalf("unexpected dns env: %+v", cfg.DNSEnv)
	}
}
