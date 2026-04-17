package conf

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"encoding/json/v2"
)

func TestConf_LoadFromPath(t *testing.T) {
	c := New()
	t.Log(c.LoadFromPath("../example/config.json"), c.NodeConfig)
}

func TestConfSupportsTopLevelV2Machines(t *testing.T) {
	c := New()
	data := []byte(`{
		"Nodes": {
			"V1": [
				{
					"ApiHost": "http://127.0.0.1",
					"ApiKey": "test",
					"NodeID": 1,
					"NodeType": "vmess"
				}
			]
		},
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
	if err := json.Unmarshal(data, c); err != nil {
		t.Fatalf("unmarshal config error: %v", err)
	}
	if len(c.NodeConfig.V1) != 1 {
		t.Fatalf("unexpected v1 node count: %d", len(c.NodeConfig.V1))
	}
	if len(c.NodeConfig.Machines) != 1 {
		t.Fatalf("unexpected machine count: %d", len(c.NodeConfig.Machines))
	}
	if len(c.NodeConfig.RuntimeNodeConfigs()) != 1 {
		t.Fatalf("unexpected runtime node count: %d", len(c.NodeConfig.RuntimeNodeConfigs()))
	}
}

func TestConfTopLevelV2OverridesLegacyNodesV2(t *testing.T) {
	c := New()
	data := []byte(`{
		"Nodes": {
			"V1": [],
			"V2": {
				"Machines": [
					{
						"ApiHost": "http://legacy.example",
						"ApiKey": "legacy",
						"MachineID": 1
					}
				]
			}
		},
		"V2": {
			"Machines": [
				{
					"ApiHost": "http://new.example",
					"ApiKey": "new",
					"MachineID": 2
				}
			]
		}
	}`)
	if err := json.Unmarshal(data, c); err != nil {
		t.Fatalf("unmarshal config error: %v", err)
	}
	if len(c.NodeConfig.Machines) != 1 {
		t.Fatalf("unexpected machine count: %d", len(c.NodeConfig.Machines))
	}
	if c.NodeConfig.Machines[0].ApiConfig.MachineID != 2 {
		t.Fatalf("unexpected machine id: %d", c.NodeConfig.Machines[0].ApiConfig.MachineID)
	}
}

func TestConf_Watch(t *testing.T) {
	c := New()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(configPath, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	reloaded := make(chan struct{}, 1)
	err := c.Watch(configPath, "", "", func() {
		select {
		case reloaded <- struct{}{}:
		default:
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte(`{"changed":true}`), 0644); err != nil {
		t.Fatal(err)
	}
	select {
	case <-reloaded:
	case <-time.After(8 * time.Second):
		t.Fatal("watch callback timeout")
	}
}
