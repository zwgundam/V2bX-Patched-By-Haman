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
	if len(nodes.V2Nodes) != 0 || len(nodes.Machines) != 0 {
		t.Fatalf("unexpected v2 config: %+v %+v", nodes.V2Nodes, nodes.Machines)
	}
	runtimeNodes := nodes.RuntimeNodeConfigs()
	if len(runtimeNodes) != 1 {
		t.Fatalf("unexpected runtime node count: %d", len(runtimeNodes))
	}
	if runtimeNodes[0].ApiConfig.NodeType != "vmess" {
		t.Fatalf("unexpected runtime node type: %s", runtimeNodes[0].ApiConfig.NodeType)
	}
}

func TestNodesConfigSupportsGroupedConfig(t *testing.T) {
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
			"Nodes": [
				{
					"ApiHost": "http://127.0.0.1",
					"ApiKey": "test",
					"MachineID": 9,
					"NodeID": 2
				}
			],
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
	if len(nodes.V2Nodes) != 1 {
		t.Fatalf("unexpected v2 node count: %d", len(nodes.V2Nodes))
	}
	if len(nodes.Machines) != 1 {
		t.Fatalf("unexpected machine count: %d", len(nodes.Machines))
	}
	runtimeNodes := nodes.RuntimeNodeConfigs()
	if len(runtimeNodes) != 2 {
		t.Fatalf("unexpected runtime node count: %d", len(runtimeNodes))
	}
	if runtimeNodes[1].ApiConfig.NodeType != "v2node" {
		t.Fatalf("unexpected v2 runtime node type: %s", runtimeNodes[1].ApiConfig.NodeType)
	}
	if runtimeNodes[1].ApiConfig.MachineID != 9 {
		t.Fatalf("unexpected machine id: %d", runtimeNodes[1].ApiConfig.MachineID)
	}
}

func TestV2MachineConfigRuntimeAPIConfig(t *testing.T) {
	cfg := V2MachineConfig{
		ApiConfig: V2MachineApiConfig{
			APIHost:   "http://127.0.0.1",
			APISendIP: "127.0.0.2",
			MachineID: 9,
			Key:       "test",
			Timeout:   30,
		},
	}
	apiConfig := cfg.RuntimeAPIConfig(12)
	if apiConfig.NodeType != "v2node" {
		t.Fatalf("unexpected node type: %s", apiConfig.NodeType)
	}
	if apiConfig.NodeID != 12 {
		t.Fatalf("unexpected node id: %d", apiConfig.NodeID)
	}
	if apiConfig.MachineID != 9 {
		t.Fatalf("unexpected machine id: %d", apiConfig.MachineID)
	}
}
