package panel

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MoeclubM/V2bX/conf"
)

func TestClientGetNodeInfoFallsBackToV1(t *testing.T) {
	var handshakeHits int32
	var v1ConfigHits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/server/handshake":
			atomic.AddInt32(&handshakeHits, 1)
			http.NotFound(w, r)
		case "/api/v1/server/UniProxy/config":
			atomic.AddInt32(&v1ConfigHits, 1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"server_port":443,"tls":1,"network":"ws","networkSettings":{"path":"/ws"},"base_config":{"push_interval":60,"pull_interval":60}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := New(&conf.ApiConfig{
		APIHost:  server.URL,
		Key:      "token",
		NodeID:   1,
		NodeType: "vmess",
	})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}

	node, err := client.GetNodeInfo()
	if err != nil {
		t.Fatalf("get node info error: %v", err)
	}
	if node == nil {
		t.Fatal("expected node info")
	}
	if atomic.LoadInt32(&handshakeHits) != 1 {
		t.Fatalf("expected handshake probe once, got %d", atomic.LoadInt32(&handshakeHits))
	}
	if atomic.LoadInt32(&v1ConfigHits) != 1 {
		t.Fatalf("expected v1 config request once, got %d", atomic.LoadInt32(&v1ConfigHits))
	}
	if client.useV2API {
		t.Fatal("expected v1 fallback")
	}
	if client.serverPathPrefix != "/api/v1/server/UniProxy" {
		t.Fatalf("unexpected server path prefix: %s", client.serverPathPrefix)
	}
	if node.Type != "vmess" {
		t.Fatalf("unexpected node type: %s", node.Type)
	}
}

func TestClientGetNodeInfoDetectsV2AndNormalizesType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/server/handshake":
			if r.URL.Query().Get("machine_id") != "9" {
				t.Fatalf("missing machine_id in handshake query: %s", r.URL.RawQuery)
			}
			if r.URL.Query().Get("node_type") != "v2node" {
				t.Fatalf("unexpected node_type in handshake query: %s", r.URL.RawQuery)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"websocket":{"enabled":false}}`)
		case "/api/v2/server/config":
			if r.URL.Query().Get("machine_id") != "9" {
				t.Fatalf("missing machine_id in config query: %s", r.URL.RawQuery)
			}
			if r.URL.Query().Get("node_type") != "v2node" {
				t.Fatalf("unexpected node_type in config query: %s", r.URL.RawQuery)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"protocol":"hysteria","server_port":8443,"version":2,"up_mbps":100,"down_mbps":200,"obfs":"salamander","obfs-password":"secret","base_config":{"push_interval":60,"pull_interval":60}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := New(&conf.ApiConfig{
		APIHost:   server.URL,
		Key:       "token",
		MachineID: 9,
		NodeID:    3,
		NodeType:  "v2node",
	})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}

	node, err := client.GetNodeInfo()
	if err != nil {
		t.Fatalf("get node info error: %v", err)
	}
	if node == nil {
		t.Fatal("expected node info")
	}
	if !client.useV2API {
		t.Fatal("expected v2 api mode")
	}
	if client.serverPathPrefix != "/api/v2/server" {
		t.Fatalf("unexpected server path prefix: %s", client.serverPathPrefix)
	}
	if node.Type != "hysteria2" {
		t.Fatalf("unexpected normalized node type: %s", node.Type)
	}
	if client.NodeType != "hysteria2" {
		t.Fatalf("expected client node type to update, got %s", client.NodeType)
	}
}

func TestClientReportUserTrafficUsesV2Report(t *testing.T) {
	var body struct {
		Traffic map[string][]int64 `json:"traffic"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/server/handshake":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"websocket":{"enabled":false}}`)
		case "/api/v2/server/report":
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode report body error: %v", err)
			}
			_, _ = io.WriteString(w, `{"data":true}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := New(&conf.ApiConfig{
		APIHost:  server.URL,
		Key:      "token",
		NodeID:   1,
		NodeType: "v2node",
	})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}

	if err := client.ReportUserTraffic([]UserTraffic{{
		UID:      7,
		Upload:   11,
		Download: 22,
	}}); err != nil {
		t.Fatalf("report user traffic error: %v", err)
	}
	if len(body.Traffic) != 1 {
		t.Fatalf("unexpected report body: %+v", body)
	}
	if got := body.Traffic["7"]; len(got) != 2 || got[0] != 11 || got[1] != 22 {
		t.Fatalf("unexpected traffic payload: %+v", body.Traffic)
	}
}

func TestClientGetMachineNodes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/server/machine/nodes":
			if r.URL.Query().Get("machine_id") != "9" {
				t.Fatalf("missing machine_id in machine query: %s", r.URL.RawQuery)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"nodes":[{"id":3,"type":"vmess","name":"node-3"}],"base_config":{"push_interval":30,"pull_interval":60}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := New(&conf.ApiConfig{
		APIHost:   server.URL,
		Key:       "token",
		MachineID: 9,
		NodeType:  "v2node",
	})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}

	resp, err := client.GetMachineNodes()
	if err != nil {
		t.Fatalf("get machine nodes error: %v", err)
	}
	if len(resp.Nodes) != 1 || resp.Nodes[0].Id != 3 {
		t.Fatalf("unexpected machine nodes: %+v", resp.Nodes)
	}
	if resp.PushInterval() != 30*time.Second {
		t.Fatalf("unexpected push interval: %s", resp.PushInterval())
	}
	if resp.PullInterval() != 60*time.Second {
		t.Fatalf("unexpected pull interval: %s", resp.PullInterval())
	}
}

func TestClientReportMachineStatus(t *testing.T) {
	var body MachineStatus
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/server/machine/status":
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode machine status body error: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"data":true}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := New(&conf.ApiConfig{
		APIHost:   server.URL,
		Key:       "token",
		MachineID: 9,
		NodeType:  "v2node",
	})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}

	err = client.ReportMachineStatus(&MachineStatus{
		CPU: 12.5,
		Mem: MachineResource{
			Total: 1024,
			Used:  512,
		},
		Swap: MachineResource{
			Total: 2048,
			Used:  256,
		},
		Disk: MachineResource{
			Total: 4096,
			Used:  1024,
		},
	})
	if err != nil {
		t.Fatalf("report machine status error: %v", err)
	}
	if body.CPU != 12.5 || body.Mem.Used != 512 || body.Disk.Total != 4096 {
		t.Fatalf("unexpected machine status body: %+v", body)
	}
}

func TestStringListSupportsSingleValue(t *testing.T) {
	var node AnyTlsNode
	if err := json.Unmarshal([]byte(`{"padding_scheme":"stop=8"}`), &node); err != nil {
		t.Fatalf("unmarshal anytls node error: %v", err)
	}
	if len(node.PaddingScheme) != 1 || node.PaddingScheme[0] != "stop=8" {
		t.Fatalf("unexpected padding scheme: %+v", node.PaddingScheme)
	}
}
