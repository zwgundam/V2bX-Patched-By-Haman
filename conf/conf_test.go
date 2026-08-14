package conf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-json-experiment/json"
)

func TestConfLoadFromPath(t *testing.T) {
	c := New()
	if err := c.LoadFromPath("../example/config.json"); err != nil {
		t.Fatalf("load config error: %v", err)
	}
	if len(c.Machines) != 2 {
		t.Fatalf("unexpected machine count: %d", len(c.Machines))
	}
}

func TestConfSupportsTopLevelMachines(t *testing.T) {
	c := New()
	data := []byte(`{
		"Machines": [
			{
				"ApiHost": "http://127.0.0.1",
				"ApiKey": "test",
				"MachineID": 9
			}
		]
	}`)
	if err := json.Unmarshal(data, c); err != nil {
		t.Fatalf("unmarshal config error: %v", err)
	}
	if len(c.Machines) != 1 {
		t.Fatalf("unexpected machine count: %d", len(c.Machines))
	}
	if c.Machines[0].ApiConfig.MachineID != 9 {
		t.Fatalf("unexpected machine id: %d", c.Machines[0].ApiConfig.MachineID)
	}
}

func TestConfSupportsMultipleTopLevelMachinesWithMinimalFields(t *testing.T) {
	c := New()
	data := []byte(`{
		"Machines": [
			{
				"ApiHost": "http://a.example",
				"ApiKey": "key-a",
				"MachineID": 1
			},
			{
				"ApiHost": "http://b.example",
				"ApiKey": "key-b",
				"MachineID": 2
			}
		]
	}`)
	if err := json.Unmarshal(data, c); err != nil {
		t.Fatalf("unmarshal config error: %v", err)
	}
	if len(c.Machines) != 2 {
		t.Fatalf("unexpected machine count: %d", len(c.Machines))
	}
	if c.Machines[1].ApiConfig.MachineID != 2 {
		t.Fatalf("unexpected machine id: %d", c.Machines[1].ApiConfig.MachineID)
	}
}

func TestConfRejectsLegacyNodesConfig(t *testing.T) {
	c := New()
	data := []byte(`{
		"Nodes": [
			{
				"ApiHost": "http://127.0.0.1",
				"ApiKey": "test",
				"NodeID": 1,
				"NodeType": "vmess"
			}
		]
	}`)
	err := json.Unmarshal(data, c)
	if err == nil {
		t.Fatal("expected legacy nodes config error")
	}
	if !strings.Contains(err.Error(), "legacy Nodes field") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfRejectsLegacyV2Config(t *testing.T) {
	c := New()
	data := []byte(`{
		"V2": {
			"Machines": [
				{
					"ApiHost": "http://127.0.0.1",
					"ApiKey": "test",
					"MachineID": 1
				}
			]
		}
	}`)
	err := json.Unmarshal(data, c)
	if err == nil {
		t.Fatal("expected legacy v2 config error")
	}
	if !strings.Contains(err.Error(), "legacy grouped config") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfWatch(t *testing.T) {
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
