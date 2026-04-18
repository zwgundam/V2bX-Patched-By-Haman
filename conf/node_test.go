package conf

import (
	"strings"
	"testing"
)

func TestMachineConfigUsesDefaultRuntimeOptions(t *testing.T) {
	var cfg MachineConfig
	data := []byte(`{
		"ApiHost": "http://127.0.0.1",
		"ApiKey": "test",
		"MachineID": 9
	}`)
	if err := cfg.UnmarshalJSON(data); err != nil {
		t.Fatalf("unmarshal machine config error: %v", err)
	}
	if cfg.ApiConfig.Timeout != 30 {
		t.Fatalf("unexpected timeout: %d", cfg.ApiConfig.Timeout)
	}
	if cfg.Options.ListenIP != "::" {
		t.Fatalf("unexpected listen ip: %s", cfg.Options.ListenIP)
	}
	if cfg.Options.DeviceOnlineMinTraffic != 200 {
		t.Fatalf("unexpected device traffic threshold: %d", cfg.Options.DeviceOnlineMinTraffic)
	}
	if cfg.Options.SingOptions == nil {
		t.Fatal("expected default sing options")
	}
	if cfg.Options.CertConfig == nil {
		t.Fatal("expected default cert config")
	}
}

func TestMachineConfigRejectsUnsupportedLocalFields(t *testing.T) {
	var cfg MachineConfig
	data := []byte(`{
		"ApiHost": "http://127.0.0.1",
		"ApiKey": "test",
		"MachineID": 9,
		"ListenIP": "::",
		"DeviceOnlineMinTraffic": 1,
		"EnableSniff": false
	}`)
	err := cfg.UnmarshalJSON(data)
	if err == nil {
		t.Fatal("expected unsupported field error")
	}
	if !strings.Contains(err.Error(), "unsupported machine config field") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMachineConfigRuntimeAPIConfigUsesPanelNodeType(t *testing.T) {
	cfg := MachineConfig{
		ApiConfig: ApiConfig{
			APIHost:      "http://127.0.0.1",
			APISendIP:    "127.0.0.2",
			MachineID:    9,
			Key:          "test",
			Timeout:      10,
			RuleListPath: "/tmp/rules.txt",
		},
	}
	apiConfig := cfg.RuntimeAPIConfig(3)
	if apiConfig.MachineID != 9 {
		t.Fatalf("unexpected machine id: %d", apiConfig.MachineID)
	}
	if apiConfig.NodeID != 3 {
		t.Fatalf("unexpected node id: %d", apiConfig.NodeID)
	}
	if apiConfig.NodeType != "v2node" {
		t.Fatalf("unexpected node type: %s", apiConfig.NodeType)
	}
	if apiConfig.RuleListPath != "/tmp/rules.txt" {
		t.Fatalf("unexpected rule list path: %s", apiConfig.RuleListPath)
	}
}
