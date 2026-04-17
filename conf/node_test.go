package conf

import "testing"

func TestNodesConfigSupportsLegacyArray(t *testing.T) {
	var nodes NodesConfig
	data := []byte(`[
		{
			"ApiHost": "http://127.0.0.1",
			"ApiKey": "test",
			"NodeID": 1,
			"NodeType": "vmess",
			"ListenIP": "0.0.0.0"
		}
	]`)
	if err := nodes.UnmarshalJSON(data); err != nil {
		t.Fatalf("unmarshal legacy nodes error: %v", err)
	}
	if len(nodes.V1) != 1 {
		t.Fatalf("unexpected v1 node count: %d", len(nodes.V1))
	}
	if len(nodes.Machines) != 0 {
		t.Fatalf("unexpected machine config: %+v", nodes.Machines)
	}
	runtimeNodes := nodes.RuntimeNodeConfigs()
	if len(runtimeNodes) != 1 {
		t.Fatalf("unexpected runtime node count: %d", len(runtimeNodes))
	}
	if runtimeNodes[0].ApiConfig.NodeType != "vmess" {
		t.Fatalf("unexpected runtime node type: %s", runtimeNodes[0].ApiConfig.NodeType)
	}
}

func TestNodesConfigSupportsGroupedV1AndLegacyV2Machines(t *testing.T) {
	var nodes NodesConfig
	data := []byte(`{
		"V1": [
			{
				"ApiHost": "http://127.0.0.1",
				"ApiKey": "test",
				"NodeID": 1,
				"NodeType": "vmess"
			}
		],
		"V2": {
			"Machines": [
				{
					"ApiHost": "http://127.0.0.1",
					"ApiKey": "test",
					"MachineID": 9
				}
			]
		}
	}`)
	if err := nodes.UnmarshalJSON(data); err != nil {
		t.Fatalf("unmarshal grouped nodes error: %v", err)
	}
	if len(nodes.V1) != 1 {
		t.Fatalf("unexpected v1 node count: %d", len(nodes.V1))
	}
	if len(nodes.Machines) != 1 {
		t.Fatalf("unexpected machine count: %d", len(nodes.Machines))
	}
	runtimeNodes := nodes.RuntimeNodeConfigs()
	if len(runtimeNodes) != 1 {
		t.Fatalf("unexpected runtime node count: %d", len(runtimeNodes))
	}
}

func TestV2MachineConfigUsesDefaultRuntimeOptions(t *testing.T) {
	var cfg V2MachineConfig
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
	if cfg.Options.ListenIP != "0.0.0.0" {
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

func TestV2MachineConfigIgnoresLocalOptionsFromJSON(t *testing.T) {
	var cfg V2MachineConfig
	data := []byte(`{
		"ApiHost": "http://127.0.0.1",
		"ApiKey": "test",
		"MachineID": 9,
		"ListenIP": "::",
		"DeviceOnlineMinTraffic": 1,
		"EnableSniff": false
	}`)
	if err := cfg.UnmarshalJSON(data); err != nil {
		t.Fatalf("unmarshal machine config error: %v", err)
	}
	if cfg.Options.ListenIP != "0.0.0.0" {
		t.Fatalf("unexpected listen ip: %s", cfg.Options.ListenIP)
	}
	if cfg.Options.DeviceOnlineMinTraffic != 200 {
		t.Fatalf("unexpected device traffic threshold: %d", cfg.Options.DeviceOnlineMinTraffic)
	}
	if !cfg.Options.SingOptions.SniffEnabled {
		t.Fatal("expected default sniff setting")
	}
}
