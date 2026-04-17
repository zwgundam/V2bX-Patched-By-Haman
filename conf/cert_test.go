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

func TestCertConfigSupportsChallengeAddressAndPortAliases(t *testing.T) {
	var cfg CertConfig
	err := cfg.UnmarshalJSON([]byte(`{
		"mode":"http",
		"certificate_path":"/etc/V2bX/{machine_id}-{node_id}.pem",
		"key_path":"/etc/V2bX/{machine_id}-{node_id}.key",
		"listen_ip":"127.0.0.1",
		"listen_port":8080
	}`))
	if err != nil {
		t.Fatalf("unmarshal cert config error: %v", err)
	}
	if cfg.CertMode != "http" {
		t.Fatalf("unexpected cert mode: %s", cfg.CertMode)
	}
	if cfg.CertFile != "/etc/V2bX/{machine_id}-{node_id}.pem" {
		t.Fatalf("unexpected cert file: %s", cfg.CertFile)
	}
	if cfg.KeyFile != "/etc/V2bX/{machine_id}-{node_id}.key" {
		t.Fatalf("unexpected key file: %s", cfg.KeyFile)
	}
	if cfg.ChallengeAddress != "127.0.0.1" {
		t.Fatalf("unexpected challenge address: %s", cfg.ChallengeAddress)
	}
	if cfg.ChallengePort != "8080" {
		t.Fatalf("unexpected challenge port: %s", cfg.ChallengePort)
	}
}
