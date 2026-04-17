package conf

import (
	"encoding/json"
	"fmt"
)

type CertConfig struct {
	CertMode         string            `json:"CertMode"` // none, file, http, dns
	RejectUnknownSni bool              `json:"RejectUnknownSni"`
	CertDomain       string            `json:"CertDomain"`
	CertFile         string            `json:"CertFile"`
	KeyFile          string            `json:"KeyFile"`
	Provider         string            `json:"Provider"` // alidns, cloudflare, gandi, godaddy....
	Email            string            `json:"Email"`
	DNSEnv           map[string]string `json:"DNSEnv"`
	Certificate      []string          `json:"-"`
	Key              []string          `json:"-"`
}

func (c *CertConfig) UnmarshalJSON(data []byte) error {
	type alias CertConfig
	raw := struct {
		alias
		CertModeBack         string            `json:"cert_mode"`
		Mode                 string            `json:"mode"`
		RejectUnknownSniBack *bool             `json:"reject_unknown_sni"`
		CertDomainBack       string            `json:"cert_domain"`
		CertFileBack         string            `json:"cert_file"`
		KeyFileBack          string            `json:"key_file"`
		CertificatePath      string            `json:"certificate_path"`
		KeyPath              string            `json:"key_path"`
		ProviderBack         string            `json:"provider"`
		EmailBack            string            `json:"email"`
		DNSEnvBack           map[string]string `json:"dns_env"`
		Certificate          json.RawMessage   `json:"certificate"`
		Key                  json.RawMessage   `json:"key"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*c = CertConfig(raw.alias)
	if c.CertMode == "" {
		c.CertMode = raw.CertModeBack
	}
	if c.CertMode == "" {
		c.CertMode = raw.Mode
	}
	if raw.RejectUnknownSniBack != nil {
		c.RejectUnknownSni = *raw.RejectUnknownSniBack
	}
	if c.CertDomain == "" {
		c.CertDomain = raw.CertDomainBack
	}
	if c.CertFile == "" {
		c.CertFile = raw.CertFileBack
	}
	if c.CertFile == "" {
		c.CertFile = raw.CertificatePath
	}
	if c.KeyFile == "" {
		c.KeyFile = raw.KeyFileBack
	}
	if c.KeyFile == "" {
		c.KeyFile = raw.KeyPath
	}
	if c.Provider == "" {
		c.Provider = raw.ProviderBack
	}
	if c.Email == "" {
		c.Email = raw.EmailBack
	}
	if len(c.DNSEnv) == 0 && len(raw.DNSEnvBack) > 0 {
		c.DNSEnv = raw.DNSEnvBack
	}
	certificates, err := unmarshalStringSlice(raw.Certificate)
	if err != nil {
		return err
	}
	keys, err := unmarshalStringSlice(raw.Key)
	if err != nil {
		return err
	}
	c.Certificate = certificates
	c.Key = keys
	return nil
}

func unmarshalStringSlice(data json.RawMessage) ([]string, error) {
	if len(data) == 0 || string(data) == "null" {
		return nil, nil
	}
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		if single == "" {
			return nil, nil
		}
		return []string{single}, nil
	}
	var many []string
	if err := json.Unmarshal(data, &many); err == nil {
		return many, nil
	}
	var raw []any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	values := make([]string, 0, len(raw))
	for i := range raw {
		item, ok := raw[i].(string)
		if !ok {
			return nil, fmt.Errorf("item %d is not string", i)
		}
		values = append(values, item)
	}
	return values, nil
}

func NewCertConfig() *CertConfig {
	return &CertConfig{
		CertMode: "none",
	}
}
