package conf

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"encoding/json"

	"github.com/MoeclubM/V2bX/common/json5"
)

type NodeConfig struct {
	ApiConfig ApiConfig `json:"-"`
	Options   Options   `json:"-"`
}

type NodesConfig struct {
	V1       []V1NodeConfig    `json:"V1"`
	V2Nodes  []V2NodeConfig    `json:"-"`
	Machines []V2MachineConfig `json:"-"`
}

type V2ConfigGroup struct {
	Nodes    []V2NodeConfig    `json:"Nodes"`
	Machines []V2MachineConfig `json:"Machines"`
}

type V1NodeConfig struct {
	ApiConfig V1ApiConfig `json:"-"`
	Options   Options     `json:"-"`
}

type V2NodeConfig struct {
	ApiConfig V2NodeApiConfig `json:"-"`
	Options   Options         `json:"-"`
}

type V2MachineConfig struct {
	ApiConfig V2MachineApiConfig `json:"-"`
	Options   Options            `json:"-"`
}

type rawNodeConfig struct {
	Include string          `json:"Include"`
	ApiRaw  json.RawMessage `json:"ApiConfig"`
	OptRaw  json.RawMessage `json:"Options"`
}

type ApiConfig struct {
	APIHost      string `json:"ApiHost"`
	APISendIP    string `json:"ApiSendIP"`
	MachineID    int    `json:"MachineID"`
	NodeID       int    `json:"NodeID"`
	Key          string `json:"ApiKey"`
	NodeType     string `json:"NodeType"`
	Timeout      int    `json:"Timeout"`
	RuleListPath string `json:"RuleListPath"`
}

type V1ApiConfig struct {
	APIHost      string `json:"ApiHost"`
	APISendIP    string `json:"ApiSendIP"`
	NodeID       int    `json:"NodeID"`
	Key          string `json:"ApiKey"`
	NodeType     string `json:"NodeType"`
	Timeout      int    `json:"Timeout"`
	RuleListPath string `json:"RuleListPath"`
}

type V2NodeApiConfig struct {
	APIHost      string `json:"ApiHost"`
	APISendIP    string `json:"ApiSendIP"`
	MachineID    int    `json:"MachineID"`
	NodeID       int    `json:"NodeID"`
	Key          string `json:"ApiKey"`
	Timeout      int    `json:"Timeout"`
	RuleListPath string `json:"RuleListPath"`
}

type V2MachineApiConfig struct {
	APIHost      string `json:"ApiHost"`
	APISendIP    string `json:"ApiSendIP"`
	MachineID    int    `json:"MachineID"`
	Key          string `json:"ApiKey"`
	Timeout      int    `json:"Timeout"`
	RuleListPath string `json:"RuleListPath"`
}

func (n *NodesConfig) UnmarshalJSON(data []byte) error {
	legacy := make([]V1NodeConfig, 0)
	if err := json.Unmarshal(data, &legacy); err == nil {
		n.V1 = legacy
		n.V2Nodes = nil
		n.Machines = nil
		return nil
	}
	grouped := struct {
		V1 []V1NodeConfig `json:"V1"`
		V2 V2ConfigGroup  `json:"V2"`
	}{}
	if err := json.Unmarshal(data, &grouped); err != nil {
		return err
	}
	n.V1 = grouped.V1
	n.V2Nodes = grouped.V2.Nodes
	n.Machines = grouped.V2.Machines
	return nil
}

func (n NodesConfig) RuntimeNodeConfigs() []NodeConfig {
	configs := make([]NodeConfig, 0, len(n.V1)+len(n.V2Nodes))
	for i := range n.V1 {
		configs = append(configs, n.V1[i].RuntimeNodeConfig())
	}
	for i := range n.V2Nodes {
		configs = append(configs, n.V2Nodes[i].RuntimeNodeConfig())
	}
	return configs
}

func (n *V1NodeConfig) UnmarshalJSON(data []byte) (err error) {
	data, raw, err := loadNodeConfigData(data)
	if err != nil {
		return err
	}
	n.ApiConfig = V1ApiConfig{
		APIHost: "http://127.0.0.1",
		Timeout: 30,
	}
	if err = unmarshalNodeAPI(raw, data, &n.ApiConfig); err != nil {
		return err
	}
	n.Options = defaultNodeOptions()
	return unmarshalNodeOptions(raw, data, &n.Options)
}

func (n *V2NodeConfig) UnmarshalJSON(data []byte) (err error) {
	data, raw, err := loadNodeConfigData(data)
	if err != nil {
		return err
	}
	n.ApiConfig = V2NodeApiConfig{
		APIHost: "http://127.0.0.1",
		Timeout: 30,
	}
	if err = unmarshalNodeAPI(raw, data, &n.ApiConfig); err != nil {
		return err
	}
	n.Options = defaultNodeOptions()
	return unmarshalNodeOptions(raw, data, &n.Options)
}

