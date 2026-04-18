package panel

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MoeclubM/V2bX/conf"
)

func TestClientNewRequiresMachineID(t *testing.T) {
	_, err := New(&conf.ApiConfig{
		APIHost:  "http://127.0.0.1",
		Key:      "token",
		NodeID:   1,
		NodeType: "v2node",
	})
	if err == nil {
		t.Fatal("expected machine id error")
	}
	if !strings.Contains(err.Error(), "machine id") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientGetNodeInfoUsesPanelPathAndNormalizesType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/server/config":
			if r.URL.Query().Get("machine_id") != "9" {
				t.Fatalf("missing machine_id in config query: %s", r.URL.RawQuery)
			}
			if r.URL.Query().Get("node_id") != "3" {
				t.Fatalf("missing node_id in config query: %s", r.URL.RawQuery)
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
	if node.Type != "hysteria2" {
		t.Fatalf("unexpected normalized node type: %s", node.Type)
	}
	if client.NodeType != "hysteria2" {
		t.Fatalf("expected client node type to update, got %s", client.NodeType)
	}
}

func TestClientGetNodeInfoParsesCertAndECH(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/server/config" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"protocol":"vmess",
			"server_port":443,
			"tls":1,
			"network":"ws",
			"networkSettings":{"path":"/ws"},
			"tls_settings":{
				"server_name":"node.example.com",
				"ech":{
					"enabled":true,
					"key":"ECH-KEY",
					"config":"ECH-CONFIG",
					"query_server_name":"ech.example.com"
				}
			},
			"cert_config":{
				"cert_mode":"none",
				"certificate":"CERT-PEM",
				"key":"KEY-PEM"
			},
			"base_config":{"push_interval":60,"pull_interval":60}
		}`)
	}))
	defer server.Close()

	client, err := New(&conf.ApiConfig{
		APIHost:   server.URL,
		Key:       "token",
		MachineID: 9,
		NodeID:    1,
		NodeType:  "v2node",
	})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}

	node, err := client.GetNodeInfo()
	if err != nil {
		t.Fatalf("get node info error: %v", err)
	}
	if node == nil || node.Common == nil || node.Common.CertConfig == nil {
		t.Fatal("expected cert config")
	}
	if len(node.Common.CertConfig.Certificate) != 1 || node.Common.CertConfig.Certificate[0] != "CERT-PEM" {
		t.Fatalf("unexpected inline certificate: %+v", node.Common.CertConfig.Certificate)
	}
	ech := node.ECH()
	if ech == nil || !ech.Enabled {
		t.Fatal("expected ech config")
	}
	if len(ech.Key) != 1 || ech.Key[0] != "ECH-KEY" {
		t.Fatalf("unexpected ech key: %+v", ech.Key)
	}
	if len(ech.Config) != 1 || ech.Config[0] != "ECH-CONFIG" {
		t.Fatalf("unexpected ech config: %+v", ech.Config)
	}
	if ech.QueryServerName != "ech.example.com" {
		t.Fatalf("unexpected query server name: %s", ech.QueryServerName)
	}
}

func TestClientGetNodeInfoDoesNotBackfillCertDomainFromTLSSettings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/server/config" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"protocol":"hysteria",
			"version":2,
			"server_port":443,
			"tls_settings":{
				"server_name":"node.example.com"
			},
			"cert_config":{
				"cert_mode":"http",
				"email":"admin@example.com"
			},
			"base_config":{"push_interval":60,"pull_interval":60}
		}`)
	}))
	defer server.Close()

	client, err := New(&conf.ApiConfig{
		APIHost:   server.URL,
		Key:       "token",
		MachineID: 9,
		NodeID:    1,
		NodeType:  "v2node",
	})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}

	node, err := client.GetNodeInfo()
	if err != nil {
		t.Fatalf("get node info error: %v", err)
	}
	if node == nil || node.Common == nil || node.Common.CertConfig == nil {
		t.Fatal("expected cert config")
	}
	if node.Common.CertConfig.CertDomain != "" {
		t.Fatalf("unexpected cert domain: %s", node.Common.CertConfig.CertDomain)
	}
}

func TestClientGetNodeInfoUsesCertConfigDomain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/server/config" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"protocol":"anytls",
			"server_port":443,
			"tls_settings":{
				"server_name":"ignored.example.com"
			},
			"cert_config":{
				"cert_mode":"http",
				"domain":"node.example.com",
				"email":"admin@example.com",
				"http_port":"80"
			},
			"base_config":{"push_interval":60,"pull_interval":60}
		}`)
	}))
	defer server.Close()

	client, err := New(&conf.ApiConfig{
		APIHost:   server.URL,
		Key:       "token",
		MachineID: 9,
		NodeID:    1,
		NodeType:  "v2node",
	})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}

	node, err := client.GetNodeInfo()
	if err != nil {
		t.Fatalf("get node info error: %v", err)
	}
	if node == nil || node.Common == nil || node.Common.CertConfig == nil {
		t.Fatal("expected cert config")
	}
	if node.Common.CertConfig.CertDomain != "node.example.com" {
		t.Fatalf("unexpected cert domain: %s", node.Common.CertConfig.CertDomain)
	}
	if node.Common.CertConfig.ChallengePort != "80" {
		t.Fatalf("unexpected challenge port: %s", node.Common.CertConfig.ChallengePort)
	}
}

func TestClientReportUserTrafficUsesReportEndpoint(t *testing.T) {
	var body struct {
		Traffic map[string][]int64 `json:"traffic"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
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
		APIHost:   server.URL,
		Key:       "token",
		MachineID: 9,
		NodeID:    1,
		NodeType:  "v2node",
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
			if r.URL.Query().Get("node_id") != "" {
				t.Fatalf("unexpected node_id in machine query: %s", r.URL.RawQuery)
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
