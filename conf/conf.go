package conf

import (
	"fmt"
	"io"
	"os"

	stdjson "encoding/json"

	"github.com/MoeclubM/V2bX/common/json5"

	"encoding/json/v2"
)

type Conf struct {
	LogConfig   LogConfig       `json:"Log"`
	CoresConfig []CoreConfig    `json:"Cores"`
	Machines    []MachineConfig `json:"Machines"`
}

func (p *Conf) UnmarshalJSON(data []byte) error {
	raw := struct {
		LogConfig   LogConfig          `json:"Log"`
		CoresConfig []CoreConfig       `json:"Cores"`
		Machines    []MachineConfig    `json:"Machines"`
		Nodes       stdjson.RawMessage `json:"Nodes"`
		V2Config    stdjson.RawMessage `json:"V2"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Nodes) > 0 && string(raw.Nodes) != "null" {
		return fmt.Errorf("legacy Nodes field is no longer supported, use top-level Machines")
	}
	if len(raw.V2Config) > 0 && string(raw.V2Config) != "null" {
		return fmt.Errorf("legacy grouped config is no longer supported, use top-level Machines")
	}
	p.LogConfig = raw.LogConfig
	p.CoresConfig = raw.CoresConfig
	p.Machines = raw.Machines
	return nil
}

func New() *Conf {
	return &Conf{
		LogConfig: LogConfig{
			Level:  "info",
			Output: "",
		},
	}
}

func (p *Conf) LoadFromPath(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open config file error: %s", err)
	}
	defer f.Close()

	reader := json5.NewTrimNodeReader(f)
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("read config file error: %s", err)
	}

	err = json.Unmarshal(data, p)
	if err != nil {
		return fmt.Errorf("unmarshal config error: %s", err)
	}

	return nil
}
