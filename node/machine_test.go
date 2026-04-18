package node

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/MoeclubM/V2bX/api/panel"
	"github.com/MoeclubM/V2bX/conf"
	"github.com/MoeclubM/V2bX/core"
	"github.com/MoeclubM/V2bX/limiter"
)

type fakeCore struct {
	access sync.Mutex
	tags   map[string]struct{}
}

func (f *fakeCore) Start() error {
	return nil
}

func (f *fakeCore) Close() error {
	return nil
}

func (f *fakeCore) AddNode(tag string, info *panel.NodeInfo, config *conf.Options) error {
	f.access.Lock()
	defer f.access.Unlock()
	f.tags[tag] = struct{}{}
	return nil
}

func (f *fakeCore) DelNode(tag string) error {
	f.access.Lock()
	defer f.access.Unlock()
	delete(f.tags, tag)
	return nil
}

func (f *fakeCore) AddUsers(p *core.AddUsersParams) (int, error) {
	return len(p.Users), nil
}

func (f *fakeCore) GetUserTrafficSlice(tag string, reset bool) ([]panel.UserTraffic, error) {
	return nil, nil
}

func (f *fakeCore) DelUsers(users []panel.UserInfo, tag string, info *panel.NodeInfo) error {
	return nil
}

func (f *fakeCore) Protocols() []string {
	return nil
}

func (f *fakeCore) Type() string {
	return "fake"
}

func (f *fakeCore) TagCount() int {
	f.access.Lock()
	defer f.access.Unlock()
	return len(f.tags)
}

func TestMachineSyncNodes(t *testing.T) {
	var stage int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/server/machine/nodes":
			w.Header().Set("Content-Type", "application/json")
			switch atomic.LoadInt32(&stage) {
			case 0:
				_, _ = io.WriteString(w, `{"nodes":[{"id":1,"type":"vmess","name":"node-1"}],"base_config":{"push_interval":3600,"pull_interval":3600}}`)
			case 1:
				_, _ = io.WriteString(w, `{"nodes":[{"id":1,"type":"vmess","name":"node-1"},{"id":2,"type":"vmess","name":"node-2"}],"base_config":{"push_interval":3600,"pull_interval":3600}}`)
			default:
				_, _ = io.WriteString(w, `{"nodes":[{"id":2,"type":"vmess","name":"node-2"}],"base_config":{"push_interval":3600,"pull_interval":3600}}`)
			}
		case "/api/v2/server/handshake":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"websocket":{"enabled":false}}`)
		case "/api/v2/server/config":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"protocol":"vmess","server_port":443,"tls":0,"network":"ws","networkSettings":{"path":"/ws"},"base_config":{"push_interval":3600,"pull_interval":3600}}`)
		case "/api/v2/server/user":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"users":[{"id":1,"uuid":"uuid-1","speed_limit":0,"device_limit":0}]}`)
		case "/api/v2/server/alivelist":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"alive":{}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	core := &fakeCore{
		tags: make(map[string]struct{}),
	}
	limiter.Init()
	machine, err := NewMachine(core, &conf.MachineConfig{
		ApiConfig: conf.ApiConfig{
			APIHost:   server.URL,
			MachineID: 9,
			Key:       "token",
			Timeout:   30,
		},
		Options: conf.Options{
			ListenIP:   "0.0.0.0",
			SendIP:     "0.0.0.0",
			CertConfig: conf.NewCertConfig(),
		},
	})
	if err != nil {
		t.Fatalf("new machine error: %v", err)
	}
	defer machine.Close()

	if err = machine.Start(); err != nil {
		t.Fatalf("start machine error: %v", err)
	}
	if core.TagCount() != 1 {
		t.Fatalf("unexpected initial node count: %d", core.TagCount())
	}

	atomic.StoreInt32(&stage, 1)
	if err = machine.syncTask(); err != nil {
		t.Fatalf("sync machine add node error: %v", err)
	}
	if core.TagCount() != 2 {
		t.Fatalf("unexpected node count after add: %d", core.TagCount())
	}

	atomic.StoreInt32(&stage, 2)
	if err = machine.syncTask(); err != nil {
		t.Fatalf("sync machine remove node error: %v", err)
	}
	if core.TagCount() != 1 {
		t.Fatalf("unexpected node count after remove: %d", core.TagCount())
	}
}

func TestMachineReportStatus(t *testing.T) {
	var statusHits int32
	var statusBody panel.MachineStatus
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/server/machine/nodes":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"nodes":[{"id":1,"type":"vmess","name":"node-1"}],"base_config":{"push_interval":60,"pull_interval":3600}}`)
		case "/api/v2/server/machine/status":
			atomic.AddInt32(&statusHits, 1)
			if err := json.NewDecoder(r.Body).Decode(&statusBody); err != nil {
				t.Fatalf("decode machine status body error: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"data":true}`)
		case "/api/v2/server/handshake":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"websocket":{"enabled":false}}`)
		case "/api/v2/server/config":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"protocol":"vmess","server_port":443,"tls":0,"network":"ws","networkSettings":{"path":"/ws"},"base_config":{"push_interval":3600,"pull_interval":3600}}`)
		case "/api/v2/server/user":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"users":[{"id":1,"uuid":"uuid-1","speed_limit":0,"device_limit":0}]}`)
		case "/api/v2/server/alivelist":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"alive":{}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	core := &fakeCore{
		tags: make(map[string]struct{}),
	}
	limiter.Init()
	machine, err := NewMachine(core, &conf.MachineConfig{
		ApiConfig: conf.ApiConfig{
			APIHost:   server.URL,
			MachineID: 9,
			Key:       "token",
			Timeout:   30,
		},
		Options: conf.Options{
			ListenIP:   "0.0.0.0",
			SendIP:     "0.0.0.0",
			CertConfig: conf.NewCertConfig(),
		},
	})
	if err != nil {
		t.Fatalf("new machine error: %v", err)
	}
	machine.statusFunc = func() (*panel.MachineStatus, error) {
		return &panel.MachineStatus{
			CPU: 18.5,
			Mem: panel.MachineResource{
				Total: 1024,
				Used:  512,
			},
			Swap: panel.MachineResource{
				Total: 2048,
				Used:  256,
			},
			Disk: panel.MachineResource{
				Total: 4096,
				Used:  1024,
			},
		}, nil
	}
	defer machine.Close()

	if err = machine.Start(); err != nil {
		t.Fatalf("start machine error: %v", err)
	}
	if machine.statusTask == nil {
		t.Fatal("expected machine status task to be created")
	}
	if err = machine.reportStatusTask(); err != nil {
		t.Fatalf("report machine status error: %v", err)
	}
	if atomic.LoadInt32(&statusHits) != 1 {
		t.Fatalf("unexpected machine status hits: %d", atomic.LoadInt32(&statusHits))
	}
	if statusBody.CPU != 18.5 || statusBody.Mem.Used != 512 || statusBody.Disk.Total != 4096 {
		t.Fatalf("unexpected machine status payload: %+v", statusBody)
	}
}