func (n *V2MachineConfig) UnmarshalJSON(data []byte) (err error) {
	data, raw, err := loadNodeConfigData(data)
	if err != nil {
		return err
	}
	n.ApiConfig = V2MachineApiConfig{
		APIHost: "http://127.0.0.1",
		Timeout: 30,
	}
	if err = unmarshalNodeAPI(raw, data, &n.ApiConfig); err != nil {
		return err
	}
	n.Options = defaultNodeOptions()
	return unmarshalNodeOptions(raw, data, &n.Options)
}

func (n V1NodeConfig) RuntimeNodeConfig() NodeConfig {
	return NodeConfig{
		ApiConfig: ApiConfig{
			APIHost:      n.ApiConfig.APIHost,
			APISendIP:    n.ApiConfig.APISendIP,
			NodeID:       n.ApiConfig.NodeID,
			Key:          n.ApiConfig.Key,
			NodeType:     n.ApiConfig.NodeType,
			Timeout:      n.ApiConfig.Timeout,
			RuleListPath: n.ApiConfig.RuleListPath,
		},
		Options: n.Options,
	}
}

func (n V2NodeConfig) RuntimeNodeConfig() NodeConfig {
	return NodeConfig{
		ApiConfig: ApiConfig{
			APIHost:      n.ApiConfig.APIHost,
			APISendIP:    n.ApiConfig.APISendIP,
			MachineID:    n.ApiConfig.MachineID,
			NodeID:       n.ApiConfig.NodeID,
			Key:          n.ApiConfig.Key,
			NodeType:     "v2node",
			Timeout:      n.ApiConfig.Timeout,
			RuleListPath: n.ApiConfig.RuleListPath,
		},
		Options: n.Options,
	}
}

func (n V2MachineConfig) RuntimeAPIConfig(nodeID int) ApiConfig {
	return ApiConfig{
		APIHost:      n.ApiConfig.APIHost,
		APISendIP:    n.ApiConfig.APISendIP,
		MachineID:    n.ApiConfig.MachineID,
		NodeID:       nodeID,
		Key:          n.ApiConfig.Key,
		NodeType:     "v2node",
		Timeout:      n.ApiConfig.Timeout,
		RuleListPath: n.ApiConfig.RuleListPath,
	}
}

func loadNodeConfigData(data []byte) ([]byte, rawNodeConfig, error) {
	rn := rawNodeConfig{}
	err := json.Unmarshal(data, &rn)
	if err != nil {
		return nil, rn, err
	}
	if len(rn.Include) != 0 {
		includePath := strings.TrimSpace(rn.Include)
		parsedURL, parseErr := url.Parse(includePath)
		if parseErr == nil && (parsedURL.Scheme == "http" || parsedURL.Scheme == "https") {
			rsp, err := http.Get(includePath)
			if err != nil {
				return nil, rn, err
			}
			defer rsp.Body.Close()
			data, err = io.ReadAll(json5.NewTrimNodeReader(rsp.Body))
			if err != nil {
				return nil, rn, fmt.Errorf("open include file error: %s", err)
			}
		} else {
			f, err := os.Open(includePath)
			if err != nil {
				return nil, rn, fmt.Errorf("open include file error: %s", err)
			}
			defer f.Close()
			data, err = io.ReadAll(json5.NewTrimNodeReader(f))
			if err != nil {
				return nil, rn, fmt.Errorf("open include file error: %s", err)
			}
		}
		err = json.Unmarshal(data, &rn)
		if err != nil {
			return nil, rn, fmt.Errorf("unmarshal include file error: %s", err)
		}
	}
	return data, rn, nil
}

func unmarshalNodeAPI(raw rawNodeConfig, data []byte, api any) error {
	if len(raw.ApiRaw) > 0 {
		return json.Unmarshal(raw.ApiRaw, api)
	}
	return json.Unmarshal(data, api)
}

func defaultNodeOptions() Options {
	return Options{
		ListenIP:   "0.0.0.0",
		SendIP:     "0.0.0.0",
		CertConfig: NewCertConfig(),
	}
}

func unmarshalNodeOptions(raw rawNodeConfig, data []byte, options *Options) error {
	if len(raw.OptRaw) > 0 {
		return json.Unmarshal(raw.OptRaw, options)
	}
	return json.Unmarshal(data, options)
}

type Options struct {
	Name                   string       `json:"Name"`
	ListenIP               string       `json:"ListenIP"`
	SendIP                 string       `json:"SendIP"`
	DeviceOnlineMinTraffic int64        `json:"DeviceOnlineMinTraffic"`
	ReportMinTraffic       int64        `json:"ReportMinTraffic"`
	LimitConfig            LimitConfig  `json:"LimitConfig"`
	SingOptions            *SingOptions `json:"SingOptions"`
	CertConfig             *CertConfig  `json:"CertConfig"`
}

func (o *Options) UnmarshalJSON(data []byte) error {
	type opt Options
	err := json.Unmarshal(data, (*opt)(o))
	if err != nil {
		return err
	}
	o.SingOptions = NewSingOptions()
	return json.Unmarshal(data, o.SingOptions)
}
