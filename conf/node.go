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

type MachineConfig struct {
	ApiConfig ApiConfig `json:"-"`
	Options   Options   `json:"-"`
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

func (n *MachineConfig) UnmarshalJSON(data []byte) (err error) {
	data, raw, err := loadNodeConfigData(data)
	if err != nil {
		return err
	}
	if err = validateMachineConfigData(data); err != nil {
		return err
	}
	n.ApiConfig = ApiConfig{
		APIHost: "http://127.0.0.1",
		Timeout: 30,
	}
	if err = unmarshalNodeAPI(raw, data, &n.ApiConfig); err != nil {
		return err
	}
	n.Options = defaultOptions()
	return nil
}

func (n MachineConfig) RuntimeAPIConfig(nodeID int) ApiConfig {
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

func validateMachineConfigData(data []byte) error {
	raw := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for key := range raw {
		switch key {
		case "Include", "ApiConfig", "ApiHost", "ApiSendIP", "MachineID", "ApiKey", "Timeout", "RuleListPath":
		default:
			return fmt.Errorf("unsupported machine config field: %s", key)
		}
	}
	if apiRaw, ok := raw["ApiConfig"]; ok && len(apiRaw) > 0 && string(apiRaw) != "null" {
		apiConfig := make(map[string]json.RawMessage)
		if err := json.Unmarshal(apiRaw, &apiConfig); err != nil {
			return err
		}
		for key := range apiConfig {
			switch key {
			case "ApiHost", "ApiSendIP", "MachineID", "ApiKey", "Timeout", "RuleListPath":
			default:
				return fmt.Errorf("unsupported machine api field: %s", key)
			}
		}
	}
	return nil
}

func defaultOptions() Options {
	return Options{
		ListenIP:               "::",
		SendIP:                 "0.0.0.0",
		DeviceOnlineMinTraffic: 200,
		ReportMinTraffic:       0,
		SingOptions:            NewSingOptions(),
		CertConfig:             NewCertConfig(),
	}
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
